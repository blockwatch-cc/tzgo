// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "strconv"

    "blockwatch.cc/tzgo/micheline"
    "blockwatch.cc/tzgo/tezos"
)

// Transaction represents "transaction" operation
type Transaction struct {
    Manager
    Amount      tezos.N               `json:"amount"`
    Destination tezos.Address         `json:"destination"`
    Parameters  *micheline.Parameters `json:"parameters,omitempty"`
}

func (o Transaction) Kind() tezos.OpType {
    return tezos.OpTypeTransaction
}

func (o Transaction) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteByte(',')
    o.Manager.EncodeJSON(buf)
    buf.WriteString(`,"amount":`)
    buf.WriteString(strconv.Quote(o.Amount.String()))
    buf.WriteString(`,"destination":`)
    buf.WriteString(strconv.Quote(o.Destination.String()))
    if o.Parameters != nil {
        buf.WriteString(`,"parameters":`)
        b, _ := o.Parameters.MarshalJSON()
        buf.Write(b)
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o Transaction) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    o.Manager.EncodeBuffer(buf, p)
    o.Amount.EncodeBuffer(buf)
    buf.Write(o.Destination.Bytes22())
    if o.Parameters != nil {
        buf.WriteByte(0xff)
        o.Parameters.EncodeBuffer(buf)
    } else {
        buf.WriteByte(0x0)
    }
    return nil
}

func (o *Transaction) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    if err = o.Manager.DecodeBuffer(buf, p); err != nil {
        return err
    }
    if err = o.Amount.DecodeBuffer(buf); err != nil {
        return
    }
    if err = o.Destination.UnmarshalBinary(buf.Next(22)); err != nil {
        return
    }
    var ok bool
    ok, err = readBool(buf.Next(1))
    if err != nil {
        return
    }
    if ok {
        param := &micheline.Parameters{}
        if err = param.DecodeBuffer(buf); err != nil {
            return err
        }
        o.Parameters = param
    }
    return nil
}

func (o Transaction) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *Transaction) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
