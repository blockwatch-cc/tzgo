// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
    "bytes"
    "encoding/binary"
    "strconv"
    "time"

    "blockwatch.cc/tzgo/tezos"
)

// BlockHeader represents block header
type BlockHeader struct {
    Level            int32                `json:"level"`
    Proto            byte                 `json:"proto"`
    Predecessor      tezos.BlockHash      `json:"predecessor"`
    Timestamp        time.Time            `json:"timestamp"`
    ValidationPass   byte                 `json:"validation_pass"`
    OperationsHash   tezos.OpListListHash `json:"operations_hash"`
    Fitness          []tezos.HexBytes     `json:"fitness"`
    Context          tezos.ContextHash    `json:"context"`
    Priority         uint16               `json:"priority"`
    ProofOfWorkNonce tezos.HexBytes       `json:"proof_of_work_nonce"`
    SeedNonceHash    tezos.NonceHash      `json:"seed_nonce_hash"`
    LbEscapeVote     bool                 `json:"liquidity_baking_escape_vote"`
    Signature        tezos.Signature      `json:"signature"`
}

// Bytes serializes the block header into binary form. When no signature is set, the
// result can be used as input for signing, if a signature is set the result is
// ready for broadcast.
func (h BlockHeader) Bytes() []byte {
    buf := bytes.NewBuffer(nil)
    _ = h.EncodeBuffer(buf)
    return buf.Bytes()
}

// WatermarkedBytes serializes the block header and prefixes it with a watermark.
// This format is only used for signing.
func (h BlockHeader) WatermarkedBytes() []byte {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte(BlockWatermark)
    _ = h.EncodeBuffer(buf)
    return buf.Bytes()
}

// Digest returns a 32 byte blake2b hash for signing the block header. The pre-image
// is binary serialized (without signature) and prefixed with a watermark byte.
func (h BlockHeader) Digest() []byte {
    d := tezos.Digest(h.WatermarkedBytes())
    return d[:]
}

// Sign signs the block header using a private key and generates a generic signature.
// If a valid signature already exists, this function is a noop.
func (h *BlockHeader) Sign(key tezos.PrivateKey) error {
    if h.Signature.IsValid() {
        return nil
    }
    sig, err := key.Sign(h.Digest())
    sig.Type = tezos.SignatureTypeGeneric
    if err != nil {
        return err
    }
    h.Signature = sig
    return nil
}

// WithSignature adds an externally created signature to the block header. Converts
// any non-generic signature first. No signature validation is performed, it is
// assumed the signature is correct.
func (h *BlockHeader) WithSignature(sig tezos.Signature) *BlockHeader {
    sig = sig.Clone()
    sig.Type = tezos.SignatureTypeGeneric
    h.Signature = sig
    return h
}

func (h BlockHeader) MarshalJSON() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    buf.WriteByte('{')
    buf.WriteString(`"level":`)
    buf.WriteString(strconv.FormatInt(int64(h.Level), 10))
    buf.WriteString(`,"proto":`)
    buf.WriteString(strconv.Itoa(int(h.Proto)))
    buf.WriteString(`,"predecessor":`)
    buf.WriteString(strconv.Quote(h.Predecessor.String()))
    buf.WriteString(`,"timestamp":`)
    buf.WriteString(strconv.Quote(h.Timestamp.UTC().Format("2006-01-02T15:04:05Z")))
    buf.WriteString(`,"validation_pass":`)
    buf.WriteString(strconv.Itoa(int(h.ValidationPass)))
    buf.WriteString(`,"operations_hash":`)
    buf.WriteString(strconv.Quote(h.OperationsHash.String()))
    buf.WriteString(`,"fitness":[`)
    for i, v := range h.Fitness {
        if i > 0 {
            buf.WriteByte(',')
        }
        buf.WriteString(strconv.Quote(v.String()))
    }
    buf.WriteString(`],"context":`)
    buf.WriteString(strconv.Quote(h.Context.String()))
    buf.WriteString(`,"priority":`)
    buf.WriteString(strconv.Itoa(int(h.Priority)))
    buf.WriteString(`,"proof_of_work_nonce":`)
    buf.WriteString(strconv.Quote(h.ProofOfWorkNonce.String()))
    if h.SeedNonceHash.IsValid() {
        buf.WriteString(`,"seed_nonce_hash":`)
        buf.WriteString(strconv.Quote(h.SeedNonceHash.String()))
    }
    buf.WriteString(`,"liquidity_baking_escape_vote":`)
    buf.WriteString(strconv.FormatBool(h.LbEscapeVote))
    if h.Signature.IsValid() {
        buf.WriteString(`,"signature":`)
        buf.WriteString(strconv.Quote(h.Signature.String()))
    }
    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func (h *BlockHeader) EncodeBuffer(buf *bytes.Buffer) error {
    binary.Write(buf, enc, h.Level)
    buf.WriteByte(h.Proto)
    buf.Write(h.Predecessor.Bytes())
    binary.Write(buf, enc, h.Timestamp.Unix())
    buf.WriteByte(h.ValidationPass)
    buf.Write(h.OperationsHash.Bytes())
    var fitnessLen int
    for _, v := range h.Fitness {
        fitnessLen += len(v)
    }
    binary.Write(buf, enc, uint32(fitnessLen+4*len(h.Fitness)))
    for _, v := range h.Fitness {
        binary.Write(buf, enc, uint32(len(v)))
        buf.Write(v)
    }
    buf.Write(h.Context.Bytes())
    binary.Write(buf, enc, h.Priority)
    buf.Write(h.ProofOfWorkNonce)
    if h.SeedNonceHash.IsValid() {
        buf.WriteByte(0xff)
        buf.Write(h.SeedNonceHash.Bytes())
    } else {
        buf.WriteByte(0x0)
    }
    if h.LbEscapeVote {
        buf.WriteByte(0xff)
    } else {
        buf.WriteByte(0x0)
    }
    if h.Signature.IsValid() {
        buf.Write(h.Signature.Data) // raw, no tag!
    }
    return nil
}

func (h *BlockHeader) DecodeBuffer(buf *bytes.Buffer) (err error) {
    h.Level, err = readInt32(buf.Next(4))
    if err != nil {
        return
    }
    h.Proto, err = readByte(buf.Next(1))
    if err != nil {
        return
    }
    if err = h.Predecessor.UnmarshalBinary(buf.Next(32)); err != nil {
        return
    }
    var i64 int64
    i64, err = readInt64(buf.Next(8))
    if err != nil {
        return
    }
    h.Timestamp = time.Unix(i64, 0).UTC()
    h.ValidationPass, err = readByte(buf.Next(1))
    if err != nil {
        return
    }
    if err = h.OperationsHash.UnmarshalBinary(buf.Next(32)); err != nil {
        return
    }
    var l int32
    l, err = readInt32(buf.Next(4))
    if err != nil {
        return
    }
    h.Fitness = make([]tezos.HexBytes, 0)
    for l > 0 {
        var n int32
        n, err = readInt32(buf.Next(4))
        if err != nil {
            return
        }
        b := make([]byte, int(n))
        copy(b, buf.Next(int(n)))
        h.Fitness = append(h.Fitness, b)
        l -= n + 4
    }
    if err = h.Context.UnmarshalBinary(buf.Next(32)); err != nil {
        return
    }
    h.Priority, err = readUint16(buf.Next(2))
    if err != nil {
        return
    }
    h.ProofOfWorkNonce = make([]byte, 8)
    copy(h.ProofOfWorkNonce[:], buf.Next(8))
    var ok bool
    ok, err = readBool(buf.Next(1))
    if err != nil {
        return
    }
    if ok {
        if err = h.SeedNonceHash.UnmarshalBinary(buf.Next(32)); err != nil {
            return
        }
    }
    h.LbEscapeVote, err = readBool(buf.Next(1))
    if err != nil {
        return
    }
    // conditionally read signature
    if buf.Len() > 0 {
        err = h.Signature.UnmarshalBinary(buf.Next(64))
        if err != nil {
            return
        }
    }
    return nil
}

func (h BlockHeader) MarshalBinary() ([]byte, error) {
    buf := bytes.NewBuffer(nil)
    err := h.EncodeBuffer(buf)
    return buf.Bytes(), err
}

func (h *BlockHeader) UnmarshalBinary(data []byte) error {
    return h.DecodeBuffer(bytes.NewBuffer(data))
}
