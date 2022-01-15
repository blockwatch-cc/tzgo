// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/tezos"
	"golang.org/x/crypto/blake2b"
)

// Comparable key as used in bigmaps and maps
type Key struct {
	Type Type
	// TODO: refactor into simple Prim
	IntKey       *big.Int
	StringKey    string
	BytesKey     []byte
	BoolKey      bool
	AddrKey      tezos.Address
	KeyKey       tezos.Key
	SignatureKey tezos.Signature
	TimeKey      time.Time
	PrimKey      Prim
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
			// k.Type.OpCode = T_BYTES
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
				if num, err2 := strconv.ParseInt(key.String, 10, 64); err2 == nil {
					t = time.Unix(num, 0)
				} else {
					return Key{}, fmt.Errorf("micheline: invalid big_map key for string timestamp: %w", err)
				}
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
				return Key{}, fmt.Errorf("micheline: invalid big_map key for string type address: %w", err)
			}
			k.AddrKey = a
		} else {
			a := tezos.Address{}
			if err := a.UnmarshalBinary(key.Bytes); err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for type address: %w", err)
			}
			k.AddrKey = a
		}
	case T_KEY:
		if len(key.Bytes) == 0 && len(key.String) > 0 {
			kk, err := tezos.ParseKey(key.String)
			if err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for string type key: %w", err)
			}
			k.KeyKey = kk
		} else {
			kk := tezos.Key{}
			if err := kk.UnmarshalBinary(key.Bytes); err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for type key: %w", err)
			}
			k.KeyKey = kk
		}
	case T_SIGNATURE:
		if len(key.Bytes) == 0 && len(key.String) > 0 {
			sk, err := tezos.ParseSignature(key.String)
			if err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for string type signature: %w", err)
			}
			k.SignatureKey = sk
		} else {
			sk := tezos.Signature{}
			if err := sk.UnmarshalBinary(key.Bytes); err != nil {
				return Key{}, fmt.Errorf("micheline: invalid big_map key for type signature: %w", err)
			}
			k.SignatureKey = sk
		}
	case T_PAIR, T_OPTION, T_OR, T_CHAIN_ID, T_UNIT, T_OPERATION:
		k.PrimKey = key
		// build type details when missing
		if len(k.Type.Args) == 0 {
			k.Type = key.BuildType()
		}

	default:
		k.PrimKey = key
		// build type details when missing
		if len(k.Type.Args) == 0 {
			k.Type = key.BuildType()
		}
		return k, fmt.Errorf("micheline: big_map key type '%s' is not implemented", typ.OpCode)
	}
	return k, nil
}

func NewKeyPtr(typ Type, key Prim) (*Key, error) {
	k, err := NewKey(typ, key)
	return &k, err
}

func (k Key) IsPacked() bool {
	return k.Type.OpCode == T_BYTES && (isPackedBytes(k.BytesKey) ||
		tezos.IsAddressBytes(k.BytesKey) ||
		isASCIIBytes(k.BytesKey))
}

func (k Key) UnpackPrim() (p Prim, err error) {
	return Prim{
		Type:  PrimBytes,
		Bytes: k.BytesKey,
	}.Unpack()
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
	case T_INT, T_NAT, T_MUTEZ, T_STRING, T_BYTES, T_BOOL,
		T_KEY_HASH, T_TIMESTAMP, T_ADDRESS, T_PAIR, T_KEY, T_SIGNATURE,
		T_OPTION, T_OR, T_CHAIN_ID, T_UNIT:
		return t, nil
	default:
		return t, fmt.Errorf("micheline: unsupported big_map key type %s", t)
	}
}

// query string parsing used for lookup
func ParseKey(typ OpCode, val string) (Key, error) {
	key := Key{Type: Type{}}
	if typ.IsTypeCode() {
		key.Type.OpCode = typ
	} else {
		key.Type.OpCode = InferKeyType(val)
	}
	var err error
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
	case T_KEY:
		key.Type.Type = PrimBytes
		key.KeyKey, err = tezos.ParseKey(val)
	case T_SIGNATURE:
		key.Type.Type = PrimBytes
		key.SignatureKey, err = tezos.ParseSignature(val)
	case T_PAIR:
		// parse comma-separated list into a right-hand pair tree
		prims := []Prim{}
		for _, v := range strings.Split(val, ",") {
			parsed, err := ParseKey(InferKeyType(v), v)
			if err != nil {
				return Key{}, fmt.Errorf("micheline: decoding bigmap pair key element %s: %w", v, err)
			}
			prims = append(prims, parsed.Prim())
		}
		switch len(prims) {
		case 0:
			return Key{}, fmt.Errorf("micheline: empty bigmap pair key: %s", val)
		case 1:
			return Key{}, fmt.Errorf("micheline: single-value bigmap pair key: %s", val)
		default:
			key.PrimKey = NewSeq(prims...).FoldPair()
			key.Type.Type = PrimBinary
		}
	case T_UNIT:
		if val != D_UNIT.String() {
			return Key{}, fmt.Errorf("micheline: invalid bigmap pair key for Unit type: %s", val)
		}
		key.PrimKey = NewCode(D_UNIT)
		key.Type.Type = PrimNullary

	default:
		return Key{}, fmt.Errorf("micheline: unsupported big_map key type %s", typ)
	}

	if err != nil {
		return Key{}, fmt.Errorf("micheline: decoding bigmap key %s as %s: %w", val, typ, err)
	}
	return key, nil
}

func InferKeyType(val string) OpCode {
	if val == D_UNIT.String() {
		return T_UNIT
	}
	if _, err := tezos.ParseAddress(val); err == nil {
		// Note: can also be KEY_HASH, but used inconsistently
		return T_ADDRESS
	}
	if _, err := tezos.ParseKey(val); err == nil {
		return T_KEY
	}
	if _, err := tezos.ParseSignature(val); err == nil {
		return T_SIGNATURE
	}
	if _, err := time.Parse(time.RFC3339, val); err == nil {
		return T_TIMESTAMP
	}
	i := big.NewInt(0)
	if err := i.UnmarshalText([]byte(val)); err == nil {
		// can also be T_MUTEZ, T_NAT
		return T_INT
	}
	if _, err := hex.DecodeString(val); err == nil {
		return T_BYTES
	}
	if strings.Contains(val, ",") {
		return T_PAIR
	}
	if _, err := strconv.ParseBool(val); err == nil {
		return T_BOOL
	}
	return T_STRING
}

func DecodeKey(typ Type, b []byte) (Key, error) {
	key := Prim{}
	if err := key.UnmarshalBinary(b); err != nil {
		return Key{}, err
	}
	return NewKey(typ, key)
}

func (k Key) Bytes() []byte {
	p := Prim{}
	switch k.Type.OpCode {
	case T_INT, T_NAT, T_MUTEZ:
		p.Type = PrimInt
		p.Int = k.IntKey
	case T_STRING:
		p.Type = PrimString
		p.String = k.StringKey
	case T_BYTES:
		p.Type = PrimBytes
		p.Bytes = k.BytesKey
	case T_BOOL:
		p.Type = PrimNullary
		if k.BoolKey {
			p.OpCode = D_TRUE
		} else {
			p.OpCode = D_FALSE
		}
	case T_TIMESTAMP:
		var z tezos.Z
		z.SetInt64(k.TimeKey.Unix())
		p.Type = PrimInt
		p.Int = z.Big()
	case T_ADDRESS:
		p.Type = PrimBytes
		p.Bytes, _ = k.AddrKey.MarshalBinary()
	case T_KEY_HASH:
		p.Type = PrimBytes
		b, _ := k.AddrKey.MarshalBinary()
		p.Bytes = b[1:] // strip leading flag
	case T_KEY:
		p.Type = PrimBytes
		p.Bytes, _ = k.KeyKey.MarshalBinary()
	case T_SIGNATURE:
		p.Type = PrimBytes
		p.Bytes, _ = k.SignatureKey.MarshalBinary()
	default:
		// anthing else comes from prim tree
		if !k.PrimKey.IsValid() {
			return nil
		}
		b, _ := k.PrimKey.MarshalBinary()
		return b
	}
	buf, _ := p.MarshalBinary()
	return buf
}

func (k Key) MarshalBinary() ([]byte, error) {
	return k.Bytes(), nil
}

func (k Key) Hash() tezos.ExprHash {
	return KeyHash(k.Bytes())
}

func KeyHash(buf []byte) tezos.ExprHash {
	// blake2b with digest size 32 byte
	h, _ := blake2b.New(32, nil)

	// encode with pack byte
	h.Write([]byte{0x5})
	h.Write(buf)

	// wrap in exprhash
	return tezos.NewExprHash(h.Sum(nil))
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
		return k.KeyKey.String()
	case T_SIGNATURE:
		return k.SignatureKey.String()
	case T_PAIR:
		type lv struct {
			l string
			v string
		}
		parts := make([]lv, 0)

		val := Value{
			Type:  k.Type, // real or guessed type tree
			Value: k.PrimKey,
		}

		// walk produces a non-deterministic order
		val.Walk("", func(label string, v interface{}) error {
			part := lv{l: label}
			if stringer, ok := v.(fmt.Stringer); ok {
				part.v = stringer.String()
			} else {
				if str, ok := v.(string); ok {
					part.v = str
				} else {
					part.v = fmt.Sprint(v)
				}
			}
			parts = append(parts, part)
			return nil
		})

		// sort by label (works up to 10 pair values)
		sort.Slice(parts, func(i, j int) bool { return parts[i].l < parts[j].l })

		var b strings.Builder
		for i, v := range parts {
			if i > 0 {
				b.WriteRune(',')
			}
			b.WriteString(v.v)
		}
		return b.String()

		// TODO: simpler, but requires value tree decoration
		// return fmt.Sprint(k.PrimKey.Value(T_PAIR)

	case T_UNIT:
		return D_UNIT.String()
	default:
		if k.PrimKey.IsValid() {
			return k.PrimKey.OpCode.String()
		}
		return ""
	}
}

func (k Key) Prim() Prim {
	p := Prim{}
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
	case T_ADDRESS, T_KEY_HASH:
		p.Bytes, _ = k.AddrKey.MarshalBinary()
		p.Type = PrimBytes
	case T_KEY:
		p.Bytes, _ = k.KeyKey.MarshalBinary()
		p.Type = PrimBytes
	case T_SIGNATURE:
		p.Bytes, _ = k.SignatureKey.MarshalBinary()
		p.Type = PrimBytes
	case T_BOOL:
		p.Type = PrimNullary
		if k.BoolKey {
			p.OpCode = D_TRUE
		} else {
			p.OpCode = D_FALSE
		}
	case T_PAIR, T_UNIT:
		p = k.PrimKey
	default:
		if k.PrimKey.IsValid() {
			p = k.PrimKey
			break
		}
		if k.BytesKey != nil {
			if err := p.UnmarshalBinary(k.BytesKey); err == nil {
				break
			}
		}
		p.Bytes = k.Bytes()
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
		return []byte(strconv.Quote(k.KeyKey.String())), nil
	case T_SIGNATURE:
		return []byte(strconv.Quote(k.SignatureKey.String())), nil
	default:
		val := &Value{
			Type:  k.Type,
			Value: k.PrimKey,
		}
		return json.Marshal(val)
	}
}
