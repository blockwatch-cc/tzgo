// Copyright (c) 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//

package micheline

import (
	"encoding/json"
	"testing"
)

type entryTest struct {
	Name string
	Spec string
	Want string
}

var entryInfo = []entryTest{
	// manager.tz
	entryTest{
		Name: "manager",
		Spec: `{"args":[{"args":[{"annots":["%do"],"args":[{"prim":"unit"},{"args":[{"prim":"operation"}],"prim":"list"}],"prim":"lambda"},{"annots":["%default"],"prim":"unit"}],"prim":"or"}],"prim":"parameter"}`,
		Want: `{
	        "default": {
	          "branch": "/R",
	          "call": "default",
	          "id": 1,
	          "type": [
	            {
	              "name": "",
	              "type": "unit"
	            }
	          ]
	        },
	        "do": {
	          "branch": "/L",
	          "call": "do",
	          "id": 0,
	          "type": [
                {
                  "args": [
                    {
                      "name": "@param",
                      "type": "unit"
                    },
                    {
                      "args": [
                        {
                          "name": "@item",
                          "type": "operation"
                        }
                      ],
                      "name": "@return",
                      "type": "list"
                    }
                  ],
                  "name": "",
                  "type": "lambda"
                }
              ]
	        }
	    }`,
	},
	// FA2 (hDAO)
	entryTest{
		Name: "FA2/HDAO",
		Spec: `{"args":[{"args":[{"args":[{"args":[{"annots":["%balance_of"],"args":[{"annots":["%requests"],"args":[{"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"list"},{"annots":["%callback"],"args":[{"args":[{"args":[{"annots":["%request"],"args":[{"annots":["%owner"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"},{"annots":["%balance"],"prim":"nat"}],"prim":"pair"}],"prim":"list"}],"prim":"contract"}],"prim":"pair"},{"annots":["%hDAO_batch"],"args":[{"args":[{"annots":["%amount"],"prim":"nat"},{"annots":["%to_"],"prim":"address"}],"prim":"pair"}],"prim":"list"}],"prim":"or"},{"args":[{"annots":["%mint"],"args":[{"args":[{"annots":["%address"],"prim":"address"},{"annots":["%amount"],"prim":"nat"}],"prim":"pair"},{"args":[{"annots":["%token_id"],"prim":"nat"},{"annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"map"}],"prim":"pair"}],"prim":"pair"},{"annots":["%set_administrator"],"prim":"address"}],"prim":"or"}],"prim":"or"},{"args":[{"args":[{"annots":["%set_pause"],"prim":"bool"},{"annots":["%token_metadata"],"args":[{"annots":["%token_ids"],"args":[{"prim":"nat"}],"prim":"list"},{"annots":["%handler"],"args":[{"args":[{"args":[{"annots":["%token_id"],"prim":"nat"},{"annots":["%token_info"],"args":[{"prim":"string"},{"prim":"bytes"}],"prim":"map"}],"prim":"pair"}],"prim":"list"},{"prim":"unit"}],"prim":"lambda"}],"prim":"pair"}],"prim":"or"},{"args":[{"annots":["%transfer"],"args":[{"args":[{"annots":["%from_"],"prim":"address"},{"annots":["%txs"],"args":[{"args":[{"annots":["%to_"],"prim":"address"},{"args":[{"annots":["%token_id"],"prim":"nat"},{"annots":["%amount"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"list"}],"prim":"pair"}],"prim":"list"},{"annots":["%update_operators"],"args":[{"args":[{"annots":["%add_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"},{"annots":["%remove_operator"],"args":[{"annots":["%owner"],"prim":"address"},{"args":[{"annots":["%operator"],"prim":"address"},{"annots":["%token_id"],"prim":"nat"}],"prim":"pair"}],"prim":"pair"}],"prim":"or"}],"prim":"list"}],"prim":"or"}],"prim":"or"}],"prim":"or"}],"prim":"parameter"}`,
		Want: `{
			"balance_of":{"id":0,"call":"balance_of","branch":"/L/L/L","type":[{"name":"requests","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"owner","type":"address"},{"name":"token_id","type":"nat"}]}]},{"name":"callback","type":"contract","args":[{"name":"0","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"request","type":"struct","args":[{"name":"owner","type":"address"},{"name":"token_id","type":"nat"}]},{"name":"balance","type":"nat"}]}]}]}]},
			"hDAO_batch":{"id":1,"call":"hDAO_batch","branch":"/L/L/R","type":[{"name":"","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"amount","type":"nat"},{"name":"to_","type":"address"}]}]}]},
			"mint":{"id":2,"call":"mint","branch":"/L/R/L","type":[{"name":"address","type":"address"},{"name":"amount","type":"nat"},{"name":"token_id","type":"nat"},{"name":"token_info","type":"map","args":[{"name":"@key","type":"string"},{"name":"@value","type":"bytes"}]}]},
			"set_administrator":{"id":3,"call":"set_administrator","branch":"/L/R/R","type":[{"name":"","type":"address"}]},
			"set_pause":{"id":4,"call":"set_pause","branch":"/R/L/L","type":[{"name":"","type":"bool"}]},
			"token_metadata":{"id":5,"call":"token_metadata","branch":"/R/L/R","type":[{"name":"token_ids","type":"list","args":[{"name":"@item","type":"nat"}]},{"name":"handler","type":"lambda","args":[{"name":"@param","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"token_id","type":"nat"},{"name":"token_info","type":"map","args":[{"name":"@key","type":"string"},{"name":"@value","type":"bytes"}]}]}]},{"name":"@return","type":"unit"}]}]},
			"transfer":{"id":6,"call":"transfer","branch":"/R/R/L","type":[{"name":"","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"from_","type":"address"},{"name":"txs","type":"list","args":[{"name":"@item","type":"struct","args":[{"name":"to_","type":"address"},{"name":"token_id","type":"nat"},{"name":"amount","type":"nat"}]}]}]}]}]},
			"update_operators":{"id":7,"call":"update_operators","branch":"/R/R/R","type":[{"name":"","type":"list","args":[{"name":"@item","type":"union","args":[{"name":"add_operator","type":"struct","args":[{"name":"owner","type":"address"},{"name":"operator","type":"address"},{"name":"token_id","type":"nat"}]},{"name":"remove_operator","type":"struct","args":[{"name":"owner","type":"address"},{"name":"operator","type":"address"},{"name":"token_id","type":"nat"}]}]}]}]}
		}`,
	},

	// single option, no T_OR
	entryTest{
		Name: "single option",
		Spec: `{"prim":"parameter","args":[{"prim":"option","args":[{"prim":"address"}]}]}`,
		Want: `{
            "default": {
            	"branch": "",
            	"call": "default",
            	"id": 0,
            	"type": [{"name":"","type":"address","optional":true}]
            }
        }`,
	},
}

func TestEntrypointRendering(t *testing.T) {
	for _, test := range entryInfo {
		t.Run(test.Name, func(T *testing.T) {
			script := NewScript()
			err := script.Code.Param.UnmarshalJSON([]byte(test.Spec))
			if err != nil {
				T.Errorf("unmarshal error: %v", err)
			}
			eps, err := script.Entrypoints(false)
			if err != nil {
				T.Errorf("entrypoint list error: %v", err)
			}
			have, err := json.Marshal(eps)
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
