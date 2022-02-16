// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Contract and token examples
// - query metadata and views
// - mutations: require private key
//   - transfer
//   - approve
//   - updateOperator
//
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"blockwatch.cc/tzgo/contract"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
)

var (
	flags   = flag.NewFlagSet("contract", flag.ContinueOnError)
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
			fmt.Println("Usage: contract [flags] <cmd> [sub-args]")
			fmt.Println("\nFlags")
			flags.PrintDefaults()
			fmt.Println("\nQuery Commands")
			fmt.Printf("  run_view       <contract> <name> <data>     run view entrypoint `name` with JSON-encoded micheline input `data`\n")
			fmt.Printf("  info           <contract>                   load contract, print entrypoints and views\n")
			fmt.Printf("  metadata       <contract>                   fetch contract metadata\n")
			fmt.Printf("  balance_of     <contract> <owner>           FA2: fetch token balance for owner\n")
			fmt.Printf("  getBalance     <contract> <owner>           FA1: fetch token balance for owner\n")
			fmt.Printf("  getTotalSupply <contract>                   FA1: fetch total token supply\n")
			fmt.Printf("  getAllowance   <contract> <owner> <spender> FA1: fetch spender permit\n")
			fmt.Println("\nTransaction Commands (require private key")
			fmt.Printf("  transfer       <contract> <token_id> <amount> <receiver>  FA1+2: transfer tokens to receiver\n")
			fmt.Printf("  approve        <contract> <owner> <spender> <amount>      FA1: grant spending right\n")
			fmt.Printf("  addOperator    <contract> <token_id> <owner> <spender>    FA2: grant operator right\n")
			fmt.Printf("  removeOperator <contract> <token_id> <owner> <spender>    FA2: revoke operator right\n")
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

	switch cmd := flags.Arg(0); cmd {
	case "run_view":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return run_view(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
	case "info":
		if n < 2 {
			return fmt.Errorf("Missing contract address")
		}
		return info(ctx, c, flags.Arg(1))
	case "metadata":
		if n < 2 {
			return fmt.Errorf("Missing contract address")
		}
		return meta(ctx, c, flags.Arg(1))
	case "balance_of":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return balance_of(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
	case "getBalance":
		if n < 3 {
			return fmt.Errorf("Missing arguments")
		}
		return getBalance(ctx, c, flags.Arg(1), flags.Arg(2))
	case "getTotalSupply":
		if n < 2 {
			return fmt.Errorf("Missing arguments")
		}
		return getTotalSupply(ctx, c, flags.Arg(1))
	case "getAllowance":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return getAllowance(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
	default:
		return fmt.Errorf("Unknown command %q", cmd)
	}
}

func loadContract(ctx context.Context, c *rpc.Client, addr string, resolve bool) (*contract.Contract, error) {
	a, err := tezos.ParseAddress(addr)
	if err != nil {
		return nil, err
	}
	if a.Type != tezos.AddressTypeContract {
		return nil, fmt.Errorf("Invalid contract address")
	}
	con := contract.NewContract(a, c)
	if resolve {
		if err := con.Resolve(ctx); err != nil {
			return nil, err
		}
	}
	return con, nil
}

// go run ./examples/contract/ -node https://rpc.tzstats.com run_view KT1LRboPna9yQY9BrjtQYDS1DVxhKESK4VVd balance_of '[{"prim":"Pair","args":[{"string":"tz1UbRzhYjQKTtWYvGUWcRtVT4fN3NESDVYT"},{"int":"0"}]}]'
func run_view(ctx context.Context, c *rpc.Client, addr, name, in string) error {
	var prim micheline.Prim
	if err := prim.UnmarshalJSON([]byte(in)); err != nil {
		return err
	}
	con, err := loadContract(ctx, c, addr, true)
	if err != nil {
		return err
	}
	ep, ok := con.Entrypoint(name)
	if !ok {
		return fmt.Errorf("No such entrypoint")
	}
	if !ep.IsCallback() {
		return fmt.Errorf("Entrypoint is not a callback")
	}
	res, err := con.RunView(ctx, name, prim)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(res, "  ", "  ")
	fmt.Printf("Result:  \n%s\n", string(buf))
	return nil
}

func max(i, j int) int {
	if i < j {
		return j
	}
	return i
}

func maxColWith(rows [][]string) []int {
	if len(rows) == 0 {
		return nil
	}
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		for i := range widths {
			if len(row) < i {
				continue
			}
			widths[i] = max(widths[i], len(row[i]))
		}
	}
	return widths
}

func writeTable(w io.Writer, prefix string, names []string, rows [][]string) {
	cols := maxColWith(append([][]string{names}, rows...))
	var b strings.Builder
	b.WriteString(prefix)
	for i, n := range names {
		b.WriteString(n)
		b.WriteString(strings.Repeat(" ", cols[i]-len(n)+2))
	}
	b.WriteByte('\n')
	w.Write([]byte(b.String()))
	b.Reset()
	for _, row := range rows {
		b.WriteString(prefix)
		for j := range row {
			b.WriteString(row[j])
			b.WriteString(strings.Repeat(" ", cols[j]-len(row[j])+2))
		}
		b.WriteByte('\n')
		w.Write([]byte(b.String()))
		b.Reset()
	}
}

func info(ctx context.Context, c *rpc.Client, addr string) error {
	con, err := loadContract(ctx, c, addr, true)
	if err != nil {
		return err
	}
	fmt.Printf("Contract     %s\n", con.Address())
	fmt.Printf("FA1          %t\n", con.IsFA1())
	fmt.Printf("FA1.2        %t\n", con.IsFA12())
	fmt.Printf("FA2          %t\n", con.IsFA2())
	fmt.Printf("Manager.tz   %t\n", con.IsManagerTz())
	eps, _ := con.Script().Entrypoints(true)
	fmt.Printf("Entrypoints  %d\n", len(eps))
	rows := make([][]string, 0, len(eps))
	for n, ep := range eps {
		row := make([]string, 3)
		row[0] = n
		td := ep.Type().Typedef("")
		td.Name = ""
		row[1] = td.String()
		var std []string
		for _, iface := range micheline.WellKnownInterfaces {
			if iface.Contains(ep) {
				std = append(std, iface.String())
			}
		}
		if ep.IsCallback() {
			std = append(std, "CALLBACK")
		}
		row[2] = strings.Join(std, ",")
		rows = append(rows, row)
	}
	writeTable(os.Stdout, "  ", []string{"Name", "Type", "Interface"}, rows)
	views, _ := con.Script().Views(false, false)
	fmt.Printf("Views        %d\n", len(views))
	rows = make([][]string, 0, len(views))
	for n, v := range views {
		row := make([]string, 3)
		row[0] = n
		row[1] = v.Param.Typedef("").String()
		row[2] = v.Retval.Typedef("").String()
		rows = append(rows, row)
	}
	if len(views) > 0 {
		writeTable(os.Stdout, "  ", []string{"Name", "Params", "Return"}, rows)
	}
	return nil
}

func meta(ctx context.Context, c *rpc.Client, addr string) error {
	con, err := loadContract(ctx, c, addr, true)
	if err != nil {
		return err
	}
	m, err := con.ResolveMetadata(ctx)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(m, "  ", "  ")
	fmt.Printf("Result:  \n%s\n", string(buf))
	return nil
}

// FA2
func balance_of(ctx context.Context, c *rpc.Client, addr, owner, id string) error {
	con, err := loadContract(ctx, c, addr, false)
	if err != nil {
		return err
	}
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	own, err := tezos.ParseAddress(owner)
	if err != nil {
		return err
	}
	req := []contract.FA2BalanceRequest{
		contract.FA2BalanceRequest{
			Owner:   own,
			TokenId: tezos.NewZ(i),
		},
	}
	fa2 := con.AsFA2(i)
	res, err := fa2.GetBalances(ctx, req)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(res, "  ", "  ")
	fmt.Printf("Result:  \n%s\n", string(buf))
	return nil
}

// FA1
func getBalance(ctx context.Context, c *rpc.Client, addr, owner string) error {
	con, err := loadContract(ctx, c, addr, false)
	if err != nil {
		return err
	}
	own, err := tezos.ParseAddress(owner)
	if err != nil {
		return err
	}
	fa1 := con.AsFA1()
	res, err := fa1.GetBalance(ctx, own)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(res, "  ", "  ")
	fmt.Printf("Result: %s\n", string(buf))
	return nil
}

// FA1
func getAllowance(ctx context.Context, c *rpc.Client, addr, owner, spender string) error {
	con, err := loadContract(ctx, c, addr, false)
	if err != nil {
		return err
	}
	own, err := tezos.ParseAddress(owner)
	if err != nil {
		return err
	}
	spend, err := tezos.ParseAddress(spender)
	if err != nil {
		return err
	}
	fa1 := con.AsFA1()
	res, err := fa1.GetAllowance(ctx, own, spend)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(res, "  ", "  ")
	fmt.Printf("Result: %s\n", string(buf))
	return nil
}

// FA1
func getTotalSupply(ctx context.Context, c *rpc.Client, addr string) error {
	con, err := loadContract(ctx, c, addr, false)
	if err != nil {
		return err
	}
	fa1 := con.AsFA1()
	res, err := fa1.GetTotalSupply(ctx)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(res, "  ", "  ")
	fmt.Printf("Result: %s\n", string(buf))
	return nil
}
