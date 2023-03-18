// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// SmartRollupAddMessages represents "smart_rollup_add_messages" operation
type SmartRollupAddMessages struct {
	Manager
	Messages []tezos.HexBytes `json:"messages"`
}

func (o SmartRollupAddMessages) Kind() tezos.OpType {
	return tezos.OpTypeSmartRollupAddMessages
}

func (o SmartRollupAddMessages) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"messages":[`)
	if len(o.Messages) > 0 {
		buf.WriteString(strconv.Quote(o.Messages[0].String()))
		if len(o.Messages) > 1 {
			for _, v := range o.Messages[1:] {
				buf.WriteByte(',')
				buf.WriteString(strconv.Quote(v.String()))
			}
		}
	}
	buf.WriteString("]}")
	return buf.Bytes(), nil
}

func (o SmartRollupAddMessages) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	var sz int
	for _, v := range o.Messages {
		sz += len(v) + 4
	}
	binary.Write(buf, enc, uint32(sz))
	for _, v := range o.Messages {
		writeBytesWithLen(buf, v)
	}
	return nil
}

func (o *SmartRollupAddMessages) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	var sz int32
	sz, err = readInt32(buf.Next(4))
	if err != nil {
		return
	}
	for sz > 0 {
		var msg tezos.HexBytes
		msg, err = readBytesWithLen(buf)
		if err != nil {
			return
		}
		o.Messages = append(o.Messages, msg)
		sz -= int32(len(msg))
	}
	return
}

func (o SmartRollupAddMessages) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *SmartRollupAddMessages) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
