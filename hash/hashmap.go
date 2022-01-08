// Copyright (c) 2018 - 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
//
package hash

import (
	"bytes"
	"sync"
)

var fnvPool = sync.Pool{
	New: func() interface{} { return NewInlineFNV64a() },
}

type HashMap map[uint64][]byte

func NewHashMap() HashMap {
	return make(HashMap)
}

func (m *HashMap) Add(buf []byte) int {
	(*m)[fnv(buf)] = buf
	return len(*m)
}

func (m *HashMap) Remove(buf []byte) int {
	h := fnv(buf)
	if b, ok := (*m)[h]; ok {
		if bytes.Compare(b, buf) == 0 {
			delete((*m), h)
		}
	}
	return len(*m)
}

func (m HashMap) Contains(buf []byte) bool {
	b, ok := m[fnv(buf)]
	return ok && bytes.Compare(b, buf) == 0
}

func fnv(buf []byte) uint64 {
	ptr := fnvPool.Get()
	defer fnvPool.Put(ptr)
	h := ptr.(*InlineFNV64a)
	h.Reset()
	h.Write(buf)
	return h.Sum64()
}
