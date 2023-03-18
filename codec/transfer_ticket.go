// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"strconv"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// TransferTicket represents "transfer_ticket" operation
type TransferTicket struct {
	Manager
	Contents    micheline.Prim `json:"ticket_contents"`
	Type        micheline.Prim `json:"ticket_ty"`
	Ticketer    tezos.Address  `json:"ticket_ticketer"`
	Amount      tezos.N        `json:"ticket_amount"`
	Destination tezos.Address  `json:"destination"`
	Entrypoint  string         `json:"entrypoint"`
}

func (o TransferTicket) Kind() tezos.OpType {
	return tezos.OpTypeTransferTicket
}

func (o TransferTicket) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"ticket_contents":`)
	o.Contents.EncodeJSON(buf)
	buf.WriteString(`,"ticket_ty":`)
	o.Type.EncodeJSON(buf)
	buf.WriteString(`,"ticketer":`)
	buf.WriteString(strconv.Quote(o.Ticketer.String()))
	buf.WriteString(`,"amount":`)
	buf.WriteString(strconv.Quote(o.Amount.String()))
	buf.WriteString(`,"destination":`)
	buf.WriteString(strconv.Quote(o.Destination.String()))
	buf.WriteString(`,"entrypoint":`)
	buf.WriteString(strconv.Quote(o.Entrypoint))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o TransferTicket) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	writePrimWithLen(buf, o.Contents)
	writePrimWithLen(buf, o.Type)
	buf.Write(o.Ticketer.EncodePadded())
	o.Amount.EncodeBuffer(buf)
	buf.Write(o.Destination.EncodePadded())
	writeStringWithLen(buf, o.Entrypoint)
	return nil
}

func (o *TransferTicket) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	if o.Contents, err = readPrimWithLen(buf); err != nil {
		return
	}
	if o.Type, err = readPrimWithLen(buf); err != nil {
		return
	}
	if err = o.Ticketer.Decode(buf.Next(22)); err != nil {
		return
	}
	if err = o.Amount.DecodeBuffer(buf); err != nil {
		return
	}
	if err = o.Destination.Decode(buf.Next(22)); err != nil {
		return
	}
	o.Entrypoint, err = readStringWithLen(buf)
	return
}

func (o TransferTicket) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *TransferTicket) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
