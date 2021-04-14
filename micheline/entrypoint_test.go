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
	              "name": "_param",
	              "type": "unit"
	            },
	            {
	              "args": [
	                {
	                  "name": "_item",
	                  "type": "operation"
	                }
	              ],
	              "name": "_return",
	              "type": "list"
	            }
	          ]
	        }
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
