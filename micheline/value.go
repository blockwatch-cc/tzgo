// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

const EMPTY_LABEL = `@%%@` // illegal Michelson annotation value

type Value struct {
	Type   Type
	Value  Prim
	mapped interface{}
}

func NewValue(typ Type, val Prim) Value {
	return Value{
		Type:  typ.Clone(),
		Value: val.Clone(),
	}
}

func NewValuePtr(typ Type, val Prim) *Value {
	v := NewValue(typ, val)
	return &v
}

func (v *Value) Decode(buf []byte) error {
	return v.Value.UnmarshalBinary(buf)
}

func (v Value) IsPacked() bool {
	return v.Value.IsPacked()
}

func (v Value) IsPackedAny() bool {
	return v.Value.IsPackedAny()
}

func (v Value) Unpack() (Value, error) {
	if !v.Value.IsPacked() {
		return v, nil
	}
	up, err := v.Value.Unpack()
	if err != nil {
		return v, err
	}
	vv := Value{
		Type:  v.Type.Clone(),
		Value: up,
	}
	return vv, nil
}

func (v Value) UnpackAll() (Value, error) {
	if !v.Value.IsPackedAny() {
		return v, nil
	}
	up, err := v.Value.UnpackAll()
	if err != nil {
		return v, err
	}
	vv := Value{
		Type:  v.Type.Clone(),
		Value: up,
	}
	return vv, nil
}

func (e *Value) FixType() {
	labels := e.Type.Anno
	e.Type = e.Value.BuildType()
	e.Type.WasPacked = true
	e.Type.Anno = labels
}

func (e *Value) Map() (interface{}, error) {
	if e.mapped != nil {
		return e.mapped, nil
	}
	m := make(map[string]interface{})
	if err := walkTree(m, EMPTY_LABEL, e.Type, NewStack(e.Value), 0); err != nil {
		return nil, err
	}
	e.mapped = m

	// lift scalar values
	if len(m) == 1 {
		for n, v := range m {
			if n == "0" {
				e.mapped = v
			}
		}
	}

	return e.mapped, nil
}

func (e Value) MarshalJSON() ([]byte, error) {
	m, err := e.Map()
	if err != nil {
		log.Errorf("RENDER: %v", err)
		type xErrorMessage struct {
			Message string `json:"message"`
			Type    Prim   `json:"type"`
			Value   Prim   `json:"value"`
		}
		resp := struct {
			Error xErrorMessage `json:"error"`
		}{
			Error: xErrorMessage{
				Message: err.Error(),
				Type:    e.Type.Prim,
				Value:   e.Value,
			},
		}
		return json.Marshal(resp)
	}

	return json.Marshal(m)
}

func walkTree(m map[string]interface{}, label string, typ Type, stack *Stack, lvl int) error {
	// abort infinite type recursions
	if lvl > 99 {
		return fmt.Errorf("micheline: max nesting level reached")
	}

	// take next value from stack
	val := stack.Pop()

	// Unwrap pairs and pair-like values unless they are containers and push them
	// onto the existing value stack. Skip all values that
	//
	// - are at the top-level of the value tree (we always start with the outermost type first)
	// - look like lambdas or containers (list/set/map will be handled separatly)
	// - have been unpacked (we will guess their type below)
	//
	// If we see an unfoldable type (i.e. one that looks like a container) but at the
	// same time we also see that the next element type is scalar, then we unfold too.
	// This helps resolve ambiguities in value trees.
	//
	for {
		if lvl > 0 && !val.WasPacked && !val.LooksLikeLambda() && (val.CanUnfold(typ) || (val.IsSequence() && typ.IsScalarType())) {
			// fmt.Printf("L%0d: %s UNFOLD SEQ PAIR args[%d(+%d)]=%s\n", lvl, label, stack.Len(), len(val.Args), val.Dump())
			// stack.Push(val.Args...)
			unfolded := val.UnfoldPair(typ)
			stack.Push(unfolded...)
			val = stack.Pop()
		} else {
			// fmt.Printf("L%0d: %s NO UNFOLD SEQ PAIR canunf=%t isseq=%t isscal=%t islambda=%t waspack=%t iscont=%t len=%d %s/%s\n",
			// 	lvl, label, val.CanUnfold(typ), val.IsSequence(), typ.IsScalarType(), val.LooksLikeLambda(), val.WasPacked,
			// 	val.LooksLikeContainer(), len(val.Args), val.Args[0].OpCode, val.Args[0].Type)
			break
		}
	}

	// // Trace Helper
	// ps := func(p Prim) string {
	// 	if p.WasPacked {
	// 		return "unpacked"
	// 	}
	// 	return ""
	// }
	// oc := func(p Prim) string {
	// 	if p.OpCode == 0 {
	// 		return p.Type.String()
	// 	}
	// 	return p.OpCode.String()
	// }
	// fmt.Printf("L%0d: %s/%s %s typ=%s %s\n", lvl, label, typ.Label(), ps(typ.Prim), typ.OpCode, typ.Dump())
	// fmt.Printf("L%0d: %s/%s %s val=%s %s\n", lvl, label, typ.Label(), ps(val), oc(val), val.Dump())

	// detect type for unpacked values
	if val.WasPacked && (!val.IsScalar() || typ.OpCode == T_BYTES) {
		labels := typ.Anno
		typ = val.BuildType()
		typ.WasPacked = true
		typ.Anno = labels
	}

	// make sure value we're going to process actually matches next type
	// we accept pairs which will be recursively unfolded
	if !typ.IsPair() && !val.matchOpCode(typ.OpCode) {
		return fmt.Errorf("micheline: type mismatch: type[%s]=%s value[%s/%d]=%s",
			typ.OpCode, typ.DumpLimit(512), val.Type, val.OpCode, val.DumpLimit(512))
	}

	// get the label from our type tree
	typeLabel := typ.Label()
	haveTypeLabel := len(typeLabel) > 0
	haveKeyLabel := label != EMPTY_LABEL && len(label) > 0
	if label == EMPTY_LABEL {
		if haveTypeLabel {
			// overwrite struct field label from type annotation
			label = typeLabel
		} else {
			// or use sequence number when type annotation is empty
			label = strconv.Itoa(len(m))
		}
	}

	// based on the type code we attach new sub-records, array elements from values
	switch typ.OpCode {
	case T_SET:
		// set <comparable type>
		arr := make([]interface{}, 0, len(val.Args))
		for _, v := range val.Args {
			if v.IsScalar() && !v.IsSequence() {
				// array of scalar types
				arr = append(arr, v.Value(typ.Args[0].OpCode))
			} else {
				// array of complex types
				mm := make(map[string]interface{})
				if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[0]}, NewStack(v), lvl+1); err != nil {
					return err
				}
				arr = append(arr, mm)
			}
		}
		m[label] = arr

	case T_LIST:
		// list <type>
		arr := make([]interface{}, 0, len(val.Args))
		for i, v := range val.Args {
			// lists may contain different types, i.e. when unpack+detect is used
			valType := typ.Args[0]
			if len(typ.Args) > i {
				valType = typ.Args[i]
			}
			if v.IsScalar() && !v.IsSequence() {
				// array of scalar types
				arr = append(arr, v.Value(valType.OpCode))
			} else {
				// array of complex types
				mm := make(map[string]interface{})
				if err := walkTree(mm, EMPTY_LABEL, Type{valType}, NewStack(v), lvl+1); err != nil {
					return err
				}
				arr = append(arr, mm)
			}
		}
		m[label] = arr

	case T_LAMBDA:
		// LAMBDA <type> <type> { <instruction> ... }
		// value_type, return_type, code
		// fmt.Printf("LAMBDA typ=%s\n", typ.Dump())
		// fmt.Printf("LAMBDA val=%s\n", val.Dump())
		// if len(typ.Args) > 0 {
		// 	// only well-defined lambdas have all data available
		// 	mm := make(map[string]interface{})
		// 	n := "_params"
		// 	if typ.Args[0].HasAnno() {
		// 		n = typ.Args[0].GetVarAnnoAny()
		// 	}
		// 	mm[n] = typ.Args[0]
		// 	// n = "_code"
		// 	// if typ.Args[1].Args[0].HasAnno() {
		// 	// 	n = typ.Args[1].Args[0].GetVarAnnoAny()
		// 	// }
		// 	mm["_code"] = val
		// 	n = "_return"
		// 	if typ.Args[1].HasAnno() {
		// 		n = typ.Args[1].GetVarAnnoAny()
		// 	}
		// 	mm[n] = typ.Args[1]
		// 	m[label] = mm
		// } else {
		// 	// unpacked lambdas lack type info
		// 	m[label] = val
		// }
		m[label] = val

	case T_MAP, T_BIG_MAP:
		// map <comparable type> <type>
		// big_map <comparable type> <type>
		// sequence of Elt (key/value) pairs

		// render bigmap reference
		if typ.OpCode == T_BIG_MAP && len(val.Args) == 0 {
			// log.Debugf("%*s> marshal %s ref key %s into map %p", lvl, "", typ.OpCode, label, m)
			switch val.Type {
			case PrimInt:
				// Babylon bigmaps contain a reference here
				m[label] = val.Value(T_INT)
			case PrimSequence:
				// pre-babylon there's only an empty sequence
				// FIXME: we could insert the bigmap id, but this is unknown at ths point
				m[label] = nil
			}
			return nil
		}

		switch val.Type {
		case PrimBinary: // single ELT
			// log.Debugf("%*s> T_MAP marshal single %s key %s len %d into map %p adding sub map %p", lvl, "", typ.OpCode, label, len(val.Args), m, mm)
			keyType := Type{typ.Args[0]}
			valType := Type{typ.Args[1]}

			// build type info if prim was packed
			if val.Args[0].WasPacked {
				keyType = val.Args[0].BuildType()
			}

			// build type info if prim was packed
			if val.Args[1].WasPacked {
				valType = val.Args[1].BuildType()
			}

			// prepare key
			key, err := NewKey(keyType, val.Args[0])
			if err != nil {
				return err
			}

			mm := make(map[string]interface{})
			if err := walkTree(mm, key.String(), valType, NewStack(val.Args[1]), lvl+1); err != nil {
				return err
			}
			m[label] = mm

		case PrimSequence: // sequence of ELTs
			mm := make(map[string]interface{})
			for _, v := range val.Args {
				if v.OpCode != D_ELT {
					return fmt.Errorf("micheline: unexpected type %s [%s] for %s Elt item", v.Type, v.OpCode, typ.OpCode)
				}

				keyType := Type{typ.Args[0]}
				valType := Type{typ.Args[1]}

				// build type info if prim was packed
				if v.Args[0].WasPacked {
					keyType = v.Args[0].BuildType()
				}

				// build type info if prim was packed
				if v.Args[1].WasPacked {
					valType = v.Args[1].BuildType()
				}

				key, err := NewKey(keyType, v.Args[0])
				if err != nil {
					return err
				}

				if err := walkTree(mm, key.String(), valType, NewStack(v.Args[1]), lvl+1); err != nil {
					return err
				}
			}
			m[label] = mm

		default:
			buf, _ := json.Marshal(val)
			return fmt.Errorf("%*s> micheline: unexpected type %s [%s] for %s Elt sequence: %s",
				lvl, "", val.Type, val.OpCode, typ.OpCode, buf)
		}

	case T_PAIR:
		// pair <type> <type> or COMB
		// FIXME: does this catch all COMB pairs??
		// - for n=2, [Pair x1 x2],
		// - for n=3, [Pair x1 (Pair x2 x3)],
		// - for n>=4, [{x1; x2; ...; xn}].

		mm := m
		if haveTypeLabel || haveKeyLabel {
			mm = make(map[string]interface{})
		}

		// Try unfolding value (again) when type is T_PAIR,
		// reuse the existing stack and push unfolded values
		if val.IsPair() {
			// unfold regular pair
			unfolded := val.UnfoldPair(typ)
			// fmt.Printf("L%0d: %s UNFOLD PAIR args[%d(+%d)]=%s\n", lvl, label, stack.Len(), len(unfolded), seq(unfolded...).Dump())
			stack.Push(unfolded...)
		} else if val.CanUnfold(typ) {
			// comb pair
			// fmt.Printf("L%0d: %s PUSH COMB args[%d(+%d)]=%s\n", lvl, label, stack.Len(), len(val.Args), val.Dump())
			stack.Push(val.Args...)
		} else {
			// push value back on stack
			// fmt.Printf("L%0d: %s PUSH VAL args[%d(+1)]=%s\n", lvl, label, stack.Len(), val.Dump())
			stack.Push(val)
		}

		for _, t := range typ.Args {
			// fmt.Printf("L%0d: %s/%s[%d/%d] CHILD=%s\n", lvl, label, t.GetVarAnnoAny(), i, len(typ.Args), stack.Peek().Dump())
			if err := walkTree(mm, EMPTY_LABEL, Type{t}, stack, lvl+1); err != nil {
				return err
			}
		}

		if haveTypeLabel || haveKeyLabel {
			m[label] = mm
		}

	case T_OPTION:
		// option <type>
		switch val.OpCode {
		case D_NONE:
			// add empty option values as null
			m[label] = nil
		case D_SOME:
			// with annots (name) use it for scalar or complex render
			if val.IsScalar() {
				if err := walkTree(m, label, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
			} else {
				mm := make(map[string]interface{})
				if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
				m[label] = mm
			}
		default:
			return fmt.Errorf("micheline: unexpected T_OPTION code %s [%s]", val.OpCode, val.OpCode)
		}

	case T_OR:
		// or <type> <type>
		mm := m
		if haveTypeLabel || haveKeyLabel {
			mm = make(map[string]interface{})
		}
		switch val.OpCode {
		case D_LEFT:
			if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
				return err
			}
		case D_RIGHT:
			if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[1]}, NewStack(val.Args[0]), lvl+1); err != nil {
				return err
			}
		default:
			return fmt.Errorf("micheline: unexpected T_OR branch with value opcode %s", val.OpCode)
		}
		if haveTypeLabel || haveKeyLabel {
			m[label] = mm
		}

	case T_TICKET:
		// always Pair( ticketer:address, Pair( original_type, int ))
		stack.Push(val)
		if err := walkTree(m, label, TicketType(typ.Args[0]), stack, lvl+1); err != nil {
			return err
		}

	case T_SAPLING_STATE:
		mm := make(map[string]interface{})
		if err := walkTree(mm, "memo_size", Type{NewPrim(T_INT)}, NewStack(typ.Args[0]), lvl+1); err != nil {
			return err
		}
		if err := walkTree(mm, "content", val.BuildType(), NewStack(val), lvl+1); err != nil {
			return err
		}
		m[label] = mm

	default:
		// int
		// nat
		// string
		// bytes
		// mutez
		// bool
		// key_hash
		// timestamp
		// address
		// key
		// unit
		// signature
		// operation
		// contract <type> (??)
		// chain_id
		// never
		// append scalar or other complex value
		if val.IsScalar() {
			m[label] = val.Value(typ.OpCode)
		} else {
			mm := make(map[string]interface{})
			if err := walkTree(mm, EMPTY_LABEL, typ, NewStack(val), lvl+1); err != nil {
				return err
			}
			m[label] = mm
		}
	}
	return nil
}

func (p Prim) matchOpCode(oc OpCode) bool {
	mismatch := false
	switch p.Type {
	case PrimSequence:
		switch oc {
		case T_LIST, T_MAP, T_BIG_MAP, T_SET, T_LAMBDA, T_OR, T_OPTION, T_PAIR,
			T_SAPLING_STATE, T_TICKET:
		default:
			mismatch = true
		}

	case PrimInt:
		switch oc {
		case T_INT, T_NAT, T_MUTEZ, T_TIMESTAMP, T_BIG_MAP, T_OR, T_OPTION, T_SAPLING_STATE,
			T_BLS12_381_G1, T_BLS12_381_G2, T_BLS12_381_FR, // maybe stored as bytes
			T_TICKET:
			// accept references to bigmap and sapling states
		default:
			mismatch = true
		}

	case PrimString:
		// sometimes timestamps and addresses can be strings
		switch oc {
		case T_BYTES, T_STRING, T_ADDRESS, T_CONTRACT, T_KEY_HASH, T_KEY,
			T_SIGNATURE, T_TIMESTAMP, T_OR, T_CHAIN_ID, T_OPTION,
			T_TICKET:
		default:
			mismatch = true
		}

	case PrimBytes:
		switch oc {
		case T_BYTES, T_STRING, T_BOOL, T_ADDRESS, T_KEY_HASH, T_KEY,
			T_CONTRACT, T_SIGNATURE, T_OPERATION, T_LAMBDA, T_OR,
			T_CHAIN_ID, T_OPTION, T_SAPLING_STATE, T_SAPLING_TRANSACTION,
			T_BLS12_381_G1, T_BLS12_381_G2, T_BLS12_381_FR, // maybe stored as bytes
			T_TICKET: // allow ticket since first value is ticketer address
		default:
			mismatch = true
		}

	default:
		switch p.OpCode {
		case D_PAIR:
			mismatch = oc != T_PAIR && oc != T_OR && oc != T_LIST && oc != T_OPTION && oc != T_TICKET
		case D_SOME, D_NONE:
			mismatch = oc != T_OPTION
		case D_UNIT:
			mismatch = oc != T_UNIT && oc != K_PARAMETER
		case D_LEFT, D_RIGHT:
			mismatch = oc != T_OR
		}
	}

	return !mismatch
}

func (v *Value) GetValue(label string) (interface{}, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			return vv, ok
		}
	}
	return nil, false
}

func (v *Value) GetString(label string) (string, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			if s, ok := vv.(string); ok {
				return s, true
			} else {
				return fmt.Sprint(s), true
			}
		}
	}
	return "", false
}

func (v *Value) GetBytes(label string) ([]byte, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// hex string or nil
			if vv == nil {
				return nil, ok
			}
			if s, ok := vv.(string); ok {
				h, err := hex.DecodeString(s)
				if err == nil {
					return h, true
				}
			}
		}
	}
	return nil, false
}

func (v *Value) GetInt64(label string) (int64, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// big, string or nil
			if vv == nil {
				return 0, ok
			}
			switch t := vv.(type) {
			case *big.Int:
				return t.Int64(), true
			case string:
				i, err := strconv.ParseInt(t, 10, 64)
				if err == nil {
					return i, true
				}
			}
		}
	}
	return 0, false
}

func (v *Value) GetBig(label string) (*big.Int, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// big, string or nil
			if vv == nil {
				return big.NewInt(0), ok
			}
			switch t := vv.(type) {
			case *big.Int:
				return t, true
			case string:
				return big.NewInt(0).SetString(t, 10)
			}
		}
	}
	return nil, false
}

func (v *Value) GetBool(label string) (bool, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// bool, string or nil
			if vv == nil {
				return false, ok
			}
			switch t := vv.(type) {
			case bool:
				return t, true
			case string:
				if b, err := strconv.ParseBool(t); err == nil {
					return b, true
				}
			}
		}
	}
	return false, false
}

func (v *Value) GetTime(label string) (time.Time, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// time, string or nil
			if vv == nil {
				return time.Time{}, ok
			}
			switch t := vv.(type) {
			case time.Time:
				return t, true
			case string:
				if b, err := time.Parse(t, time.RFC3339); err == nil {
					return b, true
				}
			}
		}
	}
	return time.Time{}, false
}

func (v *Value) GetAddress(label string) (tezos.Address, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// Adddress, string or nil
			if vv == nil {
				return tezos.InvalidAddress, ok
			}
			switch t := vv.(type) {
			case tezos.Address:
				return t, true
			case string:
				if b, err := tezos.ParseAddress(t); err == nil {
					return b, true
				}
			}
		}
	}
	return tezos.InvalidAddress, false
}

func (v *Value) GetKey(label string) (tezos.Key, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// Key, string or nil
			if vv == nil {
				return tezos.InvalidKey, ok
			}
			switch t := vv.(type) {
			case tezos.Key:
				return t, true
			case string:
				if b, err := tezos.ParseKey(t); err == nil {
					return b, true
				}
			}
		}
	}
	return tezos.InvalidKey, false
}

func (v *Value) GetSignature(label string) (tezos.Signature, bool) {
	if m, err := v.Map(); err == nil {
		if vv, ok := getPath(m, label); ok {
			// Signature, string or nil
			if vv == nil {
				return tezos.InvalidSignature, ok
			}
			switch t := vv.(type) {
			case tezos.Signature:
				return t, true
			case string:
				if b, err := tezos.ParseSignature(t); err == nil {
					return b, true
				}
			}
		}
	}
	return tezos.InvalidSignature, false
}

type ValueWalkerFunc func(label string, value interface{}) error

func (v *Value) Walk(label string, fn ValueWalkerFunc) error {
	val, ok := v.GetValue(label)
	if !ok {
		return nil
	}
	return walkValueMap(label, val, fn)
}
