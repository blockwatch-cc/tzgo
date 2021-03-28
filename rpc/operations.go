// Copyright (c) 2018 ECAD Labs Inc. MIT License
// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/tezos"
)

// OperationHeader represents a single operation included into a block
type OperationHeader struct {
	Protocol  tezos.ProtocolHash `json:"protocol"`
	ChainID   tezos.ChainIdHash  `json:"chain_id"`
	Hash      tezos.OpHash       `json:"hash"`
	Branch    tezos.BlockHash    `json:"branch"`
	Contents  Operations         `json:"contents"`
	Signature string             `json:"signature"`
}

// Operation must be implemented by all operations
type Operation interface {
	OpKind() tezos.OpType
}

type OperationError struct {
	GenericError
	Contract *tezos.Address `json:"contract,omitempty"`
	Amount   int64          `json:"amount,string,omitempty"`
	Balance  int64          `json:"balance,string,omitempty"`
}

// GenericOp is a most generic type
type GenericOp struct {
	Kind tezos.OpType `json:"kind"`
}

// OpKind implements Operation
func (e *GenericOp) OpKind() tezos.OpType {
	return e.Kind
}

// Operations is a slice of Operation (interface type) with custom JSON unmarshaller
type Operations []Operation

// UnmarshalJSON implements json.Unmarshaler
func (e *Operations) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*e = make(Operations, len(raw))

opLoop:
	for i, r := range raw {
		if r == nil {
			continue
		}
		var tmp GenericOp
		if err := json.Unmarshal(r, &tmp); err != nil {
			return fmt.Errorf("rpc: generic operation: %v", err)
		}

		switch tmp.Kind {
		// anonymous operations
		case tezos.OpTypeActivateAccount:
			(*e)[i] = &AccountActivationOp{}
		case tezos.OpTypeDoubleBakingEvidence:
			(*e)[i] = &DoubleBakingOp{}
		case tezos.OpTypeDoubleEndorsementEvidence:
			(*e)[i] = &DoubleEndorsementOp{}
		case tezos.OpTypeSeedNonceRevelation:
			(*e)[i] = &SeedNonceOp{}
		// manager operations
		case tezos.OpTypeTransaction:
			(*e)[i] = &TransactionOp{}
		case tezos.OpTypeOrigination:
			(*e)[i] = &OriginationOp{}
		case tezos.OpTypeDelegation:
			(*e)[i] = &DelegationOp{}
		case tezos.OpTypeReveal:
			(*e)[i] = &RevelationOp{}
		// consensus operations
		case tezos.OpTypeEndorsement:
			(*e)[i] = &EndorsementOp{}
		// amendment operations
		case tezos.OpTypeProposals:
			(*e)[i] = &ProposalsOp{}
		case tezos.OpTypeBallot:
			(*e)[i] = &BallotOp{}

		default:
			log.Warnf("unsupported op '%s'", tmp.Kind)
			(*e)[i] = &tmp
			continue opLoop
		}

		if err := json.Unmarshal(r, (*e)[i]); err != nil {
			return fmt.Errorf("rpc: operation kind %s: %v", tmp.Kind, err)
		}
	}

	return nil
}
