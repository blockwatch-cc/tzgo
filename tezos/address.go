// Copyright (c) 2020-2022 Blockwatch Data Inc.
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
	// ErrChecksumMismatch describes an error where decoding failed due
	// to a bad checksum.
	ErrChecksumMismatch = errors.New("tezos: checksum mismatch")

	// ErrUnknownAddressType describes an error where an address can not
	// decoded as a specific address type due to the string encoding
	// begining with an identifier byte unknown to any standard or
	// registered (via Register) network.
	ErrUnknownAddressType = errors.New("tezos: unknown address type")

	// InvalidAddress is an empty invalid address
	InvalidAddress = Address{Type: AddressTypeInvalid, Hash: nil}

	// ZeroAddress is a tz1 address with all bytes zero
	ZeroAddress = Address{Type: AddressTypeEd25519, Hash: make([]byte, HashTypePkhEd25519.Len())}

	// Burn Address
	BurnAddress = MustParseAddress("tz1burnburnburnburnburnburnburjAYjjX")
)

// AddressType represents the type of a Tezos signature.
type AddressType byte

const (
	AddressTypeInvalid AddressType = iota
	AddressTypeEd25519
	AddressTypeSecp256k1
	AddressTypeP256
	AddressTypeContract
	AddressTypeBlinded
	AddressTypeSapling
	AddressTypeBls12_381
	AddressTypeToru
	AddressTypeDekuContract
)

func ParseAddressType(s string) AddressType {
	switch s {
	case "ed25519", ED25519_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeEd25519
	case "secp256k1", SECP256K1_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeSecp256k1
	case "p256", P256_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeP256
	case "contract", NOCURVE_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeContract
	case "blinded", BLINDED_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeBlinded
	case "sapling", SAPLING_ADDRESS_PREFIX:
		return AddressTypeSapling
	case "bls12_381", BLS12_381_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeBls12_381
	case "txrollup", "toru", TORU_ADDRESS_PREFIX:
		return AddressTypeToru
	case "deku-contract", DEKU_CONTRACT_HASH_PREFIX:
		return AddressTypeDekuContract
	default:
		return AddressTypeInvalid
	}
}

func (t AddressType) IsValid() bool {
	return t != AddressTypeInvalid
}

func (t AddressType) String() string {
	switch t {
	case AddressTypeEd25519:
		return "ed25519"
	case AddressTypeSecp256k1:
		return "secp256k1"
	case AddressTypeP256:
		return "p256"
	case AddressTypeContract:
		return "contract"
	case AddressTypeBlinded:
		return "blinded"
	case AddressTypeSapling:
		return "sapling"
	case AddressTypeBls12_381:
		return "bls12_381"
	case AddressTypeToru:
		return "txrollup"
	case AddressTypeDekuContract:
		return "deku-contract"
	default:
		return "invalid"
	}
}

func (t AddressType) Prefix() string {
	switch t {
	case AddressTypeEd25519:
		return ED25519_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeSecp256k1:
		return SECP256K1_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeP256:
		return P256_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeContract:
		return NOCURVE_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeBlinded:
		return BLINDED_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeSapling:
		return SAPLING_ADDRESS_PREFIX
	case AddressTypeBls12_381:
		return BLS12_381_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeToru:
		return TORU_ADDRESS_PREFIX
	case AddressTypeDekuContract:
		return DEKU_CONTRACT_HASH_PREFIX
	default:
		return ""
	}
}

func (t AddressType) Tag() byte {
	switch t {
	case AddressTypeEd25519:
		return 0
	case AddressTypeSecp256k1:
		return 1
	case AddressTypeP256:
		return 2
	case AddressTypeBlinded:
		return 3
	case AddressTypeBls12_381:
		return 4
	case AddressTypeDekuContract:
		return 254
	default:
		return 255
	}
}

func ParseAddressTag(b byte) AddressType {
	switch b {
	case 0:
		return AddressTypeEd25519
	case 1:
		return AddressTypeSecp256k1
	case 2:
		return AddressTypeP256
	case 3:
		return AddressTypeBlinded
	case 4:
		return AddressTypeBls12_381
	case 254:
		return AddressTypeDekuContract
	default:
		return AddressTypeInvalid
	}
}

func (t *AddressType) UnmarshalText(data []byte) error {
	typ := ParseAddressType(string(data))
	if !typ.IsValid() {
		return ErrUnknownAddressType
	}
	*t = typ
	return nil
}

func (t AddressType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func HasAddressPrefix(s string) bool {
	for _, prefix := range []string{
		ED25519_PUBLIC_KEY_HASH_PREFIX,
		SECP256K1_PUBLIC_KEY_HASH_PREFIX,
		P256_PUBLIC_KEY_HASH_PREFIX,
		NOCURVE_PUBLIC_KEY_HASH_PREFIX,
		BLINDED_PUBLIC_KEY_HASH_PREFIX,
		SAPLING_ADDRESS_PREFIX,
		BLS12_381_PUBLIC_KEY_HASH_PREFIX,
		TORU_ADDRESS_PREFIX,
		DEKU_CONTRACT_HASH_PREFIX,
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func (t AddressType) HashType() HashType {
	switch t {
	case AddressTypeEd25519:
		return HashTypePkhEd25519
	case AddressTypeSecp256k1:
		return HashTypePkhSecp256k1
	case AddressTypeP256:
		return HashTypePkhP256
	case AddressTypeContract:
		return HashTypePkhNocurve
	case AddressTypeBlinded:
		return HashTypePkhBlinded
	case AddressTypeSapling:
		return HashTypeSaplingAddress
	case AddressTypeBls12_381:
		return HashTypePkhBls12_381
	case AddressTypeToru:
		return HashTypeToruAddress
	case AddressTypeDekuContract:
		return HashTypeDekuContract
	default:
		return HashTypeInvalid
	}
}

func (t AddressType) KeyType() KeyType {
	switch t {
	case AddressTypeEd25519:
		return KeyTypeEd25519
	case AddressTypeSecp256k1:
		return KeyTypeSecp256k1
	case AddressTypeP256:
		return KeyTypeP256
	case AddressTypeBls12_381:
		return KeyTypeBls12_381
	default:
		return KeyTypeInvalid
	}
}

type Address struct {
	Type AddressType
	Hash []byte
}

func NewAddress(typ AddressType, hash []byte) Address {
	a := Address{
		Type: typ,
		Hash: make([]byte, len(hash)),
	}
	copy(a.Hash, hash)
	return a
}

func (a Address) IsValid() bool {
	return a.Type != AddressTypeInvalid && len(a.Hash) == a.Type.HashType().Len()
}

func (a Address) IsEOA() bool {
	switch a.Type {
	case AddressTypeEd25519, AddressTypeSecp256k1, AddressTypeP256, AddressTypeBls12_381:
		return true
	default:
		return false
	}
}

func (a Address) IsContract() bool {
	return a.Type == AddressTypeContract
}

func (a Address) IsRollup() bool {
	return a.Type == AddressTypeToru
}

func (a Address) IsDekuContract() bool {
	return a.Type == AddressTypeDekuContract
}

func (a Address) Equal(b Address) bool {
	return a.Type == b.Type && bytes.Equal(a.Hash, b.Hash)
}

func (a Address) Clone() Address {
	x := Address{
		Type: a.Type,
		Hash: make([]byte, len(a.Hash)),
	}
	copy(x.Hash, a.Hash)
	return x
}

// String returns the string encoding of the address.
func (a Address) String() string {
	s, _ := EncodeAddress(a.Type, a.Hash)
	return s
}

func (a Address) Short() string {
	s := a.String()
	if len(s) < 12 {
		return s
	}
	return s[:8] + "..." + s[32:]
}

func (a *Address) UnmarshalText(data []byte) error {
	astr := strings.Split(string(data), "%")[0]
	addr, err := ParseAddress(astr)
	if err != nil {
		return err
	}
	*a = addr
	return nil
}

func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// Bytes returns the 21 (implicit) or 22 byte (contract) tagged and optionally padded
// binary hash value of the address.
func (a Address) Bytes() []byte {
	switch a.Type {
	case AddressTypeInvalid:
		return nil
	case AddressTypeContract:
		buf := append([]byte{01}, a.Hash...)
		buf = append(buf, byte(0)) // padding
		return buf
	case AddressTypeToru:
		buf := append([]byte{02}, a.Hash...)
		buf = append(buf, byte(0)) // padding
		return buf
	case AddressTypeDekuContract:
		buf := append([]byte{0xfe}, a.Hash...) // custom, non-Tezos
		buf = append(buf, byte(0))             // padding
		return buf
	default:
		return append([]byte{a.Type.Tag()}, a.Hash...)
	}
}

// Bytes22 returns the 22 byte tagged and padded binary encoding for contracts
// and EOAs (tz1/2/3). In contrast to Bytes which outputs the 21 byte address for EOAs
// here we add a leading 0-byte.
func (a Address) Bytes22() []byte {
	switch a.Type {
	case AddressTypeInvalid:
		return nil
	case AddressTypeContract:
		buf := append([]byte{01}, a.Hash...)
		buf = append(buf, byte(0)) // padding
		return buf
	case AddressTypeToru:
		buf := append([]byte{02}, a.Hash...)
		buf = append(buf, byte(0)) // padding
		return buf
	case AddressTypeDekuContract:
		buf := append([]byte{0xfe}, a.Hash...) // custom, non-Tezos
		buf = append(buf, byte(0))             // padding
		return buf
	default:
		return append([]byte{00, a.Type.Tag()}, a.Hash...)
	}
}

// MarshalBinary always output the 22 byte version for contracts and EOAs.
func (a Address) MarshalBinary() ([]byte, error) {
	if a.Type == AddressTypeInvalid {
		return nil, ErrUnknownAddressType
	}
	return a.Bytes22(), nil
}

// UnmarshalBinary reads a 21 byte or 22 byte address versions and is
// resilient to longer byte strings that contain extra padding or a suffix
// (e.g. an entrypoint suffix as found in smart contract data).
func (a *Address) UnmarshalBinary(b []byte) error {
	switch true {
	case len(b) >= 22 && (b[0] == 0 || b[0] == 1 || b[0] == 2):
		switch b[0] {
		case 0:
			a.Type = ParseAddressTag(b[1])
			b = b[2:22]
		case 1:
			a.Type = AddressTypeContract
			b = b[1:21]
		case 2:
			a.Type = AddressTypeToru
			b = b[1:21]
		case 0xfe:
			a.Type = AddressTypeDekuContract
			b = b[1:21]
		default:
			return fmt.Errorf("tezos: invalid binary address prefix %x", b[0])
		}
	case len(b) >= 21:
		a.Type = ParseAddressTag(b[0])
		b = b[1:21]
	default:
		return fmt.Errorf("tezos: invalid binary address length %d", len(b))
	}
	if !a.Type.IsValid() {
		return ErrUnknownAddressType
	}
	if cap(a.Hash) < 20 {
		a.Hash = make([]byte, 20)
	} else {
		a.Hash = a.Hash[:20]
	}
	copy(a.Hash, b)
	return nil
}

// IsAddressBytes checks whether a buffer likely contains a binary encoded address.
func IsAddressBytes(b []byte) bool {
	if len(b) < 21 {
		return false
	}
	switch true {
	case len(b) == 22 && (b[0] == 0 || b[0] == 1 || b[0] == 2 || b[0] == 0xfe):
		return true
	case len(b) == 21:
		return ParseAddressTag(b[0]) != AddressTypeInvalid
	default:
		return false
	}
}

// ContractAddress returns the string encoding of the address when used
// as originated contract.
func (a Address) ContractAddress() string {
	s, _ := EncodeAddress(AddressTypeContract, a.Hash)
	return s
}

// ToruAddress returns the string encoding of the address when used
// as rollup contract.
func (a Address) ToruAddress() string {
	s, _ := EncodeAddress(AddressTypeToru, a.Hash)
	return s
}

// DekuAddress returns the string encoding of the address when used
// as rollup contract.
func (a Address) DekuAddress() string {
	s, _ := EncodeAddress(AddressTypeDekuContract, a.Hash)
	return s
}

func MustParseAddress(addr string) Address {
	a, err := ParseAddress(addr)
	if err != nil {
		panic(err)
	}
	return a
}

func ParseAddress(addr string) (Address, error) {
	if len(addr) == 0 {
		return InvalidAddress, nil
	}
	a := Address{}
	sz := 3
	if strings.HasPrefix(addr, BLINDED_PUBLIC_KEY_HASH_PREFIX) ||
		strings.HasPrefix(addr, TORU_ADDRESS_PREFIX) {
		sz = 4
	}
	decoded, version, err := base58.CheckDecode(addr, sz, nil)
	if err != nil {
		if err == base58.ErrChecksum {
			return a, ErrChecksumMismatch
		}
		return a, fmt.Errorf("tezos: decoded address is of unknown format: %w", err)
	}
	if len(decoded) != 20 {
		return a, errors.New("tezos: decoded address hash is of invalid length")
	}
	switch true {
	case bytes.Equal(version, ED25519_PUBLIC_KEY_HASH_ID):
		return Address{Type: AddressTypeEd25519, Hash: decoded}, nil
	case bytes.Equal(version, SECP256K1_PUBLIC_KEY_HASH_ID):
		return Address{Type: AddressTypeSecp256k1, Hash: decoded}, nil
	case bytes.Equal(version, P256_PUBLIC_KEY_HASH_ID):
		return Address{Type: AddressTypeP256, Hash: decoded}, nil
	case bytes.Equal(version, NOCURVE_PUBLIC_KEY_HASH_ID):
		return Address{Type: AddressTypeContract, Hash: decoded}, nil
	case bytes.Equal(version, BLINDED_PUBLIC_KEY_HASH_ID):
		return Address{Type: AddressTypeBlinded, Hash: decoded}, nil
	case bytes.Equal(version, SAPLING_ADDRESS_ID):
		return Address{Type: AddressTypeSapling, Hash: decoded}, nil
	case bytes.Equal(version, BLS12_381_PUBLIC_KEY_HASH_ID):
		return Address{Type: AddressTypeBls12_381, Hash: decoded}, nil
	case bytes.Equal(version, TORU_ADDRESS_ID):
		return Address{Type: AddressTypeToru, Hash: decoded}, nil
	case bytes.Equal(version, DEKU_CONTRACT_HASH_ID):
		return Address{Type: AddressTypeDekuContract, Hash: decoded}, nil
	default:
		return a, fmt.Errorf("tezos: decoded address %s is of unknown type %x", addr, version)
	}
}

func EncodeAddress(typ AddressType, addrhash []byte) (string, error) {
	if len(addrhash) != 20 {
		return "", fmt.Errorf("tezos: invalid address hash")
	}
	switch typ {
	case AddressTypeEd25519:
		return base58.CheckEncode(addrhash, ED25519_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeSecp256k1:
		return base58.CheckEncode(addrhash, SECP256K1_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeP256:
		return base58.CheckEncode(addrhash, P256_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeContract:
		return base58.CheckEncode(addrhash, NOCURVE_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeBlinded:
		return base58.CheckEncode(addrhash, BLINDED_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeSapling:
		return base58.CheckEncode(addrhash, SAPLING_ADDRESS_ID), nil
	case AddressTypeBls12_381:
		return base58.CheckEncode(addrhash, BLS12_381_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeToru:
		return base58.CheckEncode(addrhash, TORU_ADDRESS_ID), nil
	case AddressTypeDekuContract:
		return base58.CheckEncode(addrhash, DEKU_CONTRACT_HASH_ID), nil
	default:
		return "", fmt.Errorf("tezos: unknown address type %s for hash=%x", typ, addrhash)
	}
}
