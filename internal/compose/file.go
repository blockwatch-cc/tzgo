// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

func ReadJsonFile[T any](name string) (*T, error) {
	name, jsonPath, hasPath := strings.Cut(name, "#")
	var val *T
	buf, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	if hasPath {
		res := gjson.GetBytes(buf, jsonPath)
		if !res.Exists() {
			return nil, fmt.Errorf("missing path %q in file %s", jsonPath, name)
		}
		buf = []byte(res.Raw)
	}
	err = json.Unmarshal(buf, &val)
	if err != nil {
		return nil, err
	}
	return val, nil
}
