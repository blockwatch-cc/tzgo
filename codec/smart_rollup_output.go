// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// SmartRollupExecuteOutboxMessage represents "smart_rollup_execute_outbox_message" operation
type SmartRollupExecuteOutboxMessage struct {
	Manager
	Rollup   tezos.Address               `json:"rollup"`
	Cemented tezos.SmartRollupCommitHash `json:"cemented_commitment"`
	Proof    tezos.HexBytes              `json:"output_proof"`
}

func (o SmartRollupExecuteOutboxMessage) Kind() tezos.OpType {
	return tezos.OpTypeSmartRollupExecuteOutboxMessage
}

func (o SmartRollupExecuteOutboxMessage) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"rollup":`)
	buf.WriteString(strconv.Quote(o.Rollup.String()))
	buf.WriteString(`,"cemented_commitment":`)
	buf.WriteString(strconv.Quote(o.Cemented.String()))
	buf.WriteString(`,"output_proof":`)
	buf.WriteString(strconv.Quote(o.Proof.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o SmartRollupExecuteOutboxMessage) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	buf.Write(o.Rollup.Hash()) // 20 byte only
	buf.Write(o.Cemented[:])
	writeBytesWithLen(buf, o.Proof)
	return nil
}

func (o *SmartRollupExecuteOutboxMessage) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	o.Rollup = tezos.NewAddress(tezos.AddressTypeSmartRollup, buf.Next(20))
	o.Cemented = tezos.NewSmartRollupCommitHash(buf.Next(32))
	o.Proof, err = readBytesWithLen(buf)
	return
}

func (o SmartRollupExecuteOutboxMessage) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *SmartRollupExecuteOutboxMessage) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
