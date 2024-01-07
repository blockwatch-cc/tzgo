// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package signer

import (
	"context"
	"errors"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/tezos"
)

var ErrAddressMismatch = errors.New("signer: address mismatch")

type MemorySigner struct {
	key tezos.PrivateKey
}

func NewFromKey(k tezos.PrivateKey) *MemorySigner {
	return &MemorySigner{
		key: k,
	}
}

func (s MemorySigner) ListAddresses(_ context.Context) ([]tezos.Address, error) {
	return []tezos.Address{s.key.Address()}, nil
}

func (s MemorySigner) GetKey(_ context.Context, addr tezos.Address) (tezos.Key, error) {
	pk := s.key.Public()
	if !pk.Address().Equal(addr) {
		return tezos.InvalidKey, ErrAddressMismatch
	}
	return pk, nil
}

func (s MemorySigner) SignMessage(_ context.Context, addr tezos.Address, msg string) (tezos.Signature, error) {
	if !s.key.Address().Equal(addr) {
		return tezos.InvalidSignature, ErrAddressMismatch
	}
	op := codec.NewOp().
		WithBranch(tezos.ZeroBlockHash).
		WithContents(&codec.FailingNoop{
			Arbitrary: msg,
		})
	digest := tezos.Digest(op.Bytes())
	return s.key.Sign(digest[:])
}

func (s MemorySigner) SignOperation(_ context.Context, addr tezos.Address, op *codec.Op) (tezos.Signature, error) {
	if !s.key.Address().Equal(addr) {
		return tezos.InvalidSignature, ErrAddressMismatch
	}
	err := op.Sign(s.key)
	return op.Signature, err
}

func (s MemorySigner) SignBlock(_ context.Context, addr tezos.Address, head *codec.BlockHeader) (tezos.Signature, error) {
	if !s.key.Address().Equal(addr) {
		return tezos.InvalidSignature, ErrAddressMismatch
	}
	err := head.Sign(s.key)
	return head.Signature, err
}
