// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

type MetadataMode string

const (
	MetadataModeUnset  MetadataMode = ""
	MetadataModeNever  MetadataMode = "never"
	MetadataModeAlways MetadataMode = "always"
)

// Operation represents a single operation or batch of operations included in a block
type Operation struct {
	Protocol  tezos.ProtocolHash `json:"protocol"`
	ChainID   tezos.ChainIdHash  `json:"chain_id"`
	Hash      tezos.OpHash       `json:"hash"`
	Branch    tezos.BlockHash    `json:"branch"`
	Contents  OperationList      `json:"contents"`
	Signature tezos.Signature    `json:"signature"`
	Errors    []OperationError   `json:"error,omitempty"`    // mempool only
	Metadata  string             `json:"metadata,omitempty"` // contains `too large` when stripped, this is BAD!!
}

// TotalCosts returns the sum of costs across all batched and internal operations.
func (o Operation) TotalCosts() tezos.Costs {
	var c tezos.Costs
	for _, op := range o.Contents {
		c = c.Add(op.Costs())
	}
	return c
}

// Costs returns ta list of individual costs for all batched operations.
func (o Operation) Costs() []tezos.Costs {
	list := make([]tezos.Costs, len(o.Contents))
	for i, op := range o.Contents {
		list[i] = op.Costs()
	}
	return list
}

// TypedOperation must be implemented by all operations
type TypedOperation interface {
	Kind() tezos.OpType
	Meta() OperationMetadata
	Result() OperationResult
	Costs() tezos.Costs
	Limits() tezos.Limits
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
	Delegate            tezos.Address `json:"delegate"`
	Slots               []int         `json:"slots,omitempty"`
	EndorsementPower    int           `json:"endorsement_power,omitempty"`    // v12+
	PreendorsementPower int           `json:"preendorsement_power,omitempty"` // v12+

	// some rollup ops only, FIXME: is this correct here or is this field in result?
	Level int64 `json:"level"`

	// v18 slashing ops may block a baker
	ForbiddenDelegate tezos.Address `json:"forbidden_delegate"` // v18+
}

// Address returns the delegate address for endorsements.
func (m OperationMetadata) Address() tezos.Address {
	return m.Delegate
}

// OperationResult contains receipts for executed operations, both success and failed.
// This type is a generic container for all possible results. Which fields are actually
// used depends on operation type and performed actions.
type OperationResult struct {
	Status               tezos.OpStatus   `json:"status"`
	BalanceUpdates       BalanceUpdates   `json:"balance_updates"`
	ConsumedGas          int64            `json:"consumed_gas,string"`      // deprecated in v015
	ConsumedMilliGas     int64            `json:"consumed_milligas,string"` // v007+
	Errors               []OperationError `json:"errors,omitempty"`
	Allocated            bool             `json:"allocated_destination_contract"` // tx only
	Storage              *micheline.Prim  `json:"storage,omitempty"`              // tx, orig
	OriginatedContracts  []tezos.Address  `json:"originated_contracts"`           // orig only
	StorageSize          int64            `json:"storage_size,string"`            // tx, orig, const
	PaidStorageSizeDiff  int64            `json:"paid_storage_size_diff,string"`  // tx, orig
	BigmapDiff           json.RawMessage  `json:"big_map_diff,omitempty"`         // tx, orig, <v013
	LazyStorageDiff      json.RawMessage  `json:"lazy_storage_diff,omitempty"`    // v008+ tx, orig
	GlobalAddress        tezos.ExprHash   `json:"global_address"`                 // const
	TicketUpdatesCorrect []TicketUpdate   `json:"ticket_updates"`                 // v015
	TicketReceipts       []TicketUpdate   `json:"ticket_receipt"`                 // v015, name on internal

	// v013 tx rollup
	TxRollupResult

	// v016 smart rollup
	SmartRollupResult
}

// Always use this helper to retrieve Ticket updates. This is because due to
// lack of quality control Tezos Lima protocol ended up with 2 distinct names
// for ticket updates in external call receipts versus internal call receipts.
func (r OperationResult) TicketUpdates() []TicketUpdate {
	if len(r.TicketUpdatesCorrect) > 0 {
		return r.TicketUpdatesCorrect
	}
	return r.TicketReceipts
}

func (r OperationResult) BigmapEvents() micheline.BigmapEvents {
	if r.LazyStorageDiff != nil {
		res := make(micheline.LazyEvents, 0)
		_ = json.Unmarshal(r.LazyStorageDiff, &res)
		return res.BigmapEvents()
	}
	if r.BigmapDiff != nil {
		res := make(micheline.BigmapEvents, 0)
		_ = json.Unmarshal(r.BigmapDiff, &res)
		return res
	}
	return nil
}

func (r OperationResult) IsSuccess() bool {
	return r.Status == tezos.OpStatusApplied
}

func (r OperationResult) Gas() int64 {
	if r.ConsumedMilliGas > 0 {
		var corr int64
		if r.ConsumedMilliGas%1000 > 0 {
			corr++
		}
		return r.ConsumedMilliGas/1000 + corr
	}
	return r.ConsumedGas
}

func (r OperationResult) MilliGas() int64 {
	if r.ConsumedMilliGas > 0 {
		return r.ConsumedMilliGas
	}
	return r.ConsumedGas * 1000
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
	OpKind   tezos.OpType      `json:"kind"`
	Metadata OperationMetadata `json:"metadata"`
}

// Kind returns the operation's type. Implements TypedOperation interface.
func (e Generic) Kind() tezos.OpType {
	return e.OpKind
}

// Meta returns an empty operation metadata to implement TypedOperation interface.
func (e Generic) Meta() OperationMetadata {
	return e.Metadata
}

// Result returns an empty operation result to implement TypedOperation interface.
func (e Generic) Result() OperationResult {
	return e.Metadata.Result
}

// Costs returns empty operation costs to implement TypedOperation interface.
func (e Generic) Costs() tezos.Costs {
	return tezos.Costs{}
}

// Limits returns empty operation limits to implement TypedOperation interface.
func (e Generic) Limits() tezos.Limits {
	return tezos.Limits{}
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

// Limits returns manager operation limits to implement TypedOperation interface.
func (e Manager) Limits() tezos.Limits {
	return tezos.Limits{
		Fee:          e.Fee,
		GasLimit:     e.GasLimit,
		StorageLimit: e.StorageLimit,
	}
}

// OperationList is a slice of TypedOperation (interface type) with custom JSON unmarshaller
type OperationList []TypedOperation

// Contains returns true when the list contains an operation of kind typ.
func (o OperationList) Contains(typ tezos.OpType) bool {
	for _, v := range o {
		if v.Kind() == typ {
			return true
		}
	}
	return false
}

func (o OperationList) Select(typ tezos.OpType, n int) TypedOperation {
	var cnt int
	for _, v := range o {
		if v.Kind() != typ {
			continue
		}
		if cnt == n {
			return v
		}
		cnt++
	}
	return nil
}

func (o OperationList) Len() int {
	return len(o)
}

func (o OperationList) N(n int) TypedOperation {
	if n < 0 {
		n += len(o)
	}
	return o[n]
}

// UnmarshalJSON implements json.Unmarshaler
func (e *OperationList) UnmarshalJSON(data []byte) error {
	if len(data) <= 2 {
		return nil
	}

	if data[0] != '[' {
		return fmt.Errorf("rpc: expected operation array")
	}

	// fmt.Printf("Decoding ops: %s\n", string(data))
	dec := json.NewDecoder(bytes.NewReader(data))

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return fmt.Errorf("rpc: %v", err)
	}

	for dec.More() {
		// peek into `{"kind":"...",` field
		start := int(dec.InputOffset()) + 9
		// after first JSON object, decoder pos is at `,`
		if data[start] == '"' {
			start += 1
		}
		end := start + bytes.IndexByte(data[start:], '"')
		kind := tezos.ParseOpType(string(data[start:end]))
		var op TypedOperation
		switch kind {
		// anonymous operations
		case tezos.OpTypeActivateAccount:
			op = &Activation{}
		case tezos.OpTypeDoubleBakingEvidence:
			op = &DoubleBaking{}
		case tezos.OpTypeDoubleEndorsementEvidence,
			tezos.OpTypeDoublePreendorsementEvidence:
			op = &DoubleEndorsement{}
		case tezos.OpTypeSeedNonceRevelation:
			op = &SeedNonce{}
		case tezos.OpTypeDrainDelegate:
			op = &DrainDelegate{}

		// consensus operations
		case tezos.OpTypeEndorsement,
			tezos.OpTypeEndorsementWithSlot,
			tezos.OpTypePreendorsement:
			op = &Endorsement{}

		// amendment operations
		case tezos.OpTypeProposals:
			op = &Proposals{}
		case tezos.OpTypeBallot:
			op = &Ballot{}

		// manager operations
		case tezos.OpTypeTransaction:
			op = &Transaction{}
		case tezos.OpTypeOrigination:
			op = &Origination{}
		case tezos.OpTypeDelegation:
			op = &Delegation{}
		case tezos.OpTypeReveal:
			op = &Reveal{}
		case tezos.OpTypeRegisterConstant:
			op = &ConstantRegistration{}
		case tezos.OpTypeSetDepositsLimit:
			op = &SetDepositsLimit{}
		case tezos.OpTypeIncreasePaidStorage:
			op = &IncreasePaidStorage{}
		case tezos.OpTypeVdfRevelation:
			op = &VdfRevelation{}
		case tezos.OpTypeTransferTicket:
			op = &TransferTicket{}
		case tezos.OpTypeUpdateConsensusKey:
			op = &UpdateConsensusKey{}

			// DEPRECATED: tx rollup operations, kept for testnet backward compatibility
		case tezos.OpTypeTxRollupOrigination,
			tezos.OpTypeTxRollupSubmitBatch,
			tezos.OpTypeTxRollupCommit,
			tezos.OpTypeTxRollupReturnBond,
			tezos.OpTypeTxRollupFinalizeCommitment,
			tezos.OpTypeTxRollupRemoveCommitment,
			tezos.OpTypeTxRollupRejection,
			tezos.OpTypeTxRollupDispatchTickets:
			op = &TxRollup{}

		case tezos.OpTypeSmartRollupOriginate:
			op = &SmartRollupOriginate{}
		case tezos.OpTypeSmartRollupAddMessages:
			op = &SmartRollupAddMessages{}
		case tezos.OpTypeSmartRollupCement:
			op = &SmartRollupCement{}
		case tezos.OpTypeSmartRollupPublish:
			op = &SmartRollupPublish{}
		case tezos.OpTypeSmartRollupRefute:
			op = &SmartRollupRefute{}
		case tezos.OpTypeSmartRollupTimeout:
			op = &SmartRollupTimeout{}
		case tezos.OpTypeSmartRollupExecuteOutboxMessage:
			op = &SmartRollupExecuteOutboxMessage{}
		case tezos.OpTypeSmartRollupRecoverBond:
			op = &SmartRollupRecoverBond{}
		case tezos.OpTypeDalAttestation:
			op = &DalAttestation{}
		case tezos.OpTypeDalPublishSlotHeader:
			op = &DalPublishSlotHeader{}

		default:
			return fmt.Errorf("rpc: unsupported op %q", string(data[start:end]))
		}

		if err := dec.Decode(op); err != nil {
			return fmt.Errorf("rpc: operation kind %s: %v", kind, err)
		}
		(*e) = append(*e, op)
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
	if c.MetadataMode != "" {
		u += "?metadata=" + string(c.MetadataMode)
	}
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
	if c.MetadataMode != "" {
		u += "?metadata=" + string(c.MetadataMode)
	}
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
	if c.MetadataMode != "" {
		u += "?metadata=" + string(c.MetadataMode)
	}
	if err := c.Get(ctx, u, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}
