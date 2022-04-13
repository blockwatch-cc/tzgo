// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

import (
	"github.com/legonian/tzgo/tezos"
)

// Represents Tzip12 token metadata used by FA1 and FA2 tokens
type TokenMetadata struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type TokenKind byte

const (
	TokenKindInvalid TokenKind = iota
	TokenKindFA1
	TokenKindFA1_2
	TokenKindFA2
	TokenKindNFT
)

func (k TokenKind) String() string {
	switch k {
	case TokenKindFA1:
		return "fa1"
	case TokenKindFA1_2:
		return "fa1_2"
	case TokenKindFA2:
		return "fa2"
	case TokenKindNFT:
		return "nft"
	default:
		return ""
	}
}

type TokenBalance struct {
	Owner   tezos.Address
	Token   tezos.Address
	TokenId tezos.Z
	Balance tezos.Z
}
