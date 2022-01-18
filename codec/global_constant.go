// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "encoding/binary"
    "strconv"

    "blockwatch.cc/tzgo/micheline"
    "blockwatch.cc/tzgo/tezos"
)

// RegisterGlobalConstant represents "register_global_constant" operation
type RegisterGlobalConstant struct {
    Manager
    Value micheline.Prim `json:"value"`
}

func (o RegisterGlobalConstant) Kind() tezos.OpType {
    return tezos.OpTypeRegisterConstant
}

func (o RegisterGlobalConstant) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteByte(',')
    o.Manager.EncodeJSON(buf)
    buf.WriteString(`,"value":`)
    b, _ := o.Value.MarshalJSON()
    buf.Write(b)
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o RegisterGlobalConstant) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    o.Manager.EncodeBuffer(buf, p)
    b2 := bytes.NewBuffer(nil)
    o.Value.EncodeBuffer(b2)
    binary.Write(buf, enc, uint32(b2.Len()))
    buf.Write(b2.Bytes())
    return nil
}

func (o *RegisterGlobalConstant) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    if err = o.Manager.DecodeBuffer(buf, p); err != nil {
        return err
    }
    _ = buf.Next(4)
    if err = o.Value.DecodeBuffer(buf); err != nil {
        return err
    }
    return nil
}

func (o RegisterGlobalConstant) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *RegisterGlobalConstant) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
