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
	Delegate tezos.Address `json:"delegate,omitempty"`
}

// Cost returns operation cost to implement TypedOperation interface.
func (d Delegation) Costs() tezos.Costs {
	return tezos.Costs{
		Fee:     d.Manager.Fee,
		GasUsed: d.Metadata.Result.ConsumedGas,
	}
}
