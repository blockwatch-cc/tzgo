// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc
package task

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/internal/compose/alpha"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"

	"github.com/pkg/errors"
)

var _ alpha.TaskBuilder = (*DoubleEndorseTask)(nil)

func init() {
	alpha.RegisterTask("double_endorse", NewDoubleEndorseTask)
}

type DoubleEndorseTask struct {
	TargetTask
	BakerKey tezos.PrivateKey
}

func NewDoubleEndorseTask() alpha.TaskBuilder {
	return &DoubleEndorseTask{}
}

func (t *DoubleEndorseTask) Type() string {
	return "double_endorse"
}

func (t *DoubleEndorseTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}

	// wait for endorsing rights to appear, remember block and slot
	ctx.Log.Infof("Waiting for endorsing rights for %s...", t.BakerKey.Address())
	var (
		slot int
		head *rpc.BlockHeaderLogEntry
	)
	done := make(chan struct{})
	_, err := ctx.SubscribeBlocks(func(h *rpc.BlockHeaderLogEntry, _ int64, _ int, _ int, _ bool) bool {
		s, ok, err := t.fetchEndorsingRights(ctx, t.BakerKey.Address(), h.Hash)
		if err != nil {
			ctx.Log.Warnf("fetch endorsing rights: %v", err)
		} else if ok {
			head = h
			slot = s
			close(done)
			return true
		}
		ctx.Log.Debugf("No rights in block %d", h.Level)
		return false
	})
	if err != nil {
		return nil, nil, err
	}
	select {
	case <-done:
		ctx.Log.Infof("Found endorsing rights in block=%d slot=%d", head.Level, slot)
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

	// produce random endorsements
	e1, h1 := t.randomEndorsement(ctx, head, slot)
	e2, h2 := t.randomEndorsement(ctx, head, slot)

	// order endorsements by op hash
	if bytes.Compare(h1[:], h2[:]) > 0 {
		e1, e2 = e2, e1
	}

	// pack into evidence op
	op := codec.NewOp().
		WithSource(t.Source). // required for compose simulation mode
		WithContents(&codec.TenderbakeDoubleEndorsementEvidence{
			Op1: e1,
			Op2: e2,
		})

	// wait one block for sending
	ctx.Log.Debug("Wait next block")
	if err := ctx.WaitNumBlocks(1); err != nil {
		return nil, nil, err
	}

	opts := rpc.NewCallOptions()
	opts.Signer = signer.NewFromKey(t.Key)
	opts.IgnoreLimits = true

	return op, opts, nil
}

func (t *DoubleEndorseTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *DoubleEndorseTask) parse(ctx compose.Context, task alpha.Task) (err error) {
	if err = t.TargetTask.parse(ctx, task); err != nil {
		return err
	}
	if t.BakerKey, err = ctx.ResolvePrivateKey(task.Destination); err != nil {
		return errors.Wrap(err, "destination key")
	}
	return
}

func (t *DoubleEndorseTask) randomEndorsement(ctx compose.Context, head *rpc.BlockHeaderLogEntry, slot int) (codec.TenderbakeInlinedEndorsement, tezos.OpHash) {
	// generate a random endorsement for the latest block
	e := codec.TenderbakeEndorsement{
		Slot:  int16(slot),
		Level: int32(head.Level),
		Round: int32(head.Round()),
	}
	rand.Read(e.BlockPayloadHash[:])

	// create a tenderbake endorsement
	p := ctx.Params()
	op := codec.NewOp().
		WithParams(p).          // use protocol params for correct op tags
		WithChainId(p.ChainId). // Tenderbake endorsements need chain id
		WithContents(&e).
		WithBranch(head.Hash)

	// sign the endorsement
	if err := op.Sign(t.BakerKey); err != nil {
		ctx.Log.Errorf("signing random endorsement: %v", err)
	}

	// return as inlined endorsement
	ed := codec.TenderbakeInlinedEndorsement{
		Branch:      head.Hash,
		Endorsement: e,
		Signature:   op.Signature,
	}
	return ed, op.Hash()
}

func (t *DoubleEndorseTask) fetchEndorsingRights(ctx compose.Context, addr tezos.Address, id tezos.BlockHash) (int, bool, error) {
	u := fmt.Sprintf("/chains/main/blocks/%s/helpers/endorsing_rights?delegate=%s", id, addr)
	var rights []struct {
		Level     int64                `json:"level"`
		Delegates []rpc.EndorsingRight `json:"delegates"`
	}
	if err := ctx.Fetch(u, &rights); err != nil {
		return 0, false, err
	}
	if len(rights) == 0 {
		return 0, false, nil
	}
	return rights[0].Delegates[0].FirstSlot, true, nil
}
