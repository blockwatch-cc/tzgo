// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//

package micheline

import (
	"testing"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

type marshalTest struct {
	Name      string
	Spec      string
	Value     any
	Optimized bool
	Want      string
}

var marshalTests = []marshalTest{
	// scalars
	// int
	{
		Name:      "int",
		Spec:      `{"annots": ["%payoutDelay"],"prim": "int"}`,
		Value:     map[string]any{"payoutDelay": 1},
		Optimized: false,
		Want:      `{"int":"1"}`,
	},
	//   nat
	{
		Name:      "nat",
		Spec:      `{"annots": ["%payoutFrequency"],"prim": "nat"}`,
		Value:     map[string]any{"payoutFrequency": 1},
		Optimized: false,
		Want:      `{"int":"1"}`,
	},
	//   mutez
	{
		Name:      "mutez",
		Spec:      `{"annots": ["%money"],"prim": "mutez"}`,
		Value:     map[string]any{"money": 1},
		Optimized: false,
		Want:      `{"int":"1"}`,
	},
	//   string
	{
		Name:      "string",
		Spec:      `{"annots": ["%name"],"prim": "string"}`,
		Value:     map[string]any{"name": "hello"},
		Optimized: false,
		Want:      `{"string":"hello"}`,
	},
	//   bytes
	{
		Name:      "bytes",
		Spec:      `{"annots": ["%bakerName"],"prim": "bytes"}`,
		Value:     map[string]any{"bakerName": []byte{0xfa, 0xfe}},
		Optimized: false,
		Want:      `{"bytes":"fafe"}`,
	},
	//   bool
	{
		Name:      "bool",
		Spec:      `{"annots": ["%bakerChargesTransactionFee"],"prim": "bool"}`,
		Value:     map[string]any{"bakerChargesTransactionFee": true},
		Optimized: false,
		Want:      `{"prim":"True"}`,
	},
	//   timestamp
	{
		Name:      "timestamp_unix_noopt",
		Spec:      `{"annots": ["%last_update"],"prim": "timestamp"}`,
		Value:     map[string]any{"last_update": time.Unix(1696945397, 0)},
		Optimized: false,
		Want:      `{"string":"2023-10-10T13:43:17Z"}`,
	},
	{
		Name:      "timestamp_unix_opt",
		Spec:      `{"annots": ["%last_update"],"prim": "timestamp"}`,
		Value:     map[string]any{"last_update": time.Unix(1696945397, 0)},
		Optimized: true,
		Want:      `{"int":"1696945397"}`,
	},
	{
		Name:      "timestamp_rfc_noopt",
		Spec:      `{"annots": ["%last_update"],"prim": "timestamp"}`,
		Value:     map[string]any{"last_update": "2023-10-10T13:43:17Z"},
		Optimized: false,
		Want:      `{"string":"2023-10-10T13:43:17Z"}`,
	},
	{
		Name:      "timestamp_rfc_opt",
		Spec:      `{"annots": ["%last_update"],"prim": "timestamp"}`,
		Value:     map[string]any{"last_update": "2023-10-10T13:43:17Z"},
		Optimized: true,
		Want:      `{"int":"1696945397"}`,
	},
	//   key_hash
	{
		Name:      "key_hash_string_opt",
		Spec:      `{"annots": ["%baker"],"prim": "key_hash"}`,
		Value:     map[string]any{"baker": "tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},
		Optimized: false,
		Want:      `{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}`,
	},
	{
		Name:      "key_hash_string_noopt",
		Spec:      `{"annots": ["%baker"],"prim": "key_hash"}`,
		Value:     map[string]any{"baker": "tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},
		Optimized: true,
		Want:      `{"bytes":"00cf8dad6c9cd30672987242a8c2a94fc72816d8cf"}`,
	},
	{
		Name:      "key_hash_addr_opt",
		Spec:      `{"annots": ["%baker"],"prim": "key_hash"}`,
		Value:     map[string]any{"baker": tezos.MustParseAddress("tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc")},
		Optimized: false,
		Want:      `{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}`,
	},
	{
		Name:      "key_hash_addr_noopt",
		Spec:      `{"annots": ["%baker"],"prim": "key_hash"}`,
		Value:     map[string]any{"baker": tezos.MustParseAddress("tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc")},
		Optimized: true,
		Want:      `{"bytes":"00cf8dad6c9cd30672987242a8c2a94fc72816d8cf"}`,
	},
	//   address
	{
		Name:      "address_string_opt",
		Spec:      `{"annots": ["%reporterAccount"],"prim": "address"}`,
		Value:     map[string]any{"reporterAccount": "tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},
		Optimized: false,
		Want:      `{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}`,
	},
	{
		Name:      "address_string_noopt",
		Spec:      `{"annots": ["%reporterAccount"],"prim": "address"}`,
		Value:     map[string]any{"reporterAccount": "tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},
		Optimized: true,
		Want:      `{"bytes":"0000cf8dad6c9cd30672987242a8c2a94fc72816d8cf"}`,
	},
	{
		Name:      "address_addr_opt",
		Spec:      `{"annots": ["%reporterAccount"],"prim": "address"}`,
		Value:     map[string]any{"reporterAccount": tezos.MustParseAddress("tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc")},
		Optimized: false,
		Want:      `{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}`,
	},
	{
		Name:      "address_addr_noopt",
		Spec:      `{"annots": ["%reporterAccount"],"prim": "address"}`,
		Value:     map[string]any{"reporterAccount": tezos.MustParseAddress("tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc")},
		Optimized: true,
		Want:      `{"bytes":"0000cf8dad6c9cd30672987242a8c2a94fc72816d8cf"}`,
	},
	//   key
	{
		Name:      "key_string_noopt",
		Spec:      `{"annots": ["%pour_authorizer"],"prim": "key"}`,
		Value:     map[string]any{"pour_authorizer": "edpkvEfPbB2Q8dpo8D7DcLXC7ft4ogfeayPzxnvTiuz3iKM9TonxHh"},
		Optimized: false,
		Want:      `{"string":"edpkvEfPbB2Q8dpo8D7DcLXC7ft4ogfeayPzxnvTiuz3iKM9TonxHh"}`,
	},
	{
		Name:      "key_string_opt",
		Spec:      `{"annots": ["%pour_authorizer"],"prim": "key"}`,
		Value:     map[string]any{"pour_authorizer": "edpkvEfPbB2Q8dpo8D7DcLXC7ft4ogfeayPzxnvTiuz3iKM9TonxHh"},
		Optimized: true,
		Want:      `{"bytes":"00d1e4cae729906793005dbdfddef090c4153042bf922e4db3a99a0467c45b9898"}`,
	},
	{
		Name:      "key_key_noopt",
		Spec:      `{"annots": ["%pour_authorizer"],"prim": "key"}`,
		Value:     map[string]any{"pour_authorizer": tezos.MustParseKey("edpkvEfPbB2Q8dpo8D7DcLXC7ft4ogfeayPzxnvTiuz3iKM9TonxHh")},
		Optimized: false,
		Want:      `{"string":"edpkvEfPbB2Q8dpo8D7DcLXC7ft4ogfeayPzxnvTiuz3iKM9TonxHh"}`,
	},
	{
		Name:      "key_key_opt",
		Spec:      `{"annots": ["%pour_authorizer"],"prim": "key"}`,
		Value:     map[string]any{"pour_authorizer": tezos.MustParseKey("edpkvEfPbB2Q8dpo8D7DcLXC7ft4ogfeayPzxnvTiuz3iKM9TonxHh")},
		Optimized: true,
		Want:      `{"bytes":"00d1e4cae729906793005dbdfddef090c4153042bf922e4db3a99a0467c45b9898"}`,
	},
	//   unit
	{
		Name:      "unit_string",
		Spec:      `{"annots": ["%arg"],"prim": "unit"}`,
		Value:     map[string]any{"arg": ""},
		Optimized: false,
		Want:      `{"prim":"Unit"}`,
	},
	{
		Name:      "unit_nil",
		Spec:      `{"annots": ["%arg"],"prim": "unit"}`,
		Value:     map[string]any{"arg": nil},
		Optimized: false,
		Want:      `{"prim":"Unit"}`,
	},
	//   signature
	{
		Name:      "signature_string_noopt",
		Spec:      `{"annots": ["%sig"],"prim": "signature"}`,
		Value:     map[string]any{"sig": "sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"},
		Optimized: false,
		Want:      `{"string":"sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"}`,
	},
	{
		Name:      "signature_string_opt",
		Spec:      `{"annots": ["%sig"],"prim": "signature"}`,
		Value:     map[string]any{"sig": "sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"},
		Optimized: true,
		Want:      `{"bytes":"d3a9e1467b32104921d4e2dd93265739c1a5faee7a7f8880842b096c0b6714200c43fd5872f82581dfe1cb3a76ccdadaa4d6361d72b4abee6884cb7ed87f0b04"}`,
	},
	{
		Name:      "signature_sig_noopt",
		Spec:      `{"annots": ["%sig"],"prim": "signature"}`,
		Value:     map[string]any{"sig": tezos.MustParseSignature("sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy")},
		Optimized: false,
		Want:      `{"string":"sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"}`,
	},
	{
		Name:      "signature_sig_opt",
		Spec:      `{"annots": ["%sig"],"prim": "signature"}`,
		Value:     map[string]any{"sig": tezos.MustParseSignature("sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy")},
		Optimized: true,
		Want:      `{"bytes":"d3a9e1467b32104921d4e2dd93265739c1a5faee7a7f8880842b096c0b6714200c43fd5872f82581dfe1cb3a76ccdadaa4d6361d72b4abee6884cb7ed87f0b04"}`,
	},
	//   arg list
	{
		Name:      "args_named",
		Spec:      `{"args":[{"prim":"nat","annots":["%num1"]},{"args":[{"prim":"string","annots":["%name"]},{"prim":"int","annots":["%num2"]}],"prim": "pair"}],"prim": "pair"}`,
		Value:     map[string]any{"num1": 1, "name": "hello", "num2": 42},
		Optimized: false,
		Want:      `{"prim":"Pair","args":[{"int":"1"},{"prim":"Pair","args":[{"string":"hello"},{"int":"42"}]}]}`,
	},
	//   anon arg list
	{
		Name:      "args_anon",
		Spec:      `{"args":[{"prim":"nat"},{"args":[{"prim":"string"},{"prim":"int"}],"prim": "pair"}],"prim": "pair"}`,
		Value:     map[string]any{"0": 1, "1": "hello", "2": 42},
		Optimized: false,
		Want:      `{"prim":"Pair","args":[{"int":"1"},{"prim":"Pair","args":[{"string":"hello"},{"int":"42"}]}]}`,
	},
	// set
	{
		Name:      "set",
		Spec:      `{"annots": ["%admins"],"prim": "set", "args": [{"prim": "key_hash"}]}`,
		Value:     map[string]any{"admins": []any{"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}},
		Optimized: false,
		Want:      `[{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}]`,
	},
	{
		Name:      "set",
		Spec:      `{"annots": ["%admins"],"prim": "set", "args": [{"prim": "key_hash"}]}`,
		Value:     map[string]any{"admins": []any{tezos.MustParseAddress("tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc")}},
		Optimized: false,
		Want:      `[{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}]`,
	},
	// map
	{
		Name:      "map",
		Spec:      `{"annots":["%approvals"],"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]}`,
		Value:     map[string]any{"approvals": map[string]any{"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc": "1"}},
		Optimized: false,
		Want:      `[{"prim":"Elt","args":[{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},{"int":"1"}]}]`,
	},
	{
		Name:      "map",
		Spec:      `{"prim":"pair","annots":["%add_token"],"args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}`,
		Value:     map[string]any{"token_id": "1", "token_info": map[string]any{"": "ipfs://Qmb94zFKazBKxuYk4QyTWmgiVP3zXLkGjNDDTq3DShEs8E"}},
		Optimized: false,
		Want:      `{"prim":"Pair","args":[{"int":"1"},[{"prim":"Elt","args":[{"string":""},{"bytes":"697066733a2f2f516d6239347a464b617a424b7875596b34517954576d67695650337a584c6b476a4e444454713344536845733845"}]}]]}`,
	},

	// option
	{
		Name:      "option_no_value",
		Spec:      `{"annots":["%reporterAccount"],"prim":"option","args":[{"prim":"address"}]}`,
		Value:     map[string]any{"reporterAccount": nil},
		Optimized: false,
		Want:      `{"prim":"None"}`,
	},
	{
		Name:      "option_with_value",
		Spec:      `{"annots":["%reporterAccount"],"prim":"option","args":[{"prim":"address"}]}`,
		Value:     map[string]any{"reporterAccount": "tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},
		Optimized: false,
		Want:      `{"prim":"Some","args":[{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"}]}`,
	},

	// named union type
	{
		Name:      "named-union",
		Spec:      `{"annots":["%update_operators"],"args":[{"args":[{"annots":["%add_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"annots":["%remove_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"or"}],"prim":"list"}`,
		Value:     map[string]any{"update_operators": []any{map[string]any{"add_operator": map[string]any{"owner": "tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc", "operator": "tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc", "token_id": "0"}}}},
		Optimized: false,
		Want:      `[{"prim":"Left","args":[{"prim":"Pair","args":[{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},{"prim":"Pair","args":[{"string":"tz1eZUHkQDC1bBEbvrrUxkbWEagdZJXQyszc"},{"int":"0"}]}]}]}]`,
	},

	// TODO

	// // anonymous union type
	// {
	// 	Name: "anon-union",
	// 	Spec: `{"args":[{"args":[{"prim":"unit"},{"prim":"operation"}],"prim":"lambda"},{"args":[{"prim":"key_hash"}],"prim":"set"}],"prim":"or"}`,
	// },
	// // nested map
	// {
	// 	Name: "nested_map",
	// 	Spec: `{"annots": ["%deck"],"args": [{"prim": "int"},{"args": [{"prim": "int"},{"prim": "int"}],"prim": "map"}],"prim": "map"}`,
	// },
	// // nested list (FA2)
	// {
	// 	Name: "nested_list",
	// 	Spec: `{"annots": ["%transfer"],"args": [{"args": [{"annots": ["%from_"],"prim": "address"},{"annots": ["%txs"],"args": [{"args": [{"annots": ["%to_"],"prim": "address"},{"args": [{"annots": ["%token_id"],"prim": "nat"},{"annots": ["%amount"],"prim": "nat"}],"prim": "pair"}],"prim": "pair"}],"prim": "list"}],"prim": "pair"}],"prim": "list"}`,
	// },
	// // right-hand pair tree
	// {
	// 	Name: "right_hand_pair_tree",
	// 	Spec: `{"args":[{"annots":["%tokenPool"],"prim":"nat"},{"args":[{"annots":["%xtzPool"],"prim":"mutez"},{"args":[{"annots":["%lqtTotal"],"prim":"nat"},{"args":[{"annots":["%tokenAddress"],"prim":"address"},{"annots":["%lqtAddress"],"prim":"address"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"}`,
	// },

	// unsupported
	// //   chain_id
	// //   bls12_381_g1
	// //   bls12_381_g2
	// //   bls12_381_fr
	// {
	// 	Name: "bls",
	// 	Spec: `{"annots":["%g2"],"prim":"bls12_381_g2"}`,
	// },
	// //   sapling_state
	// {
	// 	Name: "sapling_state",
	// 	Spec: `{"prim":"sapling_state","args":[{"int":"8"}]}`,
	// },
	// //   sapling_transaction
	// {
	// 	Name: "sapling_transaction",
	// 	Spec: `{"prim":"sapling_transaction","args":[{"int":"8"}]}`,
	// },
	// //   never
	// {
	// 	Name: "never",
	// 	Spec: `{"prim":"never"}`,
	// },
	// // lambda, ticket, callbacks
	// {
	//     Name: "lambda",
	//     Spec: `{"args": [{"args": [{"args": [{"prim": "string"},{"prim": "bytes"}],"prim": "pair"},{"args": [{"prim": "bytes"},{"prim": "bytes"}],"prim": "big_map"}],"prim": "pair"},{"args": [{"args": [{"prim": "operation"}],"prim": "list"},{"args": [{"prim": "bytes"},{"prim": "bytes"}],"prim": "big_map"}],"prim": "pair"}],"prim": "lambda"}`,
	// },
	// // ticket
	// {
	//     Name: "ticket",
	//     Spec: `{"prim": "ticket", "args":[{"prim":"timestamp"}]}`,
	// },
	// // ticket 2
	// {
	//     Name: "ticket2",
	//     Spec: `{"prim": "ticket", "annots":["%save"], "args":[{"prim":"string"}]}`,
	// },
	// // contract
	// {
	//     Name: "contract",
	//     Spec: `{"annots": ["%pour_dest"],"args": [{"prim": "unit"}],"prim": "contract"}`,
	// },
}

func TestTypeMarshaling(t *testing.T) {
	for _, test := range marshalTests {
		t.Run(test.Name, func(T *testing.T) {
			typ := Type{}
			err := typ.UnmarshalJSON([]byte(test.Spec))
			if err != nil {
				T.Fatalf("unmarshal error: %v", err)
			}
			prim, err := typ.Typedef("").Marshal(test.Value, test.Optimized)
			if err != nil {
				T.Fatalf("marshal error: %v", err)
			}
			have, err := prim.MarshalJSON()
			if err != nil {
				T.Fatalf("render error: %v", err)
			}
			// fmt.Printf("HAVE: %s\n", string(have))
			if !jsonDiff(T, have, []byte(test.Want)) {
				T.Error("render mismatch, see log for details")
				t.FailNow()
			}
		})
	}
}
