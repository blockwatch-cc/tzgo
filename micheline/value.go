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

const (
	EMPTY_LABEL       = `@%%@` // illegal Michelson annotation value
	RENDER_TYPE_PRIM  = 0      // silently output primitive tree instead if human-readable
	RENDER_TYPE_FAIL  = 1      // return error if human-readable formatting fails
	RENDER_TYPE_PANIC = 2      // panic with error if human-readable formatting fails
)

type Value struct {
	Type   Type
	Value  Prim
	Render int
	mapped interface{}
}

func NewValue(typ Type, val Prim) Value {
	return Value{
		Type:   typ.Clone(),
		Value:  val.Clone(),
		Render: RENDER_TYPE_PRIM,
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
		Type:   v.Type.Clone(),
		Value:  up,
		Render: v.Render,
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
		Type:   v.Type.Clone(),
		Value:  up,
		Render: v.Render,
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
		// FIXME: this is a good place to plug in an error reporting facility
		buf, _ := json.Marshal(resp)

		switch e.Render {
		default:
			log.Errorf("RENDER: %s", string(buf))
			// render the plain prim tree
			return json.Marshal(e.Value)
		case RENDER_TYPE_FAIL:
			return buf, err
		case RENDER_TYPE_PANIC:
			panic(err)
		}
	}

	return json.Marshal(m)
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
			T_TICKET, // allow ticket since first value is ticketer address
			T_CHEST, T_CHEST_KEY:
		default:
			mismatch = true
		}

	default:
		switch p.OpCode {
		case D_PAIR:
			switch oc {
			case T_PAIR, T_OR, T_LIST, T_OPTION, T_TICKET:
			default:
				mismatch = true
			}
		case D_SOME, D_NONE:
			switch oc {
			case T_OPTION:
			default:
				mismatch = true
			}
		case D_UNIT:
			switch oc {
			case T_UNIT, K_PARAMETER:
			default:
				mismatch = true
			}
		case D_LEFT, D_RIGHT:
			switch oc {
			case T_OR:
			default:
				mismatch = true
			}
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

func (v *Value) Unmarshal(val interface{}) error {
	if m, err := v.Map(); err == nil {
		buf, _ := json.Marshal(m)
		return json.Unmarshal(buf, val)
	} else {
		return err
	}
}

type ValueWalkerFunc func(label string, value interface{}) error

func (v *Value) Walk(label string, fn ValueWalkerFunc) error {
	val, ok := v.GetValue(label)
	if !ok {
		return nil
	}
	return walkValueMap(label, val, fn)
}
