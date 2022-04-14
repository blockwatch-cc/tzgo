// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Rights examples
//
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "strconv"

    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/tezos"
    "github.com/echa/log"
)

var (
    flags   = flag.NewFlagSet("rights", flag.ContinueOnError)
    verbose bool
    node    string
    ttl     int64
)

func init() {
    flags.Usage = func() {}
    flags.BoolVar(&verbose, "v", false, "be verbose")
    flags.StringVar(&node, "node", "https://rpc.tzstats.com", "Tezos node URL")
    flags.Int64Var(&ttl, "ttl", 120, "Operation TTL")
}

func main() {
    if err := flags.Parse(os.Args[1:]); err != nil {
        if err == flag.ErrHelp {
            fmt.Println("Usage: rights [args] <cmd> [sub-args]")
            fmt.Println("\nCommands")
            fmt.Printf("  snap [<cycle>]              get snapshot info for all or selected cycle\n")
            fmt.Printf("  bake <cycle> [<address>]    get cycle baking rights for baker\n")
            fmt.Printf("  endorse <cycle> [<address>] get cycle endorsing rights for baker\n")
            fmt.Println("\nArguments")
            flags.PrintDefaults()
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

    switch {
    case verbose:
        log.SetLevel(log.LevelDebug)
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
    if err := c.Init(ctx); err != nil {
        return err
    }

    switch cmd := flags.Arg(0); cmd {
    case "snap":
        return snap(ctx, c, flags.Arg(1))
    case "bake":
        return bake(ctx, c, flags.Arg(1), flags.Arg(2))
    case "endorse":
        return endorse(ctx, c, flags.Arg(1), flags.Arg(2))
    default:
        return fmt.Errorf("Unknown command %q", cmd)
    }
}

func snap(ctx context.Context, c *rpc.Client, id string) error {
    var (
        cycle int64
        err   error
    )
    if id != "" {
        cycle, err = strconv.ParseInt(id, 10, 64)
        if err != nil {
            return err
        }
    }
    for i := cycle; cycle == 0 || i == cycle; i++ {
        height := rpc.BlockLevel(c.Params.CycleStartHeight(i))
        si, err := c.GetSnapshotIndexCycle(ctx, height, i)
        if err != nil {
            if rpc.ErrorStatus(err) == 404 {
                break
            }
            return err
        }
        fmt.Printf("c%03d snapshot at c%03d/%d\n", si.Cycle, si.Base, si.Index)
    }
    return nil
}

func bake(ctx context.Context, c *rpc.Client, id, addr string) error {
    cycle, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        return err
    }
    a, err := tezos.ParseAddress(addr)
    if err != nil && addr != "" {
        return err
    }
    height := rpc.BlockLevel(c.Params.CycleStartHeight(cycle))
    rights, err := c.ListBakingRightsCycle(ctx, height, cycle, 4)
    if err != nil {
        return err
    }
    for _, v := range rights {
        if a.IsValid() && !v.Address().Equal(a) {
            continue
        }
        fmt.Printf("%07d %s bake %d\n", v.Level, v.Address(), v.Priority+v.Round)
    }
    return nil
}

func endorse(ctx context.Context, c *rpc.Client, id, addr string) error {
    cycle, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        return err
    }
    a, err := tezos.ParseAddress(addr)
    if err != nil && addr != "" {
        return err
    }
    height := rpc.BlockLevel(c.Params.CycleStartHeight(cycle))
    rights, err := c.ListEndorsingRightsCycle(ctx, height, cycle)
    if err != nil {
        return err
    }
    for _, v := range rights {
        if a.IsValid() && !v.Address().Equal(a) {
            continue
        }
        fmt.Printf("%07d %s endorse %d\n", v.Level, v.Address(), v.Power())
    }
    return nil
}
