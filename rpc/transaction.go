// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// Ensure Transaction implements the TypedOperation interface.
var _ TypedOperation = (*Transaction)(nil)

// Transaction represents a transaction operation
type Transaction struct {
	Manager
	Destination tezos.Address         `json:"destination"`
	Amount      int64                 `json:"amount,string"`
	Parameters  *micheline.Parameters `json:"parameters,omitempty"`
}

// Costs returns operation cost to implement TypedOperation interface.
func (t Transaction) Costs() tezos.Costs {
	res := t.Metadata.Result
	cost := tezos.Costs{
		Fee:         t.Manager.Fee,
		GasUsed:     res.Gas(),
		StorageUsed: res.PaidStorageSizeDiff,
	}
	if !t.Result().IsSuccess() {
		return cost
	}
	var i int
	for _, v := range res.BalanceUpdates {
		if v.Kind != CONTRACT {
			continue
		}
		if t.Amount > 0 && v.AmountAbs() == t.Amount {
			continue
		}
		burn := v.Amount()
		if burn >= 0 {
			continue
		}
		if res.PaidStorageSizeDiff > 0 && i == 0 {
			cost.StorageBurn += -burn
			cost.Burn += -burn
			i++
		} else if res.Allocated {
			cost.AllocationBurn += -burn
			cost.Burn += -burn
			i++
		}
	}
	for _, in := range t.Metadata.InternalResults {
		cost = cost.Add(in.Costs())
	}
	return cost
}

type InternalResult struct {
	Kind          tezos.OpType          `json:"kind"`
	Source        tezos.Address         `json:"source"`
	Nonce         int64                 `json:"nonce"`
	Result        OperationResult       `json:"result"`
	Destination   *tezos.Address        `json:"destination,omitempty"` // transaction
	Delegate      *tezos.Address        `json:"delegate,omitempty"`    // delegation
	Parameters    *micheline.Parameters `json:"parameters,omitempty"`  // transaction
	Amount        int64                 `json:"amount,string"`         // transaction
	Balance       int64                 `json:"balance,string"`        // origination
	Script        *micheline.Script     `json:"script,omitempty"`      // origination
	Type          micheline.Prim        `json:"type"`                  // event
	Payload       micheline.Prim        `json:"payload"`               // event
	Tag           string                `json:"tag"`                   // event
	TicketUpdates []TicketUpdate        `json:"ticket_receipt"`        // v015
}

func (r InternalResult) Costs() tezos.Costs {
	cost := tezos.Costs{
		GasUsed:     r.Result.Gas(),
		StorageUsed: r.Result.PaidStorageSizeDiff,
	}
	var i int
	for _, v := range r.Result.BalanceUpdates {
		if v.Kind != CONTRACT {
			continue
		}
		if r.Amount > 0 && v.AmountAbs() == r.Amount {
			continue
		}
		burn := v.Amount()
		if burn >= 0 {
			continue
		}
		switch {
		case r.Result.PaidStorageSizeDiff > 0 && i == 0:
			cost.StorageBurn += -burn
			cost.Burn += -burn
			i++
		case len(r.Result.OriginatedContracts) > 0 && i == 1:
			cost.AllocationBurn += -burn
			cost.Burn += -burn
			i++
		case r.Result.Allocated:
			cost.AllocationBurn += -burn
			cost.Burn += -burn
			i++
		}
	}
	return cost
}

// found in block metadata from v010+
type ImplicitResult struct {
	Kind                tezos.OpType      `json:"kind"`
	BalanceUpdates      BalanceUpdates    `json:"balance_updates"`
	ConsumedGas         int64             `json:"consumed_gas,string"`
	ConsumedMilliGas    int64             `json:"consumed_milligas,string"`
	Storage             *micheline.Prim   `json:"storage,omitempty"`
	StorageSize         int64             `json:"storage_size,string"`
	OriginatedContracts []tezos.Address   `json:"originated_contracts,omitempty"`
	PaidStorageSizeDiff int64             `json:"paid_storage_size_diff,string"`
	Script              *micheline.Script `json:"script,omitempty"`
}

func (r ImplicitResult) Gas() int64 {
	if r.ConsumedMilliGas > 0 {
		return r.ConsumedMilliGas / 1000
	}
	return r.ConsumedGas
}

func (r ImplicitResult) MilliGas() int64 {
	if r.ConsumedMilliGas > 0 {
		return r.ConsumedMilliGas
	}
	return r.ConsumedGas * 1000
}
