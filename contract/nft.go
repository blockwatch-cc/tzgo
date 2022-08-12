// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package contract

import (
	"fmt"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

type NftLedgerSchema byte

const (
	NftLedgerSchemaInvalid NftLedgerSchema = iota
	NftLedgerSchema1
	NftLedgerSchema2
	NftLedgerSchema3
)

func (s NftLedgerSchema) IsValid() bool {
	return s != NftLedgerSchemaInvalid
}

var nftLedgerSpecs = map[NftLedgerSchema]micheline.Prim{
	// 1 @key: {0: address, 1: nat}     @value: nat
	NftLedgerSchema1: micheline.NewPairType(
		micheline.NewPairType(
			micheline.NewCode(micheline.T_ADDRESS), // owner
			micheline.NewCode(micheline.T_NAT),     // token_id
		),
		micheline.NewCode(micheline.T_NAT), // balance
	),
	// 2 @key: nat  @value: address
	NftLedgerSchema2: micheline.NewPairType(
		micheline.NewCode(micheline.T_NAT),     // token_id
		micheline.NewCode(micheline.T_ADDRESS), // owner
	),
	// 3 @key: {0: nat, 1: address}     @value: nat
	NftLedgerSchema3: micheline.NewPairType(
		micheline.NewPairType(
			micheline.NewCode(micheline.T_NAT),     // token_id
			micheline.NewCode(micheline.T_ADDRESS), // owner
		),
		micheline.NewCode(micheline.T_NAT), // balance
	),
}

// Detect NFT legder schema from bigmap key + value type.
func DetectNftLedger(key, val micheline.Prim) NftLedgerSchema {
	if !key.IsValid() || !val.IsValid() {
		return NftLedgerSchemaInvalid
	}
	for t, v := range nftLedgerSpecs {
		if !v.Args[0].IsEqual(key) {
			continue
		}
		if !v.Args[1].IsEqual(val) {
			continue
		}
		return t
	}
	return NftLedgerSchemaInvalid
}

type NftLedger struct {
	Address    tezos.Address
	Schema     NftLedgerSchema
	Bigmap     int64
	FirstBlock int64
}

func (l NftLedger) DecodeEntry(prim micheline.Prim) (bal NftLedgerEntry, err error) {
	bal.schema = l.Schema
	err = prim.Decode(&bal)
	return
}

type NftLedgerEntry struct {
	Owner   tezos.Address
	TokenId tezos.Z
	Balance tezos.Z
	schema  NftLedgerSchema
}

func (b *NftLedgerEntry) UnmarshalPrim(prim micheline.Prim) error {
	// schema switch to select the chosen struct type with custom struct tags
	switch b.schema {
	case NftLedgerSchema1:
		return prim.Decode((*NftLedger1)(b))
	case NftLedgerSchema2:
		return prim.Decode((*NftLedgerTyp2)(b))
	case NftLedgerSchema3:
		return prim.Decode((*NftLedgerTyp3)(b))
	default:
		return fmt.Errorf("unsupported NFT ledger type %d", b.schema)
	}
}

// 1 @key: {0: address, 1: nat}     @value: nat
type NftLedger1 NftLedgerEntry

func (b *NftLedger1) UnmarshalPrim(prim micheline.Prim) error {
	var alias struct {
		Owner   tezos.Address `prim:"owner,path=0/0"`
		TokenId tezos.Z       `prim:"token_id,path=0/1"`
		Balance tezos.Z       `prim:"balance,path=1"`
	}
	err := prim.Decode(&alias)
	if err == nil {
		b.Owner = alias.Owner
		b.TokenId = alias.TokenId
		b.Balance = alias.Balance
	}
	return err
}

// 2 @key: nat  @value: address
type NftLedgerTyp2 NftLedgerEntry

func (b *NftLedgerTyp2) UnmarshalPrim(prim micheline.Prim) error {
	var alias struct {
		TokenId tezos.Z       `prim:"token_id,path=0"`
		Owner   tezos.Address `prim:"owner,path=1"`
	}
	err := prim.Decode(&alias)
	if err == nil {
		b.Owner = alias.Owner
		b.TokenId = alias.TokenId
		b.Balance.SetInt64(1)
	}
	return err
}

// 3 @key: {0: nat, 1: address}     @value: nat
type NftLedgerTyp3 NftLedgerEntry

func (b *NftLedgerTyp3) UnmarshalPrim(prim micheline.Prim) error {
	var alias struct {
		TokenId tezos.Z       `prim:"token_id,path=0/0"`
		Owner   tezos.Address `prim:"owner,path=0/1"`
		Balance tezos.Z       `prim:"balance,path=1"`
	}
	err := prim.Decode(&alias)
	if err == nil {
		b.Owner = alias.Owner
		b.TokenId = alias.TokenId
		b.Balance = alias.Balance
	}
	return err
}
