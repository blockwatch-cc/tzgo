// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

type CloneMode byte

const (
	CloneModeFile CloneMode = iota
	CloneModeJson
	CloneModeBinary
	CloneModeUrl
	CloneModeArgs
)

func (m CloneMode) String() string {
	switch m {
	case CloneModeFile:
		return "file"
	case CloneModeJson:
		return "json"
	case CloneModeBinary:
		return "binary"
	case CloneModeUrl:
		return "url"
	case CloneModeArgs:
		return "args"
	default:
		return ""
	}
}

func (m *CloneMode) Set(s string) (err error) {
	switch s {
	case "file":
		*m = CloneModeFile
	case "json":
		*m = CloneModeJson
	case "bin":
		*m = CloneModeBinary
	case "url":
		*m = CloneModeUrl
	case "args":
		*m = CloneModeArgs
	default:
		err = fmt.Errorf("invalid clone mode")
	}
	return
}

type CloneConfig struct {
	Name     string
	Contract tezos.Address
	IndexUrl string
	NumOps   uint
	Path     string
	Mode     CloneMode
}

type Op struct {
	Type       string            `json:"type"`
	Hash       string            `json:"hash"`
	Height     int               `json:"height"`
	OpP        int               `json:"op_p"`
	OpC        int               `json:"op_c"`
	OpI        int               `json:"op_i"`
	IsInternal bool              `json:"is_internal"`
	Sender     string            `json:"sender"`
	Receiver   string            `json:"receiver"`
	Amount     float64           `json:"volume"`
	Script     *micheline.Script `json:"script"`
	Params     *struct {
		Entrypoint string         `json:"entrypoint"`
		Prim       micheline.Prim `json:"prim"`
	} `json:"parameters"`

	// processed
	Url           string `json:"-"`
	PackedCode    string `json:"-"`
	PackedStorage string `json:"-"`
	PackedParams  string `json:"-"`
	Args          any    `json:"-"`
}

func Clone(ctx Context, version string, cfg CloneConfig) error {
	if !HasVersion(version) {
		return ErrInvalidVersion
	}
	if !cfg.Contract.IsContract() {
		return fmt.Errorf("invalid contract address")
	}
	if cfg.Name == "" {
		cfg.Name = cfg.Contract.String()
	}
	ops, err := fetchOps(ctx, cfg)
	if err != nil {
		return err
	}
	eng := New(version)
	buf, err := eng.Clone(ctx, ops, cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(cfg.Path, buf, 0644)
	if err != nil {
		return err
	}
	ctx.Log.Infof("File %s written successfully.", cfg.Path)
	return nil
}

func fetchOps(ctx Context, cfg CloneConfig) ([]Op, error) {
	ctx.Log.Infof("Fetching contract operations...")
	u := fmt.Sprintf("%s/explorer/account/%s/operations?prim=1&storage=1&order=asc&limit=%d",
		cfg.IndexUrl, cfg.Contract, cfg.NumOps+1)
	resp, err := Fetch[[]Op](ctx, u)
	if err != nil {
		return nil, err
	}
	ops := *resp
	if len(ops) == 0 {
		return nil, fmt.Errorf("contract %q has no transactions", cfg.Contract)
	}
	switch cfg.Mode {
	case CloneModeFile:
		err = storeOps(ctx, ops, cfg)
	case CloneModeJson:
		err = encodeJson(ops)
	case CloneModeBinary:
		err = encodeBinary(ops)
	case CloneModeUrl:
		encodeUrl(ops, ctx.url)
	}
	if err != nil {
		return nil, err
	}
	if err := encodeArgs(ops); err != nil {
		ctx.Log.Warnf("Marshaling args: %v", err)
	}
	return ops, nil
}

func storeOps(ctx Context, ops []Op, cfg CloneConfig) error {
	for i := range ops {
		var (
			buf []byte
			err error
		)
		ops[i].Url = generateFilename(cfg, ops[i], i)
		switch ops[i].Type {
		case "origination":
			buf, err = json.Marshal(ops[i].Script)
		case "transaction":
			if ops[i].Params != nil {
				buf, err = json.Marshal(ops[i].Params.Prim)
			}
		}
		if err != nil {
			return err
		}
		err = os.WriteFile(ops[i].Url, buf, 0644)
		if err != nil {
			return err
		}
		ctx.Log.Infof("File %s written.", ops[i].Url)
	}
	return nil
}

func encodeJson(ops []Op) error {
	for i := range ops {
		var (
			buf []byte
			err error
		)
		switch ops[i].Type {
		case "origination":
			buf, err = json.Marshal(ops[i].Script.Code)
			if err == nil {
				ops[i].PackedCode = string(buf)
				buf, err = json.Marshal(ops[i].Script.Storage)
			}
			if err == nil {
				ops[i].PackedStorage = string(buf)
			}
		case "transaction":
			if ops[i].Params != nil {
				buf, err = json.Marshal(ops[i].Params.Prim)
				ops[i].PackedParams = string(buf)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeBinary(ops []Op) error {
	for i := range ops {
		var (
			buf []byte
			err error
		)
		switch ops[i].Type {
		case "origination":
			buf, err = ops[i].Script.Code.MarshalBinary()
			if err == nil {
				ops[i].PackedCode = hex.EncodeToString(buf)
				buf, err = ops[i].Script.Storage.MarshalBinary()
			}
			if err == nil {
				ops[i].PackedStorage = hex.EncodeToString(buf)
			}
		case "transaction":
			if ops[i].Params != nil {
				buf, err = ops[i].Params.Prim.MarshalBinary()
				ops[i].PackedParams = hex.EncodeToString(buf)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeUrl(ops []Op, host string) {
	for i, op := range ops {
		switch op.Type {
		case "origination":
			if ops[i].IsInternal {
				ops[i].Url = fmt.Sprintf(
					"%s/chains/main/blocks/%d/operations/3/%d#contents.%d.metadata.internal_operation_results.%d.script",
					host,
					op.Height,
					op.OpP,
					op.OpC,
					op.OpI,
				)
			} else {
				ops[i].Url = fmt.Sprintf(
					"%s/chains/main/blocks/%d/operations/3/%d#contents.%d.script",
					host,
					op.Height,
					op.OpP,
					op.OpC,
				)
			}
		case "transaction":
			if ops[i].IsInternal {
				ops[i].Url = fmt.Sprintf(
					"%s/chains/main/blocks/%d/operations/3/%d#contents.%d.metadata.internal_operation_results.%d.parameters.value",
					host,
					op.Height,
					op.OpP,
					op.OpC,
					op.OpI,
				)
			} else {
				ops[i].Url = fmt.Sprintf(
					"%s/chains/main/blocks/%d/operations/3/%d#contents.%d.parameters.value",
					host,
					op.Height,
					op.OpP,
					op.OpC,
				)
			}
		}
	}
}

func encodeArgs(ops []Op) error {
	var script *micheline.Script
	for i, op := range ops {
		switch op.Type {
		case "origination":
			script = op.Script
			val := micheline.NewValue(script.StorageType(), script.Storage).
				UnpackAllAsciiStrings()
			res, err := val.Map()
			if err != nil {
				return err
			}
			ops[i].Args = res

		case "transaction":
			eps, err := script.Entrypoints(true)
			if err != nil {
				return err
			}
			ep, ok := eps[op.Params.Entrypoint]
			if !ok {
				return fmt.Errorf("missing entrypoint %s", op.Params.Entrypoint)
			}
			val := micheline.NewValue(ep.Type(), op.Params.Prim).
				UnpackAllAsciiStrings()
			res, err := val.Map()
			if err != nil {
				return err
			}
			ops[i].Args = res.(map[string]any)[op.Params.Entrypoint]
		}
	}
	return nil
}

func generateFilename(cfg CloneConfig, op Op, index int) string {
	return fmt.Sprintf("%02d-%s-%s.json", index, cfg.Name, op.Type)
}
