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
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
)

var (
	flags   = flag.NewFlagSet("rpc", flag.ContinueOnError)
	verbose bool
	debug   bool
	node    string
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "Be verbose")
	flags.BoolVar(&debug, "d", false, "Enable debug mode")
	flags.StringVar(&node, "node", "https://rpc.tzstats.com", "Tezos node url")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Usage: rpc [args] <cmd> [sub-args]")
			flags.PrintDefaults()
			fmt.Printf("\nOperations\n")
			fmt.Printf("  block <hash>|head        show block info\n")
			fmt.Printf("  op <hash>:<list>:<pos>   show operation info\n")
			fmt.Printf("  contract <hash>          show contract info\n")
			fmt.Printf("  search <ops> <lvl>       output blocks containing operations in list\n")
			fmt.Printf("  bootstrap                wait until node is bootstrapped\n")
			fmt.Printf("  monitor                  wait and show new heads as they are baked\n")
			os.Exit(0)
		}
		log.Fatal("Error:", err)
	}

	if err := run(); err != nil {
		log.Fatal("Error:", err)
	}
}

func run() error {
	if flags.NArg() < 1 {
		return fmt.Errorf("Command required")
	}

	switch {
	case debug:
		log.SetLevel(log.LevelDebug)
	case verbose:
		log.SetLevel(log.LevelInfo)
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

	switch flags.Arg(0) {
	case "block":
		h := flags.Arg(1)
		if h == "" {
			return fmt.Errorf("Missing block identifier")
		}
		if h == "head" {
			return fetchHead(ctx, c)
		} else {
			// try parsing arg as height (i.e. integer)
			if height, err := strconv.ParseInt(h, 10, 64); err == nil {
				return fetchBlockHeight(ctx, c, height)
			}
			// otherwise, parse as block hash
			h, err := tezos.ParseBlockHash(flags.Arg(1))
			if err != nil {
				return err
			}
			return fetchBlock(ctx, c, h)
		}
	case "op":
		h := flags.Arg(1)
		if h == "" {
			return fmt.Errorf("Missing operation identifier")
		}
		parts := strings.SplitN(h, ":", 3)
		if len(parts) != 3 {
			return fmt.Errorf("Invalid operation identifier (form: block-hash:list:pos)")
		}
		var bid rpc.BlockID
		bh, err := tezos.ParseBlockHash(parts[0])
		if err == nil {
			bid = bh
		} else {
			// parse as height
			height, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return fmt.Errorf("Invalid block identifier: %v", err)
			}
			bid = rpc.BlockLevel(height)
		}
		list, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("Invalid list identifier: %v", err)
		}
		pos, err := strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf("Invalid position identifier: %v", err)
		}
		return showOpInfo(ctx, c, bid, list, pos)

	case "bootstrap":
		return waitBootstrap(ctx, c)
	case "monitor":
		return monitorBlocks(ctx, c)
	case "contract":
		a := flags.Arg(1)
		if a == "" {
			return fmt.Errorf("Missing contract address")
		}
		addr, err := tezos.ParseAddress(a)
		if err != nil {
			return err
		}
		if addr.Type != tezos.AddressTypeContract {
			return fmt.Errorf("%s is not a smart contract address", a)
		}
		return showContractInfo(ctx, c, addr)
	case "search":
		height, _ := strconv.ParseInt(flags.Arg(2), 10, 64)
		return searchOps(ctx, c, flags.Arg(1), height)
	default:
		return fmt.Errorf("unknown command %s", flags.Arg(0))
	}
}

func fetchBlockHeight(ctx context.Context, c *rpc.Client, height int64) error {
	b, err := c.GetBlockHeight(ctx, height)
	if err != nil {
		return err
	}
	printBlock(b)
	return nil
}

func fetchBlock(ctx context.Context, c *rpc.Client, blockID tezos.BlockHash) error {
	b, err := c.GetBlock(ctx, blockID)
	if err != nil {
		return err
	}
	printBlock(b)
	return nil
}

func fetchHead(ctx context.Context, c *rpc.Client) error {
	head, err := c.GetTipHeader(ctx)
	if err != nil {
		return err
	}
	printHead(head)
	return nil
}

func printHead(h *rpc.BlockHeader) {
	fmt.Printf("Block  %d %s %s\n", h.Level, h.Hash, h.Timestamp)
}

func printBlock(b *rpc.Block) {
	fmt.Printf("Height %d (%d)\n", b.GetLevel(), b.GetCycle())
	fmt.Printf("Block  %s\n", b.Hash)
	fmt.Printf("Parent %s\n", b.Header.Predecessor)
	fmt.Printf("Time   %s\n", b.Header.Timestamp)

	// count operations and details
	ops := make(map[tezos.OpType]int)
	var count int
	for _, v := range b.Operations {
		for _, vv := range v {
			for _, op := range vv.Contents {
				kind := op.Kind()
				count++
				if c, ok := ops[kind]; ok {
					ops[kind] = c + 1
				} else {
					ops[kind] = 1
				}
			}
		}
	}
	fmt.Printf("Ops    %d: ", count)
	comma := ""
	for n, c := range ops {
		fmt.Printf("%s%d %s", comma, c, n)
		comma = ", "
	}
	fmt.Println()
	fmt.Println()
}

func waitBootstrap(ctx context.Context, c *rpc.Client) error {
	mon := rpc.NewBootstrapMonitor()
	defer mon.Close()
	if err := c.MonitorBootstrapped(ctx, mon); err != nil {
		return err
	}
	ctx2, cancel := context.WithCancel(ctx)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		select {
		case <-stop:
			fmt.Printf("Stopping monitor\n")
			cancel()
		case <-ctx.Done():
		}
	}()

	fmt.Printf("Waiting for chain to bootstrap... (cancel with Ctrl-C)\n\n")

	for {
		b, err := mon.Recv(ctx2)
		if err != nil {
			return err
		}
		if err := fetchBlock(ctx, c, b.Block); err != nil {
			return err
		}
	}
}

func monitorBlocks(ctx context.Context, c *rpc.Client) error {
	mon := rpc.NewBlockHeaderMonitor()
	defer mon.Close()
	if err := c.MonitorBlockHeader(ctx, mon); err != nil {
		return err
	}

	ctx2, cancel := context.WithCancel(ctx)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	go func() {
		select {
		case <-stop:
			fmt.Printf("Stopping monitor\n")
			cancel()
		case <-ctx.Done():
		}
	}()

	fmt.Printf("Waiting for new blocks... (cancel with Ctrl-C)\n\n")
	for {
		h, err := mon.Recv(ctx2)
		if err != nil {
			return err
		}
		if err := fetchBlock(ctx, c, h.Hash); err != nil {
			return err
		}
	}
}

func searchOps(ctx context.Context, c *rpc.Client, ops string, start int64) error {
	if ops == "deactivated" {
		return searchDeactivations(ctx, c, start)
	}

	// find the current blockchain tip
	tips, err := c.GetTips(ctx, 1, tezos.BlockHash{})
	if err != nil {
		return err
	}
	if len(tips) == 0 || len(tips[0]) == 0 {
		return fmt.Errorf("invalid chain tip")
	}
	tip, err := c.GetBlock(ctx, tips[0][0])
	if err != nil {
		return fmt.Errorf("Block %s failed: %w", tips[0][0], err)
	}

	// parse ops
	oplist := make([]tezos.OpType, 0)
	for _, op := range strings.Split(ops, ",") {
		ot := tezos.ParseOpType(op)
		if !ot.IsValid() {
			return fmt.Errorf("invalid operation type '%s'", op)
		}
		oplist = append(oplist, ot)
	}

	// fetching blocks forward
	height := start
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	for {
		b, err := c.GetBlockHeight(ctx, height)
		if err != nil {
			return fmt.Errorf("Block %d failed: %w", height, err)
		}

		if b.GetLevel()%1000 == 0 {
			fmt.Printf("Scanning blockchain at level %d\n", b.GetLevel())
		}

		// count operations and details
		opcount := make(map[tezos.OpType]int)
		var count int
		for _, v := range b.Operations {
			for _, vv := range v {
				for _, op := range vv.Contents {
					kind := op.Kind()
					count++
					if c, ok := opcount[kind]; ok {
						opcount[kind] = c + 1
					} else {
						opcount[kind] = 1
					}
					if kind == tezos.OpTypeTransaction {
						top := op.(*rpc.Transaction)
						for _, vvv := range top.Metadata.InternalResults {
							kind = vvv.Kind
							count++
							if c, ok := opcount[kind]; ok {
								opcount[kind] = c + 1
							} else {
								opcount[kind] = 1
							}
						}
					}
				}
			}
		}
		for _, op := range oplist {
			if n, ok := opcount[op]; ok {
				fmt.Printf("%s level=%d contains %d %s(s)\n", b.Hash, b.GetLevel(), n, op)
				// output relevant ops
				if !verbose {
					continue
				}
				for _, v := range b.Operations {
					for _, vv := range v {
						for _, o := range vv.Contents {
							if op == o.Kind() {
								enc.Encode(o)
							}
						}
					}
				}
			}
		}
		height++

		// the tip has probably advanced a lot since first fetch above,
		// but this is only for illustration
		if height > tip.GetLevel() {
			break
		}
	}
	return nil
}

func searchDeactivations(ctx context.Context, c *rpc.Client, start int64) error {
	// find the current blockchain tip
	tips, err := c.GetTips(ctx, 1, tezos.BlockHash{})
	if err != nil {
		return err
	}
	if len(tips) == 0 || len(tips[0]) == 0 {
		return fmt.Errorf("invalid chain tip")
	}
	tip, err := c.GetBlock(ctx, tips[0][0])
	if err != nil {
		return err
	}

	// fetching blocks forward
	height := start
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	for {
		b, err := c.GetBlockHeight(ctx, height)
		if err != nil {
			return err
		}

		if b.Metadata.Level.Level%1000 == 0 {
			fmt.Printf("Scanning blockchain at level %d\n", b.Metadata.Level.Level)
		}

		if len(b.Metadata.Deactivated) > 0 {
			res := map[int64][]tezos.Address{
				height: b.Metadata.Deactivated,
			}
			enc.Encode(res)
		}

		height++

		// the tip has probably advanced a lot since first fetch above,
		// but this is only for illustration
		if height > tip.Metadata.Level.Level {
			break
		}
	}
	return nil
}

func showContractInfo(ctx context.Context, c *rpc.Client, addr tezos.Address) error {
	fmt.Printf("Loading data for contract %s. This may take a while. Abort with Ctrl-C.\n", addr)
	script, err := c.GetContractScript(ctx, addr)
	if err != nil {
		return err
	}

	// unfold Micheline storage into human-readable form
	val := micheline.NewValue(script.StorageType(), script.Storage)
	m, err := val.Map()
	if err != nil {
		return err
	}
	buf, err := json.MarshalIndent(m, "  ", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Storage:\n  %v\n", string(buf))

	// identify bigmaps owned by the contract from contract type and storage
	bm := script.BigmapsByName()
	if len(bm) == 0 {
		fmt.Printf("Bigmaps  (none)\n")
	} else {
		fmt.Printf("Bigmaps:\n")
	}
	for name, bigid := range bm {
		// load bigmap type
		biginfo, err := c.GetBigmapInfo(ctx, bigid, rpc.Head)
		if err != nil {
			return err
		}
		// list all bigmap keys
		bigkeys, err := c.ListBigmapKeys(ctx, bigid, rpc.Head)
		if err != nil {
			return err
		}
		fmt.Printf("  %-15s  id=%d n_keys=%d bytes=%d\n", name, bigid, len(bigkeys), biginfo.TotalBytes)
		if verbose {
			for i, key := range bigkeys {
				// visit each key
				bigval, err := c.GetBigmapValue(ctx, bigid, key, rpc.Head)
				if err != nil {
					return err
				}
				// unfold Micheline type into human readable form
				val := micheline.NewValue(micheline.NewType(biginfo.ValueType), bigval)
				m, err := val.Map()
				if err != nil {
					return err
				}
				buf, err := json.MarshalIndent(m, "          ", "  ")
				if err != nil {
					return err
				}
				fmt.Printf("    %03d: %s\n", i, string(buf))
			}
		}
	}
	return nil
}

func showOpInfo(ctx context.Context, c *rpc.Client, bh rpc.BlockID, list, pos int) error {
	fmt.Printf("Loading op %s/%d/%d\n", bh, list, pos)
	op, err := c.GetBlockOperation(ctx, bh, list, pos)
	if err != nil {
		return err
	}
	fmt.Printf("Hash   %s\n", op.Hash)
	fmt.Printf("Parts  %d\n", len(op.Contents))
	for i, o := range op.Contents {
		fmt.Printf("Part   %d\n", i+1)
		fmt.Printf("  Type   %s\n", o.Kind())
		switch o.Kind() {
		case tezos.OpTypeTransaction:
			tx := o.(*rpc.Transaction)
			fmt.Printf("  Dest       %s\n", tx.Destination)
			fmt.Printf("  Fee        %d\n", tx.Fee)
			fmt.Printf("  Counter    %d\n", tx.Counter)
			fmt.Printf("  Amount     %d\n", tx.Amount)
			fmt.Printf("  Gas        %d/%d\n", tx.Metadata.Result.ConsumedGas, tx.GasLimit)
			fmt.Printf("  Storage    %d/%d\n", tx.Metadata.Result.PaidStorageSizeDiff, tx.StorageLimit)
			fmt.Printf("  Internal   %d (not shown)\n", len(tx.Metadata.InternalResults))
			var script *micheline.Script
			if tx.Destination.IsContract() {
				script, err = c.GetContractScript(ctx, tx.Destination)
				if err != nil {
					return err
				}
			}
			if tx.Parameters != nil {
				fmt.Printf("  Entrypoint %s\n", tx.Parameters.Entrypoint)
				ep, _, err := tx.Parameters.MapEntrypoint(script.ParamType())
				if err != nil {
					return err
				}
				val := micheline.NewValue(ep.Type(), tx.Parameters.Value)
				m, err := val.Map()
				if err != nil {
					return err
				}
				buf, err := json.MarshalIndent(m, "  ", "  ")
				if err != nil {
					return err
				}
				fmt.Println("  " + string(buf))
			}
			if prim := tx.Metadata.Result.Storage; prim != nil {
				val := micheline.NewValue(script.StorageType(), *prim)
				m, err := val.Map()
				if err != nil {
					return err
				}
				buf, err := json.MarshalIndent(m, "  ", "  ")
				if err != nil {
					return err
				}
				fmt.Printf("Storage:\n  %s\n", string(buf))
			}
		}
	}

	return nil
}
