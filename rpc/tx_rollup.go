// Copyright (c) 2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"encoding/json"

	"blockwatch.cc/tzgo/tezos"
)

// Ensure TxRollup implements the TypedOperation interface.
var _ TypedOperation = (*TxRollup)(nil)

// TxRollup represents any kind of rollup operation
type TxRollup struct {
	// common
	Manager

	// rollup address (used by most ops)
	Rollup tezos.Address `json:"rollup"`

	// tx_rollup_origination has no data

	// tx_rollup_submit_batch
	Batch TxRollupBatch `json:"-"`

	// tx_rollup_rejection
	Reject TxRollupRejection `json:"-"`

	// tx_rollup_dispatch_tickets
	Dispatch TxRollupDispatch `json:"-"`

	// tx_rollup_commit
	Commit TxRollupCommit `json:"commitment"`
}

type TxRollupResult struct {
	OriginatedRollup tezos.Address `json:"originated_rollup"` // v013 tx_rollup_originate
	Level            int64         `json:"level"`             // v013 ?? here or in metadata??
}

func (r *TxRollup) UnmarshalJSON(data []byte) error {
	type alias *TxRollup
	if err := json.Unmarshal(data, alias(r)); err != nil {
		return err
	}
	switch r.Kind() {
	case tezos.OpTypeTxRollupSubmitBatch:
		return json.Unmarshal(data, &r.Batch)
	case tezos.OpTypeTxRollupRejection:
		return json.Unmarshal(data, &r.Reject)
	case tezos.OpTypeTxRollupDispatchTickets:
		return json.Unmarshal(data, &r.Dispatch)
	}
	return nil
}

func (r *TxRollup) Target() tezos.Address {
	if r.Dispatch.TxRollup.IsValid() {
		return r.Dispatch.TxRollup
	}
	return r.Rollup
}

type TxRollupBatch struct {
	Content tezos.HexBytes `json:"content"`
	// BurnLimit int64          `json:"burn_limit,string,omitempty"`
}

type TxRollupCommit struct {
	Level           int64    `json:"level"`
	Messages        []string `json:"messages"`
	Predecessor     string   `json:"predecessor,omitempty"`
	InboxMerkleRoot string   `json:"inbox_merkle_root"`
}

type TxRollupRejection struct {
	Level                     int64           `json:"level"`
	Message                   json.RawMessage `json:"commitment,omitempty"`
	MessagePosition           tezos.Z         `json:"message_position"`
	MessagePath               []string        `json:"message_path,omitempty"`
	MessageResultHash         string          `json:"message_result_hash"`
	MessageResultPath         []string        `json:"message_result_path,omitempty"`
	PreviousMessageResult     json.RawMessage `json:"previous_message_result,omitempty"`
	PreviousMessageResultPath []string        `json:"previous_message_result_path,omitempty"`
	Proof                     json.RawMessage `json:"proof,omitempty"`
}

type TxRollupDispatch struct {
	Level        int64           `json:"level"`
	TxRollup     tezos.Address   `json:"tx_rollup"`
	ContextHash  string          `json:"context_hash"`
	MessageIndex int64           `json:"message_index"`
	TicketsInfo  json.RawMessage `json:"tickets_info"`
}
