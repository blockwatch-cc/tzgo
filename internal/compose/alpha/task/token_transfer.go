// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package task

import (
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/internal/compose/alpha"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"

	"github.com/pkg/errors"
)

var _ alpha.TaskBuilder = (*TokenTransferTask)(nil)

func init() {
	alpha.RegisterTask("token_transfer", NewTokenTransferTask)
}

type TokenTransferTask struct {
	TargetTask
	Standard  string
	From      tezos.Address
	Receivers []TokenReceiver
}

type TokenReceiver struct {
	Address tezos.Address
	Amount  tezos.Z
	TokenId tezos.Z
}

func NewTokenTransferTask() alpha.TaskBuilder {
	return &TokenTransferTask{}
}

func (t *TokenTransferTask) Type() string {
	return "token_transfer"
}

func (t *TokenTransferTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}
	var xfer codec.Operation
	switch t.Standard {
	case "fa2", "":
		args := contract.NewFA2TransferArgs()
		for _, r := range t.Receivers {
			args.WithTransfer(t.From, r.Address, r.TokenId, r.Amount)
		}
		xfer = args.
			Optimize().
			WithSource(t.Source).
			WithDestination(t.Destination).
			Encode()
	case "fa1", "fa12", "fa1.2":
		xfer = contract.NewFA1TransferArgs().
			WithTransfer(t.From, t.Receivers[0].Address, t.Receivers[0].Amount).
			WithSource(t.Source).
			WithDestination(t.Destination).
			Encode()
	}

	opts := rpc.NewCallOptions()
	opts.Signer = signer.NewFromKey(t.Key)
	op := codec.NewOp().WithContents(xfer)
	return op, opts, nil
}

func (t *TokenTransferTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *TokenTransferTask) parse(ctx compose.Context, task alpha.Task) (err error) {
	if err = t.TargetTask.parse(ctx, task); err != nil {
		return err
	}
	if t.Standard, err = ctx.ResolveString(task.Args["standard"]); err != nil {
		return errors.Wrap(err, "standard")
	}
	switch t.Standard {
	case "fa2", "", "fa1", "fa12", "fa1.2":
		// skip
	default:
		return fmt.Errorf("unsupported token standard %s", t.Standard)
	}
	if t.From, err = ctx.ResolveAddress(task.Args["from"]); err != nil {
		return errors.Wrap(err, "from")
	}
	if val := task.Args["receivers"]; val == nil {
		var tr TokenReceiver
		if tr.Address, err = ctx.ResolveAddress(task.Args["to"]); err != nil {
			return errors.Wrap(err, "to")
		}
		if tr.Amount, err = ctx.ResolveZ(task.Args["amount"]); err != nil {
			return errors.Wrap(err, "amount")
		}
		// only required for fa2
		switch t.Standard {
		case "fa2", "":
			if tr.TokenId, err = ctx.ResolveZ(task.Args["token_id"]); err != nil {
				return errors.Wrap(err, "token_id")
			}
		}
		t.Receivers = append(t.Receivers, tr)
	} else {
		recvs, ok := val.([]any)
		if !ok {
			return fmt.Errorf("invalid type %T for receivers, expected list(map)", val)
		}
		for i, v := range recvs {
			var tr TokenReceiver
			recv, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("receiver[%d]: invalid type %T for receiver, expected map[string]string", i, val)
			}
			if tr.Address, err = ctx.ResolveAddress(recv["to"]); err != nil {
				return fmt.Errorf("receiver[%d] to: %v", i, err)
			}
			if tr.Amount, err = ctx.ResolveZ(recv["amount"]); err != nil {
				return fmt.Errorf("receiver[%d] amount: %v", i, err)
			}
			switch t.Standard {
			case "fa2", "":
				if tr.TokenId, err = ctx.ResolveZ(recv["token_id"]); err != nil {
					return fmt.Errorf("receiver[%d] amount: %v", i, err)
				}
			}
			t.Receivers = append(t.Receivers, tr)
		}
	}
	return
}
