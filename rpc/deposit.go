// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// Ensure SetDepositsLimit implements the TypedOperation interface.
var _ TypedOperation = (*SetDepositsLimit)(nil)

// SetDepositsLimit represents a baker deposit limit update operation.
type SetDepositsLimit struct {
	Manager
	Limit    int64             `json:"limit,string"`
	Metadata OperationMetadata `json:"metadata"`
}

// Meta returns operation metadata to implement TypedOperation interface.
func (r SetDepositsLimit) Meta() OperationMetadata {
	return r.Metadata
}

// Result returns operation result to implement TypedOperation interface.
func (r SetDepositsLimit) Result() OperationResult {
	return r.Metadata.Result
}

// Costs returns operation cost to implement TypedOperation interface.
func (r SetDepositsLimit) Costs() tezos.Costs {
	return tezos.Costs{
		Fee:     r.Manager.Fee,
		GasUsed: r.Metadata.Result.Gas(),
	}
}
