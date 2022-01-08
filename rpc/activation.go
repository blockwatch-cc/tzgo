// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// AccountActivationOp represents a transaction operation
type AccountActivationOp struct {
	GenericOp
	Pkh      tezos.Address                `json:"pkh"`
	Secret   tezos.HexBytes               `json:"secret"`
	Metadata *AccountActivationOpMetadata `json:"metadata"`
}

// AccountActivationOpMetadata represents a transaction operation metadata
type AccountActivationOpMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"` // initial funding
}
