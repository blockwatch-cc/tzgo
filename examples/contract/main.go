// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Contract and token examples
// - display contract info
// - query contract and token metadata
// - execute views
// - mutations (private key required)
//   - transfer
//   - approve
//   - updateOperator
package main

import (
	"bytes"
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
	"blockwatch.cc/tzgo/signer"
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
	flags.StringVar(&node, "node", "https://rpc.tzpro.io", "Tezos node URL")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Usage: contract [flags] <cmd> [sub-args]")
			fmt.Println("\nFlags")
			flags.PrintDefaults()
			fmt.Println("\nQuery Commands")
			fmt.Printf("  run_view       <contract> <name> <data>     run on-chain view `name` with JSON-encoded micheline input `data`\n")
			fmt.Printf("  run_callback   <contract> <name> <data>     run tzip4 view/callback entrypoint `name` with JSON-encoded micheline input `data`\n")
			fmt.Printf("  run_tz16       <contract> <name> <data>     run tzip16 view code `name` with JSON-encoded micheline input `data`\n")
			fmt.Printf("  info           <contract>                   load contract, print entrypoints and views\n")
			fmt.Printf("  metadata       <contract>                   fetch contract metadata\n")
			fmt.Printf("  token_metadata <contract> <token_id>        fetch token metadata\n")
			fmt.Printf("  balance_of     <contract> <owner>           FA2: fetch token balance for owner\n")
			fmt.Printf("  getBalance     <contract> <owner>           FA1: fetch token balance for owner\n")
			fmt.Printf("  getTotalSupply <contract>                   FA1: fetch total token supply\n")
			fmt.Printf("  getAllowance   <contract> <owner> <spender> FA1: fetch spender permit\n")
			fmt.Println("\nTransaction Commands (require private key")
			fmt.Printf("  transfer       <contract> <token_id> <amount> <receiver> <privkey> FA1+2: transfer tokens to receiver\n")
			fmt.Printf("  approve        <contract> <spender> <amount> <privkey>     FA1: grant spending right\n")
			fmt.Printf("  revoke         <contract> <spender> <amount> <privkey>     FA1: revoke spending right\n")
			fmt.Printf("  addOperator    <contract> <token_id> <spender> <privkey>   FA2: grant full operator permissions\n")
			fmt.Printf("  removeOperator <contract> <token_id> <spender> <privkey>   FA2: revoke full operator permissions\n")
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
	case "run_view":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return run_view(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
	case "run_callback":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return run_callback(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
	case "run_tz16":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return run_tz16(ctx, c, flags.Arg(1), flags.Arg(2), flags.Arg(3))
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
	case "token_metadata":
		if n < 3 {
			return fmt.Errorf("Missing arguments")
		}
		return token_meta(ctx, c, flags.Arg(1), flags.Arg(2))
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
	case "transfer":
		if n < 5 {
			return fmt.Errorf("Missing arguments")
		}
		return transfer(ctx, c)
	case "approve":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return approve(ctx, c)
	case "revoke":
		if n < 3 {
			return fmt.Errorf("Missing arguments")
		}
		return revoke(ctx, c)
	case "addOperator":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return addOperator(ctx, c)
	case "removeOperator":
		if n < 4 {
			return fmt.Errorf("Missing arguments")
		}
		return removeOperator(ctx, c)
	default:
		return fmt.Errorf("Unknown command %q", cmd)
	}
}

func loadContract(ctx context.Context, c *rpc.Client, addr string, resolve bool) (*contract.Contract, error) {
	a, err := tezos.ParseAddress(addr)
	if err != nil {
		return nil, err
	}
	if a.Type() != tezos.AddressTypeContract {
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

// go run ./examples/contract/ -node https://rpc.tzpro.io run_callback KT1LRboPna9yQY9BrjtQYDS1DVxhKESK4VVd balance_of '[{"prim":"Pair","args":[{"string":"tz1UbRzhYjQKTtWYvGUWcRtVT4fN3NESDVYT"},{"int":"0"}]}]'
func run_tz16(ctx context.Context, c *rpc.Client, addr, name, in string) error {
	var prim micheline.Prim
	if err := prim.UnmarshalJSON([]byte(in)); err != nil {
		return err
	}
	con, err := loadContract(ctx, c, addr, true)
	if err != nil {
		return err
	}
	meta, err := con.ResolveMetadata(ctx)
	if err != nil {
		return fmt.Errorf("No tz16 metadata: %v", err)
	}
	view := meta.GetView(name)
	if view.Name != name {
		return fmt.Errorf("No such tz16 view")
	}
	res, err := view.Run(ctx, con, prim)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(res, "  ", "  ")
	fmt.Printf("Result:  \n%s\n", string(buf))
	return nil
}

// go run ./examples/contract/ -node https://rpc.tzpro.io run_callback KT1LRboPna9yQY9BrjtQYDS1DVxhKESK4VVd balance_of '[{"prim":"Pair","args":[{"string":"tz1UbRzhYjQKTtWYvGUWcRtVT4fN3NESDVYT"},{"int":"0"}]}]'
func run_callback(ctx context.Context, c *rpc.Client, addr, name, in string) error {
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
	res, err := con.RunCallback(ctx, name, prim)
	if err != nil {
		return err
	}
	buf, _ := json.MarshalIndent(res, "  ", "  ")
	fmt.Printf("Result:  \n%s\n", string(buf))
	return nil
}

// go run ./examples/contract/ -node https://rpc.tzpro.io run_view KT1LjmAdYQCLBjwv4S2oFkEzyHVkomAf5MrW royalty '{"int":"31212"}'
func run_view(ctx context.Context, c *rpc.Client, addr, name, in string) error {
	var prim micheline.Prim
	if err := prim.UnmarshalJSON([]byte(in)); err != nil {
		return err
	}
	con, err := loadContract(ctx, c, addr, true)
	if err != nil {
		return err
	}
	_, ok := con.View(name)
	if !ok {
		return fmt.Errorf("No such view")
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
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("  ", "  ")
	_ = enc.Encode(m)
	fmt.Printf("Result:  \n%s\n", buf.String())
	return nil
}

func token_meta(ctx context.Context, c *rpc.Client, addr, id string) error {
	con, err := loadContract(ctx, c, addr, true)
	if err != nil {
		return err
	}
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	var meta *contract.TokenMetadata
	if con.IsFA1() || con.IsFA12() {
		meta, err = con.AsFA1().ResolveMetadata(ctx)
	} else if con.IsFA2() {
		meta, err = con.AsFA2(i).ResolveMetadata(ctx)
	}
	if err != nil {
		return err
	}
	if meta == nil {
		return fmt.Errorf("Not FA1/2 compatible")
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("  ", "  ")
	_ = enc.Encode(meta)
	fmt.Printf("Result:  \n%s\n", buf.String())
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

// FA1/2
// <contract> <token_id> <amount> <receiver> <privkey>
func transfer(ctx context.Context, c *rpc.Client) error {
	fContract := flags.Arg(1)
	fId := flags.Arg(2)
	fAmount := flags.Arg(3)
	fReceiver := flags.Arg(4)
	fKey := flags.Arg(5)

	id, err := strconv.ParseInt(fId, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid token id %q: %v", fId, err)
	}
	var amount tezos.Z
	if err := amount.UnmarshalText([]byte(fAmount)); err != nil {
		return fmt.Errorf("Invalid amount %q: %v", fAmount, err)
	}
	to, err := tezos.ParseAddress(fReceiver)
	if err != nil {
		return fmt.Errorf("Invalid receiver %q: %v", fReceiver, err)
	}
	key, err := tezos.ParsePrivateKey(fKey)
	if err != nil {
		return fmt.Errorf("Invalid private key %q: %v", fKey, err)
	}

	opts := rpc.DefaultOptions
	opts.Signer = signer.NewFromKey(key)
	from := key.Address()

	// load contract
	con, err := loadContract(ctx, c, fContract, true)
	if err != nil {
		return err
	}

	if con.IsFA1() || con.IsFA12() {
		args := con.AsFA1().Transfer(from, to, amount)
		_, err := con.Call(ctx, args, &opts)
		return err
	}

	if con.IsFA2() {
		args := con.AsFA2(id).Transfer(from, to, amount)
		_, err := con.Call(ctx, args, &opts)
		return err
	}

	return fmt.Errorf("Contract is not FA1 or FA2 compatible.")
}

// FA1 approve  <contract> <spender> <amount> <privkey>
func approve(ctx context.Context, c *rpc.Client) error {
	fContract := flags.Arg(1)
	fReceiver := flags.Arg(2)
	fAmount := flags.Arg(3)
	fKey := flags.Arg(4)

	var amount tezos.Z
	if err := amount.UnmarshalText([]byte(fAmount)); err != nil {
		return fmt.Errorf("Invalid amount %q: %v", fAmount, err)
	}
	to, err := tezos.ParseAddress(fReceiver)
	if err != nil {
		return fmt.Errorf("Invalid spender %q: %v", fReceiver, err)
	}
	key, err := tezos.ParsePrivateKey(fKey)
	if err != nil {
		return fmt.Errorf("Invalid private key %q: %v", fKey, err)
	}

	opts := rpc.DefaultOptions
	opts.Signer = signer.NewFromKey(key)

	// load contract
	con, err := loadContract(ctx, c, fContract, true)
	if err != nil {
		return err
	}

	if !(con.IsFA1() || con.IsFA12()) {
		return fmt.Errorf("Contract is not FA1 compatible.")
	}

	args := con.AsFA1().Approve(to, amount)
	_, err = con.Call(ctx, args, &opts)
	return err
}

// FA1 revoke <contract> <spender> <privkey>
func revoke(ctx context.Context, c *rpc.Client) error {
	fContract := flags.Arg(1)
	fReceiver := flags.Arg(2)
	fKey := flags.Arg(3)

	to, err := tezos.ParseAddress(fReceiver)
	if err != nil {
		return fmt.Errorf("Invalid spender %q: %v", fReceiver, err)
	}
	key, err := tezos.ParsePrivateKey(fKey)
	if err != nil {
		return fmt.Errorf("Invalid private key %q: %v", fKey, err)
	}

	opts := rpc.DefaultOptions
	opts.Signer = signer.NewFromKey(key)

	// load contract
	con, err := loadContract(ctx, c, fContract, true)
	if err != nil {
		return err
	}

	if !(con.IsFA1() || con.IsFA12()) {
		return fmt.Errorf("Contract is not FA1 compatible.")
	}

	args := con.AsFA1().Revoke(to)
	_, err = con.Call(ctx, args, &opts)
	return err
}

// FA2 addOperator    <contract> <token_id> <spender> <privkey>
func addOperator(ctx context.Context, c *rpc.Client) error {
	fContract := flags.Arg(1)
	fId := flags.Arg(2)
	fReceiver := flags.Arg(3)
	fKey := flags.Arg(4)

	id, err := strconv.ParseInt(fId, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid token id %q: %v", fId, err)
	}
	to, err := tezos.ParseAddress(fReceiver)
	if err != nil {
		return fmt.Errorf("Invalid spender %q: %v", fReceiver, err)
	}
	key, err := tezos.ParsePrivateKey(fKey)
	if err != nil {
		return fmt.Errorf("Invalid private key %q: %v", fKey, err)
	}

	opts := rpc.DefaultOptions
	opts.Signer = signer.NewFromKey(key)
	from := key.Address()

	// load contract
	con, err := loadContract(ctx, c, fContract, true)
	if err != nil {
		return err
	}

	if !con.IsFA2() {
		return fmt.Errorf("Contract is not FA2 compatible.")
	}

	args := con.AsFA2(id).AddOperator(from, to)
	_, err = con.Call(ctx, args, &opts)
	return err
}

// FA2 removeOperator <contract> <token_id> <spender> <privkey>
func removeOperator(ctx context.Context, c *rpc.Client) error {
	fContract := flags.Arg(1)
	fId := flags.Arg(2)
	fReceiver := flags.Arg(3)
	fKey := flags.Arg(4)

	id, err := strconv.ParseInt(fId, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid token id %q: %v", fId, err)
	}
	to, err := tezos.ParseAddress(fReceiver)
	if err != nil {
		return fmt.Errorf("Invalid spender %q: %v", fReceiver, err)
	}
	key, err := tezos.ParsePrivateKey(fKey)
	if err != nil {
		return fmt.Errorf("Invalid private key %q: %v", fKey, err)
	}

	opts := rpc.DefaultOptions
	opts.Signer = signer.NewFromKey(key)
	from := key.Address()

	// load contract
	con, err := loadContract(ctx, c, fContract, true)
	if err != nil {
		return err
	}

	if !con.IsFA2() {
		return fmt.Errorf("Contract is not FA2 compatible.")
	}

	args := con.AsFA2(id).RemoveOperator(from, to)
	_, err = con.Call(ctx, args, &opts)
	return err
}
