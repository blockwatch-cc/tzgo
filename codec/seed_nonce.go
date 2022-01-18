// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "encoding/binary"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// SeedNonceRevelation represents "seed_nonce_revelation" operation
type SeedNonceRevelation struct {
    Level int32          `json:"level,string"`
    Nonce tezos.HexBytes `json:"nonce"`
}

func (o SeedNonceRevelation) Kind() tezos.OpType {
    return tezos.OpTypeSeedNonceRevelation
}

func (o SeedNonceRevelation) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteString(`,"level":`)
    buf.WriteString(strconv.Itoa(int(o.Level)))
    buf.WriteString(`,"nonce":`)
    buf.WriteString(strconv.Quote(o.Nonce.String()))
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o SeedNonceRevelation) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    binary.Write(buf, enc, o.Level)
    buf.Write(o.Nonce.Bytes())
    return nil
}

func (o *SeedNonceRevelation) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return
    }
    o.Level, err = readInt32(buf.Next(4))
    if err != nil {
        return
    }
    o.Nonce = make([]byte, 32)
    copy(o.Nonce, buf.Next(32))
    return nil
}

func (o SeedNonceRevelation) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *SeedNonceRevelation) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
