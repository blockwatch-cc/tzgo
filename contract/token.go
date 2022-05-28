// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

import (
	"context"
	"fmt"
	"strconv"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

type TokenKind byte

const (
	TokenKindInvalid TokenKind = iota
	TokenKindTez
	TokenKindFA1
	TokenKindFA1_2
	TokenKindFA2
	TokenKindNFT
	TokenKindNoView
)

func (k TokenKind) String() string {
	switch k {
	case TokenKindTez:
		return "tez"
	case TokenKindFA1:
		return "fa1"
	case TokenKindFA1_2:
		return "fa1_2"
	case TokenKindFA2:
		return "fa2"
	case TokenKindNFT:
		return "nft"
	case TokenKindNoView:
		return "noview"
	default:
		return ""
	}
}

func (k TokenKind) IsValid() bool {
	return k != TokenKindInvalid
}

const TOKEN_METADATA = "token_metadata"

// Represents Tzip12 token metadata used by FA1 and FA2 tokens
type TokenMetadata struct {
	// normative (only decimals is required)
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`

	// non-standard (people mix this in from Tzip21)
	Description        string `json:"description,omitempty"`
	ThumbnailUri       string `json:"thumbnailUri,omitempty"`
	ShouldPreferSymbol bool   `json:"shouldPreferSymbol,omitempty"`
	IsBooleanAmount    bool   `json:"is_boolean_amount,omitempty"`
	IsTransferable     bool   `json:"is_transferable,omitempty"`

	// internal
	uri string `json:"-"`
}

// (pair (nat %token_id) (map %token_info string bytes))
func (t *TokenMetadata) UnmarshalPrim(prim micheline.Prim) error {
	return prim.Args[1].Walk(func(p micheline.Prim) error {
		if p.IsSequence() {
			return nil
		}
		if !p.IsElt() {
			return fmt.Errorf("unexpected map item %q", p.Dump())
		}
		field := p.Args[0].String
		data := p.Args[1].Bytes
		// unpack packed bytes (some people do that, yes)
		if p.Args[1].IsPacked() {
			p, _ := p.Args[1].Unpack()
			if p.Type == micheline.PrimBytes {
				data = p.Bytes
			} else {
				data = []byte(p.String)
			}
		}
		switch field {
		case "":
			t.uri = string(data)
		case "name":
			t.Name = string(data)
		case "description":
			t.Description = string(data)
		case "symbol":
			t.Symbol = string(data)
		case "icon", "logo", "thumbnailUri", "thumbnail_uri":
			t.ThumbnailUri = string(data)
		case "decimals":
			d, err := strconv.Atoi(string(data))
			if err != nil {
				return fmt.Errorf("%q: %v", field, err)
			}
			t.Decimals = d
		case "shouldPreferSymbol", "should_prefer_symbol":
			b, err := strconv.ParseBool(string(data))
			if err != nil {
				return fmt.Errorf("%q: %v", field, err)
			}
			t.ShouldPreferSymbol = b
		case "isBooleanAmount", "is_boolean_amount":
			b, err := strconv.ParseBool(string(data))
			if err != nil {
				return fmt.Errorf("%q: %v", field, err)
			}
			t.IsBooleanAmount = b
		case "isTransferable", "is_transferable":
			b, err := strconv.ParseBool(string(data))
			if err != nil {
				return fmt.Errorf("%q: %v", field, err)
			}
			t.IsTransferable = b
		default:
			fmt.Printf("token metadata: unsupported field %q\n", field)
		}
		return micheline.PrimSkip
	})
}

func ResolveTokenMetadata(ctx context.Context, contract *Contract, tokenid tezos.Z) (*TokenMetadata, error) {
	var (
		store micheline.Prim
		err   error
	)

	// we need contract script and storage
	if err = contract.Resolve(ctx); err != nil {
		return nil, err
	}

	// lookup well known (pre-tz16 or wrong) tokens
	if m, ok := wellKnown[contract.Address().String()]; ok {
		return m, nil
	}

	// prefer off-chain view via run_code, but don't fail if not present
	tz16, _ := contract.ResolveMetadata(ctx)
	if tz16 != nil && tz16.HasView(TOKEN_METADATA) {
		view := tz16.GetView(TOKEN_METADATA)
		args := micheline.NewNat(tokenid.Big())
		store, err = view.Run(ctx, contract, args)
		if err != nil {
			return nil, err
		}
	} else {
		// read token_metadata bigmap
		bigmaps := contract.script.BigmapsByName()
		id, ok := bigmaps[TOKEN_METADATA]
		if !ok {
			return nil, fmt.Errorf("%s/%d: missing token metadata, have %v", contract.addr, tokenid.Int64(), bigmaps)
		}
		hash := (micheline.Key{
			Type:   micheline.NewType(micheline.NewPrim(micheline.T_NAT)),
			IntKey: tokenid.Big(),
		}).Hash()
		store, err = contract.rpc.GetActiveBigmapValue(ctx, id, hash)
		if err != nil {
			return nil, err
		}
	}

	// parse storage: (pair (nat %token_id) (map %token_info string bytes))
	meta := &TokenMetadata{}
	if err := meta.UnmarshalPrim(store); err != nil {
		return nil, err
	}

	// should forward?
	if meta.uri != "" {
		if err := contract.ResolveTz16Uri(ctx, meta.uri, meta, nil); err != nil {
			return nil, err
		}
	}

	// fill empty token name from contract metadata
	if meta.Name == "" && tz16 != nil {
		meta.Name = tz16.Name
	}

	return meta, nil
}

type TokenBalance struct {
	Owner   tezos.Address
	Token   tezos.Address
	TokenId tezos.Z
	Balance tezos.Z
}
