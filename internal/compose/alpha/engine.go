// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package alpha

import (
	"blockwatch.cc/tzgo/internal/compose"
)

var (
	_ compose.Engine = (*Engine)(nil)

	VERSION = "alpha"
)

func init() {
	compose.RegisterEngine(VERSION, New)
}

type Engine struct{}

func New() compose.Engine {
	return &Engine{}
}
