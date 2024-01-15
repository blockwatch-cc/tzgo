// ☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️
//
// ATTENTION!! THIS PROGRAM WILL BURN YOUR FUNDS!! HANDLE WITH CARE
//
// ☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️☠️
//
// Generates fake double_endorsement_evidence for the given baker
// which triggers slashing code in the protocol.
//
// How it works
// - watch for next block, read level, round, payload hash
// - check if baker has endorsement rights for this block, extract slot
// - create and sign 2 endorsement ops from real + random payload hashes
// - wait one block
// - send the 2 signed endorsements in as double_endorsement_evidence
//
// To specify the baker key set env var TZGO_PRIVATE_KEY

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/signer"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
	"golang.org/x/crypto/blake2b"
)

var (
	flags   = flag.NewFlagSet("doublex", flag.ContinueOnError)
	verbose bool
	node    string
	sk      tezos.PrivateKey
)

func init() {
	if k := os.Getenv("TZGO_PRIVATE_KEY"); k != "" {
		sk = tezos.MustParsePrivateKey(k)
	}
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
	flags.StringVar(&node, "node", "https://rpc.oxford.tzstats.com", "Tezos node URL")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Usage: doublex [args]")
			fmt.Println("\nArguments")
			flags.PrintDefaults()
			os.Exit(0)
		}
		fmt.Println("Error:", err)
		return
	}

	switch {
	case verbose:
		log.SetLevel(log.LevelTrace)
	default:
		log.SetLevel(log.LevelInfo)
	}
	rpc.UseLogger(log.Log)

	if err := run(); err != nil {
		if !errors.Is(err, context.Canceled) {
			fmt.Println("Error:", err)
		}
	}
}

func run() error {
	if !sk.IsValid() {
		return fmt.Errorf("missing secret key")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := rpc.NewClient(node, nil)
	if err != nil {
		return err
	}

	log.Info("Using key for baker ", sk.Address())
	c.Signer = signer.NewFromKey(sk)

	if err := c.Init(ctx); err != nil {
		return err
	}
	return monitorBlocks(ctx, c)
}

// watch for next block with endorsing rights
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
			log.Info("Stopping monitor")
			cancel()
		case <-ctx.Done():
		}
	}()

	log.Info("Waiting for new blocks... (cancel with Ctrl-C)")
	for {
		h, err := mon.Recv(ctx2)
		if err != nil {
			return err
		}
		select {
		case <-ctx2.Done():
			return nil
		default:
		}
		if err := handleBlock(ctx2, c, h); err != nil {
			log.Error(err)
		}
	}
}

var evidence *codec.TenderbakeDoubleEndorsementEvidence

func handleBlock(ctx context.Context, c *rpc.Client, b *rpc.BlockHeaderLogEntry) error {
	// send evidence in next block
	if evidence != nil {
		log.Infof("Block %d %s round=%d", b.Level, b.Hash, b.Round())
		ev := evidence
		evidence = nil
		return sendDoubleEndorse(ctx, c, b, ev)
	}

	// prepare evidence
	slot, ok, err := fetchEndorsingRights(ctx, c, b.Hash)
	if err != nil {
		return err
	}
	if !ok {
		log.Infof("Block %d %s round=%d", b.Level, b.Hash, b.Round())
		return nil
	}
	log.Infof("Block %d %s round=%d with endorsing rights", b.Level, b.Hash, b.Round())

	evidence = createDoubleEndorse(c, b, slot)
	return nil
}

func fetchEndorsingRights(ctx context.Context, c *rpc.Client, id tezos.BlockHash) (int, bool, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/endorsing_rights?delegate=%s", id, sk.Address())
	var rights []struct {
		Level         int64                `json:"level"`
		Delegates     []rpc.EndorsingRight `json:"delegates"`
		EstimatedTime time.Time            `json:"estimated_time"`
	}
	if err := c.Get(ctx, u, &rights); err != nil {
		return 0, false, err
	}
	if len(rights) == 0 {
		return 0, false, nil
	}
	return rights[0].Delegates[0].FirstSlot, true, nil
}

func createDoubleEndorse(c *rpc.Client, b *rpc.BlockHeaderLogEntry, slot int) *codec.TenderbakeDoubleEndorsementEvidence {
	log.Infof("Creating 2endorse evidence")
	o1, oh1 := signEndorsement(c, b, slot, false)
	o2, oh2 := signEndorsement(c, b, slot, true)
	// FIXME: order endorsements by op hash
	if bytes.Compare(oh1[:], oh2[:]) > 0 {
		o1, o2 = o2, o1
	}
	return &codec.TenderbakeDoubleEndorsementEvidence{
		Op1: o1,
		Op2: o2,
	}
}

func sendDoubleEndorse(ctx context.Context, c *rpc.Client, b *rpc.BlockHeaderLogEntry, ev *codec.TenderbakeDoubleEndorsementEvidence) error {
	op := codec.NewOp().
		WithParams(c.Params). // use protocol params for correct op tags
		WithContents(ev).
		WithBranch(b.Hash)
	buf, _ := op.MarshalJSON()
	log.Infof("Sending 2endorse: %s", string(buf))
	res, err := c.Send(ctx, op, &rpc.CallOptions{IgnoreLimits: true})
	if err != nil {
		return err
	}
	if !res.IsSuccess() {
		return res.Error()
	}
	log.Infof("Success")
	return nil
}

func signEndorsement(c *rpc.Client, b *rpc.BlockHeaderLogEntry, slot int, random bool) (codec.TenderbakeInlinedEndorsement, tezos.OpHash) {
	e := codec.TenderbakeEndorsement{
		Slot:             int16(slot),
		Level:            int32(b.Level),
		Round:            int32(b.Round()),
		BlockPayloadHash: b.PayloadHash(),
	}
	if random {
		_, _ = rand.Read(e.BlockPayloadHash[:])
	}
	op := codec.NewOp().
		WithParams(c.Params).   // use protocol params for correct op tags
		WithChainId(c.ChainId). // Tenderbake endorsements need chain id
		WithContents(&e).
		WithBranch(b.Hash)
	if err := op.Sign(sk); err != nil {
		log.Errorf("signing endorsement: %v", err)
	}
	return codec.TenderbakeInlinedEndorsement{
		Branch:      b.Hash,
		Endorsement: e,
		Signature:   op.Signature,
	}, ophash(op.Digest())
}

// FIXME: what's the correct method to calculate op hash from contents?
func ophash(buf []byte) (oh tezos.OpHash) {
	h, _ := blake2b.New(32, nil)
	h.Write(buf)
	copy(oh[:], h.Sum(nil))
	return
}
