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
        CompressedState string `json:"compressed_state"`
        InboxLevel      int64  `json:"inbox_level"`
        Predecessor     string `json:"predecessor"`
        NumberOfTicks   int64  `json:"number_of_ticks,string"`
    } `json:"commitment"`
}

type SmartRollupRefute struct {
    Manager
    Rollup     tezos.Address `json:"rollup"`
    Opponent   tezos.Address `json:"opponent"`
    Refutation struct {
        Kind         string          `json:"refutation_kind"`
        PlayerHash   string          `json:"player_commitment_hash"`
        OpponentHash string          `json:"opponent_commitment_hash"`
        Choice       tezos.Z         `json:"choice"`
        Step         json.RawMessage `json:"step"`
    } `json:"refutation"`
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
    Rollup             tezos.Address  `json:"rollup"`
    CementedCommitment string         `json:"cemented_commitment"`
    OutputProof        tezos.HexBytes `json:"output_proof"`
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
        type alias *GameStatus
        return json.Unmarshal(buf, alias(s))
    default:
        return fmt.Errorf("Invalid game status data %q", string(buf))
    }
    return nil
}
