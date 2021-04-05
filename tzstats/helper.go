// Copyright (c) 2013-2020 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

var stringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

func ToString(t interface{}) string {
	val := reflect.Indirect(reflect.ValueOf(t))
	if !val.IsValid() {
		return ""
	}
	if val.Type().Implements(stringerType) {
		return t.(fmt.Stringer).String()
	}
	if s, err := ToRawString(val.Interface()); err == nil {
		return s
	}
	return fmt.Sprintf("%v", val.Interface())
}

func IsBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

func ToRawString(t interface{}) (string, error) {
	val := reflect.Indirect(reflect.ValueOf(t))
	if !val.IsValid() {
		return "", nil
	}
	typ := val.Type()
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'g', -1, val.Type().Bits()), nil
	case reflect.String:
		return val.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), nil
	case reflect.Array:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		// [...]byte
		var b []byte
		if val.CanAddr() {
			b = val.Slice(0, val.Len()).Bytes()
		} else {
			b = make([]byte, val.Len())
			reflect.Copy(reflect.ValueOf(b), val)
		}
		return hex.EncodeToString(b), nil
	case reflect.Slice:
		if typ.Elem().Kind() != reflect.Uint8 {
			break
		}
		// []byte
		b := val.Bytes()
		return hex.EncodeToString(b), nil
	case reflect.Map:
		return fmt.Sprintf("%#v", t), nil
	}
	return "", fmt.Errorf("no method for converting type %s (%v) to string", typ.String(), val.Kind())
}

type ValueWalkerFunc func(path string, value interface{}) error

func walkValueMap(name string, val interface{}, fn ValueWalkerFunc) error {
	switch t := val.(type) {
	case map[string]interface{}:
		if len(name) > 0 {
			name += "."
		}
		for n, v := range t {
			child := name + n
			if err := walkValueMap(child, v, fn); err != nil {
				return err
			}
		}
	case []interface{}:
		if len(name) > 0 {
			name += "."
		}
		for i, v := range t {
			child := name + strconv.Itoa(i)
			if err := walkValueMap(child, v, fn); err != nil {
				return err
			}
		}
	default:
		return fn(name, val)
	}
	return nil
}

// Access nested map or array contents
func getPathString(val interface{}, path string) (string, bool) {
	if val == nil {
		return "", false
	}
	frag := strings.Split(path, ".")
	next := val
	for i, v := range frag {
		switch t := next.(type) {
		case map[string]interface{}:
			var ok bool
			next, ok = t[v]
			if !ok {
				return "", false
			}
		case []interface{}:
			idx, err := strconv.Atoi(v)
			if err != nil || len(t) < idx {
				return "", false
			}
			next = t[idx]
		default:
			return ToString(next), i == len(frag)-1
		}
	}
	return ToString(next), true
}

func getPathInt64(val interface{}, path string) (int64, bool) {
	str, ok := getPathString(val, path)
	if !ok {
		return 0, ok
	}
	i, err := strconv.ParseInt(str, 10, 64)
	return i, err == nil
}

func getPathBig(val interface{}, path string) (*big.Int, bool) {
	str, ok := getPathString(val, path)
	if !ok {
		return nil, ok
	}
	n := new(big.Int)
	_, err := fmt.Sscan(str, n)
	return n, err == nil
}

func getPathTime(val interface{}, path string) (time.Time, bool) {
	str, ok := getPathString(val, path)
	if !ok {
		return time.Time{}, ok
	}
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		// try parse as UNIX seconds
		if i, err2 := strconv.ParseInt(str, 10, 64); err2 == nil {
			t = time.Unix(i, 0)
			err = nil
		}
	}
	return t, err == nil
}

func getPathAddress(val interface{}, path string) (tezos.Address, bool) {
	str, ok := getPathString(val, path)
	if !ok {
		return tezos.InvalidAddress, ok
	}
	a, err := tezos.ParseAddress(str)
	return a, err == nil
}

func getPathValue(val interface{}, path string) (interface{}, bool) {
	if tree, ok := val.(map[string]interface{}); ok {
		frag := strings.Split(path, ".")
		for i, v := range frag {
			next, ok := tree[v]
			if !ok {
				return nil, false
			}
			switch t := next.(type) {
			case map[string]interface{}:
				tree = t
			default:
				return next, i == len(frag)-1
			}
		}
		return tree, true
	} else {
		return val, path == ""
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func nonNil(vals ...interface{}) interface{} {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}
