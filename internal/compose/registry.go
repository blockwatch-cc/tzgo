// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"strings"
)

type EngineFactory func() Engine

var (
	engineRegistry = make(map[string]EngineFactory)
	lastVersion    string
)

func RegisterEngine(ver string, v EngineFactory) {
	ver = strings.ToLower(ver)
	engineRegistry[ver] = v
	lastVersion = ver
}

func HasVersion(ver string) bool {
	_, ok := engineRegistry[ver]
	return ok
}

func New(ver string) Engine {
	fn, ok := engineRegistry[ver]
	if !ok {
		return nil
	}
	return fn()
}
