// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Domain specific data types
//
// see http://tezos.gitlab.io/whitedoc/michelson.html#full-grammar
//
// - timestamp: Dates in the real world.
// - mutez: A specific type for manipulating tokens.
// - address: An untyped address (implicit account or smart contract).
// - contract 'param: A contract, with the type of its code,
//   contract unit for implicit accounts.
// - operation: An internal operation emitted by a contract.
// - key: A public cryptographic key.
// - key_hash: The hash of a public cryptographic key.
// - signature: A cryptographic signature.
// - chain_id: An identifier for a chain, used to distinguish the test
//   and the main chains.
//
// PACK prefixes with 0x05!
// So that when contracts checking signatures (multisigs etc) do the current
// best practice, PACK; ...; CHECK_SIGNATURE, the 0x05 byte distinguishes the
// message from blocks, endorsements, transactions, or tezos-signer authorization
// requests (0x01-0x04)

package micheline

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

var (
	InvalidPrim = Prim{}
	EmptyPrim   = Prim{Type: PrimNullary, OpCode: 255}
)

type PrimType byte

const (
	PrimInt          PrimType = iota // 00 {name: 'int'}
	PrimString                       // 01 {name: 'string'}
	PrimSequence                     // 02 []
	PrimNullary                      // 03 {name: 'prim', len: 0, annots: false},
	PrimNullaryAnno                  // 04 {name: 'prim', len: 0, annots: true},
	PrimUnary                        // 05 {name: 'prim', len: 1, annots: false},
	PrimUnaryAnno                    // 06 {name: 'prim', len: 1, annots: true},
	PrimBinary                       // 07 {name: 'prim', len: 2, annots: false},
	PrimBinaryAnno                   // 08 {name: 'prim', len: 2, annots: true},
	PrimVariadicAnno                 // 09 {name: 'prim', len: n, annots: true},
	PrimBytes                        // 0A {name: 'bytes' }
)

func (t PrimType) TypeCode() OpCode {
	switch t {
	case PrimInt:
		return T_INT
	case PrimString:
		return T_STRING
	default:
		return T_BYTES
	}
}

func (t PrimType) IsValid() bool {
	return t <= PrimBytes
}

func ParsePrimType(val string) (PrimType, error) {
	switch val {
	case "int":
		return PrimInt, nil
	case "string":
		return PrimString, nil
	case "bytes":
		return PrimBytes, nil
	default:
		return 0, fmt.Errorf("micheline: invalid prim type '%s'", val)
	}
}

// non-normative strings, use for debugging only
func (t PrimType) String() string {
	switch t {
	case PrimInt:
		return "int"
	case PrimString:
		return "string"
	case PrimSequence:
		return "sequence"
	case PrimNullary:
		return "prim"
	case PrimNullaryAnno:
		return "prim+"
	case PrimUnary:
		return "unary"
	case PrimUnaryAnno:
		return "unary+"
	case PrimBinary:
		return "binary"
	case PrimBinaryAnno:
		return "binary+"
	case PrimVariadicAnno:
		return "variadic"
	case PrimBytes:
		return "bytes"
	default:
		return "invalid"
	}
}

func (t PrimType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type Prim struct {
	Type      PrimType // primitive type
	OpCode    OpCode   // primitive opcode (invalid on sequences, strings, bytes, int)
	Args      []Prim   // optional arguments
	Anno      []string // optional type annotations
	Int       *big.Int // optional data
	String    string   // optional data
	Bytes     []byte   // optional data
	WasPacked bool     // true when content was unpacked (and no type info is available)
}

func (p Prim) IsValid() bool {
	return p.Type.IsValid() && (p.Type > 0 || p.Int != nil)
}

func (p Prim) Clone() Prim {
	clone := Prim{
		Type:      p.Type,
		OpCode:    p.OpCode,
		String:    p.String,
		WasPacked: p.WasPacked,
	}
	if p.Args != nil {
		clone.Args = make([]Prim, len(p.Args))
		for i, arg := range p.Args {
			clone.Args[i] = arg.Clone()
		}
	}
	if p.Anno != nil {
		clone.Anno = make([]string, len(p.Anno))
		for i, anno := range p.Anno {
			clone.Anno[i] = anno
		}
	}
	if p.Int != nil {
		clone.Int = big.NewInt(0)
		clone.Int.Set(p.Int)
	}
	if p.Bytes != nil {
		clone.Bytes = make([]byte, len(p.Bytes))
		copy(clone.Bytes, p.Bytes)
	}
	return clone
}

func (p Prim) IsEqual(p2 Prim) bool {
	return IsEqualPrim(p, p2, false)
}

func (p Prim) IsEqualWithAnno(p2 Prim) bool {
	return IsEqualPrim(p, p2, true)
}

func IsEqualPrim(p1, p2 Prim, withAnno bool) bool {
	// opcode
	if p1.OpCode != p2.OpCode {
		fmt.Printf("Opcode mismatch %s <> %s\n", p1.OpCode, p2.OpCode)
		return false
	}

	// type
	t1, t2 := p1.Type, p2.Type
	if !withAnno {
		switch t1 {
		case PrimNullaryAnno, PrimUnaryAnno, PrimBinaryAnno:
			t1--
		}
		switch t2 {
		case PrimNullaryAnno, PrimUnaryAnno, PrimBinaryAnno:
			t2--
		}
	}
	if t1 != t2 {
		fmt.Println("Prim type mismatch")
		return false
	}

	// arg len
	if len(p1.Args) != len(p2.Args) {
		fmt.Println("Arg len mismatch")
		return false
	}

	// anno
	if withAnno {
		if len(p1.Anno) != len(p2.Anno) {
			fmt.Println("Anno len mismatch")
			return false
		}
		for i := range p1.Anno {
			if p1.Anno[i] != p2.Anno[i] {
				fmt.Println("Anno mismatch")
				return false
			}
		}
	}

	// contents
	if p1.String != p2.String {
		fmt.Println("String content mismatch")
		return false
	}
	if (p1.Int == nil) != (p2.Int == nil) {
		fmt.Println("Int ptr content mismatch")
		return false
	}
	if p1.Int != nil {
		if p1.Int.Cmp(p2.Int) != 0 {
			fmt.Println("Int content mismatch")
			return false
		}
	}
	if (p1.Bytes == nil) != (p2.Bytes == nil) {
		fmt.Println("Bytes ptr content mismatch")
		return false
	}
	if p1.Bytes != nil {
		if bytes.Compare(p1.Bytes, p2.Bytes) != 0 {
			fmt.Println("Bytes content mismatch")
			return false
		}
	}

	// recurse
	for i := range p1.Args {
		if !IsEqualPrim(p1.Args[i], p2.Args[i], withAnno) {
			fmt.Println("Nested content mismatch")
			return false
		}
	}

	// all equal
	return true
}

type PrimWalkerFunc func(p Prim) error

func (p Prim) Walk(f PrimWalkerFunc) error {
	if err := f(p); err != nil {
		return err
	}
	for _, v := range p.Args {
		if err := v.Walk(f); err != nil {
			return err
		}
	}
	return nil
}

// returns true when the prim can be expressed as a single value
// key/value pairs (ie. prims with annots) do not fit into this category
// used when mapping complex big map values to JSON objects
func (p Prim) IsScalar() bool {
	switch p.Type {
	case PrimInt, PrimString, PrimBytes, PrimNullary:
		// generally ok
		return true
	case PrimSequence:
		return len(p.Args) == 1 && !p.HasAnno()
	case PrimNullaryAnno, PrimUnaryAnno, PrimBinaryAnno, PrimVariadicAnno:
		// all annotated types become JSON properties
		return false
	case PrimUnary:
		switch p.OpCode {
		case D_LEFT, D_RIGHT, D_SOME:
			return p.Args[0].IsScalar()
		}
		return false
	case PrimBinary:
		// mostly not ok, unless type is option/or and sub-types are scalar
		switch p.OpCode {
		case T_OPTION, T_OR:
			return p.Args[0].IsScalar()
		}
		return false
	}
	return false
}

func (p Prim) IsSequence() bool {
	return p.Type == PrimSequence
}

func (p Prim) IsTicket() bool {
	return p.OpCode == T_TICKET
}

func (p Prim) IsSet() bool {
	return p.OpCode == T_SET
}

func (p Prim) IsList() bool {
	return p.OpCode == T_LIST
}

func (p Prim) IsMap() bool {
	return p.OpCode == T_MAP
}

func (p Prim) IsLambda() bool {
	return p.OpCode == T_LAMBDA
}

func (p Prim) IsElt() bool {
	return p.OpCode == D_ELT
}

func (p Prim) IsPair() bool {
	switch p.OpCode {
	case T_PAIR, D_PAIR:
		return true
	default:
		return false
	}
}

// Detects whether a primitive contains a regular pair or a comb pair
// which may be converted into a flat comb sequence.
//
// The Michelson spec says
// For combs, three notations are supported:
//  - a) [Pair x1 (Pair x2 ... (Pair xn-1 xn) ...)],
//  - b) [Pair x1 x2 ... xn-1 xn], and
//  - c) [{x1; x2; ...; xn-1; xn}].
//  In readable mode, we always use b),
//  in optimized mode we use the shortest to serialize:
//  - for n=2, [Pair x1 x2],
//  - for n=3, [Pair x1 (Pair x2 x3)],
//  - for n>=4, [{x1; x2; ...; xn}].
func (p Prim) IsComb(typ Type) bool {
	// regular pairs (works for type and value trees)
	if p.IsPair() {
		return true
	}

	// Note: due to Tezos encoding issues, comb pairs in block receipts
	// are naked sequences (they lack an enclosing pair wrapper), this means
	// we have to distinquish them from other container types who also use
	// a sequence as container, such as lists, sets, maps, lambdas
	switch typ.OpCode {
	case T_LIST, T_SET, T_MAP, T_LAMBDA:
		return false
	}

	// When multiple optimized combs (i.e. naked sequences) are nested
	// it is sometimes not possible to track the correct type in the
	// type tree. Hence we revert to a heuristic that checks if the
	// current primitive looks like a container type by checking its
	// contents
	if p.Type == PrimSequence && !p.LooksLikeContainer() {
		return true
	}

	return false
}

// Checks if a Prim looks like an optimized (i.e. flat) comb sequence.
func (p Prim) IsConvertedComb() bool {
	if !p.IsSequence() {
		return false
	}
	for _, v := range p.Args {
		if v.IsPair() {
			return false
		}
	}
	return true
}

// Checks if a Prim looks like a container type. This is necessary to
// distinguish optimized comb pairs from other container types.
func (p Prim) LooksLikeContainer() bool {
	// must be a sequence
	if !p.IsSequence() {
		return false
	}

	// empty container
	if len(p.Args) == 0 {
		return true
	}

	// contains Elt's
	if p.Args[0].IsElt() {
		return true
	}

	// contains a single scalar
	if len(p.Args) == 1 {
		switch p.Args[0].Type {
		case PrimInt, PrimString, PrimBytes, PrimNullary, PrimSequence:
			return true
		}
	}

	// all elements have the same prim type and opcode
	oc := p.Args[0].OpCode
	typ := p.Args[0].Type
	for _, v := range p.Args[1:] {
		if v.OpCode != oc || v.Type != typ {
			return false
		}
	}

	return true
}

// Converts a pair tree into a flattened sequence. While Michelson
// optimized comb pairs are only used for right-side combs, this
// function applies all pairs. It makes use of the type definition
// to identify which contained type is a regular pair, an already
// converted comb pair or any other container type.
//
// - Works both on value trees and type trees.
// - Will skip (i.e. reduce) annotated type pairs, so nested struct
//   information will be lost.
// - When called on already converted comb sequences this function is a noop.
//
func (p Prim) ConvertComb(typ Type) []Prim {
	flat := make([]Prim, 0)
	for i, v := range p.Args {
		t := Type{}
		if len(typ.Args) > i {
			t = Type{typ.Args[i]}
		}
		if v.IsComb(t) {
			flat = append(flat, v.ConvertComb(t)...)
		} else {
			flat = append(flat, v)
		}
	}
	return flat
}

func (p Prim) Text() string {
	switch p.Type {
	case PrimInt:
		return p.Int.Text(10)
	case PrimString:
		return p.String
	case PrimBytes:
		return hex.EncodeToString(p.Bytes)
	default:
		v, _ := p.Value(p.OpCode).(string)
		return v
	}
}

func (p Prim) IsPacked() bool {
	return (p.OpCode == T_BYTES || p.Type == PrimBytes) && len(p.Bytes) > 1 && p.Bytes[0] == 0x5
}

func (p Prim) PackedType() PrimType {
	if !p.IsPacked() {
		return PrimNullary
	}
	return PrimType(p.Bytes[1])
}

func (p Prim) Unpack() (pp Prim, err error) {
	if !p.IsPacked() {
		return p, fmt.Errorf("prim is not packed")
	}
	defer func() {
		if e := recover(); e != nil {
			pp = p
			err = fmt.Errorf("prim is not packed")
		}
	}()
	pp = Prim{WasPacked: true}
	if err := pp.UnmarshalBinary(p.Bytes[1:]); err != nil {
		return p, err
	}
	if pp.IsPackedAny() {
		if up, err := pp.UnpackAll(); err == nil {
			pp = up
		}
	}
	return pp, nil
}

func (p Prim) IsPackedAny() bool {
	if p.IsPacked() {
		return true
	}
	for _, v := range p.Args {
		if v.IsPackedAny() {
			return true
		}
	}
	return false
}

func (p Prim) UnpackAll() (Prim, error) {
	if p.IsPacked() {
		return p.Unpack()
	}
	pp := p
	pp.Args = make([]Prim, len(p.Args))
	for i, v := range p.Args {
		if v.IsPackedAny() {
			if up, err := v.UnpackAll(); err == nil {
				pp.Args[i] = up
			}
			continue
		}
		pp.Args[i] = v
	}
	return pp, nil
}

func (p Prim) Value(as OpCode) interface{} {
	var warn bool
	switch p.Type {
	case PrimInt:
		switch as {
		case T_TIMESTAMP:
			tm := time.Unix(p.Int.Int64(), 0).UTC()
			if y := tm.Year(); y < 0 || y >= 10000 {
				return p.Int.Text(10)
			}
			return tm
		default:
			return p.Int.Text(10)
		}

	case PrimString:
		switch as {
		case T_TIMESTAMP:
			if t, err := time.Parse(time.RFC3339, p.String); err == nil {
				return t
			}
			return p.String
		default:
			return p.String
		}

	case PrimBytes:
		switch as {
		case T_BYTES:
			// try unpack
			if p.IsPacked() {
				if up, err := p.Unpack(); err == nil {
					return up.Value(as)
				}
			}
			// try address
			a := tezos.Address{}
			if err := a.UnmarshalBinary(p.Bytes); err == nil {
				return a
			}
			// try ascii string
			if s := string(p.Bytes); isASCII(s) {
				return s
			}

		case T_KEY_HASH, T_ADDRESS, T_CONTRACT:
			a := tezos.Address{}
			if err := a.UnmarshalBinary(p.Bytes); err == nil {
				return a
			} else {
				log.Errorf("Rendering prim type %s as %s: %v (%s)", p.Type, as, err, hex.EncodeToString(p.Bytes))
			}
		case T_KEY:
			k := tezos.Key{}
			if err := k.UnmarshalBinary(p.Bytes); err == nil {
				return k
			} else {
				log.Errorf("Rendering prim type %s as %s: %v (%s)", p.Type, as, err, hex.EncodeToString(p.Bytes))
			}

		case T_SIGNATURE:
			s := tezos.Signature{}
			if err := s.UnmarshalBinary(p.Bytes); err == nil {
				return s
			} else {
				log.Errorf("Rendering prim type %s as %s: %v (%s)", p.Type, as, err, p.Bytes, hex.EncodeToString(p.Bytes))
			}

		case T_CHAIN_ID:
			if len(p.Bytes) == tezos.HashTypeChainId.Len() {
				return tezos.NewChainIdHash(p.Bytes).String()
			}

		case T_BLS12_381_G1, T_BLS12_381_G2, T_BLS12_381_FR, T_SAPLING_STATE:
			// as hex, fallthrough

		default:
			// case T_LAMBDA:
			// case T_LIST, T_MAP, T_BIG_MAP, T_SET:
			// case T_OPTION, T_OR, T_PAIR, T_UNIT:
			// case T_OPERATION:
			warn = true
		}

		// default is to render bytes as hex string
		return hex.EncodeToString(p.Bytes)

	case PrimUnary, PrimUnaryAnno:
		switch as {
		case T_OR, T_OPTION:
			// expected, just render as prim tree
		default:
			warn = true
		}

	case PrimNullary, PrimNullaryAnno:
		switch p.OpCode {
		case D_FALSE:
			return false
		case D_TRUE:
			return true
		case D_UNIT:
			return nil
		default:
			return p.OpCode.String()
		}

	case PrimBinary, PrimBinaryAnno:
		switch p.OpCode {
		case D_PAIR, T_PAIR:
			// mangle pair contents into string, used when rendering complex keys
			left := p.Args[0].Value(p.Args[0].OpCode)
			if _, ok := left.(fmt.Stringer); !ok {
				if _, ok := left.(string); !ok {
					left = p.Args[0].OpCode.String()
				}
			}
			right := p.Args[1].Value(p.Args[1].OpCode)
			if _, ok := right.(fmt.Stringer); !ok {
				if _, ok := right.(string); !ok {
					right = p.Args[1].OpCode.String()
				}
			}
			return fmt.Sprintf("%s,%s", left, right)
		}

	case PrimSequence:
		switch p.OpCode {
		case D_PAIR, T_PAIR:
			var b strings.Builder
			for i, v := range p.Args {
				if i > 0 {
					b.WriteByte(',')
				}
				val := v.Value(v.OpCode)
				if stringer, ok := val.(fmt.Stringer); !ok {
					if str, ok := val.(string); !ok {
						b.WriteString(v.OpCode.String())
					} else {
						b.WriteString(str)
					}
				} else {
					b.WriteString(stringer.String())
				}
			}
			return b.String()

		default:
			switch as {
			case T_UNIT, T_LAMBDA, T_LIST, T_MAP, T_BIG_MAP, T_SET, T_SAPLING_STATE:
				return p
			default:
				warn = true
			}
		}

	default:
		switch as {
		case T_BOOL:
			if p.OpCode == D_TRUE {
				return true
			} else if p.OpCode == D_FALSE {
				return false
			}
		case T_OPERATION:
			return p.OpCode.String()
		case T_BYTES:
			return hex.EncodeToString(p.Bytes)
		default:
			warn = true
		}
	}

	if warn && !p.WasPacked {
		buf, _ := json.Marshal(p)
		log.Warnf("Rendering prim type %s as %s: not implemented (%s)", p.Type, as, string(buf))
	}

	return p
}

func (p Prim) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	switch p.Type {
	case PrimSequence:
		return json.Marshal(p.Args)
	case PrimInt:
		m["int"] = p.Int.Text(10)
	case PrimString:
		m["string"] = p.String
	case PrimBytes:
		m["bytes"] = hex.EncodeToString(p.Bytes)
	default:
		m["prim"] = p.OpCode.String()
		if len(p.Anno) > 0 {
			m["annots"] = p.Anno
		}
		if len(p.Args) > 0 {
			args := make([]json.RawMessage, 0, len(p.Args))
			for _, v := range p.Args {
				arg, err := json.Marshal(v)
				if err != nil {
					return nil, err
				}
				args = append(args, arg)
			}
			m["args"] = args
		}
	}
	return json.Marshal(m)
}

func (p Prim) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := p.EncodeBuffer(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p Prim) EncodeBuffer(buf *bytes.Buffer) error {
	buf.WriteByte(byte(p.Type))
	switch p.Type {
	case PrimInt:
		var z Z
		z.Set(p.Int)
		if err := z.EncodeBuffer(buf); err != nil {
			return err
		}

	case PrimString:
		binary.Write(buf, binary.BigEndian, uint32(len(p.String)))
		buf.WriteString(p.String)

	case PrimSequence:
		seq := bytes.NewBuffer(nil)
		binary.Write(seq, binary.BigEndian, uint32(0))
		for _, v := range p.Args {
			if err := v.EncodeBuffer(seq); err != nil {
				return err
			}
		}
		res := seq.Bytes()
		binary.BigEndian.PutUint32(res[:], uint32(len(res)-4))
		buf.Write(res)

	case PrimNullary:
		buf.WriteByte(byte(p.OpCode))

	case PrimNullaryAnno:
		buf.WriteByte(byte(p.OpCode))
		anno := strings.Join(p.Anno, " ")
		binary.Write(buf, binary.BigEndian, uint32(len(anno)))
		buf.WriteString(anno)

	case PrimUnary:
		buf.WriteByte(byte(p.OpCode))
		for _, v := range p.Args {
			if err := v.EncodeBuffer(buf); err != nil {
				return err
			}
		}

	case PrimUnaryAnno:
		buf.WriteByte(byte(p.OpCode))
		for _, v := range p.Args {
			if err := v.EncodeBuffer(buf); err != nil {
				return err
			}
		}
		anno := strings.Join(p.Anno, " ")
		binary.Write(buf, binary.BigEndian, uint32(len(anno)))
		buf.WriteString(anno)

	case PrimBinary:
		buf.WriteByte(byte(p.OpCode))
		for _, v := range p.Args {
			if err := v.EncodeBuffer(buf); err != nil {
				return err
			}
		}

	case PrimBinaryAnno:
		buf.WriteByte(byte(p.OpCode))
		for _, v := range p.Args {
			if err := v.EncodeBuffer(buf); err != nil {
				return err
			}
		}
		anno := strings.Join(p.Anno, " ")
		binary.Write(buf, binary.BigEndian, uint32(len(anno)))
		buf.WriteString(anno)

	case PrimVariadicAnno:
		buf.WriteByte(byte(p.OpCode))

		seq := bytes.NewBuffer(nil)
		binary.Write(seq, binary.BigEndian, uint32(0))
		for _, v := range p.Args {
			if err := v.EncodeBuffer(seq); err != nil {
				return err
			}
		}
		res := seq.Bytes()
		binary.BigEndian.PutUint32(res[:], uint32(len(res)-4))
		buf.Write(res)

		anno := strings.Join(p.Anno, " ")
		binary.Write(buf, binary.BigEndian, uint32(len(anno)))
		buf.WriteString(anno)

	case PrimBytes:
		binary.Write(buf, binary.BigEndian, uint32(len(p.Bytes)))
		buf.Write(p.Bytes)
	}

	return nil
}

func (p *Prim) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	switch data[0] {
	case '[':
		var m []interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		return p.UnpackSequence(m)
	case '{':
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		return p.UnpackJSON(m)
	default:
		var m interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		return p.UnpackScalar(m)
	}
}

func (p *Prim) UnpackJSON(val interface{}) error {
	switch t := val.(type) {
	case map[string]interface{}:
		return p.UnpackPrimitive(t)
	case []interface{}:
		return p.UnpackSequence(t)
	default:
		return fmt.Errorf("micheline: unexpected json type %T", val)
	}
}

func (p *Prim) UnpackScalar(val interface{}) error {
	switch v := val.(type) {
	case json.Number:
		i := big.NewInt(0)
		if err := i.UnmarshalText([]byte(v.String())); err != nil {
			return err
		}
		p.Int = i
		p.Type = PrimInt
	default:
		oc, err := ParseOpCode(val.(string))
		if err == nil {
			p.OpCode = oc
			p.Type = PrimNullary
		} else {
			// fallback (should not happen)
			p.OpCode = T_STRING
			p.String = val.(string)
			p.Type = PrimString
		}
	}
	return nil
}

func (p *Prim) UnpackSequence(val []interface{}) error {
	p.Type = PrimSequence
	p.Args = make([]Prim, 0)
	for _, v := range val {
		prim := Prim{}
		if err := prim.UnpackJSON(v); err != nil {
			return err
		}
		p.Args = append(p.Args, prim)
	}
	return nil
}

func (p *Prim) UnpackPrimitive(val map[string]interface{}) error {
	p.Args = make([]Prim, 0)
	for n, v := range val {
		switch n {
		case "prim":
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("micheline: invalid prim value type %T %v", v, v)
			}
			oc, err := ParseOpCode(str)
			if err != nil {
				return err
			}
			p.OpCode = oc
			p.Type = PrimNullary
		case "int":
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("micheline: invalid int value type %T %v", v, v)
			}
			i := big.NewInt(0)
			if err := i.UnmarshalText([]byte(str)); err != nil {
				return err
			}
			p.Int = i
			p.Type = PrimInt
		case "string":
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("micheline: invalid string value type %T %v", v, v)
			}
			p.String = str
			p.Type = PrimString
		case "bytes":
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("micheline: invalid bytes value type %T %v", v, v)
			}
			b, err := hex.DecodeString(str)
			if err != nil {
				return err
			}
			p.Bytes = b
			p.Type = PrimBytes
		case "annots":
			slist, ok := v.([]interface{})
			if !ok {
				return fmt.Errorf("micheline: invalid annots value type %T %v", v, v)
			}
			for _, s := range slist {
				p.Anno = append(p.Anno, s.(string))
			}
		}
	}

	// update type when annots are present, but no more args are defined
	if len(p.Anno) > 0 && p.Type == PrimNullary {
		p.Type = PrimNullaryAnno
	}

	// process args separately and detect type based on number of args
	if a, ok := val["args"]; ok {
		args, ok := a.([]interface{})
		if !ok {
			return fmt.Errorf("micheline: invalid args value type %T %v", a, a)
		}

		switch len(args) {
		case 0:
			p.Type = PrimNullary
		case 1:
			if len(p.Anno) > 0 {
				p.Type = PrimUnaryAnno
			} else {
				p.Type = PrimUnary
			}
		case 2:
			if len(p.Anno) > 0 {
				p.Type = PrimBinaryAnno
			} else {
				p.Type = PrimBinary
			}
		default:
			p.Type = PrimVariadicAnno
		}

		// every arg is handled as embedded primitive
		for _, v := range args {
			prim := Prim{}
			if err := prim.UnpackJSON(v); err != nil {
				return err
			}
			p.Args = append(p.Args, prim)
		}
	}
	return nil
}

func (p *Prim) UnmarshalBinary(data []byte) error {
	return p.DecodeBuffer(bytes.NewBuffer(data))
}

func (p *Prim) DecodeBuffer(buf *bytes.Buffer) error {
	b := buf.Next(1)
	if len(b) == 0 {
		return io.ErrShortBuffer
	}
	tag := PrimType(b[0])
	switch tag {
	case PrimInt:
		// data is a zarith number
		var z Z
		if err := z.DecodeBuffer(buf); err != nil {
			return err
		}
		p.Int = z.Big()

	case PrimString:
		// cross-check content size
		size := int(binary.BigEndian.Uint32(buf.Next(4)))
		if buf.Len() < size {
			return io.ErrShortBuffer
		}
		p.String = string(buf.Next(size))

	case PrimSequence:
		// cross-check content size
		size := int(binary.BigEndian.Uint32(buf.Next(4)))
		if buf.Len() < size {
			return io.ErrShortBuffer
		}
		// extract sub-buffer
		seq := bytes.NewBuffer(buf.Next(size))
		// decode contained primitives
		p.Args = make([]Prim, 0)
		for seq.Len() > 0 {
			prim := Prim{}
			if err := prim.DecodeBuffer(seq); err != nil {
				return err
			}
			p.Args = append(p.Args, prim)
		}

	case PrimNullary:
		// opcode only
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		p.OpCode = OpCode(b[0])

	case PrimNullaryAnno:
		// opcode with annotations
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		p.OpCode = OpCode(b[0])

		// annotation array byte size
		size := int(binary.BigEndian.Uint32(buf.Next(4)))
		if buf.Len() < size {
			return io.ErrShortBuffer
		}
		anno := buf.Next(size)
		p.Anno = strings.Split(string(anno), " ")

	case PrimUnary:
		// opcode with single argument
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		p.OpCode = OpCode(b[0])

		// argument
		prim := Prim{}
		if err := prim.DecodeBuffer(buf); err != nil {
			return err
		}
		p.Args = append(p.Args, prim)

	case PrimUnaryAnno:
		// opcode with single argument and annotations
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		p.OpCode = OpCode(b[0])

		// argument
		prim := Prim{}
		if err := prim.DecodeBuffer(buf); err != nil {
			return err
		}
		p.Args = append(p.Args, prim)

		// annotation array byte size
		size := int(binary.BigEndian.Uint32(buf.Next(4)))
		if buf.Len() < size {
			return io.ErrShortBuffer
		}
		anno := buf.Next(size)
		p.Anno = strings.Split(string(anno), " ")

	case PrimBinary:
		// opcode with two arguments
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		p.OpCode = OpCode(b[0])

		// 2 arguments
		for i := 0; i < 2; i++ {
			prim := Prim{}
			if err := prim.DecodeBuffer(buf); err != nil {
				return err
			}
			p.Args = append(p.Args, prim)
		}

	case PrimBinaryAnno:
		// opcode with two arguments and annotations
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		p.OpCode = OpCode(b[0])

		// 2 arguments
		for i := 0; i < 2; i++ {
			prim := Prim{}
			if err := prim.DecodeBuffer(buf); err != nil {
				return err
			}
			p.Args = append(p.Args, prim)
		}

		// annotation array byte size
		size := int(binary.BigEndian.Uint32(buf.Next(4)))
		if buf.Len() < size {
			return io.ErrShortBuffer
		}
		anno := buf.Next(size)
		p.Anno = strings.Split(string(anno), " ")

	case PrimVariadicAnno:
		// opcode with N arguments and optional annotations
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		p.OpCode = OpCode(b[0])

		// argument array byte size
		size := int(binary.BigEndian.Uint32(buf.Next(4)))

		// extract sub-buffer
		seq := bytes.NewBuffer(buf.Next(size))

		// decode contained primitives
		for seq.Len() > 0 {
			prim := Prim{}
			if err := prim.DecodeBuffer(seq); err != nil {
				return err
			}
			p.Args = append(p.Args, prim)
		}
		// annotation array byte size
		size = int(binary.BigEndian.Uint32(buf.Next(4)))
		if buf.Len() < size {
			return io.ErrShortBuffer
		}
		anno := buf.Next(size)
		p.Anno = strings.Split(string(anno), " ")

	case PrimBytes:
		// cross-check content size
		size := int(binary.BigEndian.Uint32(buf.Next(4)))
		if buf.Len() < size {
			return io.ErrShortBuffer
		}
		p.Bytes = buf.Next(size)

	default:
		return fmt.Errorf("micheline: unknown primitive type 0x%x", tag)
	}
	p.Type = tag
	return nil
}

func (p Prim) FindOpCodes(typ OpCode) ([]Prim, bool) {
	if p.OpCode == typ {
		return []Prim{p}, true
	}
	found := make([]Prim, 0)
	for i := range p.Args {
		x, ok := p.Args[i].FindOpCodes(typ)
		if ok {
			found = append(found, x...)
		}
	}
	return found, len(found) > 0
}

func (p Prim) ContainsOpCode(typ OpCode) bool {
	if p.OpCode == typ {
		return true
	}
	for i := range p.Args {
		if p.Args[i].ContainsOpCode(typ) {
			return true
		}
	}
	return false
}

func (p Prim) FindLabels(label string) ([]Prim, bool) {
	if p.MatchesAnno(label) {
		return []Prim{p}, true
	}
	found := make([]Prim, 0)
	for i := range p.Args {
		x, ok := p.Args[i].FindLabels(label)
		if ok {
			found = append(found, x...)
		}
	}
	return found, len(found) > 0
}

func (p Prim) Index(label string) ([]int, bool) {
	if p.MatchesAnno(label) {
		return nil, true
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
	index := make([]int, len(path))
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	for i, v := range strings.Split(path, "/") {
		switch v {
		case "L", "l", "0":
			index[i] = 0
		case "R", "r", "1":
			index[i] = 1
		default:
			idx, err := strconv.Atoi(v)
			if err != nil {
				return InvalidPrim, fmt.Errorf("micheline: invalid path component '%v' at pos %d", v, i)
			}
			index[i] = idx
		}
	}
	return p.GetIndex(index)
}

func (p Prim) GetIndex(index []int) (Prim, error) {
	prim := p
	for _, v := range index {
		if v < 0 || len(prim.Args) < v {
			return InvalidPrim, fmt.Errorf("micheline: index %d out of bounds", v)
		}
		prim = prim.Args[v]
	}
	return prim, nil
}
