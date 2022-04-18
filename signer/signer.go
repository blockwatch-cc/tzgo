// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package signer

import (
	"context"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/tezos"
)

type Signer interface {
	// Return a list of addresses the signer manages.
	ListAddresses(context.Context) ([]tezos.Address, error)

	// Returns the public key for a managed address. Required for reveal ops.
	GetKey(context.Context, tezos.Address) (tezos.Key, error)

	// Sign an arbitrary text message wrapped into a failing noop
	SignMessage(context.Context, tezos.Address, string) (tezos.Signature, error)

	// Sign an operation.
	SignOperation(context.Context, tezos.Address, *codec.Op) (tezos.Signature, error)

	// Sign a block header.
	SignBlock(context.Context, tezos.Address, *codec.BlockHeader) (tezos.Signature, error)
}
