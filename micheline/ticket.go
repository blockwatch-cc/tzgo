// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

// call with inner value, not outer ticket type
func TicketType(t Prim) Type {
	tt := t.Clone()
	if len(tt.Anno) == 0 || tt.Anno[0] == "" {
		tt.Anno = []string{":value"}
		switch tt.Type {
		case PrimNullary, PrimUnary, PrimBinary:
			tt.Type++
		}
	}
	return Type{tpair(
		prim(T_ADDRESS, ":ticketer"),
		tpair(
			tt,
			prim(T_INT, ":amount"),
		),
	)}
}

// FIXME: ideally this is not necessary and handled while walking trees
func (t *Type) ExpandTickets() Type {
	if t.OpCode == T_TICKET {
		return TicketType(t.Args[0])
	}
	for i, v := range t.Args {
		t.Args[i] = v.ExpandTickets()
	}
	return *t
}

func (p *Prim) ExpandTickets() Prim {
	if p.OpCode == T_TICKET {
		return TicketType(p.Args[0]).Prim
	}
	for i, v := range p.Args {
		p.Args[i] = v.ExpandTickets()
	}
	return *p
}
