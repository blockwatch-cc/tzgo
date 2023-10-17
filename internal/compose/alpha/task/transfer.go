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

var _ alpha.TaskBuilder = (*TransferTask)(nil)

func init() {
	alpha.RegisterTask("transfer", NewTransferTask)
}

type TransferTask struct {
	TargetTask
	Amount int64
}

func NewTransferTask() alpha.TaskBuilder {
	return &TransferTask{}
}

func (t *TransferTask) Type() string {
	return "transfer"
}

func (t *TransferTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}
	opts := rpc.NewCallOptions()
	opts.Signer = signer.NewFromKey(t.Key)
	op := codec.NewOp().WithSource(t.Source).WithTransfer(t.Destination, t.Amount)
	return op, opts, nil
}

func (t *TransferTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *TransferTask) parse(ctx compose.Context, task alpha.Task) (err error) {
	if err = t.TargetTask.parse(ctx, task); err != nil {
		return err
	}
	t.Amount = int64(task.Amount)
	return
}
