// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "fmt"
    "io"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// ActivateAccount represents "activate_account" operation
type ActivateAccount struct {
    PublicKeyHash tezos.Address  `json:"pkh"`
    Secret        tezos.HexBytes `json:"secret"`
}

func (o ActivateAccount) Kind() tezos.OpType {
    return tezos.OpTypeActivateAccount
}

func (o ActivateAccount) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteString(`,"pkh":`)
    buf.WriteString(strconv.Quote(o.PublicKeyHash.String()))
    buf.WriteString(`,"secret":`)
    buf.WriteString(strconv.Quote(o.Secret.String()))
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o ActivateAccount) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    buf.Write(o.PublicKeyHash.Hash) // only place where a 20 byte address is used (!)
    buf.Write(o.Secret.Bytes())
    return nil
}

func (o *ActivateAccount) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    if err := ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
        return err
    }
    o.PublicKeyHash = tezos.NewAddress(tezos.AddressTypeEd25519, buf.Next(20))
    if !o.PublicKeyHash.IsValid() {
        return fmt.Errorf("invalid address type=%s len=%d", o.PublicKeyHash.Type, len(o.PublicKeyHash.Hash))
    }
    o.Secret = make([]byte, 20)
    copy(o.Secret, buf.Next(20))
    if len(o.Secret) != 20 {
        return io.ErrShortBuffer
    }
    return nil
}

func (o ActivateAccount) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *ActivateAccount) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
