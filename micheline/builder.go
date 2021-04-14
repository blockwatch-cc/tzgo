// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"math/big"
)

func NewCode(c OpCode, args ...Prim) Prim {
	typ := PrimNullary
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

func NewBig(i *big.Int) Prim {
	return Prim{Type: PrimInt, Int: i}
}

func NewBytes(b []byte) Prim {
	return Prim{Type: PrimBytes, Bytes: b}
}

func NewString(s string) Prim {
	return Prim{Type: PrimString, String: s}
}

func NewPairType(l, r Prim, anno ...string) Prim {
	typ := PrimBinary
	if len(anno) > 0 {
		typ = PrimBinaryAnno
	}
	return Prim{Type: typ, OpCode: T_PAIR, Args: []Prim{l, r}, Anno: anno}
}

func NewPairValue(l, r Prim, anno ...string) Prim {
	typ := PrimBinary
	if len(anno) > 0 {
		typ = PrimBinaryAnno
	}
	return Prim{Type: typ, OpCode: D_PAIR, Args: []Prim{l, r}, Anno: anno}
}

func NewPrim(c OpCode, anno ...string) Prim {
	typ := PrimNullary
	if len(anno) > 0 {
		typ = PrimNullaryAnno
	}
	return Prim{Type: typ, OpCode: c, Anno: anno}
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
