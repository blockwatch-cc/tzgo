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

	"github.com/legonian/tzgo/tezos"
)

type PrimUnmarshaler interface {
	UnmarshalPrim(Prim) error
}

type PrimMarshaler interface {
	MarshalPrim() (Prim, error)
}

func (p Prim) Index(label string) ([]int, bool) {
	if p.MatchesAnno(label) {
		return []int{}, true
	}
	found := make([]int, 0)
	for i := range p.Args {
		x, ok := p.Args[i].Index(label)
		if ok {
			found = append(found, x...)
		}
	}
	return found, len(found) > 0
}

func (p Prim) GetPath(path string) (Prim, error) {
	index := make([]int, 0)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	for i, v := range strings.Split(path, "/") {
		switch v {
		case "L", "l", "0":
			index = append(index, 0)
		case "R", "r", "1":
			index = append(index, 1)
		default:
			idx, err := strconv.Atoi(v)
			if err != nil {
				return InvalidPrim, fmt.Errorf("micheline: invalid path component '%v' at pos %d", v, i)
			}
			index = append(index, idx)
		}
	}
	return p.GetIndex(index)
}

func (p Prim) GetPathExt(path string, typ OpCode) (Prim, error) {
	prim, err := p.GetPath(path)
	if err != nil {
		return InvalidPrim, err
	}
	if prim.OpCode != typ {
		return InvalidPrim, fmt.Errorf("micheline: unexpected type %s at path %v", prim.OpCode, path)
	}
	return prim, nil
}

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
			return err
		}
		switch finfo.typ {
		case T_BYTES:
			if dst.CanAddr() {
				pv := dst.Addr()
				if pv.CanInterface() && pv.Type().Implements(binaryUnmarshalerType) {
					if err := pv.Interface().(encoding.BinaryUnmarshaler).UnmarshalBinary(pp.Bytes); err != nil {
						return err
					}
					break
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
							return err
						}
						break
					}
					if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(pp.String)); err != nil {
						return err
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
							return err
						}
						break
					}
					if err := pv.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(pp.String)); err != nil {
						return err
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
					return err
				}
				dst.Set(reflect.ValueOf(tm))
			}
		case T_ADDRESS:
			var (
				addr tezos.Address
				err  error
			)
			if pp.Bytes != nil {
				err = addr.UnmarshalBinary(pp.Bytes)
			} else {
				err = addr.UnmarshalText([]byte(pp.String))
			}
			if err != nil {
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
			if err != nil {
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
			if err != nil {
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
			if err != nil {
				return err
			}
			dst.Set(reflect.ValueOf(chain))
		default:
			return fmt.Errorf("micheline: unsupported prim %#v for struct field %s", pp, finfo.name)

		}
	}
	return nil
}
