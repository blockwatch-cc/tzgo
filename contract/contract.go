// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

import (
	"context"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

type CallArguments interface {
	WithSource(tezos.Address)
	WithDestination(tezos.Address)
	WithAmount(tezos.N)
	Encode() *codec.Transaction
	Parameters() *micheline.Parameters
}

type TxArgs struct {
	Source      tezos.Address
	Destination tezos.Address
	Amount      tezos.N
	Params      micheline.Parameters
}

func NewTxArgs() *TxArgs {
	return &TxArgs{}
}

func (a *TxArgs) WithSource(addr tezos.Address) {
	a.Source = addr.Clone()
}

func (a *TxArgs) WithDestination(addr tezos.Address) {
	a.Destination = addr.Clone()
}

func (a *TxArgs) WithAmount(amount tezos.N) {
	a.Amount = amount
}

func (a *TxArgs) WithParameters(params micheline.Parameters) {
	a.Params = params
}

func (a *TxArgs) Parameters() *micheline.Parameters {
	return &a.Params
}

func (a *TxArgs) Encode() *codec.Transaction {
	return &codec.Transaction{
		Manager: codec.Manager{
			Source: a.Source,
		},
		Destination: a.Destination,
		Parameters:  &a.Params,
	}
}

type Contract struct {
	addr   tezos.Address     // contract address
	script *micheline.Script // script (type info + code)
	store  *micheline.Prim   // current storage value
	meta   *Tz16             // Tzip16 metadata
	rpc    *rpc.Client       // the RPC client to use for queries and calls
}

func NewContract(addr tezos.Address, cli *rpc.Client) *Contract {
	return &Contract{
		addr: addr,
		rpc:  cli,
	}
}

func NewEmptyContract(cli *rpc.Client) *Contract {
	return &Contract{
		rpc: cli,
	}
}

func (c *Contract) Resolve(ctx context.Context) error {
	// use normalized script to have the node embed global constants
	script, err := c.rpc.GetNormalizedScript(ctx, c.addr, rpc.UnparsingModeOptimized)
	if err != nil {
		return err
	}
	store, err := c.rpc.GetContractStorage(ctx, c.addr, rpc.Head)
	if err != nil {
		return err
	}
	c.script = script
	c.store = &store
	return nil
}

func (c *Contract) ResolveMetadata(ctx context.Context) (*Tz16, error) {
	// TODO
	return nil, nil
}

func (c *Contract) WithScript(script *micheline.Script) *Contract {
	c.script = script
	return c
}

func (c *Contract) DecodeScript(data []byte) error {
	var script micheline.Script
	if err := script.UnmarshalBinary(data); err != nil {
		return err
	}
	c.script = &script
	return nil
}

func (c *Contract) WithStorage(store *micheline.Prim) *Contract {
	c.store = store
	return c
}

func (c *Contract) DecodeStorage(data []byte) error {
	var store micheline.Prim
	if err := store.UnmarshalBinary(data); err != nil {
		return err
	}
	c.store = &store
	return nil
}

func (c Contract) Address() tezos.Address {
	return c.addr
}

func (c Contract) IsManagerTz() bool {
	return c.script != nil && c.script.Implements(micheline.IManager)
}

func (c Contract) IsFA1() bool {
	return c.script != nil && c.script.Implements(micheline.ITzip5)
}

func (c Contract) IsFA12() bool {
	return c.script != nil && c.script.Implements(micheline.ITzip7)
}

func (c Contract) IsFA2() bool {
	return c.script != nil && c.script.Implements(micheline.ITzip12)
}

// func (c *Contract) IsNFT() bool {}

func (c *Contract) AsFA1() *FA1Token {
	return &FA1Token{
		Address:  c.addr,
		contract: c,
	}
}

func (c *Contract) AsFA2(id int64) *FA2Token {
	fa2 := &FA2Token{
		Address:  c.addr,
		contract: c,
	}
	fa2.TokenId.SetInt64(id)
	return fa2
}

// func (c *Contract) AsNFT() (*NFTToken, error) {}

func (c *Contract) Metadata() *Tz16 {
	return c.meta
}

func (c Contract) Script() *micheline.Script {
	return c.script
}

func (c Contract) Storage() *micheline.Prim {
	return c.store
}

// entrypoints and callbacks
func (c *Contract) Entrypoint(name string) (micheline.Entrypoint, bool) {
	if c.script == nil {
		return micheline.Entrypoint{}, false
	}
	eps, _ := c.script.Entrypoints(true)
	ep, ok := eps[name]
	return ep, ok
}

// on-chain views
func (c *Contract) View(name string) (micheline.View, bool) {
	if c.script == nil {
		return micheline.View{}, false
	}
	views, _ := c.script.Views(false, false)
	view, ok := views[name]
	return view, ok
}

// func (c *Contract) GetStorageValue(path string) (*micheline.Value, error) {}

// func (c *Contract) GetBigmapValue(path string, key micheline.Key) (*micheline.Value, error) {}

// Executes TZIP-4 fake views from callback entrypoints
func (c *Contract) RunView(ctx context.Context, name string, args micheline.Prim) (micheline.Prim, error) {
	req := rpc.RunViewRequest{
		Contract:   c.addr,
		Entrypoint: name,
		Input:      args,
		ChainId:    c.rpc.ChainId,
		Source:     tezos.ZeroAddress,
		Payer:      tezos.ZeroAddress,
		Gas:        tezos.N(1_000_000), // guess
		Mode:       "Readable",
	}
	var res rpc.RunViewResponse
	err := c.rpc.RunView(ctx, rpc.Head, &req, &res)
	return res.Data, err
}

func (c *Contract) RunViewExt(ctx context.Context, name string, args micheline.Prim, source, payer tezos.Address, gas int64) (micheline.Prim, error) {
	req := rpc.RunViewRequest{
		Contract:   c.addr,
		Entrypoint: name,
		Input:      args,
		ChainId:    c.rpc.ChainId,
		Source:     source,
		Payer:      payer,
		Gas:        tezos.N(gas),
		Mode:       "Readable",
	}
	var res rpc.RunViewResponse
	err := c.rpc.RunView(ctx, rpc.Head, &req, &res)
	return res.Data, err
}

func (c *Contract) Call(ctx context.Context, args CallArguments, opts *rpc.CallOptions) (*rpc.Receipt, error) {
	return c.CallMulti(ctx, []CallArguments{args}, opts)
}

func (c *Contract) CallMulti(ctx context.Context, args []CallArguments, opts *rpc.CallOptions) (*rpc.Receipt, error) {
	if opts == nil {
		opts = &rpc.DefaultOptions
	}

	// assemble batch transaction
	op := codec.NewOp().WithTTL(opts.TTL)
	for _, arg := range args {
		arg.WithDestination(c.addr)
		op.WithContents(arg.Encode())
	}

	// prepare, sign and broadcast
	return c.rpc.Send(ctx, op, opts)
}

func (c *Contract) Deploy(ctx context.Context, opts *rpc.CallOptions) (*rpc.Receipt, error) {
	return c.DeployExt(ctx, tezos.ZeroAddress, 0, opts)
}

func (c *Contract) DeployExt(ctx context.Context, delegate tezos.Address, balance tezos.N, opts *rpc.CallOptions) (*rpc.Receipt, error) {
	if opts == nil {
		opts = &rpc.DefaultOptions
	}

	// assemble origination op
	orig := &codec.Origination{
		Script: *c.script,
	}
	if delegate.IsValid() {
		orig.Delegate = delegate
	}
	if !balance.IsZero() {
		orig.Balance = balance
	}
	op := codec.NewOp().WithTTL(opts.TTL).WithContents(orig)

	// prepare, sign and broadcast
	rcpt, err := c.rpc.Send(ctx, op, opts)
	if err != nil {
		return nil, err
	}

	// set contract address from deployment result if successful
	if !rcpt.IsSuccess() {
		return nil, rcpt.Error()
	}
	c.addr = rcpt.Op.Contents[0].Result().OriginatedContracts[0]
	return rcpt, nil
}
