// Copyright (c) 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//

package micheline

import (
	"encoding/json"
	"testing"
)

type entryTest struct {
	Name  string
	Spec  string
	Want  string
	Calls []entryCallTest
}

type entryCallTest struct {
	Params     string
	WantName   string
	WantParams string
}

var entryInfo = []entryTest{
	// manager.tz
	{
		Name: "manager",
		Spec: `{"args":[{"args":[{"annots":["%do"],"args":[{"prim":"unit"},{"args":[{"prim":"operation"}],"prim":"list"}],"prim":"lambda"},{"annots":["%default"],"prim":"unit"}],"prim":"or"}],"prim":"parameter"}`,
		Want: `{
	        "default": {
	          "branch": "/R",
	          "name": "default",
	          "id": 1,
	          "type": [
	            {
	              "name": "",
	              "path": [],
	              "type": "unit"
	            }
	          ]
	        },
	        "do": {
	          "branch": "/L",
	          "name": "do",
	          "id": 0,
	          "type": [
                {
                  "args": [
                    {
                      "name": "@param",
                      "path": [0],
                      "type": "unit"
                    },
                    {
                      "args": [
                        {
                          "name": "@item",
                          "path": [1,0],
                          "type": "operation"
                        }
                      ],
                      "name": "@return",
                      "path": [1],
                      "type": "list"
                    }
                  ],
                  "name": "",
                  "path": [],
                  "type": "lambda"
                }
              ]
	        }
	    }`,
	},
	// FA2 (hDAO)
	{
		Name: "FA2/HDAO",
		Spec: `{"args":[{"args":[{"args":[{"args":[{"annots":["%balance_of"],"args":[{"annots":["%requests"],"args":[{"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"list"},{"annots":["%callback"],"args":[{"args":[{"args":[{"annots":["%request"],"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"},{"annots":["%balance"],"prim":"nat"}],"prim":"pair"}],"prim":"list"}],"prim":"contract"}],"prim":"pair"},{"annots":["%hDAO_batch"],"args":[{"args":[{"annots":["%amount"],"prim":"nat"},{"annots":["%to_"],"prim":"address"}],"prim":"pair"}],"prim":"list"}],"prim":"or"},{"args":[{"annots":["%mint"],"args":[{"args":[{"annots":["%address"],"prim":"address"},{"annots":["%amount"],"prim":"nat"}],"prim":"pair"},{"args":[{"annots":["%token_id"],"prim":"nat"},{"annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"map"}],"prim":"pair"}],"prim":"pair"},{"annots":["%set_administrator"],"prim":"address"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"annots":["%set_pause"],"prim":"bool"},{"annots":["%token_metadata"],"args":[{"annots":["%token_ids"],"args":[{"prim":"nat"}],"prim":"list"},{"annots":["%handler"],"args":[{"args":[{"args":[{"annots":["%token_id"],"prim":"nat"},{"annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"map"}],"prim":"pair"}],"prim":"list"},{"prim":"unit"}],"prim":"lambda"}],"prim":"pair"}],"prim":"or"},{"args":[{"annots":["%transfer"],"args":[{"args":[{"annots":["%from_"],"prim":"address"},{"annots":["%txs"],"args":[{"args":[{"annots":["%to_"],"prim":"address"},{"args":[{"annots":["%token_id"],"prim":"nat"},{"annots":["%amount"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"list"}],"prim":"pair"}],"prim":"list"},{"annots":["%update_operators"],"args":[{"args":[{"annots":["%add_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"annots":["%remove_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"or"}],"prim":"list"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"parameter"}`,
		Want: `{
			"balance_of":{"id":0,"name":"balance_of","branch":"/L/L/L","type":[{"name":"requests","path":[0],"type":"list","args":[{"name":"@item","path":[0,0],"type":"struct","args":[{"name":"owner","path":[0,0,0],"type":"address"},{"name":"token_id","path":[0,0,1],"type":"nat"}]}]},{"name":"callback","path":[1],"type":"contract","args":[{"name":"0","path":[1,0],"type":"list","args":[{"name":"@item","path":[1,0,0],"type":"struct","args":[{"name":"request","path":[1,0,0,0],"type":"struct","args":[{"name":"owner","path":[1,0,0,0,0],"type":"address"},{"name":"token_id","path":[1,0,0,0,1],"type":"nat"}]},{"name":"balance","path":[1,0,0,1],"type":"nat"}]}]}]}]},
			"balance_of":{"branch":"/L/L/L","id":0,"name":"balance_of","type":[{"args":[{"args":[{"name":"owner","path":[0,0,0],"type":"address"},{"name":"token_id","path":[0,0,1],"type":"nat"}],"name":"@item","path":[0,0],"type":"struct"}],"name":"requests","path":[0],"type":"list"},{"args":[{"args":[{"args":[{"args":[{"name":"owner","path":[1,0,0,0,0],"type":"address"},{"name":"token_id","path":[1,0,0,0,1],"type":"nat"}],"name":"request","path":[1,0,0,0],"type":"struct"},{"name":"balance","path":[1,0,0,1],"type":"nat"}],"name":"@item","path":[1,0,0],"type":"struct"}],"name":"0","path":[1,0],"type":"list"}],"name":"callback","path":[1],"type":"contract"}]},
			"hDAO_batch":{"branch":"/L/L/R","id":1,"name":"hDAO_batch","type":[{"args":[{"args":[{"name":"amount","path":[0,0],"type":"nat"},{"name":"to_","path":[0,1],"type":"address"}],"name":"@item","path":[0],"type":"struct"}],"name":"","path":[],"type":"list"}]},
			"mint":{"branch":"/L/R/L","id":2,"name":"mint","type":[{"name":"address","path":[0,0],"type":"address"},{"name":"amount","path":[0,1],"type":"nat"},{"name":"token_id","path":[1,0],"type":"nat"},{"args":[{"name":"@key","path":[0],"type":"string"},{"name":"@value","path":[1],"type":"bytes"}],"name":"token_info","path":[1,1],"type":"map"}]},
			"set_administrator":{"branch":"/L/R/R","id":3,"name":"set_administrator","type":[{"name":"","path":[],"type":"address"}]},"set_pause":{"branch":"/R/L/L","id":4,"name":"set_pause","type":[{"name":"","path":[],"type":"bool"}]},
			"token_metadata":{"branch":"/R/L/R","id":5,"name":"token_metadata","type":[{"args":[{"name":"@item","path":[0,0],"type":"nat"}],"name":"token_ids","path":[0],"type":"list"},{"args":[{"args":[{"args":[{"name":"token_id","path":[0,0,0],"type":"nat"},{"args":[{"name":"@key","path":[0],"type":"string"},{"name":"@value","path":[1],"type":"bytes"}],"name":"token_info","path":[0,0,1],"type":"map"}],"name":"@item","path":[0,0],"type":"struct"}],"name":"@param","path":[0],"type":"list"},{"name":"@return","path":[1],"type":"unit"}],"name":"handler","path":[1],"type":"lambda"}]},
			"transfer":{"branch":"/R/R/L","id":6,"name":"transfer","type":[{"args":[{"args":[{"name":"from_","path":[0,0],"type":"address"},{"args":[{"args":[{"name":"to_","path":[0,1,0,0],"type":"address"},{"name":"token_id","path":[0,1,0,1,0],"type":"nat"},{"name":"amount","path":[0,1,0,1,1],"type":"nat"}],"name":"@item","path":[0,1,0],"type":"struct"}],"name":"txs","path":[0,1],"type":"list"}],"name":"@item","path":[0],"type":"struct"}],"name":"","path":[],"type":"list"}]},
			"update_operators":{"branch":"/R/R/R","id":7,"name":"update_operators","type":[{"args":[{"args":[{"args":[{"name":"owner","path":[0,0,0],"type":"address"},{"name":"operator","path":[0,0,1,0],"type":"address"},{"name":"token_id","path":[0,0,1,1],"type":"nat"}],"name":"add_operator","path":[0,0],"type":"struct"},{"args":[{"name":"owner","path":[0,1,0],"type":"address"},{"name":"operator","path":[0,1,1,0],"type":"address"},{"name":"token_id","path":[0,1,1,1],"type":"nat"}],"name":"remove_operator","path":[0,1],"type":"struct"}],"name":"@item","path":[0],"type":"union"}],"name":"","path":[],"type":"list"}]}
		}`,
	},

	// single option, no T_OR
	{
		Name: "single option",
		Spec: `{"prim":"parameter","args":[{"prim":"option","args":[{"prim":"address"}]}]}`,
		Want: `{
	          "default": {
	          	"branch": "",
	          	"name": "default",
	          	"id": 0,
	          	"type": [{"name":"","path":[],"type":"address","optional":true}]
	          }
	      }`,
	},

	// complex nested T_OR with intermediate names (why is this even possible? - what a mess)
	{
		Name: "nested T_OR",
		Spec: `{"args":[{"args":[{"annots":["%default"],"args":[{"args":[{"args":[{"annots":["%accept_administrator"],"prim":"unit"},{"annots":["%balance_of"],"args":[{"annots":["%requests"],"args":[{"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"list"},{"annots":["%callback"],"args":[{"args":[{"args":[{"annots":["%request"],"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"},{"annots":["%balance"],"prim":"nat"}],"prim":"pair"}],"prim":"list"}],"prim":"contract"}],"prim":"pair"}],"prim":"or"},{"args":[{"annots":["%burn"],"args":[{"args":[{"annots":["%from_"],"prim":"address"},{"args":[{"annots":["%amount"],"prim":"nat"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"list"},{"annots":["%mint"],"args":[{"args":[{"annots":["%to_"],"prim":"address"},{"args":[{"annots":["%amount"],"prim":"nat"},{"annots":["%token"],"args":[{"annots":["%new"],"args":[{"annots":["%metadata"],"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"map"},{"annots":["%royalties"],"args":[{"prim":"address"},{"prim":"nat"}],"prim":"map"}],"prim":"pair"},{"annots":["%existing"],"prim":"nat"}],"prim":"or"}],"prim":"pair"}],"prim":"pair"}],"prim":"list"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"annots":["%set_metadata"],"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"big_map"},{"annots":["%transfer"],"args":[{"args":[{"annots":["%from_"],"prim":"address"},{"annots":["%txs"],"args":[{"args":[{"annots":["%to_"],"prim":"address"},{"args":[{"annots":["%token_id"],"prim":"nat"},{"annots":["%amount"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"list"}],"prim":"pair"}],"prim":"list"}],"prim":"or"},{"args":[{"annots":["%transfer_administrator"],"prim":"address"},{"args":[{"annots":["%update_adhoc_operators"],"args":[{"annots":["%add_adhoc_operators"],"args":[{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"set"},{"annots":["%clear_adhoc_operators"],"prim":"unit"}],"prim":"or"},{"annots":["%update_operators"],"args":[{"args":[{"annots":["%add_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"annots":["%remove_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"or"}],"prim":"list"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"or"},{"annots":["%set_parent"],"prim":"address"}],"prim":"or"}],"prim":"parameter"}`,
		Want: `{
			"accept_administrator": {"branch": "/L/L/L/L","id": 0,"name": "accept_administrator","type": [{"name": "","path": [],"type": "unit"}]},
			"add_adhoc_operators": {"branch": "/L/R/R/R/L/L","id": 7,"name": "add_adhoc_operators","type": [{"args": [{"args": [{"name": "operator","path": [0,0],"type": "address"},{"name": "token_id","path": [0,1],"type": "nat"}],"name": "@item","path": [0],"type": "struct"}],"name": "","path": [],"type": "set"}]},
			"balance_of": {"branch": "/L/L/L/R","id": 1,"name": "balance_of","type": [{"args": [{"args": [{"name": "owner","path": [0,0,0],"type": "address"},{"name": "token_id","path": [0,0,1],"type": "nat"}],"name": "@item","path": [0,0],"type": "struct"}],"name": "requests","path": [0],"type": "list"},{"args": [{"args": [{"args": [{"args": [{"name": "owner","path": [1,0,0,0,0],"type": "address"},{"name": "token_id","path": [1,0,0,0,1],"type": "nat"}],"name": "request","path": [1,0,0,0],"type": "struct"},{"name": "balance","path": [1,0,0,1],"type": "nat"}],"name": "@item","path": [1,0,0],"type": "struct"}],"name": "0","path": [1,0],"type": "list"}],"name": "callback","path": [1],"type": "contract"}]},
			"burn": {"branch": "/L/L/R/L","id": 2,"name": "burn","type": [{"args": [{"args": [{"name": "from_","path": [0,0],"type": "address"},{"name": "amount","path": [0,1,0],"type": "nat"},{"name": "token_id","path": [0,1,1],"type": "nat"}],"name": "@item","path": [0],"type": "struct"}],"name": "","path": [],"type": "list"}]},
			"clear_adhoc_operators": {"branch": "/L/R/R/R/L/R","id": 8,"name": "clear_adhoc_operators","type": [{"name": "","path": [],"type": "unit"}]},
			"mint": {"branch": "/L/L/R/R","id": 3,"name": "mint","type": [{"args": [{"args": [{"name": "to_","path": [0,0],"type": "address"},{"name": "amount","path": [0,1,0],"type": "nat"},{"args": [{"args": [{"args": [{"name": "@key","path": [0],"type": "string"},{"name": "@value","path": [1],"type": "bytes"}],"name": "metadata","path": [0,1,1,0,0],"type": "map"},{"args": [{"name": "@key","path": [0],"type": "address"},{"name": "@value","path": [1],"type": "nat"}],"name": "royalties","path": [0,1,1,0,1],"type": "map"}],"name": "new","path": [0,1,1,0],"type": "struct"},{"name": "existing","path": [0,1,1,1],"type": "nat"}],"name": "token","path": [0,1,1],"type": "union"}],"name": "@item","path": [0],"type": "struct"}],"name": "","path": [],"type": "list"}]},
			"set_metadata": {"branch": "/L/R/L/L","id": 4,"name": "set_metadata","type": [{"name": "@key","path": [0],"type": "string"},{"name": "@value","path": [1],"type": "bytes"}]},
			"set_parent": {"branch": "/R","id": 10,"name": "set_parent","type": [{"name": "","path": [],"type": "address"}]},
			"transfer": {"branch": "/L/R/L/R","id": 5,"name": "transfer","type": [{"args": [{"args": [{"name": "from_","path": [0,0],"type": "address"},{"args": [{"args": [{"name": "to_","path": [0,1,0,0],"type": "address"},{"name": "token_id","path": [0,1,0,1,0],"type": "nat"},{"name": "amount","path": [0,1,0,1,1],"type": "nat"}],"name": "@item","path": [0,1,0],"type": "struct"}],"name": "txs","path": [0,1],"type": "list"}],"name": "@item","path": [0],"type": "struct"}],"name": "","path": [],"type": "list"}]},
			"transfer_administrator": {"branch": "/L/R/R/L","id": 6,"name": "transfer_administrator","type": [{"name": "","path": [],"type": "address"}]},
			"update_operators": {"branch": "/L/R/R/R/R","id": 9,"name": "update_operators","type": [{"args": [{"args": [{"args": [{"name": "owner","path": [0,0,0],"type": "address"},{"name": "operator","path": [0,0,1,0],"type": "address"},{"name": "token_id","path": [0,0,1,1],"type": "nat"}],"name": "add_operator","path": [0,0],"type": "struct"},{"args": [{"name": "owner","path": [0,1,0],"type": "address"},{"name": "operator","path": [0,1,1,0],"type": "address"},{"name": "token_id","path": [0,1,1,1],"type": "nat"}],"name": "remove_operator","path": [0,1],"type": "struct"}],"name": "@item","path": [0],"type": "union"}],"name": "","path": [],"type": "list"}]}
		}`,
		Calls: []entryCallTest{
			{
				Params:     `{"entrypoint":"update_adhoc_operators","value":{"prim":"Left","args":[[{"prim":"Pair","args":[{"string":"KT1977zpPmwDqiDRqoGS47HRhQUaxcQigVYc"},{"int":"0"}]}]]}}`,
				WantName:   `add_adhoc_operators`,
				WantParams: `[{"prim":"Pair","args":[{"string":"KT1977zpPmwDqiDRqoGS47HRhQUaxcQigVYc"},{"int":"0"}]}]`,
			},
		},
	},

	// quipuswap %use
	{
		Name: "Quipu %use",
		Spec: `{"args":[{"args":[{"args":[{"args":[{"annots":["%balance_of"],"args":[{"annots":["%requests"],"args":[{"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"list"},{"annots":["%callback"],"args":[{"args":[{"args":[{"annots":["%request"],"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"},{"annots":["%balance"],"prim":"nat"}],"prim":"pair"}],"prim":"list"}],"prim":"contract"}],"prim":"pair"},{"annots":["%default"],"prim":"unit"}],"prim":"or"},{"args":[{"annots":["%get_reserves"],"args":[{"args":[{"prim":"nat"},{"prim":"nat"}],"prim":"pair"}],"prim":"contract"},{"annots":["%transfer"],"args":[{"args":[{"annots":["%from_"],"prim":"address"},{"annots":["%txs"],"args":[{"annots":[""],"args":[{"annots":["%to_"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"},{"annots":["%amount"],"prim":"nat"}],"prim":"pair"}],"prim":"list"}],"prim":"pair"}],"prim":"list"}],"prim":"or"}],"prim":"or"},{"args":[{"annots":["%update_operators"],"args":[{"args":[{"annots":["%add_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"},{"annots":["%remove_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"or"}],"prim":"list"},{"annots":["%use"],"args":[{"args":[{"args":[{"annots":["%divestLiquidity"],"args":[{"args":[{"annots":["%min_tez"],"prim":"nat"},{"annots":["%min_tokens"],"prim":"nat"}],"prim":"pair"},{"annots":["%shares"],"prim":"nat"}],"prim":"pair"},{"annots":["%initializeExchange"],"prim":"nat"}],"prim":"or"},{"args":[{"annots":["%investLiquidity"],"prim":"nat"},{"annots":["%tezToTokenPayment"],"args":[{"annots":["%min_out"],"prim":"nat"},{"annots":["%receiver"],"prim":"address"}],"prim":"pair"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"annots":["%tokenToTezPayment"],"args":[{"args":[{"annots":["%amount"],"prim":"nat"},{"annots":["%min_out"],"prim":"nat"}],"prim":"pair"},{"annots":["%receiver"],"prim":"address"}],"prim":"pair"},{"annots":["%veto"],"args":[{"annots":["%value"],"prim":"nat"},{"annots":["%voter"],"prim":"address"}],"prim":"pair"}],"prim":"or"},{"args":[{"annots":["%vote"],"args":[{"args":[{"annots":["%candidate"],"prim":"key_hash"},{"annots":["%value"],"prim":"nat"}],"prim":"pair"},{"annots":["%voter"],"prim":"address"}],"prim":"pair"},{"annots":["%withdrawProfit"],"prim":"address"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"parameter"}`,
		Want: `{
			"balance_of": {"branch": "/L/L/L","id": 0,"name": "balance_of","type": [{"args": [{"args": [{"name": "owner","path": [0,0,0],"type": "address"},{"name": "token_id","path": [0,0,1],"type": "nat"}],"name": "@item","path": [0,0],"type": "struct"}],"name": "requests","path": [0],"type": "list"},{"args": [{"args": [{"args": [{"args": [{"name": "owner","path": [1,0,0,0,0],"type": "address"},{"name": "token_id","path": [1,0,0,0,1],"type": "nat"}],"name": "request","path": [1,0,0,0],"type": "struct"},{"name": "balance","path": [1,0,0,1],"type": "nat"}],"name": "@item","path": [1,0,0],"type": "struct"}],"name": "0","path": [1,0],"type": "list"}],"name": "callback","path": [1],"type": "contract"}]},
			"default": {"branch": "/L/L/R","id": 1,"name": "default","type": [{"name": "","path": [],"type": "unit"}]},
			"divestLiquidity": {"branch": "/R/R/L/L/L","id": 5,"name": "divestLiquidity","type": [{"name": "min_tez","path": [0,0],"type": "nat"},{"name": "min_tokens","path": [0,1],"type": "nat"},{"name": "shares","path": [1],"type": "nat"}]},
			"get_reserves": {"branch": "/L/R/L","id": 2,"name": "get_reserves","type": [{"args": [{"args": [{"name": "0","path": [0,0],"type": "nat"},{"name": "1","path": [0,1],"type": "nat"}],"name": "0","path": [0],"type": "struct"}],"name": "","path": [],"type": "contract"}]},
			"initializeExchange": {"branch": "/R/R/L/L/R","id": 6,"name": "initializeExchange","type": [{"name": "","path": [],"type": "nat"}]},
			"investLiquidity": {"branch": "/R/R/L/R/L","id": 7,"name": "investLiquidity","type": [{"name": "","path": [],"type": "nat"}]},
			"tezToTokenPayment": {"branch": "/R/R/L/R/R","id": 8,"name": "tezToTokenPayment","type": [{"name": "min_out","path": [0],"type": "nat"},{"name": "receiver","path": [1],"type": "address"}]},
			"tokenToTezPayment": {"branch": "/R/R/R/L/L","id": 9,"name": "tokenToTezPayment","type": [{"name": "amount","path": [0,0],"type": "nat"},{"name": "min_out","path": [0,1],"type": "nat"},{"name": "receiver","path": [1],"type": "address"}]},
			"transfer": {"branch": "/L/R/R","id": 3,"name": "transfer","type": [{"args": [{"args": [{"name": "from_","path": [0,0],"type": "address"},{"args": [{"args": [{"name": "to_","path": [0,1,0,0],"type": "address"},{"name": "token_id","path": [0,1,0,1],"type": "nat"},{"name": "amount","path": [0,1,0,2],"type": "nat"}],"name": "@item","path": [0,1,0],"type": "struct"}],"name": "txs","path": [0,1],"type": "list"}],"name": "@item","path": [0],"type": "struct"}],"name": "","path": [],"type": "list"}]},
			"update_operators": {"branch": "/R/L","id": 4,"name": "update_operators","type": [{"args": [{"args": [{"args": [{"name": "owner","path": [0,0,0],"type": "address"},{"name": "operator","path": [0,0,1],"type": "address"},{"name": "token_id","path": [0,0,2],"type": "nat"}],"name": "add_operator","path": [0,0],"type": "struct"},{"args": [{"name": "owner","path": [0,1,0],"type": "address"},{"name": "operator","path": [0,1,1],"type": "address"},{"name": "token_id","path": [0,1,2],"type": "nat"}],"name": "remove_operator","path": [0,1],"type": "struct"}],"name": "@item","path": [0],"type": "union"}],"name": "","path": [],"type": "list"}]},
			"veto": {"branch": "/R/R/R/L/R","id": 10,"name": "veto","type": [{"name": "value","path": [0],"type": "nat"},{"name": "voter","path": [1],"type": "address"}]},
			"vote": {"branch": "/R/R/R/R/L","id": 11,"name": "vote","type": [{"name": "candidate","path": [0,0],"type": "key_hash"},{"name": "value","path": [0,1],"type": "nat"},{"name": "voter","path": [1],"type": "address"}]},
			"withdrawProfit": {"branch": "/R/R/R/R/R","id": 12,"name": "withdrawProfit","type": [{"name": "","path": [],"type": "address"}]}
		}`,
		Calls: []entryCallTest{
			{
				Params:     `{"entrypoint":"use","value":{"args":[{"args":[{"args":[{"args":[{"int":"3310026693359"},{"string":"tz1UqpNLoPRpAPKCTsgXyBRitTdpKJWEjVKT"}],"prim":"Pair"}],"prim":"Right"}],"prim":"Right"}],"prim":"Left"}}`,
				WantName:   `tezToTokenPayment`,
				WantParams: `{"args":[{"int":"3310026693359"},{"string":"tz1UqpNLoPRpAPKCTsgXyBRitTdpKJWEjVKT"}],"prim":"Pair"}`,
			},
		},
	},

	// TF vesting contracts
	{
		Name: "TF vesting",
		Spec: `{"args":[{"args":[{"annots":["%Action"],"args":[{"annots":["%action_input"],"args":[{"args":[{"annots":["%Transfer"],"args":[{"annots":["%dest"],"args":[{"prim":"unit"}],"prim":"contract"},{"annots":["%transfer_amount"],"prim":"mutez"}],"prim":"pair"},{"annots":["%Set_pour"],"args":[{"args":[{"annots":["%pour_dest"],"args":[{"prim":"unit"}],"prim":"contract"},{"annots":["%pour_authorizer"],"prim":"key"}],"prim":"pair"}],"prim":"option"}],"prim":"or"},{"args":[{"annots":["%Set_keys"],"args":[{"annots":["%key_groups"],"args":[{"args":[{"annots":["%signatories"],"args":[{"prim":"key"}],"prim":"list"},{"annots":["%group_threshold"],"prim":"nat"}],"prim":"pair"}],"prim":"list"},{"annots":["%overall_threshold"],"prim":"nat"}],"prim":"pair"},{"annots":["%Set_delegate"],"args":[{"annots":["%new_delegate"],"prim":"key_hash"}],"prim":"option"}],"prim":"or"}],"prim":"or"},{"annots":["%signatures"],"args":[{"args":[{"args":[{"prim":"signature"}],"prim":"option"}],"prim":"list"}],"prim":"list"}],"prim":"pair"},{"args":[{"annots":["%Pour"],"args":[{"annots":["%pour_auth"],"prim":"signature"},{"annots":["%pour_amount"],"prim":"mutez"}],"prim":"pair"}],"prim":"option"}],"prim":"or"}],"prim":"parameter"}`,
		Want: `{
			"@entrypoint_1": {"branch": "/R","id": 1,"name": "@entrypoint_1","type": [{"name": "pour_auth","path": [0,0],"type": "signature"},{"name": "pour_amount","path": [0,1],"type": "mutez"}]},
			"Action": {"branch": "/L","id": 0,"name": "Action","type": [{"args": [{"args": [{"args": [{"name": "0","path": [0,0,0,0,0],"type": "unit"}],"name": "dest","path": [0,0,0,0],"type": "contract"},{"name": "transfer_amount","path": [0,0,0,1],"type": "mutez"}],"name": "Transfer","path": [0,0,0],"type": "struct"},{"args": [{"args": [{"name": "0","path": [0,0,1,0,0,0],"type": "unit"}],"name": "pour_dest","path": [0,0,1,0,0],"type": "contract"},{"name": "pour_authorizer","path": [0,0,1,0,1],"type": "key"}],"name": "Set_pour","optional": true,"path": [0,0,1],"type": "struct"},{"args": [{"args": [{"args": [{"args": [{"name": "@item","path": [0,1,0,0,0,0,0],"type": "key"}],"name": "signatories","path": [0,1,0,0,0,0],"type": "list"},{"name": "group_threshold","path": [0,1,0,0,0,1],"type": "nat"}],"name": "@item","path": [0,1,0,0,0],"type": "struct"}],"name": "key_groups","path": [0,1,0,0],"type": "list"},{"name": "overall_threshold","path": [0,1,0,1],"type": "nat"}],"name": "Set_keys","path": [0,1,0],"type": "struct"},{"name": "Set_delegate","optional": true,"path": [0,1,1],"type": "key_hash"}],"name": "action_input","path": [0],"type": "union"},{"args": [{"args": [{"name": "@item","optional": true,"path": [1,0,0],"type": "signature"}],"name": "@item","path": [1,0],"type": "list"}],"name": "signatures","path": [1],"type": "list"}]}
		}`,
		Calls: []entryCallTest{
			{
				Params:     `{"args":[{"args":[{"args":[{"string":"edsigu4LoC5DEwZ749VX4gtpgkieJcmKEWmuqbDn9Yj1MJ177xvZ7J3AkT68hYVCF8gHBMfZX6oDSJwTsD5VmKdtTTkWSuaJ7mh"},{"int":"199041301565"}],"prim":"Pair"}],"prim":"Some"}],"prim":"Right"}`,
				WantName:   `@entrypoint_1`,
				WantParams: `{"args":[{"args":[{"string":"edsigu4LoC5DEwZ749VX4gtpgkieJcmKEWmuqbDn9Yj1MJ177xvZ7J3AkT68hYVCF8gHBMfZX6oDSJwTsD5VmKdtTTkWSuaJ7mh"},{"int":"199041301565"}],"prim":"Pair"}],"prim":"Some"}`,
			},
		},
	},

	// Ticket
	{
		Name: "Ticket save",
		Spec: `{"prim":"parameter","args":[{"prim":"or","args":[{"prim":"ticket","annots":["%save"],"args":[{"prim":"string"}]},{"prim":"address","annots":["%send"]}]}]}`,
		Want: `{
			"save": {"branch": "/L","id": 0,"name": "save","type": [{"args": [{"name": "@value","path": [0],"type": "string"}],"name": "","path": [],"type": "ticket"}]},
			"send": {"branch": "/R","id": 1,"name": "send","type": [{"name": "","path": [],"type": "address"}]}
		}`,
		Calls: []entryCallTest{
			{
				Params:     `{"entrypoint":"save","value":{"prim":"Pair","args":[{"bytes":"01841548c7ac7da287a25f527e5838e5f714e9449000"},{"prim":"Pair","args":[{"string":"Ticket"},{"int":"1"}]}]}}`,
				WantName:   `save`,
				WantParams: `{"prim":"Pair","args":[{"bytes":"01841548c7ac7da287a25f527e5838e5f714e9449000"},{"prim":"Pair","args":[{"string":"Ticket"},{"int":"1"}]}]}`,
			},
		},
	},
}

func TestEntrypointRendering(t *testing.T) {
	for _, test := range entryInfo {
		t.Run(test.Name, func(T *testing.T) {
			script := NewScript()
			err := script.Code.Param.UnmarshalJSON([]byte(test.Spec))
			if err != nil {
				T.Fatalf("unmarshal: %v", err)
			}
			eps, err := script.Entrypoints(false)
			if err != nil {
				T.Errorf("entrypoint list: %v", err)
			}
			have, err := json.Marshal(eps)
			if err != nil {
				T.Fatalf("render: %v", err)
			}
			if !jsonDiff(T, have, []byte(test.Want)) {
				T.Error("entrypoint type detection mismatch, see log for details")
			}
			for i, call := range test.Calls {
				var params Parameters
				if err := json.Unmarshal([]byte(call.Params), &params); err != nil {
					T.Fatalf("call %d unmarshal: %v", i, err)
				}
				ep, prim, err := params.MapEntrypoint(script.ParamType())
				if err != nil {
					T.Fatalf("cannot detect entrypoint %s: %v", params.Entrypoint, err)
				}
				if have, want := ep.Name, call.WantName; have != want {
					T.Errorf("mismatched entrypoint have=%s want=%s", have, want)
				}
				have, err := json.Marshal(prim)
				if err != nil {
					T.Fatalf("render: %v", err)
				}
				if !jsonDiff(T, have, []byte(call.WantParams)) {
					T.Error("parameter extraction mismatch, see log for details")
				}
			}
		})
	}
}
