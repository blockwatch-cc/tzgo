// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// ConstantRegistrationOp represents a global constant registration operation
type ConstantRegistrationOp struct {
	GenericOp
	Source       tezos.Address                   `json:"source"`
	Fee          int64                           `json:"fee,string"`
	Counter      int64                           `json:"counter,string"`
	GasLimit     int64                           `json:"gas_limit,string"`
	StorageLimit int64                           `json:"storage_limit,string"`
	Value        micheline.Prim                  `json:"value,omitempty"`
	Metadata     *ConstantRegistrationOpMetadata `json:"metadata"`
}

// ConstantRegistrationOpMetadata represents a transaction operation metadata
type ConstantRegistrationOpMetadata struct {
	BalanceUpdates BalanceUpdates              `json:"balance_updates"` // fee-related
	Result         *ConstantRegistrationResult `json:"operation_result"`
}

// ConstantRegistrationResult represents a transaction result
type ConstantRegistrationResult struct {
	Status         tezos.OpStatus   `json:"status"`
	BalanceUpdates BalanceUpdates   `json:"balance_updates"` // sender related
	ConsumedGas    int64            `json:"consumed_gas,string"`
	Errors         []OperationError `json:"errors,omitempty"`

	StorageSize   int64          `json:"storage_size,string"`
	GlobalAddress tezos.ExprHash `json:"global_address"`
}
