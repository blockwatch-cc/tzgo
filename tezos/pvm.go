// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"fmt"
)

type PvmKind byte

const (
	PvmKindArith   PvmKind = iota // 0
	PvmKindWasm200                // 1
	PvmKindInvalid = 255
)

func (t PvmKind) IsValid() bool {
	return t != PvmKindInvalid
}

func (t *PvmKind) UnmarshalText(data []byte) error {
	v := ParsePvmKind(string(data))
	if !v.IsValid() {
		return fmt.Errorf("tezos: invalid PVM kind '%s'", string(data))
	}
	*t = v
	return nil
}

func (t PvmKind) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func ParsePvmKind(s string) PvmKind {
	switch s {
	case "arith":
		return PvmKindArith
	case "wasm_2_0_0":
		return PvmKindWasm200
	default:
		return PvmKindInvalid
	}
}

func (t PvmKind) String() string {
	switch t {
	case PvmKindArith:
		return "arith"
	case PvmKindWasm200:
		return "wasm_2_0_0"
	default:
		return ""
	}
}
