// Copyright (c) 2023 Blockwatch Data Inc.
// Authors
// - jean.schmitt@ubisoft.com
// - abdul@blockwatch.cc

package main

import (
	"errors"
	"fmt"
)

func main() {
	if err := run(); err != nil {
		if errors.Is(err, errExit) {
			return
		}
		fmt.Printf("Error: %v\n", err)
	}
}

func run() error {
	if err := parseFlags(); err != nil {
		return err
	}
	if err := runCommand(); err != nil {
		return err
	}
	return nil
}
