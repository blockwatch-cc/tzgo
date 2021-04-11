// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

type Stack []Prim

func NewStack(args ...Prim) *Stack {
	s := Stack(args)
	return &s
}

func (s *Stack) Pop() Prim {
	if l := s.Len(); l > 0 {
		p := (*s)[l-1]
		*s = (*s)[:l-1]
		return p
	}
	return InvalidPrim
}

func (s *Stack) Push(args ...Prim) {
	for i := len(args) - 1; i >= 0; i-- {
		*s = append(*s, args[i])
	}
}

func (s *Stack) Len() int {
	return len(*s)
}

func (s *Stack) Empty() bool {
	return len(*s) == 0
}

func (s *Stack) Peek() Prim {
	if l := s.Len(); l > 0 {
		return (*s)[l-1]
	}
	return InvalidPrim
}
