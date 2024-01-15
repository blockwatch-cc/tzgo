// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

const (
	EmmyBlockWatermark                byte = 0x01 // deprecated
	EmmyEndorsementWatermark          byte = 0x02 // deprecated
	OperationWatermark                byte = 0x03
	TenderbakeBlockWatermark          byte = 0x11
	TenderbakePreendorsementWatermark byte = 0x12
	TenderbakeEndorsementWatermark    byte = 0x13
)

var (
	// enc defines the default wire encoding used for Tezos messages
	enc = binary.BigEndian
)

// Operation is a generic type used to handle different Tezos operation
// types inside an operation's contents list.
type Operation interface {
	Kind() tezos.OpType
	Limits() tezos.Limits
	GetCounter() int64
	WithSource(tezos.Address)
	WithCounter(int64)
	WithLimits(tezos.Limits)
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	MarshalJSON() ([]byte, error)
	EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error
	DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) error
}

// Op is a container used to collect, serialize and sign Tezos operations.
// It serves as a low level building block for constructing and serializing
// operations, but is agnostic to the order/lifecycle in which data is added
// or updated.
type Op struct {
	Branch    tezos.BlockHash    `json:"branch"`    // used for TTL handling
	Contents  []Operation        `json:"contents"`  // non-zero list of transactions
	Signature tezos.Signature    `json:"signature"` // added during the lifecycle
	ChainId   *tezos.ChainIdHash `json:"-"`         // optional, used for remote signing only
	TTL       int64              `json:"-"`         // optional, specify TTL in blocks
	Params    *tezos.Params      `json:"-"`         // optional, define protocol to encode for
	Source    tezos.Address      `json:"-"`         // optional, used as manager/sender
}

// NewOp creates a new empty operation that uses default params and a
// default operation TTL.
func NewOp() *Op {
	return &Op{
		Params: tezos.DefaultParams,
		TTL:    tezos.DefaultParams.MaxOperationsTTL - 2, // Ithaca recommendation
	}
}

// NeedCounter returns true if any of the contained operations has not assigned
// a valid counter value.
func (o Op) NeedCounter() bool {
	for _, v := range o.Contents {
		if v.GetCounter() == 0 {
			return true
		}
	}
	return false
}

// WithParams defines the protocol and other chain configuration params for which
// the operation will be encoded. If unset, defaults to tezos.DefaultParams.
func (o *Op) WithParams(p *tezos.Params) *Op {
	o.Params = p
	return o
}

// WithContents adds a Tezos operation to the end of the contents list.
func (o *Op) WithContents(op Operation) *Op {
	o.Contents = append(o.Contents, op)
	return o
}

// WithContentsFront adds a Tezos operation to the front of the contents list.
func (o *Op) WithContentsFront(op Operation) *Op {
	o.Contents = append([]Operation{op}, o.Contents...)
	return o
}

// WithSource sets the source for all manager operations to addr. It is required
// before calling other WithXXX functions.
func (o *Op) WithSource(addr tezos.Address) *Op {
	for _, v := range o.Contents {
		v.WithSource(addr)
	}
	o.Source = addr
	return o
}

// WithTransfer adds a simple value transfer transaction to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithTransfer(to tezos.Address, amount int64) *Op {
	o.Contents = append(o.Contents, &Transaction{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Amount:      tezos.N(amount),
		Destination: to,
	})
	return o
}

// WithCall adds a contract call transaction to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithCall(to tezos.Address, params micheline.Parameters) *Op {
	o.Contents = append(o.Contents, &Transaction{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Destination: to,
		Parameters:  &params,
	})
	return o
}

// WithCallExt adds a contract call with value transfer transaction to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithCallExt(to tezos.Address, params micheline.Parameters, amount int64) *Op {
	o.Contents = append(o.Contents, &Transaction{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Amount:      tezos.N(amount),
		Destination: to,
		Parameters:  &params,
	})
	return o
}

// WithOrigination adds a contract origination transaction to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithOrigination(script micheline.Script) *Op {
	o.Contents = append(o.Contents, &Origination{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Script: script,
	})
	return o
}

// WithOriginationExt adds a contract origination transaction with optional delegation to
// baker and an optional value transfer to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithOriginationExt(script micheline.Script, baker tezos.Address, amount int64) *Op {
	o.Contents = append(o.Contents, &Origination{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Balance:  tezos.N(amount),
		Delegate: baker,
		Script:   script,
	})
	return o
}

// WithDelegation adds a delegation transaction to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithDelegation(to tezos.Address) *Op {
	o.Contents = append(o.Contents, &Delegation{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Delegate: to,
	})
	return o
}

// WithUndelegation adds a delegation transaction that resets the callers baker to null
// to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithUndelegation() *Op {
	o.Contents = append(o.Contents, &Delegation{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
	})
	return o
}

// WithRegisterBaker adds a delegation transaction that registers the caller
// as baker to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithRegisterBaker() *Op {
	o.Contents = append(o.Contents, &Delegation{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Delegate: o.Source,
	})
	return o
}

// WithSetBakerParams adds a set_delegate_parameters call where target is
// source. The caller must be a registered baker.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithSetBakerParams(edge, limit int64) *Op {
	return o.WithCall(
		o.Source,
		micheline.Parameters{
			Entrypoint: micheline.SET_DELEGATE_PARAMETERS,
			Value: micheline.NewCombPair(
				micheline.NewInt64(edge),
				micheline.NewInt64(limit),
				micheline.Unit,
			),
		},
	)
}

// WithStake sends a stake pseudo call to source to lock tokens for staking.
// The caller must delegate to a baker and this baker is implictly chosen to
// stake with.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithStake(amount int64) *Op {
	return o.WithCallExt(
		o.Source,
		micheline.Parameters{
			Entrypoint: micheline.STAKE,
			Value:      micheline.Unit,
		},
		amount,
	)
}

// WithUnstake sends an unstake pseudo call to source which creates an
// unstake request for amount tokens.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithUnstake(amount int64) *Op {
	return o.WithCallExt(
		o.Source,
		micheline.Parameters{
			Entrypoint: micheline.UNSTAKE,
			Value:      micheline.Unit,
		},
		amount,
	)
}

// WithUnstakeAll sends an unstake pseudo call to source which creates an
// unstake request for all currently staked tokens.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithUnstakeAll(amount int64) *Op {
	return o.WithCallExt(
		o.Source,
		micheline.Parameters{
			Entrypoint: micheline.UNSTAKE,
			Value:      micheline.Unit,
		},
		9223372036854775807,
	)
}

// WithFinalizeUnstake sends a finalize_unstake pseudo call to source which
// moves all unfrozen unstaked tokens back to spendable balance.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithFinalizeUnstake() *Op {
	return o.WithCall(
		o.Source,
		micheline.Parameters{
			Entrypoint: micheline.FINALIZE_UNSTAKE,
			Value:      micheline.Unit,
		},
	)
}

// WithRegisterConstant adds a global constant registration transaction to the contents list.
// Source must be defined via WithSource() before calling this function.
func (o *Op) WithRegisterConstant(value micheline.Prim) *Op {
	o.Contents = append(o.Contents, &RegisterGlobalConstant{
		Manager: Manager{
			Source:  o.Source,
			Counter: 0,
		},
		Value: value,
	})
	return o
}

// WithTTL sets a time-to-live for the operation in number of blocks. This may be
// used as a convenience method instead of setting a branch directly, but requires
// to use an autocomplete handler, wallet or custom function that fetches the hash
// of block head~N as branch. Note that serialization will fail until a brach is set.
func (o *Op) WithTTL(n int64) *Op {
	if n > o.Params.MaxOperationsTTL {
		n = o.Params.MaxOperationsTTL - 2 // Ithaca adjusted
	} else if n < 0 {
		n = 1
	}
	o.TTL = n
	return o
}

// WithBranch sets the branch for this operation to hash.
func (o *Op) WithBranch(hash tezos.BlockHash) *Op {
	o.Branch = hash
	return o
}

// WithChainId sets chain_id for this operation to id. Use this only for remote signing
// of (pre)endorsements as it creates an invalid binary encoding otherwise.
func (o *Op) WithChainId(id tezos.ChainIdHash) *Op {
	clone := id.Clone()
	o.ChainId = &clone
	return o
}

// WithLimits sets the limits (fee, gas and storage limit) of each
// contained operation to provided limits. Use this to apply values from
// simulation with an optional safety margin on gas and storage. This will also
// calculate the minFee for each operation in the list and add the minFee
// for header bytes (branch and signature) to the first operation in a list.
//
// Setting a user-defined fee for each individual operation is only honored
// when its higher than minFee. Note that when sending batch operations all
// fees must be >= the individual minFee. Otherwise the minFee rule will
// apply to all zero/lower fee operations and the entire batch may overpay
// (e.g. if you have the first operation pay all fees for example and set
// remaining fees to zero).
func (o *Op) WithLimits(limits []tezos.Limits, margin int64) *Op {
	for i, v := range o.Contents {
		if len(limits) < i {
			continue
		}
		// apply simulated limit to get a better size estimate
		v.WithLimits(limits[i])

		// re-calculate limits with safety margins
		gas := limits[i].GasLimit + margin
		storage := limits[i].StorageLimit
		if storage > 0 {
			storage += margin
		}
		adj := tezos.Limits{
			GasLimit:     gas,
			StorageLimit: storage,
		}

		// Apply limits, and re-compute the fee if needed.
		// This is required, because fee value has an impact on operation size.
		var lastFee int64 = -1
		for lastFee < adj.Fee {
			lastFee = adj.Fee

			adj.Fee = max64(limits[i].Fee, CalculateMinFee(v, gas, i == 0, o.Params))
			v.WithLimits(adj)
		}
	}
	return o
}

func (o *Op) WithMinFee() *Op {
	for i, v := range o.Contents {
		// extend current limit with minimum fee estimate based on size + gas
		lim := v.Limits()

		adj := tezos.Limits{
			GasLimit:     lim.GasLimit,
			StorageLimit: lim.StorageLimit,
			Fee:          max64(lim.Fee, CalculateMinFee(v, lim.GasLimit, i == 0, o.Params)),
		}

		// use adjusted limits
		v.WithLimits(adj)
	}
	return o
}

// Limits returns the sum of all limits (fee, gas, storage limit) currently
// set for all contained operations.
func (o Op) Limits() tezos.Limits {
	var l tezos.Limits
	for _, v := range o.Contents {
		l = l.Add(v.Limits())
	}
	return l
}

// Bytes serializes the operation into binary form. When no signature is set, the
// result can be used as input for signing, if a signature is set the result is
// ready to be broadcast. Returns a nil slice when branch or contents are empty.
func (o *Op) Bytes() []byte {
	if len(o.Contents) == 0 || !o.Branch.IsValid() {
		return nil
	}
	p := o.Params
	if p == nil {
		p = tezos.DefaultParams
	}
	buf := bytes.NewBuffer(nil)
	buf.Write(o.Branch.Bytes())
	for _, v := range o.Contents {
		_ = v.EncodeBuffer(buf, p)
	}
	switch o.Contents[0].Kind() {
	case tezos.OpTypeEndorsementWithSlot:
		// no signature
	default:
		if o.Signature.IsValid() {
			buf.Write(o.Signature.Data) // raw, without type (!)
		}
	}
	return buf.Bytes()
}

// WatermarkedBytes serializes the operation and prefixes it with a watermark.
// This format is only used for signing. Watermarked data is not useful anywhere
// else.
func (o *Op) WatermarkedBytes() []byte {
	if len(o.Contents) == 0 || !o.Branch.IsValid() {
		return nil
	}
	p := o.Params
	if p == nil {
		p = tezos.DefaultParams
	}
	buf := bytes.NewBuffer(nil)
	switch o.Contents[0].Kind() {
	case tezos.OpTypeEndorsement, tezos.OpTypeEndorsementWithSlot:
		if p.OperationTagsVersion < 2 {
			buf.WriteByte(EmmyEndorsementWatermark)
		} else {
			buf.WriteByte(TenderbakeEndorsementWatermark)
		}
		if o.ChainId != nil {
			buf.Write(o.ChainId.Bytes())
		}
	case tezos.OpTypePreendorsement:
		buf.WriteByte(TenderbakePreendorsementWatermark)
		if o.ChainId != nil {
			buf.Write(o.ChainId.Bytes())
		}
	default:
		buf.WriteByte(OperationWatermark)
	}
	buf.Write(o.Branch.Bytes())
	for _, v := range o.Contents {
		_ = v.EncodeBuffer(buf, p)
	}
	return buf.Bytes()
}

// Digest returns a 32 byte blake2b hash for signing the operation. The pre-image
// is the binary serialized operation (without signature) prefixed with a
// type-dependent watermark byte.
func (o *Op) Digest() []byte {
	d := tezos.Digest(o.WatermarkedBytes())
	return d[:]
}

// WithSignature adds an externally created signature to the operation.
// No signature validation is performed, it is assumed the signature is correct.
func (o *Op) WithSignature(sig tezos.Signature) *Op {
	o.Signature = sig
	return o
}

// Sign signs the operation using provided private key. If a valid signature
// already exists this function is a noop. Fails when either branch or contents
// are empty.
func (o *Op) Sign(key tezos.PrivateKey) error {
	if !o.Branch.IsValid() {
		return fmt.Errorf("tezos: missing branch")
	}
	if len(o.Contents) == 0 {
		return fmt.Errorf("tezos: empty operation contents")
	}
	sig, err := key.Sign(o.Digest())
	if err != nil {
		return err
	}
	o.Signature = sig
	return nil
}

// Hash calculates the operation hash. For the hash to be correct, the operation
// must contain a valid signature.
func (o *Op) Hash() (h tezos.OpHash) {
	d := tezos.Digest(o.Bytes())
	copy(h[:], d[:])
	return
}

// MarshalJSON conditionally marshals the JSON format of the operation with checks
// for required fields. Omits signature for unsigned ops so that the encoding is
// compatible with remote forging.
func (o *Op) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"branch":`)
	buf.WriteString(strconv.Quote(o.Branch.String()))
	buf.WriteString(`,"contents":[`)
	for i, op := range o.Contents {
		if i > 0 {
			buf.WriteByte(',')
		}
		if b, err := op.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buf.Write(b)
		}
	}
	buf.WriteByte(']')
	sig := o.Signature
	if len(o.Contents) > 0 && o.Contents[0].Kind() == tezos.OpTypeEndorsementWithSlot {
		// no signature
		sig = tezos.InvalidSignature
	}
	if sig.IsValid() {
		buf.WriteString(`,"signature":`)
		buf.WriteString(strconv.Quote(sig.String()))
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// DecodeOp decodes an operation from its binary representation. The encoded
// data may or may not contain a signature.
func DecodeOp(data []byte) (*Op, error) {
	// check for shortest message
	if len(data) < 32+5 {
		return nil, io.ErrShortBuffer
	}

	// decode
	buf := bytes.NewBuffer(data)
	o := &Op{
		Contents: make([]Operation, 0),
		Params:   tezos.DefaultParams,
	}
	if err := o.Branch.UnmarshalBinary(buf.Next(32)); err != nil {
		return nil, err
	}
	for buf.Len() > 0 {
		var op Operation
		tag, _ := buf.ReadByte()
		buf.UnreadByte()
		switch tezos.ParseOpTag(tag) {
		case tezos.OpTypeEndorsement:
			if o.Params.OperationTagsVersion < 2 {
				op = new(Endorsement)
			} else {
				op = new(TenderbakeEndorsement)
			}
		case tezos.OpTypePreendorsement:
			op = new(TenderbakePreendorsement)
		case tezos.OpTypeEndorsementWithSlot:
			op = new(EndorsementWithSlot)
		case tezos.OpTypeSeedNonceRevelation:
			op = new(SeedNonceRevelation)
		case tezos.OpTypeDoubleEndorsementEvidence:
			if o.Params.OperationTagsVersion < 2 {
				op = new(DoubleEndorsementEvidence)
			} else {
				op = new(TenderbakeDoubleEndorsementEvidence)
			}
		case tezos.OpTypeDoublePreendorsementEvidence:
			op = new(TenderbakeDoublePreendorsementEvidence)
		case tezos.OpTypeDoubleBakingEvidence:
			op = new(DoubleBakingEvidence)
		case tezos.OpTypeActivateAccount:
			op = new(ActivateAccount)
		case tezos.OpTypeProposals:
			op = new(Proposals)
		case tezos.OpTypeBallot:
			op = new(Ballot)
		case tezos.OpTypeReveal:
			op = new(Reveal)
		case tezos.OpTypeTransaction:
			op = new(Transaction)
		case tezos.OpTypeOrigination:
			op = new(Origination)
		case tezos.OpTypeDelegation:
			op = new(Delegation)
		case tezos.OpTypeFailingNoop:
			op = new(FailingNoop)
		case tezos.OpTypeRegisterConstant:
			op = new(RegisterGlobalConstant)
		case tezos.OpTypeSetDepositsLimit:
			op = new(SetDepositsLimit)
		case tezos.OpTypeTransferTicket:
			op = new(TransferTicket)
		case tezos.OpTypeVdfRevelation:
			op = new(VdfRevelation)
		case tezos.OpTypeIncreasePaidStorage:
			op = new(IncreasePaidStorage)
		case tezos.OpTypeDrainDelegate:
			op = new(DrainDelegate)
		case tezos.OpTypeUpdateConsensusKey:
			op = new(UpdateConsensusKey)
		case tezos.OpTypeSmartRollupOriginate:
			op = new(SmartRollupOriginate)
		case tezos.OpTypeSmartRollupAddMessages:
			op = new(SmartRollupAddMessages)
		case tezos.OpTypeSmartRollupCement:
			op = new(SmartRollupCement)
		case tezos.OpTypeSmartRollupPublish:
			op = new(SmartRollupPublish)
		// TODO
		// case tezos.OpTypeSmartRollupRefute:
		// 	op = new(SmartRollupRefute)
		case tezos.OpTypeSmartRollupTimeout:
			op = new(SmartRollupTimeout)
		case tezos.OpTypeSmartRollupExecuteOutboxMessage:
			op = new(SmartRollupExecuteOutboxMessage)
		case tezos.OpTypeSmartRollupRecoverBond:
			op = new(SmartRollupRecoverBond)
		case tezos.OpTypeDalAttestation:
			op = new(DalAttestation)
		case tezos.OpTypeDalPublishSlotHeader:
			op = new(DalPublishSlotHeader)

		default:
			// stop if rest looks like a signature
			// FIXME: BLS sigs are 96 bytes, but accepting this here will
			// collide with detecting valid operation types in a batch
			if buf.Len() == 64 {
				break
			}
			return nil, fmt.Errorf("tezos: unsupported operation tag %d", tag)
		}
		if err := op.DecodeBuffer(buf, tezos.DefaultParams); err != nil {
			return nil, err
		}
		o.Contents = append(o.Contents, op)
	}

	if buf.Len() > 0 {
		// FIXME: BLS sigs are 96 byte
		if err := o.Signature.UnmarshalBinary(buf.Next(64)); err != nil {
			return nil, err
		}
	}
	return o, nil
}
