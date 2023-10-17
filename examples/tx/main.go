// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Transaction examples
//
// # Requirements
//
// - private key for a funded testnet or mainnet account (https://teztnets.xyz/)
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/signer/remote"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
)

var (
	flags   = flag.NewFlagSet("tx", flag.ContinueOnError)
	verbose bool
	node    string
	sk      tezos.PrivateKey
)

func init() {
	if k := os.Getenv("TZGO_PRIVATE_KEY"); k != "" {
		sk = tezos.MustParsePrivateKey(k)
	}
}

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
	flags.StringVar(&node, "node", "https://rpc.tzpro.io", "Tezos node URL")
	// flags.Var(&sk, "sk", "")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Usage: tx [args] <cmd> [sub-args]")
			fmt.Println("\nArguments")
			flags.PrintDefaults()
			fmt.Println("\nCommands")
			fmt.Printf("  describe <type>            describe JSON data required for op `type`\n")
			fmt.Printf("  encode <type> <data>       generate operation `type` from JSON `data`\n")
			fmt.Printf("  validate <type> <data>     compare local encoding against remote encoding\n")
			fmt.Printf("  decode <msg>               decode binary operation\n")
			fmt.Printf("  digest <msg>               generate operation digest for signing\n")
			fmt.Printf("  sign <msg>                 sign message digest\n")
			fmt.Printf("  sign-remote <addr> <msg>   sign message digest using remote signer\n")
			fmt.Printf("  simulate <msg>             simulate executing an operation using a fake signature\n")
			fmt.Printf("  broadcast <msg> <sig>      broadcast signed operation\n")
			fmt.Printf("  wait <ophash> [<n>]        waits for operation to be included after n confirmations (optional)\n")
			fmt.Printf("  send <type> <data>         perform all steps at once\n")
			fmt.Println("\nOperation types")
			fmt.Println("  endorsement")
			fmt.Println("  preendorsement")
			fmt.Println("  double_baking_evidence")
			fmt.Println("  double_endorsement_evidence")
			fmt.Println("  double_preendorsement_evidence")
			fmt.Println("  seed_nonce_revelation")
			fmt.Println("  activate_account")
			fmt.Println("  reveal")
			fmt.Println("  transaction")
			fmt.Println("  origination")
			fmt.Println("  delegation")
			fmt.Println("  proposals")
			fmt.Println("  ballot")
			fmt.Println("  register_global_constant")
			fmt.Println("  set_deposits_limit")
			fmt.Println("  failing_noop")
			fmt.Println("  transfer_ticket")
			fmt.Println("  vdf_revelation")
			fmt.Println("  increase_paid_storage")
			fmt.Println("  drain_delegate")
			fmt.Println("  update_consensus_key")
			fmt.Println("  smart_rollup_originate")
			fmt.Println("  smart_rollup_add_messages")
			fmt.Println("  smart_rollup_cement")
			fmt.Println("  smart_rollup_publish")
			fmt.Println("  smart_rollup_refute")
			fmt.Println("  smart_rollup_timeout")
			fmt.Println("  smart_rollup_execute_outbox_message")
			fmt.Println("  smart_rollup_recover_bond")
			fmt.Println("  dal_attestation")
			fmt.Println("  dal_publish_slot_header")
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
	case "describe":
		if n < 2 {
			return fmt.Errorf("Missing type")
		}
		return describe(ctx, c, flags.Arg(1))
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
		if n < 2 {
			return fmt.Errorf("Missing message")
		}
		return sign(ctx, c, flags.Arg(1))
	case "sign_remote":
		if n < 3 {
			return fmt.Errorf("Missing address or message")
		}
		return sign_remote(ctx, c, flags.Arg(1), flags.Arg(2))
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
	case "send":
		return send(ctx, c, flags.Arg(1), flags.Arg(2))
	default:
		return fmt.Errorf("Unknown command %q", cmd)
	}
}

func makeOp(c *rpc.Client, t, data string) (codec.Operation, error) {
	var o codec.Operation
	switch tezos.ParseOpType(t) {
	case tezos.OpTypeActivateAccount:
		o = new(codec.ActivateAccount)
	case tezos.OpTypeDoubleBakingEvidence:
		o = new(codec.DoubleBakingEvidence)
	case tezos.OpTypeDoubleEndorsementEvidence:
		if c.Params.OperationTagsVersion < 2 {
			o = new(codec.DoubleEndorsementEvidence)
		} else {
			o = new(codec.TenderbakeDoubleEndorsementEvidence)
		}
	case tezos.OpTypeDoublePreendorsementEvidence:
		o = new(codec.TenderbakeDoublePreendorsementEvidence)
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
	case tezos.OpTypePreendorsement:
		o = new(codec.TenderbakePreendorsement)
	case tezos.OpTypeEndorsement:
		if c.Params.OperationTagsVersion < 2 {
			o = new(codec.Endorsement)
		} else {
			o = new(codec.TenderbakeEndorsement)
		}
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
	case tezos.OpTypeSetDepositsLimit:
		o = new(codec.SetDepositsLimit)
	case tezos.OpTypeTransferTicket:
		o = new(codec.TransferTicket)
	case tezos.OpTypeVdfRevelation:
		o = new(codec.VdfRevelation)
	case tezos.OpTypeIncreasePaidStorage:
		o = new(codec.IncreasePaidStorage)
	case tezos.OpTypeDrainDelegate:
		o = new(codec.DrainDelegate)
	case tezos.OpTypeUpdateConsensusKey:
		o = new(codec.UpdateConsensusKey)
	case tezos.OpTypeSmartRollupOriginate:
		o = new(codec.SmartRollupOriginate)
	case tezos.OpTypeSmartRollupAddMessages:
		o = new(codec.SmartRollupAddMessages)
	case tezos.OpTypeSmartRollupCement:
		o = new(codec.SmartRollupCement)
	case tezos.OpTypeSmartRollupPublish:
		o = new(codec.SmartRollupPublish)
	case tezos.OpTypeSmartRollupRefute:
		o = new(codec.SmartRollupRefute)
	case tezos.OpTypeSmartRollupTimeout:
		o = new(codec.SmartRollupTimeout)
	case tezos.OpTypeSmartRollupExecuteOutboxMessage:
		o = new(codec.SmartRollupExecuteOutboxMessage)
	case tezos.OpTypeSmartRollupRecoverBond:
		o = new(codec.SmartRollupRecoverBond)
	case tezos.OpTypeDalAttestation:
		o = new(codec.DalAttestation)
	case tezos.OpTypeDalPublishSlotHeader:
		o = new(codec.DalPublishSlotHeader)
	default:
		return nil, fmt.Errorf("Unsupported op type %q", t)
	}
	if err := json.Unmarshal([]byte(data), &o); err != nil {
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

func describe(ctx context.Context, c *rpc.Client, typ string) error {
	o, err := makeOp(c, typ, "{}")
	if err != nil {
		return err
	}
	fmt.Println("op type", typ)
	fmt.Println("-----------------------------------------------")
	printTypeInfo(reflect.Indirect(reflect.ValueOf(o)).Type(), "")
	return nil
}

func printTypeInfo(typ reflect.Type, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		tag, aux, _ := strings.Cut(f.Tag.Get("json"), ",")
		typname := f.Type.String()

		// skip private fields
		if (f.PkgPath != "" && !f.Anonymous) || tag == "-" {
			continue
		}

		// embedded structs
		if f.Anonymous {
			t := f.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if t.Kind() == reflect.Struct {
				printTypeInfo(t, prefix+tag)
				continue
			}
		}

		// print nested structs
		if f.Type.Kind() == reflect.Struct && (strings.HasPrefix(typname, "struct") || strings.HasPrefix(typname, "codec.")) {
			printTypeInfo(f.Type, prefix+tag)
			continue
		}

		if aux != "" {
			aux = "(" + aux + ")"
		}

		fmt.Printf("%-30s %s %s\n", prefix+tag, typname, aux)
	}
}

func encode(ctx context.Context, c *rpc.Client, typ, data string) error {
	if !sk.IsValid() {
		return fmt.Errorf("Invalid private key. use -sk or TZGO_PRIVATE_KEY")
	}
	op := codec.NewOp()
	if data[0] == '[' {
		var batch []json.RawMessage
		if err := json.Unmarshal([]byte(data), &batch); err != nil {
			return err
		}
		for _, v := range batch {
			o, err := makeOp(c, typ, string(v))
			if err != nil {
				return err
			}
			op.WithContents(o)
		}
	} else {
		o, err := makeOp(c, typ, data)
		if err != nil {
			return err
		}
		op.WithContents(o)
	}
	hash, err := c.GetBlockHash(ctx, rpc.Head)
	if err != nil {
		return err
	}
	op.WithBranch(hash)
	if err := c.Complete(ctx, op, sk.Public()); err != nil {
		return err
	}
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
	o, err := makeOp(c, typ, data)
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
	if err := c.Validate(ctx, op); err != nil {
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

func sign(ctx context.Context, c *rpc.Client, msg string) error {
	if !sk.IsValid() {
		return fmt.Errorf("Invalid private key. use -sk or TZGO_PRIVATE_KEY")
	}
	buf, err := hex.DecodeString(msg)
	if err != nil {
		return err
	}
	op, err := codec.DecodeOp(buf)
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

func sign_remote(ctx context.Context, c *rpc.Client, addr, msg string) error {
	buf, err := hex.DecodeString(msg)
	if err != nil {
		return err
	}
	op, err := codec.DecodeOp(buf)
	if err != nil {
		return err
	}
	a, err := tezos.ParseAddress(addr)
	if err != nil {
		return err
	}
	rs, err := remote.New("http://localhost:6732", nil)

	sig, err := rs.SignOperation(ctx, a, op)
	if err != nil {
		return err
	}
	op.WithSignature(sig)
	fmt.Println("Signature:", sig.String())
	fmt.Println("Binary:", hex.EncodeToString(sig.Bytes()))
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
	// special treatment for endorsements (wrapper&content branches must match)
	if es, ok := op.Contents[0].(*codec.EndorsementWithSlot); ok {
		head, _ := c.GetBlockHash(ctx, rpc.Head)
		fmt.Println("Setting block hash", head)
		es.Endorsement.Branch = head
		op.WithBranch(head)
	}
	res, err := c.Simulate(ctx, op, nil)
	if err != nil {
		return err
	}
	buf, err = json.MarshalIndent(res.Op, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("Result\n", string(buf))
	buf, err = json.MarshalIndent(res.TotalCosts(), "", "  ")
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
	hash, err := c.Broadcast(ctx, op)
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
	fut := rpc.NewResult(oh)
	if n, err := strconv.ParseInt(conf, 10, 64); err == nil {
		fut.WithConfirmations(n)
	}
	if n, err := strconv.ParseInt(ttl, 10, 64); err == nil {
		fut.WithTTL(n)
	}
	// have rpc client listen for block updates
	c.Listen()
	// subscribe result to block observer
	fut.Listen(c.BlockObserver)
	// wait until confirmed, TTL expires or context is canceled
	fut.WaitContext(ctx)
	// fetch op receipt
	res, err := fut.GetReceipt(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Op included in %s with %d confirmations\n", res.Block, fut.Confirmations())
	buf, err := json.MarshalIndent(res.Op, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("Result\n", string(buf))
	buf, err = json.MarshalIndent(res.TotalCosts(), "", "  ")
	if err != nil {
		return err
	}
	fmt.Println("Cost\n", string(buf))
	return nil
}

func send(ctx context.Context, c *rpc.Client, typ, data string) error {
	if !sk.IsValid() {
		return fmt.Errorf("Invalid private key. use -sk or TZGO_PRIVATE_KEY")
	}
	c.Signer = signer.NewFromKey(sk)
	op := codec.NewOp()
	if data[0] == '[' {
		var batch []json.RawMessage
		if err := json.Unmarshal([]byte(data), &batch); err != nil {
			return err
		}
		for _, v := range batch {
			o, err := makeOp(c, typ, string(v))
			if err != nil {
				return err
			}
			op.WithContents(o)
		}
	} else {
		o, err := makeOp(c, typ, data)
		if err != nil {
			return err
		}
		op.WithContents(o)
	}
	rcpt, err := c.Send(ctx, op, &rpc.CallOptions{
		Confirmations:    2,
		TTL:              10,
		MaxFee:           1_000_000,
		ExtraGasMargin:   0,
		SimulationOffset: 5,
	})
	if err != nil {
		return err
	}
	if !rcpt.IsSuccess() {
		return rcpt.Error()
	} else {
		fmt.Printf("Op included in at %s/chains/main/blocks/%s/operations/%d/%d\n", node, rcpt.Block, rcpt.List, rcpt.Pos)
		buf, err := json.MarshalIndent(rcpt.Op, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println("Result\n", string(buf))
		buf, err = json.MarshalIndent(rcpt.TotalCosts(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println("Cost\n", string(buf))
	}

	return nil
}
