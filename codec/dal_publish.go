// Copyright (c) 2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// DalPublishCommitment represents "Dal_publish_commitment" operation
type DalPublishCommitment struct {
	Manager
	Level      int32          `json:"level"`
	Index      byte           `json:"index"`
	Commitment tezos.HexBytes `json:"commitment"`
	Proof      tezos.HexBytes `json:"commitment_proof"`
}

func (o DalPublishCommitment) Kind() tezos.OpType {
	return tezos.OpTypeDalPublishCommitment
}

func (o DalPublishCommitment) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"slot_header":{`)
	buf.WriteString(`"level":`)
	buf.WriteString(strconv.Itoa(int(o.Level)))
	buf.WriteString(`,"index":`)
	buf.WriteString(strconv.Itoa(int(o.Index)))
	buf.WriteString(`,"commitment":`)
	buf.WriteString(strconv.Quote(o.Commitment.String()))
	buf.WriteString(`,"commitment_proof":`)
	buf.WriteString(strconv.Quote(o.Proof.String()))
	buf.WriteString("}}")
	return buf.Bytes(), nil
}

func (o DalPublishCommitment) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	binary.Write(buf, enc, o.Level)
	binary.Write(buf, enc, o.Index)
	buf.Write(o.Commitment.Bytes())
	buf.Write(o.Proof.Bytes())
	return nil
}

func (o *DalPublishCommitment) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	if o.Level, err = readInt32(buf.Next(4)); err != nil {
		return
	}
	if o.Index, err = readByte(buf.Next(1)); err != nil {
		return
	}
	if err = o.Commitment.ReadBytes(buf, 48); err != nil {
		return
	}
	if err = o.Proof.ReadBytes(buf, 48); err != nil {
		return
	}
	return
}

func (o DalPublishCommitment) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *DalPublishCommitment) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
