// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// TransactionOp represents a transaction operation
type TransactionOp struct {
	GenericOp
	Source       tezos.Address          `json:"source"`
	Destination  tezos.Address          `json:"destination"`
	Fee          int64                  `json:"fee,string"`
	Amount       int64                  `json:"amount,string"`
	Counter      int64                  `json:"counter,string"`
	GasLimit     int64                  `json:"gas_limit,string"`
	StorageLimit int64                  `json:"storage_limit,string"`
	Parameters   *micheline.Parameters  `json:"parameters,omitempty"`
	Metadata     *TransactionOpMetadata `json:"metadata"`
}

// TransactionOpMetadata represents a transaction operation metadata
type TransactionOpMetadata struct {
	BalanceUpdates  BalanceUpdates     `json:"balance_updates"` // fee-related
	Result          *TransactionResult `json:"operation_result"`
	InternalResults []*InternalResult  `json:"internal_operation_results,omitempty"`
}

// TransactionResult represents a transaction result
type TransactionResult struct {
	BalanceUpdates      BalanceUpdates   `json:"balance_updates"` // tx or contract related
	ConsumedGas         int64            `json:"consumed_gas,string"`
	ConsumedMilliGas    int64            `json:"consumed_milligas,string"`
	Status              tezos.OpStatus   `json:"status"`
	Allocated           bool             `json:"allocated_destination_contract"` // new addr created and payed
	Errors              []OperationError `json:"errors,omitempty"`
	Storage             *micheline.Prim  `json:"storage,omitempty"`
	StorageSize         int64            `json:"storage_size,string"`
	PaidStorageSizeDiff int64            `json:"paid_storage_size_diff,string"`

	// deprecated in v008
	BigMapDiff micheline.BigMapDiff `json:"big_map_diff,omitempty"`

	// v008
	LazyStorageDiff LazyStorageDiff `json:"lazy_storage_diff,omitempty"`

	// when reused as internal origination result
	OriginatedContracts []tezos.Address `json:"originated_contracts,omitempty"`
}

type InternalResult struct {
	GenericOp
	Source      tezos.Address         `json:"source"`
	Nonce       int64                 `json:"nonce"`
	Result      *TransactionResult    `json:"result"`
	Destination *tezos.Address        `json:"destination,omitempty"` // transaction
	Delegate    *tezos.Address        `json:"delegate,omitempty"`    // delegation
	Parameters  *micheline.Parameters `json:"parameters,omitempty"`  // transaction
	Amount      int64                 `json:"amount,string"`         // transaction
	Balance     int64                 `json:"balance,string"`        // origination
	Script      *micheline.Script     `json:"script,omitempty"`      // origination
}
