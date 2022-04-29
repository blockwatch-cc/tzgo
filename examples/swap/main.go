// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Quipuswap examples
//
// currently works with
// - quipuswap v1 type DEX
// - quipuswap token to token DEX
//
// Examples
// - query pool info (ctez on Quipu v1)
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
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "math/big"
    "os"
    "strconv"

    "blockwatch.cc/tzgo/contract"
    "blockwatch.cc/tzgo/micheline"
    "blockwatch.cc/tzgo/rpc"
    // "blockwatch.cc/tzgo/signer"
    "blockwatch.cc/tzgo/tezos"
    "github.com/echa/log"
)

var (
    flags   = flag.NewFlagSet("swap", flag.ContinueOnError)
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
            fmt.Println("Usage: swap [flags] <cmd> [sub-args]")
            fmt.Println("\nFlags")
            flags.PrintDefaults()
            fmt.Println("\nCommands")
            fmt.Printf("  info  <KT1> <pool_id>             show Quipuswap pool info\n")
            fmt.Printf("  sim   <KT1> <pool_id> <in>        dry run a swap on `in` tokens\n")
            fmt.Printf("  swap  <KT1> <pool_id> <in> <pk>   execute swap of `in` tokens\n")
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
    case "info":
        if n < 2 {
            return fmt.Errorf("Missing arguments")
        }
        return info(ctx, c, flags.Arg(1), flags.Arg(2))
    case "sim":
        if n < 3 {
            return fmt.Errorf("Missing arguments")
        }
        return simulate(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
    case "swap":
        if n < 4 {
            return fmt.Errorf("Missing arguments")
        }
        return swap(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3), flags.Arg(4))
    default:
        return fmt.Errorf("Unknown command %q", cmd)
    }
}

type QuipuToken struct {
    FA12 tezos.Address `json:"fa12"`
    FA2  struct {
        TokenAddress tezos.Address `json:"token_address"`
        TokenId      int64         `json:"token_id,string"`
    } `json:"fa2"`
    IsTez    bool                    `json:"-"`
    Meta     *contract.TokenMetadata `json:"-"`
    Contract *contract.Contract      `json:"-"`
    Amount   tezos.Z                 `json:"-"`
}

var Tez = QuipuToken{
    IsTez: true,
    Meta: &contract.TokenMetadata{
        Symbol:   "tez",
        Name:     "Tezos",
        Decimals: 6,
    },
}

func (t QuipuToken) String() string {
    var (
        decimals int
        symbol   string
    )
    if t.Meta != nil {
        decimals = t.Meta.Decimals
        symbol = t.Meta.Symbol
    }
    return t.Amount.Decimals(decimals) + " " + symbol
}

func (t *QuipuToken) WithAmount(z tezos.Z) *QuipuToken {
    t.Amount = z.Clone()
    return t
}

func (t QuipuToken) ToAmount(in float64) tezos.Z {
    f := new(big.Float).SetFloat64(in)
    f = f.Mul(f, pow10f(t.Meta.Decimals))
    i, _ := f.Int(nil)
    var z tezos.Z
    z.Set(i)
    return z
}

func (t QuipuToken) Address() tezos.Address {
    if t.FA12.IsValid() {
        return t.FA12
    }
    return t.FA2.TokenAddress
}

func (t QuipuToken) TokenKind() contract.TokenKind {
    switch {
    case t.IsTez:
        return contract.TokenKindTez
    case t.FA12.IsValid():
        return contract.TokenKindFA1_2
    case t.FA2.TokenAddress.IsValid():
        return contract.TokenKindFA2
    default:
        return contract.TokenKindInvalid
    }
}

func (t *QuipuToken) Resolve(ctx context.Context, c *rpc.Client) (err error) {
    switch {
    case t.IsTez:
        // empty
    case t.FA12.IsValid():
        if t.Contract == nil {
            t.Contract = contract.NewContract(t.FA12, c)
        }
        if t.Meta == nil {
            t.Meta, err = t.Contract.AsFA1().ResolveMetadata(ctx)
        }
    case t.FA2.TokenAddress.IsValid():
        if t.Contract == nil {
            t.Contract = contract.NewContract(t.FA2.TokenAddress, c)
        }
        if t.Meta == nil {
            t.Meta, err = t.Contract.AsFA2(t.FA2.TokenId).ResolveMetadata(ctx)
        }
    default:
        err = fmt.Errorf("token not initialized")
    }
    return
}

// storage.pairs
type QuipuPoolContent struct {
    A      tezos.Z `json:"token_a_pool"`
    B      tezos.Z `json:"token_b_pool"`
    Supply tezos.Z `json:"total_supply"`
}

func (c *QuipuPoolContent) UnmarshalJSON(buf []byte) error {
    if bytes.Contains(buf, []byte(`"tez_pool"`)) {
        type alias struct {
            A      tezos.Z `json:"tez_pool"`
            B      tezos.Z `json:"token_pool"`
            Supply tezos.Z `json:"total_supply"`
        }
        type store struct {
            Storage *alias `json:"storage"`
        }
        a := store{Storage: (*alias)(c)}
        return json.Unmarshal(buf, &a)
    } else {
        type alias QuipuPoolContent
        return json.Unmarshal(buf, (*alias)(c))
    }
}

// storage.tokens
type QuipuPoolInfo struct {
    Address  tezos.Address      `json:"-"`
    Id       int64              `json:"-"`
    Contract *contract.Contract `json:"-"`
    A        QuipuToken         `json:"token_a_type"`
    B        QuipuToken         `json:"token_b_type"`
    LP       QuipuToken         `json:"-"`
}

func (i *QuipuPoolInfo) Resolve(ctx context.Context, c *rpc.Client) error {
    if err := i.A.Resolve(ctx, c); err != nil {
        return err
    }
    if err := i.B.Resolve(ctx, c); err != nil {
        return err
    }
    return i.LP.Resolve(ctx, c)
}

func (i *QuipuPoolInfo) WithContentFrom(val micheline.Value) error {
    var content QuipuPoolContent
    if err := val.Unmarshal(&content); err != nil {
        return fmt.Errorf("decoding pool info from value: %v", err)
    }
    i.WithContent(content)
    return nil
}

func (i *QuipuPoolInfo) WithContent(c QuipuPoolContent) *QuipuPoolInfo {
    i.A.Amount = c.A.Clone()
    i.B.Amount = c.B.Clone()
    i.LP.Amount = c.Supply.Clone()
    return i
}

// 10**decimals
func pow10(i int) *big.Int {
    return big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(i)), nil)
}

func pow10f(i int) *big.Float {
    return new(big.Float).SetInt(pow10(i))
}

// Current price of token A denominated in B? How many B can one A buy?
func (i QuipuPoolInfo) MidPrice() QuipuToken {
    var z tezos.Z
    if !i.B.Amount.IsZero() && !i.A.Amount.IsZero() {
        price := big.NewInt(0).Set(i.B.Amount.Big())
        price = price.Mul(price, pow10(i.A.Meta.Decimals))
        price = price.Div(price, i.A.Amount.Big())
        z.Set(price)
    }
    tok := i.B
    tok.Amount = z
    return tok
}

func (i QuipuPoolInfo) MidPriceInverse() QuipuToken {
    var z tezos.Z
    if !i.B.Amount.IsZero() && !i.A.Amount.IsZero() {
        price := big.NewInt(0).Set(i.A.Amount.Big())
        price = price.Mul(price, pow10(i.B.Meta.Decimals))
        price = price.Div(price, i.B.Amount.Big())
        z.Set(price)
    }
    tok := i.A
    tok.Amount = z
    return tok
}

// how many token A would the AMM deduct before swapping?
func (i QuipuPoolInfo) Fee(amount tezos.Z) QuipuToken {
    in := new(big.Int).Mul(amount.Big(), big.NewInt(3))
    fee := in.Div(in, big.NewInt(1000))
    var z tezos.Z
    z.Set(fee)
    tok := i.A
    tok.Amount = z
    return tok
}

// how many token B can given amount of A buy, excluding fees?
func (i QuipuPoolInfo) Price(amount tezos.Z) QuipuToken {
    tok := i.B
    tok.Amount = quipuPrice(amount, i.A.Amount, i.B.Amount)
    return tok
}

// how many token A can given amount of B buy, excluding fees?
func (i QuipuPoolInfo) PriceInverse(amount tezos.Z) QuipuToken {
    tok := i.A
    tok.Amount = quipuPrice(amount, i.B.Amount, i.A.Amount)
    return tok
}

// How much does token B price change when selling amount A for B?
func (i QuipuPoolInfo) PriceImpact(amount tezos.Z) float64 {
    return quipuPriceImpact(amount, i.A.Amount, i.B.Amount)
}

// func (i QuipuPoolInfo) PriceImpactInverse(amount tezos.Z) float64 {
//     return quipuPriceImpact(amount, i.B.Amount, i.A.Amount)
// }

func quipuPrice(amount, a, b tezos.Z) tezos.Z {
    var z tezos.Z
    if a.IsZero() || b.IsZero() {
        return z
    }

    //     b * amount * 997
    // -----------------------
    // a * 1000 + amount * 997
    in := big.NewInt(0).Set(amount.Big())
    inWithFee := in.Mul(in, big.NewInt(997))
    num := big.NewInt(0).Mul(inWithFee, b.Big())
    den := big.NewInt(0).Set(a.Big())
    den = den.Mul(den, big.NewInt(1000))
    den = den.Add(den, inWithFee)
    out := num.Div(num, den)
    z.Set(out)
    return z
}

func quipuPriceImpact(amount, a, b tezos.Z) float64 {
    // swap price impact
    // `1 - mid_price_after_trade / mid_price_current` where
    // mid-price is `tokenPool / xtzPool`
    af, _ := new(big.Float).SetInt(a.Big()).Float64()
    bf, _ := new(big.Float).SetInt(b.Big()).Float64()
    before := bf / af

    out := quipuPrice(amount, a, b)
    aDiff, _ := new(big.Float).SetInt(amount.Big()).Float64()
    bDiff, _ := new(big.Float).SetInt(out.Big()).Float64()
    af += aDiff
    bf -= bDiff
    after := bf / af

    return 1.0 - after/before
}

func loadDex(ctx context.Context, c *rpc.Client, addr, pool string) (*QuipuPoolInfo, error) {
    a, err := tezos.ParseAddress(addr)
    if err != nil {
        return nil, err
    }
    if a.Type != tezos.AddressTypeContract {
        return nil, fmt.Errorf("Not a contract address")
    }
    dex := contract.NewContract(a, c)
    if err := dex.Resolve(ctx); err != nil {
        return nil, err
    }
    if _, err := dex.ResolveMetadata(ctx); err != nil {
        return nil, err
    }

    info := QuipuPoolInfo{
        Address:  dex.Address(),
        Contract: dex,
    }

    // we detect which Quipswap dex it is from its metadata
    dexmeta := dex.Metadata()
    dexstore := dex.StorageValue()

    switch dexmeta.Name {
    case "Quipuswap Exchange: Token Edition":
        // token
        poolid, err := strconv.ParseInt(pool, 10, 64)
        if err != nil {
            return nil, fmt.Errorf("Invalid pool id %q: %v", pool, err)
        }
        // identities
        tokens, err := dex.GetBigmapValue(ctx, "storage.tokens", micheline.NewNat(big.NewInt(poolid)))
        if err != nil {
            return nil, fmt.Errorf("Fetching token data: %v", err)
        }
        if err := tokens.Unmarshal(&info); err != nil {
            return nil, fmt.Errorf("Decoding pool info: %v", err)
        }
        info.Id = poolid
        info.LP = QuipuToken{Contract: dex}
        info.LP.FA2.TokenAddress = dex.Address()

        // resolve tokens
        if err := info.Resolve(ctx, c); err != nil {
            return nil, fmt.Errorf("Resolve token metadata: %v", err)
        }

        // resolve pool supply
        pairpool, err := dex.GetBigmapValue(ctx, "storage.pairs", micheline.NewNat(big.NewInt(poolid)))
        if err != nil {
            return nil, fmt.Errorf("Fetching pool data: %v", err)
        }
        if err := info.WithContentFrom(*pairpool); err != nil {
            return nil, fmt.Errorf("Decoding pool data: %v", err)
        }

    case "Quipu Token":
        // v1: only tez -> FA1.2/FA2 pools
        // 1 identities
        tokaddr, ok := dexstore.GetAddress("storage.token_address")
        if !ok {
            return nil, fmt.Errorf("%s: missing or invalid token address in storage %s", tokaddr, dex.Storage().Dump())
        }
        tokenid, ok := dexstore.GetInt64("storage.token_id")
        var token QuipuToken
        if ok {
            token.FA2.TokenAddress = tokaddr
            token.FA2.TokenId = tokenid
        } else {
            token.FA12 = tokaddr
        }
        if err := token.Resolve(ctx, c); err != nil {
            return nil, fmt.Errorf("%s: loading token: %v", tokaddr, err)
        }
        lp := QuipuToken{Contract: dex}
        lp.FA2.TokenAddress = dex.Address()
        if err := lp.Resolve(ctx, c); err != nil {
            return nil, fmt.Errorf("%s: loading LP token: %v", dex.Address(), err)
        }
        info.A = Tez
        info.B = token
        info.LP = lp

        // 2 supply
        if err := info.WithContentFrom(dexstore); err != nil {
            return nil, fmt.Errorf("Decoding pool data: %v", err)
        }

    default:
        return nil, fmt.Errorf("Contract does not look like a Quipuswap pool, says: %s - %s",
            dex.Metadata().Name, dex.Metadata().Description)
    }

    return &info, nil
}

func info(ctx context.Context, c *rpc.Client, addr, pool string) error {
    dex, err := loadDex(ctx, c, addr, pool)
    if err != nil {
        return err
    }

    // info
    fmt.Printf("Contract     %s\n", dex.Address)
    fmt.Printf("Name         %s\n", dex.Contract.Metadata().Name)
    fmt.Printf("Token A ------------------------------------------\n")
    fmt.Printf("  Kind       %s\n", dex.A.TokenKind())
    fmt.Printf("  Address    %s\n", dex.A.Address())
    fmt.Printf("  Name       %s\n", dex.A.Meta.Name)
    fmt.Printf("  Symbol     %s\n", dex.A.Meta.Symbol)
    fmt.Printf("  Decimals   %d\n", dex.A.Meta.Decimals)
    fmt.Printf("Token B ------------------------------------------\n")
    fmt.Printf("  Kind       %s\n", dex.B.TokenKind())
    fmt.Printf("  Address    %s\n", dex.B.Address())
    if dex.B.TokenKind() == contract.TokenKindFA2 {
        fmt.Printf("  Token Id   %d\n", dex.B.FA2.TokenId)
    }
    fmt.Printf("  Name       %s\n", dex.B.Meta.Name)
    fmt.Printf("  Symbol     %s\n", dex.B.Meta.Symbol)
    fmt.Printf("  Decimals   %d\n", dex.B.Meta.Decimals)
    fmt.Printf("Pool Contents ------------------------------------\n")
    fmt.Printf("  A Supply   %s\n", dex.A)
    fmt.Printf("  B Supply   %s\n", dex.B)
    fmt.Printf("  LP Supply  %s\n", dex.LP)
    fmt.Printf("Mid Price ----------------------------------------\n")
    fmt.Printf("  1 %-6s = %s\n", dex.A.Meta.Symbol, dex.MidPrice())
    fmt.Printf("  1 %-6s = %s\n", dex.B.Meta.Symbol, dex.MidPriceInverse())
    fmt.Printf("Swap Price (incl. fee) ---------------------------\n")
    amt := dex.A.ToAmount(1.0)
    fmt.Printf("  1 %-6s = %s\n", dex.A.Meta.Symbol, dex.Price(amt))
    fmt.Printf("  1 %-6s = %s\n", dex.B.Meta.Symbol, dex.PriceInverse(dex.B.ToAmount(1.0)))
    if p := dex.PriceImpact(amt) * 100; p < 0.01 {
        fmt.Printf("  Impact   < 0.01 %%\n")
    } else {
        fmt.Printf("  Impact     %.2f %%\n", p)
    }
    return nil
}

func simulate(ctx context.Context, c *rpc.Client, addr, pool, amount string) error {
    amountf, err := strconv.ParseFloat(amount, 64)
    if err != nil {
        return fmt.Errorf("Invalid amount %s: %v", amount, err)
    }
    dex, err := loadDex(ctx, c, addr, pool)
    if err != nil {
        return err
    }

    amountz := dex.A.ToAmount(amountf)

    fmt.Printf("Pool Contents ------------------------------------\n")
    fmt.Printf("  A Supply   %s\n", dex.A)
    fmt.Printf("  B Supply   %s\n", dex.B)
    fmt.Printf("  LP Supply  %s\n", dex.LP)
    fmt.Printf("Mid Price ----------------------------------------\n")
    fmt.Printf("  1 %-6s = %s\n", dex.A.Meta.Symbol, dex.MidPrice())
    fmt.Printf("  1 %-6s = %s\n", dex.B.Meta.Symbol, dex.MidPriceInverse())
    fmt.Printf("Swap Price (incl. fee) ---------------------------\n")
    fmt.Printf("  %s %s = %s\n", amountz.Decimals(dex.A.Meta.Decimals), dex.A.Meta.Symbol, dex.Price(amountz))
    fmt.Printf("  Fee        %s\n", dex.Fee(amountz))
    fmt.Printf("  Impact     %.2f %%\n", dex.PriceImpact(amountz)*100)

    return nil
}

func swap(ctx context.Context, c *rpc.Client, addr, pool, amount, key string) error {
    return fmt.Errorf("Swap is not implemented yet, hang on.")
}
