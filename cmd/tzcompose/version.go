// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package main

import (
	"fmt"
	"runtime"
	"time"
)

var (
	appName           = "tzcompose"
	appVersion string = "v0.1"
)

func printVersion() {
	fmt.Printf("(c) Copyright %d Blockwatch Data Inc.\n", time.Now().Year())
	fmt.Printf("%s, version %s\n", appName, appVersion)
	fmt.Printf("Go version: %s\n", runtime.Version())
}
