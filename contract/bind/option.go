package bind

import (
	"blockwatch.cc/tzgo/micheline"
	"fmt"
	"github.com/pkg/errors"
)

// Option is a type that can either contain a value T, or be None.
type Option[T any] struct {
	v      T
	isSome bool
}

// Some returns a Some option with v as a value.
func Some[T any](v T) Option[T] {
	return Option[T]{v: v, isSome: true}
}

// None returns a None option for type T.
func None[T any]() Option[T] {
	return Option[T]{isSome: false}
}

// Get returns the inner value of the Option and a boolean
// indicating if the Option is Some.
//
// If it is none, the returned value is the default value for T.
func (o Option[T]) Get() (v T, isSome bool) {
	return o.v, o.isSome
}

// Unwrap returns the inner value of the Option, expecting
// that it is Some.
//
// Panics if the option is None.
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("Unwrap() called on a `None` Option")
	}
	return o.v
}

// UnwrapOr returns the inner value of the Option if it is Some,
// or the provided default value if it is None.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.IsNone() {
		return defaultValue
	}
	return o.v
}

// UnwrapOrZero returns the inner value of the Option if it is Some,
// or T's zero value if it is None.
func (o Option[T]) UnwrapOrZero() T {
	// o.v == zero value if o is none
	return o.v
}

// SetSome replaces o's value with Some(v).
func (o *Option[T]) SetSome(v T) {
	o.v = v
	o.isSome = true
}

// SetNone replaces o's value with None.
func (o *Option[T]) SetNone() {
	var zeroVal T
	o.v = zeroVal
	o.isSome = false
}

func (o Option[T]) IsSome() bool {
	return o.isSome
}

func (o Option[T]) IsNone() bool {
	return !o.isSome
}

// GetUntyped is the equivalent of Get, but it returns v as an empty interface
// instead of T.
//
// This method is useful when generic parameters cannot be used, for example with reflection.
func (o Option[T]) GetUntyped() (v any, isSome bool) {
	return o.Get()
}

// SetUntyped is the equivalent of SetSome or SetNone, but uses v as an empty interface
// instead of T.
//
// If v is nil, then the Option will be set to None.
// Else, it will cast v to T and set the Option to Some.
//
// Returns an error if the cast failed.
//
// This method is useful when generic parameters cannot be used, for example with reflection.
func (o *Option[T]) SetUntyped(v any) error {
	if v == nil {
		o.SetNone()
		return nil
	}
	casted, ok := v.(T)
	if !ok {
		return errors.Errorf("bad type (want %T, got %T)", o.v, v)
	}
	o.SetSome(casted)
	return nil
}

func (o Option[T]) String() string {
	if o.isSome {
		return fmt.Sprintf("Some(%v)", o.v)
	}
	return "None"
}

func (o Option[T]) MarshalPrim(optimized bool) (micheline.Prim, error) {
	if o.isSome {
		inner, err := MarshalPrim(o.v, optimized)
		if err != nil {
			return micheline.Prim{}, err
		}
		return micheline.NewCode(micheline.D_SOME, inner), nil
	}
	return micheline.NewCode(micheline.D_NONE), nil
}

func (o *Option[T]) UnmarshalPrim(prim micheline.Prim) error {
	switch prim.OpCode {
	case micheline.D_SOME:
		if len(prim.Args) != 1 {
			return errors.New("prim Some should have 1 arg")
		}
		o.isSome = true
		return UnmarshalPrim(prim.Args[0], &o.v)
	case micheline.D_NONE:
		*o = None[T]()
		return nil
	default:
		return errors.Errorf("unexpected opCode when unmarshalling Option: %s", prim.OpCode)
	}
}

func (o Option[T]) keyHash() hashType {
	if v, ok := o.Get(); ok {
		return hashFunc(zero[T]())(v)
	} else {
		return hashType{0}
	}
}
