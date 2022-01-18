// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "encoding/binary"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// Proposals represents "proposals" operation
type Proposals struct {
    Source    tezos.Address        `json:"source"`
    Period    int32                `json:"period"`
    Proposals []tezos.ProtocolHash `json:"proposals"`
}

func (o Proposals) Kind() tezos.OpType {
    return tezos.OpTypeProposals
}

func (o Proposals) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteString(`,"source":`)
    buf.WriteString(strconv.Quote(o.Source.String()))
    buf.WriteString(`,"period":`)
    buf.WriteString(strconv.Itoa(int(o.Period)))
    buf.WriteString(`,"proposals":[`)
    for i, v := range o.Proposals {
        if i > 0 {
            buf.WriteByte(',')
        }
        buf.WriteString(strconv.Quote(v.String()))
    }
    buf.WriteString(`]}`)
    return buf.Bytes(), nil
}

func (o Proposals) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    buf.Write(o.Source.Bytes())
    binary.Write(buf, enc, o.Period)
    binary.Write(buf, enc, int32(len(o.Proposals)*tezos.HashTypeProtocol.Len()))
    for _, v := range o.Proposals {
        buf.Write(v.Bytes())
    }
    return nil
}

func (o *Proposals) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    if err = o.Source.UnmarshalBinary(buf.Next(21)); err != nil {
        return
    }
    o.Period, err = readInt32(buf.Next(4))
    if err != nil {
        return
    }
    l, err := readInt32(buf.Next(4))
    if err != nil {
        return err
    }
    o.Proposals = make([]tezos.ProtocolHash, l/32)
    for i := range o.Proposals {
        if err = o.Proposals[i].UnmarshalBinary(buf.Next(32)); err != nil {
            return
        }
    }
    return nil
}

func (o Proposals) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *Proposals) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
