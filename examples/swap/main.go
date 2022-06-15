// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Dex examples
//
// - Quipuswap v1      KT1WxgZ1ZSfMgmsSDDcUn8Xn577HwnQ7e1Lb
// - Quipuswap Token   KT1VNEzpf631BLsdPJjt2ZhgUitR392x6cSi
// - Dexter v1         KT1Puc9St8wdNoGtLiD2WXaHbWU7styaxYhD
// - Dexter v2         KT1AbYeDbjjcAnV1QK7EZUUdqku77CdkTuv6
// - Dexter LB         KT1TxqZ8QtKvLu3V3JH7Gx58n7Co8pgtpQU5
// - Youves            KT1Xbx9pykNd38zag4yZvnmdSNBknmCETvQV
//
// Quipu tez/USDtez
// go run ./examples/swap/ -slippage 5/1000 sim KT1WxgZ1ZSfMgmsSDDcUn8Xn577HwnQ7e1Lb buy 1
//
// Youves info
// go run ./examples/swap/ info KT1Xbx9pykNd38zag4yZvnmdSNBknmCETvQV
//
// Dexter LB
// go run ./examples/swap/ info KT1TxqZ8QtKvLu3V3JH7Gx58n7Co8pgtpQU5
//
// Quipu info
// go run ./examples/swap/ info KT1WxgZ1ZSfMgmsSDDcUn8Xn577HwnQ7e1Lb
//

// Examples
// - query pool info (e.g. ctez on Quipu v1)
//
//   go run ./examples/swap/ info KT1FbYwEWU8BTfrvNoL5xDEC5owsDxv9nqKT 0
//
// - estimate swap cost (ctez on Quipu v1)
//
//   go run ./examples/swap/ sim KT1FbYwEWU8BTfrvNoL5xDEC5owsDxv9nqKT 0 1000
//
// - execute swaps
//

// TBD
// - [ ] reconceil post execution data (exact coins received, exact fees paid)
// - [ ] execute swaps
// - [ ] plenty, vortex dex

package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "strconv"
    "strings"
    "syscall"

    "blockwatch.cc/tzgo/contract"
    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/signer"
    "blockwatch.cc/tzgo/signer/remote"
    "blockwatch.cc/tzgo/tezos"
    "github.com/echa/log"
    "golang.org/x/term"
)

var (
    flags      = flag.NewFlagSet("swap", flag.ContinueOnError)
    verbose    bool
    node       string
    key        string
    pass       string
    signerAddr string
    signerUrl  string
    slippage   Percent = NewPercent(5, 1000)
)

func init() {
    flags.Usage = func() {}
    flags.BoolVar(&verbose, "v", false, "be verbose")
    flags.StringVar(&node, "node", "https://rpc.tzstats.com", "Tezos node URL")
    flags.StringVar(&key, "key", "", "Secret key (env TEZOS_KEY")
    flags.StringVar(&pass, "pass", "", "password for encrypted keys (env TEZOS_KEY_PASSPHRASE)")
    flags.StringVar(&signerAddr, "addr", "", "Remote signer Tezos address")
    flags.StringVar(&signerUrl, "signer", "http://localhost:6732", "Remote signer URL")
    flags.Var(&slippage, "slippage", "Trade execuation slippage (0.5% = 5/1000)")
}

func main() {
    if err := flags.Parse(os.Args[1:]); err != nil {
        if err == flag.ErrHelp {
            fmt.Println("Usage: swap [flags] <cmd> [sub-args]")
            fmt.Println("\nFlags")
            flags.PrintDefaults()
            fmt.Println("\nCommands")
            fmt.Printf("  pools <KT1>                          list pools\n")
            fmt.Printf("  info  <KT1>[/<pool_id>]              show pool info\n")
            fmt.Printf("  sim   <KT1>[/<pool_id>] <dir> <in>   dry run a swap on `in` tokens (dir: buy|sell)\n")
            fmt.Printf("  swap  <KT1>[/<pool_id>] <dir> <in>   execute swap of `in` tokens (requires key or remote signer)\n")
            os.Exit(0)
        }
        fmt.Println("Error:", err)
        return
    }

    if err := run(); err != nil {
        fmt.Println("Error:", err)
    }
}

func parsePool(s string) (addr tezos.Address, pool int64, err error) {
    parts := strings.Split(s, "/")
    switch len(parts) {
    case 1:
        addr, err = tezos.ParseAddress(parts[0])
    case 2:
        addr, err = tezos.ParseAddress(parts[0])
        if err == nil {
            pool, err = strconv.ParseInt(parts[1], 10, 64)
        }
    default:
        err = fmt.Errorf("invalid dex pool address %q", s)
    }
    return
}

func run() error {
    n := flags.NArg()
    if n < 1 {
        return fmt.Errorf("Command required")
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

    // load key from env
    if key == "" {
        key = os.Getenv("TEZOS_KEY")
    }

    // load signer from key or remote URL
    if key != "" {
        var sk tezos.PrivateKey
        if tezos.IsEncryptedKey(key) {
            sk, err = tezos.ParseEncryptedPrivateKey(key, readPassword())
        } else {
            sk, err = tezos.ParsePrivateKey(key)
        }
        if err != nil {
            return fmt.Errorf("Invalid private key %q: %v", key, err)
        }
        if !sk.IsValid() {
            return fmt.Errorf("Invalid private key %q", key)
        }
        c.Signer = signer.NewFromKey(sk)
        fmt.Printf("Using local signing address %s\n", sk.Public().Address())
    } else if signerUrl != "" {
        a, err := tezos.ParseAddress(signerAddr)
        if err != nil {
            return fmt.Errorf("Invalid signer address %q: %v", signerAddr, err)
        }
        rs, err := remote.New(signerUrl, nil)
        if err != nil {
            return err
        }
        c.Signer = rs.WithAddress(a)
        fmt.Printf("Using remote signing address %s\n", a)
    }

    switch cmd := flags.Arg(0); cmd {
    case "pools":
        if n < 1 {
            return fmt.Errorf("Missing arguments")
        }
        addr, _, err := parsePool(flags.Arg(1))
        if err != nil {
            return err
        }
        return pools(ctx, c, addr)
    case "info":
        if n < 2 {
            return fmt.Errorf("Missing arguments: info <KT1>[/<pool_id>]")
        }
        addr, pool, err := parsePool(flags.Arg(1))
        if err != nil {
            return err
        }
        return info(ctx, c, addr, pool)
    case "sim":
        if n < 4 {
            return fmt.Errorf("Missing arguments: sim <KT1>[/<pool_id>] <dir> <in>")
        }
        addr, pool, err := parsePool(flags.Arg(1))
        if err != nil {
            return err
        }
        dir, err := ParseDirection(flags.Arg(2))
        if err != nil {
            return err
        }
        amount, err := strconv.ParseFloat(flags.Arg(3), 64)
        if err != nil {
            return err
        }
        return simulate(ctx, c, addr, pool, dir, amount)
    case "swap":
        if n < 4 {
            return fmt.Errorf("Missing arguments: swap <KT1>[/<pool_id>] <dir> <in>")
        }
        addr, pool, err := parsePool(flags.Arg(1))
        if err != nil {
            return err
        }
        dir, err := ParseDirection(flags.Arg(2))
        if err != nil {
            return err
        }
        amount, err := strconv.ParseFloat(flags.Arg(3), 64)
        if err != nil {
            return err
        }
        c.Listen()
        return swap(ctx, c, addr, pool, dir, amount)
    default:
        return fmt.Errorf("Unknown command %q", cmd)
    }
}

func readPassword() tezos.PassphraseFunc {
    pwd := pass
    source := "command line"
    if pwd == "" {
        pwd = os.Getenv("TEZOS_KEY_PASSPHRASE")
        source = "env TEZOS_KEY_PASSPHRASE"
    }

    if pwd != "" {
        fmt.Printf("Using password from %s %s", source, strings.Repeat("*", len(pwd)))
        return func() ([]byte, error) { return []byte(pwd), nil }
    } else {
        return func() ([]byte, error) {
            fmt.Print("Enter Password: ")
            buf, err := term.ReadPassword(int(syscall.Stdin))
            fmt.Println()
            return buf, err
        }
    }
}

func pools(ctx context.Context, c *rpc.Client, addr tezos.Address) error {
    dex, err := OpenDex(ctx, c, addr)
    if err != nil {
        return err
    }
    fmt.Printf("Contract     %s\n", dex.Address())
    fmt.Printf("Name         %s\n", dex.Name())
    fmt.Printf("Pools  ------------------------------------------\n")
    table := make([][]string, 0)
    table = append(table, []string{
        "Id",
        "Pair",
        "Token A",
        "Token B",
        "Price",
        "LP",
        "A/LP",
        "B/LP",
    })
    for _, pool := range dex.Pools() {
        table = append(table, []string{
            fmt.Sprintf("%02d", pool.Id()),
            pool.A().Meta.Symbol + "/" + pool.B().Meta.Symbol,
            pool.A().String(),
            pool.B().String(),
            fmt.Sprintf("1 %s = %s", pool.A().Meta.Symbol, pool.MidPrice(AtoB)),
            pool.LP().String(),
            pool.A().Div(pool.LP()).String(),
            pool.B().Div(pool.LP()).String(),
        })
    }
    widths := make([]int, len(table[0]))
    for _, row := range table {
        for i := range row {
            widths[i] = max(widths[i], len(row[i]))
        }
    }
    for _, row := range table {
        for i := range row {
            fmt.Printf("%[1]*[2]s   ", -widths[i], row[i])
        }
        fmt.Println()
    }
    return nil
}

func info(ctx context.Context, c *rpc.Client, addr tezos.Address, id int64) error {
    dex, err := OpenDex(ctx, c, addr, id) // one pool only
    if err != nil {
        return err
    }
    pool, ok := dex.GetPool(id)
    if !ok {
        return fmt.Errorf("No such pool")
    }

    // info
    fmt.Printf("Contract     %s\n", dex.Address())
    fmt.Printf("Name         %s\n", dex.Name())
    fmt.Printf("Pool         %d\n", pool.Id())
    fmt.Printf("Token A ------------------------------------------\n")
    fmt.Printf("  Kind       %s\n", pool.A().Kind)
    if pool.A().Address.IsValid() {
        fmt.Printf("  Address    %s\n", pool.A().Address)
    }
    fmt.Printf("  Name       %s\n", pool.A().Meta.Name)
    fmt.Printf("  Symbol     %s\n", pool.A().Meta.Symbol)
    fmt.Printf("  Decimals   %d\n", pool.A().Meta.Decimals)
    fmt.Printf("Token B ------------------------------------------\n")
    fmt.Printf("  Kind       %s\n", pool.B().Kind)
    if pool.B().Address.IsValid() {
        fmt.Printf("  Address    %s\n", pool.B().Address)
    }
    if pool.B().Kind == contract.TokenKindFA2 {
        fmt.Printf("  Token Id   %d\n", pool.B().TokenId)
    }
    fmt.Printf("  Name       %s\n", pool.B().Meta.Name)
    fmt.Printf("  Symbol     %s\n", pool.B().Meta.Symbol)
    fmt.Printf("  Decimals   %d\n", pool.B().Meta.Decimals)
    fmt.Printf("Token LP ------------------------------------------\n")
    fmt.Printf("  Kind       %s\n", pool.LP().Kind)
    if pool.LP().Address.IsValid() {
        fmt.Printf("  Address    %s\n", pool.LP().Address)
    }
    if pool.LP().Kind == contract.TokenKindFA2 {
        fmt.Printf("  Token Id   %d\n", pool.LP().TokenId)
    }
    fmt.Printf("  Name       %s\n", pool.LP().Meta.Name)
    fmt.Printf("  Symbol     %s\n", pool.LP().Meta.Symbol)
    fmt.Printf("  Decimals   %d\n", pool.LP().Meta.Decimals)
    fmt.Printf("Pool Contents ------------------------------------\n")
    fmt.Printf("  A Supply   %s\n", pool.A())
    fmt.Printf("  B Supply   %s\n", pool.B())
    fmt.Printf("  LP Supply  %s\n", pool.LP())
    fmt.Printf("Mid Price ----------------------------------------\n")
    fmt.Printf("  1 %-6s = %s\n", pool.A().Meta.Symbol, pool.MidPrice(AtoB))
    fmt.Printf("  1 %-6s = %s\n", pool.B().Meta.Symbol, pool.MidPrice(BtoA))
    fmt.Printf("Simulated Swap (incl. %s fee and %s slippage) ----\n", pool.Fee(), slippage)
    amt := pool.A().ToAmount(1.0)
    fmt.Printf("  1 %-6s = %s\n", pool.A().Meta.Symbol, pool.MinimumOut(amt, AtoB, slippage))
    fmt.Printf("  1 %-6s = %s\n", pool.B().Meta.Symbol, pool.MinimumOut(pool.B().ToAmount(1.0), BtoA, slippage))
    if p := pool.PriceImpact(amt, AtoB) * 100; p < 0.01 {
        fmt.Printf("  Impact   < 0.01 %%\n")
    } else {
        fmt.Printf("  Impact     %.2f %%\n", p)
    }

    // load own token balances when signer is defined
    if c.Signer != nil {
        addrs, _ := c.Signer.ListAddresses(ctx)
        if len(addrs) > 0 {
            fmt.Printf("Balances for %s \n", addrs[0])
            bal, err := pool.A().GetBalance(ctx, addrs[0])
            if err != nil {
                return err
            }
            fmt.Printf("  A          %s\n", bal)
            bal, err = pool.B().GetBalance(ctx, addrs[0])
            if err != nil {
                return err
            }
            fmt.Printf("  B          %s\n", bal)
            bal, err = pool.LP().GetBalance(ctx, addrs[0])
            if err != nil {
                return err
            }
            fmt.Printf("  LP         %s\n", bal)
        }
    }

    return nil
}

func simulate(ctx context.Context, c *rpc.Client, addr tezos.Address, id int64, dir Direction, amount float64) error {
    dex, err := OpenDex(ctx, c, addr)
    if err != nil {
        return err
    }
    pool, ok := dex.GetPool(id)
    if !ok {
        return fmt.Errorf("No such pool")
    }

    addrs, _ := c.Signer.ListAddresses(ctx)

    in := pool.A()
    if dir == BtoA {
        in = pool.B()
    }
    in.WithAmountFloat64(amount)
    out := pool.MinimumOut(in.Amount, dir, slippage)

    if dir == BtoA {
        fmt.Printf("%s: Sell %s for %s on %s\n", dex.Name(), in, out, pool.Name())
    } else {
        fmt.Printf("%s: Buy %s for %s on %s\n", dex.Name(), out, in, pool.Name())
    }

    res, err := pool.SimulateTrade(in, dir, addrs[0], DefaultTimeout, slippage)
    if err != nil {
        return err
    }

    fmt.Printf("Pool Contents ------------------------------------\n")
    fmt.Printf("  A Supply   %s\n", pool.A())
    fmt.Printf("  B Supply   %s\n", pool.B())
    fmt.Printf("  LP Supply  %s\n", pool.LP())
    fmt.Printf("  Fee        %s\n", pool.Fee())
    fmt.Printf("Mid Price ----------------------------------------\n")
    fmt.Printf("  Before     1 %s = %s\n", in.Meta.Symbol, res.MidPrice)
    fmt.Printf("  After      1 %s = %s\n", in.Meta.Symbol, res.NextMidPrice)
    fmt.Printf("Simulated Swap (incl. %s fee and %s slippage) ----\n", pool.Fee(), slippage)
    fmt.Printf("  In         %s\n", res.VolumeBase)
    fmt.Printf("  MinOut     %s\n", res.VolumeQuote)
    fmt.Printf("  Fee        %s\n", res.TradingFee)
    fmt.Printf("  Price      %s\n", res.ExecutionPrice)
    fmt.Printf("  Impact     %.2f %%\n", res.PriceImpact*100)
    fmt.Printf("  Tx Fee     %s\n", Tez.WithAmount(tezos.NewZ(res.TxFee.Fee)))
    fmt.Printf("  Tx Burn    %s\n", Tez.WithAmount(tezos.NewZ(res.TxFee.Burn)))

    // amountz := pool.A().ToAmount(amount)
    // fmt.Printf("Pool Contents ------------------------------------\n")
    // fmt.Printf("  A Supply   %s\n", pool.A())
    // fmt.Printf("  B Supply   %s\n", pool.B())
    // fmt.Printf("  LP Supply  %s\n", pool.LP())
    // fmt.Printf("Mid Price ----------------------------------------\n")
    // fmt.Printf("  1 %-6s = %s\n", pool.A().Meta.Symbol, pool.MidPrice())
    // fmt.Printf("  1 %-6s = %s\n", pool.B().Meta.Symbol, pool.MidPriceInverse())
    // fmt.Printf("Swap Price (incl. fee) ---------------------------\n")
    // fmt.Printf("  %s %s = %s\n", amountz.Decimals(pool.A().Meta.Decimals), pool.A().Meta.Symbol, pool.Price(amountz))
    // fmt.Printf("  Fee        %s\n", pool.Fee(amountz))
    // fmt.Printf("  Impact     %.2f %%\n", pool.PriceImpact(amountz)*100)

    return nil
}

func swap(ctx context.Context, c *rpc.Client, addr tezos.Address, id int64, dir Direction, amount float64) error {
    return fmt.Errorf("Swap is not implemented yet, hang on.")
}

func max(x, y int) int {
    if x < y {
        return y
    }
    return x
}
