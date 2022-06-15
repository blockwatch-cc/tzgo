// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// Ensure ConstantRegistration implements the TypedOperation interface.
var _ TypedOperation = (*ConstantRegistration)(nil)

// ConstantRegistration represents a global constant registration operation
type ConstantRegistration struct {
	Manager
	Value micheline.Prim `json:"value,omitempty"`
}

// Costs returns operation cost to implement TypedOperation interface.
func (c ConstantRegistration) Costs() tezos.Costs {
	res := c.Metadata.Result
	burn := res.BalanceUpdates[0].Amount()
	return tezos.Costs{
		Fee:         c.Manager.Fee,
		GasUsed:     res.Gas(),
		Burn:        -burn,
		StorageUsed: res.StorageSize,
		StorageBurn: -burn,
	}
}
