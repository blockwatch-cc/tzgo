// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// Ensure SeedNonce implements the TypedOperation interface.
var _ TypedOperation = (*SeedNonce)(nil)

// SeedNonce represents a seed_nonce_revelation operation
type SeedNonce struct {
	Generic
	Level    int64             `json:"level"`
	Nonce    tezos.HexBytes    `json:"nonce"`
	Metadata OperationMetadata `json:"metadata"`
}

// Meta returns an empty operation metadata to implement TypedOperation interface.
func (s SeedNonce) Meta() OperationMetadata {
	return s.Metadata
}
