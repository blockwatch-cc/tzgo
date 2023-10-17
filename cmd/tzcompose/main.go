// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/tezos"

	_ "blockwatch.cc/tzgo/internal/compose/alpha"
	_ "blockwatch.cc/tzgo/internal/compose/alpha/task"
)

var (
	flags      = flag.NewFlagSet(appName, flag.ContinueOnError)
	runflags   = flag.NewFlagSet("run", flag.ContinueOnError)
	cloneflags = flag.NewFlagSet("clone", flag.ContinueOnError)
	errExit    = errors.New("exit")
	errNoCmd   = errors.New("unsupported command")
	verbose    bool
	vtrace     bool
	vdebug     bool
	cmd        string = "[cmd]"

	// run config
	fpath  string
	resume bool
	rpcUrl string

	// clone config
	name                   string
	mode                   compose.CloneMode
	addr                   tezos.Address
	numOpsAfterOrigination uint
	version                string
	indexUrl               string
	outputPath             string
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", true, "be verbose")
	flags.BoolVar(&vdebug, "vv", false, "debug mode")
	flags.BoolVar(&vtrace, "vvv", false, "trace mode")

	runflags.Usage = func() {}
	runflags.StringVar(&fpath, "f", "tzcompose.yaml", "configuration `file` or path")
	runflags.StringVar(&fpath, "file", "tzcompose.yaml", "configuration `file` or path")
	runflags.BoolVar(&resume, "resume", false, "continue pipeline execution")
	runflags.StringVar(&rpcUrl, "rpc", "https://rpc.tzpro.io", "Tezos node RPC url")

	cloneflags.Usage = func() {}
	cloneflags.StringVar(&indexUrl, "index", "https://api.tzpro.io", "Tezos indexer url")
	cloneflags.StringVar(&rpcUrl, "rpc", "https://rpc.tzpro.io", "Tezos node RPC url")
	cloneflags.StringVar(&version, "version", "alpha", "compose engine version")
	cloneflags.Var(&addr, "contract", "address of the contract to clone")
	cloneflags.Var(&mode, "mode", "output mode for cloned micheline data (file, json, bin, url)")
	cloneflags.StringVar(&name, "name", "contract", "project name")
	cloneflags.StringVar(&outputPath, "out", "tzcompose.yaml", "output path for generated files")
	cloneflags.UintVar(&numOpsAfterOrigination, "n", 0, "number of operations after origination")
}

func main() {
	if err := parseFlags(); err != nil {
		if err != errExit {
			fmt.Println("Error:", err)
		}
		return
	}
	initLogging()

	if err := run(); err != nil {
		fmt.Println("Error:", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ectx := compose.NewContext(ctx)
	ectx.WithLogger(taskLog).
		WithUrl(rpcUrl).
		WithApiKey(os.Getenv("TZCOMPOSE_API_KEY")).
		WithBase(os.Getenv("TZCOMPOSE_BASE_KEY")).
		WithResume(resume)

	var err error
	switch cmd {
	case "version":
		printVersion()
	case "validate":
		err = compose.Run(ectx, fpath, compose.RunModeValidate)
	case "simulate":
		err = compose.Run(ectx, fpath, compose.RunModeSimulate)
	case "run":
		err = compose.Run(ectx, fpath, compose.RunModeExecute)
	case "clone":
		err = compose.Clone(ectx, version, compose.CloneConfig{
			Name:     name,
			Contract: addr,
			IndexUrl: indexUrl,
			NumOps:   numOpsAfterOrigination,
			Path:     outputPath,
			Mode:     mode,
		})
	default:
		err = errNoCmd
	}
	return err
}

func parseFlags() error {
	if len(os.Args) < 2 {
		printHelp()
		return errExit
	}

	n := 1
	if !strings.HasPrefix(os.Args[n], "-") {
		cmd = os.Args[n]
		n++
	}

	switch cmd {
	case "validate", "simulate", "run", "clone", "version", "[cmd]":
		// ok
	default:
		return errNoCmd
	}

	// parse global flags
	err := flags.Parse(filterFlags(flags, os.Args[n:]))
	if err != nil {
		if err == flag.ErrHelp {
			printHelp()
			return errExit
		}
	}

	// parse command flags
	switch cmd {
	case "validate", "simulate", "run":
		err = runflags.Parse(filterFlags(runflags, os.Args[2:]))
	case "clone":
		err = cloneflags.Parse(filterFlags(cloneflags, os.Args[2:]))
	}
	if err != nil {
		if err == flag.ErrHelp {
			printHelp()
			return errExit
		}
		return err
	}
	return nil
}

func filterFlags(set *flag.FlagSet, args []string) []string {
	res := make([]string, 0)
	var maybeCopyNext bool
	for _, v := range args {
		if strings.HasPrefix(v, "-") {
			f := set.Lookup(v[1:])
			if f == nil && v != "-h" {
				maybeCopyNext = false
				continue
			}
			maybeCopyNext = true
			res = append(res, v)
		} else if maybeCopyNext {
			maybeCopyNext = false
			res = append(res, v)
		}
	}
	return res
}

func printHelp() {
	fmt.Printf("(c) Copyright (c) %d Blockwatch Data Inc.\n", time.Now().Year())
	fmt.Printf("Usage:  %s %s [flags]\n", appName, cmd)
	switch cmd {
	case "validate", "simulate", "run":
		fmt.Printf("\nEnv\n")
		fmt.Println("  TZCOMPOSE_BASE_KEY  private key for base account")
		fmt.Println("  TZCOMPOSE_API_KEY   API key for RPC and index calls (optional)")
		fmt.Println("\nFlags")
		runflags.PrintDefaults()
		fmt.Println("  -h	print help and exit")
		flags.PrintDefaults()
	case "clone":
		fmt.Printf("\nEnv\n")
		fmt.Println("  TZCOMPOSE_BASE_KEY  private key for base account")
		fmt.Println("  TZCOMPOSE_API_KEY   API key for RPC and index calls (optional)")
		fmt.Println("\nFlags")
		cloneflags.PrintDefaults()
		fmt.Println("  -h	print help and exit")
		flags.PrintDefaults()
	default:
		fmt.Printf("\nCommands\n")
		fmt.Println("  clone     Clone a contract and its transactions")
		fmt.Println("  validate  Check compose file syntax and parameters")
		fmt.Println("  simulate  Simulate compose file execution")
		fmt.Println("  run       Execute compose file(s)")
		fmt.Println("  version   Print version and exit")
		fmt.Println("\nFlags")
		fmt.Println("  -h	print help and exit")
		flags.PrintDefaults()
	}
}
