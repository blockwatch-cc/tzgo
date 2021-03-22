// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

// call with inner value, not outer ticket type
func TicketType(t *Prim) *Prim {
	tt := t.Clone()
	if len(tt.Anno) == 0 || tt.Anno[0] == "" {
		tt.Anno = []string{":value"}
		switch tt.Type {
		case PrimNullary, PrimUnary, PrimBinary:
			tt.Type++
		}
	}
	return tpair(
		prim(T_ADDRESS, ":ticketer"),
		tpair(
			tt,
			prim(T_INT, ":amount"),
		),
	)
}

func (p *Prim) ExpandTickets() *Prim {
	if p.OpCode == T_TICKET {
		return TicketType(p.Args[0])
	}
	for i, v := range p.Args {
		p.Args[i] = v.ExpandTickets()
	}
	return p
}
