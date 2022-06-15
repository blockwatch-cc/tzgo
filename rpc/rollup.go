// Copyright (c) 2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
    "encoding/json"

    "blockwatch.cc/tzgo/micheline"
    "blockwatch.cc/tzgo/tezos"
)

// Ensure Rollup implements the TypedOperation interface.
var _ TypedOperation = (*Rollup)(nil)

// Rollup represents any kind of rollup operation
type Rollup struct {
    // common
    Manager

    // rollup address (used by most ops)
    Rollup tezos.Address `json:"rollup"`

    // tx_rollup_origination has no data

    // transfer_ticket contents
    Transfer TransferTicket `json:"-"`

    // tx_rollup_submit_batch
    Batch RollupBatch `json:"-"`

    // tx_rollup_rejection
    Reject RollupRejection `json:"-"`

    // tx_rollup_dispatch_tickets
    Dispatch RollupDispatch `json:"-"`

    // tx_rollup_commit
    Commit RollupCommit `json:"commitment"`

    // // sc_rollup_originate
    // Kind       json.RawMessage `json:"kind"`
    // BootSector json.RawMessage `json:"boot_sector"`
}

func (r *Rollup) UnmarshalJSON(data []byte) error {
    type alias *Rollup
    if err := json.Unmarshal(data, alias(r)); err != nil {
        return err
    }
    switch r.Kind() {
    case tezos.OpTypeTransferTicket:
        return json.Unmarshal(data, &r.Transfer)
    case tezos.OpTypeToruSubmitBatch:
        return json.Unmarshal(data, &r.Batch)
    case tezos.OpTypeToruRejection:
        return json.Unmarshal(data, &r.Reject)
    case tezos.OpTypeToruDispatchTickets:
        return json.Unmarshal(data, &r.Dispatch)
    }
    return nil
}

func (r *Rollup) Target() tezos.Address {
    if r.Transfer.Destination.IsValid() {
        return r.Transfer.Destination
    }
    if r.Dispatch.TxRollup.IsValid() {
        return r.Dispatch.TxRollup
    }
    return r.Rollup
}

type RollupBatch struct {
    Content tezos.HexBytes `json:"content"`
    // BurnLimit int64          `json:"burn_limit,string,omitempty"`
}

type TransferTicket struct {
    Destination tezos.Address  `json:"destination"`
    Entrypoint  string         `json:"entrypoint"`
    Type        micheline.Prim `json:"ticket_ty"`
    Contents    micheline.Prim `json:"ticket_contents"`
    Ticketer    tezos.Address  `json:"ticket_ticketer"`
    Amount      tezos.Z        `json:"ticket_amount"`
}

type RollupCommit struct {
    Level           int64        `json:"level"`
    Messages        []tezos.Hash `json:"messages"`
    Predecessor     *tezos.Hash  `json:"predecessor,omitempty"`
    InboxMerkleRoot tezos.Hash   `json:"inbox_merkle_root"`
}

type RollupRejection struct {
    Level                     int64           `json:"level"`
    Message                   json.RawMessage `json:"commitment,omitempty"`
    MessagePosition           tezos.Z         `json:"message_position"`
    MessagePath               []tezos.Hash    `json:"message_path,omitempty"`
    MessageResultHash         tezos.Hash      `json:"message_result_hash"`
    MessageResultPath         []tezos.Hash    `json:"message_result_path,omitempty"`
    PreviousMessageResult     json.RawMessage `json:"previous_message_result,omitempty"`
    PreviousMessageResultPath []tezos.Hash    `json:"previous_message_result_path,omitempty"`
    Proof                     json.RawMessage `json:"proof,omitempty"`
}

type RollupDispatch struct {
    Level        int64           `json:"level"`
    TxRollup     tezos.Address   `json:"tx_rollup"`
    ContextHash  tezos.Hash      `json:"context_hash"`
    MessageIndex int64           `json:"message_index"`
    TicketsInfo  json.RawMessage `json:"tickets_info"`
}
