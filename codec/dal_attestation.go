// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// DalAttestation represents "dal_attestation" operation
type DalAttestation struct {
	Simple
	Attestor    tezos.Address `json:"attestor"`
	Attestation tezos.Z       `json:"attestation"`
	Level       int32         `json:"level"`
}

func (o DalAttestation) Kind() tezos.OpType {
	return tezos.OpTypeDalAttestation
}

func (o DalAttestation) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"attestor":`)
	buf.WriteString(strconv.Quote(o.Attestor.String()))
	buf.WriteString(`,"attestation":`)
	buf.WriteString(strconv.Quote(o.Attestation.String()))
	buf.WriteString(`,"level":`)
	buf.WriteString(strconv.Itoa(int(o.Level)))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o DalAttestation) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	buf.Write(o.Attestor.Encode())
	buf.Write(o.Attestation.Bytes())
	binary.Write(buf, enc, o.Level)
	return nil
}

func (o *DalAttestation) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Attestor.Decode(buf.Next(21)); err != nil {
		return
	}
	if err = o.Attestation.DecodeBuffer(buf); err != nil {
		return
	}
	o.Level, err = readInt32(buf.Next(4))
	return
}

func (o DalAttestation) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *DalAttestation) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
