// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"blockwatch.cc/tzgo/hash"
)

type AddressSet struct {
	set  map[uint64]Address
	coll []Address
}

func NewAddressSet(addrs ...Address) *AddressSet {
	set := &AddressSet{
		set: make(map[uint64]Address, len(addrs)),
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
	a, ok := s.set[h]
	if ok {
		if !a.Equal(addr) {
			// hash collision
			s.coll = append(s.coll, addr.Clone())
			ok = false
		}
	} else {
		s.set[h] = addr.Clone()
	}
	return !ok
}

func (s *AddressSet) Add(addr Address) {
	s.AddUnique(addr)
}

func (s *AddressSet) Remove(addr Address) {
	delete(s.set, s.hash(addr))
	for i := range s.coll {
		if !s.coll[i].Equal(addr) {
			continue
		}
		s.coll = append(s.coll[:idx], s.coll[idx+1:]...)
		return
	}
}

func (s AddressSet) Contains(addr Address) bool {
	a, ok := s.set[s.hash(addr)]
	if !ok {
		return false
	}
	if a.Equal(addr) {
		return true
	}
	for _, v := range s.coll {
		if v.Equal(addr) {
			return true
		}
	}
	return false
}

func (s *AddressSet) Merge(b *AddressSet) {
	for bHash, bAddr := range b.set {
		if aAddr, ok := s.set[bHash]; ok {
			if !aAddr.Equal(bAddr) {
				var found bool
				for _, v := range s.coll {
					if v.Equal(bAddr) {
						found = true
						break
					}
				}
				if !found {
					s.coll = append(s.coll, bAddr)
				}
			}
		} else {
			s.set[bHash] = bAddr.Clone()
		}
	}
}

func (s *AddressSet) Len() int {
	return len(s.set) + len(s.coll)
}

func (s AddressSet) Map() map[uint64]Address {
	return s.set
}

func (s AddressSet) HasCollisions() bool {
	return len(s.coll) > 0
}

func (s AddressSet) Slice() []Address {
	if len(s.set) == 0 {
		return nil
	}
	a := make([]Address, 0, len(s.set))
	for _, v := range s.Map() {
		a = append(a, v)
	}
	a = append(a, s.coll...)
	return a
}
