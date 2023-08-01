package ast

import (
	"blockwatch.cc/tzgo/micheline"
)

type Entrypoint struct {
	Name   string
	Raw    *micheline.Entrypoint `json:"-"`
	Params []*Struct
}

// Getter is a read-only Entrypoint, with a return value.
// It is implemented with TZIP-4.
type Getter struct {
	Entrypoint
	ReturnType *Struct
}
