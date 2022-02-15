// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

// Represents a generic FA2 (tzip12) token
type FA2Token struct {
	Address  tezos.Address
	TokenId  tezos.Z
	contract *Contract
}

func NewFA2Token(addr tezos.Address, id int64, cli *rpc.Client) *FA2Token {
	t := &FA2Token{Address: addr, contract: NewContract(addr, cli)}
	t.TokenId.SetInt64(id)
	return t
}

func (t FA2Token) Contract() *Contract {
	return t.contract
}

func (t FA2Token) Equal(v FA2Token) bool {
	return t.Address.Equal(v.Address) && t.TokenId.Equal(v.TokenId)
}

func (t FA2Token) GetMetadata(ctx context.Context) (TokenMetadata, error) {
	// TODO
	return TokenMetadata{}, nil
}

type FA2BalanceRequest struct {
	Owner   tezos.Address `json:"owner"`
	TokenId tezos.Z       `json:"token_id"`
}

type FA2BalanceResponse struct {
	Request FA2BalanceRequest `json:"request"`
	Balance tezos.Z           `json:"balance"`
}

func (t FA2Token) GetBalances(ctx context.Context, req []FA2BalanceRequest) ([]FA2BalanceResponse, error) {
	args := micheline.NewSeq()
	for _, r := range req {
		args.Args = append(args.Args, micheline.NewPair(
			micheline.NewBytes(r.Owner.Bytes22()),
			micheline.NewNat(r.TokenId.Big()),
		))
	}
	prim, err := t.contract.RunView(ctx, "balance_of", args)
	if err != nil {
		return nil, err
	}
	val := micheline.NewValue(
		micheline.NewType(micheline.ITzip12.FuncPrim("balance_of").Args[1].Args[0]),
		prim,
	)
	resp := make([]FA2BalanceResponse, 0)
	err = val.Unmarshal(&resp)
	return resp, err
}

type FA2Approval struct {
	Owner    tezos.Address `json:"owner"`
	Operator tezos.Address `json:"operator"`
	TokenId  tezos.Z       `json:"token_id"`
	Add      bool          `json:"-"`
}

func (p *FA2Approval) UnmarshalJSON(data []byte) error {
	nested := make(map[string]json.RawMessage)
	err := json.Unmarshal(data, &nested)
	if err != nil {
		return err
	}
	v, ok := nested["add_operator"]
	if ok {
		err = json.Unmarshal(v, p)
		if err != nil {
			return err
		}
		p.Add = true
	} else {
		v, ok = nested["remove_operator"]
		if !ok {
			return fmt.Errorf("invalid FA2 approval data")
		}
		err = json.Unmarshal(v, p)
		if err != nil {
			return err
		}
	}
	return nil
}

type FA2ApprovalArgs struct {
	TxArgs
	Approvals []FA2Approval `json:"update_operators"`
}

var _ CallArguments = (*FA2ApprovalArgs)(nil)

func (p FA2ApprovalArgs) Parameters() *micheline.Parameters {
	params := &micheline.Parameters{
		Entrypoint: "update_operators",
		Value:      micheline.NewSeq(),
	}
	for _, v := range p.Approvals {
		branch := micheline.D_LEFT // add_operator
		if !v.Add {
			branch = micheline.D_RIGHT // remove_operator
		}
		params.Value.Args = append(params.Value.Args, micheline.NewCode(
			branch,
			micheline.NewPair(
				micheline.NewBytes(v.Owner.Bytes22()),
				micheline.NewPair(
					micheline.NewBytes(v.Operator.Bytes22()),
					micheline.NewNat(v.TokenId.Big()),
				),
			),
		))
	}
	return params
}

func NewFA2ApprovalArgs() *FA2ApprovalArgs {
	return &FA2ApprovalArgs{
		Approvals: make([]FA2Approval, 0),
	}
}

func (p *FA2ApprovalArgs) AddOperator(owner, operator tezos.Address, id tezos.Z) *FA2ApprovalArgs {
	if p.Approvals == nil {
		p.Approvals = make([]FA2Approval, 0)
	}
	p.Approvals = append(p.Approvals, FA2Approval{
		Owner:    owner.Clone(),
		Operator: operator.Clone(),
		TokenId:  id,
		Add:      true,
	})
	return p
}

func (p *FA2ApprovalArgs) RemoveOperator(owner, operator tezos.Address, id tezos.Z) *FA2ApprovalArgs {
	if p.Approvals == nil {
		p.Approvals = make([]FA2Approval, 0)
	}
	p.Approvals = append(p.Approvals, FA2Approval{
		Owner:    owner.Clone(),
		Operator: operator.Clone(),
		TokenId:  id.Clone(),
		Add:      true,
	})
	return p
}

func (p FA2ApprovalArgs) Encode() *codec.Transaction {
	return &codec.Transaction{
		Manager: codec.Manager{
			Source: p.Source,
		},
		Destination: p.Destination,
		Parameters:  p.Parameters(),
	}
}

type FA2Transfer struct {
	From    tezos.Address
	To      tezos.Address
	TokenId tezos.Z
	Amount  tezos.Z
}

func (t FA2Transfer) Prim() micheline.Prim {
	return micheline.NewPair(
		micheline.NewBytes(t.To.Bytes22()),
		micheline.NewPair(
			micheline.NewNat(t.TokenId.Big()),
			micheline.NewNat(t.Amount.Big()),
		),
	)
}

type FA2TransferList []FA2Transfer

func (l FA2TransferList) Len() int { return len(l) }
func (l FA2TransferList) Less(i, j int) bool {
	return bytes.Compare(l[i].From.Bytes22(), l[j].From.Bytes22()) < 0
}
func (l FA2TransferList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// compatible with micheline.Value.Unmarshal()
func (t *FA2TransferList) UnmarshalJSON(data []byte) error {
	var xfer struct {
		Transfers []struct {
			From tezos.Address `json:"from_"`
			Txs  []struct {
				To      tezos.Address `json:"to_"`
				TokenId tezos.Z       `json:"token_id"`
				Amount  tezos.Z       `json:"amount"`
			} `json:"txs"`
		} `json:"transfer"`
	}
	if err := json.Unmarshal(data, &xfer); err != nil {
		return err
	}
	if *t == nil {
		*t = make(FA2TransferList, 0, len(xfer.Transfers))
	}
	for i := range xfer.Transfers {
		for j := range xfer.Transfers[i].Txs {
			// Note: token address is unknown here
			tx := FA2Transfer{
				From:    xfer.Transfers[i].From,
				To:      xfer.Transfers[i].Txs[j].To,
				TokenId: xfer.Transfers[i].Txs[j].TokenId,
				Amount:  xfer.Transfers[i].Txs[j].Amount,
			}
			*t = append(*t, tx)
		}
	}
	return nil
}

type FA2TransferArgs struct {
	TxArgs
	Transfers FA2TransferList
}

var _ CallArguments = (*FA2TransferArgs)(nil)

func NewFA2TransferArgs() *FA2TransferArgs {
	return &FA2TransferArgs{
		Transfers: make(FA2TransferList, 0),
	}
}

func (p *FA2TransferArgs) WithTransfer(from, to tezos.Address, id, amount tezos.Z) *FA2TransferArgs {
	if p.Transfers == nil {
		p.Transfers = make(FA2TransferList, 0)
	}
	p.Transfers = append(p.Transfers, FA2Transfer{
		From:    from.Clone(),
		To:      to.Clone(),
		TokenId: id.Clone(),
		Amount:  amount.Clone(),
	})
	return p
}

func (p *FA2TransferArgs) Optimize() *FA2TransferArgs {
	// stable-sort by `from` address
	sort.Stable(p.Transfers)
	return p
}

func (t FA2TransferArgs) Parameters() *micheline.Parameters {
	// collate by `from` address
	var k int
	seq := micheline.NewSeq()
	for i, v := range t.Transfers {
		if i == 0 || !v.From.Equal(t.Transfers[i-1].From) {
			seq.Args = append(seq.Args,
				micheline.NewPair(
					micheline.NewBytes(v.From.Bytes22()),
					micheline.NewSeq(),
				),
			)
			k = len(seq.Args) - 1
		}
		seq.Args[k].Args[1].Args = append(seq.Args[k].Args[1].Args, v.Prim())
	}
	return &micheline.Parameters{
		Entrypoint: "transfer",
		Value:      seq,
	}
}

func (p FA2TransferArgs) Encode() *codec.Transaction {
	return &codec.Transaction{
		Manager: codec.Manager{
			Source: p.Source,
		},
		Destination: p.Destination,
		Parameters:  p.Parameters(),
	}
}

// TODO: make it work for internal results as well (so we can use it for crawling)
type FA2TransferReceipt struct {
	tx *rpc.Transaction
}

func NewFA2TransferReceipt(tx *rpc.Transaction) (*FA2TransferReceipt, error) {
	if tx.Parameters == nil {
		return nil, fmt.Errorf("missing transaction parameters")
	}
	if tx.Parameters.Entrypoint != "transfer" {
		return nil, fmt.Errorf("invalid transfer entrypoint name %q", tx.Parameters.Entrypoint)
	}
	return &FA2TransferReceipt{tx: tx}, nil
}

func (r FA2TransferReceipt) IsSuccess() bool {
	return r.tx.Result().Status.IsSuccess()
}

func (r FA2TransferReceipt) Request() FA2TransferList {
	typ := micheline.ITzip12.FuncType("transfer")
	val := micheline.NewValue(typ, r.tx.Parameters.Value)
	xfer := make(FA2TransferList, 0)
	// FIXME: works only for strictly compliant contracts (i.e. type + annots)
	_ = val.Unmarshal(&xfer)
	return xfer
}

func (r FA2TransferReceipt) Result() *rpc.Transaction {
	return r.tx
}

func (r FA2TransferReceipt) Costs() tezos.Costs {
	return r.tx.Costs()
}

func (r FA2TransferReceipt) BalanceUpdates() []TokenBalance {
	// TODO: read from bigmap update
	return nil
}
