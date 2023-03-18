package hash

import (
	"encoding/hex"
	"hash/fnv"
	"testing"
	"testing/quick"
)

// test implementation of FNV Hash64
func TestNewHash64(t *testing.T) {
	expect := func(data []byte) uint64 {
		h := fnv.New64a()
		if _, err := h.Write(data); err != nil {
			t.Fatal(err)
		}
		return h.Sum64()
	}
	hash := func(data []byte) uint64 {
		h := NewInlineFNV64a()
		h.Write(data)
		return h.Sum64()
	}
	if err := quick.CheckEqual(hash, expect, nil); err != nil {
		t.Fatal(err)
	}
}

func TestStaticHash64(t *testing.T) {
	expect := func(data []byte) uint64 {
		h := fnv.New64a()
		if _, err := h.Write(data); err != nil {
			t.Fatal(err)
		}
		return h.Sum64()
	}
	if err := quick.CheckEqual(Hash64, expect, nil); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkNewHash64(b *testing.B) {
	buf, _ := hex.DecodeString("029d4ed3161d644bedccb8673f30c6682b6e0a11756a3f75d7a739dede1cf29e")
	b.SetBytes(32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h := NewInlineFNV64a()
		h.Write(buf)
		_ = h.Sum64()
	}
}

func BenchmarkStaticHash64(b *testing.B) {
	buf, _ := hex.DecodeString("029d4ed3161d644bedccb8673f30c6682b6e0a11756a3f75d7a739dede1cf29e")
	b.SetBytes(32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Hash64(buf)
	}
}
