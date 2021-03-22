// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
	"encoding/json"
)

// BallotOp represents a ballot operation
type BallotOp struct {
	GenericOp
	Source   tezos.Address      `json:"source"`
	Period   int                `json:"period"`
	Ballot   tezos.BallotVote   `json:"ballot"` // yay, nay, pass
	Proposal tezos.ProtocolHash `json:"proposal"`
	Metadata json.RawMessage    `json:"metadata"` // missing example
}
