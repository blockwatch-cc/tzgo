// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "encoding/binary"
    "encoding/json"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// DoubleBakingEvidence represents "double_baking_evidence" operation
type DoubleBakingEvidence struct {
    Bh1 BlockHeader `json:"bh1"`
    Bh2 BlockHeader `json:"bh2"`
}

func (o DoubleBakingEvidence) Kind() tezos.OpType {
    return tezos.OpTypeDoubleBakingEvidence
}

func (o DoubleBakingEvidence) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteString(`,"bh1":`)
    enc := json.NewEncoder(buf)
    enc.Encode(o.Bh1)
    buf.WriteString(`,"bh2":`)
    enc.Encode(o.Bh2)
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o DoubleBakingEvidence) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    b2 := bytes.NewBuffer(nil)
    o.Bh1.EncodeBuffer(b2)
    binary.Write(buf, enc, uint32(b2.Len()))
    buf.Write(b2.Bytes())
    b2.Reset()
    o.Bh2.EncodeBuffer(b2)
    binary.Write(buf, enc, uint32(b2.Len()))
    buf.Write(b2.Bytes())
    return nil
}

func (o *DoubleBakingEvidence) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    var l int32
    l, err = readInt32(buf.Next(4))
    if err != nil {
        return
    }
    if err = o.Bh1.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l)))); err != nil {
        return
    }
    l, err = readInt32(buf.Next(4))
    if err != nil {
        return
    }
    if err = o.Bh2.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l)))); err != nil {
        return
    }
    return
}

func (o DoubleBakingEvidence) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *DoubleBakingEvidence) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
