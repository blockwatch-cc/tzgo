// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"blockwatch.cc/tzgo/base58"
)

var (
	// ErrUnknownSignatureType describes an error where a type for a
	// signature is undefined.
	ErrUnknownSignatureType = errors.New("unknown signature type")

	// ErrSignature is returned when signature verification fails
	ErrSignature = errors.New("signature mismatch")

	// InvalidSignature represents an empty invalid signature
	InvalidSignature = Signature{Type: SignatureTypeInvalid, Data: nil}

	// ZeroSignature represents a valid signature derived from null bytes
	ZeroSignature = MustParseSignature("sigMzJ4GVAvXEd2RjsKGfG2H9QvqTSKCZsuB2KiHbZRGFz72XgF6KaKADznh674fQgBatxw3xdHqTtMHUZAGRprxy64wg1aq")
)

// SignatureType represents the type of a Tezos signature.
type SignatureType byte

const (
	SignatureTypeEd25519 SignatureType = iota
	SignatureTypeSecp256k1
	SignatureTypeP256
	SignatureTypeBls12_381
	SignatureTypeGeneric
	SignatureTypeGenericAggregate
	SignatureTypeInvalid
)

func (t SignatureType) IsValid() bool {
	return t < SignatureTypeInvalid
}

func (t SignatureType) HashType() HashType {
	switch t {
	case SignatureTypeEd25519:
		return HashTypeSigEd25519
	case SignatureTypeSecp256k1:
		return HashTypeSigSecp256k1
	case SignatureTypeP256:
		return HashTypeSigP256
	case SignatureTypeBls12_381:
		return HashTypeSigBls12_381
	case SignatureTypeGeneric:
		return HashTypeSigGeneric
	case SignatureTypeGenericAggregate:
		return HashTypeSigGenericAggregate
	default:
		return HashTypeInvalid
	}
}

func (t SignatureType) PrefixBytes() []byte {
	switch t {
	case SignatureTypeEd25519:
		return ED25519_SIGNATURE_ID
	case SignatureTypeSecp256k1:
		return SECP256K1_SIGNATURE_ID
	case SignatureTypeP256:
		return P256_SIGNATURE_ID
	case SignatureTypeBls12_381:
		return BLS12_381_SIGNATURE_ID
	case SignatureTypeGeneric:
		return GENERIC_SIGNATURE_ID
	case SignatureTypeGenericAggregate:
		return GENERIC_AGGREGATE_SIGNATURE_ID
	default:
		return nil
	}
}

func (t SignatureType) Prefix() string {
	switch t {
	case SignatureTypeEd25519:
		return ED25519_SIGNATURE_PREFIX
	case SignatureTypeSecp256k1:
		return SECP256K1_SIGNATURE_PREFIX
	case SignatureTypeP256:
		return P256_SIGNATURE_PREFIX
	case SignatureTypeBls12_381:
		return BLS12_381_SIGNATURE_PREFIX
	case SignatureTypeGeneric:
		return GENERIC_SIGNATURE_PREFIX
	case SignatureTypeGenericAggregate:
		return GENERIC_AGGREGATE_SIGNATURE_PREFIX
	default:
		return ""
	}
}

func (t SignatureType) String() string {
	return t.Prefix()
}

func (t SignatureType) Tag() byte {
	switch t {
	case SignatureTypeEd25519:
		return 0
	case SignatureTypeSecp256k1:
		return 1
	case SignatureTypeP256:
		return 2
	case SignatureTypeBls12_381:
		return 3
	default:
		return 255
	}
}

func ParseSignatureTag(b byte) SignatureType {
	switch b {
	case 0:
		return SignatureTypeEd25519
	case 1:
		return SignatureTypeSecp256k1
	case 2:
		return SignatureTypeP256
	case 3:
		return SignatureTypeBls12_381
	default:
		return SignatureTypeGeneric
	}
}

func HasSignaturePrefix(s string) bool {
	for _, prefix := range []string{
		ED25519_SIGNATURE_PREFIX,
		SECP256K1_SIGNATURE_PREFIX,
		P256_SIGNATURE_PREFIX,
		BLS12_381_SIGNATURE_PREFIX,
		GENERIC_SIGNATURE_PREFIX,
		GENERIC_AGGREGATE_SIGNATURE_PREFIX,
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func (t SignatureType) Len() int {
	switch t {
	case SignatureTypeBls12_381, SignatureTypeGenericAggregate:
		return 96
	case SignatureTypeInvalid:
		return 0
	default:
		return 64
	}
}

func IsSignature(s string) bool {
	for _, prefix := range []string{
		ED25519_SIGNATURE_PREFIX,
		SECP256K1_SIGNATURE_PREFIX,
		P256_SIGNATURE_PREFIX,
		BLS12_381_SIGNATURE_PREFIX,
		GENERIC_SIGNATURE_PREFIX,
		GENERIC_AGGREGATE_SIGNATURE_PREFIX,
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// Signature represents a typed Tezos signature.
type Signature struct {
	Type SignatureType
	Data []byte
}

func NewSignature(typ SignatureType, data []byte) Signature {
	return Signature{
		Type: typ,
		Data: data,
	}
}

func (s Signature) IsValid() bool {
	return s.Type.IsValid() && s.Type.Len() == len(s.Data)
}

func (s Signature) Equal(s2 Signature) bool {
	return s.Type == s2.Type && bytes.Equal(s.Data, s2.Data)
}

func (s Signature) Clone() Signature {
	buf := make([]byte, len(s.Data))
	copy(buf, s.Data)
	return Signature{
		Type: s.Type,
		Data: buf,
	}
}

// Signature converts a typed Tezos signature into a generic signature string.
func (s Signature) Generic() string {
	if !s.IsValid() {
		return ""
	}
	return base58.CheckEncode(s.Data, GENERIC_SIGNATURE_ID)
}

func (s Signature) String() string {
	if !s.IsValid() {
		return ""
	}
	return base58.CheckEncode(s.Data, s.Type.PrefixBytes())
}

func (s Signature) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *Signature) UnmarshalText(data []byte) error {
	sig, err := ParseSignature(string(data))
	if err != nil {
		return err
	}
	*s = sig
	return nil
}

func (s Signature) Bytes() []byte {
	tag := s.Type.Tag()
	if tag == 255 {
		return s.Data
	}
	return append([]byte{tag}, s.Data...)
}

func (s *Signature) DecodeBuffer(buf *bytes.Buffer) error {
	l := buf.Len()
	if l < 64 {
		return fmt.Errorf("tezos: invalid binary signature length %d", l)
	}
	// default to generic without tag
	s.Type = SignatureTypeGeneric
	// tagged signatures are either 65 byte or 97 byte (BLS)
	if l == 65 || l == 97 {
		tag := buf.Next(1)[0]
		if typ := ParseSignatureTag(tag); !typ.IsValid() {
			return fmt.Errorf("tezos: invalid binary signature type %x", tag)
		} else {
			s.Type = typ
		}
	}
	l = s.Type.Len()
	s.Data = make([]byte, l)
	copy(s.Data, buf.Next(l))
	if !s.IsValid() {
		return fmt.Errorf("tezos: invalid %s signature length %d", s.Type, l)
	}
	return nil
}

func (s Signature) MarshalBinary() ([]byte, error) {
	if !s.Type.IsValid() {
		return nil, ErrUnknownSignatureType
	}
	return s.Bytes(), nil
}

func (s *Signature) UnmarshalBinary(b []byte) error {
	switch len(b) {
	case 64:
		s.Type = SignatureTypeGeneric
	case 96:
		s.Type = SignatureTypeGenericAggregate
	case 65, 97:
		if typ := ParseSignatureTag(b[0]); !typ.IsValid() {
			return fmt.Errorf("tezos: invalid binary signature type %x", b[0])
		} else {
			s.Type = typ
		}
		b = b[1:]
	default:
		return fmt.Errorf("tezos: invalid binary signature length %d", len(b))
	}
	if cap(s.Data) < s.Type.Len() {
		s.Data = make([]byte, s.Type.Len())
	} else {
		s.Data = s.Data[:s.Type.Len()]
	}
	copy(s.Data, b)
	return nil
}

func ParseSignature(s string) (sig Signature, err error) {
	var (
		dec, ver []byte
		typ      SignatureType
	)
	switch {
	case strings.HasPrefix(s, ED25519_SIGNATURE_PREFIX):
		dec, ver, err = base58.CheckDecode(s, 5, nil)
		typ = SignatureTypeEd25519

	case strings.HasPrefix(s, SECP256K1_SIGNATURE_PREFIX):
		dec, ver, err = base58.CheckDecode(s, 5, nil)
		typ = SignatureTypeSecp256k1

	case strings.HasPrefix(s, P256_SIGNATURE_PREFIX):
		dec, ver, err = base58.CheckDecode(s, 4, nil)
		typ = SignatureTypeP256

	case strings.HasPrefix(s, BLS12_381_SIGNATURE_PREFIX):
		dec, ver, err = base58.CheckDecode(s, 4, nil)
		typ = SignatureTypeBls12_381

	case strings.HasPrefix(s, GENERIC_SIGNATURE_PREFIX):
		dec, ver, err = base58.CheckDecode(s, 3, nil)
		typ = SignatureTypeGeneric

	case strings.HasPrefix(s, GENERIC_AGGREGATE_SIGNATURE_PREFIX):
		dec, ver, err = base58.CheckDecode(s, 4, nil)
		typ = SignatureTypeGenericAggregate

	default:
		err = fmt.Errorf("tezos: unknown signature prefix %s", s)
		return
	}

	if err != nil {
		if err == base58.ErrChecksum {
			err = ErrChecksumMismatch
			return
		}
		err = fmt.Errorf("tezos: unknown signature format: %w", err)
		return
	}

	if !bytes.Equal(ver, typ.PrefixBytes()) {
		err = fmt.Errorf("tezos: invalid signature type %x for %s", ver, typ.Prefix())
		return
	}

	if l := len(dec); l < typ.Len() {
		err = fmt.Errorf("tezos: invalid length %d for %s signature data", l, typ.Prefix())
		return
	}

	sig.Type = typ
	sig.Data = dec[:typ.Len()]
	return
}

func MustParseSignature(s string) Signature {
	sig, err := ParseSignature(s)
	if err != nil {
		panic(err)
	}
	return sig
}

// Set implements the flags.Value interface for use in command line
// argument parsing.
func (s *Signature) Set(sig string) (err error) {
	*s, err = ParseSignature(sig)
	return
}
