// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

type PrimMarshaler interface {
	MarshalPrim() (Prim, error)
}

// SetPath replaces a nested primitive at path with dst.
// Path segments are separated by slash (/).
// Works on both type and value primitive trees.
func (p *Prim) SetPath(path string, dst Prim) error {
	index, err := p.getIndex(path)
	if err != nil {
		return err
	}
	return p.SetIndex(index, dst)
}

// SetPathExt replaces a nested primitive at path with dst if the primitive matches
// the expected type. Path segments are separated by slash (/).
// Works on best on value primitive trees.
func (p *Prim) SetPathExt(path string, typ PrimType, dst Prim) error {
	index, err := p.getIndex(path)
	if err != nil {
		return err
	}
	return p.SetIndexExt(index, typ, dst)
}

// SetIndex replaces a nested primitive at path index with dst.
func (p *Prim) SetIndex(index []int, dst Prim) error {
	prim := p
	for _, v := range index {
		if v < 0 || len(prim.Args) <= v {
			return fmt.Errorf("micheline: index %d out of bounds", v)
		}
		prim = &prim.Args[v]
	}
	*prim = dst
	return nil
}

// SetIndexExt replaces a nested primitive at path index if the primitive matches the
// expected primitive type. This function works best with value trees which
// lack opcode info. Use as extra cross-check when replacing prims.
func (p *Prim) SetIndexExt(index []int, typ PrimType, dst Prim) error {
	prim := p
	for _, v := range index {
		if v < 0 || len(prim.Args) <= v {
			return fmt.Errorf("micheline: index %d out of bounds", v)
		}
		prim = &prim.Args[v]
	}
	if prim.Type != typ {
		return fmt.Errorf("micheline: unexpected type %s at path %v", prim.Type, index)
	}
	*prim = dst
	return nil
}

// Marshal takes a scalar or nested Go type and populates a Micheline
// primitive tree compatible with type t. This method is compatible
// with most contract entrypoints, contract storage, bigmap values, etc.
// Use optimized to control whether the target prims contain values in
// optimized form (binary addresses, numeric timestamps) or string form.
//
// Note: This is work in progress. Several data types are still unsupported
// and entrypoint mapping requires some extra boilerplate:
//
//	// Entrypoint example (without error handling for brevity)
//	eps, _ := script.Entrypoints(true)
//	ep, _ := eps["name"]
//
//	// marshal to prim tree
//	// Note: be mindful of the way entrypoint typedefs are structured:
//	// - 1 arg: use scalar value in ep.Typedef[0]
//	// - >1 arg: use entire list in ep.Typedef but wrap into struct
//	typ := ep.Typedef[0]
//	if len(ep.Typedef) > 1 {
//	    typ = micheline.Typedef{
//	        Name: micheline.CONST_ENTRYPOINT,
//	        Type: micheline.TypeStruct,
//	        Args: ep.Typedef,
//	    }
//	}
//
//	// then use the type to marshal into primitives
//	prim, err := typ.Marshal(args, true)
func (t Typedef) Marshal(v any, optimized bool) (Prim, error) {
	return t.marshal(v, optimized, 0)
}

func (t Typedef) marshal(v any, optimized bool, depth int) (Prim, error) {
	// fmt.Printf("Marshal %T %v => %#v\n", v, v, t)
	if t.Optional {
		val := v
		if t.Name != "" && val != nil {
			vals, ok := v.(map[string]any)
			if ok {
				val, ok = vals[t.Name]
				if !ok {
					return InvalidPrim, fmt.Errorf("missing arg %s", t.Name)
				}
			}
		}
		if val != nil {
			t.Optional = false
			p, err := t.marshal(val, optimized, depth+1)
			if err != nil {
				return InvalidPrim, err
			}
			return NewOption(p), nil
		} else {
			return NewOption(), nil
		}
	}
	switch t.Type {
	case TypeUnion:
		// find the named union element in map
		vals, ok := v.(map[string]any)
		if !ok {
			return InvalidPrim, fmt.Errorf("invalid type %T on union %s", v, t.Name)
		}
		var child Typedef
		for _, n := range t.Args {
			if _, ok := vals[n.Name]; ok {
				child = n
				break
			}
		}
		// marshal child type
		p, err := child.marshal(vals[child.Name], optimized, depth+1)
		if err != nil {
			return InvalidPrim, err
		}
		// produce OR tree for child's path
		return NewUnion(child.Path[depth:], p), nil

	case TypeStruct:
		vals, ok := v.(map[string]any)
		if !ok {
			return InvalidPrim, fmt.Errorf("invalid type %T on struct %s", v, t.Name)
		}
		// for values with nested named structs try if name exists
		if m, ok := vals[t.Name]; t.Name != "" && ok {
			fmt.Printf("Unpacking nested struct %s\n", t.Name)
			vals, ok = m.(map[string]any)
			if !ok {
				return InvalidPrim, fmt.Errorf("invalid type %T on nested struct %s", m, t.Name)
			}
		}
		prims := []Prim{}
		for _, v := range t.Args {
			p, err := v.marshal(vals[v.Name], optimized, depth+1)
			if err != nil {
				return InvalidPrim, err
			}
			prims = append(prims, p)
		}
		if len(prims) > 2 {
			// reconstruct struct structure as Pair tree from type paths
			var root Prim
			for i, v := range prims {
				root.Insert(v, t.Args[i].Path[depth:])
			}
			return root, nil
		}
		return NewPair(prims[0], prims[1]), nil

	case "list", "set":
		if v == nil {
			return NewSeq(), nil
		}
		listVals, ok := v.([]any)
		if !ok {
			// use nested value for named lists
			vals, ok := v.(map[string]any)
			if !ok {
				return InvalidPrim, fmt.Errorf("invalid list/set type %T on field %s, must be map[string]any", v, t.Name)
			}
			list, ok := vals[t.Name]
			if !ok {
				return InvalidPrim, fmt.Errorf("missing list/set arg %s", t.Name)
			}
			listVals, ok = list.([]any)
			if !ok {
				return InvalidPrim, fmt.Errorf("invalid list/set type %T on field %s, must be []any", list, t.Name)
			}
		}
		prims := []Prim{}
		for _, v := range listVals {
			p, err := t.Args[0].marshal(v, optimized, depth+1)
			if err != nil {
				return InvalidPrim, err
			}
			prims = append(prims, p)
		}
		return NewSeq(prims...), nil

	case "map", "big_map":
		if v == nil {
			return NewMap(), nil
		}
		vals, ok := v.(map[string]any)
		if !ok {
			return InvalidPrim, fmt.Errorf("invalid map type %T on field %s, must be map[string]any", v, t.Name)
		}
		// for top-level maps (in entrypoints etc) try if map name is part of value tree
		if depth == 0 {
			if m, ok := vals[t.Name]; ok {
				vals, ok = m.(map[string]any)
				if !ok {
					return InvalidPrim, fmt.Errorf("invalid map type %T on field %s, must be map[string]any", m, t.Name)
				}
			}
		}
		prims := []Prim{}
		for n, v := range vals {
			key, err := ParsePrim(t.Left(), n, optimized)
			if err != nil {
				return InvalidPrim, err
			}
			value, err := t.Right().marshal(v, optimized, depth+1)
			if err != nil {
				return InvalidPrim, err
			}
			prims = append(prims, NewMapElem(key, value))
		}
		return NewMap(prims...), nil

	case "lambda":
		switch val := v.(type) {
		case string:
			var p Prim
			err := p.UnmarshalJSON([]byte(val))
			return p, err
		case PrimMarshaler:
			return val.MarshalPrim()
		case Prim:
			return val, nil
		default:
			return InvalidPrim, fmt.Errorf("unsupported type %T for lambda on field %s", v, t.Name)
		}

	default:
		// scalar
		oc := t.OpCode()
		if !oc.IsValid() {
			return InvalidPrim, fmt.Errorf("invalid type code %s on field %s", t.Type, t.Name)
		}
		if v == nil {
			if oc == T_UNIT {
				return NewCode(D_UNIT), nil
			}
			return InvalidPrim, fmt.Errorf("missing arg %s (%s)", t.Name, t.Type)
		}
		switch val := v.(type) {
		case map[string]any:
			// recurse unpack the named value from this map
			return t.marshal(val[t.Name], optimized, depth)
		case Prim:
			return val, nil
		case PrimMarshaler:
			return val.MarshalPrim( /* optimized */ )
		case string:
			// parse anything from string (supports config file and API map[string]string)
			return ParsePrim(t, val, optimized)
		case []byte:
			return NewBytes(val), nil
		case bool:
			if val {
				return NewCode(D_TRUE), nil
			}
			return NewCode(D_FALSE), nil
		case int:
			switch oc {
			case T_BYTES:
				return NewBytes([]byte(strconv.FormatInt(int64(val), 10))), nil
			case T_STRING:
				return NewString(strconv.FormatInt(int64(val), 10)), nil
			case T_TIMESTAMP:
				if optimized {
					return NewInt64(int64(val)), nil
				}
				return NewString(time.Unix(int64(val), 0).UTC().Format(time.RFC3339)), nil
			case T_INT, T_NAT, T_MUTEZ:
				return NewInt64(int64(val)), nil
			default:
				return InvalidPrim, fmt.Errorf("unsupported type conversion %T to opcode %s for on field %s", v, t.Type, t.Name)
			}
		case int64:
			switch oc {
			case T_BYTES:
				return NewBytes([]byte(strconv.FormatInt(val, 10))), nil
			case T_STRING:
				return NewString(strconv.FormatInt(val, 10)), nil
			case T_TIMESTAMP:
				if optimized {
					return NewInt64(val), nil
				}
				return NewString(time.Unix(val, 0).UTC().Format(time.RFC3339)), nil
			case T_INT, T_NAT, T_MUTEZ:
				return NewInt64(val), nil
			default:
				return InvalidPrim, fmt.Errorf("unsupported type conversion %T to opcode %s on field %s", v, t.Type, t.Name)
			}
		case time.Time:
			if optimized {
				return NewInt64(val.Unix()), nil
			}
			return NewString(val.UTC().Format(time.RFC3339)), nil
		case tezos.Address:
			if optimized {
				switch oc {
				case T_KEY_HASH:
					return NewKeyHash(val), nil
				case T_ADDRESS:
					return NewAddress(val), nil
				default:
					return InvalidPrim, fmt.Errorf("unsupported type conversion from %T to opcode %s on field %s", v, t.Type, t.Name)
				}
			}
			return NewString(val.String()), nil
		case tezos.Key:
			if optimized {
				return NewBytes(val.Bytes()), nil
			}
			return NewString(val.String()), nil
		case tezos.Signature:
			if optimized {
				return NewBytes(val.Bytes()), nil
			}
			return NewString(val.String()), nil
		case tezos.ChainIdHash:
			return NewString(val.String()), nil

		default:
			// TODO
			return InvalidPrim, fmt.Errorf("unsupported type %T for opcode %s on field %s", v, t.Type, t.Name)
		}
	}
}

func ParsePrim(typ Typedef, val string, optimized bool) (p Prim, err error) {
	p = InvalidPrim
	if !typ.OpCode().IsTypeCode() {
		err = fmt.Errorf("invalid type code %q", typ)
		return
	}
	switch typ.OpCode() {
	case T_INT, T_NAT, T_MUTEZ:
		i := big.NewInt(0)
		err = i.UnmarshalText([]byte(val))
		p = NewBig(i)
	case T_STRING:
		p = NewString(val)
	case T_BYTES:
		if buf, err2 := hex.DecodeString(val); err2 != nil {
			p = NewBytes([]byte(val))
		} else {
			p = NewBytes(buf)
		}
	case T_BOOL:
		var b bool
		b, err = strconv.ParseBool(val)
		if b {
			p = NewCode(D_TRUE)
		} else {
			p = NewCode(D_FALSE)
		}
	case T_TIMESTAMP:
		// either RFC3339 or UNIX seconds
		var tm time.Time
		if strings.Contains(val, "T") {
			tm, err = time.Parse(time.RFC3339, val)
		} else {
			var i int64
			i, err = strconv.ParseInt(val, 10, 64)
			tm = time.Unix(i, 0).UTC()
		}
		if optimized {
			p = NewInt64(tm.Unix())
		} else {
			p = NewString(tm.Format(time.RFC3339))
		}
	case T_KEY_HASH:
		var addr tezos.Address
		addr, err = tezos.ParseAddress(val)
		if optimized {
			p = NewKeyHash(addr)
		} else {
			p = NewString(addr.String())
		}
	case T_ADDRESS:
		var addr tezos.Address
		addr, err = tezos.ParseAddress(val)
		if optimized {
			p = NewAddress(addr)
		} else {
			p = NewString(addr.String())
		}
	case T_KEY:
		var key tezos.Key
		key, err = tezos.ParseKey(val)
		if optimized {
			p = NewBytes(key.Bytes())
		} else {
			p = NewString(key.String())
		}

	case T_SIGNATURE:
		var sig tezos.Signature
		sig, err = tezos.ParseSignature(val)
		if optimized {
			p = NewBytes(sig.Bytes())
		} else {
			p = NewString(sig.String())
		}

	case T_UNIT:
		if val == D_UNIT.String() || val == "" {
			p = NewCode(D_UNIT)
		} else {
			err = fmt.Errorf("micheline: invalid value %q for unit type", val)
		}

	case T_PAIR:
		// parse comma-separated list into map using type lables from typedef
		// note: this only supports simple structs which is probably enough
		// because bigmap keys must be comparable types
		m := make(map[string]any)
		for i, v := range strings.Split(val, ",") {
			// find i-th child in typedef
			if len(typ.Args) < i-1 {
				err = fmt.Errorf("micheline: invalid value for bigmap key struct type %s", typ.Name)
				return
			}
			m[typ.Args[i].Name] = v
		}
		return typ.marshal(m, optimized, 0)

	default:
		err = fmt.Errorf("micheline: unsupported big_map key type %s", typ)
	}

	if err != nil {
		p = InvalidPrim
	}
	return
}

func (p *Prim) Insert(src Prim, path []int) {
	if !p.IsValid() {
		*p = NewPair(Prim{}, Prim{})
	}

	if len(p.Args) <= path[0] {
		cp := make([]Prim, path[0]+1)
		copy(cp, p.Args)
		p.Args = cp
		// convert to sequence
		p.Type = PrimSequence
		p.OpCode = 0
	}

	if len(path) == 1 {
		p.Args[path[0]] = src
		return
	}

	p.Args[path[0]].Insert(src, path[1:])
}
