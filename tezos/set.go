// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"blockwatch.cc/tzgo/hash"
)

type AddressSet struct {
	set map[uint64]Address
}

func NewAddressSet(addrs ...Address) *AddressSet {
	set := &AddressSet{
		set: make(map[uint64]Address),
	}
	for _, v := range addrs {
		if !v.IsValid() {
			continue
		}
		set.AddUnique(v)
	}
	return set
}

func (s AddressSet) hash(addr Address) uint64 {
	h := hash.NewInlineFNV64a()
	h.Write([]byte{byte(addr.Type)})
	h.Write(addr.Hash)
	return h.Sum64()
}

func (s *AddressSet) AddUnique(addr Address) bool {
	h := s.hash(addr)
	_, ok := s.set[h]
	s.set[h] = addr.Clone()
	return !ok
}

func (s *AddressSet) Add(addr Address) {
	s.set[s.hash(addr)] = addr.Clone()
}

func (s *AddressSet) Remove(addr Address) {
	delete(s.set, s.hash(addr))
}

func (s AddressSet) Contains(addr Address) bool {
	_, ok := s.set[s.hash(addr)]
	return ok
}

func (s *AddressSet) Merge(b *AddressSet) {
	for n, v := range b.set {
		s.set[n] = v.Clone()
	}
}

func (s *AddressSet) Len() int {
	return len(s.set)
}

func (s AddressSet) Map() map[uint64]Address {
	return s.set
}

func (s AddressSet) Slice() []Address {
	if len(s.set) == 0 {
		return nil
	}
	a := make([]Address, 0, len(s.set))
	for _, v := range s.Map() {
		a = append(a, v)
	}
	return a
}
