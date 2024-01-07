// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Transfer examples
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
)

var (
	flags   = flag.NewFlagSet("transfer", flag.ContinueOnError)
	verbose bool
	node    string
	key     string
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
	flags.StringVar(&key, "key", "", "private key")
	flags.StringVar(&node, "node", "https://rpc.tzpro.io", "Tezos node URL")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Usage: transfer [flags] <cmd> [sub-args]")
			fmt.Println("\nFlags")
			flags.PrintDefaults()
			fmt.Println("\nTransaction Commands")
			fmt.Printf("  transfer   {<receiver> <amount>}+  transfer tez to single or multiple receiver(s)\n")
			os.Exit(0)
		}
		fmt.Println("Error:", err)
		return
	}

	if err := run(); err != nil {
		fmt.Println("Error:", err)
	}
}

func run() error {
	n := flags.NArg()
	if n < 1 {
		return fmt.Errorf("Command required")
	}

	if key == "" {
		return fmt.Errorf("Key required")
	}

	switch {
	case verbose:
		log.SetLevel(log.LevelTrace)
	default:
		log.SetLevel(log.LevelWarn)
	}
	rpc.UseLogger(log.Log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := rpc.NewClient(node, nil)
	if err != nil {
		return err
	}
	err = c.Init(ctx)
	if err != nil {
		return err
	}
	c.Listen()

	switch cmd := flags.Arg(0); cmd {
	case "transfer":
		if n < 3 || (n-1)%2 == 1 {
			return fmt.Errorf("Missing arguments")
		}
		return transfer(ctx, c)
	default:
		return fmt.Errorf("Unknown command %q", cmd)
	}
}

func transfer(ctx context.Context, c *rpc.Client) error {
	sk, err := tezos.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("Invalid private key %q: %v", key, err)
	}
	c.Signer = signer.NewFromKey(sk)

	// construct batch operation
	op := codec.NewOp().WithSource(sk.Address())
	for i := 1; i < flags.NArg(); i += 2 {
		fReceiver := flags.Arg(i)
		fAmount := flags.Arg(i + 1)
		recv, err := tezos.ParseAddress(fReceiver)
		if err != nil {
			return fmt.Errorf("Invalid receiver %q: %v", fReceiver, err)
		}
		amount, err := strconv.ParseInt(fAmount, 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid amount %q: %v", fAmount, err)
		}
		op.WithTransfer(recv, amount)
	}

	// send operation with default options
	rcpt, err := c.Send(ctx, op, nil)
	if err != nil {
		return err
	}

	total := rcpt.TotalCosts()
	fmt.Println("Costs")
	fmt.Printf("  Total             %d\n", total.Fee+total.StorageBurn+total.AllocationBurn)
	fmt.Printf("    Baker Fee       %d\n", total.Fee)
	fmt.Printf("    Storage burn    %d\n", total.StorageBurn)
	fmt.Printf("    Allocation burn %d\n", total.AllocationBurn)
	fmt.Printf("  Gas used          %d\n", total.GasUsed)
	fmt.Printf("  Storage bytes     %d\n", total.StorageUsed)

	if !rcpt.IsSuccess() {
		return fmt.Errorf("Transfer failed: %v", rcpt.Error())
	}
	fmt.Printf("Success.")
	return nil
}
