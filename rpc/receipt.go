// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"errors"
	"sync"

	"blockwatch.cc/tzgo/tezos"
)

var (
	Canceled    = errors.New("operation confirm canceled")
	TTLExceeded = errors.New("operation ttl exceeded")
)

type Receipt struct {
	Block tezos.BlockHash
	List  int
	Pos   int
	Op    *Operation
}

// TotalCosts returns the sum of costs across all batched and internal operations.
func (r *Receipt) TotalCosts() tezos.Costs {
	if r.Op != nil {
		return r.Op.TotalCosts()
	}
	return tezos.Costs{}
}

// Costs returns a list of individual costs for all batched operations.
func (r *Receipt) Costs() []tezos.Costs {
	if r.Op != nil {
		return r.Op.Costs()
	}
	return nil
}

// MapLimits returns a list of individual operation costs mapped to limits for use
// in simulation results.
func (r *Receipt) MapLimits() []tezos.Limits {
	if r.Op != nil {
		lims := make([]tezos.Limits, len(r.Op.Contents))
		for i, v := range r.Op.Costs() {
			lims[i].Fee = v.Fee
			lims[i].GasLimit = v.GasUsed
			lims[i].StorageLimit = v.StorageUsed
		}
		return lims
	}
	return nil
}

type Result struct {
	oh     tezos.OpHash    // the operation hash to watch
	block  tezos.BlockHash // the block hash where op was included
	list   int             // the list where op was included
	pos    int             // the list position where op was included
	err    error           // saves any error
	ttl    int64           // number of blocks before wait fails
	wait   int64           // number of confirmations required
	blocks int64           // number of confirmation blocks seen
	obs    *Observer       // blockchain observer
	subId  int             // monitor subscription id
	done   chan struct{}   // channel used to signal completion
	once   sync.Once       // ensures only one completion state exists
}

func NewResult(oh tezos.OpHash) *Result {
	return &Result{
		oh:   oh,
		wait: 1,
		done: make(chan struct{}),
	}
}

func (r *Result) Hash() tezos.OpHash {
	return r.oh
}

func (r *Result) Listen(o *Observer) {
	if o != nil {
		r.obs = o
		r.subId = r.obs.Subscribe(r.oh, r.callback)
	}
}

func (r *Result) Cancel() {
	r.once.Do(func() {
		if r.subId > 0 {
			r.obs.Unsubscribe(r.subId)
			r.err = Canceled
			r.subId = 0
		}
		close(r.done)
	})
}

func (r *Result) WithConfirmations(n int64) *Result {
	r.wait = n
	return r
}

func (r *Result) WithTTL(n int64) *Result {
	r.ttl = n
	return r
}

func (r *Result) Confirmations() int64 {
	return r.blocks
}

func (r *Result) Done() <-chan struct{} {
	return r.done
}

func (r *Result) Err() error {
	return r.err
}

func (r *Result) GetReceipt(ctx context.Context) (*Receipt, error) {
	if r.err != nil {
		return nil, r.err
	}
	rec := &Receipt{
		Block: r.block,
		Pos:   r.pos,
		List:  r.list,
	}
	if r.obs != nil {
		op, err := r.obs.c.GetBlockOperation(ctx, r.block, r.list, r.pos)
		if err != nil {
			return rec, err
		}
		rec.Op = op
	}
	return rec, nil
}

func (r *Result) Wait() {
	<-r.done
}

func (r *Result) WaitContext(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-r.done:
	}
}

func (r *Result) callback(block tezos.BlockHash, list, pos int, force bool) bool {
	if force {
		r.block = block.Clone()
		r.list = list
		r.pos = pos
		return false
	}
	if !r.block.IsValid() {
		r.block = block.Clone()
		r.list = list
		r.pos = pos
	}
	r.blocks++
	if r.ttl > 0 && r.blocks >= r.ttl {
		r.once.Do(func() {
			r.err = TTLExceeded
			r.subId = 0
			close(r.done)
		})
		return true
	}
	if r.blocks >= r.wait {
		r.once.Do(func() {
			r.subId = 0
			close(r.done)
		})
		return true
	}
	return false
}
