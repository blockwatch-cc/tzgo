// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package task

import (
	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/internal/compose/alpha"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"

	"github.com/pkg/errors"
)

var _ alpha.TaskBuilder = (*FinalizeUnstakeTask)(nil)

func init() {
	alpha.RegisterTask("finalize_unstake", NewFinalizeUnstakeTask)
}

type FinalizeUnstakeTask struct {
	BaseTask
}

func NewFinalizeUnstakeTask() alpha.TaskBuilder {
	return &FinalizeUnstakeTask{}
}

func (t *FinalizeUnstakeTask) Type() string {
	return "finalize_unstake"
}

func (t *FinalizeUnstakeTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}
	opts := rpc.NewCallOptions()
	opts.Signer = signer.NewFromKey(t.Key)
	op := codec.NewOp().
		WithSource(t.Source).
		WithFinalizeUnstake()
	return op, opts, nil
}

func (t *FinalizeUnstakeTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *FinalizeUnstakeTask) parse(ctx compose.Context, task alpha.Task) (err error) {
	return t.BaseTask.parse(ctx, task)
}
