// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"
)

const GasSafetyMargin int64 = 100

var (
	// for reveal
	DefaultRevealLimits = tezos.Limits{
		Fee:      1000,
		GasLimit: 1000,
	}
	// for transfers to tz1/2/3
	DefaultTransferLimitsEOA = tezos.Limits{
		Fee:      1000,
		GasLimit: 1420, // 1820 when source is emptied
	}
	// for transfers to manager.tz
	DefaultTransferLimitsKT1 = tezos.Limits{
		Fee:      1000,
		GasLimit: 2078,
	}
	// for delegation
	DefaultDelegationLimitsEOA = tezos.Limits{
		Fee:      1000,
		GasLimit: 1000,
	}
	// for simulating contract calls and other operations
	// used when no explicit costs are set
	DefaultSimulationLimits = tezos.Limits{
		GasLimit:     tezos.DefaultParams.HardGasLimitPerOperation,
		StorageLimit: tezos.DefaultParams.HardStorageLimitPerOperation,
	}
)

type CallOptions struct {
	Confirmations     int64         // number of confirmations to wait after broadcast
	MaxFee            int64         // max acceptable fee, optional (default = 0)
	TTL               int64         // max lifetime for operations in blocks
	IgnoreLimits      bool          // ignore simulated limits and use user-defined limits from op
	SimulationBlockID BlockID       // custom block id to simulate operation
	Signer            signer.Signer // optional signer interface to use for signing the transaction
	Sender            tezos.Address // optional address to sign for (use when signer manages multiple addresses)
	Observer          *Observer     // optional custom block observer for waiting on confirmations
}

var DefaultOptions = CallOptions{
	Confirmations: 2,
	TTL:           tezos.DefaultParams.MaxOperationsTTL - 2,
	MaxFee:        1_000_000,
}

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

type RunCodeRequest struct {
	ChainId    tezos.ChainIdHash `json:"chain_id"`
	Script     micheline.Code    `json:"script"`
	Storage    micheline.Prim    `json:"storage"`
	Input      micheline.Prim    `json:"input"`
	Amount     tezos.N           `json:"amount"`
	Balance    tezos.N           `json:"balance"`
	Source     *tezos.Address    `json:"source,omitempty"`
	Payer      *tezos.Address    `json:"payer,omitempty"`
	Gas        *tezos.N          `json:"gas,omitempty"`
	Entrypoint string            `json:"entrypoint,omitempty"`
}

// RunCodeResponse -
type RunCodeResponse struct {
	Operations      []Operation            `json:"operations"`
	Storage         micheline.Prim         `json:"storage"`
	BigmapDiff      micheline.BigmapEvents `json:"big_map_diff,omitempty"`
	LazyStorageDiff micheline.LazyEvents   `json:"lazy_storage_diff,omitempty"`
}

// Complete ensures an operation is compatible with the current source account's
// on-chain state. Sets branch for TTL control, replay counters, and reveals
// the sender's pubkey if not published yet.
func (c *Client) Complete(ctx context.Context, o *codec.Op, key tezos.Key) error {
	needBranch := !o.Branch.IsValid()
	needCounter := o.NeedCounter()
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
			reveal.WithLimits(DefaultRevealLimits)
			o.WithContentsFront(reveal)
			needCounter = true
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
func (c *Client) Simulate(ctx context.Context, o *codec.Op, opts *CallOptions) (*Receipt, error) {
	sim := &codec.Op{
		Branch:    o.Branch,
		Contents:  o.Contents,
		Signature: tezos.ZeroSignature,
		TTL:       o.TTL,
		Params:    o.Params,
	}

	if sim.TTL == 0 && opts != nil {
		sim.TTL = opts.TTL
	}

	if !sim.Branch.IsValid() {
		ofs := o.Params.MaxOperationsTTL - sim.TTL
		hash, err := c.GetBlockHash(ctx, NewBlockOffset(Head, -ofs))
		if err != nil {
			return nil, err
		}
		sim.Branch = hash
	}

	if opts == nil || !opts.IgnoreLimits {
		// use default gas/storage limits, set min fee
		for _, op := range o.Contents {
			l := op.Limits()
			if l.GasLimit == 0 {
				l.GasLimit = DefaultSimulationLimits.GasLimit / int64(len(o.Contents))
			}
			if l.StorageLimit == 0 {
				l.StorageLimit = DefaultSimulationLimits.StorageLimit / int64(len(o.Contents))
			}
			op.WithLimits(l)
		}
	}

	blockID := BlockID(Head)
	if opts != nil && opts.SimulationBlockID != nil {
		blockID = opts.SimulationBlockID
	}

	req := RunOperationRequest{
		Operation: sim,
		ChainId:   c.ChainId,
	}
	resp := &Operation{}
	if err := c.RunOperation(ctx, blockID, req, resp); err != nil {
		return nil, err
	}

	// TODO: adjust min fee using known gas units before return so that res.Cost()
	// reflects the entire cost that Send() will pay
	rcpt := &Receipt{
		Op: resp,
	}

	// fail with Tezos error when simulation failed
	if !rcpt.IsSuccess() {
		return rcpt, rcpt.Error()
	}
	return rcpt, nil
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

// Send is a convenience wrapper for sending operations. It auto-completes gas and storage limit,
// ensures minimum fees are set, protects against fee overpayment, signs and broadcasts the final
// operation and waits for a defined number of confirmations.
func (c *Client) Send(ctx context.Context, op *codec.Op, opts *CallOptions) (*Receipt, error) {
	if opts == nil {
		opts = &DefaultOptions
	}

	signer := c.Signer
	if opts.Signer != nil {
		signer = opts.Signer
	}

	// identify the sender address for signing the message
	addr := opts.Sender
	if !addr.IsValid() {
		addrs, err := signer.ListAddresses(ctx)
		if err != nil {
			return nil, err
		}
		addr = addrs[0]
	}

	key, err := signer.GetKey(ctx, addr)
	if err != nil {
		return nil, err
	}

	// set source on all ops
	op.WithSource(key.Address())

	// auto-complete op with branch/ttl, source counter, reveal
	err = c.Complete(ctx, op, key)
	if err != nil {
		return nil, err
	}

	// simulate to check tx validity and estimate cost
	sim, err := c.Simulate(ctx, op, opts)
	if err != nil {
		return nil, err
	}

	// fail with Tezos error when simulation failed
	if !sim.IsSuccess() {
		return nil, sim.Error()
	}

	// apply simulated cost as limits to tx list
	if !opts.IgnoreLimits {
		op.WithLimits(sim.MinLimits(), GasSafetyMargin)
	}

	// log info about tx costs
	logDebug(func() {
		costs := sim.Costs()
		for i, v := range op.Contents {
			verb := "used"
			if opts.IgnoreLimits {
				verb = "forced"
			}
			limits := v.Limits()
			log.Debugf("OP#%03d: %s gas_used(sim)=%d storage_used(sim)=%d storage_burn(sim)=%d alloc_burn(sim)=%d fee(%s)=%d gas_limit(%s)=%d storage_limit(%s)=%d ",
				i, v.Kind(), costs[i].GasUsed, costs[i].StorageUsed, costs[i].StorageBurn, costs[i].AllocationBurn,
				verb, limits.Fee, verb, limits.GasLimit, verb, limits.StorageLimit,
			)
		}
	})

	// check minFee calc against maxFee if set
	if opts.MaxFee > 0 {
		if l := op.Limits(); l.Fee > opts.MaxFee {
			return nil, fmt.Errorf("estimated cost %d > max %d", l.Fee, opts.MaxFee)
		}
	}

	// sign digest
	sig, err := signer.SignOperation(ctx, addr, op)
	if err != nil {
		return nil, err
	}
	op.WithSignature(sig)

	// broadcast
	hash, err := c.Broadcast(ctx, op)
	if err != nil {
		return nil, err
	}

	// wait for confirmations
	res := NewResult(hash).WithTTL(op.TTL).WithConfirmations(opts.Confirmations)

	// use custom observer when provided
	mon := c.BlockObserver
	if opts.Observer != nil {
		mon = opts.Observer
	}

	// wait for confirmations
	res.Listen(mon)
	res.WaitContext(ctx)
	if err := res.Err(); err != nil {
		return nil, err
	}

	// return receipt
	return res.GetReceipt(ctx)
}

// RunCode simulates executing of provided code on the context of a contract at selected block.
func (c *Client) RunCode(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/scripts/run_code", id)
	return c.Post(ctx, u, body, resp)
}

// RunView simulates executing of on on-chain view on the context of a contract at selected block.
func (c *Client) RunView(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/scripts/run_view", id)
	return c.Post(ctx, u, body, resp)
}

// TraceCode simulates executing of code on the context of a contract at selected block and
// returns a full execution trace.
func (c *Client) TraceCode(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/scripts/trace_code", id)
	return c.Post(ctx, u, body, resp)
}
