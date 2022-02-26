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
		return fmt.Errorf("tezos: invalid operation status '%s'", string(data))
	}
	*t = v
	return nil
}

func (t OpStatus) MarshalText() ([]byte, error) {
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

// enums are allocated in chronological order
const (
	OpTypeInvalid                      OpType = iota
	OpTypeActivateAccount                     // 1
	OpTypeDoubleBakingEvidence                // 2
	OpTypeDoubleEndorsementEvidence           // 3
	OpTypeSeedNonceRevelation                 // 4
	OpTypeTransaction                         // 5
	OpTypeOrigination                         // 6
	OpTypeDelegation                          // 7
	OpTypeReveal                              // 8
	OpTypeEndorsement                         // 9
	OpTypeProposals                           // 10
	OpTypeBallot                              // 11
	OpTypeFailingNoop                         // 12 v009
	OpTypeEndorsementWithSlot                 // 13 v009
	OpTypeRegisterConstant                    // 14 v011
	OpTypePreendorsement                      // 15 v012
	OpTypeDoublePreendorsementEvidence        // 16 v012
	OpTypeSetDepositsLimit                    // 17 v012
)

func (t OpType) IsValid() bool {
	return t != OpTypeInvalid
}

func (t *OpType) UnmarshalText(data []byte) error {
	v := ParseOpType(string(data))
	if !v.IsValid() {
		return fmt.Errorf("tezos: invalid operation type '%s'", string(data))
	}
	*t = v
	return nil
}

func (t OpType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func ParseOpType(s string) OpType {
	switch s {
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
	case "endorsement":
		return OpTypeEndorsement
	case "endorsement_with_slot":
		return OpTypeEndorsementWithSlot
	case "proposals":
		return OpTypeProposals
	case "ballot":
		return OpTypeBallot
	case "failing_noop":
		return OpTypeFailingNoop
	case "register_global_constant":
		return OpTypeRegisterConstant
	case "preendorsement":
		return OpTypePreendorsement
	case "double_preendorsement_evidence":
		return OpTypeDoublePreendorsementEvidence
	case "set_deposits_limit":
		return OpTypeSetDepositsLimit
	default:
		return OpTypeInvalid
	}
}

func (t OpType) String() string {
	switch t {
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
	case OpTypeEndorsementWithSlot:
		return "endorsement_with_slot"
	case OpTypeProposals:
		return "proposals"
	case OpTypeBallot:
		return "ballot"
	case OpTypeFailingNoop:
		return "failing_noop"
	case OpTypeRegisterConstant:
		return "register_global_constant"
	case OpTypePreendorsement:
		return "preendorsement"
	case OpTypeDoublePreendorsementEvidence:
		return "double_preendorsement_evidence"
	case OpTypeSetDepositsLimit:
		return "set_deposits_limit"
	default:
		return ""
	}
}

var (
	// before Babylon v005
	opTagV0 = map[OpType]byte{
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
	// Babylon v005 until Hangzhou v011
	opTagV1 = map[OpType]byte{
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
		OpTypeEndorsementWithSlot:       10,  // v009
		OpTypeFailingNoop:               17,  // v009
		OpTypeRegisterConstant:          111, // v011
	}
	// Itahca v012 and up
	opTagV2 = map[OpType]byte{
		OpTypeSeedNonceRevelation:          1,
		OpTypeDoubleEndorsementEvidence:    2,
		OpTypeDoubleBakingEvidence:         3,
		OpTypeActivateAccount:              4,
		OpTypeProposals:                    5,
		OpTypeBallot:                       6,
		OpTypeReveal:                       107, // v005
		OpTypeTransaction:                  108, // v005
		OpTypeOrigination:                  109, // v005
		OpTypeDelegation:                   110, // v005
		OpTypeFailingNoop:                  17,  // v009
		OpTypeRegisterConstant:             111, // v011
		OpTypePreendorsement:               20,  // v012
		OpTypeEndorsement:                  21,  // v012
		OpTypeDoublePreendorsementEvidence: 7,   // v012
		OpTypeSetDepositsLimit:             112, // v012
	}
)

func (t OpType) TagVersion(ver int) byte {
	var (
		tag byte
		ok  bool
	)
	switch ver {
	case 0:
		tag, ok = opTagV0[t]
	case 1:
		tag, ok = opTagV1[t]
	default:
		tag, ok = opTagV2[t]
	}
	if !ok {
		return 255
	}
	return tag
}

func (t OpType) Tag() byte {
	tag, ok := opTagV2[t]
	if !ok {
		tag = 255
	}
	return tag
}

var (
	// before Babylon v005
	opMinSizeV0 = map[byte]int{
		0:  5,         // OpTypeEndorsement
		1:  37,        // OpTypeSeedNonceRevelation
		2:  9 + 2*101, // OpTypeDoubleEndorsementEvidence
		3:  9 + 2*189, // OpTypeDoubleBakingEvidence (w/o seed_nonce_hash)
		4:  41,        // OpTypeActivateAccount
		5:  30,        // OpTypeProposals
		6:  59,        // OpTypeBallot
		7:  26 + 32,   // OpTypeReveal (assuming shortest pk)
		8:  49,        // OpTypeTransaction
		9:  53,        // OpTypeOrigination
		10: 26,        // OpTypeDelegation
	}
	// Babylon v005 until Hangzhou v011
	opMinSizeV1 = map[byte]int{
		0:   5,          // OpTypeEndorsement // <v009
		1:   37,         // OpTypeSeedNonceRevelation
		2:   11 + 2*101, // OpTypeDoubleEndorsementEvidence
		3:   9 + 2*189,  // OpTypeDoubleBakingEvidence (w/o seed_nonce_hash, lb_escape_vote)
		4:   41,         // OpTypeActivateAccount
		5:   30,         // OpTypeProposals
		6:   59,         // OpTypeBallot
		107: 26 + 32,    // OpTypeReveal // v005 (assuming shortest pk)
		108: 50,         // OpTypeTransaction // v005
		109: 28,         // OpTypeOrigination // v005
		110: 27,         // OpTypeDelegation // v005
		10:  108,        // OpTypeEndorsementWithSlot // v009
		17:  5,          // OpTypeFailingNoop  // v009
		111: 30,         // OpTypeRegisterConstant // v011
	}
	// Ithaca v012 and up
	opMinSizeV2 = map[byte]int{
		1:   37,               // OpTypeSeedNonceRevelation
		2:   9 + 2*(32+43+64), // OpTypeDoubleEndorsementEvidence
		3:   9 + 2*237,        // OpTypeDoubleBakingEvidence (w/o seed_nonce_hash, min fitness size)
		4:   41,               // OpTypeActivateAccount
		5:   30,               // OpTypeProposals
		6:   59,               // OpTypeBallot
		107: 26 + 32,          // OpTypeReveal // v005 (assuming shortest pk)
		108: 50,               // OpTypeTransaction // v005
		109: 28,               // OpTypeOrigination // v005
		110: 27,               // OpTypeDelegation // v005
		17:  5,                // OpTypeFailingNoop  // v009
		111: 30,               // OpTypeRegisterConstant // v011
		7:   9 + 2*(32+43+64), // OpTypeDoublePreendorsementEvidence // v012
		20:  43,               // OpTypePreendorsement // v012
		21:  43,               // OpTypeEndorsement // v012
		112: 27,               // OpTypeSetDepositsLimit // v012
	}
)

func (t OpType) MinSizeVersion(ver int) int {
	switch ver {
	case 0:
		return opMinSizeV0[t.TagVersion(ver)]
	case 1:
		return opMinSizeV1[t.TagVersion(ver)]
	default:
		return opMinSizeV2[t.TagVersion(ver)]
	}
}

func (t OpType) MinSize() int {
	return opMinSizeV2[t.Tag()]
}

func (t OpType) ListId() int {
	switch t {
	case OpTypeEndorsement, OpTypeEndorsementWithSlot, OpTypePreendorsement:
		return 0
	case OpTypeProposals, OpTypeBallot:
		return 1
	case OpTypeActivateAccount,
		OpTypeDoubleBakingEvidence,
		OpTypeDoubleEndorsementEvidence,
		OpTypeSeedNonceRevelation,
		OpTypeDoublePreendorsementEvidence:
		return 2
	case OpTypeTransaction, // generic user operations
		OpTypeOrigination,
		OpTypeDelegation,
		OpTypeReveal,
		OpTypeRegisterConstant,
		OpTypeSetDepositsLimit:
		return 3
	default:
		return -1 // invalid
	}
}

func ParseOpTag(t byte) OpType {
	switch t {
	case 0:
		return OpTypeEndorsement
	case 10:
		return OpTypeEndorsementWithSlot
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
	case 107:
		return OpTypeReveal
	case 108:
		return OpTypeTransaction
	case 109:
		return OpTypeOrigination
	case 110:
		return OpTypeDelegation
	case 17:
		return OpTypeFailingNoop
	case 111:
		return OpTypeRegisterConstant
	case 20:
		return OpTypePreendorsement
	case 21:
		return OpTypeEndorsement
	case 7:
		return OpTypeDoublePreendorsementEvidence
	case 112:
		return OpTypeSetDepositsLimit
	default:
		return OpTypeInvalid
	}
}

func ParseOpTagVersion(t byte, ver int) OpType {
	tags := opTagV0
	switch ver {
	case 1:
		tags = opTagV1
	case 2:
		tags = opTagV2
	}
	for typ, tag := range tags {
		if tag == t {
			return typ
		}
	}
	return OpTypeInvalid
}
