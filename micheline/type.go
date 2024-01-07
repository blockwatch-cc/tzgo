// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/tezos"
	"golang.org/x/exp/slices"
)

type Type struct {
	Prim
}

// Extra Types
const (
	TypeStruct = "struct"
	TypeUnion  = "union"
)

// Default names
const (
	CONST_ENTRYPOINT  = "@entrypoint"
	CONST_KEY         = "@key"
	CONST_VALUE       = "@value"
	CONST_ITEM        = "@item"
	CONST_PARAM       = "@param"
	CONST_RETURN      = "@return"
	CONST_UNION_LEFT  = "@or_0"
	CONST_UNION_RIGHT = "@or_1"
)

type Typedef struct {
	Name     string    `json:"name"`               // annotation label | @key | @value | @item | @params | @return
	Type     string    `json:"type"`               // opcode or struct | union
	Optional bool      `json:"optional,omitempty"` // Union only
	Args     []Typedef `json:"args,omitempty"`
	Path     []int     `json:"path"` // type tree original path to this element
}

func (a Typedef) IsValid() bool {
	return a.Name != "" || a.Type != "" || len(a.Args) > 0
}

func (a Typedef) Equal(b Typedef) bool {
	if a.Type != b.Type {
		return false
	}
	if a.Optional != b.Optional {
		return false
	}
	if len(a.Args) != len(b.Args) {
		return false
	}
	for i, av := range a.Args {
		if !av.Equal(b.Args[i]) {
			return false
		}
	}
	return true
}

func (a Typedef) Similar(b Typedef) bool {
	if a.Optional != b.Optional {
		return false
	}
	if ((a.Type == "list" || a.Type == "set") && len(a.Args) == 0 && b.Type == "map") ||
		((b.Type == "list" || b.Type == "set") && len(b.Args) == 0 && a.Type == "map") {
		return true
	}
	if a.Type != b.Type && !a.Optional {
		return false
	}
	if len(a.Args) != len(b.Args) && !a.Optional {
		return false
	}
	if len(a.Args) == len(b.Args) {
		for i, av := range a.Args {
			if !av.Similar(b.Args[i]) {
				return false
			}
		}
	}
	return true
}

func (t Typedef) Unfold() Typedef {
	b := Typedef{
		Name:     t.Name,
		Type:     t.Type,
		Optional: t.Optional,
	}
	for _, v := range t.Args {
		b.Args = append(b.Args, v.unfold()...)
	}
	return b
}

func (t Typedef) unfold() []Typedef {
	switch t.Type {
	case TypeStruct:
		// unfold nested structs
		args := make([]Typedef, 0, len(t.Args))
		for _, v := range t.Args {
			args = append(args, v.unfold()...)
		}
		if t.Optional {
			// keep struct header when optional
			t.Args = args
			return []Typedef{t}
		} else {
			return args
		}
	case TypeUnion:
		// unfold each arg independently
		for i, v := range t.Args {
			args := make([]Typedef, 0, len(v.Args))
			for _, vv := range v.Args {
				args = append(args, vv.unfold()...)
			}
			t.Args[i].Args = args
		}
	case "list", "set":
		// unfold nested structs inside list
		args := make([]Typedef, 0, len(t.Args))
		for _, v := range t.Args {
			args = append(args, v.unfold()...)
		}
		t.Args = args
	case "map":
		// unfold nested structs inside map key and map value independently
		if t.Args[0].Name == CONST_KEY && t.Args[0].Type == TypeStruct {
			t.Args[0].Args = t.Args[0].unfold()
		}
		if t.Args[1].Name == CONST_VALUE && t.Args[1].Type == TypeStruct {
			t.Args[1].Args = t.Args[1].unfold()
		}
	case "lambda":
		// unfold arg and result structs inside independently
		if len(t.Args) == 2 {
			if t.Args[0].Name == CONST_PARAM && t.Args[0].Type == TypeStruct {
				t.Args[0].Args = t.Args[0].unfold()
			}
			if t.Args[1].Name == CONST_RETURN && t.Args[1].Type == TypeStruct {
				t.Args[1].Args = t.Args[1].unfold()
			}
		}
	}
	return []Typedef{t}
}

func (t Typedef) StrictEqual(v Typedef) bool {
	if t.Name != v.Name {
		return false
	}
	return t.Equal(v)
}

func (t Typedef) Left() Typedef {
	if len(t.Args) > 0 {
		return t.Args[0]
	}
	return Typedef{}
}

func (t Typedef) Right() Typedef {
	if len(t.Args) > 1 {
		return t.Args[1]
	}
	return Typedef{}
}

func (t Typedef) OpCode() OpCode {
	switch t.Type {
	case TypeStruct:
		return T_PAIR
	case TypeUnion:
		return T_OR
	default:
		oc, _ := ParseOpCode(t.Type)
		return oc
	}
}

func (t Typedef) String() string {
	var b strings.Builder
	if t.Name != "" {
		b.WriteString(t.Name)
		b.WriteString(": ")
	}
	if t.Optional {
		b.WriteByte('?')
	}
	switch t.Type {
	case "map":
		b.WriteString("map[")
		n := t.Args[0].Name
		t.Args[0].Name = ""
		b.WriteString(t.Args[0].String())
		t.Args[0].Name = n
		b.WriteString("](")
		n = t.Args[1].Name
		t.Args[1].Name = ""
		b.WriteString(t.Args[1].String())
		t.Args[1].Name = n
		b.WriteString(")")
	case "set", "list":
		b.WriteByte('[')
		for i, v := range t.Args {
			if i > 0 {
				b.WriteString(", ")
			}
			n := v.Name
			v.Name = ""
			b.WriteString(v.String())
			v.Name = n
		}
		b.WriteByte(']')
	case "struct":
		b.WriteByte('{')
		for i, v := range t.Args {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(v.String())
		}
		b.WriteByte('}')
	case "union":
		b.WriteByte('(')
		for i, v := range t.Args {
			if i > 0 {
				b.WriteString(" | ")
			}
			b.WriteString(v.String())
		}
		b.WriteByte(')')
	case "contract":
		b.WriteString(t.Type)
		b.WriteByte('(')
		for i, v := range t.Args {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(v.String())
		}
		b.WriteByte(')')
	default:
		b.WriteString(t.Type)
		if len(t.Args) > 0 {
			b.WriteByte('(')
			for i, v := range t.Args {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(v.String())
			}
			b.WriteByte(')')
		}
	}
	// if len(t.Path) > 0 {
	// 	b.WriteString(" [")
	// 	for i, v := range t.Path {
	// 		if i > 0 {
	// 			b.WriteString(", ")
	// 		}
	// 		b.WriteString(strconv.Itoa(v))
	// 	}
	// 	b.WriteByte(']')
	// }
	return b.String()
}

func NewType(p Prim) Type {
	return Type{p.Clone()}
}

func NewTypePtr(p Prim) *Type {
	return &Type{p.Clone()}
}

func ParseType(s string) (t Type, err error) {
	err = t.UnmarshalJSON([]byte(s))
	return
}

func MustParseType(s string) (t Type) {
	if err := t.UnmarshalJSON([]byte(s)); err != nil {
		panic(err)
	}
	return
}

func (t *Type) UnmarshalJSON(buf []byte) error {
	return t.Prim.UnmarshalJSON(buf)
}

func (t *Type) UnmarshalBinary(buf []byte) error {
	return t.Prim.UnmarshalBinary(buf)
}

func (t Type) Clone() Type {
	return Type{t.Prim.Clone()}
}

func (t Type) Label() string {
	return t.GetVarAnnoAny()
}

func (t Type) HasLabel() bool {
	return t.HasAnno()
}

func (t Type) IsEqual(t2 Type) bool {
	return IsEqualPrim(t.Prim, t2.Prim, false)
}

func (t Type) IsEqualWithAnno(t2 Type) bool {
	return IsEqualPrim(t.Prim, t2.Prim, true)
}

func (t Type) Left() Type {
	if len(t.Args) > 0 {
		return Type{t.Args[0]}
	}
	return Type{}
}

func (t Type) Right() Type {
	if len(t.Args) > 1 {
		return Type{t.Args[1]}
	}
	return Type{}
}

func (t Type) Typedef(name string) Typedef {
	return buildTypedef(name, t.Prim, []int{})
}

func (t Type) TypedefPtr(name string) *Typedef {
	td := buildTypedef(name, t.Prim, []int{})
	return &td
}

func (t Type) IsSimilar(t2 Type) bool {
	u1 := t.Typedef("").Unfold()
	u2 := t2.Typedef("").Unfold()
	return u1.Similar(u2)
}

func (t Type) MarshalJSON() ([]byte, error) {
	if !t.IsValid() {
		return []byte("{}"), nil
	}
	return json.Marshal(buildTypedef("", t.Prim, []int{}))
}

func (p Prim) Implements(t Type) bool {
	td := buildTypedef("", t.Prim, []int{})
	return p.ImplementsType(td)
}

func (p Prim) ImplementsType(t Typedef) bool {
	err := p.Walk(func(p Prim) error {
		// fmt.Printf("CMP typ=%#v val=%s\n", t, p.Dump())
		switch p.OpCode {
		case D_PAIR:
			if t.Type == TypeStruct {
				// fmt.Println("> handle struct")
				for i, v := range p.UnfoldPair(Type{}) {
					if i >= len(t.Args) || !v.ImplementsType(t.Args[i]) {
						// fmt.Println("> BAD struct elem")
						return ErrTypeMismatch
					}
				}
				return PrimSkip
			}
		case D_SOME, D_NONE:
			if t.Optional {
				// fmt.Println("> OK optional")
				return PrimSkip
			}
		case D_TRUE, D_FALSE:
			if t.Type == T_BOOL.String() {
				// fmt.Println("> OK bool")
				return PrimSkip
			}
		case D_UNIT:
			if t.Type == T_UNIT.String() {
				// fmt.Println("> OK unit")
				return PrimSkip
			}
		case D_ELT:
			if len(t.Args) == 2 && p.Args[0].ImplementsType(t.Args[0]) && p.Args[1].ImplementsType(t.Args[1]) {
				// fmt.Println("> OK map")
				return PrimSkip
			}
		case D_LEFT:
			// walk left tree by clipping off right handled types
			if t.Type == TypeUnion {
				// fmt.Println("> UNION left")
				if len(t.Args) == 1 {
					t = t.Args[0]
				} else {
					t.Args = t.Args[:len(t.Args)-1]
				}
				if p.Args[0].ImplementsType(t) {
					// fmt.Println("> OK union left")
					return PrimSkip
				}
			}
		case D_RIGHT:
			if t.Type == TypeUnion && p.Args[0].ImplementsType(t.Args[len(t.Args)-1]) {
				// fmt.Println("> OK union right")
				return PrimSkip
			}
		default:
			oc, err := ParseOpCode(t.Type)
			if err != nil {
				return err
			}
			// fmt.Printf("> handle %s in %s prim\n", oc, p.Type)
			switch p.Type {
			case PrimSequence:
				switch oc {
				case T_MAP:
					for _, v := range p.Args {
						if !v.ImplementsType(t) {
							// fmt.Println("> BAD map elem")
							return ErrTypeMismatch
						}
					}
					return PrimSkip
				case T_SET:
					for _, v := range p.Args {
						if !v.ImplementsType(t) {
							// fmt.Println("> BAD set elem")
							return ErrTypeMismatch
						}
					}
					return PrimSkip
				case T_LIST:
					for _, v := range p.Args {
						if !v.ImplementsType(t.Args[0]) {
							// fmt.Println("> BAD list elem")
							return ErrTypeMismatch
						}
					}
					return PrimSkip
				case T_LAMBDA:
					if len(p.Args) > 0 && p.Args[0].IsInstruction() {
						// fmt.Println("> OK lambda")
						return PrimSkip
					}
				}

			case PrimInt:
				switch oc {
				case T_INT, T_NAT, T_MUTEZ, T_TIMESTAMP, T_BIG_MAP:
					// fmt.Println("> OK int")
					return PrimSkip
				}

			case PrimString:
				// sometimes timestamps and addresses can be strings
				switch oc {
				case T_STRING, T_ADDRESS, T_CONTRACT, T_KEY_HASH, T_KEY,
					T_SIGNATURE, T_TIMESTAMP, T_CHAIN_ID, T_TX_ROLLUP_L2_ADDRESS:
					// fmt.Println("> OK string")
					return PrimSkip
				}

			case PrimBytes:
				switch oc {
				case T_BYTES, T_ADDRESS, T_KEY_HASH, T_KEY,
					T_CONTRACT, T_SIGNATURE, T_CHAIN_ID,
					T_BLS12_381_G1, T_BLS12_381_G2, T_BLS12_381_FR,
					T_CHEST, T_CHEST_KEY,
					T_TX_ROLLUP_L2_ADDRESS:
					// fmt.Println("> OK bytes")
					return PrimSkip
				}
			default:
				// FIXME
				// T_SAPLING_STATE, T_SAPLING_TRANSACTION,
				// T_TICKET
				return PrimSkip

			}
		}
		// fmt.Printf("> no case matched\n")
		return ErrTypeMismatch
	})
	return err == nil
}

func buildTypedef(name string, typ Prim, path []int) Typedef {
	if typ.HasAnno() {
		n := typ.GetVarAnnoAny()
		if n != "" {
			name = n
		}
	}
	td := Typedef{
		Name: name,
		Type: typ.OpCode.String(),
		Path: slices.Clone(path),
	}

	switch typ.OpCode {
	case T_LIST, T_SET:
		if len(typ.Args) > 0 {
			td.Args = []Typedef{
				buildTypedef(CONST_ITEM, typ.Args[0], append(path, 0)),
			}
		}

	case T_MAP, T_BIG_MAP:
		td.Args = []Typedef{
			buildTypedef(CONST_KEY, typ.Args[0], []int{0}),
			buildTypedef(CONST_VALUE, typ.Args[1], []int{1}),
		}

	case T_CONTRACT:
		td.Args = make([]Typedef, len(typ.Args))
		for i, v := range typ.Args {
			td.Args[i] = buildTypedef(strconv.Itoa(i), v, append(path, i))
		}

	case T_TICKET:
		td.Args = []Typedef{
			buildTypedef(CONST_VALUE, typ.Args[0], append(path, 0)),
		}

	case T_LAMBDA:
		td.Args = make([]Typedef, len(typ.Args))
		if len(typ.Args) > 0 {
			td.Args[0] = buildTypedef(CONST_PARAM, typ.Args[0], []int{0})
		}
		if len(typ.Args) > 1 {
			td.Args[1] = buildTypedef(CONST_RETURN, typ.Args[1], []int{1})
		}

	case T_PAIR:
		args := typ.UnfoldTypeRecursive(path)
		td.Type = TypeStruct
		td.Args = make([]Typedef, len(args))
		for i, v := range args {
			td.Args[i] = buildTypedef(strconv.Itoa(i), v, v.Path)
		}

	case T_OPTION:
		td.Optional = true
		if len(typ.Args) > 0 {
			child := buildTypedef(name, typ.Args[0], append(path, 0))
			td.Type = child.Type
			td.Args = child.Args
		} else {
			td.Type = "unknown"
		}

	case T_OR:
		td.Type = TypeUnion
		td.Args = make([]Typedef, 0)
		label := CONST_UNION_LEFT
		for i, v := range typ.Args {
			child := buildTypedef(label, v, append(path, i))
			if child.Type == TypeUnion {
				td.Args = append(td.Args, child.Args...)
			} else {
				td.Args = append(td.Args, child)
			}
			label = CONST_UNION_RIGHT
		}

	case T_SAPLING_STATE, T_SAPLING_TRANSACTION:
		td.Type += fmt.Sprintf("(%d)", typ.Args[0].Int.Int64())

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
		// chain_id
		// unit
		// bls12_381_g1
		// bls12_381_g2
		// bls12_381_fr
		// sapling_state
		// sapling_transaction
		// never
		return Typedef{
			Name: name,
			Type: typ.OpCode.String(),
			Path: slices.Clone(path),
		}
	}

	return td
}

// build matching type tree for value
func (p Prim) BuildType() Type {
	// Note: don't set WasPacked flag recursively on all children; we set this flag
	// once on the top level type during dynamic type detection so that comb unfolding
	// works
	t := Prim{}
	// t := Prim{WasPacked: true}
	if p.OpCode.IsTypeCode() {
		t.OpCode = p.OpCode
	}
	switch p.Type {
	case PrimInt:
		t.OpCode = p.Type.TypeCode()
		t.Type = PrimNullary

	case PrimBytes:
		t.Type = PrimNullary
		// detect address encoding first
		var addr tezos.Address
		if err := addr.Decode(p.Bytes); err == nil {
			if addr.IsRollup() {
				t.OpCode = T_TX_ROLLUP_L2_ADDRESS
			} else {
				t.OpCode = T_ADDRESS
			}
		}
		if t.OpCode == 0 {
			t.OpCode = p.Type.TypeCode()
		}

	case PrimString:
		t.Type = PrimNullary
		if len(p.String) > 0 {
			// detect timestamp and address encoding first
			if _, err := time.Parse(time.RFC3339, p.String); err == nil {
				t.OpCode = T_TIMESTAMP
			} else if addr, err := tezos.ParseAddress(p.String); err == nil {
				if addr.IsRollup() {
					t.OpCode = T_TX_ROLLUP_L2_ADDRESS
				} else {
					t.OpCode = T_ADDRESS
				}
			} else if _, err := tezos.ParseSignature(p.String); err == nil {
				t.OpCode = T_SIGNATURE
			}
		}
		// fallback to string
		if t.OpCode == 0 {
			t.OpCode = p.Type.TypeCode()
		}

	case PrimSequence:
		switch {
		case p.LooksLikeCode():
			t.Type = PrimNullary // we don't know in/out types
			t.OpCode = T_LAMBDA
		case p.LooksLikeMap():
			t.OpCode = T_MAP
			t.Type = PrimBinary
			t.Args = []Prim{
				p.Args[0].Args[0].BuildType().Prim, // key type
				p.Args[0].Args[1].BuildType().Prim, // value type
			}
		case p.LooksLikeSet():
			t.OpCode = T_SET // guess, can also be LIST
			t.Type = PrimUnary
			t.Args = []Prim{
				p.Args[0].BuildType().Prim, // single set type
			}
		case len(p.Args) == 0:
			t.OpCode = T_LIST // guess, can be MAP, SET, LIST
			t.Type = PrimNullary
		case len(p.Args) == 1:
			t.OpCode = T_LIST // guess, can be SET, LIST
			t.Type = PrimUnary
			t.Args = []Prim{p.Args[0].BuildType().Prim}
		case len(p.Args) == 2:
			t.OpCode = T_PAIR
			t.Type = PrimBinary
			t.Args = []Prim{
				p.Args[0].BuildType().Prim,
				p.Args[1].BuildType().Prim,
			}
		default:
			// struct
			t.OpCode = T_PAIR
			t.Type = PrimVariadicAnno
			t.Args = make([]Prim, len(p.Args))
			for i, v := range p.Args {
				t.Args[i] = v.BuildType().Prim
			}
		}
	case PrimNullary, PrimNullaryAnno:
		t.Type = PrimNullary
		t.OpCode = p.OpCode.TypeCode()
	case PrimUnary, PrimUnaryAnno:
		t.OpCode = p.OpCode.TypeCode()
		switch t.OpCode {
		case T_LAMBDA:
			t.Type = PrimNullary
		case T_OR:
			// in data we only see one branch, so we have to guess the other type
			t.Type = PrimBinary
			inner := p.Args[0].BuildType().Prim
			t.Args = []Prim{inner, inner}
		case T_OPTION:
			// we only know the embedded type on D_SOME
			if p.OpCode == D_SOME {
				t.Type = PrimUnary
				t.Args = []Prim{p.Args[0].BuildType().Prim}
			} else {
				t.Type = PrimNullary
			}
		case T_BOOL, T_UNIT:
			t.Type = PrimNullary
		case T_TICKET:
			t.Type = PrimUnary
			t.Args = []Prim{p.Args[0].BuildType().Prim}
		}
	case PrimBinary, PrimBinaryAnno:
		if p.OpCode == D_ELT {
			t.OpCode = T_MAP
			t.Type = PrimBinary
			t.Args = []Prim{
				p.Args[0].BuildType().Prim,
				p.Args[1].BuildType().Prim,
			}
		} else {
			// probably a regular pair
			t.Type = PrimBinary
			t.OpCode = p.OpCode.TypeCode()
			t.Args = []Prim{
				p.Args[0].BuildType().Prim,
				p.Args[1].BuildType().Prim,
			}
		}

	case PrimVariadicAnno:
		// ? probably an operation
		t.Type = PrimNullary
		t.OpCode = p.OpCode.TypeCode()
	}
	return Type{t}
}

func (p Prim) CanUnfoldType() bool {
	if p.IsPair() {
		return true
	}
	if p.IsContainerType() || p.LooksLikeCode() {
		return false
	}
	if p.IsSequence() {
		return true
	}
	return false
}

func (p Prim) UnfoldTypeRecursive(path []int) []Prim {
	flat := make([]Prim, 0)
	for i, v := range p.Args {
		v.Path = append(slices.Clone(path), i)
		if !v.WasPacked && v.CanUnfoldType() && !v.HasAnno() {
			flat = append(flat, v.UnfoldTypeRecursive(v.Path)...)
		} else {
			flat = append(flat, v)
		}
	}
	return flat
}
