// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"blockwatch.cc/tzgo/hash"
)

type AddressFilter struct {
	set map[uint64]struct{}
}

func NewAddressFilter(addrs ...Address) *AddressFilter {
	set := &AddressFilter{
		set: make(map[uint64]struct{}),
	}
	for _, v := range addrs {
		set.Add(v)
	}
	return set
}

func (s AddressFilter) hash(addr Address) uint64 {
	h := hash.NewInlineFNV64a()
	h.Write([]byte{byte(addr.Type)})
	h.Write(addr.Hash)
	return h.Sum64()
}

func (s *AddressFilter) AddUnique(addr Address) bool {
	h := s.hash(addr)
	_, ok := s.set[h]
	s.set[h] = struct{}{}
	return !ok
}

func (s *AddressFilter) Add(addr Address) {
	s.set[s.hash(addr)] = struct{}{}
}

func (s *AddressFilter) Remove(addr Address) {
	delete(s.set, s.hash(addr))
}

func (s AddressFilter) Contains(addr Address) bool {
	_, ok := s.set[s.hash(addr)]
	return ok
}

func (s *AddressFilter) Merge(b *AddressFilter) {
	for n, _ := range b.set {
		s.set[n] = struct{}{}
	}
}

func (s *AddressFilter) Len() int {
	return len(s.set)
}
