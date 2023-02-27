// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    // "encoding/binary"
    // "io"
    // "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// Smart_rollup_refute (tag 204)
// =============================

// | Name          | Size                 | Contents               |
// +===============+======================+========================+
// | Tag           | 1 byte               | unsigned 8-bit integer |
// | source        | 21 bytes             | $public_key_hash       |
// | fee           | Determined from data | $N.t                   |
// | counter       | Determined from data | $N.t                   |
// | gas_limit     | Determined from data | $N.t                   |
// | storage_limit | Determined from data | $N.t                   |
// | rollup        | 20 bytes             | bytes                  |
// | opponent      | 21 bytes             | $public_key_hash       |
// | refutation    | Determined from data | $X_24                  |

// X_20 (Determined from data, 8-bit tag)
// **************************************

// raw data proof (tag 0)
// ======================

// | Name                  | Size     | Contents                |
// +=======================+==========+=========================+
// | Tag                   | 1 byte   | unsigned 8-bit integer  |
// | # bytes in next field | 2 bytes  | unsigned 16-bit integer |
// | raw_data              | Variable | bytes                   |

// metadata proof (tag 1)
// ======================

// | Name | Size   | Contents               |
// +======+========+========================+
// | Tag  | 1 byte | unsigned 8-bit integer |

// X_21 (Determined from data, 8-bit tag)
// **************************************

// inbox proof (tag 0)
// ===================

// | Name                  | Size                 | Contents                |
// +=======================+======================+=========================+
// | Tag                   | 1 byte               | unsigned 8-bit integer  |
// | level                 | 4 bytes              | signed 32-bit integer   |
// | message_counter       | Determined from data | $N.t                    |
// | # bytes in next field | 4 bytes              | unsigned 30-bit integer |
// | serialized_proof      | Variable             | bytes                   |

// reveal proof (tag 1)
// ====================

// | Name         | Size                 | Contents               |
// +==============+======================+========================+
// | Tag          | 1 byte               | unsigned 8-bit integer |
// | reveal_proof | Determined from data | $X_20                  |

// first input (tag 2)
// ===================

// | Name | Size   | Contents               |
// +======+========+========================+
// | Tag  | 1 byte | unsigned 8-bit integer |

// X_22
// ****

// | Name                        | Size                 | Contents                            |
// +=============================+======================+=====================================+
// | ? presence of field "state" | 1 byte               | boolean (0 for false, 255 for true) |
// | state                       | 32 bytes             | bytes                               |
// | tick                        | Determined from data | $N.t                                |

// X_23 (Determined from data, 8-bit tag)
// **************************************

// Dissection (tag 0)
// ==================

// | Name                  | Size     | Contents                |
// +=======================+==========+=========================+
// | Tag                   | 1 byte   | unsigned 8-bit integer  |
// | # bytes in next field | 4 bytes  | unsigned 30-bit integer |
// | Unnamed field 0       | Variable | sequence of $X_22       |

// Proof (tag 1)
// =============

// | Name                              | Size                 | Contents                            |
// +===================================+======================+=====================================+
// | Tag                               | 1 byte               | unsigned 8-bit integer              |
// | # bytes in next field             | 4 bytes              | unsigned 30-bit integer             |
// | pvm_step                          | Variable             | bytes                               |
// | ? presence of field "input_proof" | 1 byte               | boolean (0 for false, 255 for true) |
// | input_proof                       | Determined from data | $X_21                               |

// X_24 (Determined from data, 8-bit tag)
// **************************************

// Start (tag 0)
// =============

// | Name                     | Size     | Contents               |
// +==========================+==========+========================+
// | Tag                      | 1 byte   | unsigned 8-bit integer |
// | player_commitment_hash   | 32 bytes | bytes                  |
// | opponent_commitment_hash | 32 bytes | bytes                  |

// Move (tag 1)
// ============

// | Name   | Size                 | Contents               |
// +========+======================+========================+
// | Tag    | 1 byte               | unsigned 8-bit integer |
// | choice | Determined from data | $N.t                   |
// | step   | Determined from data | $X_23                  |

// SmartRollupRefute represents "smart_rollup_refute" operation
type SmartRollupRefute struct {
    Manager
    Rollup     tezos.Address `json:"rollup"`
    Opponent   tezos.Address `json:"opponent"`
    Refutation struct {
        Kind string `json:"refutation_kind"`
        // start
        Player   *tezos.SmartRollupCommitHash `json:"player_commitment_hash,omitempty"`
        Opponent *tezos.SmartRollupCommitHash `json:"opponent_commitment_hash,omitempty"`
        // move
        Choice *tezos.N `json:"choice,omitempty"`
        Step   any      `json:"step,omitempty"`
    } `json:"refutation"`
}

type SmartRollupRefuteStep struct {
    State tezos.SmartRollupStateHash `json:"state"`
    Tick  tezos.N                    `json:"tick"`
}

type SmartRollupRefuteProof struct {
    PvmStep    tezos.HexBytes `json:"pvm_step"`
    InputProof struct {
        InputProofKind   string         `json:"input_proof_kind"`
        Level            int32          `json:"level,omitempty"`
        MessageCounter   *tezos.Z       `json:"message_counter,omitempty"`
        Serialized_proof tezos.HexBytes `json:"serialized_proof,omitempty"`
        RevealProof      *struct {
            RawData   tezos.HexBytes `json:"raw_data",omitempty`
            DalPageId *struct {
                PublishedLevel int32 `json:"published_level"`
                SlotIndex      byte  `json:"slot_index"`
                PageIndex      int16 `json:"page_index"`
            } `json:"dal_page_id,omitempty"`
            DalProof tezos.HexBytes `json:"dal_proof"`
        } `json:"reveal_proof,omitempty"`
    } `json:"input_proof"`
}

func (o SmartRollupRefute) Kind() tezos.OpType {
    return tezos.OpTypeSmartRollupRefute
}

func (o SmartRollupRefute) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    // buf.WriteByte('{')
    // buf.WriteString(`"kind":`)
    // buf.WriteString(strconv.Quote(o.Kind().String()))
    // buf.WriteByte(',')
    // o.Manager.EncodeJSON(buf)
    // buf.WriteString(`,"rollup":`)
    // buf.WriteString(strconv.Quote(o.Rollup.String()))
    // buf.WriteString(`,"commitment":{`)
    // buf.WriteString(`"compressed_state":`)
    // buf.WriteString(strconv.Quote(o.Commitment.State.String()))
    // buf.WriteString(`,"inbox_level":`)
    // buf.WriteString(strconv.FormatInt(o.Commitment.InboxLevel, 10))
    // buf.WriteString(`",predecessor":`)
    // buf.WriteString(strconv.Quote(o.Commitment.Predecessor.String()))
    // buf.WriteString(`,"number_of_ticks":`)
    // buf.WriteString(strconv.FormatInt(o.Commitment.NumberOfTicks, 10))
    // buf.WriteString("}}")
    return buf.Bytes(), nil
}

func (o SmartRollupRefute) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    o.Manager.EncodeBuffer(buf, p)
    // buf.Write(o.Rollup.Hash()) // 20 byte only
    // buf.Write(o.Commitment.State)
    // binary.Write(buf, enc, uint32(o.Commitment.InboxLevel))
    // buf.Write(o.Commitment.Predecessor)
    // binary.Write(buf, enc, o.Commitment.NumberOfTicks)
    return nil
}

func (o *SmartRollupRefute) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    if err = o.Manager.DecodeBuffer(buf, p); err != nil {
        return
    }
    // o.Rollup = tezos.NewAddress(tezos.SmartRollupAddress, buf.Next(20))
    // o.Commitment.State = tezos.NewSmartRollupStateHash(buf.Next(32))
    // o.Commitment.InboxLevel, err = readInt32(buf.Next(4))
    // if err != nil {
    //     return
    // }
    // o.Commitment.Predecessor = tezos.NewSmartRollupStateHash(buf.Next(32))
    // o.Commitment.NumberOfTicks, err = readInt64(buf.Next(8))
    return
}

func (o SmartRollupRefute) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *SmartRollupRefute) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
