// Copyright (c) 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//

package micheline

import (
	"strconv"
	"testing"
)

type typedefTest struct {
	Name string
	Spec string
	Want string
}

var typedefTests = []typedefTest{
	// scalars
	//   int
	{
		Name: "int",
		Spec: `{"annots": ["%payoutDelay"],"prim": "int"}`,
		Want: `{"name":"payoutDelay","path":[],"type":"int"}`,
	},
	//   nat
	{
		Name: "nat",
		Spec: `{"annots": ["%payoutFrequency"],"prim": "nat"}`,
		Want: `{"name":"payoutFrequency","path":[],"type":"nat"}`,
	},
	//   string
	{
		Name: "string",
		Spec: `{"annots": ["%name"],"prim": "string"}`,
		Want: `{"name":"name","path":[],"type":"string"}`,
	},
	//   bytes
	{
		Name: "bytes",
		Spec: `{"annots": ["%bakerName"],"prim": "bytes"}`,
		Want: `{"name":"bakerName","path":[],"type":"bytes"}`,
	},
	//   mutez
	{
		Name: "mutez",
		Spec: `{"annots": ["%signup_fee"],"prim": "mutez"}`,
		Want: `{"name":"signup_fee","path":[],"type":"mutez"}`,
	},
	//   bool
	{
		Name: "bool",
		Spec: `{"annots": ["%bakerChargesTransactionFee"],"prim": "bool"}`,
		Want: `{"name":"bakerChargesTransactionFee","path":[],"type":"bool"}`,
	},
	//   key_hash
	{
		Name: "key_hash",
		Spec: `{"annots": ["%baker"],"prim": "key_hash"}`,
		Want: `{"name":"baker","path":[],"type":"key_hash"}`,
	},
	//   timestamp
	{
		Name: "timestamp",
		Spec: `{"annots": ["%last_update"],"prim": "timestamp"}`,
		Want: `{"name":"last_update","path":[],"type":"timestamp"}`,
	},
	//   address
	{
		Name: "address",
		Spec: `{"annots": ["%reporterAccount"],"prim": "address"}`,
		Want: `{"name":"reporterAccount","path":[],"type":"address"}`,
	},
	//   key
	{
		Name: "key",
		Spec: `{"annots": ["%pour_authorizer"],"prim": "key"}`,
		Want: `{"name":"pour_authorizer","path":[],"type":"key"}`,
	},
	//   unit
	//   signature
	{
		Name: "signature",
		Spec: `{"args":[{"args":[{"prim":"nat"},{"args":[{"prim":"key"},{"prim":"signature"}],"prim": "pair"}],"prim": "pair"}],"prim": "pair"}`,
		Want: `{"name":"","path":[],"type":"struct","args":[{"name":"0","path":[0,0],"type":"nat"},{"name":"1","path":[0,1,0],"type":"key"},{"name":"2","path":[0,1,1],"type":"signature"}]}`,
	},
	//   chain_id
	//   bls12_381_g1
	//   bls12_381_g2
	//   bls12_381_fr
	{
		Name: "bls",
		Spec: `{"annots":["%g2"],"prim":"bls12_381_g2"}`,
		Want: `{"name":"g2","path":[],"type":"bls12_381_g2"}`,
	},
	//   sapling_state
	{
		Name: "sapling_state",
		Spec: `{"prim":"sapling_state","args":[{"int":"8"}]}`,
		Want: `{"name":"","path":[],"type":"sapling_state(8)"}`,
	},
	//   sapling_transaction
	{
		Name: "sapling_transaction",
		Spec: `{"prim":"sapling_transaction","args":[{"int":"8"}]}`,
		Want: `{"name":"","path":[],"type":"sapling_transaction(8)"}`,
	},
	//   never
	{
		Name: "never",
		Spec: `{"prim":"never"}`,
		Want: `{"name":"","path":[],"type":"never"}`,
	},
	// set
	{
		Name: "set",
		Spec: `{"annots": ["%admins"],"prim": "set", "args": [{"prim": "key_hash"}]}`,
		Want: `{"name":"admins","path":[],"type":"set","args":[{"name":"@item","path":[0],"type":"key_hash"}]}`,
	},
	// map
	{
		Name: "map",
		Spec: `{"annots":["%approvals"],"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]}`,
		Want: `{"name":"approvals","path":[],"type":"map","args":[{"name":"@key","path":[0],"type":"address"},{"name":"@value","path":[1],"type":"nat"}]}`,
	},
	// bigmap with scalar key
	// bigmap with pair key
	{
		Name: "bigmap",
		Spec: `{"annots": ["%ledger"],"args": [{"args": [{"prim": "address"},{"prim": "nat"}],"prim": "pair"},{"prim": "nat"}],"prim": "big_map"}`,
		Want: `{"name": "ledger", "path":[], "type": "big_map", "args":[{"name":"@key","path":[0],"type":"struct","args":[{"name":"0","path":[0,0],"type":"address"},{"name":"1","path":[0,1],"type":"nat"}]},{"name":"@value","path":[1],"type":"nat"}]}`,
	},
	// contract
	{
		Name: "contract",
		Spec: `{"annots": ["%pour_dest"],"args": [{"prim": "unit"}],"prim": "contract"}`,
		Want: `{"name":"pour_dest","path":[],"type":"contract","args":[{"name":"0","path":[0],"type":"unit"}]}`,
	},
	// lambda, list, operation
	{
		Name: "lambda",
		Spec: `{"args": [{"args": [{"args": [{"prim": "string"},{"prim": "bytes"}],"prim": "pair"},{"args": [{"prim": "bytes"},{"prim": "bytes"}],"prim": "big_map"}],"prim": "pair"},{"args": [{"args": [{"prim": "operation"}],"prim": "list"},{"args": [{"prim": "bytes"},{"prim": "bytes"}],"prim": "big_map"}],"prim": "pair"}],"prim": "lambda"}`,
		Want: `{"name":"","path":[],"type":"lambda","args":[{"name":"@param","path":[0],"type":"struct","args":[{"name":"0","path":[0,0,0],"type":"string"},{"name":"1","path":[0,0,1],"type":"bytes"},{"name":"2","path":[0,1],"type":"big_map","args":[{"name":"@key","path":[0],"type":"bytes"},{"name":"@value","path":[1],"type":"bytes"}]}]},{"name":"@return","path":[1],"type":"struct","args":[{"name":"0","path":[1,0],"type":"list","args":[{"name":"@item","path":[1,0,0],"type":"operation"}]},{"name":"1","path":[1,1],"type":"big_map","args":[{"name":"@key","path":[0],"type":"bytes"},{"name":"@value","path":[1],"type":"bytes"}]}]}]}`,
	},
	// ticket
	{
		Name: "ticket",
		Spec: `{"prim": "ticket", "args":[{"prim":"timestamp"}]}`,
		Want: `{"name":"","path":[],"type":"ticket","args":[{"name":"@value","path":[0],"type":"timestamp"}]}`,
	},
	// ticket 2
	{
		Name: "ticket2",
		Spec: `{"prim": "ticket", "annots":["%save"], "args":[{"prim":"string"}]}`,
		Want: `{"name":"save","path":[],"type":"ticket","args":[{"name":"@value","path":[0],"type":"string"}]}`,
	},
	// option
	{
		Name: "option",
		Spec: `{"annots":["%reporterAccount"],"prim":"option","args":[{"prim":"address"}]}`,
		Want: `{"name":"reporterAccount","path":[],"type":"address","optional":true}`,
	},
	// named union type
	{
		Name: "named-union",
		Spec: `{"args":[{"annots":["%do"],"args":[{"prim":"unit"},{"args":[{"prim":"operation"}],"prim":"list"}],"prim":"lambda"},{"annots":["%default"],"prim":"unit"}],"prim":"or"}`,
		Want: `{"name":"","path":[],"type":"union","args":[{"name":"do","path":[0],"type":"lambda","args":[{"name":"@param","path":[0],"type":"unit"},{"name":"@return","path":[1],"type":"list","args":[{"name":"@item","path":[1,0],"type":"operation"}]}]},{"name":"default","path":[1],"type":"unit"}]}`,
	},
	// anonymous union type
	{
		Name: "anon-union",
		Spec: `{"args":[{"args":[{"prim":"unit"},{"prim":"operation"}],"prim":"lambda"},{"args":[{"prim":"key_hash"}],"prim":"set"}],"prim":"or"}`,
		Want: `{"name":"","path":[],"type":"union","args":[{"name":"@or_0","path":[0],"type":"lambda","args":[{"name":"@param","path":[0],"type":"unit"},{"name":"@return","path":[1],"type":"operation"}]},{"name":"@or_1","path":[1],"type":"set","args":[{"name":"@item","path":[1,0],"type":"key_hash"}]}]}`,
	},
	// nested map
	{
		Name: "nested_map",
		Spec: `{"annots": ["%deck"],"args": [{"prim": "int"},{"args": [{"prim": "int"},{"prim": "int"}],"prim": "map"}],"prim": "map"}`,
		Want: `{"name":"deck","path":[],"type":"map","args":[{"name":"@key","path":[0],"type":"int"},{"name":"@value","path":[1],"type":"map","args":[{"name":"@key","path":[0],"type":"int"},{"name":"@value","path":[1],"type":"int"}]}]}`,
	},
	// nested list (FA2)
	{
		Name: "nested_list",
		Spec: `{"annots": ["%transfer"],"args": [{"args": [{"annots": ["%from_"],"prim": "address"},{"annots": ["%txs"],"args": [{"args": [{"annots": ["%to_"],"prim": "address"},{"args": [{"annots": ["%token_id"],"prim": "nat"},{"annots": ["%amount"],"prim": "nat"}],"prim": "pair"}],"prim": "pair"}],"prim": "list"}],"prim": "pair"}],"prim": "list"}`,
		Want: `{"name":"transfer","path":[],"type":"list","args":[{"name":"@item","path":[0],"type":"struct","args":[{"name":"from_","path":[0,0],"type":"address"},{"name":"txs","path":[0,1],"type":"list","args":[{"name":"@item","path":[0,1,0],"type":"struct","args":[{"name":"to_","path":[0,1,0,0],"type":"address"},{"name":"token_id","path":[0,1,0,1,0],"type":"nat"},{"name":"amount","path":[0,1,0,1,1],"type":"nat"}]}]}]}]}`,
	},
	// right-hand pair tree
	{
		Name: "right_hand_pair_tree",
		Spec: `{"args":[{"annots":["%tokenPool"],"prim":"nat"},{"args":[{"annots":["%xtzPool"],"prim":"mutez"},{"args":[{"annots":["%lqtTotal"],"prim":"nat"},{"args":[{"annots":["%tokenAddress"],"prim":"address"},{"annots":["%lqtAddress"],"prim":"address"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"}`,
		Want: `{"name":"","path":[],"type":"struct","args":[{"name":"tokenPool","path":[0],"type":"nat"},{"name":"xtzPool","path":[1,0],"type":"mutez"},{"name":"lqtTotal","path":[1,1,0],"type":"nat"},{"name":"tokenAddress","path":[1,1,1,0],"type":"address"},{"name":"lqtAddress","path":[1,1,1,1],"type":"address"}]}`,
	},
}

func TestTypeRendering(t *testing.T) {
	for _, test := range typedefTests {
		t.Run(test.Name, func(T *testing.T) {
			prim := Prim{}
			err := prim.UnmarshalJSON([]byte(test.Spec))
			if err != nil {
				T.Errorf("unmarshal error: %v", err)
			}
			have, err := Type{prim}.MarshalJSON()
			if err != nil {
				T.Errorf("render error: %v", err)
			}
			if !jsonDiff(T, have, []byte(test.Want)) {
				T.Error("render mismatch, see log for details")
				t.FailNow()
			}
		})
	}
}

type interfaceTest struct {
	Name   string
	Type   string
	Value  string
	Expect bool
}

const (
	fa1TransferType = `{"prim":"pair","args":[{"prim":"address","annots":[":from"]},{"prim":"pair","args":[{"prim":"address","annots":[":to"]},{"prim":"nat","annots":[":value"]}]}]}`
	fa2TransferType = `{"prim":"list","annots":["%transfer"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%from_"]},{"prim":"list","annots":["%txs"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%to_"]},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}]}]}]}]}]}`
)

var interfaceTests = []interfaceTest{
	// FA1
	{
		Name:   "fa1",
		Type:   fa1TransferType,
		Value:  `{"prim":"Pair","args":[{"bytes":"019c0931f0ebac06db1063abe651e773db6c353ce900"},{"prim":"Pair","args":[{"bytes": "0000c2bb77ac9a2c86ca05fdaea1888408471b9c1468"},{"int":"740596954"}]}]}`,
		Expect: true,
	},
	{
		Name:   "fa1_wrong_value",
		Type:   fa1TransferType,
		Value:  `[{"prim":"Pair","args":[{"string":"tz1bq5QazD43hBGxYxX8mv4sa1r1kz5sRGWz"},{"string":"tz1QgLtW1kJehxYhtypPyWJcutUTY6xZfDQf"},{"int":"21367500000"}]},{"prim":"Pair","args":[{"string":"tz1bq5QazD43hBGxYxX8mv4sa1r1kz5sRGWz"},{"string":"tz1QgLtW1kJehxYhtypPyWJcutUTY6xZfDQf"},{"int":"21367500000"}]}]`,
		Expect: false,
	},
	// FA2 single (nocomb)
	{
		Name:   "fa2_single_recv",
		Type:   fa2TransferType,
		Value:  `[{"prim":"Pair","args":[{"bytes":"01b39686f116bb35115559f7e781200850e02854c400"},[{"prim":"Pair","args":[{"bytes":"0000f0ddca1cdfa0c48c92d162f3f72b8144ee2045ba"},{"prim":"Pair","args":[{"int":"0"},{"int":"1027681"}]}]}]]}]`,
		Expect: true,
	},
	// FA2 multi (nocomb)
	{
		Name:   "fa2_multi_recv",
		Type:   fa2TransferType,
		Value:  `[{"prim":"Pair","args":[{"bytes":"01b39686f116bb35115559f7e781200850e02854c400"},[{"prim":"Pair","args":[{"bytes":"0000f0ddca1cdfa0c48c92d162f3f72b8144ee2045ba"},{"prim":"Pair","args":[{"int":"0"},{"int":"1027681"}]}]},{"prim":"Pair","args":[{"bytes":"0000f0ddca1cdfa0c48c92d162f3f72b8144ee2045ba"},{"prim":"Pair","args":[{"int":"0"},{"int":"1027681"}]}]}]]}]`,
		Expect: true,
	},
	// FA2 single (comb)
	{
		Name:   "fa2_multi_recv_comb",
		Type:   fa2TransferType,
		Value:  `[{"prim":"Pair","args":[{"bytes":"01b39686f116bb35115559f7e781200850e02854c400"},[{"prim":"Pair","args":[{"bytes":"0000f0ddca1cdfa0c48c92d162f3f72b8144ee2045ba"},{"int":"0"},{"int":"1027681"}]}]]}]`,
		Expect: true,
	},
	// TODO
	// union type
	// optional flag
	// set
	// map
	// bigmap
	// lambda
	// ticket
	// sapling
}

func TestInterfaceCheck(t *testing.T) {
	for _, test := range interfaceTests {
		t.Run(test.Name, func(T *testing.T) {
			var typ Type
			if err := typ.Prim.UnmarshalJSON([]byte(test.Type)); err != nil {
				T.Fatalf("unmarshal type: %v", err)
			}
			var val Prim
			if err := val.UnmarshalJSON([]byte(test.Value)); err != nil {
				T.Fatalf("unmarshal value: %v", err)
			}
			if have, want := val.Implements(typ), test.Expect; have != want {
				T.Errorf("mismatch want=%t have=%t", want, have)
			}
		})
	}
}

type bigmapDetectTest struct {
	Name         string
	Type         string
	Value        string
	Expect       map[string]int64
	SkipTypetest bool
}

var bigmapDetectTests = []bigmapDetectTest{
	{
		Name:   "HEN",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%administrator"]},{"prim":"pair","args":[{"prim":"nat","annots":["%all_tokens"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"big_map","annots":["%operators"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"pair","args":[{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]}]},{"prim":"unit"}]}]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1Hkg5qeNhfwpKW4fXvq7HGZB9z2EnmCCA9"},{"int":"800633"},{"int":"511"}]},{"prim":"Pair","args":[{"int":"512"},{"int":"513"}]},{"prim":"False"},{"int":"514"}]}`,
		Expect: map[string]int64{"ledger": 511, "metadata": 512, "operators": 513, "token_metadata": 514},
	},
	{
		Name:   "TzTacos",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%base_timestamp"]},{"prim":"nat","annots":["%curr_id"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"nat"},{"prim":"address"}]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%operators"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit"}]},{"prim":"mutez","annots":["%price"]}]},{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"2020-06-09T00:00:01Z"},{"int":"21224"}]},{"int":"808"},{"int":"809"}]},{"prim":"Pair","args":[{"int":"810"},{"int":"2122400"}]},{"int":"811"}]}`,
		Expect: map[string]int64{"ledger": 808, "metadata": 809, "operators": 810, "token_metadata": 811},
	},
	{
		Name:   "USDtz",
		Type:   `{"prim":"pair","args":[{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address","annots":[":user"]},{"prim":"pair","args":[{"prim":"nat","annots":[":balance"]},{"prim":"map","annots":[":approvals"],"args":[{"prim":"address","annots":[":spender"]},{"prim":"nat","annots":[":value"]}]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"nat","annots":["%totalSupply"]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"int":"36"},{"string":"KT1DTxmuuEARtynHP2r6EX14PieaPmTu1vxF"},{"prim":"False"},{"int":"516370330738"}]}`,
		Expect: map[string]int64{"ledger": 36},
	},
	{
		Name:   "Ramp",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%administrator"]},{"prim":"pair","args":[{"prim":"big_map","annots":["%balances"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"map","annots":["%approvals"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%balance"]}]}]},{"prim":"option","annots":["%operator_contract"],"args":[{"prim":"address"}]}]}]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"pair","args":[{"prim":"nat","annots":["%totalSupply"]},{"prim":"big_map","annots":["%whitelist"],"args":[{"prim":"address"},{"prim":"nat"}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1MFdUjp818c96671eBzfCPnwwBh7eGaSj1"},{"int":"329"},{"prim":"Some","args":[{"string":"KT1VgXHLXRgh6J5iGw4zkk7vUfjLoPhRnt9L"}]}]},{"prim":"False"},{"int":"3343000000"},{"int":"330"}]}`,
		Expect: map[string]int64{"balances": 329, "whitelist": 330},
	},
	{
		Name:   "QLKUSD",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%balances"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"map","annots":["%approvals"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%balance"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%governorAddress"]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%ovenRegistryAddress"]},{"prim":"address","annots":["%quipuswapAddress"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rewardChangeAllowedLevel"]},{"prim":"nat","annots":["%rewardPercent"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%savedState_depositor"],"args":[{"prim":"address"}]},{"prim":"option","annots":["%savedState_redeemer"],"args":[{"prim":"address"}]}]},{"prim":"pair","args":[{"prim":"option","annots":["%savedState_tokensToDeposit"],"args":[{"prim":"nat"}]},{"prim":"option","annots":["%savedState_tokensToRedeem"],"args":[{"prim":"nat"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"int","annots":["%state"]},{"prim":"address","annots":["%tokenAddress"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat"},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"nat","annots":["%totalSupply"]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1600"},{"string":"KT1FC288PPN87gtmTKRuHVUhoJ1Z8T4KBUAD"},{"int":"1601"}]},{"prim":"Pair","args":[{"string":"KT1Ldn1XWQmk7J4pYgGFjjwV57Ew8NYvcNtJ"},{"string":"tz1Ke2h7sDdakHJQh8WX4Z372du1KChsksyU"}]},{"int":"1522673"},{"int":"10"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"None"},{"prim":"None"}]},{"prim":"None"},{"prim":"None"}]},{"prim":"Pair","args":[{"int":"0"},{"string":"KT1K9gCRgaLRFKTErYt1wVxA3Frb9FjasjTV"}]},{"int":"1602"},{"int":"27974308647677254253603734093909520253599"}]}`,
		Expect: map[string]int64{"balances": 1600, "metadata": 1601, "token_metadata": 1602},
	},
	{
		Name:   "Werecoin",
		Type:   `{"prim":"pair","args":[{"prim":"nat","annots":["%onetoken"]},{"prim":"pair","args":[{"prim":"big_map","annots":["%allowance"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"address"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"int":"1000000"},{"int":"130"},{"int":"131"},{"int":"132"}]}`,
		Expect: map[string]int64{"allowance": 130, "ledger": 131, "metadata": 132},
	},
	{
		Name:   "sCasino",
		Type:   `{"prim":"pair","args":[{"prim":"pair","annots":["%assets"],"args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"big_map","annots":["%operators"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit"}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"big_map","annots":["%token_total_supply"],"args":[{"prim":"nat"},{"prim":"nat"}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"option","annots":["%pauseable_admin"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"bool","annots":["%paused"]}]},{"prim":"option","annots":["%pending_admin"],"args":[{"prim":"address"}]}]}]},{"prim":"pair","args":[{"prim":"pair","annots":["%permit_currency"],"args":[{"prim":"address","annots":["%fa2_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","args":[{"prim":"map","annots":["%active_game_info"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","annots":["%game_currency"],"args":[{"prim":"address","annots":["%fa2_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","args":[{"prim":"or","annots":["%stage"],"args":[{"prim":"or","args":[{"prim":"unit","annots":["%game_Resolved"]},{"prim":"unit","annots":["%house_Committed"]}]},{"prim":"unit","annots":["%maker_Committed"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%maker"]},{"prim":"pair","args":[{"prim":"bytes","annots":["%maker_hash"]},{"prim":"pair","args":[{"prim":"option","annots":["%house_hash"],"args":[{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"option","annots":["%maker_secret"],"args":[{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%bet"]},{"prim":"pair","args":[{"prim":"nat","annots":["%guess"]},{"prim":"pair","args":[{"prim":"nat","annots":["%payout"]},{"prim":"pair","args":[{"prim":"option","annots":["%ended"],"args":[{"prim":"timestamp"}]},{"prim":"option","annots":["%winner"],"args":[{"prim":"bool"}]}]}]}]}]}]}]}]}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%resolved_game_info"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","annots":["%game_currency"],"args":[{"prim":"address","annots":["%fa2_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","args":[{"prim":"or","annots":["%stage"],"args":[{"prim":"or","args":[{"prim":"unit","annots":["%game_Resolved"]},{"prim":"unit","annots":["%house_Committed"]}]},{"prim":"unit","annots":["%maker_Committed"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%maker"]},{"prim":"pair","args":[{"prim":"bytes","annots":["%maker_hash"]},{"prim":"pair","args":[{"prim":"option","annots":["%house_hash"],"args":[{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"option","annots":["%maker_secret"],"args":[{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%start_time"]},{"prim":"pair","args":[{"prim":"nat","annots":["%bet"]},{"prim":"pair","args":[{"prim":"nat","annots":["%guess"]},{"prim":"pair","args":[{"prim":"nat","annots":["%payout"]},{"prim":"pair","args":[{"prim":"option","annots":["%ended"],"args":[{"prim":"timestamp"}]},{"prim":"option","annots":["%winner"],"args":[{"prim":"bool"}]}]}]}]}]}]}]}]}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%bankroll_ledger"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%fa2_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","args":[{"prim":"pair","annots":["%game_currency"],"args":[{"prim":"address","annots":["%fa2_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","args":[{"prim":"pair","annots":["%share_currency"],"args":[{"prim":"address","annots":["%fa2_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%total_bankroll"]},{"prim":"nat","annots":["%total_shares_issued"]}]}]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%current_game_id"]},{"prim":"pair","args":[{"prim":"nat","annots":["%next_share_token_id"]},{"prim":"pair","args":[{"prim":"address","annots":["%house_wallet"]},{"prim":"pair","args":[{"prim":"address","annots":["%burn_wallet"]},{"prim":"pair","args":[{"prim":"nat","annots":["%max_bet"]},{"prim":"pair","args":[{"prim":"big_map","annots":["%edge_permits"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%low_edge"]},{"prim":"pair","args":[{"prim":"nat","annots":["%high_edge"]},{"prim":"pair","args":[{"prim":"bool","annots":["%hh"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%hh_start_time"]},{"prim":"pair","args":[{"prim":"int","annots":["%hh_time"]},{"prim":"int","annots":["%time_between_hh"]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"3908"},{"int":"3909"}]},{"int":"3910"},{"int":"3911"}]},{"int":"3912"},{"prim":"Some","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1YFxWGfE7K8wQkKBVerB21HbNEiLpA2ch9"},{"prim":"True"}]},{"prim":"None"}]}]},{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},[{"prim":"Elt","args":[{"int":"456"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1QfdfpmTbBn8kQqp7BTibYRGGC6cMPw8Wi"},{"bytes":"0d41c3cdf4c42672d38bb2adf3406b5767c0f845830a90e2e3ee0e28c614c835"},{"prim":"Some","args":[{"bytes":"f4c0702144423e9b3b40e1aed13679c7bbe7ee9366f3b905e0d54c5a394436ff"}]},{"prim":"None"},{"string":"2021-05-31T03:35:18Z"},{"int":"20"},{"int":"2"},{"int":"38"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"536"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1M2Ws52krJrwJi1ZFsmVfazBiafWYKZTvd"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hYc8FKJSztPJb8a9b4V4yQBGtk9t1FkEj"},{"bytes":"d3a6ad0bed2212638e05134d269354db99a510cb6ed47e562dbb769b3c9bfc38"},{"prim":"Some","args":[{"bytes":"197119b5f4e4260fcbc6fe344c6b4d16ade935b4ee9e7c6deadf87e3b077345e"}]},{"prim":"None"},{"string":"2021-06-01T22:13:38Z"},{"int":"10000000"},{"int":"3"},{"int":"10000000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"537"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1M2Ws52krJrwJi1ZFsmVfazBiafWYKZTvd"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hYc8FKJSztPJb8a9b4V4yQBGtk9t1FkEj"},{"bytes":"f19100e36e6781ba8eb9384a81d9f68875215b3724818caa608a5aa7c21e78c0"},{"prim":"Some","args":[{"bytes":"3b5541226e623939ffd422024d954557bfe2a87baba54716ec4856027965532c"}]},{"prim":"None"},{"string":"2021-06-01T22:17:38Z"},{"int":"10000000"},{"int":"3"},{"int":"10000000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"538"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1M2Ws52krJrwJi1ZFsmVfazBiafWYKZTvd"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hYc8FKJSztPJb8a9b4V4yQBGtk9t1FkEj"},{"bytes":"4b3487dd82c3f68f9735f86d6aa486cb46bac79e8cd1cd6eb8eae69001766b87"},{"prim":"Some","args":[{"bytes":"b413732269f28358e91599c58da889de5c5d3b94f86026cbf3d48f0e8f0e2194"}]},{"prim":"None"},{"string":"2021-06-01T22:17:38Z"},{"int":"10000000"},{"int":"3"},{"int":"10000000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"544"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1M2Ws52krJrwJi1ZFsmVfazBiafWYKZTvd"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hYc8FKJSztPJb8a9b4V4yQBGtk9t1FkEj"},{"bytes":"c8067c638657143a4112f12b3084779cbd2f31e414d3f99f7de314a1e001ba5e"},{"prim":"Some","args":[{"bytes":"d1c23f30c4d3f6bdd9158200ae175bcaaa41945f04d07c01eaae10e995f99240"}]},{"prim":"None"},{"string":"2021-06-02T02:59:06Z"},{"int":"10000000"},{"int":"3"},{"int":"10000000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"545"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1M2Ws52krJrwJi1ZFsmVfazBiafWYKZTvd"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hYc8FKJSztPJb8a9b4V4yQBGtk9t1FkEj"},{"bytes":"c8067c638657143a4112f12b3084779cbd2f31e414d3f99f7de314a1e001ba5e"},{"prim":"Some","args":[{"bytes":"33e77eb4068f055f0ae257a7dbb4477bb2e74f6b3692863bd2aeee435098c4d1"}]},{"prim":"None"},{"string":"2021-06-02T03:03:06Z"},{"int":"10000000"},{"int":"3"},{"int":"10000000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"546"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1M2Ws52krJrwJi1ZFsmVfazBiafWYKZTvd"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hYc8FKJSztPJb8a9b4V4yQBGtk9t1FkEj"},{"bytes":"c8067c638657143a4112f12b3084779cbd2f31e414d3f99f7de314a1e001ba5e"},{"prim":"Some","args":[{"bytes":"832b7dae019c23320bf51b4233f2b395749b5f4caa362b1d5559e2987f67a617"}]},{"prim":"None"},{"string":"2021-06-02T03:05:06Z"},{"int":"10000000"},{"int":"3"},{"int":"10000000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"547"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1M2Ws52krJrwJi1ZFsmVfazBiafWYKZTvd"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hYc8FKJSztPJb8a9b4V4yQBGtk9t1FkEj"},{"bytes":"c8067c638657143a4112f12b3084779cbd2f31e414d3f99f7de314a1e001ba5e"},{"prim":"Some","args":[{"bytes":"52d337577be86aa48574d188944fbbe8ef5d729eaae7debfae05eadb71fedd07"}]},{"prim":"None"},{"string":"2021-06-02T03:14:46Z"},{"int":"10000000"},{"int":"3"},{"int":"10000000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"560"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VEe2PES4F9nqsgQHdonPsFkTvKXLbbdha"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz2H4E4JizKxqZYJ39Z9B9JHHmVwucTib2ZM"},{"bytes":"fa6395d1a21479c93dd6586aec5c27f09d6d5d6593602a963924246573629250"},{"prim":"Some","args":[{"bytes":"7d53455cb1a2c8b53fba9d4112b8c85d04500bc8c1c000845ff1460cdf1b381b"}]},{"prim":"None"},{"string":"2021-06-02T22:54:46Z"},{"int":"100000"},{"int":"5"},{"int":"17600"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"620"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1ZW1Z6YqZKjeCbb5LVTWJ7UaSTooXmMajq"},{"bytes":"80b032de77abbe46ba0f6f206a2e35b5b6cab850356019ee64a1ea58c6f9879b"},{"prim":"Some","args":[{"bytes":"598304b29ffe0e8f6ef8de7ef0ada7411c5185cb952e4de11349c62174017d10"}]},{"prim":"None"},{"string":"2021-06-03T18:10:58Z"},{"int":"100"},{"int":"3"},{"int":"100"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"652"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VEe2PES4F9nqsgQHdonPsFkTvKXLbbdha"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1TVZgy4rejPUkEF3bsxLGNCp6H4gz4PaGf"},{"bytes":"4298ee7fb942f132fcbea8437ba4cb64c40cb42440cb46af81f8cd4b459f5853"},{"prim":"Some","args":[{"bytes":"d837efccce572d477d1bbd3dcb59fe65dae8a63908341b8828d49fc53b31f7fa"}]},{"prim":"None"},{"string":"2021-06-04T07:59:02Z"},{"int":"1000000"},{"int":"1"},{"int":"4880000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"677"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1Cjx8hYwzaCAke6rLWoZBLp8w89VeAduAR"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz2MRhfAKUyNZDKH8XBzmXyCwd2h5ySbgZJJ"},{"bytes":"e743e1884a144c06e76a87c1804b70a836803bcc9c5442030b07693534539a52"},{"prim":"Some","args":[{"bytes":"e5ad2891ad674744617bad1918ac3fafb5700a0eb9567c5c90511522986b3062"}]},{"prim":"None"},{"string":"2021-06-05T18:59:54Z"},{"int":"316500000"},{"int":"1"},{"int":"1582500000"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"704"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1ZW1Z6YqZKjeCbb5LVTWJ7UaSTooXmMajq"},{"bytes":"7d1dabc4fe8ebc99b85f5df6a964d5eac6653545a356ab36f6300fafcaf94bde"},{"prim":"Some","args":[{"bytes":"ff444665e29b0bfb86a639cc97cc4a237342c16595a7e88d55659401de1ef295"}]},{"prim":"None"},{"string":"2021-06-06T20:57:06Z"},{"int":"1000"},{"int":"3"},{"int":"980"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"806"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz2CYbLMDbiG54yz8Z5gNd61ZA2x6XJUZ1Gb"},{"bytes":"e3c3488df1f8c87b872c758a054ddfa4f211d8c0e6cdaddba0b6ee16d7f18e18"},{"prim":"Some","args":[{"bytes":"a45851868f90ef28e2bd01e5d79a4a53de0f33ce619c83f254bbdfee15fbdda5"}]},{"prim":"None"},{"string":"2021-06-09T19:29:34Z"},{"int":"2000"},{"int":"2"},{"int":"3940"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"1462"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1hFpyGJ8t6bthUQ8wpXiAP3EnGLtkKJme9"},{"bytes":"e33a6ee621b666bfdaebd85640a2c83d136b86840d29a6d898d7389bfb9cff89"},{"prim":"Some","args":[{"bytes":"fcaecc08d2bb60c57d75f58f3025d054dbb5eeb53a781789bab74a27e8e2eb7f"}]},{"prim":"None"},{"string":"2021-06-18T15:24:46Z"},{"int":"710"},{"int":"2"},{"int":"1398"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"2759"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1fwd9Ub1AwXaRNLUfdDCtQfWyeq8YcJvCV"},{"bytes":"7c0b55d7f56fb662f1739205a09ba0d75fd6aa744fe372daedcea7eb9a9f522e"},{"prim":"Some","args":[{"bytes":"c3b9b0ec2c4d29f0e0821132a895a19fe1a8b05dd6513fc4730ee198361a8356"}]},{"prim":"None"},{"string":"2021-08-12T02:04:40Z"},{"int":"29442"},{"int":"3"},{"int":"28853"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"2768"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1cjUnZmKsfHNqAXarvgvHcd6WZY4U3WE7K"},{"bytes":"4f9eaf57c5118481136ba695484038c1f5ca2de01bd2aff339a2e5a89049a52f"},{"prim":"Some","args":[{"bytes":"9ae902cfd41391814d258bebfd38278d58aef24779f9e9920330f09a93a174ed"}]},{"prim":"None"},{"string":"2021-08-12T02:04:40Z"},{"int":"148"},{"int":"2"},{"int":"291"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"2797"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1NUUa7iqcTMRmepuGiKvj88tcq6TnKHBBX"},{"bytes":"d862978e2369fd1415d07f3c9844017bf372a7197e112467e1a1df22fb904748"},{"prim":"Some","args":[{"bytes":"757b212eed981a7fd34ac136688dd1668e2d2ae23ee06b1ac8655d88d33e15b1"}]},{"prim":"None"},{"string":"2021-08-23T18:41:06Z"},{"int":"10"},{"int":"3"},{"int":"9"},{"prim":"None"},{"prim":"None"}]}]},{"prim":"Elt","args":[{"int":"2811"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1VHd7ysjnvxEzwtjBAmYAmasvVCfPpSkiG"},{"int":"0"}]},{"prim":"Left","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"string":"tz1XwZGo4xbE71PqMziQ2ZdsKEskJt1VJw79"},{"bytes":"610275dc4feb45aa8885f2ec14a77325515ccd88441ebd7afdac20ea81d6c660"},{"prim":"Some","args":[{"bytes":"66f11bf3ea4028c9c40970daaa602acf7cb465ba36226884e96d069dd1b6cb4e"}]},{"prim":"None"},{"string":"2021-09-02T06:00:58Z"},{"int":"10000"},{"int":"3"},{"int":"9600"},{"prim":"None"},{"prim":"None"}]}]}],{"int":"3913"},{"int":"3914"},{"int":"2962"},{"int":"27"},{"string":"tz1YFxWGfE7K8wQkKBVerB21HbNEiLpA2ch9"},{"string":"KT1HTfs9vb1TgnLZCJwXSDNgw1dg4mK4bCSs"},{"int":"2"},{"int":"3915"},{"int":"594"},{"int":"588"},{"prim":"False"},{"string":"2022-05-22T05:19:59Z"},{"int":"3600"},{"int":"43200"}]}`,
		Expect: map[string]int64{"bankroll_ledger": 3914, "edge_permits": 3915, "ledger": 3908, "metadata": 3912, "operators": 3909, "resolved_game_info": 3913, "token_metadata": 3910, "token_total_supply": 3911},
	},
	{
		Name:   "QUIPU",
		Type:   `{"prim":"pair","args":[{"prim":"big_map","annots":["%account_info"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"map","annots":["%balances"],"args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"set","annots":["%allowances"],"args":[{"prim":"address"}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%token_info"],"args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"map","annots":["%minters_info"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%last_token_id"]},{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"nat","annots":["%permit_counter"]},{"prim":"pair","args":[{"prim":"big_map","annots":["%permits"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"option","annots":["%expiry"],"args":[{"prim":"nat"}]},{"prim":"map","annots":["%permits"],"args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"timestamp","annots":["%created_at"]},{"prim":"option","annots":["%expiry"],"args":[{"prim":"nat"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%default_expiry"]},{"prim":"nat","annots":["%total_minter_shares"]}]}]}]}]}]}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"int":"12043"},{"int":"12044"},{"int":"12045"},{"int":"12046"},[{"prim":"Elt","args":[{"string":"tz1RAqmz7uwftjrrSP6b97x8z1zMY3eEX87R"},{"int":"71000000"}]},{"prim":"Elt","args":[{"string":"tz1RieztE7wgV1M2qxxbHPxL6b1cm8fARZH2"},{"int":"5000000"}]},{"prim":"Elt","args":[{"string":"tz1TBQxAct5ij3Dg34PKwkhR9mibTdd2Rdqg"},{"int":"10000000"}]},{"prim":"Elt","args":[{"string":"tz1Zu5JJV5p9RBK6BL48UwYVtgC2ASUqFT4i"},{"int":"500000"}]},{"prim":"Elt","args":[{"string":"tz1fttB1fAdqmM5wSy9uBndErYtyCK6teBBd"},{"int":"13500000"}]}],{"int":"1"},{"string":"tz1be39XdGQhLFCfATabBr5yLYMGhmFSbdQv"},{"int":"0"},{"int":"12047"},{"int":"1000"},{"int":"100000000"}]}`,
		Expect: map[string]int64{"account_info": 12043, "metadata": 12045, "permits": 12047, "token_info": 12044, "token_metadata": 12046},
	},
	{
		Name:         "AKA-Royalties",
		SkipTypetest: true,
		Type:         `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%akaMinter"]},{"prim":"address","annots":["%akaNFTContract"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%constant_royalties"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"address","annots":["%creator"]},{"prim":"pair","args":[{"prim":"map","annots":["%royalties"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%total_royalties"]}]}]}]},{"prim":"big_map","annots":["%get_royalty_list"],"args":[{"prim":"address"},{"prim":"unit"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%manager"]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"map","annots":["%royalties"],"args":[{"prim":"address"},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"address","annots":["%creator"]},{"prim":"pair","args":[{"prim":"map","annots":["%royalties"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%total_royalties"]}]}]}]}]},{"prim":"big_map","annots":["%royalties_updater"],"args":[{"prim":"address"},{"prim":"unit"}]}]}]}]}`,
		Value:        `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"KT1ULea6kxqiYe1A7CZVfMuGmTx7NmDGAph1"},{"string":"KT1AFq5XorPduoYyWxs5gEyrFK6fVjJVbtCj"}]},{"int":"55654"},{"int":"55655"}]},{"prim":"Pair","args":[{"string":"tz1WCYsbPyHTBcnj4saWG6SRFHECCj2TTzC6"},{"int":"55656"}]},[{"prim":"Elt","args":[{"string":"KT1AFq5XorPduoYyWxs5gEyrFK6fVjJVbtCj"},{"int":"55678"}]},{"prim":"Elt","args":[{"string":"KT1DEwdmXvjbdCz3HcehrYGiV46rVAwDiYVk"},{"int":"292144"}]},{"prim":"Elt","args":[{"string":"KT1KEa8z6vWXDJrVqtMrAeDVzsvxat3kHaCE"},{"int":"259052"}]},{"prim":"Elt","args":[{"string":"KT1MYSapB87YGSm1zxN3pbBGWDxea9YCkPH8"},{"int":"300594"}]},{"prim":"Elt","args":[{"string":"KT1RJ6PbjHpwc3M5rw5s2Nbmefwbuwbdxton"},{"int":"55710"}]},{"prim":"Elt","args":[{"string":"KT1U6EHmNxJTkvaWJ4ThczG4FSDaHC21ssvi"},{"int":"259051"}]}],{"int":"55657"}]}`,
		Expect:       map[string]int64{"royalties_updater": 55657, "constant_royalties": 55654, "get_royalty_list": 55655, "metadata": 55656, "KT1AFq5XorPduoYyWxs5gEyrFK6fVjJVbtCj": 55678, "KT1RJ6PbjHpwc3M5rw5s2Nbmefwbuwbdxton": 55710, "KT1U6EHmNxJTkvaWJ4ThczG4FSDaHC21ssvi": 259051, "KT1KEa8z6vWXDJrVqtMrAeDVzsvxat3kHaCE": 259052, "KT1DEwdmXvjbdCz3HcehrYGiV46rVAwDiYVk": 292144, "KT1MYSapB87YGSm1zxN3pbBGWDxea9YCkPH8": 300594},
	},
	{
		Name:   "QuipuLP",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%dex_lambdas"],"args":[{"prim":"nat"},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","annots":["%divestLiquidity"],"args":[{"prim":"pair","args":[{"prim":"nat","annots":["%min_tez"]},{"prim":"nat","annots":["%min_tokens"]}]},{"prim":"nat","annots":["%shares"]}]},{"prim":"nat","annots":["%initializeExchange"]}]},{"prim":"or","args":[{"prim":"nat","annots":["%investLiquidity"]},{"prim":"pair","annots":["%tezToTokenPayment"],"args":[{"prim":"nat","annots":["%min_out"]},{"prim":"address","annots":["%receiver"]}]}]}]},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","annots":["%tokenToTezPayment"],"args":[{"prim":"pair","args":[{"prim":"nat","annots":["%amount"]},{"prim":"nat","annots":["%min_out"]}]},{"prim":"address","annots":["%receiver"]}]},{"prim":"pair","annots":["%veto"],"args":[{"prim":"nat","annots":["%value"]},{"prim":"address","annots":["%voter"]}]}]},{"prim":"or","args":[{"prim":"pair","annots":["%vote"],"args":[{"prim":"pair","args":[{"prim":"key_hash","annots":["%candidate"]},{"prim":"nat","annots":["%value"]}]},{"prim":"address","annots":["%voter"]}]},{"prim":"address","annots":["%withdrawProfit"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%current_candidate"],"args":[{"prim":"key_hash"}]},{"prim":"option","annots":["%current_delegated"],"args":[{"prim":"key_hash"}]}]},{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_update_time"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_veto"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","annots":["%allowances"],"args":[{"prim":"address"}]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}]}]},{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%reward_paid"]},{"prim":"nat","annots":["%reward_per_sec"]}]},{"prim":"nat","annots":["%reward_per_share"]},{"prim":"nat","annots":["%tez_pool"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%total_supply"]},{"prim":"nat","annots":["%total_votes"]}]},{"prim":"big_map","annots":["%user_rewards"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"nat","annots":["%veto"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%vetos"],"args":[{"prim":"key_hash"},{"prim":"timestamp"}]},{"prim":"big_map","annots":["%voters"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%candidate"],"args":[{"prim":"key_hash"}]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}]},{"prim":"big_map","annots":["%votes"],"args":[{"prim":"key_hash"},{"prim":"nat"}]}]},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%current_candidate"],"args":[{"prim":"key_hash"}]},{"prim":"option","annots":["%current_delegated"],"args":[{"prim":"key_hash"}]}]},{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_update_time"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_veto"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","annots":["%allowances"],"args":[{"prim":"address"}]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}]}]},{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%reward_paid"]},{"prim":"nat","annots":["%reward_per_sec"]}]},{"prim":"nat","annots":["%reward_per_share"]},{"prim":"nat","annots":["%tez_pool"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%total_supply"]},{"prim":"nat","annots":["%total_votes"]}]},{"prim":"big_map","annots":["%user_rewards"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"nat","annots":["%veto"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%vetos"],"args":[{"prim":"key_hash"},{"prim":"timestamp"}]},{"prim":"big_map","annots":["%voters"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%candidate"],"args":[{"prim":"key_hash"}]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}]},{"prim":"big_map","annots":["%votes"],"args":[{"prim":"key_hash"},{"prim":"nat"}]}]}]}]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","annots":["%storage"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%current_candidate"],"args":[{"prim":"key_hash"}]},{"prim":"option","annots":["%current_delegated"],"args":[{"prim":"key_hash"}]}]},{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_update_time"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_veto"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","annots":["%allowances"],"args":[{"prim":"address"}]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}]}]},{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%reward_paid"]},{"prim":"nat","annots":["%reward_per_sec"]}]},{"prim":"nat","annots":["%reward_per_share"]},{"prim":"nat","annots":["%tez_pool"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%total_supply"]},{"prim":"nat","annots":["%total_votes"]}]},{"prim":"big_map","annots":["%user_rewards"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"nat","annots":["%veto"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%vetos"],"args":[{"prim":"key_hash"},{"prim":"timestamp"}]},{"prim":"big_map","annots":["%voters"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%candidate"],"args":[{"prim":"key_hash"}]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}]},{"prim":"big_map","annots":["%votes"],"args":[{"prim":"key_hash"},{"prim":"nat"}]}]},{"prim":"big_map","annots":["%token_lambdas"],"args":[{"prim":"nat"},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","annots":["%iBalance_of"],"args":[{"prim":"list","annots":["%requests"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}]}]},{"prim":"contract","annots":["%callback"],"args":[{"prim":"list","args":[{"prim":"pair","args":[{"prim":"pair","annots":["%request"],"args":[{"prim":"address","annots":["%owner"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"nat","annots":["%balance"]}]}]}]}]},{"prim":"list","annots":["%iTransfer"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%from_"]},{"prim":"list","annots":["%txs"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%to_"]},{"prim":"nat","annots":["%token_id"]},{"prim":"nat","annots":["%amount"]}]}]}]}]}]},{"prim":"list","annots":["%iUpdate_operators"],"args":[{"prim":"or","args":[{"prim":"pair","annots":["%add_operator"],"args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","annots":["%remove_operator"],"args":[{"prim":"address","annots":["%owner"]},{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%current_candidate"],"args":[{"prim":"key_hash"}]},{"prim":"option","annots":["%current_delegated"],"args":[{"prim":"key_hash"}]}]},{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_update_time"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_veto"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","annots":["%allowances"],"args":[{"prim":"address"}]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}]}]},{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%reward_paid"]},{"prim":"nat","annots":["%reward_per_sec"]}]},{"prim":"nat","annots":["%reward_per_share"]},{"prim":"nat","annots":["%tez_pool"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%total_supply"]},{"prim":"nat","annots":["%total_votes"]}]},{"prim":"big_map","annots":["%user_rewards"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"nat","annots":["%veto"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%vetos"],"args":[{"prim":"key_hash"},{"prim":"timestamp"}]},{"prim":"big_map","annots":["%voters"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%candidate"],"args":[{"prim":"key_hash"}]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}]},{"prim":"big_map","annots":["%votes"],"args":[{"prim":"key_hash"},{"prim":"nat"}]}]},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%current_candidate"],"args":[{"prim":"key_hash"}]},{"prim":"option","annots":["%current_delegated"],"args":[{"prim":"key_hash"}]}]},{"prim":"nat","annots":["%invariant"]},{"prim":"timestamp","annots":["%last_update_time"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_veto"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"set","annots":["%allowances"],"args":[{"prim":"address"}]},{"prim":"nat","annots":["%balance"]}]},{"prim":"nat","annots":["%frozen_balance"]}]}]}]},{"prim":"timestamp","annots":["%period_finish"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%reward_paid"]},{"prim":"nat","annots":["%reward_per_sec"]}]},{"prim":"nat","annots":["%reward_per_share"]},{"prim":"nat","annots":["%tez_pool"]}]},{"prim":"pair","args":[{"prim":"address","annots":["%token_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"nat","annots":["%token_pool"]},{"prim":"nat","annots":["%total_reward"]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%total_supply"]},{"prim":"nat","annots":["%total_votes"]}]},{"prim":"big_map","annots":["%user_rewards"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"nat","annots":["%reward_paid"]}]}]},{"prim":"nat","annots":["%veto"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%vetos"],"args":[{"prim":"key_hash"},{"prim":"timestamp"}]},{"prim":"big_map","annots":["%voters"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%candidate"],"args":[{"prim":"key_hash"}]},{"prim":"timestamp","annots":["%last_veto"]}]},{"prim":"nat","annots":["%veto"]},{"prim":"nat","annots":["%vote"]}]}]}]},{"prim":"big_map","annots":["%votes"],"args":[{"prim":"key_hash"},{"prim":"nat"}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1033"},{"int":"1034"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Some","args":[{"string":"tz1RCFbB9GpALpsZtu6J58sb74dm8qe6XBzv"}]},{"prim":"Some","args":[{"string":"tz1RCFbB9GpALpsZtu6J58sb74dm8qe6XBzv"}]}]},{"int":"0"},{"string":"2021-04-24T14:00:24Z"}]},{"prim":"Pair","args":[{"string":"2021-03-31T13:57:03Z"},{"int":"1035"}]},{"string":"2021-04-30T13:57:03Z"},{"int":"47100000"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"46814486"},{"int":"0"}]},{"int":"0"},{"int":"0"}]},{"prim":"Pair","args":[{"string":"KT1AFA2mwNUMNd4SsujE1YYp29vd8BZejyKW"},{"int":"0"}]},{"int":"0"},{"int":"0"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"0"},{"int":"7924948"}]},{"int":"1036"},{"int":"0"}]},{"prim":"Pair","args":[{"int":"1037"},{"int":"1038"}]},{"int":"1039"}]},{"int":"1040"}]}`,
		Expect: map[string]int64{"dex_lambdas": 1033, "ledger": 1035, "metadata": 1034, "token_lambdas": 1040, "user_rewards": 1036, "vetos": 1037, "voters": 1038, "votes": 1039},
	},
	{
		Name:   "FlameDAO",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"set","annots":["%allowances"],"args":[{"prim":"address"}]},{"prim":"nat","annots":["%balance"]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"set","annots":["%minters"],"args":[{"prim":"address"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bool","annots":["%operators_allowed"]},{"prim":"bool","annots":["%paused"]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"nat","annots":["%total_supply"]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1YtUbTURpWeX1CzHFamxS7fGdaKwKKgMzq"},{"int":"4828"}]},{"int":"4829"},[]]},{"prim":"Pair","args":[{"prim":"True"},{"prim":"False"}]},{"int":"4830"},{"int":"100000000000"}]}`,
		Expect: map[string]int64{"ledger": 4828, "metadata": 4829, "token_metadata": 4830},
	},
	{
		Name:   "FlameDEX",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"option","annots":["%admin_candidate"],"args":[{"prim":"address"}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%allowances"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"address"}]},{"prim":"address"}]},{"prim":"unit"}]},{"prim":"big_map","annots":["%buckets"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%bucket_id"]},{"prim":"nat","annots":["%fee"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%last_rewards_per_tez"]},{"prim":"nat","annots":["%rewards_per_share"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"or","annots":["%token_a"],"args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","annots":["%fa2"],"args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit","annots":["%tz"]}]},{"prim":"nat","annots":["%token_a_res"]}]},{"prim":"pair","args":[{"prim":"or","annots":["%token_b"],"args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","annots":["%fa2"],"args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit","annots":["%tz"]}]},{"prim":"nat","annots":["%token_b_res"]}]}]}]},{"prim":"nat","annots":["%total_supply"]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%buckets_count"]},{"prim":"big_map","annots":["%info_to_id"],"args":[{"prim":"pair","args":[{"prim":"or","annots":["%token_a"],"args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","annots":["%fa2"],"args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit","annots":["%tz"]}]},{"prim":"or","annots":["%token_b"],"args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","annots":["%fa2"],"args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit","annots":["%tz"]}]}]},{"prim":"nat"}]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%last_update_time"]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%rewards_per_sec"]},{"prim":"nat","annots":["%rewards_per_tez"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%rewards_period_end"]},{"prim":"big_map","annots":["%shares"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"nat","annots":["%last_rewards_per_share"]}]},{"prim":"nat","annots":["%rewards"]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%tez_rewards"]},{"prim":"nat","annots":["%tez_total"]}]},{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1YtUbTURpWeX1CzHFamxS7fGdaKwKKgMzq"},{"prim":"None"}]},{"int":"22657"},{"int":"22658"}]},{"prim":"Pair","args":[{"int":"39"},{"int":"22659"}]},{"string":"2022-12-11T09:39:14Z"},{"int":"22660"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"0"},{"int":"0"}]},{"string":"2021-11-02T17:09:12Z"},{"int":"22661"}]},{"prim":"Pair","args":[{"int":"0"},{"int":"1486029453"}]},{"int":"22662"}]}`,
		Expect: map[string]int64{"allowances": 22657, "buckets": 22658, "info_to_id": 22659, "metadata": 22660, "shares": 22661, "token_metadata": 22662},
	},
	{
		Name:   "MAG",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"pair","args":[{"prim":"address","annots":["%bakerValidatorContract"]},{"prim":"address","annots":["%bestUser"]}]}]},{"prim":"pair","args":[{"prim":"map","annots":["%candidates"],"args":[{"prim":"address"},{"prim":"key_hash"}]},{"prim":"pair","args":[{"prim":"key_hash","annots":["%currentBaker"]},{"prim":"nat","annots":["%dividends"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"timestamp","annots":["%finishTime"]},{"prim":"pair","args":[{"prim":"nat","annots":["%lastPot"]},{"prim":"address","annots":["%lastWinner"]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","annots":["%approvals"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%balance"]},{"prim":"nat","annots":["%frozenBalance"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%reward"]},{"prim":"pair","args":[{"prim":"nat","annots":["%rewardDebt"]},{"prim":"nat","annots":["%usedVotes"]}]}]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%liqAddress"]},{"prim":"nat","annots":["%maxGameSupply"]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%piggyBank"]},{"prim":"timestamp","annots":["%piggyBankBreakTime"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%rewardPerToken"]},{"prim":"pair","args":[{"prim":"key_hash","annots":["%secondNextBaker"]},{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%totalIssued"]},{"prim":"pair","args":[{"prim":"nat","annots":["%totalPot"]},{"prim":"nat","annots":["%totalStaked"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%totalSupply"]},{"prim":"nat","annots":["%usedLiqSupply"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%usedSupply"]},{"prim":"map","annots":["%votes"],"args":[{"prim":"key_hash"},{"prim":"nat"}]}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1euJLUBuMwXWdK6uH7KqP5me5BjnsbDHAw"},{"string":"KT1By1szyPt4Z9ajjd2ADH9eadR5tVUzb5vx"},{"string":"tz1iF5y4x9AhAPkKZFywWV3HtcZ3CjsFxmJk"}]},[{"prim":"Elt","args":[{"string":"tz1LKJHvS2Sppb85z4dHcVtGTzDE7fpYenRe"},{"string":"tz1aRoaRhSpRYvFdyvgWLL6TGyRoGF51wDjM"}]},{"prim":"Elt","args":[{"string":"tz1LQmXJFKvNnidGPUawCQgCCgdnp2nRPdEV"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1LRRHo9esP9PPpNorqzp1amLCHVYC1ynCT"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1LrMXpAmdVrP5daG1dyQnTia36EtXQGhzC"},{"string":"tz1dRKU4FQ9QRRQPdaH4zCR6gmCmXfcvcgtB"}]},{"prim":"Elt","args":[{"string":"tz1Lu3STErLHjS9ct8d9SXHDYsjRuyT6mRmx"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1MUDrv9zVNkKK9cvYonXhvCFPrHbWSmZGn"},{"string":"tz1dRKU4FQ9QRRQPdaH4zCR6gmCmXfcvcgtB"}]},{"prim":"Elt","args":[{"string":"tz1PEMYdeJoMXqHeu4343Dc15kaDRarX4fHj"},{"string":"tz1aqcYgG6NuViML5vdWhohHJBYxcDVLNUsE"}]},{"prim":"Elt","args":[{"string":"tz1PQp79DG41Xh9tpFwYxdmMBLXph6BxjK15"},{"string":"tz1SUgyRB8T5jXgXAwS33pgRHAKrafyg87Yc"}]},{"prim":"Elt","args":[{"string":"tz1PwxV2kaHhubkYaRWTnGBvsxXswz7V9GWc"},{"string":"tz3cF8X5V6DGBdZCC1hoqLzjtq95BcDvDKqe"}]},{"prim":"Elt","args":[{"string":"tz1Q4foF5PyRs1VsvdyehhdKb8zMniz48mi6"},{"string":"tz1V4qCyvPKZ5UeqdH14HN42rxvNPQfc9UZg"}]},{"prim":"Elt","args":[{"string":"tz1QkNLU3BLJjLBsSeaUTBEo2QCNsqE2x9pL"},{"string":"tz1VQnqCCqX4K5sP3FNkVSNKTdCAMJDd3E1n"}]},{"prim":"Elt","args":[{"string":"tz1RAKBnmd1X43pkkGqywYCRgf1Uo6D77zMx"},{"string":"tz1dRKU4FQ9QRRQPdaH4zCR6gmCmXfcvcgtB"}]},{"prim":"Elt","args":[{"string":"tz1RmkVNSQtvVEN6EXsSzDi8jHLkHK9TSSWv"},{"string":"tz1gg5bjopPcr9agjamyu9BbXKLibNc2rbAq"}]},{"prim":"Elt","args":[{"string":"tz1SC1LzWawZ7i72djS4jibWYpH1XxFRMoTL"},{"string":"tz1SUgyRB8T5jXgXAwS33pgRHAKrafyg87Yc"}]},{"prim":"Elt","args":[{"string":"tz1SRqikWtR2KNpugCApreA9QeZuqh8swQMa"},{"string":"tz1S5WxdZR5f9NzsPXhr7L9L1vrEb5spZFur"}]},{"prim":"Elt","args":[{"string":"tz1ScoARpp9pPSewsBsiWhgik9nFhZybeJtt"},{"string":"tz1S5WxdZR5f9NzsPXhr7L9L1vrEb5spZFur"}]},{"prim":"Elt","args":[{"string":"tz1SpRwsAdsWshBeugtEaWD5PhesgzRvHVrn"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1Uh9fH9yYJF9ZLJtopz4xCkJyhEXVPuB8E"},{"string":"tz1gg5bjopPcr9agjamyu9BbXKLibNc2rbAq"}]},{"prim":"Elt","args":[{"string":"tz1Uza3jMhipiQMtmeEidizjbSqnqZMQyRv7"},{"string":"tz1R664EP6wjcM1RSUVJ7nrJisTpBW9QyJzP"}]},{"prim":"Elt","args":[{"string":"tz1V3fPD3AKPNwZf42PKGzFaWU3vZ83Gjvfq"},{"string":"tz1WnfXMPaNTBmH7DBPwqCWs9cPDJdkGBTZ8"}]},{"prim":"Elt","args":[{"string":"tz1Vq5mYKXw1dD9js26An8dXdASuzo3bfE2w"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1Wq1vm9C2rDNoNxs5Qb9AiAfPFm7PLRnuc"},{"string":"tz1RHgE6W7SnPdwqL1SAtGWEdynA7HeTbpbN"}]},{"prim":"Elt","args":[{"string":"tz1XhCVoisfsmZoSmQxqXN3wAF5pGy71JF77"},{"string":"tz1dRKU4FQ9QRRQPdaH4zCR6gmCmXfcvcgtB"}]},{"prim":"Elt","args":[{"string":"tz1Xqwfw4o3EJeZuBhrJ6JCHgFparnZWoqnT"},{"string":"tz1VQnqCCqX4K5sP3FNkVSNKTdCAMJDd3E1n"}]},{"prim":"Elt","args":[{"string":"tz1Y8v9LUURRffBVoJNysYTMzb2NranuM3fd"},{"string":"tz1aDiEJf9ztRrAJEXZfcG3CKimoKsGhwVAi"}]},{"prim":"Elt","args":[{"string":"tz1YjsPAFc2RmuKa784bkeQhdVLzbnRVtB7c"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1aHVkqAAU5vybCQYYJa7TWjVrtSdKUHjuF"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1aQbe92zZy6MNeRFtgmgV5dV5R4BNTdqK1"},{"string":"tz1WnfXMPaNTBmH7DBPwqCWs9cPDJdkGBTZ8"}]},{"prim":"Elt","args":[{"string":"tz1bQNoRzcT2qHL8tnZCNgfcoFpWZ9GPMV6s"},{"string":"tz1dRKU4FQ9QRRQPdaH4zCR6gmCmXfcvcgtB"}]},{"prim":"Elt","args":[{"string":"tz1cD31qUc7Vz5g3ThFb3yvUkcj2e3wcetsB"},{"string":"tz1WnfXMPaNTBmH7DBPwqCWs9cPDJdkGBTZ8"}]},{"prim":"Elt","args":[{"string":"tz1cL17YdEHjqEHNi3oAhbP7LZu5Lbmo7eEY"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1dkkfj76D6KkhqwZbDcBy2hRvNc45TEaVM"},{"string":"tz1RHgE6W7SnPdwqL1SAtGWEdynA7HeTbpbN"}]},{"prim":"Elt","args":[{"string":"tz1drEV5soQL85qJMNts3JfkgZxD1oL88gsh"},{"string":"tz1SUgyRB8T5jXgXAwS33pgRHAKrafyg87Yc"}]},{"prim":"Elt","args":[{"string":"tz1eRtn8FnaLVL97RsyJrSX6iDTeKHuBBh48"},{"string":"tz1WnfXMPaNTBmH7DBPwqCWs9cPDJdkGBTZ8"}]},{"prim":"Elt","args":[{"string":"tz1ebqDQxvC6Sx9n1UBgq81ZnoVPQd1cs9w1"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1esBLpDvS35fg9zbzRtj11hzTSyjEyck8Y"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1ez5hc7wfNKrZ65vVRyC6JBBxZF6UYX3aD"},{"string":"tz1R664EP6wjcM1RSUVJ7nrJisTpBW9QyJzP"}]},{"prim":"Elt","args":[{"string":"tz1fE99znpJ8M7ix3Ve874BxR6JsQAMebb7A"},{"string":"tz1NEKxGEHsFufk87CVZcrqWu8o22qh46GK6"}]},{"prim":"Elt","args":[{"string":"tz1gp8fJ8FgDUeJCTnq1HQeCP6CUVTeWDSpA"},{"string":"tz1NEKxGEHsFufk87CVZcrqWu8o22qh46GK6"}]},{"prim":"Elt","args":[{"string":"tz1gz2WQnokTw7PSRB6WjNDKAhSzcfVWxKXZ"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1hS2WbwV27KDdPmW3AWkpUfrpTwZhpnzLR"},{"string":"tz1aDiEJf9ztRrAJEXZfcG3CKimoKsGhwVAi"}]},{"prim":"Elt","args":[{"string":"tz1hxcAckH4YNDv9k5kYPUondLLg65JwxA88"},{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"}]},{"prim":"Elt","args":[{"string":"tz1iF5y4x9AhAPkKZFywWV3HtcZ3CjsFxmJk"},{"string":"tz1dRKU4FQ9QRRQPdaH4zCR6gmCmXfcvcgtB"}]},{"prim":"Elt","args":[{"string":"tz1iPtfG5KaDZtAc5Z5YAm4zwuBw3fbfNgMm"},{"string":"tz1aRoaRhSpRYvFdyvgWLL6TGyRoGF51wDjM"}]},{"prim":"Elt","args":[{"string":"tz1ichH92L9WtocJMmoAqFbjGzoWmRM5EF9k"},{"string":"tz1R664EP6wjcM1RSUVJ7nrJisTpBW9QyJzP"}]},{"prim":"Elt","args":[{"string":"tz2ALu8Jxf2SMgzSZaZEwfiJQbogpUDPStwz"},{"string":"tz1aRoaRhSpRYvFdyvgWLL6TGyRoGF51wDjM"}]},{"prim":"Elt","args":[{"string":"tz2WYko4XXaV4B5mzjguTeRkkgw8LyB2jA2d"},{"string":"tz1S5WxdZR5f9NzsPXhr7L9L1vrEb5spZFur"}]}],{"string":"tz1aRoaRhSpRYvFdyvgWLL6TGyRoGF51wDjM"},{"int":"70803880"}]},{"prim":"Pair","args":[{"string":"2022-06-11T04:14:14Z"},{"int":"80000"},{"string":"tz1M9ejBgrCueVEnRqhCDSCTFQTyRfook4rT"}]},{"int":"5950"},{"string":"KT1UstGGfND1RTuVfhSQGni6WyGaAZuTc487"},{"int":"10000000000000"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"5951"},{"int":"1576242894"},{"string":"2024-12-30T09:22:30Z"}]},{"int":"386732327850434"},{"string":"tz1aDiEJf9ztRrAJEXZfcG3CKimoKsGhwVAi"},{"int":"5952"}]},{"prim":"Pair","args":[{"int":"1668837446"},{"int":"80000"},{"int":"3686851603024"}]},{"prim":"Pair","args":[{"int":"153869000000"},{"int":"9588238481106"}]},{"int":"156779000000"},[{"prim":"Elt","args":[{"string":"tz1NEKxGEHsFufk87CVZcrqWu8o22qh46GK6"},{"int":"0"}]},{"prim":"Elt","args":[{"string":"tz1R664EP6wjcM1RSUVJ7nrJisTpBW9QyJzP"},{"int":"30260863"}]},{"prim":"Elt","args":[{"string":"tz1RHgE6W7SnPdwqL1SAtGWEdynA7HeTbpbN"},{"int":"0"}]},{"prim":"Elt","args":[{"string":"tz1S5WxdZR5f9NzsPXhr7L9L1vrEb5spZFur"},{"int":"0"}]},{"prim":"Elt","args":[{"string":"tz1SUgyRB8T5jXgXAwS33pgRHAKrafyg87Yc"},{"int":"28925009"}]},{"prim":"Elt","args":[{"string":"tz1V4qCyvPKZ5UeqdH14HN42rxvNPQfc9UZg"},{"int":"42076324"}]},{"prim":"Elt","args":[{"string":"tz1VQnqCCqX4K5sP3FNkVSNKTdCAMJDd3E1n"},{"int":"0"}]},{"prim":"Elt","args":[{"string":"tz1WnfXMPaNTBmH7DBPwqCWs9cPDJdkGBTZ8"},{"int":"0"}]},{"prim":"Elt","args":[{"string":"tz1aDiEJf9ztRrAJEXZfcG3CKimoKsGhwVAi"},{"int":"132013390380"}]},{"prim":"Elt","args":[{"string":"tz1aRoaRhSpRYvFdyvgWLL6TGyRoGF51wDjM"},{"int":"427067777809"}]},{"prim":"Elt","args":[{"string":"tz1aqcYgG6NuViML5vdWhohHJBYxcDVLNUsE"},{"int":"0"}]},{"prim":"Elt","args":[{"string":"tz1ci1ARnm8JoYV16Hbe4FoxX17yFEAQVytg"},{"int":"1422287748"}]},{"prim":"Elt","args":[{"string":"tz1dRKU4FQ9QRRQPdaH4zCR6gmCmXfcvcgtB"},{"int":"84407032442"}]},{"prim":"Elt","args":[{"string":"tz1gg5bjopPcr9agjamyu9BbXKLibNc2rbAq"},{"int":"0"}]},{"prim":"Elt","args":[{"string":"tz3cF8X5V6DGBdZCC1hoqLzjtq95BcDvDKqe"},{"int":"10514934402"}]}]]}`,
		Expect: map[string]int64{"ledger": 5950, "metadata": 5951, "token_metadata": 5952},
	},
	{
		Name:   "tzDomains",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%actions"],"args":[{"prim":"string"},{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes"},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%data"],"args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","annots":["%expiry_map"],"args":[{"prim":"bytes"},{"prim":"timestamp"}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","annots":["%records"],"args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%address"],"args":[{"prim":"address"}]},{"prim":"map","annots":["%data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"option","annots":["%expiry_key"],"args":[{"prim":"bytes"}]},{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","annots":["%tzip12_token_id"],"args":[{"prim":"nat"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%reverse_records"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"option","annots":["%name"],"args":[{"prim":"bytes"}]}]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"big_map","annots":["%tzip12_tokens"],"args":[{"prim":"nat"},{"prim":"bytes"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%data"],"args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","annots":["%expiry_map"],"args":[{"prim":"bytes"},{"prim":"timestamp"}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","annots":["%records"],"args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%address"],"args":[{"prim":"address"}]},{"prim":"map","annots":["%data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"option","annots":["%expiry_key"],"args":[{"prim":"bytes"}]},{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","annots":["%tzip12_token_id"],"args":[{"prim":"nat"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%reverse_records"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"option","annots":["%name"],"args":[{"prim":"bytes"}]}]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"big_map","annots":["%tzip12_tokens"],"args":[{"prim":"nat"},{"prim":"bytes"}]}]}]}]}]}]}]},{"prim":"pair","annots":["%store"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%data"],"args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","annots":["%expiry_map"],"args":[{"prim":"bytes"},{"prim":"timestamp"}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","annots":["%records"],"args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%address"],"args":[{"prim":"address"}]},{"prim":"map","annots":["%data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"option","annots":["%expiry_key"],"args":[{"prim":"bytes"}]},{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","annots":["%tzip12_token_id"],"args":[{"prim":"nat"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%reverse_records"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"option","annots":["%name"],"args":[{"prim":"bytes"}]}]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"big_map","annots":["%tzip12_tokens"],"args":[{"prim":"nat"},{"prim":"bytes"}]}]}]}]}]},{"prim":"set","annots":["%trusted_senders"],"args":[{"prim":"address"}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1260"},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"1261"},{"int":"1262"}]},{"int":"1263"},{"int":"127441"}]},{"prim":"Pair","args":[{"string":"KT1W56o8dK5En5hM46VsD1zKtgpqWPhs3bLh"},{"int":"1264"}]},{"int":"1265"},{"int":"1266"}]},[{"string":"KT1F7JKNqwaoLzRsMio1MQC7zv3jG9dHcDdJ"},{"string":"KT1H1MqmUM4aK9i1833EBmYCCEfkbt6ZdSBc"},{"string":"KT1J9VpjiH5cmcsskNb8gEXpBtjD4zrAx4Vo"},{"string":"KT1Mqx5meQbhufngJnUAGEGpa4ZRxhPSiCgB"},{"string":"KT1QHLk1EMUA8BPH3FvRUeUmbTspmAhb7kpd"},{"string":"KT1TnTr6b2YxSx2xUQ8Vz3MoWy771ta66yGx"}]]}`,
		Expect: map[string]int64{"actions": 1260, "data": 1261, "expiry_map": 1262, "metadata": 1263, "records": 1264, "reverse_records": 1265, "tzip12_tokens": 1266},
	},
	{
		Name:   "Rarible",
		Type:   `{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"pair","args":[{"prim":"option","annots":["%owner_candidate"],"args":[{"prim":"address"}]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"pair","args":[{"prim":"big_map","annots":["%royalties"],"args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%partAccount"]},{"prim":"nat","annots":["%partValue"]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"address"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%operator"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat"},{"prim":"address"}]}]},{"prim":"unit"}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%permits"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"nat","annots":["%counter"]},{"prim":"pair","args":[{"prim":"option","annots":["%user_expiry"],"args":[{"prim":"nat"}]},{"prim":"map","annots":["%user_permits"],"args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"option","annots":["%expiry"],"args":[{"prim":"nat"}]},{"prim":"timestamp","annots":["%created_at"]}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%operator_for_all"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"address"}]},{"prim":"unit"}]},{"prim":"pair","args":[{"prim":"nat","annots":["%default_expiry"]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]}]}]}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"string":"tz1PyW1EznU9ADpocaauSi41NCPynBuqf1Kc"},{"prim":"None"},{"prim":"False"},{"int":"55541"},{"int":"55542"},{"int":"55543"},{"int":"55544"},{"int":"55545"},{"int":"55546"},{"int":"31556952"},{"int":"55547"}]}`,
		Expect: map[string]int64{"ledger": 55542, "metadata": 55547, "operator": 55543, "operator_for_all": 55546, "permits": 55545, "royalties": 55541, "token_metadata": 55544},
	},
	{
		Name:   "Tezos.com NFT Gallery",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","annots":["%admin"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%admin"]},{"prim":"bool","annots":["%paused"]}]},{"prim":"option","annots":["%pending_admin"],"args":[{"prim":"address"}]}]},{"prim":"pair","annots":["%assets"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"big_map","annots":["%operators"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit"}]}]},{"prim":"pair","args":[{"prim":"pair","annots":["%permissions"],"args":[{"prim":"or","annots":["%operator"],"args":[{"prim":"unit","annots":["%no_transfer"]},{"prim":"or","args":[{"prim":"unit","annots":["%owner_transfer"]},{"prim":"unit","annots":["%owner_or_operator_transfer"]}]}]},{"prim":"pair","args":[{"prim":"or","annots":["%receiver"],"args":[{"prim":"unit","annots":["%owner_no_hook"]},{"prim":"or","args":[{"prim":"unit","annots":["%optional_owner_hook"]},{"prim":"unit","annots":["%required_owner_hook"]}]}]},{"prim":"pair","args":[{"prim":"or","annots":["%sender"],"args":[{"prim":"unit","annots":["%owner_no_hook"]},{"prim":"or","args":[{"prim":"unit","annots":["%optional_owner_hook"]},{"prim":"unit","annots":["%required_owner_hook"]}]}]},{"prim":"option","annots":["%custom"],"args":[{"prim":"pair","args":[{"prim":"string","annots":["%tag"]},{"prim":"option","annots":["%config_api"],"args":[{"prim":"address"}]}]}]}]}]}]},{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]}]}]},{"prim":"big_map","annots":["%token_total_supply"],"args":[{"prim":"nat"},{"prim":"nat"}]}]}]},{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1YEF3iQydxFfcuGVnzUbt9N7ike6Yzvph8"},{"prim":"False"}]},{"prim":"None"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"8215"},{"int":"8216"}]},{"prim":"Pair","args":[{"prim":"Right","args":[{"prim":"Right","args":[{"prim":"Unit"}]}]},{"prim":"Right","args":[{"prim":"Left","args":[{"prim":"Unit"}]}]},{"prim":"Left","args":[{"prim":"Unit"}]},{"prim":"None"}]},{"int":"8217"}]},{"int":"8218"}]},{"int":"8219"}]}`,
		Expect: map[string]int64{"ledger": 8215, "metadata": 8219, "operators": 8216, "token_metadata": 8217, "token_total_supply": 8218},
	},
	{
		Name:   "KT1E9vnFywiWig3xpp9NDj1k7U6UFMFGqhrT",
		Type:   `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%administrator"]},{"prim":"pair","args":[{"prim":"nat","annots":["%all_tokens"]},{"prim":"address","annots":["%ballondor"]}]}]},{"prim":"pair","args":[{"prim":"list","annots":["%default_royalties"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%partAccount"]},{"prim":"nat","annots":["%partValue"]}]}]},{"prim":"pair","args":[{"prim":"bytes","annots":["%default_token_metadata"]},{"prim":"big_map","annots":["%ledger"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%limit_supply"]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"set","annots":["%operator_for_all"],"args":[{"prim":"address"}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%operators"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"pair","args":[{"prim":"address","annots":["%operator"]},{"prim":"nat","annots":["%token_id"]}]}]},{"prim":"unit"}]},{"prim":"pair","args":[{"prim":"bool","annots":["%paused"]},{"prim":"timestamp","annots":["%presale_end_time"]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%presale_price"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%presale_start_time"]},{"prim":"nat","annots":["%presale_supply"]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%revealer"]},{"prim":"pair","args":[{"prim":"big_map","annots":["%royalties"],"args":[{"prim":"nat"},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%partAccount"]},{"prim":"nat","annots":["%partValue"]}]}]}]},{"prim":"nat","annots":["%sale_limit"]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%sale_price"]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%sale_start_time"]},{"prim":"list","annots":["%shares"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%partAccount"]},{"prim":"nat","annots":["%partValue"]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%token_metadata"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"nat","annots":["%token_id"]},{"prim":"map","annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"big_map","annots":["%total_supply"],"args":[{"prim":"nat"},{"prim":"nat"}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%whitelist"]},{"prim":"nat","annots":["%whitelist_token_id"]}]}]}]}]}]}`,
		Value:  `{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"prim":"Pair","args":[{"string":"tz1Z9bSe4rCT6uTseBu7FjvY2RJAkcK2Ux6a"},{"int":"2022"},{"string":"tz1hCKMo9r9sb2epHzUBZgQKCX1uNAY32vM7"}]},[{"prim":"Pair","args":[{"string":"tz1hCKMo9r9sb2epHzUBZgQKCX1uNAY32vM7"},{"int":"50"}]},{"prim":"Pair","args":[{"string":"tz1WyZ4mwysxkgiMbRQn5DXzwxw9c6eVdc42"},{"int":"500"}]},{"prim":"Pair","args":[{"string":"tz1Xur15cBJSsKU3iDfnWg7sQmk5cjkhashP"},{"int":"200"}]}],{"bytes":"697066733a2f2f516d577833705056374a3371624551763853546d693551764d716a5837756d41506a44654d6f5935337275714476"},{"int":"315235"}]},{"prim":"Pair","args":[{"int":"2022"},{"int":"315236"},[]]},{"int":"315237"},{"prim":"False"},{"string":"2022-10-25T16:00:00Z"}]},{"prim":"Pair","args":[{"prim":"Pair","args":[{"int":"145000000"},{"string":"2022-10-23T16:00:00Z"},{"int":"1772"}]},{"string":"tz1h7T8bTvzjVNCJFycXyYu5gX4oi56HrkNz"},{"int":"315238"},{"int":"10"}]},{"prim":"Pair","args":[{"int":"180000000"},{"string":"2022-10-25T16:00:00Z"},[{"prim":"Pair","args":[{"string":"tz1WyZ4mwysxkgiMbRQn5DXzwxw9c6eVdc42"},{"int":"500"}]},{"prim":"Pair","args":[{"string":"tz1Xur15cBJSsKU3iDfnWg7sQmk5cjkhashP"},{"int":"300"}]}]]},{"prim":"Pair","args":[{"int":"315239"},{"int":"315240"}]},{"string":"KT1PAZMHzFtZPGHxXBnfZpYiNzb3XqfnCKdR"},{"int":"0"}]}`,
		Expect: map[string]int64{"ledger": 315235, "metadata": 315236, "operators": 315237, "royalties": 315238, "token_metadata": 315239, "total_supply": 315240},
	},
	// {
	// 	Name:   "",
	// 	Type:   ``,
	// 	Value:  ``,
	// 	Expect: map[string]int64{},
	// },
}

func TestBigmapDetect(t *testing.T) {
	for _, test := range bigmapDetectTests {
		t.Run(test.Name, func(T *testing.T) {
			var typ Prim
			if err := typ.UnmarshalJSON([]byte(test.Type)); err != nil {
				T.Fatalf("unmarshal type: %v", err)
			}
			var val Prim
			if err := val.UnmarshalJSON([]byte(test.Value)); err != nil {
				T.Fatalf("unmarshal value: %v", err)
			}
			detected := DetectBigmaps(typ, val)
			if have, want := len(detected), len(test.Expect); have != want {
				T.Errorf("mismatch count want=%d have=%d", want, have)
			}
			for n, i := range detected {
				if j, ok := test.Expect[n]; !ok {
					T.Errorf("unexpected detected bigmap %s %d", n, i)
				} else if i != j {
					T.Errorf("wrong id for bigmap %s want=%d have=%d", n, j, i)
				}
			}
			for n, i := range test.Expect {
				if j, ok := detected[n]; !ok {
					T.Errorf("undetected bigmap %s %d", n, i)
				} else if i != j {
					T.Errorf("wrong id for bigmap %s want=%d have=%d", n, i, j)
				}
			}
		})
	}
}

func TestBigmapTypeDetect(t *testing.T) {
	for _, test := range bigmapDetectTests {
		t.Run(test.Name, func(T *testing.T) {
			if test.SkipTypetest {
				return
			}
			var typ Prim
			if err := typ.UnmarshalJSON([]byte(test.Type)); err != nil {
				T.Fatalf("unmarshal type: %v", err)
			}
			var val Prim
			if err := val.UnmarshalJSON([]byte(test.Value)); err != nil {
				T.Fatalf("unmarshal value: %v", err)
			}
			found := detectBigmapTypes(typ)
			detected := DetectBigmapTypes(typ)
			if have, want := len(detected), len(found); have != want {
				T.Errorf("mismatch count want=%d have=%d", want, have)
			}
			for n, i := range detected {
				if j, ok := found[n]; !ok {
					T.Errorf("unexpected detected bigmap %s %s", n, i.Dump())
				} else if !i.IsEqual(j) {
					T.Errorf("type mismatch for bigmap %s want=%s have=%s", n, j.Dump(), i.Dump())
				}
			}
			for n, i := range found {
				if j, ok := detected[n]; !ok {
					T.Errorf("undetected bigmap %s %s", n, i.Dump())
				} else if !i.IsEqual(j) {
					T.Errorf("type mismatch for bigmap %s want=%s have=%s", n, i.Dump(), j.Dump())
				}
			}
		})
	}
}

// previously used algo that just scans for bigmap type opcodes; this algo
// cannot reuse outer type annotations to properly name bigmaps inside maps or lists
func detectBigmapTypes(p Prim) map[string]Type {
	named := make(map[string]Type)
	bigmaps, _ := p.FindOpCodes(T_BIG_MAP)
	for i := range bigmaps {
		n := bigmaps[i].GetVarAnnoAny()
		if n == "" {
			n = "bigmap_" + strconv.Itoa(i)
		}
		if _, ok := named[n]; ok {
			n += "_" + strconv.Itoa(i)
		}
		named[n] = NewType(bigmaps[i])
	}
	return named
}

// Test comparison between annotated and non-annotated bigmaps, must be resilient to
// format differences introduced by comb pairs
type bigmapTypeCompareTest struct {
	Name         string
	SrcKeyType   string
	SrcValueType string
	DstKeyType   string
	DstValueType string
	Expect       bool
}

var bigmapCompareTests = []bigmapTypeCompareTest{
	{
		Name:         "KT1UKKJeQ7wbppyzyLWMoCSKVhFpMVPHgoPm_204645",
		SrcKeyType:   `{"prim":"address"}`,
		SrcValueType: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]}`,
		DstKeyType:   `{"prim":"address"}`,
		DstValueType: `{"prim":"pair","args":[{"prim":"pair","annots":["%accountBorrows"],"args":[{"prim":"nat","annots":["%interestIndex"]},{"prim":"nat","annots":["%principal"]}]},{"prim":"pair","args":[{"prim":"map","annots":["%approvals"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%balance"]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1VPz82qKWshaAAxDURmZfYQRmUY9WQRjLu_203763",
		SrcKeyType:   `{"prim":"nat"}`,
		SrcValueType: `{"prim":"pair","args":[{"prim":"bool"},{"prim":"pair","args":[{"prim":"address"},{"prim":"mutez"},{"prim":"mutez"},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"},{"prim":"nat"}]}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"},{"prim":"nat"}]}]}]},{"prim":"bool"}]}`,
		DstKeyType:   `{"prim":"nat"}`,
		DstValueType: `{"prim":"pair","args":[{"prim":"bool","annots":["%executed"]},{"prim":"pair","args":[{"prim":"pair","annots":["%proposal"],"args":[{"prim":"address","annots":["%user1"]},{"prim":"pair","args":[{"prim":"mutez","annots":["%mutez_amount1"]},{"prim":"pair","args":[{"prim":"mutez","annots":["%mutez_amount2"]},{"prim":"pair","args":[{"prim":"list","annots":["%tokens1"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%fa2"]},{"prim":"pair","args":[{"prim":"nat","annots":["%id"]},{"prim":"nat","annots":["%amount"]}]}]}]},{"prim":"list","annots":["%tokens2"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%fa2"]},{"prim":"pair","args":[{"prim":"nat","annots":["%id"]},{"prim":"nat","annots":["%amount"]}]}]}]}]}]}]}]},{"prim":"bool","annots":["%user1_accepted"]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1MWV1oaRAZaQvXcnJWf3tA8B52sb9cEv5i_6232",
		SrcKeyType:   `{"prim":"string"}`,
		SrcValueType: `{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes"},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}]}]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"address"},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"option","args":[{"prim":"bytes"}]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"address"}]},{"prim":"option","args":[{"prim":"nat"}]}]}]}]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"option","args":[{"prim":"bytes"}]}]},{"prim":"address"}]}]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"timestamp"}]}]},{"prim":"big_map","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"address"},{"prim":"big_map","args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"address"}]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"option","args":[{"prim":"bytes"}]},{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"address"}]},{"prim":"option","args":[{"prim":"nat"}]}]}]}]},{"prim":"big_map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"option","args":[{"prim":"bytes"}]}]},{"prim":"address"}]}]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"bytes"}]}]}]}`,
		DstKeyType:   `{"prim":"string"}`,
		DstValueType: `{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"bytes"},{"prim":"address"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%data"],"args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","annots":["%expiry_map"],"args":[{"prim":"bytes"},{"prim":"timestamp"}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","annots":["%records"],"args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%address"],"args":[{"prim":"address"}]},{"prim":"map","annots":["%data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"option","annots":["%expiry_key"],"args":[{"prim":"bytes"}]},{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","annots":["%tzip12_token_id"],"args":[{"prim":"nat"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%reverse_records"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"option","annots":["%name"],"args":[{"prim":"bytes"}]}]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"big_map","annots":["%tzip12_tokens"],"args":[{"prim":"nat"},{"prim":"bytes"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"list","args":[{"prim":"operation"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"big_map","annots":["%data"],"args":[{"prim":"bytes"},{"prim":"bytes"}]},{"prim":"big_map","annots":["%expiry_map"],"args":[{"prim":"bytes"},{"prim":"timestamp"}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"nat","annots":["%next_tzip12_token_id"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%owner"]},{"prim":"big_map","annots":["%records"],"args":[{"prim":"bytes"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%address"],"args":[{"prim":"address"}]},{"prim":"map","annots":["%data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]},{"prim":"pair","args":[{"prim":"option","annots":["%expiry_key"],"args":[{"prim":"bytes"}]},{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"address","annots":["%owner"]}]},{"prim":"option","annots":["%tzip12_token_id"],"args":[{"prim":"nat"}]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%reverse_records"],"args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"map","annots":["%internal_data"],"args":[{"prim":"string"},{"prim":"bytes"}]},{"prim":"option","annots":["%name"],"args":[{"prim":"bytes"}]}]},{"prim":"address","annots":["%owner"]}]}]},{"prim":"big_map","annots":["%tzip12_tokens"],"args":[{"prim":"nat"},{"prim":"bytes"}]}]}]}]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1LDM81PK3gg5QZ4wt6m5tHUC3PrdrXmYPf_238400 ",
		SrcKeyType:   `{"prim":"nat"}`,
		SrcValueType: `{"prim":"pair","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"mutez"},{"prim":"nat"}]}`,
		DstKeyType:   `{"prim":"nat"}`,
		DstValueType: `{"prim":"pair","args":[{"prim":"address","annots":["%seller"]},{"prim":"pair","annots":["%sale_data"],"args":[{"prim":"pair","annots":["%sale_token"],"args":[{"prim":"address","annots":["%fa2_address"]},{"prim":"nat","annots":["%token_id"]}]},{"prim":"pair","args":[{"prim":"mutez","annots":["%price"]},{"prim":"nat","annots":["%amount"]}]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1LxPGwkrvj8gG8k8CkpyKQaWyQAsnLfHLg_256783",
		SrcKeyType:   `{"prim":"address"}`,
		SrcValueType: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]}`,
		DstKeyType:   `{"prim":"address"}`,
		DstValueType: `{"prim":"pair","args":[{"prim":"pair","annots":["%accountBorrows"],"args":[{"prim":"nat","annots":["%interestIndex"]},{"prim":"nat","annots":["%principal"]}]},{"prim":"pair","args":[{"prim":"map","annots":["%approvals"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%balance"]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1VH3JQSZaSq3qeUTBmeJjiqSfP4nTkMo9X_311744",
		SrcKeyType:   `{"prim":"nat"}`,
		SrcValueType: `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]}`,
		DstKeyType:   `{"prim":"nat"}`,
		DstValueType: `{"prim":"list","args":[{"prim":"pair","args":[{"prim":"address","annots":["%partAccount"]},{"prim":"nat","annots":["%partValue"]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1M8tp17Y2GWsYrZDBRXqdRNawgwyFDzmp4_306129",
		SrcKeyType:   `{"prim":"or","args":[{"prim":"or","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"unit"}]}`,
		SrcValueType: `{"prim":"nat"}`,
		DstKeyType:   `{"prim":"or","args":[{"prim":"or","args":[{"prim":"address","annots":["%fa12"]},{"prim":"pair","annots":["%fa2"],"args":[{"prim":"address","annots":["%token"]},{"prim":"nat","annots":["%id"]}]}]},{"prim":"unit","annots":["%tez"]}]}`,
		DstValueType: `{"prim":"nat"}`,
		Expect:       true,
	},
	{
		Name:         "KT1QH9RJYFdopvxzo7Z8TzpU7hQsvfcAqz5L_301197",
		SrcKeyType:   `{"prim":"nat"}`,
		SrcValueType: `{"prim":"map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"timestamp"},{"prim":"nat"}]}]}`,
		DstKeyType:   `{"prim":"nat"}`,
		DstValueType: `{"prim":"map","args":[{"prim":"address"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%rate"]},{"prim":"nat","annots":["%reward"]}]},{"prim":"pair","args":[{"prim":"timestamp","annots":["%timestamp"]},{"prim":"nat","annots":["%value"]}]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1NRLjyE7wxeSZ6La6DfuhSKCAAnc9Lnvdg_209035",
		SrcKeyType:   `{"prim":"nat"}`,
		SrcValueType: `{"prim":"or","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"or","args":[{"prim":"nat"},{"prim":"nat"}]}]}`,
		DstKeyType:   `{"prim":"nat"}`,
		DstValueType: `{"prim":"or","args":[{"prim":"pair","annots":["%mission"],"args":[{"prim":"nat","annots":["%mission_id"]},{"prim":"nat","annots":["%start_block"]}]},{"prim":"or","args":[{"prim":"nat","annots":["%rest"]},{"prim":"nat","annots":["%train"]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1Nr4CLi7hY7QrZe8D4ar6uihpnr7nRtGJH_256939",
		SrcKeyType:   `{"prim":"string"}`,
		SrcValueType: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string"},{"prim":"timestamp"}]},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"nat"},{"prim":"option","args":[{"prim":"timestamp"}]},{"prim":"bool"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"mutez"}]},{"prim":"string"},{"prim":"bool"}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"timestamp"}]},{"prim":"bool"}]},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"timestamp"}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"timestamp"}]}]}]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]}]}`,
		DstKeyType:   `{"prim":"string"}`,
		DstValueType: `{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"string","annots":["%bid"]},{"prim":"timestamp","annots":["%challenge_time"]}]},{"prim":"pair","args":[{"prim":"pair","annots":["%challenged"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%challenged_damage"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","annots":["%challenger"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat","annots":["%challenger_damage"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%experience_gained"]},{"prim":"pair","args":[{"prim":"option","annots":["%finish_time"],"args":[{"prim":"timestamp"}]},{"prim":"bool","annots":["%finished"]}]}]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%looser"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]},{"prim":"mutez","annots":["%loot"]}]},{"prim":"pair","args":[{"prim":"string","annots":["%mode"]},{"prim":"bool","annots":["%resolved"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%start_time"],"args":[{"prim":"timestamp"}]},{"prim":"bool","annots":["%started"]}]},{"prim":"pair","args":[{"prim":"pair","annots":["%turn"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"pair","annots":["%turns"],"args":[{"prim":"option","annots":["%latest"],"args":[{"prim":"timestamp"}]},{"prim":"list","annots":["%turns"],"args":[{"prim":"pair","args":[{"prim":"nat","annots":["%damage"]},{"prim":"pair","args":[{"prim":"pair","annots":["%hero"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"timestamp","annots":["%timestamp"]}]}]}]}]},{"prim":"option","annots":["%victor"],"args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]}]}]}]}]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1RwwjAS89iotPSKGWmabcPaUSZ1fdGjS2Z_328858",
		SrcKeyType:   `{"prim":"nat"}`,
		SrcValueType: `{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"nat"}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"unit"}]},{"prim":"nat"}]}]}]}]},{"prim":"address"}]}]},{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"nat"}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"unit"}]},{"prim":"nat"}]}]}]}]},{"prim":"address"}]}]},{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"nat"}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"unit"}]},{"prim":"nat"}]}]}]}]},{"prim":"address"}]}]},{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"},{"prim":"nat"}]}]}]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"nat"}]}]},{"prim":"nat"}]}]},{"prim":"big_map","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"nat"}]},{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"list","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"unit"},{"prim":"unit"}]},{"prim":"nat"}]}]}]}]},{"prim":"address"}]}]},{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"},{"prim":"nat"}]}]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"},{"prim":"nat"}]}]},{"prim":"address"},{"prim":"address"}]},{"prim":"or","args":[{"prim":"operation"},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"operation"}]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]}]}]}]}]}`,
		DstKeyType:   `{"prim":"nat"}`,
		DstValueType: `{"prim":"lambda","args":[{"prim":"pair","args":[{"prim":"or","args":[{"prim":"pair","annots":["%sell"],"args":[{"prim":"pair","args":[{"prim":"pair","annots":["%dex"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%contract_address"]},{"prim":"or","annots":["%ext_type"],"args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%defaultCP"]},{"prim":"nat","annots":["%quipuDex2"]}]},{"prim":"or","args":[{"prim":"pair","annots":["%quipuStableSwap"],"args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"nat","annots":["%pool_token1_id"]}]},{"prim":"nat","annots":["%pool_token2_id"]}]},{"prim":"list","annots":["%quipuTokenToToken"],"args":[{"prim":"pair","args":[{"prim":"or","annots":["%operation"],"args":[{"prim":"unit","annots":["%a_to_b"]},{"prim":"unit","annots":["%b_to_a"]}]},{"prim":"nat","annots":["%pair_id"]}]}]}]}]},{"prim":"address","annots":["%siriusLB"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%fee"]},{"prim":"nat","annots":["%provider"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%token1_id"]},{"prim":"nat","annots":["%token2_id"]}]}]},{"prim":"nat","annots":["%min_out"]}]},{"prim":"nat","annots":["%x_amount"]}]},{"prim":"or","args":[{"prim":"pair","annots":["%buy"],"args":[{"prim":"pair","args":[{"prim":"pair","annots":["%dex"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%contract_address"]},{"prim":"or","annots":["%ext_type"],"args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%defaultCP"]},{"prim":"nat","annots":["%quipuDex2"]}]},{"prim":"or","args":[{"prim":"pair","annots":["%quipuStableSwap"],"args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"nat","annots":["%pool_token1_id"]}]},{"prim":"nat","annots":["%pool_token2_id"]}]},{"prim":"list","annots":["%quipuTokenToToken"],"args":[{"prim":"pair","args":[{"prim":"or","annots":["%operation"],"args":[{"prim":"unit","annots":["%a_to_b"]},{"prim":"unit","annots":["%b_to_a"]}]},{"prim":"nat","annots":["%pair_id"]}]}]}]}]},{"prim":"address","annots":["%siriusLB"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%fee"]},{"prim":"nat","annots":["%provider"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%token1_id"]},{"prim":"nat","annots":["%token2_id"]}]}]},{"prim":"nat","annots":["%min_out"]}]},{"prim":"nat","annots":["%y_amount"]}]},{"prim":"pair","annots":["%getReserves"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%contract_address"]},{"prim":"or","annots":["%ext_type"],"args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%defaultCP"]},{"prim":"nat","annots":["%quipuDex2"]}]},{"prim":"or","args":[{"prim":"pair","annots":["%quipuStableSwap"],"args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"nat","annots":["%pool_token1_id"]}]},{"prim":"nat","annots":["%pool_token2_id"]}]},{"prim":"list","annots":["%quipuTokenToToken"],"args":[{"prim":"pair","args":[{"prim":"or","annots":["%operation"],"args":[{"prim":"unit","annots":["%a_to_b"]},{"prim":"unit","annots":["%b_to_a"]}]},{"prim":"nat","annots":["%pair_id"]}]}]}]}]},{"prim":"address","annots":["%siriusLB"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%fee"]},{"prim":"nat","annots":["%provider"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%token1_id"]},{"prim":"nat","annots":["%token2_id"]}]}]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%tokens"],"args":[{"prim":"nat"},{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","annots":["%fa12"],"args":[{"prim":"address","annots":["%contract_address"]},{"prim":"nat","annots":["%decimals"]}]},{"prim":"pair","annots":["%fa2"],"args":[{"prim":"pair","args":[{"prim":"address","annots":["%contract_address"]},{"prim":"nat","annots":["%decimals"]}]},{"prim":"nat","annots":["%fa2_id"]}]}]},{"prim":"nat","annots":["%xtz"]}]}]},{"prim":"pair","args":[{"prim":"big_map","annots":["%dexes"],"args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address","annots":["%contract_address"]},{"prim":"or","annots":["%ext_type"],"args":[{"prim":"or","args":[{"prim":"or","args":[{"prim":"unit","annots":["%defaultCP"]},{"prim":"nat","annots":["%quipuDex2"]}]},{"prim":"or","args":[{"prim":"pair","annots":["%quipuStableSwap"],"args":[{"prim":"pair","args":[{"prim":"nat","annots":["%pool_id"]},{"prim":"nat","annots":["%pool_token1_id"]}]},{"prim":"nat","annots":["%pool_token2_id"]}]},{"prim":"list","annots":["%quipuTokenToToken"],"args":[{"prim":"pair","args":[{"prim":"or","annots":["%operation"],"args":[{"prim":"unit","annots":["%a_to_b"]},{"prim":"unit","annots":["%b_to_a"]}]},{"prim":"nat","annots":["%pair_id"]}]}]}]}]},{"prim":"address","annots":["%siriusLB"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%fee"]},{"prim":"nat","annots":["%provider"]}]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%token1_id"]},{"prim":"nat","annots":["%token2_id"]}]}]}]},{"prim":"pair","args":[{"prim":"option","annots":["%sirius_estimates"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%level"]},{"prim":"nat","annots":["%lqt"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%token"]},{"prim":"nat","annots":["%xtz"]}]}]}]},{"prim":"pair","args":[{"prim":"address","annots":["%ctez_admin"]},{"prim":"address","annots":["%dex_referral"]}]}]}]}]}]},{"prim":"or","args":[{"prim":"operation","annots":["%swap"]},{"prim":"pair","annots":["%getReserves"],"args":[{"prim":"option","annots":["%operation"],"args":[{"prim":"operation"}]},{"prim":"option","annots":["%reserves"],"args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]}]}]}]}]}`,
		Expect:       true,
	},
	{
		Name:         "KT1Sm4gYTvQ6PRN49vhH3ZHXGc46ZQJWSKJY_408375",
		SrcKeyType:   `{"prim":"nat"}`,
		SrcValueType: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"nat"},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"option","args":[{"prim":"nat"}]}]},{"prim":"nat"}]},{"prim":"option","args":[{"prim":"nat"}]}]},{"prim":"option","args":[{"prim":"nat"}]}]}]},{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat"}]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","args":[{"prim":"nat"}]},{"prim":"nat"}]},{"prim":"pair","args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"option","args":[{"prim":"nat"}]}]}]}]}]}`,
		DstKeyType:   `{"prim":"nat"}`,
		DstValueType: `{"prim":"or","args":[{"prim":"or","args":[{"prim":"pair","annots":["%branch"],"args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%left"]},{"prim":"nat","annots":["%left_height"]}]},{"prim":"pair","args":[{"prim":"nat","annots":["%left_tok"]},{"prim":"nat","annots":["%parent"]}]}]},{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"nat","annots":["%right"]},{"prim":"nat","annots":["%right_height"]}]},{"prim":"nat","annots":["%right_tok"]}]}]},{"prim":"pair","annots":["%leaf"],"args":[{"prim":"nat","annots":["%parent"]},{"prim":"pair","annots":["%value"],"args":[{"prim":"pair","args":[{"prim":"pair","annots":["%contents"],"args":[{"prim":"pair","args":[{"prim":"pair","annots":["%burrow"],"args":[{"prim":"address"},{"prim":"nat"}]},{"prim":"option","annots":["%min_kit_for_unwarranted"],"args":[{"prim":"nat"}]}]},{"prim":"nat","annots":["%tok"]}]},{"prim":"option","annots":["%older"],"args":[{"prim":"nat"}]}]},{"prim":"option","annots":["%younger"],"args":[{"prim":"nat"}]}]}]}]},{"prim":"pair","annots":["%root"],"args":[{"prim":"option","args":[{"prim":"nat"}]},{"prim":"option","args":[{"prim":"pair","args":[{"prim":"pair","args":[{"prim":"option","annots":["%older_auction"],"args":[{"prim":"nat"}]},{"prim":"nat","annots":["%sold_tok"]}]},{"prim":"pair","args":[{"prim":"pair","annots":["%winning_bid"],"args":[{"prim":"address","annots":["%address"]},{"prim":"nat","annots":["%kit"]}]},{"prim":"option","annots":["%younger_auction"],"args":[{"prim":"nat"}]}]}]}]}]}]}`,
		Expect:       true,
	},
	// {
	// Name: "",
	// SrcKeyType: ``,
	// SrcValueType: ``,
	// DstKeyType: ``,
	// DstValueType: ``,
	// Expect: true,
	// },
}

func TestBigmapTypeCompare(t *testing.T) {
	for _, test := range bigmapCompareTests {
		t.Run(test.Name, func(T *testing.T) {
			var sk, sv, dk, dv Prim
			if err := sk.UnmarshalJSON([]byte(test.SrcKeyType)); err != nil {
				T.Fatalf("unmarshal type: %v", err)
			}
			if err := sv.UnmarshalJSON([]byte(test.SrcValueType)); err != nil {
				T.Fatalf("unmarshal value: %v", err)
			}
			if err := dk.UnmarshalJSON([]byte(test.DstKeyType)); err != nil {
				T.Fatalf("unmarshal type: %v", err)
			}
			if err := dv.UnmarshalJSON([]byte(test.DstValueType)); err != nil {
				T.Fatalf("unmarshal value: %v", err)
			}
			if want, have := test.Expect, NewType(sk).Typedef("").Unfold().Equal(NewType(dk).Typedef("").Unfold()); want != have {
				T.Errorf("key type want=%t have=%t\nsrc=%s\ndst=%s", want, have, sk.Dump(), dk.Dump())
			}
			if want, have := test.Expect, NewType(sv).Typedef("").Unfold().Equal(NewType(dv).Typedef("").Unfold()); want != have {
				T.Errorf("value type want=%t have=%t\nsrc=%s\ndst=%s", want, have, sv.Dump(), dv.Dump())
			}
		})
	}
}
