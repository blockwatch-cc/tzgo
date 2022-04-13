// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/legonian/tzgo/codec"
	"github.com/legonian/tzgo/micheline"
	"github.com/legonian/tzgo/tezos"
)

const GasSafetyMargin int64 = 100

var (
	defaultRevealLimits = tezos.Limits{
		Fee:      1000,
		GasLimit: 1000,
	}
	// for transfers to tz1/2/3
	defaultTransferLimitsEOA = tezos.Limits{
		Fee:      1000,
		GasLimit: 1420,
	}
	// for transfers to manager.tz
	defaultTransferLimitsKT1 = tezos.Limits{
		Fee:      1000,
		GasLimit: 2078,
	}
	defaultDelegationLimitsEOA = tezos.Limits{
		Fee:      1000,
		GasLimit: 1000,
	}
)

type RunOperationRequest struct {
	Operation *codec.Op         `json:"operation"`
	ChainId   tezos.ChainIdHash `json:"chain_id"`
}

type RunViewRequest struct {
	Contract   tezos.Address     `json:"contract"`
	Entrypoint string            `json:"entrypoint"`
	Input      micheline.Prim    `json:"input"`
	ChainId    tezos.ChainIdHash `json:"chain_id"`
	Source     tezos.Address     `json:"source"`
	Payer      tezos.Address     `json:"payer"`
	Gas        tezos.N           `json:"gas"`
	Mode       string            `json:"unparsing_mode"` // "Readable" | "Optimized"
}

type RunViewResponse struct {
	Data micheline.Prim `json:"data"`
}

// Complete ensures an operation is compatible with the current source account's
// on-chain state. Sets branch for TTL control, replay counters, and reveals
// the sender's pubkey if not published yet.
func (c *Client) Complete(ctx context.Context, o *codec.Op, key tezos.Key) error {
	needBranch := !o.Branch.IsValid()
	needCounter := len(o.Contents) > 0 && o.Contents[0].GetCounter() == 0
	mayNeedReveal := len(o.Contents) > 0 && o.Contents[0].Kind() != tezos.OpTypeReveal

	if !needBranch && !mayNeedReveal && !needCounter {
		return nil
	}

	// add branch for TTL control
	if needBranch {
		ofs := o.Params.MaxOperationsTTL - o.TTL
		hash, err := c.GetBlockHash(ctx, NewBlockOffset(Head, -ofs))
		if err != nil {
			return err
		}
		o.WithBranch(hash)
	}

	if needCounter || mayNeedReveal {
		// fetch current state
		state, err := c.GetContractExt(ctx, key.Address(), Head)
		if err != nil {
			return err
		}

		// add reveal if necessary
		if mayNeedReveal && !state.IsRevealed() {
			reveal := &codec.Reveal{
				Manager: codec.Manager{
					Source: key.Address(),
				},
				PublicKey: key,
			}
			reveal.WithLimits(defaultRevealLimits)
			o.WithContentsFront(reveal)
		}

		// add counters
		if needCounter {
			nextCounter := state.Counter + 1
			for _, op := range o.Contents {
				// skip non-manager ops
				if op.GetCounter() < 0 {
					continue
				}
				op.WithCounter(nextCounter)
				nextCounter++
			}
		}
	}
	return nil
}

// Simulate dry-runs the execution of the operation against the current state
// of a Tezos node in order to estimate execution costs and fees (fee/burn/gas/storage).
func (c *Client) Simulate(ctx context.Context, o *codec.Op) (*Receipt, error) {
	sim := &codec.Op{
		Branch:    o.Branch,
		Contents:  o.Contents,
		Signature: tezos.ZeroSignature,
		TTL:       o.TTL,
		Params:    o.Params,
	}

	if !sim.Branch.IsValid() {
		ofs := o.Params.MaxOperationsTTL - sim.TTL
		hash, err := c.GetBlockHash(ctx, NewBlockOffset(Head, -ofs))
		if err != nil {
			return nil, err
		}
		sim.Branch = hash
	}

	req := RunOperationRequest{
		Operation: sim,
		ChainId:   c.ChainId,
	}
	resp := &Operation{}
	if err := c.RunOperation(ctx, Head, req, resp); err != nil {
		return nil, err
	}

	res := &Receipt{
		Op: resp,
	}
	return res, nil
}

// Validate compares local serializiation against remote RPC serialization of the
// operation and returns an error on mismatch.
func (c *Client) Validate(ctx context.Context, o *codec.Op) error {
	op := &codec.Op{
		Branch:   o.Branch,
		Contents: o.Contents,
	}
	local := op.Bytes()
	var remote tezos.HexBytes
	if err := c.ForgeOperation(ctx, Head, op, &remote); err != nil {
		return err
	}
	if !bytes.Equal(local, remote.Bytes()) {
		return fmt.Errorf("tezos: mismatch between local and remote serialized operations:\n local=%s\n remote=%s",
			hex.EncodeToString(local), hex.EncodeToString(remote))
	}
	return nil
}

// Broadcast sends the signed operation to network and returns the operation hash
// on successful pre-validation.
func (c *Client) Broadcast(ctx context.Context, o *codec.Op) (tezos.OpHash, error) {
	return c.BroadcastOperation(ctx, o.Bytes())
}
