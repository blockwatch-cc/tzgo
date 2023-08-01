package bind

import (
	"blockwatch.cc/tzgo/micheline"
	"github.com/pkg/errors"
)

// Or is a type that can be either L or R.
//
// It maps to michelson's `or` type.
type Or[L, R any] struct {
	l       L
	r       R
	isRight bool
}

// Left returns a new Or[L, R] filled with the left value.
func Left[L, R any](l L) Or[L, R] {
	return Or[L, R]{l: l}
}

// Right returns a new Or[L, R] filled with the right value.
func Right[L, R any](r R) Or[L, R] {
	return Or[L, R]{r: r, isRight: true}
}

func (o Or[L, R]) IsLeft() bool {
	return !o.isRight
}

func (o Or[L, R]) IsRight() bool {
	return o.isRight
}

// Left returns the left value and true, if the Or is Left.
func (o Or[L, R]) Left() (L, bool) {
	return o.l, !o.isRight
}

// Right returns the right value and true, if the Or is Right.
func (o Or[L, R]) Right() (R, bool) {
	return o.r, o.isRight
}

func (o Or[L, R]) MarshalPrim(optimized bool) (micheline.Prim, error) {
	if o.isRight {
		inner, err := MarshalPrim(o.r, optimized)
		if err != nil {
			return micheline.Prim{}, err
		}
		return micheline.NewCode(micheline.D_RIGHT, inner), nil
	} else {
		inner, err := MarshalPrim(o.l, optimized)
		if err != nil {
			return micheline.Prim{}, err
		}
		return micheline.NewCode(micheline.D_LEFT, inner), nil
	}
}

func (o *Or[L, R]) UnmarshalPrim(prim micheline.Prim) error {
	switch prim.OpCode {
	case micheline.D_LEFT:
		if len(prim.Args) != 1 {
			return errors.New("prim Left should have 1 arg")
		}
		o.isRight = false
		return UnmarshalPrim(prim.Args[0], &o.l)
	case micheline.D_RIGHT:
		if len(prim.Args) != 1 {
			return errors.New("prim Right should have 1 arg")
		}
		o.isRight = true
		return UnmarshalPrim(prim.Args[0], &o.r)
	default:
		return errors.Errorf("unexpected opCode when unmarshalling Or: %s", prim.OpCode)
	}
}

func (o Or[L, R]) keyHash() hashType {
	if l, ok := o.Left(); ok {
		valHash := hashFunc(zero[L]())(l)
		return hashBytes(append([]byte{0}, valHash[:]...))
	} else {
		r, _ := o.Right()
		valHash := hashFunc(zero[R]())(r)
		return hashBytes(append([]byte{1}, valHash[:]...))
	}
}
