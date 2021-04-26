// Copyright (c) 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//

package micheline

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"testing"

	"blockwatch.cc/tzgo/tezos"
)

type keyTest struct {
	Name   string
	Type   string
	Hash   tezos.ExprHash
	Hex    string
	Value  json.RawMessage
	String string
	Prim   Prim
}

var keyInfo = []keyTest{
	// scalars
	//   int
	keyTest{
		Name:   "int",
		Type:   "int",
		Hash:   tezos.MustParseExprHash("exprv6n4YrvfCD2N6JmSF9aZxtcrcDCDV5YAFpaJDhJU6bhmNHz3YK"),
		Hex:    "00a005",
		Value:  []byte(`"352"`),
		String: "352",
		Prim:   NewInt64(352),
	},
	//   nat
	keyTest{
		Name:   "nat",
		Type:   "nat",
		Hash:   tezos.MustParseExprHash("expruE5MGe6oKRLTiog6iBZzpztj5kCGzMEYBfWzsVebPnhn43ndYa"),
		Hex:    "008eb818",
		Value:  []byte(`"200206"`),
		String: "200206",
		Prim:   NewInt64(200206),
	},
	//   mutez
	keyTest{
		Name:   "mutez",
		Type:   "mutez",
		Hash:   tezos.MustParseExprHash("expruE5MGe6oKRLTiog6iBZzpztj5kCGzMEYBfWzsVebPnhn43ndYa"),
		Hex:    "008eb818",
		Value:  []byte(`"200206"`),
		String: "200206",
		Prim:   NewInt64(200206),
	},
	//   unit
	keyTest{
		Name:   "unit",
		Type:   "unit",
		Hash:   tezos.MustParseExprHash("expruaDPoTWXcTR6fiQPy4KZSW72U6Swc1rVmMiP1KdwmCceeEpVjd"),
		Hex:    "030b",
		Value:  []byte(`null`),
		String: "Unit",
		Prim:   NewCode(D_UNIT),
	},
	//   string
	keyTest{
		Name:   "string",
		Type:   "string",
		Hash:   tezos.MustParseExprHash("exprtiRSZkLKYRess9GZ3ryb4cVQD36WLo2oysZBFxKTZ2jXqcHWGj"),
		Hex:    "010000000947616d65206f6e6521",
		Value:  []byte(`"Game one!"`),
		String: "Game one!",
		Prim:   NewString("Game one!"),
	},
	//   bytes
	keyTest{
		Name:   "bytes",
		Type:   "bytes",
		Hash:   tezos.MustParseExprHash("expruLUtQBGu3aw4onM4nA9A8UM7PrDh3pbcrqnrFpzAAuTd12Ggdv"),
		Hex:    "0a000000209e2dce28b861646c13c2b9cf4245e483a837525f2d1547a410293e29c3734d6e",
		Value:  []byte(`"9e2dce28b861646c13c2b9cf4245e483a837525f2d1547a410293e29c3734d6e"`),
		String: "9e2dce28b861646c13c2b9cf4245e483a837525f2d1547a410293e29c3734d6e",
		Prim:   NewBytes([]byte{0x9e, 0x2d, 0xce, 0x28, 0xb8, 0x61, 0x64, 0x6c, 0x13, 0xc2, 0xb9, 0xcf, 0x42, 0x45, 0xe4, 0x83, 0xa8, 0x37, 0x52, 0x5f, 0x2d, 0x15, 0x47, 0xa4, 0x10, 0x29, 0x3e, 0x29, 0xc3, 0x73, 0x4d, 0x6e}),
	},
	//   key_hash
	keyTest{
		Name:   "tz1 key_hash",
		Type:   "key_hash",
		Hash:   tezos.MustParseExprHash("exprtc5kFMbFCzkSq1hZkDrLWqQBUdpPx3URB7KvEu72XBFBqv7k72"),
		Hex:    "0a000000150046a2c00eb115343242347fa1cd672a2bc1dcc609",
		Value:  []byte(`"tz1S5WxdZR5f9NzsPXhr7L9L1vrEb5spZFur"`),
		String: "tz1S5WxdZR5f9NzsPXhr7L9L1vrEb5spZFur",
		Prim:   NewBytes(tezos.MustParseAddress("tz1S5WxdZR5f9NzsPXhr7L9L1vrEb5spZFur").Bytes22()),
	},
	//   address
	keyTest{
		Name:   "tz1 address",
		Type:   "address",
		Hash:   tezos.MustParseExprHash("expruQacisQeiLaWSgSHFeLA4BdLfS6yswqYQ8gjYSmJABQ9Sf53Y4"),
		Hex:    "0a000000160000b8c25930f179a13ffeefa8b0026318f7e508a8fc",
		Value:  []byte(`"tz1cUwqynCFDp1D22kLNtWMKxpoZFDHg5eZH"`),
		String: "tz1cUwqynCFDp1D22kLNtWMKxpoZFDHg5eZH",
		Prim:   NewBytes(tezos.MustParseAddress("tz1cUwqynCFDp1D22kLNtWMKxpoZFDHg5eZH").Bytes22()),
	},
	keyTest{
		Name:   "KT1 address",
		Type:   "address",
		Hash:   tezos.MustParseExprHash("exprvAHu1SyoiSzyh9w7GPfifvyrNiMb442y7Q2MA8tcPCGPajxRH6"),
		Hex:    "0a0000001601a3d0f58d8964bd1b37fb0a0c197b38cf46608d4900",
		Value:  []byte(`"KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn"`),
		String: "KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn",
		Prim:   NewBytes(tezos.MustParseAddress("KT1PWx2mnDueood7fEmfbBDKx1D9BAnnXitn").Bytes22()),
	},

	//   timestamp as string
	//   timestamp as unix seconds
	//   timestamp with out-of-range year
	//   key
	keyTest{
		Name:   "key",
		Type:   "key",
		Hash:   tezos.MustParseExprHash("exprv1Vjr2jWEzSALFrHaoubi3jELpXvnMtGNG4ZJPDMRHxrQtyBDW"),
		Hex:    "0a000000210078149c2d111816aaef9e329970c344fc32375dd5eff99eeeed3b37a9d51beacd",
		Value:  []byte(`"edpkuZ7ERiU5B8knLqQsVMH86j9RLMUyHyL665oCXDkPQxF7HGqSeJ"`),
		String: "edpkuZ7ERiU5B8knLqQsVMH86j9RLMUyHyL665oCXDkPQxF7HGqSeJ",
		Prim:   NewBytes(tezos.MustParseKey("edpkuZ7ERiU5B8knLqQsVMH86j9RLMUyHyL665oCXDkPQxF7HGqSeJ").Bytes()),
	},
	//   signature
	//   bool

	// composites
	//   pair
	keyTest{
		Name:   "Pair(address,nat)",
		Type:   "pair",
		Hash:   tezos.MustParseExprHash("exprvD1v8DxXvrsCqbx7BA2ZqxYuUk9jXE1QrXuL46i3MWG6o1szUq"),
		Hex:    "07070a0000001600005db799bf9b0dc319ba1cf21ab01461a9639043ca009902",
		Value:  []byte(`{"0":"tz1UBZUkXpKGhYsP5KtzDNqLLchwF4uHrGjw","1":"153"}`),
		String: "tz1UBZUkXpKGhYsP5KtzDNqLLchwF4uHrGjw,153",
		Prim: NewPairValue(
			NewBytes(tezos.MustParseAddress("tz1UBZUkXpKGhYsP5KtzDNqLLchwF4uHrGjw").Bytes22()),
			NewInt64(153),
		),
	},
	keyTest{
		Name:   "Pair(address,string)",
		Type:   "pair",
		Hash:   tezos.MustParseExprHash("exprtdyqcWJgj564TtpwqXkkHQu728pP4hVM7vdc16RVXaSbWoJttS"),
		Hex:    "07070a000000160000fe533bc9b5847c783653835de5889ba6a946a2b601000000094449435245524a3238",
		Value:  []byte(`{"0":"tz1ipn31fhqk47Tr3f7KZAeCamyqMBXBAKBi","1":"DICRERJ28"}`),
		String: "tz1ipn31fhqk47Tr3f7KZAeCamyqMBXBAKBi,DICRERJ28",
		Prim: NewPairValue(
			NewBytes(tezos.MustParseAddress("tz1ipn31fhqk47Tr3f7KZAeCamyqMBXBAKBi").Bytes22()),
			NewString("DICRERJ28"),
		),
	},
	keyTest{
		Name:   "Pair(address,address,nat)",
		Type:   "pair",
		Hash:   tezos.MustParseExprHash("exprtXiCYp3hWQMDQNszcmsigcU13M32bzQYLDaQp35t2F1Nqj6tiW"),
		Hex:    "07070a00000016000060d8a7d4dc6eee130387b916a79193e65e6ef65107070a0000001601c953466e6cf9295ecad052acc742247dd0e3c91900008f01",
		Value:  []byte(`{"0":"tz1UU772ew1GALQ2Uh8fCCN4uhzWBzSQH4Az","1":"KT1SwH9P1Tx8a58Mm6qBExQFTcy2rwZyZiXS","2":"79"}`),
		String: "tz1UU772ew1GALQ2Uh8fCCN4uhzWBzSQH4Az,KT1SwH9P1Tx8a58Mm6qBExQFTcy2rwZyZiXS,79",
		Prim: NewPairValue(
			NewBytes(tezos.MustParseAddress("tz1UU772ew1GALQ2Uh8fCCN4uhzWBzSQH4Az").Bytes22()),
			NewPairValue(
				NewBytes(tezos.MustParseAddress("KT1SwH9P1Tx8a58Mm6qBExQFTcy2rwZyZiXS").Bytes22()),
				NewInt64(79),
			),
		),
	},
	//   option
	//   or / union

	// specials
	//   packed key

}

func TestKeyParser(t *testing.T) {
	for _, test := range keyInfo {
		t.Run(test.Name, func(T *testing.T) {
			typ, err := ParseKeyType(test.Type)
			if err != nil {
				T.Errorf("key type error: %v", err)
				T.FailNow()
			}
			key, err := ParseKey(typ, test.String)
			if err != nil {
				T.Errorf("key parse error: %v", err)
				T.FailNow()
			}

			// encode and check against bytes
			buf := key.Bytes()
			hx, _ := hex.DecodeString(test.Hex)
			if bytes.Compare(buf, hx) != 0 {
				T.Errorf("binary encoding mismatch:\n    want: %x\n    got:  %x", hx, buf)
			}

			// check against prim
			prim := key.Prim()
			if !prim.IsEqual(test.Prim) {
				T.Errorf("prim mismatch:\n    want: %s\n    got:  %s", test.Prim.Dump(), prim.Dump())
			}

			// generate hash and compare
			hs := key.Hash()
			if !hs.Equal(test.Hash) {
				T.Errorf("hash mismatch:\n    want: %s\n    got:  %s", test.Hash, hs)
			}
		})
	}
}

func TestKeyRendering(t *testing.T) {
	for _, test := range keyInfo {
		t.Run(test.Name, func(T *testing.T) {
			typ, err := ParseKeyType(test.Type)
			if err != nil {
				T.Errorf("key type error: %v", err)
				T.FailNow()
			}
			key, err := NewKey(NewType(NewPrim(typ)), test.Prim)
			if err != nil {
				T.Errorf("new key error: %v", err)
				T.FailNow()
			}

			// marshal string
			if want, have := test.String, key.String(); want != have {
				T.Errorf("string mismatch:\n    want: %s\n    got:  %s", want, have)
			}

			// encode and check against bytes
			buf := key.Bytes()
			hx, _ := hex.DecodeString(test.Hex)
			if bytes.Compare(buf, hx) != 0 {
				T.Errorf("binary encoding mismatch:\n    want: %x\n    got:  %x", hx, buf)
			}

			// marshal JSON
			have, err := key.MarshalJSON()
			if err != nil {
				T.Errorf("marshal error: %v", err)
			}
			if !jsonDiff(T, have, []byte(test.Value)) {
				T.Error("render mismatch, see log for details")
				t.FailNow()
			}

			// generate hash and compare
			hs := key.Hash()
			if !hs.Equal(test.Hash) {
				T.Errorf("hash mismatch:\n    want: %s\n    got:  %s", test.Hash, hs)
			}
		})
	}
}
