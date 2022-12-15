// Copyright (c) 2022 Blockwatch Data Inc.
// Author: stefan@blockwatch.cc

package tezos

import (
	"bytes"
	"io"
	"math/big"
	"math/rand"
	"testing"
)

type ZarithDecodeTest struct {
	name string
	buf  []byte
	res  []uint64
	sign int
	err  error
}

var zarithDecodeCases = []ZarithDecodeTest{
	{
		name: "e0",
		buf:  []byte{},
		err:  io.ErrShortBuffer,
	},
	{
		name: "e1",
		buf:  []byte{0xc0},
		err:  io.ErrShortBuffer,
	},
	{
		name: "e9",
		buf:  []byte{0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0},
		err:  io.ErrShortBuffer,
	},
	{
		name: "l1",
		buf:  []byte{0x20},
		res:  []uint64{0x20},
	},
	{
		name: "l9",
		buf:  []byte{0xa0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x2040810204081020},
	},
	{
		name: "l10",
		buf:  []byte{0xa0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x10, 0x2040810204081020},
	},
	{
		name: "l18",
		buf:  []byte{0xa0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x1020408102040810, 0x2040810204081020},
	},
	{
		name: "l19",
		buf:  []byte{0xa0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x08, 0x1020408102040810, 0x2040810204081020},
	},
	{
		name: "n1",
		buf:  []byte{0x60},
		res:  []uint64{0x20},
		sign: 1,
	},
	{
		name: "n9",
		buf:  []byte{0xe0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x2040810204081020},
		sign: 1,
	},
	{
		name: "n10",
		buf:  []byte{0xe0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x10, 0x2040810204081020},
		sign: 1,
	},
	{
		name: "n18",
		buf:  []byte{0xe0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x1020408102040810, 0x2040810204081020},
		sign: 1,
	},
	{
		name: "n19",
		buf:  []byte{0xe0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0xc0, 0x40},
		res:  []uint64{0x08, 0x1020408102040810, 0x2040810204081020},
		sign: 1,
	},
}

func TestDecodeBuffer(t *testing.T) {
	for _, c := range zarithDecodeCases {
		var z Z
		err := z.DecodeBuffer(bytes.NewBuffer(c.buf))
		if got, want := err, c.err; got != want {
			t.Errorf("%s: unexpected error %v, expected %v", c.name, got, want)
		}
		if err != nil {
			continue
		}
		res := big.NewInt(0)
		n := new(big.Int)
		for _, v := range c.res {
			n.SetUint64(v)
			res.Or(res.Lsh(res, 64), n)
		}
		if c.sign != 0 {
			res.Neg(res)
		}
		if got, want := z, (Z)(*res); got.Cmp(want) != 0 {
			t.Errorf("%s: unexpected result %v, expected %v", c.name, got, want)
		}
	}
}

type benchmarkSize struct {
	name string
	l    int
}

var benchmarkSizes = []benchmarkSize{
	{"6bit", 1},
	{"62bit", 9},
	{"125bit", 18},
	{"251bit", 36},
	{"510bit", 73},
}

func randZarithSlice(n int) []byte {
	s := make([]byte, n)
	if n == 1 {
		s[0] = byte(rand.Intn(0x40))
		return s
	}

	s[0] = byte(rand.Intn(0x40)) | 0x80
	for i := 1; i < n-1; i++ {
		s[i] = byte(rand.Intn(0x80)) | 0x80
	}
	s[n-1] = byte(rand.Intn(0x80))
	return s
}

func BenchmarkDecodeBuffer(b *testing.B) {
	for _, bm := range benchmarkSizes {
		buf := randZarithSlice(bm.l)
		var z Z
		b.Run(bm.name, func(b *testing.B) {
			b.SetBytes(int64(bm.l))
			for i := 0; i < b.N; i++ {
				z.DecodeBuffer(bytes.NewBuffer(buf))
			}
		})
	}
}
