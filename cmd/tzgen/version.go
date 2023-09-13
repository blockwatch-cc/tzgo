// Copyright (c) 2023 Blockwatch Data Inc.
// Authors
// - jean.schmitt@ubisoft.com
// - abdul@blockwatch.cc

package main

import (
	"fmt"
	"runtime"
)

var (
	appName        = "tzgen"
	version string = "v0.1"
	commit  string = "dev"
)

func printVersion() {
	fmt.Printf("%s, version %s (%s)\n", appName, version, commit)
	fmt.Printf("Go version: %s\n", runtime.Version())
}
