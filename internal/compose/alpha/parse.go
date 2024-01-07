// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package alpha

import (
	"encoding"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"blockwatch.cc/tzgo/contract/bind"
	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
	"github.com/pkg/errors"
)

func ParseScript(ctx compose.Context, task Task) (*micheline.Script, error) {
	var (
		script *micheline.Script
		err    error
	)
	if task.Script.ValueSource.IsUsed() {
		script, err = loadSource[micheline.Script](ctx, task.Script.ValueSource)
		if err != nil {
			return nil, errors.Wrap(err, "loading script")
		}
	} else {
		var (
			code  *micheline.Code
			store *micheline.Prim
		)
		if task.Script.Code == nil || !task.Script.Code.ValueSource.IsUsed() {
			return nil, fmt.Errorf("missing script code source")
		}
		code, err = loadSource[micheline.Code](ctx, task.Script.Code.ValueSource)
		if err != nil {
			return nil, errors.Wrap(err, "loading script code")
		}
		if task.Script.Storage == nil || (!task.Script.Storage.ValueSource.IsUsed() && task.Script.Storage.Args == nil) {
			return nil, fmt.Errorf("missing initial storage source")
		}
		if task.Script.Storage.Args != nil {
			// processed below
			store = &micheline.Prim{}
		} else {
			store, err = loadSource[micheline.Prim](ctx, task.Script.Storage.ValueSource)
			if err != nil {
				return nil, errors.Wrap(err, "loading initial storage")
			}
		}
		script = &micheline.Script{
			Code:    *code,
			Storage: *store,
		}
	}
	if !script.IsValid() {
		return nil, fmt.Errorf("invalid script")
	}

	// Note: code patching is not supported

	// patch storage
	if task.Script.Storage != nil {
		// ctx.Log.Infof("Storage Type: %s", script.Code.Storage.Dump())
		// ctx.Log.Infof("Original Storage: %s", script.Storage.Dump())
		for _, p := range task.Script.Storage.Patch {
			if p.Path == nil {
				ctx.Log.Debugf("Resolving storage path for %q...", *p.Key)
				idx, ok := script.StorageType().LabelIndex(*p.Key)
				if !ok {
					return nil, fmt.Errorf("storage key %q not found", *p.Key)
				}
				s := idxToPath(idx)
				p.Path = &s
			}
			ctx.Log.Debugf("Patching script storage...")
			err = patch(ctx, p, func(path string, oc micheline.OpCode, prim micheline.Prim) error {
				ctx.Log.Debugf("> path=%s type=%s val=%s", path, oc, prim.Dump())
				return script.Storage.SetPath(path, prim)
			})
			if err != nil {
				return nil, errors.Wrap(err, "patching storage")
			}
		}
		if task.Script.Storage.Args != nil {
			// replace variables in args
			if _, _, err := resolveArgs(ctx, task.Script.Storage.Args); err != nil {
				return nil, err
			}
			// map args to storage spec
			// ctx.Log.Infof("Args: %#v", task.Script.Storage.Args)
			// ctx.Log.Infof("Typedef: %s", script.StorageType().Typedef(""))
			script.Storage, err = script.StorageType().Typedef("").Marshal(task.Script.Storage.Args, false)
			if err != nil {
				return nil, err
			}
			// ctx.Log.Infof("Gen: %s", script.Storage.Dump())
			// ctx.Log.Infof("Gen: %#v", script.Storage)
			// buf, err := script.Storage.MarshalBinary()
			// if err != nil {
			// 	return nil, err
			// }
			// ctx.Log.Infof("Gen: %s", tezos.HexBytes(buf))
		}
	}
	if !script.Storage.IsValid() {
		return nil, fmt.Errorf("invalid storage %s", script.Storage.Dump())
	}

	return script, nil
}

func ParseParams(ctx compose.Context, task Task) (*micheline.Prim, error) {
	var (
		params *micheline.Prim
		err    error
	)
	if task.Params.ValueSource.IsUsed() {
		params, err = loadSource[micheline.Prim](ctx, task.Params.ValueSource)
		if err != nil {
			return nil, err
		}
	}
	for _, p := range task.Params.Patch {
		if p.Path == nil {
			ctx.Log.Debugf("Resolving destination script for %s...", task.Destination)
			// load script
			addr, err := tezos.ParseAddress(task.Destination)
			if err != nil {
				return nil, err
			}
			script, err := ctx.ResolveScript(addr)
			if err != nil {
				return nil, err
			}
			ctx.Log.Debugf("Resolving storage path for %q...", *p.Key)
			idx, ok := script.StorageType().LabelIndex(*p.Key)
			if !ok {
				return nil, fmt.Errorf("storage key %q not found", *p.Key)
			}
			s := idxToPath(idx)
			p.Path = &s
		}

		ctx.Log.Debugf("Patching params...")
		err = patch(ctx, p, func(path string, oc micheline.OpCode, prim micheline.Prim) error {
			ctx.Log.Debugf("> path=%s type=%s val=%s", path, oc, prim.Dump())
			return params.SetPath(path, prim)
		})
		if err != nil {
			return nil, errors.Wrap(err, "patching params")
		}
	}
	if task.Params.Args != nil {
		// replace variables in args
		if _, _, err := resolveArgs(ctx, task.Params.Args); err != nil {
			return nil, err
		}
		addr, err := ctx.ResolveAddress(task.Destination)
		if err != nil {
			return nil, err
		}
		// load script
		ctx.Log.Debugf("Resolving destination script for %s...", addr)
		script, err := ctx.ResolveScript(addr)
		if err != nil {
			return nil, err
		}
		// map args to entrypoint spec
		eps, err := script.Entrypoints(true)
		if err != nil {
			return nil, err
		}
		ep, ok := eps[task.Params.Entrypoint]
		if !ok {
			return nil, fmt.Errorf("entrypoint %q not found, have %#v", task.Params.Entrypoint, eps)
		}
		// marshal to prim tree
		// Note: be mindful of the way entrypoint typedefs are structured:
		// - 1 arg: use scalar value in ep.Typedef[0]
		// - >1 arg: use entire list in ep.Typedef but wrap into struct
		ctx.Log.Debugf("Marshaling params...")
		// ctx.Log.Debugf("EP prim=%s", ep.Prim.Dump())
		// ctx.Log.Debugf("EP type=%s", ep.Typedef)
		typ := ep.Typedef[0]
		if len(ep.Typedef) > 1 {
			typ = micheline.Typedef{
				Name: micheline.CONST_ENTRYPOINT,
				Type: micheline.TypeStruct,
				Args: ep.Typedef,
			}
		}
		// ctx.Log.Debugf("USE type=%s", typ)
		prim, err := typ.Marshal(task.Params.Args, true)
		if err != nil {
			return nil, errors.Wrap(err, "marshal params")
		}
		ctx.Log.Debugf("Result %s", prim.Dump())
		params = &prim
	}
	return params, nil
}

func loadSource[T any](ctx compose.Context, src ValueSource) (val *T, err error) {
	switch {
	case src.Value != "":
		val = new(T)
		if !isHex(src.Value) {
			err = json.Unmarshal([]byte(src.Value), val)
		} else {
			var buf []byte
			buf, err = hex.DecodeString(src.Value)
			if err == nil {
				x := any(val)
				if u, ok := x.(encoding.BinaryUnmarshaler); ok {
					err = u.UnmarshalBinary(buf)
				} else {
					err = fmt.Errorf("type %T does not implement encoding.BinaryUnmarshaler", val)
				}
			}
		}
	case src.File != "":
		fname := filepath.Join(ctx.Filepath(), src.File)
		val, err = compose.ReadJsonFile[T](fname)
	case src.Url != "":
		val, err = compose.Fetch[T](ctx, src.Url)
	default:
		err = fmt.Errorf("invalid source")
	}
	return
}

func patch(ctx compose.Context, p Patch, updater func(string, micheline.OpCode, micheline.Prim) error) error {
	opCode, err := micheline.ParseOpCode(p.Type)
	if err != nil {
		return err
	}
	if !opCode.IsTypeCode() {
		return fmt.Errorf("%s is not a valid type", p.Type)
	}
	str, err := ctx.ResolveString(*p.Value)
	if err != nil {
		return err
	}
	val, err := compose.ParseValue(opCode, str)
	if err != nil {
		return err
	}
	prim, err := bind.MarshalPrim(val, p.Optimized)
	if err != nil {
		return err
	}
	return updater(*p.Path, opCode, prim)
}

func idxToPath(idx []int) string {
	if len(idx) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(strconv.Itoa(idx[0]))
	for _, v := range idx[1:] {
		b.WriteByte('/')
		b.WriteString(strconv.Itoa(v))
	}
	return b.String()
}

func resolveArgs(ctx compose.Context, args any) (string, bool, error) {
	if args == nil {
		return "", false, nil
	}
	switch val := args.(type) {
	case map[string]any:
		for n, v := range val {
			s, ok, err := resolveArgs(ctx, v)
			if err != nil {
				return "", false, fmt.Errorf("arg %s: %v", n, err)
			}
			// try convert key (may be composite key)
			if newKey, ok := ctx.ResolveNestedVars(n); ok {
				delete(val, n)
				n = newKey
				val[n] = v
			}
			if ok {
				val[n] = s
			}
		}
	case []any:
		for i, v := range val {
			s, ok, err := resolveArgs(ctx, v)
			if err != nil {
				return "", false, fmt.Errorf("arg %d: %v", i, err)
			}
			if ok {
				val[i] = s
			}
		}
	case string:
		if s, err := ctx.ResolveString(val); err != nil {
			return "", false, err
		} else {
			return s, true, nil
		}
	case int:
		// skip
	case bool:
		// skip
	default:
		return "", false, fmt.Errorf("unsupported arg type %T", args)
	}
	return "", false, nil
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// func isJson(s string) bool {
// 	if len(s) == 0 {
// 		return true
// 	}
// 	switch s[0] {
// 	case '{', '[', '"', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
// 		return true
// 	default:
// 		return false
// 	}
// }
