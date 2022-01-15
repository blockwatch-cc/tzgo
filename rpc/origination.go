// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// Ensure Origination implements the TypedOperation interface.
var _ TypedOperation = (*Origination)(nil)

// Origination represents a contract creation operation
type Origination struct {
	Manager
	ManagerPubkey  tezos.Address     `json:"manager_pubkey"` // proto v1 & >= v4
	ManagerPubkey2 tezos.Address     `json:"managerPubkey"`  // proto v2, v3
	Balance        int64             `json:"balance,string"`
	Spendable      *bool             `json:"spendable"`   // true when missing before v5 Babylon
	Delegatable    *bool             `json:"delegatable"` // true when missing before v5 Babylon
	Delegate       *tezos.Address    `json:"delegate"`
	Script         *micheline.Script `json:"script"`
	Metadata       OperationMetadata `json:"metadata"`
}

func (o Origination) ManagerAddress() tezos.Address {
	if o.ManagerPubkey2.IsValid() {
		return o.ManagerPubkey2
	}
	return o.ManagerPubkey
}

// Meta returns an empty operation metadata to implement TypedOperation interface.
func (o Origination) Meta() OperationMetadata {
	return o.Metadata
}

// Result returns an empty operation result to implement TypedOperation interface.
func (o Origination) Result() OperationResult {
	return o.Metadata.Result
}

// Cost returns operation cost to implement TypedOperation interface.
func (o Origination) Cost() OperationCost {
	res := o.Metadata.Result
	cost := OperationCost{
		Fee:          o.Manager.Fee,
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
	if len(res.OriginatedContracts) > 0 {
		burn := res.BalanceUpdates[i].Amount()
		cost.AllocationBurn += -burn
		cost.Burn += -burn
		i++
	}
	return cost
}
