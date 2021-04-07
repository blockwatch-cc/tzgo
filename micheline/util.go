// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"strconv"
	"strings"
	"unicode"
)

const PATH_SEPARATOR = "."

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 32 || s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func isASCIIBytes(b []byte) bool {
	return isASCII(string(b))
}

func isPackedBytes(b []byte) bool {
	return len(b) > 1 && b[0] == 0x5 && b[1] <= 0x0A // first primitive is valid
}

func limit(s string, l int) string {
	return s[:min(len(s), l)]
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func getPath(val interface{}, path string) (interface{}, bool) {
	if val == nil {
		return nil, false
	}
	if path == "" {
		return val, true
	}
	frag := strings.Split(path, PATH_SEPARATOR)
	next := val
	for i, v := range frag {
		switch t := next.(type) {
		case map[string]interface{}:
			var ok bool
			next, ok = t[v]
			if !ok {
				return nil, false
			}
		case []interface{}:
			idx, err := strconv.Atoi(v)
			if err != nil || idx < 0 || len(t) < idx {
				return nil, false
			}
			next = t[idx]
		default:
			return next, i == len(frag)-1
		}
	}
	return next, true
}

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
