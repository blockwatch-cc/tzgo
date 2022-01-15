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
	Metadata    OperationMetadata     `json:"metadata"`
}

// Meta returns operation metadata to implement TypedOperation interface.
func (t Transaction) Meta() OperationMetadata {
	return t.Metadata
}

// Result returns operation result to implement TypedOperation interface.
func (t Transaction) Result() OperationResult {
	return t.Metadata.Result
}

// Cost returns operation cost to implement TypedOperation interface.
func (t Transaction) Cost() OperationCost {
	res := t.Metadata.Result
	cost := OperationCost{
		Fee:          t.Manager.Fee,
		Gas:          res.ConsumedGas,
		StorageBytes: res.PaidStorageSizeDiff,
	}
	var i int
	if res.PaidStorageSizeDiff > 0 {
		burn := res.BalanceUpdates[i].Amount()
		cost.StorageBurn += -burn
		cost.Burn += -burn
		i++
	}
	if res.Allocated {
		burn := res.BalanceUpdates[i].Amount()
		cost.AllocationBurn += -burn
		cost.Burn += -burn
		i++
	}
	for _, in := range t.Metadata.InternalResults {
		cost.Gas += in.Result.ConsumedGas
		cost.StorageBytes += in.Result.PaidStorageSizeDiff
		var i int
		if in.Amount > 0 {
			i += 2
		}
		if in.Result.PaidStorageSizeDiff > 0 {
			burn := in.Result.BalanceUpdates[i].Amount()
			cost.StorageBurn += -burn
			cost.Burn += -burn
			i++
		}
		if len(in.Result.OriginatedContracts) > 0 || in.Result.Allocated {
			burn := in.Result.BalanceUpdates[i].Amount()
			cost.AllocationBurn += -burn
			cost.Burn += -burn
			i++
		}
	}
	return cost
}

type InternalResult struct {
	Kind        tezos.OpType          `json:"kind"`
	Source      tezos.Address         `json:"source"`
	Nonce       int64                 `json:"nonce"`
	Result      OperationResult       `json:"result"`
	Destination *tezos.Address        `json:"destination,omitempty"` // transaction
	Delegate    *tezos.Address        `json:"delegate,omitempty"`    // delegation
	Parameters  *micheline.Parameters `json:"parameters,omitempty"`  // transaction
	Amount      int64                 `json:"amount,string"`         // transaction
	Balance     int64                 `json:"balance,string"`        // origination
	Script      *micheline.Script     `json:"script,omitempty"`      // origination
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
