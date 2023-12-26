// Copyright (c) 2020-2023 Blockwatch Data Inc.
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

func BuildAddressSet(s ...string) (*AddressSet, error) {
	set := NewAddressSet()
	for _, v := range s {
		addr, err := ParseAddress(v)
		if err != nil {
			return nil, err
		}
		set.AddUnique(addr)
	}
	return set, nil
}

func MustBuildAddressSet(s ...string) *AddressSet {
	set, err := BuildAddressSet(s...)
	if err != nil {
		panic(err)
	}
	return set
}

func (s *AddressSet) AddUnique(addr Address) bool {
	h := hash.Hash64(addr[:])
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
	delete(s.set, hash.Hash64(addr[:]))
	for i := range s.coll {
		if !s.coll[i].Equal(addr) {
			continue
		}
		s.coll = append(s.coll[:i], s.coll[i+1:]...)
		return
	}
}

func (s *AddressSet) Clear() {
	for n := range s.set {
		delete(s.set, n)
	}
	s.coll = s.coll[:0]
}

func (s *AddressSet) Contains(addr Address) bool {
	if s == nil || len(s.set) == 0 {
		return false
	}
	a, ok := s.set[hash.Hash64(addr[:])]
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

func (s AddressSet) HasIntersect(t *AddressSet) bool {
	for k, v := range s.set {
		a, ok := t.set[k]
		if !ok {
			continue
		}
		if v.Equal(a) {
			return true
		}
		for _, v := range t.coll {
			if v.Equal(a) {
				return true
			}
		}
	}
	for _, v := range s.coll {
		if t.Contains(v) {
			return true
		}
	}
	return false
}

func (s AddressSet) Intersect(t *AddressSet) *AddressSet {
	i := NewAddressSet()
	for k, v := range s.set {
		a, ok := t.set[k]
		if !ok {
			continue
		}
		if v.Equal(a) {
			i.AddUnique(a)
		} else {
			for _, a := range t.coll {
				if v.Equal(a) {
					i.AddUnique(a)
				}
			}
		}
	}
	for _, v := range s.coll {
		if t.Contains(v) {
			i.AddUnique(v)
		}
	}
	return i
}
