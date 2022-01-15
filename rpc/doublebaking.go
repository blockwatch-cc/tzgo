// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

// Ensure DoubleBaking implements the TypedOperation interface.
var _ TypedOperation = (*DoubleBaking)(nil)

// DoubleBaking represents a double_baking_evidence operation
type DoubleBaking struct {
	Generic
	BH1      BlockHeader       `json:"bh1"`
	BH2      BlockHeader       `json:"bh2"`
	Metadata OperationMetadata `json:"metadata"`
}

// Meta returns operation metadata to implement TypedOperation interface.
func (d DoubleBaking) Meta() OperationMetadata {
	return d.Metadata
}

// Cost returns operation cost to implement TypedOperation interface.
func (d DoubleBaking) Cost() OperationCost {
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
	return OperationCost{
		Burn: -burn,
	}
}
