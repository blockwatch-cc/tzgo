// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"encoding/hex"
	"fmt"
	"io"
	"sync"
)

// HexBytes represents bytes as a JSON string of hexadecimal digits
type HexBytes []byte

// UnmarshalText umarshals a hex string to bytes. It implements the
// encoding.TextUnmarshaler interface, so that HexBytes can be used in Go
// structs in combination with the standard JSON library.
func (h *HexBytes) UnmarshalText(data []byte) error {
	dst := make([]byte, hex.DecodedLen(len(data)))
	if _, err := hex.Decode(dst, data); err != nil {
		return err
	}
	*h = dst
	return nil
}

// ReadBytes copies size bytes from r into h. If h is nil or too short,
// a new byte slice is allocated. Fails with io.ErrShortBuffer when
// less than size bytes can be read from r.
func (h *HexBytes) ReadBytes(r io.Reader, size int) error {
	if cap(*h) < size {
		*h = make([]byte, size)
	} else {
		*h = (*h)[:size]
	}
	n, err := r.Read(*h)
	if err != nil {
		return err
	}
	if n < size {
		return io.ErrShortBuffer
	}
	return nil
}

// MarshalText converts a byte slice to a hex encoded string. It implements the
// encoding.TextMarshaler interface, so that HexBytes can be used in Go
// structs in combination with the standard JSON library.
func (h HexBytes) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h)), nil
}

// String converts a byte slice to a hex encoded string
func (h HexBytes) String() string {
	return hex.EncodeToString(h)
}

// Bytes type-casts HexBytes back to a byte slice
func (h HexBytes) Bytes() []byte {
	return []byte(h)
}

// UnmarshalBinary umarshals a binary slice. It implements the
// encoding.BinaryUnmarshaler interface.
func (h *HexBytes) UnmarshalBinary(data []byte) error {
	if cap(*h) < len(data) {
		*h = make([]byte, len(data))
	}
	*h = (*h)[:len(data)]
	copy(*h, data)
	return nil
}

// MarshalBinary marshals as binary slice. It implements the
// encoding.BinaryMarshaler interface.
func (h HexBytes) MarshalBinary() ([]byte, error) {
	return h, nil
}

// Ratio represents a numeric ratio used in Ithaca constants
type Ratio struct {
	Num int `json:"numerator"`
	Den int `json:"denominator"`
}

func (r Ratio) Float64() float64 {
	if r.Den == 0 {
		return 0
	}
	return float64(r.Num) / float64(r.Den)
}

func Short(v any) string {
	var s string
	if str, ok := v.(fmt.Stringer); ok {
		s = str.String()
	} else {
		s = v.(string)
	}
	if len(s) <= 12 {
		return s
	}
	return s[:8] + "..." + s[len(s)-4:]
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func b2i(b bool) (i int) {
	// The compiler currently only optimizes this form into movzbl.
	// See https://0x0f.me/blog/golang-compiler-optimization/
	// See issue 6011.
	if b {
		i = 1
	}
	return
}

var bufPool32 = &sync.Pool{
	New: func() any { return make([]byte, 32) },
}
