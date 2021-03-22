// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// OriginationOp represents a contract creation operation
type OriginationOp struct {
	GenericOp
	Source         tezos.Address          `json:"source"`
	Fee            int64                  `json:"fee,string"`
	Counter        int64                  `json:"counter,string"`
	GasLimit       int64                  `json:"gas_limit,string"`
	StorageLimit   int64                  `json:"storage_limit,string"`
	ManagerPubkey  tezos.Address          `json:"manager_pubkey"` // proto v1 & >= v4
	ManagerPubkey2 tezos.Address          `json:"managerPubkey"`  // proto v2, v3
	Balance        int64                  `json:"balance,string"`
	Spendable      *bool                  `json:"spendable"`   // true when missing before v5 Babylon
	Delegatable    *bool                  `json:"delegatable"` // true when missing before v5 Babylon
	Delegate       *tezos.Address         `json:"delegate"`
	Script         *micheline.Script      `json:"script"`
	Metadata       *OriginationOpMetadata `json:"metadata"`
}

func (o OriginationOp) Manager() tezos.Address {
	if o.ManagerPubkey2.IsValid() {
		return o.ManagerPubkey2
	}
	return o.ManagerPubkey
}

// OriginationOpMetadata represents a transaction operation metadata
type OriginationOpMetadata struct {
	BalanceUpdates BalanceUpdates    `json:"balance_updates"` // fee-related
	Result         OriginationResult `json:"operation_result"`
}

// OriginationResult represents a contract creation result
type OriginationResult struct {
	BalanceUpdates      BalanceUpdates   `json:"balance_updates"` // burned fees
	OriginatedContracts []tezos.Address  `json:"originated_contracts"`
	ConsumedGas         int64            `json:"consumed_gas,string"`
	StorageSize         int64            `json:"storage_size,string"`
	PaidStorageSizeDiff int64            `json:"paid_storage_size_diff,string"`
	Status              tezos.OpStatus   `json:"status"`
	Errors              []OperationError `json:"errors,omitempty"`

	// v007
	ConsumedMilliGas int64 `json:"consumed_milligas,string"`

	// deprecated in v008
	BigMapDiff micheline.BigMapDiff `json:"big_map_diff,omitempty"`

	// v008
	LazyStorageDiff LazyStorageDiff `json:"lazy_storage_diff,omitempty"`
}
