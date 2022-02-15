// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// BalanceUpdate is a variable structure depending on the Kind field
type BalanceUpdate struct {
	Kind   string `json:"kind"`
	Change int64  `json:"change,string"`
	Origin string `json:"origin"` // block, migration, subsidy

	// Contract only
	Contract tezos.Address `json:"contract"`

	// Freezer only
	Category string        `json:"category"`
	Delegate tezos.Address `json:"delegate"`
	Level_   int64         `json:"level"` // wrongly called level, it's cycle
	Cycle_   int64         `json:"cycle"` // v4 fix
}

func (b BalanceUpdate) Address() tezos.Address {
	switch b.Kind {
	case "contract":
		return b.Contract
	case "freezer":
		return b.Delegate
	}
	return tezos.Address{}
}

func (b BalanceUpdate) Amount() int64 {
	return b.Change
}

func (b BalanceUpdate) Cycle() int64 {
	if b.Level_ > 0 {
		return b.Level_
	}
	return b.Cycle_
}

// BalanceUpdates is a list of balance update operations
type BalanceUpdates []BalanceUpdate
