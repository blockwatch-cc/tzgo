// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// SmartRollupTimeout represents "smart_rollup_timeout" operation
type SmartRollupTimeout struct {
	Manager
	Rollup  tezos.Address `json:"rollup"`
	Stakers struct {
		Alice tezos.Address `json:"alice"`
		Bob   tezos.Address `json:"bob"`
	} `json:"stakers"`
}

func (o SmartRollupTimeout) Kind() tezos.OpType {
	return tezos.OpTypeSmartRollupTimeout
}

func (o SmartRollupTimeout) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"rollup":`)
	buf.WriteString(strconv.Quote(o.Rollup.String()))
	buf.WriteString(`,"stakers":{`)
	buf.WriteString(`"alice":`)
	buf.WriteString(strconv.Quote(o.Stakers.Alice.String()))
	buf.WriteString(`,"bob":`)
	buf.WriteString(strconv.Quote(o.Stakers.Bob.String()))
	buf.WriteString("}}")
	return buf.Bytes(), nil
}

func (o SmartRollupTimeout) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	buf.Write(o.Rollup.Hash()) // 20 byte only
	buf.Write(o.Stakers.Alice.Encode())
	buf.Write(o.Stakers.Bob.Encode())
	return nil
}

func (o *SmartRollupTimeout) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	o.Rollup = tezos.NewAddress(tezos.AddressTypeSmartRollup, buf.Next(20))
	if err = o.Stakers.Alice.Decode(buf.Next(21)); err != nil {
		return
	}
	if err = o.Stakers.Bob.Decode(buf.Next(21)); err != nil {
		return
	}
	return
}

func (o SmartRollupTimeout) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *SmartRollupTimeout) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
