// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Mempool examples
//
package main

import (
    "context"
    // "encoding/hex"
    // "encoding/json"
    "flag"
    "fmt"
    "io"
    "os"

    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/tezos"
    "github.com/echa/log"
)

var (
    flags   = flag.NewFlagSet("mempool", flag.ContinueOnError)
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
            fmt.Println("Usage: mempool [args] <cmd> [sub-args]")
            fmt.Println("\nCommands")
            fmt.Printf("  stream [<filter>]   stream new ops entering the mempool, optional filter\n")
            fmt.Printf("  wait <ophash>       wait for operation to be visible in mempool\n")
            fmt.Printf("  info                print info about mempool\n")
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

    switch cmd := flags.Arg(0); cmd {
    case "info":
        return info(ctx, c)
    case "stream":
        return stream(ctx, c, flags.Arg(1))
    case "wait":
        if n != 2 {
            return fmt.Errorf("Missing ophash")
        }
        return wait(ctx, c, flags.Arg(1))
    default:
        return fmt.Errorf("Unknown command %q", cmd)
    }
}

func info(ctx context.Context, c *rpc.Client) error {
    mem, err := c.GetMempool(ctx)
    if err != nil {
        return err
    }
    cnt := make([]map[tezos.OpType]int, 5)
    for i, v := range [][]*rpc.Operation{
        mem.Applied,
        mem.Refused,
        mem.BranchRefused,
        mem.BranchDelayed,
        mem.Unprocessed,
    } {
        if cnt[i] == nil {
            cnt[i] = make(map[tezos.OpType]int)
        }
        m := cnt[i]
        for _, op := range v {
            for _, o := range op.Contents {
                k := o.Kind()
                if c, ok := m[k]; ok {
                    m[k] = c + 1
                } else {
                    m[k] = 1
                }
            }
        }
    }
    fmt.Println("Applied:", len(mem.Applied))
    for n, c := range cnt[0] {
        fmt.Printf("  %s (%d)\n", n, c)
    }
    fmt.Println("Refused:", len(mem.Refused))
    for n, c := range cnt[1] {
        fmt.Printf("  %s (%d)\n", n, c)
    }
    fmt.Println("Branch refused:", len(mem.BranchRefused))
    for n, c := range cnt[2] {
        fmt.Printf("  %s (%d)\n", n, c)
    }
    fmt.Println("Branch delayed:", len(mem.BranchDelayed))
    for n, c := range cnt[3] {
        fmt.Printf("  %s (%d)\n", n, c)
    }
    fmt.Println("Unprocessed:", len(mem.Unprocessed))
    for n, c := range cnt[4] {
        fmt.Printf("  %s (%d)\n", n, c)
    }
    return nil
}

func stream(ctx context.Context, c *rpc.Client, flt string) error {
    var mon *rpc.MempoolMonitor
    for {
        if mon == nil {
            fmt.Println("Connecting monitor")
            mon = rpc.NewMempoolMonitor()
            if err := c.MonitorMempool(ctx, mon); err != nil {
                mon.Close()
                return err
            }
        }
        ops, err := mon.Recv(ctx)
        if err != nil {
            if err != io.EOF {
                fmt.Println("Monitor failed")
                return err
            }
            head, err := c.GetTipHeader(ctx)
            if err == nil {
                fmt.Println("Monitor closed on new block", head.Level, head.Hash)
            } else {
                fmt.Println("Monitor closed:", err)
            }
            mon = nil
            continue
        }
        for _, op := range ops {
            fmt.Println(op.Hash, op.Contents[0].Kind())
            // calculate cost but discard result to catch errors
            _ = op.Cost()
        }
    }
}

func wait(ctx context.Context, c *rpc.Client, hash string) error {
    fmt.Println("Not implemented yet")
    return nil
}
