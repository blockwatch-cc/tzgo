// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// Michelson type spec
// see http://tezos.gitlab.io/whitedoc/michelson.html#full-grammar
//
// PACK prefixes with 0x05!
// So that when contracts checking signatures (multisigs etc) do the current
// best practice, PACK; ...; CHECK_SIGNATURE, the 0x05 byte distinguishes the
// message from blocks, endorsements, transactions, or tezos-signer authorization
// requests (0x01-0x04)

package micheline

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
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

	PrimSkip        = errors.New("skip branch")
	ErrTypeMismatch = errors.New("type mismatch")
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
	case INT:
		return PrimInt, nil
	case STRING:
		return PrimString, nil
	case BYTES:
		return PrimBytes, nil
	default:
		return 0, fmt.Errorf("micheline: invalid prim type '%s'", val)
	}
}

// non-normative strings, use for debugging only
func (t PrimType) String() string {
	switch t {
	case PrimInt:
		return INT
	case PrimString:
		return STRING
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
		return BYTES
	default:
		return "invalid"
	}
}

func (t PrimType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

type PrimList []Prim

func (l PrimList) Last() Prim {
	n := len(l)
	if n == 0 {
		return InvalidPrim
	}
	return l[n-1]
}

type Prim struct {
	Type      PrimType // primitive type
	OpCode    OpCode   // primitive opcode (invalid on sequences, strings, bytes, int)
	Args      PrimList // optional arguments
	Anno      []string // optional type annotations
	Int       *big.Int // optional data
	String    string   // optional data
	Bytes     []byte   // optional data
	WasPacked bool     // true when content was unpacked (and no type info is available)
	Path      []int    // optional path to this prim (use to track type structure)
}

func (p Prim) IsValid() bool {
	return p.Type.IsValid() && (p.Type > 0 || p.Int != nil)
}

func (p Prim) IsEmpty() bool {
	return p.Type == PrimNullary && p.OpCode == 255
}

func (p Prim) Size() int {
	sz := szPrim + len(p.String) + len(p.Bytes)
	if p.Int != nil {
		sz += len(p.Int.Bytes())
	}
	for _, v := range p.Anno {
		sz += len(v)
	}
	for _, v := range p.Args {
		sz += v.Size()
	}
	return sz
}

func (p Prim) Hash64() uint64 {
	var buf []byte
	if p.IsValid() {
		buf, _ = p.MarshalBinary()
	}
	h := sha256.Sum256(buf)
	return binary.BigEndian.Uint64(h[:8])
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
		copy(clone.Anno, p.Anno)
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

func (p Prim) CloneNoAnnots() Prim {
	typ := p.Type
	switch typ {
	case PrimNullaryAnno, PrimUnaryAnno, PrimBinaryAnno:
		typ--
	}
	clone := Prim{
		Type:      typ,
		OpCode:    p.OpCode,
		String:    p.String,
		WasPacked: p.WasPacked,
	}
	if p.Args != nil {
		clone.Args = make([]Prim, len(p.Args))
		for i, arg := range p.Args {
			clone.Args[i] = arg.CloneNoAnnots()
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

func (p Prim) Compare(p2 Prim) int {
	if p.Type != p2.Type {
		return 0
	}
	switch p.Type {
	case PrimString:
		return strings.Compare(p.String, p2.String)
	case PrimBytes:
		return bytes.Compare(p.Bytes, p2.Bytes)
	case PrimInt:
		return p.Int.Cmp(p2.Int)
	default:
		return 0
	}
}

func IsEqualPrim(p1, p2 Prim, withAnno bool) bool {
	if p1.OpCode != p2.OpCode {
		return false
	}

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
		return false
	}

	if len(p1.Args) != len(p2.Args) {
		return false
	}

	if withAnno {
		if len(p1.Anno) != len(p2.Anno) {
			return false
		}
		for i := range p1.Anno {
			l1, l2 := len(p1.Anno[i]), len(p2.Anno[i])
			if l1 != l2 {
				return false
			}
			if l1 > 0 && l2 > 0 && p1.Anno[i][1:] != p2.Anno[i][1:] {
				return false
			}
		}
	}

	if p1.String != p2.String {
		return false
	}
	if (p1.Int == nil) != (p2.Int == nil) {
		return false
	}
	if p1.Int != nil {
		if p1.Int.Cmp(p2.Int) != 0 {
			return false
		}
	}
	if (p1.Bytes == nil) != (p2.Bytes == nil) {
		return false
	}
	if p1.Bytes != nil {
		if !bytes.Equal(p1.Bytes, p2.Bytes) {
			return false
		}
	}

	for i := range p1.Args {
		if !IsEqualPrim(p1.Args[i], p2.Args[i], withAnno) {
			return false
		}
	}

	return true
}

// PrimWalkerFunc is the callback function signature used while
// traversing a prim tree in read-only mode.
type PrimWalkerFunc func(p Prim) error

// Walk traverses the prim tree in pre-order in read-only mode, forwarding
// value copies to the callback.
func (p Prim) Walk(f PrimWalkerFunc) error {
	if err := f(p); err != nil {
		if err == PrimSkip {
			return nil
		}
		return err
	}
	for _, v := range p.Args {
		if err := v.Walk(f); err != nil {
			return err
		}
	}
	return nil
}

// PrimWalkerFunc is the callback function signature used while
// traversing a prim tree. The callback may change the contents of
// the visited node, including altering nested child nodes and annotations.
type PrimVisitorFunc func(p *Prim) error

// Visit traverses the prim tree in pre-order and allows the callback to
// alter the contents of a visited node.
func (p *Prim) Visit(f PrimVisitorFunc) error {
	if err := f(p); err != nil {
		if err == PrimSkip {
			return nil
		}
		return err
	}
	for i := range p.Args {
		if err := p.Args[i].Visit(f); err != nil {
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
		return true
	case PrimSequence:
		return len(p.Args) == 1 && !p.HasAnno()
	case PrimNullaryAnno, PrimUnaryAnno, PrimBinaryAnno, PrimVariadicAnno:
		return false
	case PrimUnary:
		switch p.OpCode {
		case D_LEFT, D_RIGHT, D_SOME:
			return p.Args[0].IsScalar()
		}
		return false
	case PrimBinary:
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

func (p Prim) IsInstruction() bool {
	return p.OpCode.TypeCode() == T_LAMBDA
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

func (p Prim) IsConstant() bool {
	return p.OpCode == H_CONSTANT
}

func (p Prim) IsPair() bool {
	switch p.OpCode {
	case T_PAIR, D_PAIR:
		return true
	}
	return false
}

func (p Prim) IsNil() bool {
	switch p.OpCode {
	case D_UNIT, D_NONE:
		return true
	}
	return false
}

func (p Prim) IsEmptyBigmap() bool {
	return p.OpCode == I_EMPTY_BIG_MAP
}

func (p Prim) IsScalarType() bool {
	switch p.OpCode {
	case T_BOOL,
		T_CONTRACT,
		T_INT,
		T_KEY,
		T_KEY_HASH,
		T_NAT,
		T_SIGNATURE,
		T_STRING,
		T_BYTES,
		T_MUTEZ,
		T_TIMESTAMP,
		T_UNIT,
		T_OPERATION,
		T_ADDRESS,
		T_CHAIN_ID,
		T_NEVER,
		T_BLS12_381_G1,
		T_BLS12_381_G2,
		T_BLS12_381_FR,
		T_SAPLING_STATE,
		T_SAPLING_TRANSACTION:
		return true
	default:
		return false
	}
}

func (p Prim) IsContainerType() bool {
	switch p.OpCode {
	case T_MAP, T_LIST, T_SET, T_LAMBDA:
		return true
	default:
		return false
	}
}

// Detects whether a primitive contains a regular pair or any form
// of container type. Pairs can be unfolded into flat sequences.
func (p Prim) CanUnfold(typ Type) bool {
	// regular pairs (works for type and value trees)
	if p.IsPair() {
		return true
	}

	// Note: due to Tezos encoding issues, comb pairs in block receipts
	// are naked sequences (they lack an enclosing pair wrapper), this means
	// we have to distinquish them from other container types who also use
	// a sequence as container, such as lists, sets, maps, lambdas
	if p.IsContainerType() || typ.IsContainerType() {
		return false
	}

	// When multiple optimized combs (i.e. naked sequences) are nested
	// it is sometimes not possible to track the correct type in the
	// type tree. Hence we revert to a heuristic that checks if the
	// current primitive looks like a container type by checking its
	// contents
	if p.IsSequence() && !p.LooksLikeContainer() && !p.LooksLikeCode() {
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
	return len(p.Args) > 0
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
	if p.Args[0].IsElt() && p.Args[len(p.Args)-1].IsElt() {
		return true
	}

	// contains a single scalar
	if len(p.Args) == 1 {
		switch p.Args[0].Type {
		case PrimInt, PrimString, PrimBytes, PrimNullary:
			return true
		}
	}

	// contains similar records
	if p.HasSimilarChildTypes() {
		return true
	}

	return false
}

// Checks if all children have the same type by generating a type tree from
// values. Can be used to identfy containers based on the existence of similar
// records.
//
// Works for simple and nested primitives but may mis-detect ambiguous
// simple types like PrimInt (used for int, nat, timestamp, mutez), or PrimString
// resp. PrimBytes. May also misdetect when optional types like T_OR, T_OPTION are
// used and their values are nil since we cannot detect embedded type here.
func (p Prim) HasSimilarChildTypes() bool {
	if len(p.Args) == 0 {
		return true
	}
	oc := p.Args[0].OpCode.TypeCode()
	firstType := p.Args[0].BuildType()
	for _, v := range p.Args[1:] {
		var isSame bool
		switch v.OpCode {
		case D_SOME, D_NONE, D_FALSE, D_TRUE, D_LEFT, D_RIGHT:
			isSame = oc == v.OpCode.TypeCode()
		default:
			isSame = firstType.IsSimilar(v.BuildType())
		}
		if !isSame {
			return false
		}
	}
	return true
}

func (p Prim) LooksLikeMap() bool {
	// must be a sequence
	if !p.IsSequence() || len(p.Args) == 0 {
		return false
	}

	// contents must be Elt
	return p.Args[0].IsElt() && p.Args[len(p.Args)-1].IsElt()
}

func (p Prim) LooksLikeSet() bool {
	// must be a sequence
	if !p.IsSequence() || len(p.Args) == 0 {
		return false
	}

	// contains similar records
	if !p.HasSimilarChildTypes() {
		return false
	}

	return true
}

// Checks if a Prim looks like a lambda type.
func (p Prim) LooksLikeCode() bool {
	if p.OpCode == T_LAMBDA || p.IsInstruction() {
		return true
	}

	if p.Type != PrimSequence || len(p.Args) == 0 {
		return false
	}

	// first prim contains instruction ocpode
	if p.Args[0].IsInstruction() {
		return true
	}

	return false
}

// Converts a pair tree into a flat sequence. While Michelson
// optimized comb pairs are only used for right-side combs, this
// function applies to all pairs. It makes use of the type definition
// to identify which contained type is a regular pair, an already
// unfolded pair sequence or anther container type.
//
// - Works both on value trees and type trees.
// - When called on already converted comb sequences this function is a noop.
func (p Prim) UnfoldPair(typ Type) []Prim {
	flat := make([]Prim, 0)
	for i, v := range p.Args {
		t := Type{}
		if len(typ.Args) > i {
			t = Type{typ.Args[i]}
		}
		if !v.WasPacked && v.CanUnfold(t) && !t.HasAnno() {
			flat = append(flat, v.Args...)
		} else {
			flat = append(flat, v)
		}
	}
	return flat
}

func (p Prim) UnfoldPairRecursive(typ Type) []Prim {
	flat := make([]Prim, 0)
	for i, v := range p.Args {
		t := Type{}
		if len(typ.Args) > i {
			t = Type{typ.Args[i]}
		}
		if !v.WasPacked && v.CanUnfold(t) && !t.HasAnno() {
			flat = append(flat, v.UnfoldPairRecursive(t)...)
		} else {
			flat = append(flat, v)
		}
	}
	return flat
}

// Turns a pair sequence into a right-hand pair tree
func (p Prim) FoldPair() Prim {
	if !p.IsSequence() || len(p.Args) < 2 {
		return p
	}
	switch len(p.Args) {
	case 2:
		return NewPair(p.Args[0], p.Args[1])
	default:
		return NewPair(p.Args[0], NewSeq(p.Args[1:]...).FoldPair())
	}
}

// Checks if a primitve contains a packed value such as a byte sequence
// generated with PACK (starting with 0x05), an address or ascii/utf string.
func (p Prim) IsPacked() bool {
	return p.Type == PrimBytes &&
		(isPackedBytes(p.Bytes) || tezos.IsAddressBytes(p.Bytes) || isASCIIBytes(p.Bytes))
}

// Packs produces a packed serialization for of a primitive's contents that
// is prefixed with a 0x5 byte.
func (p Prim) Pack() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(0x5)
	_ = p.EncodeBuffer(buf)
	return buf.Bytes()
}

// Unpacks all primitive contents that looks like packed and returns a new primitive
// tree.
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
	switch {
	case isPackedBytes(p.Bytes):
		if err := pp.UnmarshalBinary(p.Bytes[1:]); err != nil {
			return p, err
		}
		if pp.IsPackedAny() {
			if up, err := pp.UnpackAll(); err == nil {
				pp = up
			}
		}
	case tezos.IsAddressBytes(p.Bytes):
		a := tezos.Address{}
		if err := a.Decode(p.Bytes); err != nil {
			return p, err
		}
		pp.Type = PrimString
		pp.String = a.String()
	case isASCII(string(p.Bytes)):
		pp.Type = PrimString
		pp.String = string(p.Bytes)
	default:
		pp = p
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
	if p.LooksLikeCode() {
		return p, nil
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

// UnpackAsciiString converts ASCII strings inside byte prims.
func (p Prim) UnpackAsciiString() Prim {
	if p.Bytes != nil {
		if s := string(p.Bytes); isASCII(s) {
			return Prim{
				Type:   PrimString,
				String: s,
			}
		}
	}
	return p
}

// UnpackAllAsciiStrings recursively converts all ASCII strings inside byte prims.
func (p Prim) UnpackAllAsciiStrings() Prim {
	if len(p.Args) == 0 {
		return p.UnpackAsciiString()
	}
	up := p
	up.Args = make([]Prim, len(p.Args))
	for i, v := range p.Args {
		up.Args[i] = v.UnpackAllAsciiStrings()
	}
	return up
}

// Returns a typed/decoded value from an encoded primitive.
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
			var z tezos.Z
			z.SetBig(p.Int)
			return z
		}

	case PrimString:
		switch as {
		case T_TIMESTAMP:
			if t, err := time.Parse(time.RFC3339, p.String); err == nil {
				return t
			}
			return p.String

		case T_KEY_HASH, T_ADDRESS, T_CONTRACT:
			a, err := tezos.ParseAddress(p.String)
			if err == nil {
				return a
			}

		case T_KEY:
			k, err := tezos.ParseKey(p.String)
			if err == nil {
				return k
			}

		case T_SIGNATURE:
			s, err := tezos.ParseSignature(p.String)
			if err == nil {
				return s
			}

		case T_CHAIN_ID:
			id, err := tezos.ParseChainIdHash(p.String)
			if err == nil {
				return id
			}

		default:
			return p.String
		}
		return p.String

	case PrimBytes:
		switch as {
		case T_KEY_HASH, T_ADDRESS, T_CONTRACT:
			a := tezos.Address{}
			if err := a.Decode(p.Bytes); err == nil {
				return a
			}
		case T_TX_ROLLUP_L2_ADDRESS:
			return tezos.NewAddress(tezos.AddressTypeBls12_381, p.Bytes)

		case T_KEY:
			k := tezos.Key{}
			if err := k.UnmarshalBinary(p.Bytes); err == nil {
				return k
			}

		case T_SIGNATURE:
			s := tezos.Signature{}
			if err := s.UnmarshalBinary(p.Bytes); err == nil {
				return s
			}

		case T_CHAIN_ID:
			if len(p.Bytes) == tezos.HashTypeChainId.Len {
				return tezos.NewChainIdHash(p.Bytes)
			}

		default:
			// as hex, fallthrough
			// case T_BYTES:
			// case T_BLS12_381_G1, T_BLS12_381_G2, T_BLS12_381_FR:
			// case T_SAPLING_STATE:
			// case T_LAMBDA:
			// case T_LIST, T_MAP, T_BIG_MAP, T_SET:
			// case T_OPTION, T_OR, T_PAIR, T_UNIT:
			// case T_OPERATION:
		}

		return hex.EncodeToString(p.Bytes)

	case PrimUnary, PrimUnaryAnno:
		switch as {
		case T_OR:
			switch p.OpCode {
			case D_LEFT:
				return p.Args[0].Value(as)
			case D_RIGHT:
				if len(p.Args) > 1 {
					return p.Args[1].Value(as)
				} else {
					return p.Args[0].Value(as)
				}
			}
		case T_OPTION:
			switch p.OpCode {
			case D_NONE:
				return nil
			case D_SOME:
				return p.Args[0].Value(as)
			}
		default:
			warn = true
		}

	case PrimNullary, PrimNullaryAnno:
		switch p.OpCode {
		case D_FALSE:
			return false
		case D_TRUE:
			return true
		case D_UNIT, D_NONE:
			return nil
		default:
			return p.OpCode.String()
		}

	case PrimBinary, PrimBinaryAnno:
		switch p.OpCode {
		case D_PAIR, T_PAIR:
			// FIXME: requires value tree decoration (types in opcodes)
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
			// FIXME: requires value tree decoration (types in opcodes)
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
		case T_LAMBDA:
			return p.OpCode.String()
		case T_BYTES:
			return hex.EncodeToString(p.Bytes)
		default:
			warn = true
		}
	}

	if warn && !p.WasPacked {
		log.Warnf("Rendering prim type %s as %s: not implemented (%s)", p.Type, as, p.Dump())
	}

	return p
}

func (p Prim) MarshalYAML() (any, error) {
	buf, err := p.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return string(buf), nil
}

func (p Prim) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	p.EncodeJSON(buf)
	return buf.Bytes(), nil
}

func (p Prim) EncodeJSON(buf *bytes.Buffer) {
	if !p.IsValid() {
		buf.WriteString("{}")
		return
	}
	switch p.Type {
	case PrimSequence:
		buf.WriteByte('[')
		for i, v := range p.Args {
			if i > 0 {
				buf.WriteByte(',')
			}
			v.EncodeJSON(buf)
		}
		buf.WriteByte(']')

	case PrimInt:
		buf.WriteString(`{"int":"`)
		buf.WriteString(p.Int.Text(10))
		buf.WriteString(`"}`)

	case PrimString:
		buf.WriteString(`{"string":`)
		buf.WriteString(strconv.Quote(p.String))
		buf.WriteByte('}')

	case PrimBytes:
		buf.WriteString(`{"bytes":"`)
		buf.WriteString(hex.EncodeToString(p.Bytes))
		buf.WriteString(`"}`)

	default:
		buf.WriteString(`{"prim":"`)
		buf.WriteString(p.OpCode.String())
		buf.WriteByte('"')
		if len(p.Anno) > 0 && len(p.Anno[0]) > 0 {
			buf.WriteString(`,"annots":[`)
			for i, v := range p.Anno {
				if i > 0 {
					buf.WriteByte(',')
				}
				buf.WriteString(strconv.Quote(v))
			}
			buf.WriteByte(']')
		}
		if len(p.Args) > 0 {
			buf.WriteString(`,"args":[`)
			for i, v := range p.Args {
				if i > 0 {
					buf.WriteByte(',')
				}
				v.EncodeJSON(buf)
			}
			buf.WriteByte(']')
		}
		buf.WriteByte('}')
	}
}

func (p Prim) ToBytes() []byte {
	buf, _ := p.MarshalBinary()
	return buf
}

func (p Prim) MarshalBinary() ([]byte, error) {
	if !p.IsValid() {
		return nil, nil
	}
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
		var z tezos.Z
		z.SetBig(p.Int)
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
		binary.BigEndian.PutUint32(res, uint32(len(res)-4))
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
		binary.BigEndian.PutUint32(res, uint32(len(res)-4))
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
	var val any
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case []any:
		return p.UnpackSequence(v)
	case map[string]any:
		return p.UnpackPrimitive(v)
	default:
		return nil
	}
}

func (p *Prim) UnpackJSON(val any) error {
	switch t := val.(type) {
	case map[string]any:
		return p.UnpackPrimitive(t)
	case []any:
		return p.UnpackSequence(t)
	default:
		return fmt.Errorf("micheline: unexpected json type %T", val)
	}
}

func (p *Prim) UnpackSequence(val []any) error {
	p.Type = PrimSequence
	p.Args = make([]Prim, len(val))
	for i, v := range val {
		if err := p.Args[i].UnpackJSON(v); err != nil {
			return err
		}
	}
	return nil
}

func (p *Prim) UnpackPrimitive(val map[string]any) error {
	for n, v := range val {
		switch n {
		case PRIM:
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
		case INT:
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("micheline: invalid int value type %T %v", v, v)
			}
			i := big.NewInt(0)
			i.SetString(str, 0)
			p.Int = i
			p.Type = PrimInt
		case STRING:
			str, ok := v.(string)
			if !ok {
				return fmt.Errorf("micheline: invalid string value type %T %v", v, v)
			}
			p.String = str
			p.Type = PrimString
		case BYTES:
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
		case ANNOTS:
			slist, ok := v.([]any)
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
	if a, ok := val[ARGS]; ok {
		args, ok := a.([]any)
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
		p.Args = make([]Prim, len(args))
		for i, v := range args {
			if err := p.Args[i].UnpackJSON(v); err != nil {
				return err
			}
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
		var z tezos.Z
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
		return fmt.Errorf("micheline: unknown primitive type 0x%x", byte(tag))
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

func (p Prim) FindBigmapByName(name string) (Prim, bool) {
	if p.OpCode == T_BIG_MAP && p.MatchesAnno(name) {
		return p, true
	}
	for i := range p.Args {
		if pp, found := p.Args[i].FindBigmapByName(name); found {
			return pp, found
		}
	}
	return Prim{}, false
}
