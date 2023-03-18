// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// DoubleEndorsementEvidence represents "double_endorsement_evidence" operation
type DoubleEndorsementEvidence struct {
	Simple
	Op1  InlinedEndorsement `json:"op1"`
	Op2  InlinedEndorsement `json:"op2"`
	Slot int16              `json:"slot"`
}

func (o DoubleEndorsementEvidence) Kind() tezos.OpType {
	return tezos.OpTypeDoubleEndorsementEvidence
}

func (o DoubleEndorsementEvidence) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"op1":`)
	b, _ := o.Op1.MarshalJSON()
	buf.Write(b)
	buf.WriteString(`,"op2":`)
	b, _ = o.Op2.MarshalJSON()
	buf.Write(b)
	buf.WriteString(`,"slot":`)
	buf.WriteString(strconv.Itoa(int(o.Slot)))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o DoubleEndorsementEvidence) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	b2 := bytes.NewBuffer(nil)
	o.Op1.EncodeBuffer(b2, p)
	binary.Write(buf, enc, uint32(b2.Len()))
	buf.Write(b2.Bytes())
	b2.Reset()
	o.Op2.EncodeBuffer(b2, p)
	binary.Write(buf, enc, uint32(b2.Len()))
	buf.Write(b2.Bytes())
	binary.Write(buf, enc, o.Slot)
	return nil
}

func (o *DoubleEndorsementEvidence) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	l, err := readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	if err = o.Op1.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
		return err
	}
	l, err = readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	if err = o.Op2.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
		return err
	}
	o.Slot, err = readInt16(buf.Next(2))
	if err != nil {
		return err
	}
	return nil
}

func (o DoubleEndorsementEvidence) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *DoubleEndorsementEvidence) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}

// TenderbakeDoubleEndorsementEvidence represents "double_endorsement_evidence" operation
// for Tenderbake protocols
type TenderbakeDoubleEndorsementEvidence struct {
	Simple
	Op1 TenderbakeInlinedEndorsement `json:"op1"`
	Op2 TenderbakeInlinedEndorsement `json:"op2"`
}

func (o TenderbakeDoubleEndorsementEvidence) Kind() tezos.OpType {
	return tezos.OpTypeDoubleEndorsementEvidence
}

func (o TenderbakeDoubleEndorsementEvidence) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"op1":`)
	b, _ := o.Op1.MarshalJSON()
	buf.Write(b)
	buf.WriteString(`,"op2":`)
	b, _ = o.Op2.MarshalJSON()
	buf.Write(b)
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o TenderbakeDoubleEndorsementEvidence) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	b2 := bytes.NewBuffer(nil)
	o.Op1.EncodeBuffer(b2, p)
	binary.Write(buf, enc, uint32(b2.Len()))
	buf.Write(b2.Bytes())
	b2.Reset()
	o.Op2.EncodeBuffer(b2, p)
	binary.Write(buf, enc, uint32(b2.Len()))
	buf.Write(b2.Bytes())
	return nil
}

func (o *TenderbakeDoubleEndorsementEvidence) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	l, err := readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	if err = o.Op1.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
		return err
	}
	l, err = readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	if err = o.Op2.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
		return err
	}
	return nil
}

func (o TenderbakeDoubleEndorsementEvidence) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *TenderbakeDoubleEndorsementEvidence) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}

// TenderbakeDoublePreendorsementEvidence represents "double_preendorsement_evidence" operation
// for Tenderbake protocols
type TenderbakeDoublePreendorsementEvidence struct {
	Simple
	Op1 TenderbakeInlinedPreendorsement `json:"op1"`
	Op2 TenderbakeInlinedPreendorsement `json:"op2"`
}

func (o TenderbakeDoublePreendorsementEvidence) Kind() tezos.OpType {
	return tezos.OpTypeDoublePreendorsementEvidence
}

func (o TenderbakeDoublePreendorsementEvidence) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"op1":`)
	b, _ := o.Op1.MarshalJSON()
	buf.Write(b)
	buf.WriteString(`,"op2":`)
	b, _ = o.Op2.MarshalJSON()
	buf.Write(b)
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o TenderbakeDoublePreendorsementEvidence) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	b2 := bytes.NewBuffer(nil)
	o.Op1.EncodeBuffer(b2, p)
	binary.Write(buf, enc, uint32(b2.Len()))
	buf.Write(b2.Bytes())
	b2.Reset()
	o.Op2.EncodeBuffer(b2, p)
	binary.Write(buf, enc, uint32(b2.Len()))
	buf.Write(b2.Bytes())
	return nil
}

func (o *TenderbakeDoublePreendorsementEvidence) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	l, err := readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	if err = o.Op1.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
		return err
	}
	l, err = readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	if err = o.Op2.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
		return err
	}
	return nil
}

func (o TenderbakeDoublePreendorsementEvidence) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *TenderbakeDoublePreendorsementEvidence) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
