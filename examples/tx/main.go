// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Transaction examples
//
// # Requirements
//
// - private key for a funded testnet or mainnet account (https://teztnets.xyz/)
//
package main

import (
    "context"
    "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "strconv"

    "blockwatch.cc/tzgo/codec"
    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/tezos"
    "blockwatch.cc/tzgo/wallet"
    "github.com/echa/log"
)

var (
    flags   = flag.NewFlagSet("tx", flag.ContinueOnError)
    verbose bool
    node    string
)

func init() {
    flags.Usage = func() {}
    flags.BoolVar(&verbose, "v", false, "be verbose")
    flags.StringVar(&node, "node", "https://rpc.hangzhou.tzstats.com", "Tezos node URL")
}

func main() {
    if err := flags.Parse(os.Args[1:]); err != nil {
        if err == flag.ErrHelp {
            fmt.Println("Usage: tx [args] <cmd> [sub-args]")
            fmt.Println("\nArguments")
            flags.PrintDefaults()
            fmt.Println("\nCommands")
            fmt.Printf("  encode <type> <data>       generate operation `type` from JSON `data`\n")
            fmt.Printf("  validate <type> <data>     compare local encoding against remote encoding\n")
            fmt.Printf("  decode <msg>               decode binary operation\n")
            fmt.Printf("  digest <msg>               generate operation digest for signing\n")
            fmt.Printf("  sign <key> <msg>           sign message digest\n")
            fmt.Printf("  simulate <msg>             simulate executing operation using invalid signature\n")
            fmt.Printf("  broadcast <msg> <sig>      broadcast signed operation\n")
            fmt.Printf("  wait <ophash> [<n>]        waits for operation to be included after n confirmations (optional)\n")
            fmt.Println("\nOperation types & required JSON keys")
            fmt.Printf("  endorsement                 level:int\n")
            fmt.Printf("  endorsement_with_slot       level:int slot:int\n")
            fmt.Printf("  double_baking_evidence      <complex>\n")
            fmt.Printf("  double_endorsement_evidence <complex>\n")
            fmt.Printf("  seed_nonce_revelation       level:str(int) nonce:hash\n")
            fmt.Printf("  activate_account            pkh:addr secret:hex32\n")
            fmt.Printf("  reveal                      source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) public_key:key \n")
            fmt.Printf("  transaction                 source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) amount:str(int) destination:addr \n")
            fmt.Printf("  origination                 source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) balance:str(int) delegate?:addr script:prim\n")
            fmt.Printf("  delegation                  source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) delegate?:addr\n")
            fmt.Printf("  proposals                   source:addr period:str(int) proposal:[hash]\n")
            fmt.Printf("  ballot                      source:addr period:str(int) proposal:hash ballot:(yay,nay,pass)\n")
            fmt.Printf("  register_global_constant    source:addr fee:str(int) counter:str(int) gas_limit:str(int) storage_limit:str(int) value:prim\n")
            fmt.Printf("  failing_noop                arbitrary:str\n")
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
    case "Z":
        if n < 2 {
            return fmt.Errorf("Missing number")
        }
        return encodeZ(flags.Arg(1))
    case "N":
        if n < 2 {
            return fmt.Errorf("Missing number")
        }
        return encodeN(flags.Arg(1))
    case "encode":
        if n < 3 {
            return fmt.Errorf("Missing type or data")
        }
        return encode(ctx, c, flags.Arg(1), flags.Arg(2))
    case "decode":
        if n < 2 {
            return fmt.Errorf("Missing message")
        }
        return decode(flags.Arg(1))
    case "validate":
        if n < 3 {
            return fmt.Errorf("Missing type or data")
        }
        return validate(ctx, c, flags.Arg(1), flags.Arg(2))
    case "digest":
        if n < 2 {
            return fmt.Errorf("Missing message")
        }
        return digest(ctx, c, flags.Arg(1))
    case "sign":
        if n < 3 {
            return fmt.Errorf("Missing key or message")
        }
        return sign(ctx, c, flags.Arg(1), flags.Arg(2))
    case "simulate":
        if n < 2 {
            return fmt.Errorf("Missing message")
        }
        return simulate(ctx, c, flags.Arg(1))
    case "broadcast":
        if n < 3 {
            return fmt.Errorf("Missing message or signature")
        }
        return broadcast(ctx, c, flags.Arg(1), flags.Arg(2))
    case "wait":
        return wait(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
    default:
        return fmt.Errorf("Unknown command %q", cmd)
    }
}

func makeOp(t, data string) (codec.Operation, error) {
    var o codec.Operation
    switch tezos.ParseOpType(t) {
    case tezos.OpTypeActivateAccount:
        o = new(codec.ActivateAccount)
    case tezos.OpTypeDoubleBakingEvidence:
        o = new(codec.DoubleBakingEvidence)
    case tezos.OpTypeDoubleEndorsementEvidence:
        o = new(codec.DoubleEndorsementEvidence)
    case tezos.OpTypeSeedNonceRevelation:
        o = new(codec.SeedNonceRevelation)
    case tezos.OpTypeTransaction:
        o = new(codec.Transaction)
    case tezos.OpTypeOrigination:
        o = new(codec.Origination)
    case tezos.OpTypeDelegation:
        o = new(codec.Delegation)
    case tezos.OpTypeReveal:
        o = new(codec.Reveal)
    case tezos.OpTypeEndorsement:
        o = new(codec.Endorsement)
    case tezos.OpTypeEndorsementWithSlot:
        o = new(codec.EndorsementWithSlot)
    case tezos.OpTypeProposals:
        o = new(codec.Proposals)
    case tezos.OpTypeBallot:
        o = new(codec.Ballot)
    case tezos.OpTypeFailingNoop:
        o = new(codec.FailingNoop)
    case tezos.OpTypeRegisterConstant:
        o = new(codec.RegisterGlobalConstant)
    default:
        return nil, fmt.Errorf("Unsupported op type %q", t)
    }
    if err := json.Unmarshal([]byte(data), o); err != nil {
        return nil, err
    }
    return o, nil
}

func encodeZ(s string) error {
    var z, x tezos.Z
    if err := z.UnmarshalText([]byte(s)); err != nil {
        return err
    }
    fmt.Println("Input:", s)
    fmt.Println("Parsed:", z.String())
    buf, _ := z.MarshalBinary()
    fmt.Println("Binary:", hex.EncodeToString(buf))
    x.UnmarshalBinary(buf)
    fmt.Println("Decoded:", x.String())
    buf, _ = x.MarshalBinary()
    fmt.Println("Recoded:", hex.EncodeToString(buf))
    return nil
}

func encodeN(s string) error {
    var z, x tezos.N
    if err := z.UnmarshalText([]byte(s)); err != nil {
        return err
    }
    fmt.Println("Input:", s)
    fmt.Println("Parsed:", z.String())
    buf, _ := z.MarshalBinary()
    fmt.Println("Binary:", hex.EncodeToString(buf))
    x.UnmarshalBinary(buf)
    fmt.Println("Decoded:", x.String())
    buf, _ = x.MarshalBinary()
    fmt.Println("Recoded:", hex.EncodeToString(buf))
    return nil
}

func encode(ctx context.Context, c *rpc.Client, typ, data string) error {
    o, err := makeOp(typ, data)
    if err != nil {
        return err
    }
    hash, err := c.GetBlockHash(ctx, rpc.Head)
    if err != nil {
        return err
    }
    op := codec.NewOp().
        WithContents(o).
        WithBranch(hash)
    fmt.Println("Encoded:", hex.EncodeToString(op.Bytes()))
    return nil
}

func decode(msg string) error {
    buf, err := hex.DecodeString(msg)
    if err != nil {
        return err
    }
    op, err := codec.DecodeOp(buf)
    if err != nil {
        return err
    }
    buf, err = json.MarshalIndent(op, "", "  ")
    if err != nil {
        return err
    }
    fmt.Println(string(buf))
    return nil
}

func validate(ctx context.Context, c *rpc.Client, typ, data string) error {
    o, err := makeOp(typ, data)
    if err != nil {
        return err
    }
    hash, err := c.GetBlockHash(ctx, rpc.Head)
    if err != nil {
        return err
    }
    op := codec.NewOp().
        WithContents(o).
        WithBranch(hash)
    fmt.Println("Local:", hex.EncodeToString(op.Bytes()))
    if err := wallet.Validate(ctx, c, op); err != nil {
        return err
    }
    fmt.Println("OK.")
    return nil
}

func digest(ctx context.Context, c *rpc.Client, msg string) error {
    buf, err := hex.DecodeString(msg)
    if err != nil {
        return err
    }
    op, err := codec.DecodeOp(buf)
    if err != nil {
        return err
    }
    fmt.Println("Digest:", hex.EncodeToString(op.Digest()))
    return nil
}

func sign(ctx context.Context, c *rpc.Client, key, msg string) error {
    buf, err := hex.DecodeString(msg)
    if err != nil {
        return err
    }
    op, err := codec.DecodeOp(buf)
    if err != nil {
        return err
    }
    sk, err := tezos.ParsePrivateKey(key)
    if err != nil {
        return err
    }
    if err := op.Sign(sk); err != nil {
        return err
    }
    fmt.Println("Signature:", op.Signature.String())
    fmt.Println("Binary:", hex.EncodeToString(op.Signature.Bytes()))
    return nil
}

func simulate(ctx context.Context, c *rpc.Client, msg string) error {
    buf, err := hex.DecodeString(msg)
    if err != nil {
        return err
    }
    op, err := codec.DecodeOp(buf)
    if err != nil {
        return err
    }
    err = c.InitChain(ctx)
    if err != nil {
        return err
    }
    // special treatment for endorsements (wrapper&content branches must match)
    if es, ok := op.Contents[0].(*codec.EndorsementWithSlot); ok {
        head, _ := c.GetBlockHash(ctx, rpc.Head)
        fmt.Println("Setting block hash", head)
        es.Endorsement.Branch = head
        op.WithBranch(head)
    }
    res, err := wallet.Simulate(ctx, c, op)
    if err != nil {
        return err
    }
    buf, err = json.MarshalIndent(res.Op, "", "  ")
    if err != nil {
        return err
    }
    fmt.Println("Result\n", string(buf))
    buf, err = json.MarshalIndent(res.Cost(), "", "  ")
    if err != nil {
        return err
    }
    fmt.Println("Cost\n", string(buf))
    return nil
}

func broadcast(ctx context.Context, c *rpc.Client, msg, sig string) error {
    buf, err := hex.DecodeString(msg)
    if err != nil {
        return err
    }
    op, err := codec.DecodeOp(buf)
    if err != nil {
        return err
    }
    s, err := tezos.ParseSignature(sig)
    if err != nil {
        return err
    }
    op.WithSignature(s)
    hash, err := wallet.Broadcast(ctx, c, op)
    if err != nil {
        return err
    }
    fmt.Println("Broadcast:", hash.String())
    return nil
}

func wait(ctx context.Context, c *rpc.Client, op, conf, ttl string) error {
    oh, err := tezos.ParseOpHash(op)
    if err != nil {
        return err
    }
    fut := wallet.NewFutureResult(oh)
    if n, err := strconv.ParseInt(conf, 10, 64); err == nil {
        fut.WithConfirmations(n)
    }
    if n, err := strconv.ParseInt(ttl, 10, 64); err == nil {
        fut.WithTTL(n)
    }
    mon := wallet.NewMonitor()
    mon.Listen(c)
    fut.Listen(mon)
    fut.Wait()
    res, err := fut.GetResult(ctx, c)
    if err != nil {
        return err
    }
    fmt.Printf("Op included in %s with %d confirmations\n", res.Block, fut.Confirmations())
    buf, err := json.MarshalIndent(res.Op, "", "  ")
    if err != nil {
        return err
    }
    fmt.Println("Result\n", string(buf))
    buf, err = json.MarshalIndent(res.Cost(), "", "  ")
    if err != nil {
        return err
    }
    fmt.Println("Cost\n", string(buf))
    return nil
}
