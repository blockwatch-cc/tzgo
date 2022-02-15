// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//go:build ignore
// +build ignore

package contract

import (
	"context"
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"blockwatch.cc/tzgo/wallet"
)

// WIP

// Represents a generic FA1 (tzip5), FA1.2 (tzip7) or FA2 (tzip12) token
type Token struct {
	Kind     TokenKind     `json:"kind"`
	Address  tezos.Address `json:"address"`
	TokenId  tezos.Z       `json:"token_id"`
	Metadata TokenMetadata `json:"metadata"`
}

func (t Token) Equal(v Token) bool {
	return t.Address.Equal(v.Address) && t.TokenId.Equal(v.TokenId)
}

func NewToken(addr tezos.Address, id int64, cli *rpc.Client) *Token {
	t := &Token{Address: addr, contract: NewContract(addr, cli)}
	t.TokenId.SetInt64(id)
	return t
}

func GetTokenBalance(ctx context.Context, c *rpc.Client, token Token, owner tezos.Address) (tezos.Z, error) {
	// TODO: call off-chain or fake on-chain view via proxy contract
	return tezos.Z{}, nil
}

func GetTokenSupply(ctx context.Context, c *rpc.Client, token Token) (tezos.Z, error) {
	// TODO: call off-chain or fake on-chain view via proxy contract
	return tezos.Z{}, nil
}

func GetTokenAllowance(ctx context.Context, c *rpc.Client, token Token, owner, spender tezos.Address) (tezos.Z, error) {
	// TODO: call off-chain or fake on-chain view via proxy contract
	return tezos.Z{}, nil
}

func ApproveToken(ctx context.Context, c *rpc.Client, req TokenApprovalParams) error {
	// TODO: call off-chain or fake on-chain view via proxy contract
	return nil
}

type TokenApproval struct {
	Token   Token         `json:"token"`
	Spender tezos.Address `json:"spender"`
	Value   tezos.Z       `json:"value"`
}

type TokenApprovalParams struct {
	CallParams
	Approval TokenApproval `json:"approve"`
}

func (a TokenApproval) Parameters() *micheline.Parameters {
	return &micheline.Parameters{
		Entrypoint: "approve",
		Value: micheline.NewPairValue(
			micheline.NewBytes(a.Spender.Bytes22()),
			micheline.NewNat(a.Value.Big()),
		),
	}
}

type TokenTransfer struct {
	Token  Token         `json:"token"`
	From   tezos.Address `json:"from"`
	To     tezos.Address `json:"to"`
	Amount tezos.Z       `json:"value"`
}

func (t TokenTransfer) Parameters() *micheline.Parameters {
	switch t.Token.Kind {
	case TokenKindFA1, TokenKindFA1_2:
		return &micheline.Parameters{
			Entrypoint: "transfer",
			Value: micheline.NewPairValue(
				micheline.NewBytes(t.From.Bytes22()),
				micheline.NewPairValue(
					micheline.NewBytes(t.To.Bytes22()),
					micheline.NewNat(t.Amount.Big()),
				),
			),
		}
	case TokenKindFA2:
		return &micheline.Parameters{
			Entrypoint: "transfer",
			Value: micheline.NewSeq(
				micheline.NewPairValue(
					micheline.NewBytes(t.From.Bytes22()),
					micheline.NewSeq(
						micheline.NewPairValue(
							micheline.NewBytes(t.To.Bytes22()),
							micheline.NewPairValue(
								micheline.NewNat(t.Token.TokenId.Big()),
								micheline.NewNat(t.Amount.Big()),
							),
						),
					),
				),
			),
		}
	default:
		return nil
	}
}

type fa1Transfer struct {
	From   tezos.Address `json:"from"`
	To     tezos.Address `json:"to"`
	Amount tezos.Z       `json:"value"`
}

type fa2Transfer []struct {
	From tezos.Address `json:"from_"`
	Txs  []struct {
		To      tezos.Address `json:"to_"`
		TokenId tezos.Z       `json:"token_id"`
		Amount  tezos.Z       `json:"amount"`
	} `json:"txs"`
}

// compatible with micheline.Value.Unmarshal()
func (t *TokenTransfer) UnmarshalJSON(data []byte) error {
	switch t.Token.Kind {
	case TokenKindFA1:
		var xfer struct {
			Transfer fa1Transfer `json:"transfer"`
		}
		if err := json.Unmarshal(data, &xfer); err != nil {
			return err
		}
		t.From = xfer.Transfer.From
		t.To = xfer.Transfer.To
		t.Amount = xfer.Transfer.Amount
	case TokenKindFA1_2:
		var xfer struct {
			Transfer fa1Transfer `json:"transfer"`
		}
		if err := json.Unmarshal(data, &xfer); err != nil {
			return err
		}
		t.From = xfer.Transfer.From
		t.To = xfer.Transfer.To
		t.Amount = xfer.Transfer.Amount
	case TokenKindFA2:
		var xfer struct {
			Transfer fa2Transfer `json:"transfer"`
		}
		if err := json.Unmarshal(data, &xfer); err != nil {
			return err
		}
		t.From = xfer.Transfer[0].From
		t.To = xfer.Transfer[0].Txs[0].To
		t.Amount = xfer.Transfer[0].Txs[0].Amount
	}
	return nil
}

// TokenTransferParams is used as input to token.Transfer.
type TokenTransferParams struct {
	CallParams
	Transfer TokenTransfer
}

// TODO: API design
//
// a receipt can originate from
// - a direct FA1.2 transfer (single transaction in rpc.OperationList)
// - a direct FA1.2 transfer + reveal (second transaction in rpc.OperationList)
// - a batched FA1.2 transfer (pos-N transaction in rpc.OperationList)
// - an internal FA1.2 transfer (pos-M transaction in rpc.OperationList[n].InternalResults)
//
// should Receipt contain a proof or link to op hash / block hash?
//
// Howto send a transfer/approval?
// A: this package only contains types
// B: this package provides funcs to construct complete transactions

type TokenTransferReceipt struct {
	kind    TokenKind
	receipt *wallet.Receipt
	tx      *rpc.Transaction
}

func NewTokenTransferReceipt(k TokenKind, r *wallet.Receipt) (*TokenTransferReceipt, error) {
	if r.Op == nil {
		return nil, fmt.Errorf("invalid receipt")
	}
	op := r.Op.Contents.Select(tezos.OpTypeTransaction, 0)
	if op == nil {
		return nil, fmt.Errorf("missing transaction")
	}
	tx := op.(*rpc.Transaction)
	if tx.Parameters == nil {
		return nil, fmt.Errorf("missing transaction parameters")
	}
	return &TokenTransferReceipt{kind: k, receipt: r, tx: tx}, nil
}

func (r TokenTransferReceipt) IsSuccess() bool {
	return r.tx.Result().Status.IsSuccess()
}

func (r TokenTransferReceipt) Request() TokenTransfer {
	var typ micheline.Type
	switch r.kind {
	case TokenKindFA1:
		typ = micheline.NewType(micheline.WellKnownInterfaces[micheline.ITzip5][0])
	case TokenKindFA1_2:
		typ = micheline.NewType(micheline.WellKnownInterfaces[micheline.ITzip7][0])
	case TokenKindFA2:
		typ = micheline.NewType(micheline.WellKnownInterfaces[micheline.ITzip12][0])
	}
	val := micheline.NewValue(typ, r.tx.Parameters.Value)
	xfer := TokenTransfer{Token: Token{Kind: r.kind, Address: r.tx.Destination}}
	_ = val.Unmarshal(&xfer)
	return xfer
}

// TODO: what's useful to the caller?
// - pre/post balances of from/to accounts (extracted from ledger bigmap updates)
// - actual balance updates (i.e. another TokenTransfer struct from ledger updates)
// - raw list of bigmap updates
//  update requires
func (r TokenTransferReceipt) Result() *rpc.Transaction {
	return r.tx
}

func (r TokenTransferReceipt) Costs() tezos.Costs {
	return r.tx.Costs()
}
