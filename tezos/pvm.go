// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"fmt"
)

type PvmKind byte

const (
	PvmKindArith     PvmKind = iota // 0
	PvmKindWasm200                  // 1
	PvmKindWasm200r1                // 2 v1
	PvmKindWasm200r2                // 3 v2
	PvmKindWasm200r3                // 4 v3
	PvmKindWasm200r4                // 5 v4
	PvmKindInvalid   = 255
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
	case "wasm_2_0_0.r1":
		return PvmKindWasm200r1
	case "wasm_2_0_0.r2":
		return PvmKindWasm200r2
	case "wasm_2_0_0.r3":
		return PvmKindWasm200r3
	case "wasm_2_0_0.r4":
		return PvmKindWasm200r4
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
	case PvmKindWasm200r1:
		return "wasm_2_0_0.r1"
	case PvmKindWasm200r2:
		return "wasm_2_0_0.r2"
	case PvmKindWasm200r3:
		return "wasm_2_0_0.r3"
	case PvmKindWasm200r4:
		return "wasm_2_0_0.r4"
	default:
		return ""
	}
}
