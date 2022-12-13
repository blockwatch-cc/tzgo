//go:build !wasm

// Copyright (c) 2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import "reflect"

func GetFieldNameAndTag(typ reflect.Type, index []int) *fieldNameAndTag {
	f := typ.FieldByIndex(index)
	return &fieldNameAndTag{
		Name: f.Name,
		Tag:  f.Tag.Get(tagName),
	}
}
