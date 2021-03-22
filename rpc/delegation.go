// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// DelegationOp represents a transaction operation
type DelegationOp struct {
	GenericOp
	Source       tezos.Address         `json:"source"`
	Fee          int64                 `json:"fee,string"`
	Counter      int64                 `json:"counter,string"`
	GasLimit     int64                 `json:"gas_limit,string"`
	StorageLimit int64                 `json:"storage_limit,string"`
	Delegate     tezos.Address         `json:"delegate,omitempty"`
	Metadata     *DelegationOpMetadata `json:"metadata"`
}

// DelegationOpMetadata represents a transaction operation metadata
type DelegationOpMetadata struct {
	BalanceUpdates BalanceUpdates   `json:"balance_updates"` // fee-related
	Result         DelegationResult `json:"operation_result"`
}

// DelegationResult represents a transaction result
type DelegationResult struct {
	ConsumedGas int64            `json:"consumed_gas,string"`
	Status      tezos.OpStatus   `json:"status"`
	Errors      []OperationError `json:"errors,omitempty"`
}
