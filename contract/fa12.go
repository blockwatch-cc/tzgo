// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

import (
	"context"
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

// Represents a generic FA1 (tzip5) or FA1.2 (tzip7) token
type FA1Token struct {
	Address  tezos.Address
	contract *Contract
}

func NewFA1Token(addr tezos.Address, cli *rpc.Client) *FA1Token {
	return &FA1Token{Address: addr, contract: NewContract(addr, cli)}
}

func (t FA1Token) Contract() *Contract {
	return t.contract
}

func (t FA1Token) Equal(v FA1Token) bool {
	return t.Address.Equal(v.Address)
}

func (t FA1Token) GetMetadata(ctx context.Context) (TokenMetadata, error) {
	// TODO
	return TokenMetadata{}, nil
}

func (t FA1Token) GetBalance(ctx context.Context, owner tezos.Address) (tezos.Z, error) {
	var balance tezos.Z
	prim, err := t.contract.RunView(ctx, "getBalance", micheline.NewBytes(owner.Bytes22()))
	if err == nil {
		balance.Set(prim.Int)
	}
	return balance, err
}

func (t FA1Token) GetTotalSupply(ctx context.Context) (tezos.Z, error) {
	var supply tezos.Z
	prim, err := t.contract.RunView(ctx, "getTotalSupply", micheline.NewPrim(micheline.D_UNIT))
	if err == nil {
		supply.Set(prim.Int)
	}
	return supply, err
}

func (t FA1Token) GetAllowance(ctx context.Context, owner, spender tezos.Address) (tezos.Z, error) {
	var allowance tezos.Z
	prim, err := t.contract.RunView(ctx, "getAllowance",
		micheline.NewPair(
			micheline.NewBytes(owner.Bytes22()),
			micheline.NewBytes(spender.Bytes22()),
		),
	)
	if err == nil {
		allowance.Set(prim.Int)
	}
	return allowance, err
}

type FA1Approval struct {
	Spender tezos.Address `json:"spender"`
	Value   tezos.Z       `json:"value"`
}

type FA1ApprovalArgs struct {
	TxArgs
	Approval FA1Approval `json:"approve"`
}

var _ CallArguments = (*FA1ApprovalArgs)(nil)

func (a FA1ApprovalArgs) Parameters() *micheline.Parameters {
	return &micheline.Parameters{
		Entrypoint: "approve",
		Value: micheline.NewPair(
			micheline.NewBytes(a.Approval.Spender.Bytes22()),
			micheline.NewNat(a.Approval.Value.Big()),
		),
	}
}

func (p FA1ApprovalArgs) Encode() *codec.Transaction {
	return &codec.Transaction{
		Manager: codec.Manager{
			Source: p.Source,
		},
		Destination: p.Destination,
		Parameters:  p.Parameters(),
	}
}

type FA1Transfer struct {
	From   tezos.Address `json:"from"`
	To     tezos.Address `json:"to"`
	Amount tezos.Z       `json:"value"`
}

// compatible with micheline.Value.Unmarshal()
func (t *FA1Transfer) UnmarshalJSON(data []byte) error {
	var xfer struct {
		Transfer struct {
			From   tezos.Address `json:"from"`
			To     tezos.Address `json:"to"`
			Amount tezos.Z       `json:"value"`
		} `json:"transfer"`
	}
	if err := json.Unmarshal(data, &xfer); err != nil {
		return err
	}
	t.From = xfer.Transfer.From
	t.To = xfer.Transfer.To
	t.Amount = xfer.Transfer.Amount
	return nil
}

type FA1TransferArgs struct {
	TxArgs
	Transfer FA1Transfer
}

var _ CallArguments = (*FA1TransferArgs)(nil)

func (t FA1TransferArgs) Parameters() *micheline.Parameters {
	return &micheline.Parameters{
		Entrypoint: "transfer",
		Value: micheline.NewPair(
			micheline.NewBytes(t.Transfer.From.Bytes22()),
			micheline.NewPair(
				micheline.NewBytes(t.Transfer.To.Bytes22()),
				micheline.NewNat(t.Transfer.Amount.Big()),
			),
		),
	}
}

func (p FA1TransferArgs) Encode() *codec.Transaction {
	return &codec.Transaction{
		Manager: codec.Manager{
			Source: p.Source,
		},
		Destination: p.Destination,
		Parameters:  p.Parameters(),
	}
}

type FA1TransferReceipt struct {
	tx *rpc.Transaction
}

func NewFA1TransferReceipt(tx *rpc.Transaction) (*FA1TransferReceipt, error) {
	if tx.Parameters == nil {
		return nil, fmt.Errorf("missing transaction parameters")
	}
	if tx.Parameters.Entrypoint != "transfer" {
		return nil, fmt.Errorf("invalid transfer entrypoint name %q", tx.Parameters.Entrypoint)
	}
	return &FA1TransferReceipt{tx: tx}, nil
}

func (r FA1TransferReceipt) IsSuccess() bool {
	return r.tx.Result().Status.IsSuccess()
}

func (r FA1TransferReceipt) Request() FA1Transfer {
	typ := micheline.ITzip7.FuncType("transfer")
	val := micheline.NewValue(typ, r.tx.Parameters.Value)
	xfer := FA1Transfer{}
	_ = val.Unmarshal(&xfer)
	return xfer
}

func (r FA1TransferReceipt) Result() *rpc.Transaction {
	return r.tx
}

func (r FA1TransferReceipt) Costs() tezos.Costs {
	return r.tx.Costs()
}

func (r FA1TransferReceipt) BalanceUpdates() []TokenBalance {
	// TODO: read from ledger bigmap update
	return nil
}
