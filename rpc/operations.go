// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/micheline"
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

// Cost returns the sum of costs across all batched and internal operations.
func (o Operation) Cost() OperationCost {
	var c OperationCost
	for _, op := range o.Contents {
		c = c.Add(op.Cost())
	}
	return c
}

// TypedOperation must be implemented by all operations
type TypedOperation interface {
	Kind() tezos.OpType
	Meta() OperationMetadata
	Result() OperationResult
	Cost() OperationCost
}

// OperationCost represents all costs paid by an operation. Its contents depends on
// operation type and activity. On transactions with internal results costs are a
// summary.
type OperationCost struct {
	Fee            int64
	Burn           int64
	Gas            int64
	StorageBytes   int64
	StorageBurn    int64
	AllocationBurn int64
}

// Add adds two costs z = x + y and returns the sum z without changing any of the inputs.
func (x OperationCost) Add(y OperationCost) OperationCost {
	x.Fee += y.Fee
	x.Burn += y.Burn
	x.Gas += y.Gas
	x.StorageBurn += y.StorageBurn
	x.AllocationBurn += y.AllocationBurn
	return x
}

// OperationError represents data describing an error conditon that lead to a
// failed operation execution.
type OperationError struct {
	GenericError
	Contract *tezos.Address  `json:"contract,omitempty"`
	Raw      json.RawMessage `json:"-"`
}

// OperationMetadata contains execution receipts for successful and failed
// operations.
type OperationMetadata struct {
	BalanceUpdates BalanceUpdates  `json:"balance_updates"` // fee-related
	Result         OperationResult `json:"operation_result"`

	// transaction only
	InternalResults []*InternalResult `json:"internal_operation_results,omitempty"`

	// endorsement only
	Delegate tezos.Address `json:"delegate"`
	Slots    []int         `json:"slots,omitempty"`
}

// Address returns the delegate address for endorsements.
func (m OperationMetadata) Address() tezos.Address {
	return m.Delegate
}

// OperationResult contains receipts for executed operations, both success and failed.
// This type is a generic container for all possible results. Which fields are actually
// used depends on operation type and performed actions.
type OperationResult struct {
	Status              tezos.OpStatus       `json:"status"`
	BalanceUpdates      BalanceUpdates       `json:"balance_updates"` // burn, etc
	ConsumedGas         int64                `json:"consumed_gas,string"`
	ConsumedMilliGas    int64                `json:"consumed_milligas,string"` // v007+
	Errors              []OperationError     `json:"errors,omitempty"`
	Allocated           bool                 `json:"allocated_destination_contract"` // tx only
	Storage             *micheline.Prim      `json:"storage,omitempty"`              // tx, orig
	OriginatedContracts []tezos.Address      `json:"originated_contracts"`           // orig only
	StorageSize         int64                `json:"storage_size,string"`            // tx, orig, const
	PaidStorageSizeDiff int64                `json:"paid_storage_size_diff,string"`  // tx, orig
	BigmapDiff          micheline.BigmapDiff `json:"big_map_diff,omitempty"`         // tx, orig
	LazyStorageDiff     LazyStorageDiff      `json:"lazy_storage_diff,omitempty"`    // v008+ tx, orig
	GlobalAddress       tezos.ExprHash       `json:"global_address"`                 // const
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

// Generic is the most generic operation type.
type Generic struct {
	OpKind tezos.OpType `json:"kind"`
}

// Manager represents data common for all manager operations.
type Manager struct {
	Generic
	Source       tezos.Address `json:"source"`
	Fee          int64         `json:"fee,string"`
	Counter      int64         `json:"counter,string"`
	GasLimit     int64         `json:"gas_limit,string"`
	StorageLimit int64         `json:"storage_limit,string"`
}

// Kind returns the operation's type. Implements TypedOperation interface.
func (e Generic) Kind() tezos.OpType {
	return e.OpKind
}

// Meta returns an empty operation metadata to implement TypedOperation interface.
func (e Generic) Meta() OperationMetadata {
	return OperationMetadata{}
}

// Result returns an empty operation result to implement TypedOperation interface.
func (e Generic) Result() OperationResult {
	return OperationResult{}
}

// Cost returns an empty operation cost to implement TypedOperation interface.
func (e Generic) Cost() OperationCost {
	return OperationCost{}
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
		var tmp Generic
		if err := json.Unmarshal(r, &tmp); err != nil {
			return fmt.Errorf("rpc: generic operation: %w", err)
		}

		switch tmp.Kind() {
		// anonymous operations
		case tezos.OpTypeActivateAccount:
			(*e)[i] = &Activation{}
		case tezos.OpTypeDoubleBakingEvidence:
			(*e)[i] = &DoubleBaking{}
		case tezos.OpTypeDoubleEndorsementEvidence:
			(*e)[i] = &DoubleEndorsement{}
		case tezos.OpTypeSeedNonceRevelation:
			(*e)[i] = &SeedNonce{}
		// manager operations
		case tezos.OpTypeTransaction:
			(*e)[i] = &Transaction{}
		case tezos.OpTypeOrigination:
			(*e)[i] = &Origination{}
		case tezos.OpTypeDelegation:
			(*e)[i] = &Delegation{}
		case tezos.OpTypeReveal:
			(*e)[i] = &Reveal{}
		case tezos.OpTypeRegisterConstant:
			(*e)[i] = &ConstantRegistration{}
		// consensus operations
		case tezos.OpTypeEndorsement, tezos.OpTypeEndorsementWithSlot:
			(*e)[i] = &Endorsement{}
		// amendment operations
		case tezos.OpTypeProposals:
			(*e)[i] = &Proposals{}
		case tezos.OpTypeBallot:
			(*e)[i] = &Ballot{}

		default:
			log.Warnf("unsupported op '%s'", tmp.Kind)
			(*e)[i] = &tmp
			continue opLoop
		}

		if err := json.Unmarshal(r, (*e)[i]); err != nil {
			return fmt.Errorf("rpc: operation kind %s: %w", tmp.Kind(), err)
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
