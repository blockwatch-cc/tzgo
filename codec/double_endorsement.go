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

// DoubleEndorsementEvidence represents "double_endorsement_evidence" operation
type DoubleEndorsementEvidence struct {
    Op1  InlinedEndorsement `json:"op1"`
    Op2  InlinedEndorsement `json:"op2"`
    Slot int16              `json:"slot"`
}

func (o DoubleEndorsementEvidence) Kind() tezos.OpType {
    return tezos.OpTypeDoubleEndorsementEvidence
}

func (o DoubleEndorsementEvidence) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    enc := json.NewEncoder(buf)
    buf.WriteString(`,"op1":`)
    enc.Encode(o.Op1)
    buf.WriteString(`,"op2":`)
    enc.Encode(o.Op2)
    buf.WriteString(`,"slot":`)
    buf.WriteString(strconv.Itoa(int(o.Slot)))
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o DoubleEndorsementEvidence) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    b2 := bytes.NewBuffer(nil)
    o.Op1.EncodeBuffer(b2, p)
    binary.Write(buf, enc, uint32(b2.Len()))
    buf.Write(b2.Bytes())
    b2.Reset()
    o.Op2.EncodeBuffer(b2, p)
    binary.Write(buf, enc, uint32(b2.Len()))
    buf.Write(b2.Bytes())
    binary.Write(buf, enc, o.Slot)
    return nil
}

func (o *DoubleEndorsementEvidence) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    l, err := readInt32(buf.Next(4))
    if err != nil {
        return err
    }
    if err = o.Op1.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
        return err
    }
    l, err = readInt32(buf.Next(4))
    if err != nil {
        return err
    }
    if err = o.Op2.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
        return err
    }
    o.Slot, err = readInt16(buf.Next(2))
    if err != nil {
        return err
    }
    return nil
}

func (o DoubleEndorsementEvidence) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *DoubleEndorsementEvidence) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
