// Copyright (c) 2018 - 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package hash

import (
	"bytes"
)

type HashMap map[uint64][]byte

func NewHashMap() HashMap {
	return make(HashMap)
}

func (m *HashMap) Add(buf []byte) int {
	(*m)[Hash64(buf)] = buf
	return len(*m)
}

func (m *HashMap) Remove(buf []byte) int {
	h := Hash64(buf)
	if b, ok := (*m)[h]; ok {
		if bytes.Equal(b, buf) {
			delete((*m), h)
		}
	}
	return len(*m)
}

func (m HashMap) Contains(buf []byte) bool {
	b, ok := m[Hash64(buf)]
	return ok && bytes.Equal(b, buf)
}
