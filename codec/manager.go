// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// Manager contains fields common for all manager operations
type Manager struct {
    Source       tezos.Address `json:"source"`
    Fee          tezos.N       `json:"fee"`
    Counter      tezos.N       `json:"counter"`
    GasLimit     tezos.N       `json:"gas_limit"`
    StorageLimit tezos.N       `json:"storage_limit"`
}

func (o Manager) EncodeJSON(buf *bytes.Buffer) error {
    buf.WriteString(`"source":`)
    buf.WriteString(strconv.Quote(o.Source.String()))
    buf.WriteString(`,"fee":`)
    buf.WriteString(strconv.Quote(o.Fee.String()))
    buf.WriteString(`,"counter":`)
    buf.WriteString(strconv.Quote(o.Counter.String()))
    buf.WriteString(`,"gas_limit":`)
    buf.WriteString(strconv.Quote(o.GasLimit.String()))
    buf.WriteString(`,"storage_limit":`)
    buf.WriteString(strconv.Quote(o.StorageLimit.String()))
    return nil
}

func (o Manager) EncodeBuffer(buf *bytes.Buffer, _ *tezos.Params) error {
    buf.Write(o.Source.Bytes())
    o.Fee.EncodeBuffer(buf)
    o.Counter.EncodeBuffer(buf)
    o.GasLimit.EncodeBuffer(buf)
    o.StorageLimit.EncodeBuffer(buf)
    return nil
}

func (o *Manager) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
    if err = o.Source.UnmarshalBinary(buf.Next(21)); err != nil {
        return
    }
    if err = o.Fee.DecodeBuffer(buf); err != nil {
        return err
    }
    if err = o.Counter.DecodeBuffer(buf); err != nil {
        return err
    }
    if err = o.GasLimit.DecodeBuffer(buf); err != nil {
        return err
    }
    if err = o.StorageLimit.DecodeBuffer(buf); err != nil {
        return err
    }
    return nil
}

func (o *Manager) WithCounter(c int64) *Manager {
    o.Counter.SetInt64(c)
    return o
}

func (o *Manager) WithFee(fee int64) *Manager {
    o.Fee.SetInt64(fee)
    return o
}

func (o *Manager) WithGasLimit(limit int64) *Manager {
    o.GasLimit.SetInt64(limit)
    return o
}

func (o *Manager) WithStorageLimit(limit int64) *Manager {
    o.StorageLimit.SetInt64(limit)
    return o
}
