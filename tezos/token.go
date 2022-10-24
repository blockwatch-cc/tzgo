// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
    "bytes"
    "fmt"

    "blockwatch.cc/tzgo/base58"
)

var (
    InvalidToken = Token{Hash: nil}

    // ZeroToken is a placeholder token with zero hash and zero id that may be
    // used by applications to represent tez or another default option where
    // a token address is expected.
    ZeroToken = NewToken(ZeroAddress, Zero)
)

// Token represents a specialzied Tezos token address that consists of
// a smart contract KT1 address and a token id represented as big integer number.
type Token struct {
    Hash []byte // type is always KT1
    Id   Z
}

func NewToken(contract Address, id Z) Token {
    t := Token{
        Hash: make([]byte, len(contract.Hash)),
    }
    copy(t.Hash, contract.Hash)
    t.Id = id.Clone()
    return t
}

func (t Token) IsValid() bool {
    return len(t.Hash) == HashTypePkhNocurve.Len()
}

func (t Token) Contract() Address {
    return Address{Type: AddressTypeContract, Hash: t.Hash}
}

func (t Token) TokenId() Z {
    return t.Id
}

func (t Token) Equal(b Token) bool {
    return bytes.Equal(t.Hash, b.Hash) && t.Id.Equal(b.Id)
}

func (t Token) Clone() Token {
    x := Token{
        Hash: make([]byte, len(t.Hash)),
        Id:   t.Id.Clone(),
    }
    copy(x.Hash, t.Hash)
    return x
}

func (t Token) String() string {
    addr := base58.CheckEncode(t.Hash, NOCURVE_PUBLIC_KEY_HASH_ID)
    return addr + "_" + t.Id.String()
}

func (t *Token) UnmarshalText(data []byte) error {
    idx := bytes.IndexByte(data, '_')
    max := HashTypePkhNocurve.Base58Len()
    if idx > max {
        idx = max
    } else if idx < 0 {
        idx = len(data)
    }
    decoded, version, err := base58.CheckDecode(string(data[:idx]), 3, nil)
    if err != nil {
        if err == base58.ErrChecksum {
            return ErrChecksumMismatch
        }
        return fmt.Errorf("tezos: invalid token address: %w", err)
    }
    hashLen := HashTypePkhNocurve.Len()
    if len(decoded) != hashLen {
        return fmt.Errorf("tezos: invalid token address length %d", len(decoded))
    }
    if !bytes.Equal(version, NOCURVE_PUBLIC_KEY_HASH_ID) {
        return fmt.Errorf("tezos: invalid token address type %x", version)
    }
    if cap(t.Hash) != hashLen {
        t.Hash = make([]byte, 0, hashLen)
    }
    t.Hash = t.Hash[:hashLen]
    copy(t.Hash, decoded)
    if idx < len(data) {
        // token id is optional
        if err := t.Id.UnmarshalText(data[idx+1:]); err != nil {
            t.Id.SetInt64(0)
            return fmt.Errorf("tezos: invalid token id: %w", err)
        }
    } else {
        t.Id.SetInt64(0)
    }
    return nil
}

func (t Token) MarshalText() ([]byte, error) {
    return []byte(t.String()), nil
}

// Bytes returns the 20 byte (contract) hash appended with a zarith encoded token id.
func (t Token) Bytes() []byte {
    buf := bytes.NewBuffer(nil)
    buf.Write(t.Hash)
    t.Id.EncodeBuffer(buf)
    return buf.Bytes()
}

func (t Token) MarshalBinary() ([]byte, error) {
    return t.Bytes(), nil
}

func (t *Token) UnmarshalBinary(data []byte) error {
    l, exp := len(data), HashTypePkhNocurve.Len()
    if l < exp {
        return fmt.Errorf("tezos: short binary token address len %d", l)
    }
    if cap(t.Hash) < exp {
        t.Hash = make([]byte, 0, exp)
    }
    t.Hash = t.Hash[:exp]
    copy(t.Hash, data)
    if l > exp {
        if err := t.Id.UnmarshalBinary(data[exp:]); err != nil {
            return fmt.Errorf("tezos: invalid binary token id: %w", err)
        }
    }
    return nil
}

func MustParseToken(addr string) Token {
    t, err := ParseToken(addr)
    if err != nil {
        panic(err)
    }
    return t
}

func ParseToken(addr string) (Token, error) {
    t := Token{}
    if len(addr) == 0 {
        return InvalidToken, nil
    }
    err := t.UnmarshalText([]byte(addr))
    return t, err
}
