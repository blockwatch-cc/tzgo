// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"blockwatch.cc/tzgo/tezos"
)

// ProposalsOp represents a proposal operation
type ProposalsOp struct {
	GenericOp
	Source    tezos.Address        `json:"source"`
	Period    int                  `json:"period"`
	Proposals []tezos.ProtocolHash `json:"proposals"`
}
