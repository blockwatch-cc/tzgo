// Copyright (c) 2023 Blockwatch Data Inc.
// Authors
// - jean.schmitt@ubisoft.com
// - abdul@blockwatch.cc

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"blockwatch.cc/tzgo/internal/generate"
	"blockwatch.cc/tzgo/internal/parse"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	errExit = errors.New("exit")

	endpointFlag  string
	addressFlag   string
	srcFlag       string
	nameFlag      string
	pkgFlag       string
	outFlag       string
	fixupFileFlag string
)

func init() {
	flag.StringVar(&endpointFlag, "endpoint", "https://rpc.tzstats.com", "rpc endpoint")
	flag.StringVar(&addressFlag, "address", "", "address of the contract. required if -src is not set")
	flag.StringVar(&srcFlag, "src", "", "json file containing the contracts's script")
	flag.StringVar(&nameFlag, "name", "", "name of the contract")
	flag.StringVar(&pkgFlag, "pkg", "", "package name of the output go code")
	flag.StringVar(&outFlag, "out", "", "output file. Prints to Stdout if not set")
	flag.StringVar(&fixupFileFlag, "fixup", "", "yaml file to fix generated go code for automatically generated functions/variable names")
}

func parseFlags() error {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "version":
			printVersion()
			return errExit
		case "help":
			fmt.Printf("Usage: %s [flags]\n", appName)
			fmt.Println("\nFlags")
			flag.PrintDefaults()
		}
	}
	flag.Parse()
	return nil
}

func runCommand() error {
	if pkgFlag == "" {
		return errors.New("-pkg is required, to get package name")
	}
	if nameFlag == "" {
		return errors.New("-name is required to set name of contract")
	}
	src, err := getSrc()
	if err != nil {
		return errors.Wrap(err, "failed to get contract script")
	}
	generated, err := generateBindings(src)
	if err != nil {
		return errors.Wrap(err, "failed to generate bindings")
	}
	err = writeResult(generated)
	if err != nil {
		return errors.Wrap(err, "failed to write generated code to file")
	}
	return nil
}

func generateBindings(script []byte) ([]byte, error) {
	var err error
	data := generate.Data{
		Address: addressFlag,
		Package: pkgFlag,
	}
	data.Contract, data.Structs, err = parse.Parse(script, nameFlag)
	if err != nil {
		return nil, err
	}
	if fixupFileFlag != "" {
		fixupFile, err := os.ReadFile(fixupFileFlag)
		if err != nil {
			return nil, err
		}

		var fixupCfg parse.FixupConfig
		err = yaml.NewDecoder(bytes.NewReader(fixupFile)).Decode(&fixupCfg)
		if err != nil {
			return nil, err
		}

		data.Structs = parse.Fixup(fixupCfg, data.Structs, strcase.ToCamel)
	}

	return generate.Render(&data)
}

func getSrc() ([]byte, error) {
	if srcFlag != "" {
		return os.ReadFile(srcFlag)
	}

	// Get source from RPC
	// At this point, addressFlag is required
	if addressFlag == "" {
		return nil, errors.New("-address is required when getting script from rpc")
	}

	u, err := url.JoinPath(endpointFlag, "chains/main/blocks/head/context/contracts", addressFlag, "script")
	if err != nil {
		return nil, err
	}
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get contract script at url %s: %v", u, res.Status)
	}
	return io.ReadAll(res.Body)
}

func writeResult(out []byte) error {
	if outFlag == "" {
		_, err := os.Stdout.Write(out)
		if err != nil {
			return err
		}
		return nil
	}
	return os.WriteFile(outFlag, out, 0o644)
}
