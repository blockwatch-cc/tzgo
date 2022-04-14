// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package signer

import (
    "context"

    "blockwatch.cc/tzgo/codec"
    "blockwatch.cc/tzgo/tezos"
)

type MemorySigner struct {
    key tezos.PrivateKey
}

func NewFromKey(k tezos.PrivateKey) *MemorySigner {
    return &MemorySigner{
        key: k,
    }
}

func (s MemorySigner) Address(_ context.Context) (tezos.Address, error) {
    return s.key.Address(), nil
}

func (s MemorySigner) Key(_ context.Context) (tezos.Key, error) {
    return s.key.Public(), nil
}

func (s MemorySigner) SignMessage(_ context.Context, msg string) (tezos.Signature, error) {
    digest := tezos.Digest([]byte(msg))
    return s.key.Sign(digest[:])
}

func (s MemorySigner) SignOperation(_ context.Context, op *codec.Op) (tezos.Signature, error) {
    err := op.Sign(s.key)
    return op.Signature, err
}

func (s MemorySigner) SignBlock(_ context.Context, head *codec.BlockHeader) (tezos.Signature, error) {
    err := head.Sign(s.key)
    return head.Signature, err
}
