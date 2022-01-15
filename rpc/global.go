// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
)

// Ensure ConstantRegistration implements the TypedOperation interface.
var _ TypedOperation = (*ConstantRegistration)(nil)

// ConstantRegistration represents a global constant registration operation
type ConstantRegistration struct {
	Manager
	Value    micheline.Prim    `json:"value,omitempty"`
	Metadata OperationMetadata `json:"metadata"`
}

// Meta returns an empty operation metadata to implement TypedOperation interface.
func (c ConstantRegistration) Meta() OperationMetadata {
	return c.Metadata
}

// Result returns an empty operation result to implement TypedOperation interface.
func (c ConstantRegistration) Result() OperationResult {
	return c.Metadata.Result
}

// Cost returns operation cost to implement TypedOperation interface.
func (c ConstantRegistration) Cost() OperationCost {
	res := c.Metadata.Result
	burn := res.BalanceUpdates[0].Amount()
	return OperationCost{
		Fee:          c.Manager.Fee,
		Gas:          res.ConsumedGas,
		Burn:         -burn,
		StorageBytes: res.StorageSize,
		StorageBurn:  -burn,
	}
}
