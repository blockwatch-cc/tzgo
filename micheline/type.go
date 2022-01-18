// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"blockwatch.cc/tzgo/tezos"
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
}

func NewType(p Prim) Type {
	return Type{p.Clone()}
}

func NewTypePtr(p Prim) *Type {
	return &Type{p.Clone()}
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
	return buildTypedef(name, t.Prim)
}

func (t Type) TypedefPtr(name string) *Typedef {
	td := buildTypedef(name, t.Prim)
	return &td
}

func (t Type) MarshalJSON() ([]byte, error) {
	if !t.IsValid() {
		return []byte("{}"), nil
	}
	return json.Marshal(buildTypedef("", t.Prim))
}

func buildTypedef(name string, typ Prim) Typedef {
	if typ.HasAnno() {
		name = typ.GetVarAnnoAny()
	}
	td := Typedef{
		Name: name,
		Type: typ.OpCode.String(),
	}

	switch typ.OpCode {
	case T_LIST, T_SET:
		td.Args = []Typedef{
			buildTypedef(CONST_ITEM, typ.Args[0]),
		}

	case T_MAP, T_BIG_MAP:
		td.Args = []Typedef{
			buildTypedef(CONST_KEY, typ.Args[0]),
			buildTypedef(CONST_VALUE, typ.Args[1]),
		}

	case T_CONTRACT:
		td.Args = make([]Typedef, len(typ.Args))
		for i, v := range typ.Args {
			td.Args[i] = buildTypedef(strconv.Itoa(i), v)
		}

	case T_TICKET:
		td.Args = []Typedef{
			buildTypedef(CONST_VALUE, typ.Args[0]),
		}

	case T_LAMBDA:
		td.Args = make([]Typedef, len(typ.Args))
		if len(typ.Args) > 0 {
			td.Args[0] = buildTypedef(CONST_PARAM, typ.Args[0])
		}
		if len(typ.Args) > 1 {
			td.Args[1] = buildTypedef(CONST_RETURN, typ.Args[1])
		}

	case T_PAIR:
		typs := typ.UnfoldPairRecursive(Type{typ})
		td.Type = TypeStruct
		td.Args = make([]Typedef, len(typs))
		for i, v := range typs {
			td.Args[i] = buildTypedef(strconv.Itoa(i), v)
		}

	case T_OPTION:
		child := buildTypedef(name, typ.Args[0])
		td.Optional = true
		td.Type = child.Type
		td.Args = child.Args

	case T_OR:
		td.Type = TypeUnion
		td.Args = make([]Typedef, 0)
		label := CONST_UNION_LEFT
		for _, v := range typ.Args {
			child := buildTypedef(label, v)
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
		if err := addr.UnmarshalBinary(p.Bytes); err == nil {
			t.OpCode = T_ADDRESS
		} else {
			t.OpCode = p.Type.TypeCode()
		}

	case PrimString:
		t.Type = PrimNullary
		if len(p.String) > 0 {
			// detect timestamp and address encoding first
			if _, err := time.Parse(time.RFC3339, p.String); err == nil {
				t.OpCode = T_TIMESTAMP
			} else if _, err := tezos.ParseAddress(p.String); err == nil {
				t.OpCode = T_ADDRESS
			} else {
				t.OpCode = p.Type.TypeCode()
			}
		} else {
			t.OpCode = p.Type.TypeCode()
		}

	case PrimSequence:
		switch {
		case p.LooksLikeMap():
			t.OpCode = T_MAP
			t.Type = PrimBinary
			t.Args = []Prim{
				p.Args[0].Args[0].BuildType().Prim, // key type
				p.Args[0].Args[1].BuildType().Prim, // value type
			}
		case p.LooksLikeCode():
			t.Type = PrimNullary // we don't know in/out types
			t.OpCode = T_LAMBDA
		case p.LooksLikeSet():
			t.OpCode = T_SET
			t.Type = PrimUnary
			t.Args = []Prim{
				p.Args[0].BuildType().Prim, // single set type
			}
		default:
			// walk the entire list and generate types for each element in-order
			t.OpCode = T_LIST
			switch len(p.Args) {
			case 0:
				t.Type = PrimNullary
			case 1:
				t.Type = PrimUnary
				t.Args = []Prim{p.Args[0].BuildType().Prim}
			case 2:
				t.Type = PrimBinary
				t.Args = []Prim{
					p.Args[0].BuildType().Prim,
					p.Args[1].BuildType().Prim,
				}
			default:
				t.Type = PrimVariadicAnno
				t.Args = make([]Prim, len(p.Args))
				for i, v := range p.Args {
					t.Args[i] = v.BuildType().Prim
				}
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
