//go:build wasm

// Copyright (c) 2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import "reflect"

func GetFieldNameAndTag(typ reflect.Type, index []int) *fieldNameAndTag {
	return &fieldNameAndTag{
		Name: "unsupported in tinygo",
		Tag:  "unsupported in tinygo",
	}
}
