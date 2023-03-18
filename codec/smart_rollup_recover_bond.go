// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// SmartRollupRecoverBond represents "smart_rollup_recover_bond" operation
type SmartRollupRecoverBond struct {
	Manager
	Rollup tezos.Address `json:"rollup"`
	Staker tezos.Address `json:"staker"`
}

func (o SmartRollupRecoverBond) Kind() tezos.OpType {
	return tezos.OpTypeSmartRollupRecoverBond
}

func (o SmartRollupRecoverBond) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"rollup":`)
	buf.WriteString(strconv.Quote(o.Rollup.String()))
	buf.WriteString(`,"staker":`)
	buf.WriteString(strconv.Quote(o.Staker.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o SmartRollupRecoverBond) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	buf.Write(o.Rollup.Hash()) // 20 byte only
	buf.Write(o.Staker.Encode())
	return nil
}

func (o *SmartRollupRecoverBond) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	o.Rollup = tezos.NewAddress(tezos.AddressTypeSmartRollup, buf.Next(20))
	err = o.Staker.Decode(buf.Next(21))
	return
}

func (o SmartRollupRecoverBond) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *SmartRollupRecoverBond) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
