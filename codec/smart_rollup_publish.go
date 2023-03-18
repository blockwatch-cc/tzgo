// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// SmartRollupPublish represents "smart_rollup_publish" operation
type SmartRollupPublish struct {
	Manager
	Rollup     tezos.Address `json:"rollup"`
	Commitment struct {
		State         tezos.SmartRollupStateHash `json:"compressed_state"`
		InboxLevel    int32                      `json:"inbox_level"`
		Predecessor   tezos.SmartRollupStateHash `json:"predecessor"`
		NumberOfTicks int64                      `json:"number_of_ticks,string"`
	} `json:"commitment"`
}

func (o SmartRollupPublish) Kind() tezos.OpType {
	return tezos.OpTypeSmartRollupPublish
}

func (o SmartRollupPublish) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteByte(',')
	o.Manager.EncodeJSON(buf)
	buf.WriteString(`,"rollup":`)
	buf.WriteString(strconv.Quote(o.Rollup.String()))
	buf.WriteString(`,"commitment":{`)
	buf.WriteString(`"compressed_state":`)
	buf.WriteString(strconv.Quote(o.Commitment.State.String()))
	buf.WriteString(`,"inbox_level":`)
	buf.WriteString(strconv.FormatInt(int64(o.Commitment.InboxLevel), 10))
	buf.WriteString(`,"predecessor":`)
	buf.WriteString(strconv.Quote(o.Commitment.Predecessor.String()))
	buf.WriteString(`,"number_of_ticks":`)
	buf.WriteString(strconv.Quote(strconv.FormatInt(o.Commitment.NumberOfTicks, 10)))
	buf.WriteString("}}")
	return buf.Bytes(), nil
}

func (o SmartRollupPublish) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	o.Manager.EncodeBuffer(buf, p)
	buf.Write(o.Rollup.Hash()) // 20 byte only
	buf.Write(o.Commitment.State[:])
	binary.Write(buf, enc, uint32(o.Commitment.InboxLevel))
	buf.Write(o.Commitment.Predecessor[:])
	binary.Write(buf, enc, o.Commitment.NumberOfTicks)
	return nil
}

func (o *SmartRollupPublish) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	if err = o.Manager.DecodeBuffer(buf, p); err != nil {
		return
	}
	o.Rollup = tezos.NewAddress(tezos.AddressTypeSmartRollup, buf.Next(20))
	o.Commitment.State = tezos.NewSmartRollupStateHash(buf.Next(32))
	o.Commitment.InboxLevel, err = readInt32(buf.Next(4))
	if err != nil {
		return
	}
	o.Commitment.Predecessor = tezos.NewSmartRollupStateHash(buf.Next(32))
	o.Commitment.NumberOfTicks, err = readInt64(buf.Next(8))
	return
}

func (o SmartRollupPublish) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *SmartRollupPublish) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
