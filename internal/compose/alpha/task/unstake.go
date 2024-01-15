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

var _ alpha.TaskBuilder = (*UnstakeTask)(nil)

func init() {
	alpha.RegisterTask("unstake", NewUnstakeTask)
}

type UnstakeTask struct {
	BaseTask
	Amount int64
}

func NewUnstakeTask() alpha.TaskBuilder {
	return &UnstakeTask{}
}

func (t *UnstakeTask) Type() string {
	return "unstake"
}

func (t *UnstakeTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}
	opts := rpc.NewCallOptions()
	opts.Signer = signer.NewFromKey(t.Key)
	op := codec.NewOp().
		WithSource(t.Source).
		WithUnstake(t.Amount)
	return op, opts, nil
}

func (t *UnstakeTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *UnstakeTask) parse(ctx compose.Context, task alpha.Task) (err error) {
	if err = t.BaseTask.parse(ctx, task); err != nil {
		return err
	}
	t.Amount = int64(task.Amount)
	return
}
