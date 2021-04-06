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
