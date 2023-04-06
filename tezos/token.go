// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"fmt"

	"blockwatch.cc/tzgo/base58"
)

var (
	InvalidToken = Token{}

	// ZeroToken is a placeholder token with zero hash and zero id that may be
	// used by applications to represent tez or another default option where
	// a token address is expected.
	ZeroToken = NewToken(ZeroContract, Zero)
)

// Token represents a specialized Tezos token address that consists of
// a smart contract KT1 address and a token id represented as big integer number.
type Token struct {
	Hash [20]byte // type is always KT1
	Id   Z
}

func NewToken(contract Address, id Z) (t Token) {
	copy(t.Hash[:], contract[1:])
	t.Id = id.Clone()
	return
}

func (t Token) Contract() Address {
	return NewAddress(AddressTypeContract, t.Hash[:])
}

func (t Token) TokenId() Z {
	return t.Id
}

func (t Token) Equal(b Token) bool {
	return t.Hash == b.Hash && t.Id.Equal(b.Id)
}

func (t Token) Clone() Token {
	return Token{
		Hash: t.Hash,
		Id:   t.Id.Clone(),
	}
}

func (t Token) String() string {
	addr := base58.CheckEncode(t.Hash[:], NOCURVE_PUBLIC_KEY_HASH_ID)
	return addr + "_" + t.Id.String()
}

func (t *Token) UnmarshalText(data []byte) error {
	idx := bytes.IndexByte(data, '_')
	max := HashTypePkhNocurve.B58Len
	if idx > max {
		idx = max
	} else if idx < 0 {
		idx = len(data)
	}
	dec, ver, err := base58.CheckDecode(string(data[:idx]), 3, nil)
	if err != nil {
		if err == base58.ErrChecksum {
			return ErrChecksumMismatch
		}
		return fmt.Errorf("tezos: invalid token address: %w", err)
	}
	hashLen := HashTypePkhNocurve.Len
	if len(dec) != hashLen {
		return fmt.Errorf("tezos: invalid token address length %d", len(dec))
	}
	if !bytes.Equal(ver, NOCURVE_PUBLIC_KEY_HASH_ID) {
		return fmt.Errorf("tezos: invalid token address type %x", ver)
	}
	copy(t.Hash[:], dec)
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
	buf.Write(t.Hash[:])
	t.Id.EncodeBuffer(buf)
	return buf.Bytes()
}

func (t Token) MarshalBinary() ([]byte, error) {
	return t.Bytes(), nil
}

func (t *Token) UnmarshalBinary(data []byte) error {
	l, exp := len(data), HashTypePkhNocurve.Len
	if l < exp {
		return fmt.Errorf("tezos: short binary token address len %d", l)
	}
	copy(t.Hash[:], data)
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

// Set implements the flags.Value interface for use in command line argument parsing.
func (t *Token) Set(key string) (err error) {
	*t, err = ParseToken(key)
	return
}
