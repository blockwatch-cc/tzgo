// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	yamlExts = map[string]struct{}{
		".yml":  {},
		".yaml": {},
	}
)

// ParseFile parses file based on path given to spec
func ParseFile[T any](fpath string) (*T, error) {
	ext := filepath.Ext(fpath)
	if _, ok := yamlExts[ext]; !ok {
		return nil, fmt.Errorf("file extension %q is not supported", ext)
	}
	buf, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	var spec T
	err = yaml.Unmarshal(buf, &spec)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}
