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
	Rollup     tezos.Address         `json:"rollup"`
	Opponent   tezos.Address         `json:"opponent"`
	Refutation SmartRollupRefutation `json:"refutation"`
}

type SmartRollupRefutation struct {
	Kind         string                      `json:"refutation_kind"`
	PlayerHash   tezos.SmartRollupCommitHash `json:"player_commitment_hash"`
	OpponentHash tezos.SmartRollupCommitHash `json:"opponent_commitment_hash"`
	Choice       tezos.Z                     `json:"choice"`
	Step         SmartRollupRefuteStep       `json:"step"`
}

type SmartRollupRefuteStep struct {
	Ticks []SmartRollupTick
	Proof *SmartRollupProof
}

type SmartRollupProof struct {
	PvmStep    tezos.HexBytes        `json:"pvm_step"`
	InputProof SmartRollupInputProof `json:"input_proof"`
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

func (o SmartRollupRefute) Kind() tezos.OpType {
	return tezos.OpTypeSmartRollupRefute
}

func (o SmartRollupRefute) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	// TODO if needed
	return buf.Bytes(), nil
}

func (o SmartRollupRefute) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	// TODO if needed
	return nil
}

func (o *SmartRollupRefute) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	// TODO if needed
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
