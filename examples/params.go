// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"blockwatch.cc/tzgo/rpc"
)

var (
	flags   = flag.NewFlagSet("params", flag.ContinueOnError)
	verbose bool
	node    string
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
	flags.StringVar(&node, "node", "https://rpc.tzstats.com", "node url")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Params Test")
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
	if flags.NArg() < 1 {
		return fmt.Errorf("Block height required")
	}

	h, err := strconv.ParseInt(flags.Arg(0), 10, 64)
	if err != nil {
		return err
	}

	// fetch constants at height
	c, _ := rpc.NewClient(node, nil)
	cons, err := c.GetConstantsHeight(context.Background(), h)
	if err != nil {
		return err
	}

	block, err := c.GetBlockHeight(context.Background(), h)
	if err != nil {
		return err
	}

	p := cons.MapToChainParams().ForNetwork(block.ChainId).ForProtocol(block.Protocol)

	fmt.Println("Height ...................... ", h)
	fmt.Println("Protocol .................... ", block.Protocol)
	fmt.Println("Period ...................... ", block.GetVotingPeriodKind(), block.GetVotingPeriod())
	fmt.Println("StartCycle .................. ", p.StartCycle)
	fmt.Println("StartBlockOffset ............ ", p.StartBlockOffset)
	// fmt.Println("VoteBlockOffset ............. ", p.VoteBlockOffset)
	fmt.Println("BlocksPerCycle .............. ", p.BlocksPerCycle)
	fmt.Println("BlocksPerVotingPeriod ....... ", p.BlocksPerVotingPeriod)
	fmt.Println("----------------------------- ")
	fmt.Println("IsCycleStart ................ ", p.IsCycleStart(h))
	fmt.Println("IsCycleEnd .................. ", p.IsCycleEnd(h))
	fmt.Println("IsSnapshotBlock ............. ", p.IsSnapshotBlock(h))
	fmt.Println("IsSeedRequired .............. ", p.IsSeedRequired(h))
	fmt.Println("CycleFromHeight ............. ", p.CycleFromHeight(h))
	fmt.Println("CycleStartHeight ............ ", p.CycleStartHeight(p.CycleFromHeight(h)))
	fmt.Println("CycleEndHeight .............. ", p.CycleEndHeight(p.CycleFromHeight(h)))
	fmt.Println("SnapshotIndex ............... ", p.SnapshotIndex(h))
	fmt.Println("MaxSnapshotIndex ............ ", p.MaxSnapshotIndex())
	fmt.Println("VotingStartCycleFromHeight .. ", p.VotingStartCycleFromHeight(h))
	fmt.Println("IsVoteStart ................. ", p.IsVoteStart(h))
	fmt.Println("IsVoteEnd ................... ", p.IsVoteEnd(h))
	fmt.Println("VoteStartHeight ............. ", p.VoteStartHeight(h))
	fmt.Println("VoteEndHeight ............... ", p.VoteEndHeight(h))
	fmt.Println("IsPreBabylonHeight .......... ", p.IsPreBabylonHeight(h))

	return nil
}
