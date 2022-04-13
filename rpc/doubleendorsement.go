// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"github.com/legonian/tzgo/tezos"
)

// Ensure DoubleEndorsement implements the TypedOperation interface.
var _ TypedOperation = (*DoubleEndorsement)(nil)

// DoubleEndorsement represents a double_endorsement_evidence operation
type DoubleEndorsement struct {
	Generic
	OP1      InlinedEndorsement `json:"op1"`
	OP2      InlinedEndorsement `json:"op2"`
	Metadata OperationMetadata  `json:"metadata"`
}

// Meta returns operation metadata to implement TypedOperation interface.
func (d DoubleEndorsement) Meta() OperationMetadata {
	return d.Metadata
}

// Costs returns operation cost to implement TypedOperation interface.
func (d DoubleEndorsement) Costs() tezos.Costs {
	var burn int64
	upd := d.Metadata.BalanceUpdates
	// last item is accuser reward, rest is burned
	for i, v := range upd {
		if i == len(upd)-1 {
			burn += v.Amount()
		} else {
			burn += v.Amount()
		}
	}
	return tezos.Costs{
		Burn: -burn,
	}
}
