// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

type Ticket struct {
	Ticketer tezos.Address  `json:"ticketer"`
	Content  micheline.Prim `json:"content"`
	Type     micheline.Prim `json:"content_type"`
}

type TicketBalanceUpdate struct {
	Account tezos.Address `json:"account"`
	Amount  tezos.Z       `json:"amount"`
}

type TicketUpdate struct {
	Ticket  Ticket                `json:"ticket_token"`
	Updates []TicketBalanceUpdate `json:"updates"`
}
