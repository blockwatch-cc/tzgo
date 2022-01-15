// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// RPC examples
//
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
)

var (
	flags   = flag.NewFlagSet("fa12", flag.ContinueOnError)
	verbose bool
	node    string
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
	flags.StringVar(&node, "node", "https://rpc.tzstats.com", "Tezos node URL")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Usage: fa12 [args] <block> <pos>")
			fmt.Println("\n  Decodes FA1.2 transfer info from transaction.")
			fmt.Println("\nArguments")
			flags.PrintDefaults()
			os.Exit(0)
		}
		log.Fatal("Error:", err)
	}

	if err := run(); err != nil {
		log.Fatal("Error:", err)
	}
}

func run() error {
	if flags.NArg() < 2 {
		return fmt.Errorf("Argument required")
	}

	switch {
	case verbose:
		log.SetLevel(log.LevelInfo)
	default:
		log.SetLevel(log.LevelWarn)
	}
	rpc.UseLogger(log.Log)

	hash, err := tezos.ParseBlockHash(flags.Arg(0))
	if err != nil {
		return err
	}
	op_n, err := strconv.ParseInt(flags.Arg(1), 10, 64)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := rpc.NewClient(node, nil)
	if err != nil {
		return err
	}

	b, err := c.GetBlock(ctx, hash)
	if err != nil {
		return err
	}

	tx := b.Operations[3][op_n].Contents[0].(*rpc.Transaction)

	// you need the contract's script for type info
	script, err := c.GetContractScript(ctx, tx.Destination)
	if err != nil {
		return err
	}

	// unwind params for nested entrypoints
	ep, prim, err := tx.Parameters.MapEntrypoint(script.ParamType())
	if err != nil {
		return err
	}

	// convert Micheline params into human-readable form
	val := micheline.NewValue(ep.Type(), prim)

	// use Value interface to access data, you have multiple options
	// 1/ get a decoded `map[string]interface{}`
	m, err := val.Map()
	if err != nil {
		return err
	}

	buf, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("Map=", string(buf))
	fmt.Printf("Value=%s %[1]T\n", m.(map[string]interface{})["transfer"].(map[string]interface{})["value"])

	// 2/ access individual fields (ok is true when the field exists and
	//    has the correct type)
	from, ok := val.GetAddress("transfer.from")
	if !ok {
		return fmt.Errorf("No from param")
	}
	fmt.Println("Sent from", from)

	// 3/ unmarshal the decoded Micheline parameters into a Go struct
	type FA12Transfer struct {
		From  tezos.Address `json:"from"`
		To    tezos.Address `json:"to"`
		Value tezos.Z       `json:"value,string"`
	}
	type FA12TransferWrapper struct {
		Transfer FA12Transfer `json:"transfer"`
	}

	var transfer FA12TransferWrapper
	err = val.Unmarshal(&transfer)
	if err != nil {
		return err
	}
	buf, _ = json.MarshalIndent(transfer, "", "  ")
	fmt.Printf("FA transfer %s\n", string(buf))

	return nil
}
