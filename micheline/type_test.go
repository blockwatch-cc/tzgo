// Copyright (c) 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//

package micheline

import (
	"testing"
)

type typedefTest struct {
	Name string
	Spec string
	Want string
}

var typedefInfo = []typedefTest{
	// scalars
	//   int
	typedefTest{
		Name: "int",
		Spec: `{"annots": ["%payoutDelay"],"prim": "int"}`,
		Want: `{"name":"payoutDelay","type":"int"}`,
	},
	//   nat
	typedefTest{
		Name: "nat",
		Spec: `{"annots": ["%payoutFrequency"],"prim": "nat"}`,
		Want: `{"name":"payoutFrequency","type":"nat"}`,
	},
	//   string
	typedefTest{
		Name: "string",
		Spec: `{"annots": ["%name"],"prim": "string"}`,
		Want: `{"name":"name","type":"string"}`,
	},
	//   bytes
	typedefTest{
		Name: "bytes",
		Spec: `{"annots": ["%bakerName"],"prim": "bytes"}`,
		Want: `{"name":"bakerName","type":"bytes"}`,
	},
	//   mutez
	typedefTest{
		Name: "mutez",
		Spec: `{"annots": ["%signup_fee"],"prim": "mutez"}`,
		Want: `{"name":"signup_fee","type":"mutez"}`,
	},
	//   bool
	typedefTest{
		Name: "bool",
		Spec: `{"annots": ["%bakerChargesTransactionFee"],"prim": "bool"}`,
		Want: `{"name":"bakerChargesTransactionFee","type":"bool"}`,
	},
	//   key_hash
	typedefTest{
		Name: "key_hash",
		Spec: `{"annots": ["%baker"],"prim": "key_hash"}`,
		Want: `{"name":"baker","type":"key_hash"}`,
	},
	//   timestamp
	typedefTest{
		Name: "timestamp",
		Spec: `{"annots": ["%last_update"],"prim": "timestamp"}`,
		Want: `{"name":"last_update","type":"timestamp"}`,
	},
	//   address
	typedefTest{
		Name: "address",
		Spec: `{"annots": ["%reporterAccount"],"prim": "address"}`,
		Want: `{"name":"reporterAccount","type":"address"}`,
	},
	//   key
	typedefTest{
		Name: "key",
		Spec: `{"annots": ["%pour_authorizer"],"prim": "key"}`,
		Want: `{"name":"pour_authorizer","type":"key"}`,
	},
	//   unit
	//   signature
	typedefTest{
		Name: "signature",
		Spec: `{"args":[{"args":[{"prim":"nat"},{"args":[{"prim":"key"},{"prim":"signature"}],"prim": "pair"}],"prim": "pair"}],"prim": "pair"}`,
		Want: `{"name":"","type":"struct","args":[{"name":"0","type":"nat"},{"name":"1","type":"key"},{"name":"2","type":"signature"}]}`,
	},
	//   chain_id
	//   bls12_381_g1
	//   bls12_381_g2
	//   bls12_381_fr
	typedefTest{
		Name: "bls",
		Spec: `{"annots":["%g2"],"prim":"bls12_381_g2"}`,
		Want: `{"name":"g2","type":"bls12_381_g2"}`,
	},
	//   sapling_state
	typedefTest{
		Name: "sapling_state",
		Spec: `{"prim":"sapling_state","args":[{"int":"8"}]}`,
		Want: `{"name":"","type":"sapling_state(8)"}`,
	},
	//   sapling_transaction
	typedefTest{
		Name: "sapling_transaction",
		Spec: `{"prim":"sapling_transaction","args":[{"int":"8"}]}`,
		Want: `{"name":"","type":"sapling_transaction(8)"}`,
	},
	//   never
	typedefTest{
		Name: "never",
		Spec: `{"prim":"never"}`,
		Want: `{"name":"","type":"never"}`,
	},
	// set
	typedefTest{
		Name: "set",
		Spec: `{"annots": ["%admins"],"prim": "set", "args": [{"prim": "key_hash"}]}`,
		Want: `{"name":"admins","type":"set","args":[{"name":"@item","type":"key_hash"}]}`,
	},
	// map
	typedefTest{
		Name: "map",
		Spec: `{"annots":["%approvals"],"prim":"map","args":[{"prim":"address"},{"prim":"nat"}]}`,
		Want: `{"name":"approvals","type":"map","args":[{"name":"@key","type":"address"},{"name":"@value","type":"nat"}]}`,
	},
	// bigmap with scalar key
	// bigmap with pair key
	typedefTest{
		Name: "bigmap",
		Spec: `{"annots": ["%ledger"],"args": [{"args": [{"prim": "address"},{"prim": "nat"}],"prim": "pair"},{"prim": "nat"}],"prim": "big_map"}`,
		Want: `{"name": "ledger", "type": "big_map", "args":[{"name":"@key","type":"struct","args":[{"name":"0","type":"address"},{"name":"1","type":"nat"}]},{"name":"@value","type":"nat"}]}`,
	},
	// contract
	typedefTest{
		Name: "contract",
		Spec: `{"annots": ["%pour_dest"],"args": [{"prim": "unit"}],"prim": "contract"}`,
		Want: `{"name":"pour_dest","type":"contract","args":[{"name":"0","type":"unit"}]}`,
	},
	// lambda, list, operation
	typedefTest{
		Name: "lambda",
		Spec: `{"args": [{"args": [{"args": [{"prim": "string"},{"prim": "bytes"}],"prim": "pair"},{"args": [{"prim": "bytes"},{"prim": "bytes"}],"prim": "big_map"}],"prim": "pair"},{"args": [{"args": [{"prim": "operation"}],"prim": "list"},{"args": [{"prim": "bytes"},{"prim": "bytes"}],"prim": "big_map"}],"prim": "pair"}],"prim": "lambda"}`,
		Want: `{"name":"","type":"lambda","args":[{"name":"@param","type":"struct","args":[{"name":"0","type":"string"},{"name":"1","type":"bytes"},{"name":"2","type":"big_map","args":[{"name":"@key","type":"bytes"},{"name":"@value","type":"bytes"}]}]},{"name":"@return","type":"struct","args":[{"name":"0","type":"list","args":[{"name":"@item","type":"operation"}]},{"name":"1","type":"big_map","args":[{"name":"@key","type":"bytes"},{"name":"@value","type":"bytes"}]}]}]}`,
	},
	// ticket
	typedefTest{
		Name: "ticket",
		Spec: `{"prim": "ticket", "args":[{"prim":"timestamp"}]}`,
		Want: `{"name":"","type":"ticket","args":[{"name":"@value","type":"timestamp"}]}`,
	},
	// option
	typedefTest{
		Name: "option",
		Spec: `{"annots":["%reporterAccount"],"prim":"option","args":[{"prim":"address"}]}`,
		Want: `{"name":"reporterAccount","type":"address","optional":true}`,
	},
	// named union type
	typedefTest{
		Name: "named-union",
		Spec: `{"args":[{"annots":["%do"],"args":[{"prim":"unit"},{"args":[{"prim":"operation"}],"prim":"list"}],"prim":"lambda"},{"annots":["%default"],"prim":"unit"}],"prim":"or"}`,
		Want: `{"name":"","type":"union","args":[{"name":"do","type":"lambda","args":[{"name":"@param","type":"unit"},{"name":"@return","type":"list","args":[{"name":"@item","type":"operation"}]}]},{"name":"default","type":"unit"}]}`,
	},
	// anonymous union type
	typedefTest{
		Name: "anon-union",
		Spec: `{"args":[{"args":[{"prim":"unit"},{"prim":"operation"}],"prim":"lambda"},{"args":[{"prim":"key_hash"}],"prim":"set"}],"prim":"or"}`,
		Want: `{"name":"","type":"union","args":[{"name":"@or_0","type":"lambda","args":[{"name":"@param","type":"unit"},{"name":"@return","type":"operation"}]},{"name":"@or_1","type":"set","args":[{"name":"@item","type":"key_hash"}]}]}`,
	},
	// nested map
	typedefTest{
		Name: "nested_map",
		Spec: `{"annots": ["%deck"],"args": [{"prim": "int"},{"args": [{"prim": "int"},{"prim": "int"}],"prim": "map"}],"prim": "map"}`,
		Want: `{"name":"deck","type":"map","args":[{"name":"@key","type":"int"},{"name":"@value","type":"map","args":[{"name":"@key","type":"int"},{"name":"@value","type":"int"}]}]}`,
	},
	// nested list (FA2)
	typedefTest{
		Name: "nested_list",
		Spec: `{"annots": ["%transfer"],"args": [{"args": [{"annots": ["%from_"],"prim": "address"},{"annots": ["%txs"],"args": [{"args": [{"annots": ["%to_"],"prim": "address"},{"args": [{"annots": ["%token_id"],"prim": "nat"},{"annots": ["%amount"],"prim": "nat"}],"prim": "pair"}],"prim": "pair"}],"prim": "list"}],"prim": "pair"}],"prim": "list"}`,
		Want: `{"name":"transfer","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"from_","type":"address"},{"name":"txs","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"to_","type":"address"},{"name":"token_id","type":"nat"},{"name":"amount","type":"nat"}]}]}]}]}`,
	},
	// right-hand pair tree
	typedefTest{
		Name: "right_hand_pair_tree",
		Spec: `{"args":[{"annots":["%tokenPool"],"prim":"nat"},{"args":[{"annots":["%xtzPool"],"prim":"mutez"},{"args":[{"annots":["%lqtTotal"],"prim":"nat"},{"args":[{"annots":["%tokenAddress"],"prim":"address"},{"annots":["%lqtAddress"],"prim":"address"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"}],"prim":"pair"}`,
		Want: `{"name":"","type":"struct","args":[{"name":"tokenPool","type":"nat"},{"name":"xtzPool","type":"mutez"},{"name":"lqtTotal","type":"nat"},{"name":"tokenAddress","type":"address"},{"name":"lqtAddress","type":"address"}]}`,
	},
}

func TestTypeRendering(t *testing.T) {
	for _, test := range typedefInfo {
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
