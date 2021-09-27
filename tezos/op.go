// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"fmt"
)

type OpStatus byte

const (
	OpStatusInvalid OpStatus = iota // 0
	OpStatusApplied                 // 1 (success)
	OpStatusFailed
	OpStatusSkipped
	OpStatusBacktracked
)

func (t OpStatus) IsValid() bool {
	return t != OpStatusInvalid
}

func (t OpStatus) IsSuccess() bool {
	return t == OpStatusApplied
}

func (t *OpStatus) UnmarshalText(data []byte) error {
	v := ParseOpStatus(string(data))
	if !v.IsValid() {
		return fmt.Errorf("invalid operation status '%s'", string(data))
	}
	*t = v
	return nil
}

func (t *OpStatus) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func ParseOpStatus(s string) OpStatus {
	switch s {
	case "applied":
		return OpStatusApplied
	case "failed":
		return OpStatusFailed
	case "skipped":
		return OpStatusSkipped
	case "backtracked":
		return OpStatusBacktracked
	default:
		return OpStatusInvalid
	}
}

func (t OpStatus) String() string {
	switch t {
	case OpStatusApplied:
		return "applied"
	case OpStatusFailed:
		return "failed"
	case OpStatusSkipped:
		return "skipped"
	case OpStatusBacktracked:
		return "backtracked"
	default:
		return ""
	}
}

type OpType byte

const (
	OpTypeBake                      OpType = iota // 0
	OpTypeActivateAccount                         // 1
	OpTypeDoubleBakingEvidence                    // 2
	OpTypeDoubleEndorsementEvidence               // 3
	OpTypeSeedNonceRevelation                     // 4
	OpTypeTransaction                             // 5
	OpTypeOrigination                             // 6
	OpTypeDelegation                              // 7
	OpTypeReveal                                  // 8
	OpTypeEndorsement                             // 9
	OpTypeProposals                               // 10
	OpTypeBallot                                  // 11
	OpTypeUnfreeze                                // 12 indexer only
	OpTypeInvoice                                 // 13 indexer only
	OpTypeAirdrop                                 // 14 indexer only
	OpTypeSeedSlash                               // 15 indexer only
	OpTypeMigration                               // 16 indexer only
	OpTypeFailingNoop                             // 17 v009
	OpTypeRegisterConstant                        // 18 v011
	OpTypeBatch                     = 254         // indexer only, output-only
	OpTypeInvalid                   = 255
)

func (t OpType) IsValid() bool {
	return t != OpTypeInvalid
}

func (t *OpType) UnmarshalText(data []byte) error {
	v := ParseOpType(string(data))
	if !v.IsValid() {
		return fmt.Errorf("invalid operation type '%s'", string(data))
	}
	*t = v
	return nil
}

func (t *OpType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func ParseOpType(s string) OpType {
	switch s {
	case "bake":
		return OpTypeBake
	case "activate_account":
		return OpTypeActivateAccount
	case "double_baking_evidence":
		return OpTypeDoubleBakingEvidence
	case "double_endorsement_evidence":
		return OpTypeDoubleEndorsementEvidence
	case "seed_nonce_revelation":
		return OpTypeSeedNonceRevelation
	case "transaction":
		return OpTypeTransaction
	case "origination":
		return OpTypeOrigination
	case "delegation":
		return OpTypeDelegation
	case "reveal":
		return OpTypeReveal
	case "endorsement", "endorsement_with_slot":
		return OpTypeEndorsement
	case "proposals":
		return OpTypeProposals
	case "ballot":
		return OpTypeBallot
	case "unfreeze":
		return OpTypeUnfreeze
	case "invoice":
		return OpTypeInvoice
	case "airdrop":
		return OpTypeAirdrop
	case "seed_slash":
		return OpTypeSeedSlash
	case "migration":
		return OpTypeMigration
	case "batch":
		return OpTypeBatch
	case "failing_noop":
		return OpTypeFailingNoop
	case "register_global_constant":
		return OpTypeRegisterConstant
	default:
		return OpTypeInvalid
	}
}

func (t OpType) String() string {
	switch t {
	case OpTypeBake:
		return "bake"
	case OpTypeActivateAccount:
		return "activate_account"
	case OpTypeDoubleBakingEvidence:
		return "double_baking_evidence"
	case OpTypeDoubleEndorsementEvidence:
		return "double_endorsement_evidence"
	case OpTypeSeedNonceRevelation:
		return "seed_nonce_revelation"
	case OpTypeTransaction:
		return "transaction"
	case OpTypeOrigination:
		return "origination"
	case OpTypeDelegation:
		return "delegation"
	case OpTypeReveal:
		return "reveal"
	case OpTypeEndorsement:
		return "endorsement"
	case OpTypeProposals:
		return "proposals"
	case OpTypeBallot:
		return "ballot"
	case OpTypeUnfreeze:
		return "unfreeze"
	case OpTypeInvoice:
		return "invoice"
	case OpTypeAirdrop:
		return "airdrop"
	case OpTypeSeedSlash:
		return "seed_slash"
	case OpTypeMigration:
		return "migration"
	case OpTypeBatch:
		return "batch"
	case OpTypeFailingNoop:
		return "failing_noop"
	case OpTypeRegisterConstant:
		return "register_global_constant"
	default:
		return ""
	}
}

var (
	// before babylon
	opTagV1 = map[OpType]byte{
		OpTypeEndorsement:               0,
		OpTypeSeedNonceRevelation:       1,
		OpTypeDoubleEndorsementEvidence: 2,
		OpTypeDoubleBakingEvidence:      3,
		OpTypeActivateAccount:           4,
		OpTypeProposals:                 5,
		OpTypeBallot:                    6,
		OpTypeReveal:                    7,
		OpTypeTransaction:               8,
		OpTypeOrigination:               9,
		OpTypeDelegation:                10,
	}
	// Babylon v005 and up
	opTagV2 = map[OpType]byte{
		OpTypeEndorsement:               0,
		OpTypeSeedNonceRevelation:       1,
		OpTypeDoubleEndorsementEvidence: 2,
		OpTypeDoubleBakingEvidence:      3,
		OpTypeActivateAccount:           4,
		OpTypeProposals:                 5,
		OpTypeBallot:                    6,
		OpTypeReveal:                    107, // v005
		OpTypeTransaction:               108, // v005
		OpTypeOrigination:               109, // v005
		OpTypeDelegation:                110, // v005
		OpTypeFailingNoop:               17,  // v009
		OpTypeRegisterConstant:          111, // v011
	}
)

func (t OpType) Tag(p *Params) byte {
	v := 0
	if p != nil {
		v = p.OperationTagsVersion
	}
	var (
		tag byte
		ok  bool
	)
	switch v {
	case 0:
		tag, ok = opTagV1[t]
	case 1:
		tag, ok = opTagV2[t]
	default:
		tag, ok = opTagV2[t]
	}
	if !ok {
		return 255
	}
	return tag
}

func (t OpType) ListId() int {
	switch t {
	case OpTypeEndorsement:
		return 0
	case OpTypeProposals, OpTypeBallot:
		return 1
	case OpTypeActivateAccount,
		OpTypeDoubleBakingEvidence,
		OpTypeDoubleEndorsementEvidence,
		OpTypeSeedNonceRevelation:
		return 2
	case OpTypeTransaction, // generic user operations
		OpTypeOrigination,
		OpTypeDelegation,
		OpTypeReveal,
		OpTypeRegisterConstant,
		OpTypeBatch: // custom, indexer only
		return 3
	case OpTypeBake, OpTypeUnfreeze, OpTypeSeedSlash:
		return -1 // block level ops
	case OpTypeInvoice, OpTypeAirdrop, OpTypeMigration:
		return -2 // migration ops
	default:
		return -255 // invalid
	}
}

func ParseOpTag(t byte) OpType {
	switch t {
	case 0:
		return OpTypeEndorsement
	case 1:
		return OpTypeSeedNonceRevelation
	case 2:
		return OpTypeDoubleEndorsementEvidence
	case 3:
		return OpTypeDoubleBakingEvidence
	case 4:
		return OpTypeActivateAccount
	case 5:
		return OpTypeProposals
	case 6:
		return OpTypeBallot
	case 7, 107:
		return OpTypeReveal
	case 8, 108:
		return OpTypeTransaction
	case 9, 109:
		return OpTypeOrigination
	case 10, 110:
		return OpTypeDelegation
	case 17:
		return OpTypeFailingNoop
	case 111:
		return OpTypeRegisterConstant
	default:
		return OpTypeInvalid
	}
}
