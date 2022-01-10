// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package wallet

import (
    "bytes"
    "context"
    "encoding/hex"
    // "encoding/json"
    "fmt"

    "blockwatch.cc/tzgo/codec"
    // "blockwatch.cc/tzgo/micheline"
    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/tezos"
)

type RunOperationRequest struct {
    Operation *codec.Op         `json:"operation"`
    ChainId   tezos.ChainIdHash `json:"chain_id"`
}

// Simulate dry-runs the execution of the operation against the current state
// of a Tezos node in order to estimate execution costs and fees (fee/burn/gas/storage).
func Simulate(ctx context.Context, c *rpc.Client, o *codec.Op) (*Result, error) {
    sim := &codec.Op{
        Branch:    o.Branch,
        Contents:  o.Contents,
        Signature: tezos.ZeroSignature,
        TTL:       o.TTL,
        Params:    o.Params,
    }

    if !sim.Branch.IsValid() {
        ofs := o.Params.MaxOperationsTTL - sim.TTL
        hash, err := c.GetBlockHash(ctx, rpc.NewBlockOffset(rpc.Head, -ofs))
        if err != nil {
            return nil, err
        }
        sim.Branch = hash
    }

    req := RunOperationRequest{
        Operation: sim,
        ChainId:   c.ChainId,
    }
    resp := &rpc.Operation{}
    if err := c.RunOperation(ctx, rpc.Head, req, resp); err != nil {
        return nil, err
    }

    res := &Result{
        Op: resp,
    }
    return res, nil
}

// Validate compares local serializiation against remote RPC serialization of the
// operation and returns an error on mismatch.
func Validate(ctx context.Context, c *rpc.Client, o *codec.Op) error {
    op := &codec.Op{
        Branch:   o.Branch,
        Contents: o.Contents,
    }
    local := op.Bytes()
    var remote tezos.HexBytes
    if err := c.ForgeOperation(ctx, rpc.Head, op, &remote); err != nil {
        return err
    }
    if bytes.Compare(local, remote.Bytes()) != 0 {
        return fmt.Errorf("tezos: mismatch between local and remote serialized operations:\n local=%s\n remote=%s",
            hex.EncodeToString(local), hex.EncodeToString(remote))
    }
    return nil
}

// Broadcast sends the signed operation to network and returns the operation hash
// on successful pre-validation.
func Broadcast(ctx context.Context, c *rpc.Client, o *codec.Op) (tezos.OpHash, error) {
    return c.BroadcastOperation(ctx, o.Bytes())
}

// defined in rpc package
// see script_tz_error_registration.ml
// type OperationError struct {
//     ID                  string          `json:"id"`
//     Kind                string          `json:"kind"`
//     Contract            string          `json:"contract,omitempty"`
//     Raw                 json.RawMessage `json:"-"`
// ........
// specialized error contents for different ids
//     Expected            int64           `json:"expected,string,omitempty"`
//     Found               int64           `json:"found,string,omitempty"`
//     Location            int64           `json:"location,omitempty"`
//     Loc                 int64           `json:"loc,omitempty"`
//     BigMap              int64           `json:"big_map,omitempty"`
//     ExpectedForm        *micheline.Prim `json:"expectedForm,omitempty"`
//     WrongExpression     *micheline.Prim `json:"wrong_expression,omitempty"`
//     WrongExpression2    *micheline.Prim `json:"wrongExpression,omitempty"`
//     ExpectedType        *micheline.Prim `json:"expected_type,omitempty"`
//     WrongType           *micheline.Prim `json:"wrong_type,omitempty"`
//     FirstType           *micheline.Prim `json:"first_type,omitempty"`
//     OtherType           *micheline.Prim `json:"other_type,omitempty"`
//     BodyType            *micheline.Prim `json:"body_type,omitempty"`
//     BeforeStack         json.RawMessage `json:"bef_stack,omitempty"`
//     AfterStack          json.RawMessage `json:"aft_stack,omitempty"`
//     TypeSize            uint16          `json:"type_size,omitempty"`
//     MaximumTypeSize     uint16          `json:"maximum_type_size,omitempty"`
//     Identifier          string          `json:"identifier,omitempty"`
//     IllTypedExpression  *micheline.Prim `json:"ill_typed_expression,omitempty"`
//     IllFormedExpression *micheline.Prim `json:"ill_formed_expression,omitempty"`
//     IllTypedCode        *micheline.Prim `json:"ill_typed_code,omitempty"`
//     TypeMap             json.RawMessage `json:"type_map,omitempty"`
// }

// type RunCodeRequest struct {
//     Script     micheline.Prim    `json:"script"`
//     Storage    micheline.Prim    `json:"storage"`
//     Input      micheline.Prim    `json:"input"`
//     Amount     tezos.N           `json:"amount"`
//     Balance    tezos.N           `json:"balance"`
//     ChainId    tezos.ChainIdHash `json:"chain_id"`
//     Source     *tezos.Address    `json:"source,omitempty"`
//     Payer      *tezos.Address    `json:"payer,omitempty"`
//     Gas        *tezos.N          `json:"gas,omitempty"`
//     Entrypoint string            `json:"entrypoint,omitempty"`
// }

// // RunCodeResponse -
// type RunCodeResponse struct {
//     Operations []rpc.Operation      `json:"operations"`
//     Storage    micheline.Prim       `json:"storage"`
//     BigmapDiff micheline.BigmapDiff `json:"big_map_diff,omitempty"`
// }

// type RunCodeError struct {
//     ID string `json:"id"`
// }

// type TracedCodeResponse struct {
//     RunCodeResponse
//     Trace Trace `json:"trace"`
// }

// type Trace struct {
//     Location int     `json:"location"`
//     Gas      tezos.N `json:"gas"`
//     Stack    []Stack `json:"stack"`
// }

// type Stack struct {
//     Item  *json.RawMessage `json:"item"`
//     Annot string           `json:"annot,omitempty"`
// }
