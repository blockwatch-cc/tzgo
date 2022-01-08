// Copyright (c) 2018 ECAD Labs Inc. MIT License
// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/tezos"
)

// Operation represents a single operation or batch of operations included in a block
type Operation struct {
	Protocol  tezos.ProtocolHash `json:"protocol"`
	ChainID   tezos.ChainIdHash  `json:"chain_id"`
	Hash      tezos.OpHash       `json:"hash"`
	Branch    tezos.BlockHash    `json:"branch"`
	Contents  OperationList      `json:"contents"`
	Signature tezos.Signature    `json:"signature"`
	Errors    []OperationError   `json:"error,omitempty"` // mempool only
}

// TypedOperation must be implemented by all operations
type TypedOperation interface {
	OpKind() tezos.OpType
}

type OperationError struct {
	GenericError
	Contract *tezos.Address  `json:"contract,omitempty"`
	Raw      json.RawMessage `json:"-"`
}

func (o OperationError) MarshalJSON() ([]byte, error) {
	return o.Raw, nil
}

func (o *OperationError) UnmarshalJSON(data []byte) error {
	type alias OperationError
	if err := json.Unmarshal(data, (*alias)(o)); err != nil {
		return err
	}
	o.Raw = make([]byte, len(data))
	copy(o.Raw, data)
	return nil
}

// GenericOp is a most generic type
type GenericOp struct {
	Kind tezos.OpType `json:"kind"`
}

// OpKind returns the operation's type. Implements TypedOperation interface.
func (e *GenericOp) OpKind() tezos.OpType {
	return e.Kind
}

// OperationList is a slice of TypedOperation (interface type) with custom JSON unmarshaller
type OperationList []TypedOperation

// UnmarshalJSON implements json.Unmarshaler
func (e *OperationList) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*e = make(OperationList, len(raw))

opLoop:
	for i, r := range raw {
		if r == nil {
			continue
		}
		var tmp GenericOp
		if err := json.Unmarshal(r, &tmp); err != nil {
			return fmt.Errorf("rpc: generic operation: %w", err)
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
		case tezos.OpTypeRegisterConstant:
			(*e)[i] = &ConstantRegistrationOp{}
		// consensus operations
		case tezos.OpTypeEndorsement, tezos.OpTypeEndorsementWithSlot:
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
			return fmt.Errorf("rpc: operation kind %s: %w", tmp.Kind, err)
		}
	}

	return nil
}

// GetBlockOperationHash returns a single operation hashes included in block
// https://tezos.gitlab.io/active/rpc.html#get-block-id-operation-hashes-list-offset-operation-offset
func (c *Client) GetBlockOperationHash(ctx context.Context, id BlockID, l, n int) (tezos.OpHash, error) {
	var hash tezos.OpHash
	u := fmt.Sprintf("chains/main/blocks/%s/operation_hashes/%d/%d", id, l, n)
	err := c.Get(ctx, u, &hash)
	return hash, err
}

// GetBlockOperationHashes returns a list of list of operation hashes included in block
// https://tezos.gitlab.io/active/rpc.html#get-block-id-operation-hashes
func (c *Client) GetBlockOperationHashes(ctx context.Context, id BlockID) ([][]tezos.OpHash, error) {
	hashes := make([][]tezos.OpHash, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/operation_hashes", id)
	if err := c.Get(ctx, u, &hashes); err != nil {
		return nil, err
	}
	return hashes, nil
}

// GetBlockOperationListHashes returns a list of operation hashes included in block
// at a specified list position (i.e. validation pass) [0..3]
// https://tezos.gitlab.io/active/rpc.html#get-block-id-operation-hashes-list-offset
func (c *Client) GetBlockOperationListHashes(ctx context.Context, id BlockID, l int) ([]tezos.OpHash, error) {
	hashes := make([]tezos.OpHash, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/operation_hashes/%d", id, l)
	if err := c.Get(ctx, u, &hashes); err != nil {
		return nil, err
	}
	return hashes, nil
}

// GetBlockOperation returns information about a single validated Tezos operation group
// (i.e. a single operation or a batch of operations) at list l and position n
// https://tezos.gitlab.io/active/rpc.html#get-block-id-operations-list-offset-operation-offset
func (c *Client) GetBlockOperation(ctx context.Context, id BlockID, l, n int) (*Operation, error) {
	var op Operation
	u := fmt.Sprintf("chains/main/blocks/%s/operations/%d/%d", id, l, n)
	if err := c.Get(ctx, u, &op); err != nil {
		return nil, err
	}
	return &op, nil
}

// GetBlockOperationList returns information about all validated Tezos operation group
// inside operation list l (i.e. validation pass) [0..3].
// https://tezos.gitlab.io/active/rpc.html#get-block-id-operations-list-offset
func (c *Client) GetBlockOperationList(ctx context.Context, id BlockID, l int) ([]Operation, error) {
	ops := make([]Operation, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/operations/%d", id, l)
	if err := c.Get(ctx, u, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}

// GetBlockOperations returns information about all validated Tezos operation groups
// from all operation lists in block.
// https://tezos.gitlab.io/active/rpc.html#get-block-id-operations
func (c *Client) GetBlockOperations(ctx context.Context, id BlockID) ([][]Operation, error) {
	ops := make([][]Operation, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/operations", id)
	if err := c.Get(ctx, u, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}

// BroadcastOperation sends a signed operation to the network (injection).
// The call returns the operation hash on success. If theoperation was rejected
// by the node error is of type RPCError.
func (c *Client) BroadcastOperation(ctx context.Context, body []byte) (hash tezos.OpHash, err error) {
	err = c.Post(ctx, "injection/operation", hex.EncodeToString(body), &hash)
	return
}

// RunOperation simulates executing an operation without requiring a valid signature.
// The call returns the execution result as regular operation receipt.
func (c *Client) RunOperation(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/scripts/run_operation", id)
	return c.Post(ctx, u, body, resp)
}

// ForgeOperation uses a remote node to serialize an operation to its binary format.
// The result of this call SHOULD NEVER be used for signing the operation, it is only
// meant for validating the locally generated serialized output.
func (c *Client) ForgeOperation(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/forge/operations", id)
	return c.Post(ctx, u, body, resp)
}

// RunCode simulates executing of provided code on the context of a contract at selected block.
func (c *Client) RunCode(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/scripts/run_code", id)
	return c.Post(ctx, u, body, resp)
}

// RunView simulates executing of on on-chain view on the context of a contract at selected block.
func (c *Client) RunView(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/scripts/run_view", id)
	return c.Post(ctx, u, body, resp)
}

// TraceCode simulates executing of code on the context of a contract at selected block and
// returns a full execution trace.
func (c *Client) TraceCode(ctx context.Context, id BlockID, body, resp interface{}) error {
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/scripts/trace_code", id)
	return c.Post(ctx, u, body, resp)
}
