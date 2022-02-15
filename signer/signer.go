// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package signer

import (
	"context"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/tezos"
)

type Signer interface {
	Address(context.Context) (tezos.Address, error) // returns address
	Key(context.Context) (tezos.Key, error)         // returns public key
	SignMessage(context.Context, string) (tezos.Signature, error)
	SignOperation(context.Context, *codec.Op) (tezos.Signature, error)
	SignBlock(context.Context, *codec.BlockHeader) (tezos.Signature, error)
}

// https://pkg.go.dev/github.com/cosmos/cosmos-sdk@v0.45.0/crypto/keyring#Keyring
// type Signer interface {
//     // Sign sign byte messages with a user key.
//     Sign(uid string, msg []byte) ([]byte, types.PubKey, error)

//     // SignByAddress sign byte messages with a user key providing the address.
//     SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error)
// }
