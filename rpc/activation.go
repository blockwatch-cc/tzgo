// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// Ensure Activation implements the TypedOperation interface.
var _ TypedOperation = (*Activation)(nil)

// Activation represents a transaction operation
type Activation struct {
	Generic
	Pkh    tezos.Address  `json:"pkh"`
	Secret tezos.HexBytes `json:"secret"`
}
