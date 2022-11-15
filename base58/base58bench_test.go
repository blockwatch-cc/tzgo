// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Copyright (c) 2013-2014 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package base58_test

import (
	"bytes"
	"fmt"
	"testing"

	"blockwatch.cc/tzgo/base58"
)

var (
	sizes = []int{20, 32, 50, 100}
)

func BenchmarkEncodeBigInt(b *testing.B) {
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("size_%d", sz), func(b *testing.B) {
			data := bytes.Repeat([]byte{0xff}, sz)
			b.SetBytes(int64(sz))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				base58.Encode(data)
			}
		})
	}
}

func BenchmarkDecodeBigInt(b *testing.B) {
	for _, sz := range sizes {
		b.Run(fmt.Sprintf("size_%d", sz), func(b *testing.B) {
			data := bytes.Repeat([]byte{0xff}, sz)
			enc := base58.Encode(data)
			b.SetBytes(int64(len(enc)))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				base58.Decode(enc, nil)
			}
		})
	}
}
