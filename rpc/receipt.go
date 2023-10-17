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
	Block  tezos.BlockHash
	Height int64
	List   int
	Pos    int
	Op     *Operation
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

// IsSuccess returns true when all operations in this group have been applied successfully.
func (r *Receipt) IsSuccess() bool {
	for _, v := range r.Op.Contents {
		switch v.Result().Status {
		case tezos.OpStatusApplied:
			return true
		case tezos.OpStatusInvalid:
			// only manager ops contain a status field
			return true
		default:
			return false
		}
	}
	return true
}

// Error returns the first execution error found in this operation group or one of
// its internal results that is of status failed. This helper only exports the error
// as GenericError. To access error details or all errors, visit
// r.Op.Contents[].OperationResult.Errors[] and
// r.Op.Contents[].Metadata.InternalResults.Result.Errors[]
func (r *Receipt) Error() error {
	for _, v := range r.Op.Contents {
		res := v.Result()
		if len(res.Errors) > 0 && res.Status != tezos.OpStatusApplied {
			return res.Errors[len(res.Errors)-1].GenericError
		}
		for _, vv := range v.Meta().InternalResults {
			res := vv.Result
			if len(res.Errors) > 0 && res.Status != tezos.OpStatusApplied {
				return res.Errors[len(res.Errors)-1].GenericError
			}
		}
	}
	return nil
}

// OriginatedContract returns the first contract address deployed by the operation.
func (r *Receipt) OriginatedContract() (tezos.Address, bool) {
	if r.IsSuccess() {
		for _, contents := range r.Op.Contents {
			if contents.Kind() == tezos.OpTypeOrigination {
				result := contents.Result()
				if len(result.OriginatedContracts) > 0 {
					return result.OriginatedContracts[0], true
				}
			}
		}
	}
	return tezos.InvalidAddress, false
}

// MinLimits returns a list of individual operation costs mapped to limits for use
// in simulation results. Fee is reset to zero to prevent higher simulation fee from
// spilling over into real fees paid.
func (r *Receipt) MinLimits() []tezos.Limits {
	lims := make([]tezos.Limits, len(r.Op.Contents))
	for i, v := range r.Op.Costs() {
		lims[i].Fee = 0
		lims[i].GasLimit = v.GasUsed
		lims[i].StorageLimit = v.StorageUsed + v.AllocationBurn/tezos.DefaultParams.CostPerByte
	}
	return lims
}

type Result struct {
	oh     tezos.OpHash    // the operation hash to watch
	block  tezos.BlockHash // the block hash where op was included
	height int64           // block height
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
		Block:  r.block,
		Height: r.height,
		Pos:    r.pos,
		List:   r.list,
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

func (r *Result) WaitContext(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		r.err = context.Canceled
		return false
	case <-r.done:
		return true
	}
}

func (r *Result) callback(block *BlockHeaderLogEntry, height int64, list, pos int, force bool) bool {
	if force {
		r.block = block.Hash
		r.height = height
		r.list = list
		r.pos = pos
		return false
	}
	if !r.block.IsValid() {
		r.block = block.Hash
		r.height = height
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
