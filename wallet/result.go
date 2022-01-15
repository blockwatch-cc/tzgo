// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package wallet

import (
    "context"
    "errors"
    "sync"

    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/tezos"
)

var (
    Canceled    = errors.New("operation confirm canceled")
    TTLExceeded = errors.New("operation ttl exceeded")
)

type Result struct {
    Block tezos.BlockHash
    List  int
    Pos   int
    Op    *rpc.Operation
}

// Cost returns the sum of costs across all batched and internal operations.
func (r *Result) Cost() rpc.OperationCost {
    if r.Op != nil {
        return r.Op.Cost()
    }
    return rpc.OperationCost{}
}

type FutureResult struct {
    oh     tezos.OpHash    // the operation hash to watch
    block  tezos.BlockHash // the block hash where op was included
    list   int             // the list where op was included
    pos    int             // the list position where op was included
    err    error           // saves any error
    ttl    int64           // number of blocks before wait fails
    wait   int64           // number of confirmations required
    blocks int64           // number of confirmation blocks seen
    mon    *Monitor        // blockchain monitor
    subId  int             // monitor subscription id
    done   chan struct{}   // channel used to signal completion
    once   sync.Once       // ensures only one completion state exists
}

func NewFutureResult(oh tezos.OpHash) *FutureResult {
    return &FutureResult{
        oh:   oh,
        wait: 1,
        done: make(chan struct{}),
    }
}

func (r *FutureResult) Listen(mon *Monitor) {
    if r.mon == nil {
        r.mon = mon
        r.subId = mon.Subscribe(r.oh, r.callback)
    }
}

func (r *FutureResult) Cancel() {
    r.once.Do(func() {
        if r.subId > 0 {
            r.mon.Unsubscribe(r.subId)
            r.err = Canceled
            r.subId = 0
        }
        close(r.done)
    })
}

func (r *FutureResult) WithConfirmations(n int64) *FutureResult {
    r.wait = n
    return r
}

func (r *FutureResult) WithTTL(n int64) *FutureResult {
    r.ttl = n
    return r
}

func (r *FutureResult) Confirmations() int64 {
    return r.blocks
}

func (r *FutureResult) Done() <-chan struct{} {
    return r.done
}

func (r *FutureResult) Err() error {
    return r.err
}

func (r *FutureResult) GetResult(ctx context.Context, c *rpc.Client) (*Result, error) {
    if r.err != nil {
        return nil, r.err
    }
    res := &Result{
        Block: r.block,
        Pos:   r.pos,
        List:  r.list,
    }
    op, err := c.GetBlockOperation(ctx, r.block, r.list, r.pos)
    if err != nil {
        return res, err
    }
    res.Op = op
    return res, nil
}

func (r *FutureResult) Wait() {
    <-r.done
}

func (r *FutureResult) WaitContext(ctx context.Context) {
    select {
    case <-ctx.Done():
    case <-r.done:
    }
}

func (r *FutureResult) callback(block tezos.BlockHash, list, pos int, force bool) bool {
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
