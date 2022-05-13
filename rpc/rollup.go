// Copyright (c) 2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
    "blockwatch.cc/tzgo/micheline"
    "blockwatch.cc/tzgo/tezos"
)

// Ensure Rollup implements the TypedOperation interface.
var _ TypedOperation = (*Rollup)(nil)

// Rollup represents any kind of rollup operation
type Rollup struct {
    Manager
    Rollup tezos.Address `json:"rollup"`

    // // tx_rollup_origination
    // Origination json.RawMessage `json:"tx_rollup_origination"`

    // // tx_rollup_commit
    // Commitment json.RawMessage `json:"commitment"`

    // // tx_rollup_submit_batch
    // Content   json.RawMessage `json:"content"`
    // BurnLimit int64           `json:"burn_limit,string"`

    // // tx_rollup_rejection
    // Level                     int64           `json:"level"`
    // Message                   json.RawMessage `json:"commitment"`
    // MessagePosition           tezos.Z         `json:"message_position"`
    // MessagePath               []tezos.Hash    `json:"message_path"`
    // MessageResultHash         tezos.Hash      `json:"message_result_hash"`
    // MessageResultPath         []tezos.Hash    `json:"message_result_path"`
    // PreviousMessageResult     json.RawMessage `json:"previous_message_result"`
    // PreviousMessageResultPath []tezos.Hash    `json:"previous_message_result_path"`
    // Proof                     json.RawMessage `json:"proof"`

    // // tx_rollup_dispatch_tickets
    // TxRollup     tezos.Address   `json:"tx_rollup"`
    // ContextHash  tezos.Hash      `json:"context_hash"`
    // MessageIndex int64           `json:"message_index"`
    // TicketsInfo  json.RawMessage `json:"tickets_info"`

    // // sc_rollup_originate
    // Kind       json.RawMessage `json:"kind"`
    // BootSector json.RawMessage `json:"boot_sector"`
}

type TransferTicket struct {
    Manager
    Destination tezos.Address `json:"destination"`
    Entrypoint  string        `json:"entrypoint"`

    // transfer_ticket contents
    Type     micheline.Prim `json:"ticket_ty"`
    Contents micheline.Prim `json:"ticket_contents"`
    Ticketer tezos.Address  `json:"ticket_ticketer"`
    Amount   tezos.Z        `json:"ticket_amount"`
}

func (t *TransferTicket) EncodeParameters() micheline.Prim {
    return micheline.NewPair(
        micheline.TicketType(t.Type).Prim,
        micheline.TicketValue(t.Contents, t.Ticketer, t.Amount),
    )
}
