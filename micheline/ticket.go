// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"blockwatch.cc/tzgo/tezos"
)

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

// Wraps ticket content into structure that is compatible
// with ticket type. This is necessary for transfer_ticket calls which
// use explicit fields for value, amount and ticketer.
func TicketValue(v Prim, ticketer tezos.Address, amount tezos.Z) Prim {
	return NewPair(
		NewBytes(ticketer.EncodePadded()),
		NewPair(v, NewNat(amount.Big())),
	)
}
