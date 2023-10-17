// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

import (
	"context"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

type CallArguments interface {
	WithSource(tezos.Address) CallArguments
	WithDestination(tezos.Address) CallArguments
	WithAmount(tezos.N) CallArguments
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

func (a *TxArgs) WithSource(addr tezos.Address) CallArguments {
	a.Source = addr.Clone()
	return a
}

func (a *TxArgs) WithDestination(addr tezos.Address) CallArguments {
	a.Destination = addr.Clone()
	return a
}

func (a *TxArgs) WithAmount(amount tezos.N) CallArguments {
	a.Amount = amount
	return a
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
		Amount:      a.Amount,
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

func (c *Contract) Client() *rpc.Client {
	return c.rpc
}

func (c *Contract) Resolve(ctx context.Context) error {
	// use normalized script to have the node embed global constants
	script, err := c.rpc.GetNormalizedScript(ctx, c.addr, rpc.UnparsingModeReadable)
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

func (c *Contract) Reload(ctx context.Context) error {
	store, err := c.rpc.GetContractStorage(ctx, c.addr, rpc.Head)
	if err != nil {
		return err
	}
	c.store = &store
	return nil
}

func (c *Contract) ResolveMetadata(ctx context.Context) (*Tz16, error) {
	if c.meta != nil {
		return c.meta, nil
	}
	if c.script == nil {
		if err := c.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	tz16 := &Tz16{}
	if err := c.resolveStorageUri(ctx, "tezos-storage:", tz16, nil); err != nil {
		return nil, err
	}
	c.meta = tz16
	return tz16, nil
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

func (c Contract) IsToken() bool {
	return c.IsFA1() || c.IsFA12() || c.IsFA2()
}

func (c Contract) TokenKind() TokenKind {
	switch {
	case c.IsFA1():
		return TokenKindFA1
	case c.IsFA12():
		return TokenKindFA1_2
	case c.IsFA2():
		return TokenKindFA2
	default:
		return TokenKindInvalid
	}
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

func (c Contract) StorageValue() micheline.Value {
	return micheline.NewValue(c.script.StorageType(), *c.store)
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

func (c *Contract) GetBigmapValue(ctx context.Context, path string, args micheline.Prim) (*micheline.Value, error) {
	store := c.StorageValue()
	bigmap, ok := store.GetInt64(path)
	if !ok {
		return nil, fmt.Errorf("bigmap %q not found in storage", path)
	}
	typ, ok := c.script.Code.Storage.FindLabel(path)
	if !ok {
		return nil, fmt.Errorf("bigmap %q not found in type", path)
	}
	key, err := micheline.NewKey(micheline.NewType(typ.Args[0]), args)
	if err != nil {
		return nil, err
	}
	prim, err := c.rpc.GetActiveBigmapValue(ctx, bigmap, key.Hash())
	if err != nil {
		return nil, err
	}
	// fmt.Printf("Bigmap value %s typ %s\n", prim.Dump(), typ.Args[1].Dump())
	val := micheline.NewValue(micheline.NewType(typ.Args[1]), prim)
	return &val, nil
}

// Executes on-chain views from callback entrypoints
func (c *Contract) RunView(ctx context.Context, name string, args micheline.Prim) (micheline.Prim, error) {
	req := rpc.RunViewRequest{
		Contract:     c.addr,
		View:         name,
		Input:        args,
		ChainId:      c.rpc.ChainId,
		Source:       tezos.ZeroAddress,
		Payer:        tezos.ZeroAddress,
		UnlimitedGas: true,
		Mode:         "Readable",
	}
	var res rpc.RunViewResponse
	err := c.rpc.RunView(ctx, rpc.Head, &req, &res)
	return res.Data, err
}

func (c *Contract) RunViewExt(ctx context.Context, name string, args micheline.Prim, source, payer tezos.Address, gas int64) (micheline.Prim, error) {
	req := rpc.RunViewRequest{
		Contract: c.addr,
		View:     name,
		Input:    args,
		ChainId:  c.rpc.ChainId,
		Source:   source,
		Payer:    payer,
		Gas:      tezos.N(gas),
		Mode:     "Readable",
	}
	if gas == 0 {
		req.UnlimitedGas = true
	}
	var res rpc.RunViewResponse
	err := c.rpc.RunView(ctx, rpc.Head, &req, &res)
	return res.Data, err
}

// Executes TZIP-4 callback-based views from callback entrypoints
func (c *Contract) RunCallback(ctx context.Context, name string, args micheline.Prim) (micheline.Prim, error) {
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
	err := c.rpc.RunCallback(ctx, rpc.Head, &req, &res)
	return res.Data, err
}

func (c *Contract) RunCallbackExt(ctx context.Context, name string, args micheline.Prim, source, payer tezos.Address, gas int64) (micheline.Prim, error) {
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
	err := c.rpc.RunCallback(ctx, rpc.Head, &req, &res)
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
		if arg == nil {
			continue
		}
		arg.WithDestination(c.addr)
		op.WithContents(arg.Encode())
	}

	// prepare, sign and broadcast
	return c.rpc.Send(ctx, op, opts)
}

func (c *Contract) Deploy(ctx context.Context, opts *rpc.CallOptions) (*rpc.Receipt, error) {
	return c.DeployExt(ctx, tezos.Address{}, 0, opts)
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
	c.addr, _ = rcpt.OriginatedContract()
	return rcpt, nil
}
