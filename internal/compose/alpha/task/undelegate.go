// Copyright (c) 2023 Blockwatch Data Inc.
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

var _ alpha.TaskBuilder = (*UndelegateTask)(nil)

func init() {
	alpha.RegisterTask("undelegate", NewUndelegateTask)
}

type UndelegateTask struct {
	BaseTask
}

func NewUndelegateTask() alpha.TaskBuilder {
	return &UndelegateTask{}
}

func (t *UndelegateTask) Type() string {
	return "undelegate"
}

func (t *UndelegateTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}
	opts := rpc.DefaultOptions
	opts.Signer = signer.NewFromKey(t.Key)
	op := codec.NewOp().WithSource(t.Source).WithUndelegation()
	return op, &opts, nil
}

func (t *UndelegateTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *UndelegateTask) parse(ctx compose.Context, task alpha.Task) error {
	return t.BaseTask.parse(ctx, task)
}
