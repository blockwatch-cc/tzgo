// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// UpdateConsensusKey represents "update_consensus_key" operation
type UpdateConsensusKey struct {
	Manager
	Amount    tezos.Z   `json:"amount"`
	PublicKey tezos.Key `json:"pk"`
}

func (o UpdateConsensusKey) Kind() tezos.OpType {
	return tezos.OpTypeUpdateConsensusKey
}

func (o UpdateConsensusKey) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"pk":`)
	buf.WriteString(strconv.Quote(o.PublicKey.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o UpdateConsensusKey) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	buf.Write(o.PublicKey.Bytes())
	return nil
}

func (o *UpdateConsensusKey) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return err
	}
	if err = o.PublicKey.DecodeBuffer(buf); err != nil {
		return
	}
	return
}

func (o UpdateConsensusKey) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *UpdateConsensusKey) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
