// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

type PrimUnmarshaler interface {
	UnmarshalPrim(Prim) error
}

// FindLabel searches a nested type annotation path. Must be used on a type prim.
// Path segments are separated by dot (.)
func (p Prim) FindLabel(label string) (Prim, bool) {
	idx, ok := p.LabelIndex(label)
	if !ok {
		return InvalidPrim, false
	}
	prim, _ := p.GetIndex(idx)
	return prim, true
}

// LabelIndex returns the indexed path to a type annotation label and true
// if path exists. Path segments are separated by dot (.)
func (p Prim) LabelIndex(label string) ([]int, bool) {
	return p.findLabelPath(strings.Split(label, "."), nil)
}

func (p Prim) findLabelPath(path []string, idx []int) ([]int, bool) {
	prim := p
next:
	for {
		if len(path) == 0 {
			return idx, true
		}
		var found bool
		for i, v := range prim.Args {
			if v.HasAnno() && v.MatchesAnno(path[0]) {
				idx = append(idx, i)
				path = path[1:]
				prim = v
				found = true
				continue next
			}
		}

		for i, v := range prim.Args {
			if v.HasAnno() {
				continue
			}
			idx2, ok := v.findLabelPath(path, append(idx, i))
			if ok {
				path = path[:0]
				found = true
				idx = idx2
				continue next
			}
		}
		if !found {
			return nil, false
		}
	}
}

// GetPath returns a nested primitive at path. Path segments are separated by slash (/).
// Works on both type and value primitive trees.
func (p Prim) GetPath(path string) (Prim, error) {
	index, err := p.getIndex(path)
	if err != nil {
		return InvalidPrim, err
	}
	return p.GetIndex(index)
}

// GetPathExt returns a nested primitive at path if the primitive matches
// the expected opcode. Path segments are separated by slash (/).
// Works on both type and value primitive trees.
func (p Prim) GetPathExt(path string, typ OpCode) (Prim, error) {
	prim, err := p.GetPath(path)
	if err != nil {
		return prim, err
	}
	if prim.OpCode != typ {
		return InvalidPrim, fmt.Errorf("micheline: unexpected type %s at path %v", prim.OpCode, path)
	}
	return prim, nil
}

func (p Prim) getIndex(path string) ([]int, error) {
	index := make([]int, 0)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	if len(path) == 0 {
		return nil, nil
	}
	for i, v := range strings.Split(path, "/") {
		switch v {
		case "L", "l", "0":
			index = append(index, 0)
		case "R", "r", "1":
			index = append(index, 1)
		default:
			idx, err := strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("micheline: invalid path component '%v' at pos %d", v, i)
			}
			index = append(index, idx)
		}
	}
	return index, nil
}

// HasIndex returns true when a nested primitive exists at path defined by index.
func (p Prim) HasIndex(index []int) bool {
	prim := p
	for _, v := range index {
		if v < 0 || len(prim.Args) <= v {
			return false
		}
		prim = prim.Args[v]
	}
	return true
}

// GetIndex returns a nested primitive at path index.
func (p Prim) GetIndex(index []int) (Prim, error) {
	prim := p
	for _, v := range index {
		if v < 0 || len(prim.Args) <= v {
			return InvalidPrim, fmt.Errorf("micheline: index %d out of bounds", v)
		}
		prim = prim.Args[v]
	}
	return prim, nil
}

// GetIndex returns a nested primitive at path index if the primitive matches the
// expected opcode. This only works on type trees. Value trees lack opcode info.
func (p Prim) GetIndexExt(index []int, typ OpCode) (Prim, error) {
	prim, err := p.GetIndex(index)
	if err != nil {
		return InvalidPrim, err
	}
	if prim.OpCode != typ {
		return InvalidPrim, fmt.Errorf("micheline: unexpected type %s at path %v", prim.OpCode, index)
	}
	return prim, nil
}

// Decode unmarshals a prim tree into a Go struct. The mapping uses Go struct tags
// to define primitive paths that are mapped to each struct member. Types are
// converted between Micheline and Go when possible.
//
// Examples of struct field tags and their meanings:
//
//	// maps Micheline path 0/0/0 to string field and fails on type mismatch
//	Field string `prim:",path=0/0/1"`
//
//	// ignore type errors and do not update struct field
//	Field string  `prim:",path=0/0/1,nofail"`
//
//	// ignore struct field
//	Field string  `prim:"-"`
func (p Prim) Decode(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("micheline: non-pointer passed to Decode: %s %s", val.Kind(), val.Type().String())
	}
	val = reflect.Indirect(val)
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("micheline: non-struct passed to Decode %s %s", val.Kind(), val.Type().String())
	}
	return p.unmarshal(val)
}

func (p Prim) unmarshal(val reflect.Value) error {
	val = derefValue(val)
	if val.CanInterface() && val.Type().Implements(primUnmarshalerType) {
		// This is an unmarshaler with a non-pointer receiver,
		// so it's likely to be incorrect, but we do what we're told.
		return val.Interface().(PrimUnmarshaler).UnmarshalPrim(p)
	}
	if val.CanAddr() {
		pv := val.Addr()
		if pv.CanInterface() && pv.Type().Implements(primUnmarshalerType) {
			return pv.Interface().(PrimUnmarshaler).UnmarshalPrim(p)
		}
	}

	tinfo, err := getTypeInfo(indirectType(val.Type()))
	if err != nil {
		return err
	}
	for _, finfo := range tinfo.fields {
		dst := finfo.value(val)
		if !dst.IsValid() {
			continue
		}
		if dst.Kind() == reflect.Ptr {
			if dst.IsNil() && dst.CanSet() {
				dst.Set(reflect.New(dst.Type()))
			}
			dst = dst.Elem()
		}
		pp, err := p.GetIndex(finfo.path)
		if err != nil {
			if finfo.nofail {
				continue
			}
			return err
		}
		switch finfo.typ {
		case T_BYTES:
			if dst.CanAddr() {
				pv := dst.Addr()
				if pv.CanInterface() {
					if pv.Type().Implements(binaryUnmarshalerType) {
						if err := pv.Interface().(encoding.BinaryUnmarshaler).UnmarshalBinary(pp.Bytes); err != nil {
							if !finfo.nofail {
								return err
							}
						}
						break
					}
					if pv.Type().Implements(textUnmarshalerType) {
						if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText(pp.Bytes); err != nil {
							if !finfo.nofail {
								return err
							}
						}
						break
					}
				}
			}
			buf := make([]byte, len(pp.Bytes))
			copy(buf, pp.Bytes)
			dst.SetBytes(buf)
		case T_STRING:
			if dst.CanAddr() {
				pv := dst.Addr()
				if pv.CanInterface() && pv.Type().Implements(textUnmarshalerType) {
					if pp.Bytes != nil {
						if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText(pp.Bytes); err != nil {
							if !finfo.nofail {
								return err
							}
						}
						break
					}
					if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(pp.String)); err != nil {
						if !finfo.nofail {
							return err
						}
					}
					break
				}
			}
			dst.SetString(pp.String)
		case T_INT, T_NAT:
			if dst.CanAddr() {
				pv := dst.Addr()
				if pv.CanInterface() && pv.Type().Implements(textUnmarshalerType) {
					if pp.Int != nil {
						if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(pp.Int.Text(10))); err != nil {
							if !finfo.nofail {
								return err
							}
						}
						break
					}
					if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(pp.String)); err != nil {
						if !finfo.nofail {
							return err
						}
					}
					break
				}
			}
			switch dst.Type().Kind() {
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				dst.SetUint(uint64(pp.Int.Int64()))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				dst.SetInt(pp.Int.Int64())
			}

		case T_BOOL:
			dst.SetBool(pp.OpCode == D_TRUE)
		case T_TIMESTAMP:
			if pp.Int != nil {
				dst.Set(reflect.ValueOf(time.Unix(pp.Int.Int64(), 0).UTC()))
			} else {
				tm, err := time.Parse(time.RFC3339, pp.String)
				if err != nil {
					if !finfo.nofail {
						return err
					}
				}
				dst.Set(reflect.ValueOf(tm))
			}
		case T_ADDRESS:
			var (
				addr tezos.Address
				err  error
			)
			if pp.Bytes != nil {
				err = addr.Decode(pp.Bytes)
			} else {
				err = addr.UnmarshalText([]byte(pp.String))
			}
			if err != nil && !finfo.nofail {
				return err
			}
			dst.Set(reflect.ValueOf(addr))
		case T_KEY:
			var (
				key tezos.Key
				err error
			)
			if pp.Bytes != nil {
				err = key.UnmarshalBinary(pp.Bytes)
			} else {
				err = key.UnmarshalText([]byte(pp.String))
			}
			if err != nil && !finfo.nofail {
				return err
			}
			dst.Set(reflect.ValueOf(key))
		case T_SIGNATURE:
			var (
				sig tezos.Signature
				err error
			)
			if pp.Bytes != nil {
				err = sig.UnmarshalBinary(pp.Bytes)
			} else {
				err = sig.UnmarshalText([]byte(pp.String))
			}
			if err != nil && !finfo.nofail {
				return err
			}
			dst.Set(reflect.ValueOf(sig))
		case T_CHAIN_ID:
			var (
				chain tezos.ChainIdHash
				err   error
			)
			if pp.Bytes != nil {
				err = chain.UnmarshalBinary(pp.Bytes)
			} else {
				err = chain.UnmarshalText([]byte(pp.String))
			}
			if err != nil && !finfo.nofail {
				return err
			}
			dst.Set(reflect.ValueOf(chain))
		case T_LIST:
			styp := dst.Type()
			if dst.IsNil() {
				dst.Set(reflect.MakeSlice(styp, 0, len(pp.Args)))
			}
			for idx, ppp := range pp.Args {
				sval := reflect.New(styp.Elem())
				if sval.Type().Kind() == reflect.Ptr && sval.IsNil() && sval.CanSet() {
					sval.Set(reflect.New(sval.Type().Elem()))
				}
				// decode from value prim
				if err := ppp.unmarshal(sval); err != nil && !finfo.nofail {
					return err
				}
				dst.SetLen(idx + 1)
				dst.Index(idx).Set(sval.Elem())
			}
		case T_MAP:
			mtyp := dst.Type()
			switch mtyp.Key().Kind() {
			case reflect.String:
			default:
				return fmt.Errorf("micheline: only string keys are supported for map %s", finfo.name)
			}
			if dst.IsNil() {
				dst.Set(reflect.MakeMap(mtyp))
			}

			// process ELT args
			for _, ppp := range pp.Args {
				// must be an ELT
				if ppp.OpCode != D_ELT {
					return fmt.Errorf("micheline: expected ELT data for map field %s, got %s",
						finfo.name, ppp.Dump())
				}
				// decode string from ELT key
				k, err := NewKey(ppp.Args[0].BuildType(), ppp.Args[0])
				if err != nil {
					return fmt.Errorf("micheline: cannot convert ELT key for field %s val=%s: %v",
						finfo.name, ppp.Args[0].Dump(), err)
				}
				name := k.String()

				// allocate value
				mval := reflect.New(mtyp.Elem()).Elem()
				if mval.Type().Kind() == reflect.Ptr && mval.IsNil() && mval.CanSet() {
					mval.Set(reflect.New(mval.Type().Elem()))
				}

				// decode from value prim
				if err := ppp.Args[1].unmarshal(mval); err != nil && !finfo.nofail {
					return err
				}

				// assign to map
				dst.SetMapIndex(reflect.ValueOf(name), mval)
			}
		default:
			return fmt.Errorf("micheline: unsupported prim %#v for struct field %s", pp, finfo.name)

		}
	}
	return nil
}
