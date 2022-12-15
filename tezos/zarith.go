// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

// little-endian zarith encoding
// https://github.com/ocaml/Zarith

package tezos

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"sync"
)

type Bool byte

const (
	False Bool = 0x00
	True  Bool = 0xff
)

func (b Bool) EncodeBuffer(buf *bytes.Buffer) error {
	buf.WriteByte(byte(b))
	return nil
}

func (b *Bool) DecodeBuffer(buf *bytes.Buffer) error {
	if buf.Len() < 1 {
		return io.ErrShortBuffer
	}
	if buf.Next(1)[0] == 0xff {
		*b = True
	} else {
		*b = False
	}
	return nil
}

// A variable length sequence of bytes, encoding a Zarith number.
// Each byte has a running unary size bit: the most significant bit
// of each byte tells if this is the last byte in the sequence (0)
// or if there is more to read (1). The second most significant bit
// of the first byte is reserved for the sign (positive if zero).
// Size and sign bits ignored, data is then the binary representation
// of the absolute value of the number in little endian order.
type Z big.Int

var Zero = NewZ(0)

func NewZ(i int64) Z {
	var z Z
	z.SetInt64(i)
	return z
}

func NewBigZ(b *big.Int) Z {
	var z Z
	z.SetBig(b)
	return z
}

func (z Z) Big() *big.Int {
	return (*big.Int)(&z)
}

func (z Z) Equal(x Z) bool {
	return z.Big().Cmp(x.Big()) == 0
}

func (z Z) IsZero() bool {
	return (*big.Int)(&z).IsInt64() && (*big.Int)(&z).Int64() == 0
}

func (z Z) Cmp(b Z) int {
	return (*big.Int)(&z).Cmp((*big.Int)(&b))
}

func (z Z) Int64() int64 {
	return (*big.Int)(&z).Int64()
}

func (z *Z) SetBig(b *big.Int) *Z {
	(*big.Int)(z).Set(b)
	return z
}

func (z *Z) SetInt64(i int64) *Z {
	(*big.Int)(z).SetInt64(i)
	return z
}

func (z Z) Clone() Z {
	var x Z
	x.SetBig(z.Big())
	return x
}

func (z *Z) UnmarshalBinary(data []byte) error {
	return z.DecodeBuffer(bytes.NewBuffer(data))
}

func (z *Z) DecodeBuffer(buf *bytes.Buffer) error {
	tmp := make([]byte, 16)
	var (
		b   byte
		err error
	)
	// read bits [0,6)
	if b, err = buf.ReadByte(); err != nil {
		return io.ErrShortBuffer
	}
	sign := b&0x40 > 0
	tmp[0] = b & 0x3f
	// read bits [6,62)
	for i := 1; i < 9; i++ {
		if b < 0x80 {
			break
		}
		if b, err = buf.ReadByte(); err != nil {
			return io.ErrShortBuffer
		}
		tmp[i] = b & 0x7f
	}

	w := int64(tmp[0]) | int64(tmp[1])<<6 | int64(tmp[2])<<13 | int64(tmp[3])<<20 | int64(tmp[4])<<27 |
		int64(tmp[5])<<34 | int64(tmp[6])<<41 | int64(tmp[7])<<48 | int64(tmp[8])<<55

	if b < 0x80 {
		z.SetInt64(w)
		if sign {
			(*big.Int)(z).Neg((*big.Int)(z))
		}
		return nil
	}

	binary.BigEndian.PutUint64(tmp[0:8], 0)
	tmp[8] = 0

	// read bits [62,125)
	for i := 0; i < 9; i++ {
		if b < 0x80 {
			break
		}
		if b, err = buf.ReadByte(); err != nil {
			return io.ErrShortBuffer
		}
		tmp[i] = b & 0x7f
	}

	w |= int64(tmp[0]) << 62
	w2 := int64(tmp[0])>>2 | int64(tmp[1])<<5 | int64(tmp[2])<<12 | int64(tmp[3])<<19 | int64(tmp[4])<<26 |
		int64(tmp[5])<<33 | int64(tmp[6])<<40 | int64(tmp[7])<<47 | int64(tmp[8])<<54

	binary.BigEndian.PutUint64(tmp[0:8], uint64(w2))
	binary.BigEndian.PutUint64(tmp[8:16], uint64(w))
	x := (*big.Int)(z).SetBytes(tmp[0:16])

	if b < 0x80 {
		if sign {
			x.Neg(x)
		}
		return nil
	}

	var s uint = 125
	y := bigIntPool.Get().(*big.Int)

	// read bits >=125
	for b >= 0x80 {
		binary.BigEndian.PutUint64(tmp[0:8], 0)
		tmp[8] = 0
		for i := 0; i < 9; i++ {
			if b < 0x80 {
				break
			}
			if b, err = buf.ReadByte(); err != nil {
				bigIntPool.Put(y)
				return io.ErrShortBuffer
			}
			tmp[i] = b & 0x7f
		}

		w := int64(tmp[0]) | int64(tmp[1])<<7 | int64(tmp[2])<<14 | int64(tmp[3])<<21 | int64(tmp[4])<<28 |
			int64(tmp[5])<<35 | int64(tmp[6])<<42 | int64(tmp[7])<<49 | int64(tmp[8])<<56

		y.SetInt64(w)

		x = x.Or(x, y.Lsh(y, s))
		s += 63
	}

	bigIntPool.Put(y)

	if sign {
		x.Neg(x)
	}
	return nil
}

func (z Z) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := z.EncodeBuffer(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (z *Z) EncodeBuffer(buf *bytes.Buffer) error {
	xi := bigIntPool.Get()
	x := xi.(*big.Int).Set((*big.Int)(z))
	yi := bigIntPool.Get()
	y := yi.(*big.Int).SetInt64(0)
	var sign byte
	if x.Sign() < 0 {
		sign = 0x40
		x.Neg(x)
	}
	if x.IsInt64() && x.Int64() < 0x40 {
		buf.WriteByte(byte(x.Int64()) | sign)
		bigIntPool.Put(xi)
		bigIntPool.Put(yi)
		return nil
	} else {
		buf.WriteByte(byte(y.And(x, mask3f).Int64()) | 0x80 | sign)
		x.Rsh(x, 6)
	}

	for !x.IsInt64() || x.Int64() >= 0x80 {
		buf.WriteByte(byte(y.And(x, mask7f).Int64()) | 0x80)
		x.Rsh(x, 7)
	}
	buf.WriteByte(byte(x.Int64()))
	bigIntPool.Put(xi)
	bigIntPool.Put(yi)
	return nil
}

func ParseZ(s string) (Z, error) {
	var z Z
	err := (*big.Int)(&z).UnmarshalText([]byte(s))
	return z, err
}

func MustParseZ(s string) Z {
	z, err := ParseZ(s)
	if err != nil {
		panic(err)
	}
	return z
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (z *Z) Set(val string) (err error) {
	*z, err = ParseZ(val)
	return
}

func (z Z) MarshalText() ([]byte, error) {
	return (*big.Int)(&z).MarshalText()
}

func (z *Z) UnmarshalText(d []byte) error {
	return (*big.Int)(z).UnmarshalText(d)
}

func (z Z) String() string {
	return (*big.Int)(&z).Text(10)
}

func (z Z) Bytes() []byte {
	buf, _ := z.MarshalBinary()
	return buf
}

func (z Z) Decimals(d int) string {
	s := z.String()
	if d <= 0 {
		return s
	}
	l := len(s)
	if l <= d {
		s = strings.Repeat("0", d-l+1) + s
	}
	l = len(s)
	return s[:l-d] + "." + s[l-d:]
}

func (z Z) Neg() Z {
	var n Z
	n.SetBig(new(big.Int).Neg(z.Big()))
	return n
}

func (z Z) Add(y Z) Z {
	var x Z
	x.SetBig(new(big.Int).Add(z.Big(), y.Big()))
	return x
}

func (z Z) Sub(y Z) Z {
	var x Z
	x.SetBig(new(big.Int).Sub(z.Big(), y.Big()))
	return x
}

func (z Z) Mul(y Z) Z {
	var x Z
	x.SetBig(new(big.Int).Mul(z.Big(), y.Big()))
	return x
}

func (z Z) Div(y Z) Z {
	var x Z
	if !y.IsZero() {
		x.SetBig(new(big.Int).Div(z.Big(), y.Big()))
	}
	return x
}

func (z Z) Add64(y int64) Z {
	var x Z
	x.SetBig(new(big.Int).Add(z.Big(), big.NewInt(y)))
	return x
}

func (z Z) Sub64(y int64) Z {
	var x Z
	x.SetBig(new(big.Int).Sub(z.Big(), big.NewInt(y)))
	return x
}

func (z Z) Mul64(y int64) Z {
	var x Z
	x.SetBig(new(big.Int).Mul(z.Big(), big.NewInt(y)))
	return x
}

func (z Z) Div64(y int64) Z {
	var x Z
	if y != 0 {
		x.SetBig(new(big.Int).Div(z.Big(), big.NewInt(y)))
	}
	return x
}

func (z Z) IsNeg() bool {
	return z.Big().Sign() < 0
}

var (
	mask3f     = big.NewInt(0x3f)
	mask7f     = big.NewInt(0x7f)
	bigIntPool = &sync.Pool{
		New: func() interface{} { return big.NewInt(0) },
	}
)

// A variable length sequence of bytes, encoding a Zarith number.
// Each byte has a running unary size bit: the most significant bit
// of each byte tells is this is the last byte in the sequence (0)
// or if there is more to read (1). Size bits ignored, data is then
// the binary representation of the absolute value of the number in
// little endian order.
type N int64

func NewN(i int64) N {
	return N(i)
}

func (n N) Equal(x N) bool {
	return n == x
}

func (n N) IsZero() bool {
	return n == 0
}

func (n N) Int64() int64 {
	return int64(n)
}

func (n *N) SetInt64(i int64) *N {
	*n = N(i)
	return n
}

func (n N) Clone() N {
	return n
}

func (n *N) DecodeBuffer(buf *bytes.Buffer) error {
	var (
		x int64
		s uint
	)
	for i := 0; ; i++ {
		b := buf.Next(1)
		if len(b) == 0 {
			return io.ErrShortBuffer
		}
		if b[0] < 0x80 {
			if i > 9 || i == 9 && b[0] > 1 {
				return fmt.Errorf("tezos: numeric overflow")
			}
			x |= int64(b[0]) << s
			break
		}
		x |= int64(b[0]&0x7f) << s
		s += 7
	}
	*n = N(x)
	return nil
}

func (n N) EncodeBuffer(buf *bytes.Buffer) error {
	x := int64(n)
	for x >= 0x80 {
		buf.WriteByte(byte(x) | 0x80)
		x >>= 7
	}
	buf.WriteByte(byte(x))
	return nil
}

func (n *N) UnmarshalBinary(data []byte) error {
	return n.DecodeBuffer(bytes.NewBuffer(data))
}

func (n N) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := n.EncodeBuffer(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (n N) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(n), 10)), nil
}

func (n *N) UnmarshalText(d []byte) error {
	i, err := strconv.ParseInt(string(d), 10, 64)
	if err != nil {
		return err
	}
	*n = N(i)
	return nil
}

func (n N) String() string {
	return strconv.FormatInt(int64(n), 10)
}

func (n N) Decimals(d int) string {
	s := n.String()
	if d <= 0 {
		return s
	}
	l := len(s)
	if l <= d {
		s = strings.Repeat("0", d-l+1) + s
	}
	l = len(s)
	return s[:l-d] + "." + s[l-d:]
}

func ParseN(s string) (N, error) {
	i, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		return N(0), err
	}
	return N(i), nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (n *N) Set(val string) (err error) {
	*n, err = ParseN(val)
	return
}
