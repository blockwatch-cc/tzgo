package bind

import (
	"math/big"
	"reflect"
	"time"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
	"github.com/pkg/errors"
)

type PrimMarshaler interface {
	MarshalPrim(optimized bool) (micheline.Prim, error)
}

// MarshalPrim marshals v into a Prim by using reflection.
//
// If true, timestamps ,addresses, keys and signatures will be
// marshaled in their optimized format.
// See https://tezos.gitlab.io/active/michelson.html#differences-with-the-formal-notation.
func MarshalPrim(v any, optimized bool) (micheline.Prim, error) {
	// Handle types that we can process with a type switch
	switch t := v.(type) {
	case micheline.Prim:
		return t, nil
	case PrimMarshaler:
		return t.MarshalPrim(optimized)
	case *big.Int:
		return micheline.NewBig(t), nil
	case string:
		return micheline.NewString(t), nil
	case bool:
		if t {
			return micheline.NewCode(micheline.D_TRUE), nil
		}
		return micheline.NewCode(micheline.D_FALSE), nil
	case []byte:
		return micheline.NewBytes(t), nil
	case time.Time:
		if optimized {
			return micheline.NewInt64(t.Unix()), nil
		}
		return micheline.NewString(t.Format(time.RFC3339)), nil
	case tezos.Address:
		if optimized {
			return micheline.NewAddress(t), nil
		}
		return micheline.NewString(t.String()), nil
	case tezos.Key:
		if optimized {
			return micheline.NewBytes(t.Bytes()), nil
		}
		return micheline.NewString(t.String()), nil
	case tezos.Signature:
		if optimized {
			return micheline.NewBytes(t.Bytes()), nil
		}
		return micheline.NewString(t.String()), nil
	case tezos.ChainIdHash:
		return micheline.NewString(t.String()), nil
	}

	// Container types
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Slice:
		n := val.Len()
		prims := make([]micheline.Prim, 0, n)
		for i := 0; i < n; i++ {
			prim, err := MarshalPrim(val.Index(i).Interface(), optimized)
			if err != nil {
				return micheline.Prim{}, err
			}
			prims = append(prims, prim)
		}
		return micheline.NewSeq(prims...), nil
	}

	return micheline.Prim{}, errors.Errorf("type not handled: %T", v)
}

// MarshalParams marshals the provided params into a folded Prim.
func MarshalParams(optimized bool, params ...any) (micheline.Prim, error) {
	if len(params) == 0 {
		return micheline.NewPrim(micheline.D_UNIT), nil
	}

	prims := make([]micheline.Prim, 0, len(params))
	for _, p := range params {
		prim, err := MarshalPrim(p, optimized)
		if err != nil {
			return micheline.Prim{}, err
		}
		prims = append(prims, prim)
	}

	return foldRightComb(prims...), nil
}

// foldRightComb folds a list of prims into nested Pairs, by following the right-comb convention.
func foldRightComb(prims ...micheline.Prim) micheline.Prim {
	n := len(prims)
	switch n {
	case 0:
		return micheline.NewPrim(micheline.D_UNIT)
	case 1:
		return prims[0]
	default:
		return foldRightComb(append(prims[:n-2], micheline.NewPair(prims[n-2], prims[n-1]))...)
	}
}
