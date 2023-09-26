// Copyright (c) 2020-2023 Blockwatch Data Inc.
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
	OpTypeInvalid                         OpType = iota
	OpTypeActivateAccount                        // 1
	OpTypeDoubleBakingEvidence                   // 2
	OpTypeDoubleEndorsementEvidence              // 3
	OpTypeSeedNonceRevelation                    // 4
	OpTypeTransaction                            // 5
	OpTypeOrigination                            // 6
	OpTypeDelegation                             // 7
	OpTypeReveal                                 // 8
	OpTypeEndorsement                            // 9
	OpTypeProposals                              // 10
	OpTypeBallot                                 // 11
	OpTypeFailingNoop                            // 12 v009
	OpTypeEndorsementWithSlot                    // 13 v009
	OpTypeRegisterConstant                       // 14 v011
	OpTypePreendorsement                         // 15 v012
	OpTypeDoublePreendorsementEvidence           // 16 v012
	OpTypeSetDepositsLimit                       // 17 v012
	OpTypeTxRollupOrigination                    // 18 v013 DEPRECATED in v016
	OpTypeTxRollupSubmitBatch                    // 19 v013 DEPRECATED in v016
	OpTypeTxRollupCommit                         // 20 v013 DEPRECATED in v016
	OpTypeTxRollupReturnBond                     // 21 v013 DEPRECATED in v016
	OpTypeTxRollupFinalizeCommitment             // 22 v013 DEPRECATED in v016
	OpTypeTxRollupRemoveCommitment               // 23 v013 DEPRECATED in v016
	OpTypeTxRollupRejection                      // 24 v013 DEPRECATED in v016
	OpTypeTxRollupDispatchTickets                // 25 v013 DEPRECATED in v016
	OpTypeTransferTicket                         // 26 v013
	OpTypeVdfRevelation                          // 27 v014
	OpTypeIncreasePaidStorage                    // 28 v014
	OpTypeEvent                                  // 29 v014 (only in internal_operation_results)
	OpTypeDrainDelegate                          // 30 v015
	OpTypeUpdateConsensusKey                     // 31 v015
	OpTypeSmartRollupOriginate                   // 32 v016
	OpTypeSmartRollupAddMessages                 // 33 v016
	OpTypeSmartRollupCement                      // 34 v016
	OpTypeSmartRollupPublish                     // 35 v016
	OpTypeSmartRollupRefute                      // 36 v016
	OpTypeSmartRollupTimeout                     // 37 v016
	OpTypeSmartRollupExecuteOutboxMessage        // 38 v016
	OpTypeSmartRollupRecoverBond                 // 39 v016
	OpTypeDalAttestation                         // 40 v016+?
	OpTypeDalPublishSlotHeader                   // 41 v016+?
)

var (
	opTypeStrings = map[OpType]string{
		OpTypeInvalid:                         "",
		OpTypeActivateAccount:                 "activate_account",
		OpTypeDoubleBakingEvidence:            "double_baking_evidence",
		OpTypeDoubleEndorsementEvidence:       "double_endorsement_evidence",
		OpTypeSeedNonceRevelation:             "seed_nonce_revelation",
		OpTypeTransaction:                     "transaction",
		OpTypeOrigination:                     "origination",
		OpTypeDelegation:                      "delegation",
		OpTypeReveal:                          "reveal",
		OpTypeEndorsement:                     "endorsement",
		OpTypeProposals:                       "proposals",
		OpTypeBallot:                          "ballot",
		OpTypeFailingNoop:                     "failing_noop",
		OpTypeEndorsementWithSlot:             "endorsement_with_slot",
		OpTypeRegisterConstant:                "register_global_constant",
		OpTypePreendorsement:                  "preendorsement",
		OpTypeDoublePreendorsementEvidence:    "double_preendorsement_evidence",
		OpTypeSetDepositsLimit:                "set_deposits_limit",
		OpTypeTxRollupOrigination:             "tx_rollup_origination",
		OpTypeTxRollupSubmitBatch:             "tx_rollup_submit_batch",
		OpTypeTxRollupCommit:                  "tx_rollup_commit",
		OpTypeTxRollupReturnBond:              "tx_rollup_return_bond",
		OpTypeTxRollupFinalizeCommitment:      "tx_rollup_finalize_commitment",
		OpTypeTxRollupRemoveCommitment:        "tx_rollup_remove_commitment",
		OpTypeTxRollupRejection:               "tx_rollup_rejection",
		OpTypeTxRollupDispatchTickets:         "tx_rollup_dispatch_tickets",
		OpTypeTransferTicket:                  "transfer_ticket",
		OpTypeVdfRevelation:                   "vdf_revelation",
		OpTypeIncreasePaidStorage:             "increase_paid_storage",
		OpTypeEvent:                           "event",
		OpTypeDrainDelegate:                   "drain_delegate",
		OpTypeUpdateConsensusKey:              "update_consensus_key",
		OpTypeSmartRollupOriginate:            "smart_rollup_originate",
		OpTypeSmartRollupAddMessages:          "smart_rollup_add_messages",
		OpTypeSmartRollupCement:               "smart_rollup_cement",
		OpTypeSmartRollupPublish:              "smart_rollup_publish",
		OpTypeSmartRollupRefute:               "smart_rollup_refute",
		OpTypeSmartRollupTimeout:              "smart_rollup_timeout",
		OpTypeSmartRollupExecuteOutboxMessage: "smart_rollup_execute_outbox_message",
		OpTypeSmartRollupRecoverBond:          "smart_rollup_recover_bond",
		OpTypeDalAttestation:                  "dal_attestation",
		OpTypeDalPublishSlotHeader:            "dal_publish_slot_header",

		// rename: endorsement -> attetstaion
		// OpTypeDoubleEndorsementEvidence:       "double_attestation_evidence",
		// OpTypeEndorsement:                     "attestation",
		// OpTypePreendorsement:                  "preattestation",
		// OpTypeDoublePreendorsementEvidence:    "double_preattestation_evidence",
	}
	opTypeReverseStrings = make(map[string]OpType)
)

func init() {
	for n, v := range opTypeStrings {
		opTypeReverseStrings[v] = n
	}
}

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
	t, ok := opTypeReverseStrings[s]
	if !ok {
		t = OpTypeInvalid
	}
	return t
}

func (t OpType) String() string {
	return opTypeStrings[t]
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
	// Ithaca v012 and up
	opTagV2 = map[OpType]byte{
		OpTypeSeedNonceRevelation:             1,
		OpTypeDoubleEndorsementEvidence:       2,
		OpTypeDoubleBakingEvidence:            3,
		OpTypeActivateAccount:                 4,
		OpTypeProposals:                       5,
		OpTypeBallot:                          6,
		OpTypeReveal:                          107, // v005
		OpTypeTransaction:                     108, // v005
		OpTypeOrigination:                     109, // v005
		OpTypeDelegation:                      110, // v005
		OpTypeFailingNoop:                     17,  // v009
		OpTypeRegisterConstant:                111, // v011
		OpTypePreendorsement:                  20,  // v012
		OpTypeEndorsement:                     21,  // v012
		OpTypeDoublePreendorsementEvidence:    7,   // v012
		OpTypeSetDepositsLimit:                112, // v012
		OpTypeTxRollupOrigination:             150, // v013
		OpTypeTxRollupSubmitBatch:             151, // v013
		OpTypeTxRollupCommit:                  152, // v013
		OpTypeTxRollupReturnBond:              153, // v013
		OpTypeTxRollupFinalizeCommitment:      154, // v013
		OpTypeTxRollupRemoveCommitment:        155, // v013
		OpTypeTxRollupRejection:               156, // v013
		OpTypeTxRollupDispatchTickets:         157, // v013
		OpTypeTransferTicket:                  158, // v013
		OpTypeVdfRevelation:                   8,   // v014
		OpTypeIncreasePaidStorage:             113, // v014
		OpTypeDrainDelegate:                   9,   // v015
		OpTypeUpdateConsensusKey:              114, // v015
		OpTypeSmartRollupOriginate:            200, // v016
		OpTypeSmartRollupAddMessages:          201, // v016
		OpTypeSmartRollupCement:               202, // v016
		OpTypeSmartRollupPublish:              203, // v016
		OpTypeSmartRollupRefute:               204, // v016
		OpTypeSmartRollupTimeout:              205, // v016
		OpTypeSmartRollupExecuteOutboxMessage: 206, // v016
		OpTypeSmartRollupRecoverBond:          207, // v016
		OpTypeDalPublishSlotHeader:            230, // v016+
		OpTypeDalAttestation:                  22,  // v016+
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
		1:   37,                       // OpTypeSeedNonceRevelation
		2:   9 + 2*(32+43+64),         // OpTypeDoubleEndorsementEvidence
		3:   9 + 2*237,                // OpTypeDoubleBakingEvidence (w/o seed_nonce_hash, min fitness size)
		4:   41,                       // OpTypeActivateAccount
		5:   30,                       // OpTypeProposals
		6:   59,                       // OpTypeBallot
		107: 26 + 32,                  // OpTypeReveal // v005 (assuming shortest pk)
		108: 50,                       // OpTypeTransaction // v005
		109: 28,                       // OpTypeOrigination // v005
		110: 27,                       // OpTypeDelegation // v005
		17:  5,                        // OpTypeFailingNoop  // v009
		111: 30,                       // OpTypeRegisterConstant // v011
		7:   9 + 2*(32+43+64),         // OpTypeDoublePreendorsementEvidence // v012
		20:  43,                       // OpTypePreendorsement // v012
		21:  43,                       // OpTypeEndorsement // v012
		112: 27,                       // OpTypeSetDepositsLimit // v012
		8:   201,                      // OpTypeVdfRevelation // v014
		113: 27 + 22,                  // OpTypeIncreasePaidStorage // v014
		9:   1 + 3*21,                 // OpTypeDrainDelegate // v015
		114: 26 + 32,                  // OpTypeUpdateConsensusKey // v015
		158: 26 + 8 + 22 + 1 + 22 + 4, // OpTypeTransferTicket // v013
		200: 26 + 13,                  // OpTypeSmartRollupOriginate // v016
		201: 26 + 4,                   // OpTypeSmartRollupAddMessages // v016
		202: 26 + 52,                  // OpTypeSmartRollupCement // v016
		203: 26 + 96,                  // OpTypeSmartRollupPublish // v016
		204: 26 + 41,                  // OpTypeSmartRollupRefute // v016
		205: 26 + 62,                  // OpTypeSmartRollupTimeout // v016
		206: 26 + 56,                  // OpTypeSmartRollupExecuteOutboxMessage // v016
		207: 26 + 41,                  // OpTypeSmartRollupRecoverBond // v016
		230: 26 + 101,                 // OpTypeDalPublishSlotHeader // v016+
		22:  1 + 21 + 1 + 4,           // OpTypeDalAttestation  // v016+
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
		OpTypeDoublePreendorsementEvidence,
		OpTypeVdfRevelation,
		OpTypeDrainDelegate,
		OpTypeDalAttestation:
		return 2
	case OpTypeTransaction, // generic user operations
		OpTypeOrigination,
		OpTypeDelegation,
		OpTypeReveal,
		OpTypeRegisterConstant,
		OpTypeSetDepositsLimit,
		OpTypeTxRollupOrigination,
		OpTypeTxRollupSubmitBatch,
		OpTypeTxRollupCommit,
		OpTypeTxRollupReturnBond,
		OpTypeTxRollupFinalizeCommitment,
		OpTypeTxRollupRemoveCommitment,
		OpTypeTxRollupRejection,
		OpTypeTxRollupDispatchTickets,
		OpTypeTransferTicket,
		OpTypeUpdateConsensusKey,
		OpTypeSmartRollupOriginate,
		OpTypeSmartRollupAddMessages,
		OpTypeSmartRollupCement,
		OpTypeSmartRollupPublish,
		OpTypeSmartRollupRefute,
		OpTypeSmartRollupTimeout,
		OpTypeSmartRollupExecuteOutboxMessage,
		OpTypeSmartRollupRecoverBond,
		OpTypeDalPublishSlotHeader:
		return 3
	default:
		return -1 // invalid
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
	case 7:
		return OpTypeDoublePreendorsementEvidence
	case 8:
		return OpTypeVdfRevelation
	case 9:
		return OpTypeDrainDelegate
	case 10:
		return OpTypeEndorsementWithSlot
	case 17:
		return OpTypeFailingNoop
	case 20:
		return OpTypePreendorsement
	case 21:
		return OpTypeEndorsement
	case 22:
		return OpTypeDalAttestation
	case 107:
		return OpTypeReveal
	case 108:
		return OpTypeTransaction
	case 109:
		return OpTypeOrigination
	case 110:
		return OpTypeDelegation
	case 111:
		return OpTypeRegisterConstant
	case 112:
		return OpTypeSetDepositsLimit
	case 113:
		return OpTypeIncreasePaidStorage
	case 114:
		return OpTypeUpdateConsensusKey
	case 150:
		return OpTypeTxRollupOrigination
	case 151:
		return OpTypeTxRollupSubmitBatch
	case 152:
		return OpTypeTxRollupCommit
	case 153:
		return OpTypeTxRollupReturnBond
	case 154:
		return OpTypeTxRollupFinalizeCommitment
	case 155:
		return OpTypeTxRollupRemoveCommitment
	case 156:
		return OpTypeTxRollupRejection
	case 157:
		return OpTypeTxRollupDispatchTickets
	case 158:
		return OpTypeTransferTicket
	case 200:
		return OpTypeSmartRollupOriginate
	case 201:
		return OpTypeSmartRollupAddMessages
	case 202:
		return OpTypeSmartRollupCement
	case 203:
		return OpTypeSmartRollupPublish
	case 204:
		return OpTypeSmartRollupRefute
	case 205:
		return OpTypeSmartRollupTimeout
	case 206:
		return OpTypeSmartRollupExecuteOutboxMessage
	case 207:
		return OpTypeSmartRollupRecoverBond
	case 230:
		return OpTypeDalPublishSlotHeader
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
