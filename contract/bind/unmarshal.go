package bind

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
	"github.com/pkg/errors"
	"math/big"
	"reflect"
	"time"
)

// UnmarshalPrimPath unmarshals a nested prim contained in root, obtained using the given path,
// into v.
//
// v must be a non-nil pointer to the expected result.
func UnmarshalPrimPath(root micheline.Prim, path string, v any) error {
	prim, err := root.GetPath(path)
	if err != nil {
		return errors.Wrap(err, "failed to get path")
	}
	return UnmarshalPrim(prim, v)
}

// UnmarshalPrimPaths unmarshals a Prim into a map of (path => destination).
func UnmarshalPrimPaths(root micheline.Prim, dst map[string]any) error {
	for path, v := range dst {
		if err := UnmarshalPrimPath(root, path, v); err != nil {
			return errors.Wrapf(err, "failed to unmarshal for path %s", path)
		}
	}
	return nil
}

// UnmarshalPrim unmarshals a prim into v.
//
// v must be a non-nil pointer to the expected result.
func UnmarshalPrim(prim micheline.Prim, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("v should be a non-nil pointer")
	}

	return unmarshalPrimVal(prim, val)
}

var _ micheline.PrimUnmarshaler = &Bigmap[string, []byte]{}

func unmarshalPrimVal(prim micheline.Prim, val reflect.Value) error {
	// Init Elem value if val is Nil
	if val.Kind() == reflect.Ptr && val.IsNil() {
		val.Set(reflect.New(val.Type().Elem()))
	}

	// Check PrimUnmarshaler interface first
	if unmarshaler, ok := val.Interface().(micheline.PrimUnmarshaler); ok {
		return unmarshaler.UnmarshalPrim(prim)
	} else if unmarshaler, ok = val.Elem().Interface().(micheline.PrimUnmarshaler); ok {
		return unmarshalPrimVal(prim, val.Elem())
	}

	// Check scalar and container types
	switch prim.Type {
	case micheline.PrimInt:
		return unmarshalInt(prim.Int, val)
	case micheline.PrimString:
		return unmarshalString(prim.String, val)
	case micheline.PrimBytes:
		return unmarshalBytes(prim.Bytes, val)
	case micheline.PrimSequence:
		return unmarshalSlice(prim, val)
	case micheline.PrimNullary:
		switch prim.OpCode {
		case micheline.D_TRUE, micheline.D_FALSE:
			return unmarshalBool(prim.OpCode == micheline.D_TRUE, val)
		case micheline.D_UNIT:
			// v should be a struct{}, so we don't have to set anything
			return nil
		}
	}

	// Wasn't handled with *T, try with T
	if val.Kind() == reflect.Ptr {
		return unmarshalPrimVal(prim, val.Elem())
	}

	return errors.Errorf("prim type not handled: %s", prim.Type)
}

func unmarshalInt(i *big.Int, v reflect.Value) error {
	if v.Kind() == reflect.Ptr && v.Type() != tBigInt {
		return unmarshalInt(i, v.Elem())
	}

	switch v.Type() {
	case tBigInt:
		v.Set(reflect.ValueOf(i))
	case tTime:
		v.Set(reflect.ValueOf(time.Unix(i.Int64(), 0)))
	default:
		return errors.Errorf("unexpected type for int prim: %T", v.Type())
	}

	return nil
}

func unmarshalString(str string, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		return unmarshalString(str, v.Elem())
	}

	switch v.Type() {
	case tString:
		v.SetString(str)
	case tTime:
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return errors.Wrapf(err, "failed to parse time: %s", str)
		}
		v.Set(reflect.ValueOf(t))
	case tAddress:
		address, err := tezos.ParseAddress(str)
		if err != nil {
			return errors.Wrapf(err, "failed to parse address: %s", str)
		}
		v.Set(reflect.ValueOf(address))
	case tKey:
		key, err := tezos.ParseKey(str)
		if err != nil {
			return errors.Wrapf(err, "failed to parse key: %s", str)
		}
		v.Set(reflect.ValueOf(key))
	case tSignature:
		sig, err := tezos.ParseSignature(str)
		if err != nil {
			return errors.Wrapf(err, "failed to parse signature: %s", str)
		}
		v.Set(reflect.ValueOf(sig))
	case tChainIdHash:
		chainID, err := tezos.ParseChainIdHash(str)
		if err != nil {
			return errors.Wrapf(err, "failed to parse chainID: %s", str)
		}
		v.Set(reflect.ValueOf(chainID))
	default:
		return errors.Errorf("unexpected type for string prim: %T", v.Type())
	}
	return nil
}

func unmarshalBytes(b []byte, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		return unmarshalBytes(b, v.Elem())
	}

	switch v.Type() {
	case tBytes:
		v.Set(reflect.ValueOf(b))
	case tAddress:
		var addr tezos.Address
		if err := addr.UnmarshalBinary(b); err != nil {
			return errors.Wrapf(err, "failed to parse address: %v", b)
		}
		v.Set(reflect.ValueOf(addr))
	case tKey:
		var key tezos.Key
		if err := key.UnmarshalBinary(b); err != nil {
			return errors.Wrapf(err, "failed to parse key: %v", b)
		}
		v.Set(reflect.ValueOf(key))
	case tSignature:
		var sig tezos.Signature
		if err := sig.UnmarshalBinary(b); err != nil {
			return errors.Wrapf(err, "failed to parse key: %v", b)
		}
		v.Set(reflect.ValueOf(sig))
	case tChainIdHash:
		var chainID tezos.ChainIdHash
		if err := chainID.UnmarshalBinary(b); err != nil {
			return errors.Wrapf(err, "failed to parse chainID: %v", b)
		}
		v.Set(reflect.ValueOf(chainID))
	default:
		return errors.Errorf("unexpected type for bytes prim: %T", v.Type())
	}
	return nil
}

func unmarshalBool(b bool, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		return unmarshalBool(b, v.Elem())
	}

	if v.Type() != tBool {
		return errors.Errorf("unexpected type for bool prim: %T", v.Type())
	}

	v.Set(reflect.ValueOf(b))
	return nil
}

// unmarshalSlice unmarshals a Prim sequence, into a slice (through a reflect.Value).
func unmarshalSlice(prim micheline.Prim, v reflect.Value) error {
	if v.Kind() == reflect.Ptr {
		return unmarshalSlice(prim, v.Elem())
	}

	if v.Kind() != reflect.Slice {
		return errors.Errorf("unexpected type for sequence prim: %T", v.Type())
	}

	elemType := v.Type().Elem()

	// probably improvable
	lenArgs := len(prim.Args)
	newSlice := reflect.MakeSlice(v.Type(), lenArgs, lenArgs)
	for i, arg := range prim.Args {
		elem := reflect.New(elemType)
		err := UnmarshalPrim(arg, elem.Interface())
		if err != nil {
			return err
		}
		newSlice.Index(i).Set(elem.Elem())
	}
	v.Set(newSlice)

	return nil
}

// pre-computed reflect types
var (
	tBigInt      = reflect.TypeOf((*big.Int)(nil))
	tString      = reflect.TypeOf("")
	tBytes       = reflect.TypeOf(([]byte)(nil))
	tBool        = reflect.TypeOf(true)
	tTime        = reflect.TypeOf(time.Time{})
	tAddress     = reflect.TypeOf(tezos.Address{})
	tKey         = reflect.TypeOf(tezos.Key{})
	tSignature   = reflect.TypeOf(tezos.Signature{})
	tChainIdHash = reflect.TypeOf(tezos.ChainIdHash{})
)
