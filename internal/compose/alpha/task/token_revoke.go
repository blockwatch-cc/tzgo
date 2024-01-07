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

var _ alpha.TaskBuilder = (*TokenRevokeTask)(nil)

func init() {
	alpha.RegisterTask("token_revoke", NewTokenRevokeTask)
}

type TokenRevokeTask struct {
	TargetTask
	Spender  tezos.Address
	Standard string
	TokenId  tezos.Z
}

func NewTokenRevokeTask() alpha.TaskBuilder {
	return &TokenRevokeTask{}
}

func (t *TokenRevokeTask) Type() string {
	return "token_revoke"
}

func (t *TokenRevokeTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}
	var xfer codec.Operation
	switch t.Standard {
	case "fa2", "":
		xfer = contract.NewFA2ApprovalArgs().
			RemoveOperator(t.Source, t.Spender, t.TokenId).
			WithSource(t.Source).
			WithDestination(t.Destination).
			Encode()
	case "fa1", "fa12", "fa1.2":
		xfer = contract.NewFA1ApprovalArgs().
			Revoke(t.Spender).
			WithSource(t.Source).
			WithDestination(t.Destination).
			Encode()
	}

	opts := rpc.NewCallOptions()
	opts.Signer = signer.NewFromKey(t.Key)
	op := codec.NewOp().WithContents(xfer)
	return op, opts, nil
}

func (t *TokenRevokeTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *TokenRevokeTask) parse(ctx compose.Context, task alpha.Task) (err error) {
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
	if t.Spender, err = ctx.ResolveAddress(task.Args["spender"]); err != nil {
		return errors.Wrap(err, "spender")
	}
	// only required for fa2
	switch t.Standard {
	case "fa2", "":
		if t.TokenId, err = ctx.ResolveZ(task.Args["token_id"]); err != nil {
			return errors.Wrap(err, "token_id")
		}
	}
	return
}
