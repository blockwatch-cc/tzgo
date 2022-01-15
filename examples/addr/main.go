// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Address examples
//
// tz1KzpjBnunNJVABHBnzfG4iuLmphitExW2p
// tz3gN8NTLNLJg5KRsUU47NHNVHbdhcFXjjaB
// KT1GyeRktoGPEKsWpchWguyy8FAf3aNHkw2T
//
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"blockwatch.cc/tzgo/tezos"
)

var (
	flags   = flag.NewFlagSet("addr", flag.ContinueOnError)
	verbose bool
)

func init() {
	flags.Usage = func() {}
	flags.BoolVar(&verbose, "v", false, "be verbose")
}

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Println("Address Test")
			flags.PrintDefaults()
			os.Exit(0)
		}
		log.Fatalln("Error:", err)
	}

	if err := run(); err != nil {
		log.Fatalln("Error:", err)
	}
}

func run() error {
	if flags.NArg() < 1 {
		return fmt.Errorf("Address required")
	}

	if flags.NArg() == 2 {
		return blinded()
	}

	var (
		key  tezos.Key
		addr tezos.Address
		hash tezos.Hash
		err  error
	)
	// try decoding as key
	if key, err = tezos.ParseKey(flags.Arg(0)); err != nil {
		// try decoding an address
		if addr, err = tezos.ParseAddress(flags.Arg(0)); err != nil {
			// try decoding as hex
			hv, err := hex.DecodeString(flags.Arg(0))
			if err != nil {
				return err
			}
			addr.Hash = hv
			addr.Type = 1
			key.Data = hv
			key.Type = 0
		}
	} else {
		addr = key.Address()
		hash, err = tezos.ParseHash(flags.Arg(0))
		if err != nil {
			return err
		}
	}

	if key.IsValid() {
		fmt.Printf("Key     %s\n", key.String())
		fmt.Printf("ReKey   %s\n", tezos.NewKey(addr.Type.KeyType(), key.Data))
		fmt.Printf("KeyData %x\n", key.Data)
		fmt.Printf("AsHash  %x\n", hash.Hash)
		fmt.Printf("HType   %s\n", hash.Type)
		addr = key.Address()
	}
	fmt.Printf("Address %s\n", addr.String())
	fmt.Printf("Short   %s\n", addr.Short())
	fmt.Printf("PkType  %s\n", addr.Type)
	fmt.Printf("PkHash  %x\n", addr.Hash)
	return nil
}

func blinded() error {
	// Example
	// "pkh": "tz1T1rRqmAk4XtGadNJuNpq8dUdWqLv2Gtq4",
	// "secret": "06da1e038224114366831e47aee7f128f4675311",

	// try decoding an address
	addr, err := tezos.ParseAddress(flags.Arg(0))
	if err != nil {
		return err
	}
	// try blinding with secret
	secret, err := hex.DecodeString(flags.Arg(1))
	if err != nil {
		return err
	}
	blind, err := tezos.BlindAddress(addr, secret)
	if err != nil {
		return err
	}
	fmt.Printf("Address %s\n", addr.String())
	fmt.Printf("Hash    %x\n", addr.Hash)
	fmt.Printf("Blinded %s\n", blind.String())
	return nil
}
