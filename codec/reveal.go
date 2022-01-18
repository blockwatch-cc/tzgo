// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// Reveal represents "reveal" operation
type Reveal struct {
    Manager
    PublicKey tezos.Key `json:"public_key"`
}

func (o Reveal) Kind() tezos.OpType {
    return tezos.OpTypeReveal
}

func (o Reveal) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteByte(',')
    o.Manager.EncodeJSON(buf)
    buf.WriteString(`,"public_key":`)
    buf.WriteString(strconv.Quote(o.PublicKey.String()))
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o Reveal) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    o.Manager.EncodeBuffer(buf, p)
    buf.Write(o.PublicKey.Bytes())
    return nil
}

func (o *Reveal) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    if err = o.Manager.DecodeBuffer(buf, p); err != nil {
        return err
    }
    if err = o.PublicKey.DecodeBuffer(buf); err != nil {
        return
    }
    return nil
}

func (o Reveal) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *Reveal) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
