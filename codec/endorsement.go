// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding/binary"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// Endorsement represents "endorsement" operation
type Endorsement struct {
	Simple
	Level int32 `json:"level"`
}

func (o Endorsement) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"level":`)
	buf.WriteString(strconv.Itoa(int(o.Level)))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o Endorsement) Kind() tezos.OpType {
	return tezos.OpTypeEndorsement
}

func (o Endorsement) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	binary.Write(buf, enc, o.Level)
	return nil
}

func (o *Endorsement) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	o.Level, err = readInt32(buf.Next(4))
	if err != nil {
		return
	}
	return nil
}

func (o Endorsement) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *Endorsement) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}

// InlinedEndorsement represents inlined endorsement operation with signature. This
// type is uses as part of other operations, but is not a stand-alone operation.
type InlinedEndorsement struct {
	Branch      tezos.BlockHash `json:"branch"`
	Endorsement Endorsement     `json:"operations"`
	Signature   tezos.Signature `json:"signature"`
}

func (o InlinedEndorsement) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"branch":`)
	buf.WriteString(strconv.Quote(o.Branch.String()))
	buf.WriteString(`,"operations":`)
	if b, err := o.Endorsement.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buf.Write(b)
	}
	buf.WriteString(`,"signature":`)
	buf.WriteString(strconv.Quote(o.Signature.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o InlinedEndorsement) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.Write(o.Branch.Bytes())
	o.Endorsement.EncodeBuffer(buf, p)
	buf.Write(o.Signature.Data) // generic sig, no tag (!)
	return nil
}

func (o *InlinedEndorsement) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	err = o.Branch.UnmarshalBinary(buf.Next(tezos.HashTypeBlock.Len))
	if err != nil {
		return
	}
	if err = o.Endorsement.DecodeBuffer(buf, p); err != nil {
		return
	}
	if err := o.Signature.DecodeBuffer(buf); err != nil {
		return err
	}
	return nil
}

// EndorsementWithSlot represents "endorsement_with_slot" operation
type EndorsementWithSlot struct {
	Simple
	Endorsement InlinedEndorsement `json:"endorsement"`
	Slot        int16              `json:"slot"`
}

func (o EndorsementWithSlot) Kind() tezos.OpType {
	return tezos.OpTypeEndorsementWithSlot
}

func (o EndorsementWithSlot) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"endorsement":`)
	b, _ := o.Endorsement.MarshalJSON()
	buf.Write(b)
	buf.WriteString(`,"slot":`)
	buf.WriteString(strconv.Itoa(int(o.Slot)))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o EndorsementWithSlot) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	b2 := bytes.NewBuffer(nil)
	o.Endorsement.EncodeBuffer(b2, p)
	binary.Write(buf, enc, uint32(b2.Len()))
	buf.Write(b2.Bytes())
	binary.Write(buf, enc, o.Slot)
	return nil
}

func (o *EndorsementWithSlot) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	l, err := readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	if err = o.Endorsement.DecodeBuffer(bytes.NewBuffer(buf.Next(int(l))), p); err != nil {
		return err
	}
	o.Slot, err = readInt16(buf.Next(2))
	return err
}

func (o EndorsementWithSlot) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *EndorsementWithSlot) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}

// TenderbakeEndorsement represents tenderbake endorsement operation
type TenderbakeEndorsement struct {
	Simple
	Slot             int16             `json:"slot"`
	Level            int32             `json:"level"`
	Round            int32             `json:"round"`
	BlockPayloadHash tezos.PayloadHash `json:"block_payload_hash"`
}

func (o TenderbakeEndorsement) Kind() tezos.OpType {
	return tezos.OpTypeEndorsement
}

func (o TenderbakeEndorsement) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"slot":`)
	buf.WriteString(strconv.Itoa(int(o.Slot)))
	buf.WriteString(`,"level":`)
	buf.WriteString(strconv.Itoa(int(o.Level)))
	buf.WriteString(`,"round":`)
	buf.WriteString(strconv.Itoa(int(o.Round)))
	buf.WriteString(`,"block_payload_hash":`)
	buf.WriteString(strconv.Quote(o.BlockPayloadHash.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o TenderbakeEndorsement) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	binary.Write(buf, enc, o.Slot)
	binary.Write(buf, enc, o.Level)
	binary.Write(buf, enc, o.Round)
	buf.Write(o.BlockPayloadHash.Bytes())
	return nil
}

func (o *TenderbakeEndorsement) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	o.Slot, err = readInt16(buf.Next(2))
	if err != nil {
		return err
	}
	o.Level, err = readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	o.Round, err = readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	err = o.BlockPayloadHash.UnmarshalBinary(buf.Next(32))
	return err
}

func (o TenderbakeEndorsement) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *TenderbakeEndorsement) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}

// TenderbakeInlinedEndorsement represents inlined endorsement operation with signature. This
// type is uses as part of other operations, but is not a stand-alone operation.
type TenderbakeInlinedEndorsement struct {
	Branch      tezos.BlockHash       `json:"branch"`
	Endorsement TenderbakeEndorsement `json:"operations"`
	Signature   tezos.Signature       `json:"signature"`
}

func (o TenderbakeInlinedEndorsement) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"branch":`)
	buf.WriteString(strconv.Quote(o.Branch.String()))
	buf.WriteString(`,"operations":`)
	b, _ := o.Endorsement.MarshalJSON()
	buf.Write(b)
	buf.WriteString(`,"signature":`)
	buf.WriteString(strconv.Quote(o.Signature.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o TenderbakeInlinedEndorsement) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.Write(o.Branch.Bytes())
	o.Endorsement.EncodeBuffer(buf, p)
	buf.Write(o.Signature.Data) // generic sig, no tag (!)
	return nil
}

func (o *TenderbakeInlinedEndorsement) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	err = o.Branch.UnmarshalBinary(buf.Next(tezos.HashTypeBlock.Len))
	if err != nil {
		return
	}
	if err = o.Endorsement.DecodeBuffer(buf, p); err != nil {
		return
	}
	if err := o.Signature.DecodeBuffer(buf); err != nil {
		return err
	}
	return nil
}

// TenderbakePreendorsement represents tenderbake preendorsement operation
type TenderbakePreendorsement struct {
	Simple
	Slot             int16             `json:"slot"`
	Level            int32             `json:"level"`
	Round            int32             `json:"round"`
	BlockPayloadHash tezos.PayloadHash `json:"block_payload_hash"`
}

func (o TenderbakePreendorsement) Kind() tezos.OpType {
	return tezos.OpTypePreendorsement
}

func (o TenderbakePreendorsement) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"kind":`)
	buf.WriteString(strconv.Quote(o.Kind().String()))
	buf.WriteString(`,"slot":`)
	buf.WriteString(strconv.Itoa(int(o.Slot)))
	buf.WriteString(`,"level":`)
	buf.WriteString(strconv.Itoa(int(o.Level)))
	buf.WriteString(`,"round":`)
	buf.WriteString(strconv.Itoa(int(o.Round)))
	buf.WriteString(`,"block_payload_hash":`)
	buf.WriteString(strconv.Quote(o.BlockPayloadHash.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o TenderbakePreendorsement) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
	binary.Write(buf, enc, o.Slot)
	binary.Write(buf, enc, o.Level)
	binary.Write(buf, enc, o.Round)
	buf.Write(o.BlockPayloadHash.Bytes())
	return nil
}

func (o *TenderbakePreendorsement) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	if err = ensureTagAndSize(buf, o.Kind(), p.OperationTagsVersion); err != nil {
		return
	}
	o.Slot, err = readInt16(buf.Next(2))
	if err != nil {
		return err
	}
	o.Level, err = readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	o.Round, err = readInt32(buf.Next(4))
	if err != nil {
		return err
	}
	err = o.BlockPayloadHash.UnmarshalBinary(buf.Next(32))
	return err
}

func (o TenderbakePreendorsement) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := o.EncodeBuffer(buf, tezos.DefaultParams)
	return buf.Bytes(), err
}

func (o *TenderbakePreendorsement) UnmarshalBinary(data []byte) error {
	return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}

// TenderbakeInlinedPreendorsement represents inlined preendorsement operation with signature. This
// type is uses as part of other operations, but is not a stand-alone operation.
type TenderbakeInlinedPreendorsement struct {
	Branch      tezos.BlockHash          `json:"branch"`
	Endorsement TenderbakePreendorsement `json:"operations"`
	Signature   tezos.Signature          `json:"signature"`
}

func (o TenderbakeInlinedPreendorsement) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte('{')
	buf.WriteString(`"branch":`)
	buf.WriteString(strconv.Quote(o.Branch.String()))
	buf.WriteString(`,"operations":`)
	b, _ := o.Endorsement.MarshalJSON()
	buf.Write(b)
	buf.WriteString(`,"signature":`)
	buf.WriteString(strconv.Quote(o.Signature.String()))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o TenderbakeInlinedPreendorsement) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
	buf.Write(o.Branch.Bytes())
	o.Endorsement.EncodeBuffer(buf, p)
	buf.Write(o.Signature.Data) // generic sig, no tag (!)
	return nil
}

func (o *TenderbakeInlinedPreendorsement) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
	err = o.Branch.UnmarshalBinary(buf.Next(tezos.HashTypeBlock.Len))
	if err != nil {
		return
	}
	if err = o.Endorsement.DecodeBuffer(buf, p); err != nil {
		return
	}
	if err := o.Signature.DecodeBuffer(buf); err != nil {
		return err
	}
	return nil
}
