// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "encoding/binary"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// FailingNoop represents "failing_noop" operations. Used for signing arbitrary messages
// and guaranteed to be not included on-chain. This prevents an attack vector where a
// message is crafted which looks like a regular transaction.
type FailingNoop struct {
    Arbitrary string `json:"arbitrary"`
}

func (o FailingNoop) Kind() tezos.OpType {
    return tezos.OpTypeFailingNoop
}

func (o FailingNoop) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteString(`,"arbitrary":`)
    buf.WriteString(strconv.Quote(o.Arbitrary))
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o FailingNoop) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    // convert utf8 string to bytes
    val := []byte(o.Arbitrary)
    binary.Write(buf, enc, uint32(len(val)))
    buf.Write(val)
    return nil
}

func (o *FailingNoop) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    l, err := readUint32(buf.Next(4))
    if err != nil {
        return err
    }
    val := make([]byte, l)
    copy(val, buf.Next(int(l)))
    o.Arbitrary = string(val)
    return nil
}

func (o FailingNoop) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *FailingNoop) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
