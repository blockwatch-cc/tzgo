// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

// Wraps ticket value type into type structure that is compatible
// with ticket values. This is necessary because T_TICKET uses an
// implicit structure (extra fields amount, ticketer) in addition
// to the defined value.
func TicketType(t Prim) Type {
	tt := t.Clone()
	if len(tt.Anno) == 0 || tt.Anno[0] == "" {
		tt.Anno = []string{":value"}
		switch tt.Type {
		case PrimNullary, PrimUnary, PrimBinary:
			tt.Type++
		}
	}
	return Type{NewPairType(
		NewPrim(T_ADDRESS, ":ticketer"),
		NewPairType(
			tt,
			NewPrim(T_INT, ":amount"),
		),
	)}
}
