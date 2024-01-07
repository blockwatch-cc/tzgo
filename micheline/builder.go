// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"math/big"
	"sort"

	"blockwatch.cc/tzgo/tezos"
)

var (
	Unit = NewCode(D_UNIT)
)

func NewCode(c OpCode, args ...Prim) Prim {
	var typ PrimType
	switch len(args) {
	case 0:
		typ = PrimNullary
	case 1:
		typ = PrimUnary
	case 2:
		typ = PrimBinary
	default:
		typ = PrimVariadicAnno
	}
	return Prim{Type: typ, OpCode: c, Args: args}
}

func NewCodeAnno(c OpCode, anno string, args ...Prim) Prim {
	p := NewCode(c, args...)
	p.Anno = []string{anno}
	if p.Type != PrimVariadicAnno {
		p.Type++
	}
	return p
}

func NewSeq(args ...Prim) Prim {
	return Prim{Type: PrimSequence, Args: args}
}

func NewInt64(i int64) Prim {
	return NewBig(big.NewInt(i))
}

func NewZ(z tezos.Z) Prim {
	return NewBig(new(big.Int).Set(z.Big()))
}

func NewMutez(n tezos.N) Prim {
	return NewBig(big.NewInt(int64(n)))
}

func NewBig(i *big.Int) Prim {
	return Prim{Type: PrimInt, Int: i}
}

func NewNat(i *big.Int) Prim {
	return Prim{Type: PrimInt, Int: i}
}

func NewKeyHash(a tezos.Address) Prim {
	return NewBytes(a.Encode())
}

func NewAddress(a tezos.Address) Prim {
	return NewBytes(a.EncodePadded())
}

func NewBytes(b []byte) Prim {
	return Prim{Type: PrimBytes, Bytes: b}
}

func NewString(s string) Prim {
	return Prim{Type: PrimString, String: s}
}

func NewOption(p ...Prim) Prim {
	if len(p) == 0 {
		return Prim{Type: PrimNullary, OpCode: D_NONE}
	}
	return Prim{Type: PrimUnary, OpCode: D_SOME, Args: []Prim{p[0]}}
}

func NewPairType(l, r Prim, anno ...string) Prim {
	typ := PrimBinary
	if len(anno) > 0 {
		typ = PrimBinaryAnno
	}
	return Prim{Type: typ, OpCode: T_PAIR, Args: []Prim{l, r}, Anno: anno}
}

func NewMapType(k, v Prim, anno ...string) Prim {
	typ := PrimBinary
	if len(anno) > 0 {
		typ = PrimBinaryAnno
	}
	return Prim{Type: typ, OpCode: T_MAP, Args: []Prim{k, v}, Anno: anno}
}

func NewMap(elts ...Prim) Prim {
	sort.Slice(elts, func(i, j int) bool {
		return elts[i].Args[0].Compare(elts[j].Args[0]) <= 0
	})
	return Prim{Type: PrimSequence, Args: elts}
}

func NewMapElem(k, v Prim) Prim {
	return Prim{Type: PrimBinary, OpCode: D_ELT, Args: []Prim{k, v}}
}

func NewSetType(e Prim, anno ...string) Prim {
	typ := PrimUnary
	if len(anno) > 0 {
		typ = PrimUnaryAnno
	}
	return Prim{Type: typ, OpCode: T_SET, Args: []Prim{e}, Anno: anno}
}

func NewOptType(e Prim, anno ...string) Prim {
	typ := PrimUnary
	if len(anno) > 0 {
		typ = PrimUnaryAnno
	}
	return Prim{Type: typ, OpCode: T_OPTION, Args: []Prim{e}, Anno: anno}
}

func NewPair(l, r Prim) Prim {
	return Prim{Type: PrimBinary, OpCode: D_PAIR, Args: []Prim{l, r}}
}

func NewCombPair(contents ...Prim) Prim {
	return Prim{Type: PrimSequence, Args: contents}
}

func NewCombPairType(contents ...Prim) Prim {
	return Prim{Type: PrimSequence, OpCode: T_PAIR, Args: contents}
}

func NewPrim(c OpCode, anno ...string) Prim {
	typ := PrimNullary
	if len(anno) > 0 {
		typ = PrimNullaryAnno
	}
	return Prim{Type: typ, OpCode: c, Anno: anno}
}

func NewUnion(path []int, prim Prim) Prim {
	if len(path) == 0 {
		return prim
	}
	oc := D_LEFT
	if path[0] == 1 {
		oc = D_RIGHT
	}
	return NewCode(oc, NewUnion(path[1:], prim))
}

func (p Prim) WithAnno(anno string) Prim {
	p.Anno = append(p.Anno, anno)
	return p
}

// Macros
func ASSERT_CMPEQ() Prim {
	return NewSeq(
		NewSeq(NewCode(I_COMPARE), NewCode(I_EQ)),
		NewCode(I_IF, NewSeq(), NewSeq(NewSeq(NewCode(I_UNIT), NewCode(I_FAILWITH)))),
	)
}

func DUUP() Prim {
	return NewSeq(NewCode(I_DIP, NewSeq(NewCode(I_DUP)), NewCode(I_SWAP)))
}

func IFCMPNEQ(left, right Prim) Prim {
	return NewSeq(
		NewCode(I_COMPARE),
		NewCode(I_EQ),
		NewCode(I_IF, left, right),
	)
}

func UNPAIR() Prim {
	return NewSeq(NewSeq(
		NewCode(I_DUP),
		NewCode(I_CAR),
		NewCode(I_DIP, NewSeq(NewCode(I_CDR))),
	))
}
