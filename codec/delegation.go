// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// Delegation represents "delegation" operation
type Delegation struct {
    Manager
    Delegate tezos.Address `json:"delegate"`
}

func (o Delegation) Kind() tezos.OpType {
    return tezos.OpTypeDelegation
}

func (o Delegation) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteByte(',')
    o.Manager.EncodeJSON(buf)
    if o.Delegate.IsValid() {
        buf.WriteString(`,"delegate":`)
        buf.WriteString(strconv.Quote(o.Delegate.String()))
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o Delegation) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    o.Manager.EncodeBuffer(buf, p)
    if o.Delegate.IsValid() {
        buf.WriteByte(0xff)
        buf.Write(o.Delegate.Bytes())
    } else {
        buf.WriteByte(0x0)
    }
    return nil
}

func (o *Delegation) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    if err = o.Manager.DecodeBuffer(buf, p); err != nil {
        return err
    }
    var ok bool
    ok, err = readBool(buf.Next(1))
    if err != nil {
        return err
    }
    if ok {
        err = o.Delegate.UnmarshalBinary(buf.Next(21))
        if err != nil {
            return err
        }
    }
    return nil
}

func (o Delegation) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *Delegation) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
