// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// VdfRevelation represents "vdf_revelation" operation
type VdfRevelation struct {
	Simple
	Solution tezos.HexBytes `json:"solution"`
}

func (o VdfRevelation) Kind() tezos.OpType {
	return tezos.OpTypeVdfRevelation
}

func (o VdfRevelation) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"solution":`)
	buf.WriteString(strconv.Quote(o.Solution.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o VdfRevelation) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	buf.Write(o.Solution.Bytes())
	return nil
}

func (o *VdfRevelation) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	return o.Solution.ReadBytes(buf, 200)
}

func (o VdfRevelation) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *VdfRevelation) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
