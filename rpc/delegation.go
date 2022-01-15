// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// Ensure Delegation implements the TypedOperation interface.
var _ TypedOperation = (*Delegation)(nil)

// Delegation represents a transaction operation
type Delegation struct {
	Manager
	Delegate tezos.Address     `json:"delegate,omitempty"`
	Metadata OperationMetadata `json:"metadata"`
}

// Meta returns an empty operation metadata to implement TypedOperation interface.
func (d Delegation) Meta() OperationMetadata {
	return d.Metadata
}

// Result returns an empty operation result to implement TypedOperation interface.
func (d Delegation) Result() OperationResult {
	return d.Metadata.Result
}

// Cost returns operation cost to implement TypedOperation interface.
func (d Delegation) Cost() OperationCost {
	return OperationCost{
		Fee: d.Manager.Fee,
		Gas: d.Metadata.Result.ConsumedGas,
	}
}
