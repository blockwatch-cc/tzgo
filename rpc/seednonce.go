// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// SeedNonceOp represents a seed_nonce_revelation operation
type SeedNonceOp struct {
	GenericOp
	Level    int64                `json:"level"`
	Nonce    tezos.NonceHash      `json:"nonce"`
	Metadata *SeedNonceOpMetadata `json:"metadata"`
}

// SeedNonceOpMetadata represents a transaction operation metadata
type SeedNonceOpMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates"` // fee-related
}
