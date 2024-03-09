// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
)

var (
	flags   = flag.NewFlagSet("params", flag.ContinueOnError)
	verbose bool
	node    string
	proto   string
	net     string
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
	flags.StringVar(&node, "node", "https://rpc.tzpro.io", "node url")
	flags.StringVar(&proto, "proto", "", "simulate with protocol")
	flags.StringVar(&net, "net", "", "simulate with network")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("params [height/hash/alias]")
			flags.PrintDefaults()
			os.Exit(0)
		}
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run() error {
	id := rpc.Head
	if flags.NArg() > 0 {
		id = rpc.BlockAlias(flags.Arg(0))
	}
	ctx := context.Background()

	// fetch block & constants at height
	c, _ := rpc.NewClient(node, nil)
	block, err := c.GetBlock(ctx, id)
	if err != nil {
		if rpc.ErrorStatus(err) != 404 {
			return err
		}
		block, err = c.GetHeadBlock(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Block %s does not exist yet. Using constants at current head block %d\n", id, block.GetLevel())
	}
	p, err := c.GetParams(ctx, rpc.BlockLevel(block.GetLevel()))
	if err != nil {
		return err
	}

	// simulate
	if proto != "" {
		p.Protocol, err = tezos.ParseProtocolHash(proto)
		if err != nil {
			return err
		}
	}
	if net != "" {
		p.ChainId, err = tezos.ParseChainIdHash(net)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Using protocol %s on %s\n", tezos.Short(p.Protocol.String())[:8], p.Network)

	fmt.Println("Height .............................. ", block.GetLevel())
	fmt.Println("Cycle ............................... ", block.GetCycle())
	fmt.Println("--------------------------------------------- ")
	fmt.Println("Cycle (calc)......................... ", p.CycleFromHeight(block.GetLevel()))
	fmt.Println("Snap (calc).......................... ", p.SnapshotBlock(524, 1))
	fmt.Println("Protocol ............................ ", p.Protocol)
	fmt.Println("Network ............................. ", p.Network)
	fmt.Println("Version ............................. ", p.Version)
	fmt.Println("Chain_id ............................ ", p.ChainId)
	fmt.Println("Operation Tags ...................... ", p.OperationTagsVersion)
	fmt.Println("StartHeight.......................... ", p.StartHeight)
	fmt.Println("EndHeight............................ ", p.EndHeight)
	fmt.Println("StartOffset.......................... ", p.StartOffset)
	fmt.Println("StartCycle........................... ", p.StartCycle)
	fmt.Println("--------------------------------------------- ")
	fmt.Println("minimal_block_delay ................. ", p.MinimalBlockDelay)
	fmt.Println("max_operations_ttl .................. ", p.MaxOperationsTTL)
	fmt.Println("cost_per_byte ....................... ", p.CostPerByte)
	fmt.Println("origination_size .................... ", p.OriginationSize)
	fmt.Println("preserved_cycles .................... ", p.PreservedCycles)
	fmt.Println("--------------------------------------------- ")
	fmt.Println("hard_gas_limit_per_operation ........ ", p.HardGasLimitPerOperation)
	fmt.Println("hard_gas_limit_per_block ............ ", p.HardGasLimitPerBlock)
	fmt.Println("hard_storage_limit_per_operation .... ", p.HardStorageLimitPerOperation)
	fmt.Println("max_operation_data_length ........... ", p.MaxOperationDataLength)
	return nil
}
