// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package signer

import (
    "context"
    "net/http"

    "blockwatch.cc/tzgo/codec"
    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/signer"
    "blockwatch.cc/tzgo/tezos"
)

var _ signer.Signer = (*RemoteSigner)(nil)

type RemoteSigner struct {
    c *rpc.Client
}

// New creates a new remote signer client and initializes it with the remote url.
// Users may pass an optional http client with a custom configuration, otherwise
// the http.DefaultClient is used.
func New(url string, client *http.Client) (*RemoteSigner, error) {
    c, err := rpc.NewClient(url, client)
    if err != nil {
        return nil, err
    }
    return &RemoteSigner{c: c}, nil
}

// ListAddresses returns a list of addresses the remote signer can produce signatures for.
func (s RemoteSigner) ListAddresses(ctx context.Context) ([]tezos.Address, error) {
    type response struct {
        Addrs []tezos.Address `json:"authorized_keys"`
    }
    var resp response
    err := s.c.Get(ctx, "/authorized_keys", &resp)
    if err != nil {
        return nil, err
    }
    return resp.Addrs, nil
}

// GetKey returns the public key associated with address.
func (s RemoteSigner) GetKey(ctx context.Context, address tezos.Address) (tezos.Key, error) {
    type response struct {
        Pk tezos.Key `json:"public_key"`
    }
    var resp response
    err := s.c.Get(ctx, "/keys/"+address.String(), &resp)
    return resp.Pk, err
}

// SignMessage signs msg for address by wrapping it into a failing noop operation
// with zero branch hash. This prevents unintended signature of message bytes that
// represent a valid transaction.
//
// Note that most remote signers for Tezos do not support signing of operation kinds other
// than baking related operations.
func (s RemoteSigner) SignMessage(ctx context.Context, address tezos.Address, msg string) (tezos.Signature, error) {
    op := codec.NewOp().
        WithBranch(tezos.ZeroBlockHash).
        WithContents(&codec.FailingNoop{
            Arbitrary: msg,
        })
    return s.SignOperation(ctx, address, op)
}

// SignOperation signs operation op for address using the configured remote signer's
// REST API. For endorsements this call requires branch_id to be present.
//
// Note that most remote signers for Tezos do not support signing of operation kinds other
// than baking related operations.
func (s RemoteSigner) SignOperation(ctx context.Context, address tezos.Address, op *codec.Op) (tezos.Signature, error) {
    type response struct {
        Sig tezos.Signature `json:"signature"`
    }
    var resp response
    err := s.c.Post(ctx, "/keys/"+address.String(), tezos.HexBytes(op.WatermarkedBytes()), &resp)
    return resp.Sig, err
}

// SignOperation signs a block header for address using the configured remote signer's
// REST API. This call requires branch_id to be present.
func (s RemoteSigner) SignBlock(ctx context.Context, address tezos.Address, head *codec.BlockHeader) (tezos.Signature, error) {
    type response struct {
        Sig tezos.Signature `json:"signature"`
    }
    var resp response
    err := s.c.Post(ctx, "/keys/"+address.String(), tezos.HexBytes(head.WatermarkedBytes()), &resp)
    return resp.Sig, err
}
