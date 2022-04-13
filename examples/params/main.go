// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/legonian/tzgo/rpc"
	"github.com/legonian/tzgo/tezos"
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
	flags.StringVar(&node, "node", "https://rpc.tzstats.com", "node url")
	flags.StringVar(&proto, "proto", "", "simulate with protocol")
	flags.StringVar(&net, "net", "", "simulate with network")
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

	height, err := strconv.ParseInt(flags.Arg(0), 10, 64)
	if err != nil {
		return err
	}
	ctx := context.Background()

	// fetch block & constants at height
	c, _ := rpc.NewClient(node, nil)
	block, err := c.GetBlockHeight(ctx, height)
	if err != nil {
		if rpc.ErrorStatus(err) != 404 {
			return err
		}
		block, err = c.GetHeadBlock(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Block %d does not exist yet. Using constants at current head block %d\n", height, block.GetLevel())
	}
	cons, err := c.GetConstants(ctx, rpc.BlockLevel(block.GetLevel()))
	if err != nil {
		return err
	}
	forNet, forProto := block.ChainId, block.Protocol

	// simulate
	if proto != "" {
		forProto, err = tezos.ParseProtocolHash(proto)
		if err != nil {
			return err
		}
	}
	if net != "" {
		forNet, err = tezos.ParseChainIdHash(net)
		if err != nil {
			return err
		}
	}
	p := cons.MapToChainParams().ForNetwork(forNet).ForProtocol(forProto)
	fmt.Printf("Using protocol %s on %s\n", forProto.Short()[:8], p.Network)

	fmt.Println("Height ...................... ", height)
	fmt.Println("Protocol .................... ", forProto)
	fmt.Println("Period ...................... ", block.GetVotingPeriodKind(), block.GetVotingPeriod())
	fmt.Println("StartCycle .................. ", p.StartCycle)
	fmt.Println("StartBlockOffset ............ ", p.StartBlockOffset)
	fmt.Println("VoteBlockOffset ............. ", p.VoteBlockOffset)
	fmt.Println("BlocksPerCycle .............. ", p.BlocksPerCycle)
	fmt.Println("BlocksPerVotingPeriod ....... ", p.BlocksPerVotingPeriod)
	fmt.Println("----------------------------- ")
	fmt.Println("IsCycleStart ................ ", p.IsCycleStart(height))
	fmt.Println("IsCycleEnd .................. ", p.IsCycleEnd(height))
	fmt.Println("IsSnapshotBlock ............. ", p.IsSnapshotBlock(height))
	fmt.Println("IsSeedRequired .............. ", p.IsSeedRequired(height))
	fmt.Println("CycleFromHeight ............. ", p.CycleFromHeight(height))
	fmt.Println("CycleStartHeight ............ ", p.CycleStartHeight(p.CycleFromHeight(height)))
	fmt.Println("CycleEndHeight .............. ", p.CycleEndHeight(p.CycleFromHeight(height)))
	fmt.Println("SnapshotIndex ............... ", p.SnapshotIndex(height))
	fmt.Println("MaxSnapshotIndex ............ ", p.MaxSnapshotIndex())
	fmt.Println("VotingStartCycleFromHeight .. ", p.VotingStartCycleFromHeight(height))
	fmt.Println("IsVoteStart ................. ", p.IsVoteStart(height))
	fmt.Println("IsVoteEnd ................... ", p.IsVoteEnd(height))
	fmt.Println("VoteStartHeight ............. ", p.VoteStartHeight(height))
	fmt.Println("VoteEndHeight ............... ", p.VoteEndHeight(height))
	fmt.Println("IsPreBabylonHeight .......... ", p.IsPreBabylonHeight(height))

	return nil
}
