// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "encoding/binary"
    "strconv"

    "blockwatch.cc/tzgo/tezos"
)

// Ballot represents "ballot" operation
type Ballot struct {
    Source   tezos.Address      `json:"source"`
    Period   int32              `json:"period"`
    Proposal tezos.ProtocolHash `json:"proposal"`
    Ballot   tezos.BallotVote   `json:"ballot"`
}

func (o Ballot) Kind() tezos.OpType {
    return tezos.OpTypeBallot
}

func (o Ballot) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"kind":`)
    buf.WriteString(strconv.Quote(o.Kind().String()))
    buf.WriteString(`,"source":`)
    buf.WriteString(strconv.Quote(o.Source.String()))
    buf.WriteString(`,"period":`)
    buf.WriteString(strconv.Itoa(int(o.Period)))
    buf.WriteString(`,"proposal":`)
    buf.WriteString(strconv.Quote(o.Proposal.String()))
    buf.WriteString(`,"ballot":`)
    buf.WriteString(strconv.Quote(o.Ballot.String()))
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (o Ballot) EncodeBuffer(buf *bytes.Buffer, p *tezos.Params) error {
    buf.WriteByte(o.Kind().TagVersion(p.OperationTagsVersion))
    buf.Write(o.Source.Bytes())
    binary.Write(buf, enc, o.Period)
    buf.Write(o.Proposal.Bytes())
    buf.WriteByte(o.Ballot.Tag())
    return nil
}

func (o *Ballot) DecodeBuffer(buf *bytes.Buffer, p *tezos.Params) (err error) {
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
    if err = o.Proposal.UnmarshalBinary(buf.Next(32)); err != nil {
        return
    }
    if err = o.Ballot.UnmarshalBinary(buf.Next(1)); err != nil {
        return
    }
    return nil
}

func (o Ballot) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := o.EncodeBuffer(buf, tezos.DefaultParams)
    return buf.Bytes(), err
}

func (o *Ballot) UnmarshalBinary(data []byte) error {
    return o.DecodeBuffer(bytes.NewBuffer(data), tezos.DefaultParams)
}
