// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
    "encoding/json"
    "fmt"

    "blockwatch.cc/tzgo/micheline"
    "blockwatch.cc/tzgo/tezos"
)

// Ensure SmartRollup types implement the TypedOperation interface.
var (
    _ TypedOperation = (*SmartRollupOrigination)(nil)
    _ TypedOperation = (*SmartRollupAddMessages)(nil)
    _ TypedOperation = (*SmartRollupCement)(nil)
    _ TypedOperation = (*SmartRollupPublish)(nil)
    _ TypedOperation = (*SmartRollupRefute)(nil)
    _ TypedOperation = (*SmartRollupTimeout)(nil)
    _ TypedOperation = (*SmartRollupExecuteOutboxMessage)(nil)
    _ TypedOperation = (*SmartRollupRecoverBond)(nil)
)

type SmartRollupOrigination struct {
    Manager
    PvmKind          tezos.PvmKind  `json:"pvm_kind"`
    Kernel           tezos.HexBytes `json:"kernel"`
    OriginationProof tezos.HexBytes `json:"origination_proof"`
    ParametersTy     micheline.Prim `json:"parameters_ty"`
}

type SmartRollupAddMessages struct {
    Manager
    Messages []tezos.HexBytes `json:"message"`
}

type SmartRollupCement struct {
    Manager
    Rollup     tezos.Address               `json:"rollup"`
    Commitment tezos.SmartRollupCommitHash `json:"commitment"`
}

type SmartRollupPublish struct {
    Manager
    Rollup     tezos.Address `json:"rollup"`
    Commitment struct {
        CompressedState tezos.SmartRollupStateHash  `json:"compressed_state"`
        InboxLevel      int64                       `json:"inbox_level"`
        Predecessor     tezos.SmartRollupCommitHash `json:"predecessor"`
        NumberOfTicks   tezos.Z                     `json:"number_of_ticks"`
    } `json:"commitment"`
}

type SmartRollupRefute struct {
    Manager
    Rollup     tezos.Address `json:"rollup"`
    Opponent   tezos.Address `json:"opponent"`
    Refutation struct {
        Kind         string                      `json:"refutation_kind"`
        PlayerHash   tezos.SmartRollupCommitHash `json:"player_commitment_hash"`
        OpponentHash tezos.SmartRollupCommitHash `json:"opponent_commitment_hash"`
        Choice       tezos.Z                     `json:"choice"`
        Step         *SmartRollupRefuteStep      `json:"step"`
    } `json:"refutation"`
}

// Step can either be
//
// - []SmartRollupTick
// - SmartRollupInputProof
// - smth else?
//
// There is no indication in the outer parts of the refutation struct that
// suggests how to decode this.
type SmartRollupRefuteStep struct {
    Ticks []SmartRollupTick
    Proof *SmartRollupInputProof
}

func (s *SmartRollupRefuteStep) UnmarshalJSON(buf []byte) error {
    if len(buf) == 0 {
        return nil
    }
    switch buf[0] {
    case '[':
        s.Ticks = make([]SmartRollupTick, 0)
        return json.Unmarshal(buf, &s.Ticks)
    case '{':
        s.Proof = &SmartRollupInputProof{}
        return json.Unmarshal(buf, s.Proof)
    default:
        return fmt.Errorf("Invalid refute step data %q", string(buf))
    }
    return nil
}

func (s SmartRollupRefuteStep) MarshalJSON() ([]byte, error) {
    if s.Ticks != nil {
        return json.Marshal(s.Ticks)
    }
    if s.Proof != nil {
        return json.Marshal(s.Proof)
    }
    return nil, nil
}

type SmartRollupTick struct {
    State tezos.SmartRollupStateHash `json:"state"`
    Tick  tezos.Z                    `json:"tick"`
}

type SmartRollupInputProof struct {
    Kind    string         `json:"input_proof_kind"`
    Level   int64          `json:"level"`
    Counter tezos.Z        `json:"message_counter"`
    Proof   tezos.HexBytes `json:"serialized_proof"`
}

type SmartRollupTimeout struct {
    Manager
    Rollup  tezos.Address `json:"rollup"`
    Stakers struct {
        Alice tezos.Address `json:"alice"`
        Bob   tezos.Address `json:"bob"`
    } `json:"stakers"`
}

type SmartRollupExecuteOutboxMessage struct {
    Manager
    Rollup             tezos.Address               `json:"rollup"`
    CementedCommitment tezos.SmartRollupCommitHash `json:"cemented_commitment"`
    OutputProof        tezos.HexBytes              `json:"output_proof"`
}

type SmartRollupRecoverBond struct {
    Manager
    Rollup tezos.Address `json:"rollup"`
    Staker tezos.Address `json:"staker"`
}

type GameStatus struct {
    Status string        `json:"-"`
    Kind   string        `json:"kind"`
    Reason string        `json:"reason"`
    Player tezos.Address `json:"player"`
}

func (s *GameStatus) UnmarshalJSON(buf []byte) error {
    if len(buf) == 0 {
        return nil
    }
    switch buf[0] {
    case '"':
        s.Status = string(buf[1 : len(buf)-1])
    case '{':
        type alias struct {
            S *GameStatus `json:"result"`
        }
        return json.Unmarshal(buf, &alias{s})
    default:
        return fmt.Errorf("Invalid game status data %q", string(buf))
    }
    return nil
}

func (s GameStatus) MarshalJSON() ([]byte, error) {
    if s.Status != "" {
        return []byte(`"` + s.Status + `"`), nil
    }
    type alias struct {
        S GameStatus `json:"result"`
    }
    return json.Marshal(alias{s})
}
