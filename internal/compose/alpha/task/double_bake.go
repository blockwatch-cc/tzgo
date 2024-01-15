// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc
package task

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/internal/compose/alpha"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"

	"github.com/pkg/errors"
)

var _ alpha.TaskBuilder = (*DoubleBakeTask)(nil)

func init() {
	alpha.RegisterTask("double_bake", NewDoubleBakeTask)
}

type DoubleBakeTask struct {
	TargetTask
	BakerKey tezos.PrivateKey
}

func NewDoubleBakeTask() alpha.TaskBuilder {
	return &DoubleBakeTask{}
}

func (t *DoubleBakeTask) Type() string {
	return "double_bake"
}

func (t *DoubleBakeTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}

	// wait for baking rights to appear for the next block, the block must
	// actually be baked by this baker, so we look for round 0 only
	// this means to test this on sandbox the source/baker must be a
	// sandbox baker
	ctx.Log.Infof("Waiting for round-0 baking rights for %s...", t.BakerKey.Address())
	var (
		round int
		head  *rpc.BlockHeaderLogEntry
	)
	done := make(chan struct{})
	_, err := ctx.SubscribeBlocks(func(h *rpc.BlockHeaderLogEntry, _ int64, _ int, _ int, _ bool) bool {
		r, ok, err := t.fetchBakingRights(ctx, t.BakerKey.Address(), h.Hash)
		if err != nil {
			ctx.Log.Warnf("fetch baking rights: %v", err)
		} else if ok && r == 0 {
			head = h
			round = r
			close(done)
			return true
		}
		ctx.Log.Debugf("No round-0 rights in block %d", h.Level)
		return false
	})
	if err != nil {
		return nil, nil, err
	}
	select {
	case <-done:
		ctx.Log.Infof("Found baking rights in block=%d round=%d", head.Level+1, round)
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}

	// produce random block headers (for NEXT block!)
	b1 := t.randomBlock(ctx, head, round)
	b2 := t.randomBlock(ctx, head, round)

	// order blocks by hash
	if bytes.Compare(b1.Hash().Bytes(), b2.Hash().Bytes()) > 0 {
		b1, b2 = b2, b1
	}

	// pack into evidence op
	op := codec.NewOp().
		WithSource(t.Source). // required for compose simulation mode
		WithContents(&codec.DoubleBakingEvidence{
			Bh1: b1,
			Bh2: b2,
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

func (t *DoubleBakeTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *DoubleBakeTask) parse(ctx compose.Context, task alpha.Task) (err error) {
	if err = t.TargetTask.parse(ctx, task); err != nil {
		return err
	}
	if t.BakerKey, err = ctx.ResolvePrivateKey(task.Destination); err != nil {
		return errors.Wrap(err, "destination key")
	}
	return
}

func (t *DoubleBakeTask) randomBlock(ctx compose.Context, head *rpc.BlockHeaderLogEntry, round int) codec.BlockHeader {
	// generate a random block header
	h := codec.BlockHeader{
		Level:            int32(head.Level + 1),
		Proto:            byte(head.Proto),
		Predecessor:      head.Hash,
		Timestamp:        head.Timestamp,
		ValidationPass:   byte(head.ValidationPass),
		Fitness:          head.Fitness,
		ProofOfWorkNonce: head.Pow(),
		PayloadRound:     round,
		OperationsHash:   head.OperationsHash,
		Context:          head.Context,
		LbVote:           tezos.FeatureVotePass,
		AiVote:           tezos.FeatureVotePass,
	}
	// adjust fitness
	f2 := make([]byte, len(h.Fitness[1]))
	binary.BigEndian.PutUint32(f2, uint32(head.Level+1))
	h.Fitness[1] = f2

	// randomize payload hash
	rand.Read(h.PayloadHash[:])
	h.WithChainId(ctx.Params().ChainId) // Tenderbake block signing needs chain id

	// sign the block
	if err := h.Sign(t.BakerKey); err != nil {
		ctx.Log.Errorf("signing random block: %v", err)
	}
	return h
}

func (t *DoubleBakeTask) fetchBakingRights(ctx compose.Context, addr tezos.Address, id tezos.BlockHash) (int, bool, error) {
	u := fmt.Sprintf("/chains/main/blocks/%s/helpers/baking_rights?delegate=%s&max_round=64", id, addr)
	var rights []struct {
		Round int `json:"round"`
	}
	if err := ctx.Fetch(u, &rights); err != nil {
		return 0, false, err
	}
	if len(rights) == 0 {
		return 0, false, nil
	}
	return rights[0].Round, true, nil
}
