// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

// Comparable key as used in bigmaps and maps
type Key struct {
	Type      Type
	Hash      tezos.ExprHash
	IntKey    *big.Int
	StringKey string
	BytesKey  []byte
	BoolKey   bool
	AddrKey   tezos.Address
	TimeKey   time.Time
	PrimKey   Prim
}

func NewKey(typ Type, key Prim) (Key, error) {
	k := Key{
		Type: typ,
	}
	switch typ.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		k.IntKey = key.Int
	case T_STRING:
		if isASCII(key.String) {
			k.StringKey = key.String
		} else {
			key.Bytes = []byte(key.String)
		}
		// convert empty and non-ascii strings to bytes
		if len(k.StringKey) == 0 && len(key.Bytes) > 0 {
			k.Type.OpCode = T_BYTES
			k.BytesKey = key.Bytes
		}
	case T_BYTES:
		k.BytesKey = key.Bytes
	case T_BOOL:
		switch key.OpCode {
		case D_TRUE:
			k.BoolKey = true
		case D_FALSE:
			k.BoolKey = false
		default:
			return Key{}, fmt.Errorf("micheline: invalid bool big_map key opcode %s (%[1]d)", key.OpCode)
		}
	case T_TIMESTAMP:
		// in some cases (originated contract storage) timestamps are strings
		if key.Int == nil {
			t, err := time.Parse(time.RFC3339, key.String)
			if err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for string timestamp: %v", err)
			}
			k.TimeKey = t
		} else {
			k.TimeKey = time.Unix(key.Int.Int64(), 0).UTC()
		}
	case T_KEY_HASH, T_ADDRESS:
		// in some cases (originated contract storage) addresses are strings
		if len(key.Bytes) == 0 && len(key.String) > 0 {
			a, err := tezos.ParseAddress(strings.Split(key.String, "%")[0])
			if err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for string type address: %v", err)
			}
			k.AddrKey = a
		} else {
			a := tezos.Address{}
			if err := a.UnmarshalBinary(key.Bytes); err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for type address: %v", err)
			}
			k.AddrKey = a
		}

	case T_KEY:
		k.Hash = tezos.NewExprHash(key.Bytes)

	case T_PAIR, D_PAIR:
		k.PrimKey = key

	default:
		return Key{}, fmt.Errorf("micheline: big_map key type '%s' is not implemented", typ.OpCode)
	}
	return k, nil
}

func NewKeyPtr(typ Type, key Prim) (*Key, error) {
	k, err := NewKey(typ, key)
	return &k, err
}

func (k Key) IsPacked() bool {
	return k.Type.OpCode == T_BYTES && len(k.BytesKey) > 1 && k.BytesKey[0] == 0x5
}

func (k Key) UnpackPrim() (p Prim, err error) {
	p = Prim{}
	if !k.IsPacked() {
		return p, fmt.Errorf("key is not packed")
	}
	defer func() {
		if e := recover(); e != nil {
			p = Prim{}
			err = fmt.Errorf("prim is not packed")
		}
	}()
	if err = p.UnmarshalBinary(k.BytesKey[1:]); err != nil {
		return p, err
	}
	return p, nil
}

func (k Key) Unpack() (Key, error) {
	if !k.IsPacked() {
		return k, nil
	}
	p, err := k.UnpackPrim()
	if err != nil {
		return Key{}, err
	}
	return Key{
		Type:      p.BuildType(),
		IntKey:    p.Int,
		StringKey: p.String,
		BytesKey:  p.Bytes,
		PrimKey:   p,
	}, nil
}

func ParseKeyType(typ string) (OpCode, error) {
	t, err := ParseOpCode(typ)
	if err != nil {
		return t, fmt.Errorf("micheline: invalid big_map key type '%s'", typ)
	}
	switch t {
	case T_INT, T_NAT, T_MUTEZ, T_STRING, T_BYTES, T_BOOL, T_KEY_HASH, T_TIMESTAMP, T_ADDRESS, T_PAIR:
		return t, nil
	default:
		return t, fmt.Errorf("micheline: unsupported big_map key type string '%s'", t)
	}
}

// query string parsing used for lookup (does not support Pair keys)
func ParseKey(typ, val string) (Key, error) {
	var err error
	key := Key{Type: NewType()}
	key.Hash, err = tezos.ParseExprHash(val)
	if err == nil {
		key.Type.OpCode = T_KEY
		key.Type.Type = PrimBytes
		return key, nil
	}
	if typ == "" {
		key.Type.OpCode = InferKeyType(val)
		typ = key.Type.OpCode.String()
	} else {
		key.Type.OpCode, err = ParseOpCode(typ)
		if err != nil {
			return Key{}, err
		}
	}
	switch key.Type.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		key.Type.Type = PrimInt
		key.IntKey = big.NewInt(0)
		err = key.IntKey.UnmarshalText([]byte(val))
	case T_STRING:
		key.Type.Type = PrimString
		key.StringKey = val
	case T_BYTES:
		key.Type.Type = PrimBytes
		key.BytesKey, err = hex.DecodeString(val)
	case T_BOOL:
		key.Type.Type = PrimNullary
		key.BoolKey, err = strconv.ParseBool(val)
	case T_TIMESTAMP:
		// either RFC3339 or UNIX seconds
		key.Type.Type = PrimInt
		if strings.Contains(val, "T") {
			key.TimeKey, err = time.Parse(time.RFC3339, val)
		} else {
			var i int64
			i, err = strconv.ParseInt(val, 10, 64)
			key.TimeKey = time.Unix(i, 0).UTC()
		}
	case T_KEY_HASH, T_ADDRESS:
		key.Type.Type = PrimBytes
		key.AddrKey, err = tezos.ParseAddress(val)
	case T_PAIR:
		// parse comma-separated list into a comb pair tree
		prims := []Prim{}
		for _, v := range strings.Split(val, ",") {
			parsed, err := ParseKey(InferKeyType(v).String(), v)
			if err != nil {
				return Key{}, fmt.Errorf("micheline: decoding bigmap pair key element %s: %v", v, err)
			}
			prims = append(prims, parsed.Prim())
		}
		if len(prims) == 2 {
			key.PrimKey = dpair(prims[0], prims[1])
			key.Type.Type = PrimBinary
		} else {
			key.PrimKey = seq(prims...)
			key.Type.Type = PrimSequence
		}

	default:
		return Key{}, fmt.Errorf("micheline: unsupported big_map key %s type %s", key.Type.Label(), key.Type.OpCode)
	}

	if err != nil {
		return Key{}, fmt.Errorf("micheline: decoding bigmap key %s (%s): %v", val, typ, err)
	}
	return key, nil
}

func InferKeyType(val string) OpCode {
	if _, err := strconv.ParseBool(val); err == nil {
		return T_BOOL
	}
	if a, err := tezos.ParseAddress(val); err == nil {
		if a.Type == tezos.AddressTypeContract {
			return T_ADDRESS
		}
		return T_KEY_HASH
	}
	if _, err := time.Parse(time.RFC3339, val); err == nil {
		return T_TIMESTAMP
	}
	i := big.NewInt(0)
	if err := i.UnmarshalText([]byte(val)); err == nil {
		return T_INT
	}
	if _, err := hex.DecodeString(val); err == nil {
		return T_BYTES
	}
	if strings.Contains(val, ",") {
		return T_PAIR
	}
	return T_STRING
}

func (k Key) Bytes() []byte {
	p := Prim{
		Type:   k.Type.Type,
		OpCode: k.Type.OpCode,
	}
	switch k.Type.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		p.Int = k.IntKey
	case T_STRING:
		p.String = k.StringKey
	case T_BYTES:
		p.Bytes = k.BytesKey
	case T_BOOL:
		if k.BoolKey {
			p.OpCode = D_TRUE
		} else {
			p.OpCode = D_FALSE
		}
	case T_TIMESTAMP:
		var z Z
		z.SetInt64(k.TimeKey.Unix())
		p.Int = z.Big()
	case T_ADDRESS:
		p.Bytes, _ = k.AddrKey.MarshalBinary()
	case T_KEY_HASH:
		b, _ := k.AddrKey.MarshalBinary()
		p.Bytes = b[1:] // strip leading flag
	case T_KEY:
		p.Bytes = k.Hash.Hash.Hash
	case T_PAIR, D_PAIR:
		b, _ := k.PrimKey.MarshalBinary()
		p.Bytes = b
	default:
		return nil
	}
	buf, _ := p.MarshalBinary()
	return buf
}

// Note: this is not the encoding stored in bigmap table! (prim wrapper is missing)
func (k Key) MarshalBinary() ([]byte, error) {
	switch k.Type.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		var z Z
		return z.Set(k.IntKey).MarshalBinary()
	case T_STRING:
		return []byte(k.StringKey), nil
	case T_BYTES:
		return k.BytesKey, nil
	case T_BOOL:
		if k.BoolKey {
			return []byte{byte(D_TRUE)}, nil
		} else {
			return []byte{byte(D_FALSE)}, nil
		}
	case T_TIMESTAMP:
		var z Z
		z.SetInt64(k.TimeKey.Unix())
		return z.MarshalBinary()
	case T_ADDRESS:
		return k.AddrKey.MarshalBinary()
	case T_KEY_HASH:
		b, err := k.AddrKey.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return b[1:], nil // strip leading flag
	case T_KEY:
		return k.Hash.Hash.Hash, nil
	case T_PAIR, D_PAIR:
		return k.PrimKey.MarshalBinary()
	default:
		return nil, fmt.Errorf("micheline: no binary marshaller for big_map key type '%s'", k.Type.OpCode)
	}
}

func DecodeKey(typ Type, b []byte) (Key, error) {
	key := Prim{}
	if err := key.UnmarshalBinary(b); err != nil {
		return Key{}, err
	}
	return NewKey(typ, key)
}

func (k Key) String() string {
	switch k.Type.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		return k.IntKey.Text(10)
	case T_STRING:
		return k.StringKey
	case T_BYTES:
		return hex.EncodeToString(k.BytesKey)
	case T_BOOL:
		return strconv.FormatBool(k.BoolKey)
	case T_TIMESTAMP:
		return k.TimeKey.Format(time.RFC3339)
	case T_KEY_HASH, T_ADDRESS:
		return k.AddrKey.String()
	case T_KEY:
		return k.Hash.String()
	case T_PAIR, D_PAIR:
		left := k.PrimKey.Args[0].Value(k.PrimKey.Args[0].OpCode)
		if _, ok := left.(fmt.Stringer); !ok {
			if _, ok := left.(string); !ok {
				left = k.PrimKey.Args[0].OpCode.String()
			}
		}
		right := k.PrimKey.Args[1].Value(k.PrimKey.Args[1].OpCode)
		if _, ok := right.(fmt.Stringer); !ok {
			if _, ok := right.(string); !ok {
				right = k.PrimKey.Args[1].OpCode.String()
			}
		}
		return fmt.Sprintf("%s,%s", left, right)
	default:
		return ""
	}
}

func (k Key) Encode() string {
	switch k.Type.OpCode {
	case T_STRING:
		return k.StringKey
	case T_INT, T_NAT, T_MUTEZ:
		return k.IntKey.Text(10)
	case T_BYTES:
		return hex.EncodeToString(k.BytesKey)
	default:
		buf, _ := k.MarshalBinary()
		return hex.EncodeToString(buf)
	}
}

func (k Key) Prim() Prim {
	p := Prim{
		OpCode: k.Type.OpCode,
	}
	switch k.Type.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		p.Int = k.IntKey
		p.Type = PrimInt
	case T_TIMESTAMP:
		p.Int = big.NewInt(k.TimeKey.Unix())
		p.Type = PrimInt
	case T_STRING:
		p.String = k.StringKey
		p.Type = PrimString
	case T_BYTES:
		p.Bytes = k.BytesKey
		p.Type = PrimBytes
	case T_ADDRESS, T_KEY_HASH, T_KEY:
		p.Bytes, _ = k.MarshalBinary()
		p.Type = PrimBytes
	case T_BOOL:
		p.Type = PrimNullary
		if k.BoolKey {
			p.OpCode = D_TRUE
		} else {
			p.OpCode = D_FALSE
		}
	case T_PAIR:
		p = k.PrimKey
	default:
		if k.BytesKey != nil {
			if err := p.UnmarshalBinary(k.BytesKey); err == nil {
				break
			}
		}
		p.Bytes, _ = k.MarshalBinary()
		p.Type = PrimBytes
	}
	return p
}

func (k Key) PrimPtr() *Prim {
	p := k.Prim()
	return &p
}
func (k Key) MarshalJSON() ([]byte, error) {
	switch k.Type.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		return []byte(strconv.Quote(k.IntKey.Text(10))), nil
	case T_STRING:
		return []byte(strconv.Quote(k.StringKey)), nil
	case T_BYTES:
		return []byte(strconv.Quote(hex.EncodeToString(k.BytesKey))), nil
	case T_BOOL:
		return []byte(strconv.FormatBool(k.BoolKey)), nil
	case T_TIMESTAMP:
		if y := k.TimeKey.Year(); y < 0 || y >= 10000 {
			return []byte(strconv.Quote(strconv.FormatInt(k.TimeKey.Unix(), 10))), nil
		}
		return []byte(strconv.Quote(k.TimeKey.Format(time.RFC3339))), nil
	case T_KEY_HASH, T_ADDRESS:
		return []byte(strconv.Quote(k.AddrKey.String())), nil
	case T_KEY:
		return []byte(strconv.Quote(k.Hash.String())), nil
	case T_PAIR:
		val := &Value{
			Type:  k.Type,
			Value: k.PrimKey,
		}
		return json.Marshal(val)
	default:
		key, _ := k.Type.MarshalJSON()
		val, _ := k.PrimKey.MarshalJSON()
		return nil, fmt.Errorf("micheline: unsupported big_map key type '%s': typ=%s val=%s",
			k.Type.OpCode, string(key), string(val),
		)
	}
}
