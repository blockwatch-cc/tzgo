// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package task

import (
	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/internal/compose/alpha"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"

	"github.com/pkg/errors"
)

var _ alpha.TaskBuilder = (*CallTask)(nil)

func init() {
	alpha.RegisterTask("call", NewCallTask)
}

type CallTask struct {
	TargetTask
}

func NewCallTask() alpha.TaskBuilder {
	return &CallTask{}
}

func (c *CallTask) Type() string {
	return "call"
}

func (t *CallTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}
	params, err := alpha.ParseParams(ctx, task)
	if err != nil {
		return nil, nil, errors.Wrap(err, "params")
	}
	opts := rpc.NewCallOptions()
	opts.Signer = signer.NewFromKey(t.Key)
	op := codec.NewOp().
		WithSource(t.Source).
		WithCallExt(t.Destination, micheline.Parameters{
			Entrypoint: task.Params.Entrypoint,
			Value:      *params,
		}, int64(task.Amount))
	return op, opts, nil
}

func (t *CallTask) Validate(ctx compose.Context, task alpha.Task) error {
	if err := t.parse(ctx, task); err != nil {
		return err
	}
	if _, err := alpha.ParseParams(ctx, task); err != nil {
		return errors.Wrap(err, "params")
	}
	return nil
}

func (t *CallTask) parse(ctx compose.Context, task alpha.Task) (err error) {
	if err = t.TargetTask.parse(ctx, task); err != nil {
		return err
	}
	return
}
