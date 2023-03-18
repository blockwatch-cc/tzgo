// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"errors"
	"fmt"
	"io"
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
	InvalidAddress = NewAddress(AddressTypeInvalid, nil)

	// ZeroAddress is a tz1 address with all bytes zero
	ZeroAddress  = NewAddress(AddressTypeEd25519, make([]byte, HashTypePkhEd25519.Len))
	ZeroContract = NewAddress(AddressTypeContract, make([]byte, HashTypePkhNocurve.Len))

	// Burn Address
	BurnAddress = MustParseAddress("tz1burnburnburnburnburnburnburjAYjjX")
)

const MAX_ADDRESS_LEN = 37 // tx rollup address

// AddressType represents the type of a Tezos signature.
type AddressType byte

// addressTypeData is an internal type used to store address related config
// options in a single place
type addressTypeData struct {
	Id       byte
	Tag      byte
	Name     string
	HashType HashType
	KeyType  KeyType
}

const (
	AddressTypeInvalid     AddressType = iota // 0
	AddressTypeEd25519                        // 1
	AddressTypeSecp256k1                      // 2
	AddressTypeP256                           // 3
	AddressTypeContract                       // 4
	AddressTypeBlinded                        // 5
	AddressTypeBls12_381                      // 6
	AddressTypeTxRollup                       // 7
	AddressTypeSmartRollup                    // 8
)

var (
	addressTypes = []addressTypeData{
		{0, 255, "invalid", HashTypeInvalid, KeyTypeInvalid},
		{1, 0, "ed25519", HashTypePkhEd25519, KeyTypeEd25519},
		{2, 1, "secp256k1", HashTypePkhSecp256k1, KeyTypeSecp256k1},
		{3, 2, "p256", HashTypePkhP256, KeyTypeP256},
		{4, 255, "contract", HashTypePkhNocurve, KeyTypeInvalid},
		{5, 3, "blinded", HashTypePkhBlinded, KeyTypeInvalid},
		{6, 4, "bls12_381", HashTypePkhBls12_381, KeyTypeBls12_381},
		{7, 255, "tx_rollup", HashTypeTxRollupAddress, KeyTypeInvalid},
		{8, 255, "smart_rollup", HashTypeSmartRollupAddress, KeyTypeInvalid},
	}

	addressTags = []AddressType{
		AddressTypeEd25519,   // 0
		AddressTypeSecp256k1, // 1
		AddressTypeP256,      // 2
		AddressTypeBlinded,   // 3
		AddressTypeBls12_381, // 4
	}
)

func ParseAddressType(s string) AddressType {
	for _, v := range addressTypes {
		if s == v.HashType.B58Prefix || s == v.Name {
			return AddressType(v.Id)
		}
	}
	return AddressTypeInvalid
}

func (t AddressType) IsValid() bool {
	return t != AddressTypeInvalid
}

func (t AddressType) String() string {
	return addressTypes[int(t)].Name
}

func (t AddressType) Prefix() string {
	return addressTypes[int(t)].HashType.B58Prefix
}

func (t AddressType) Tag() byte {
	return addressTypes[int(t)].Tag
}

func (t AddressType) HashType() HashType {
	return addressTypes[int(t)].HashType
}

func (t AddressType) KeyType() KeyType {
	return addressTypes[int(t)].KeyType
}

func (t AddressType) asByte() byte {
	return byte(t)
}

func parseAddressTag(b byte) byte {
	t := AddressTypeInvalid
	if int(b) < len(addressTags) {
		t = addressTags[int(b)]
	}
	return t.asByte()
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
	for _, typ := range addressTypes[1:] {
		// ED25519_PUBLIC_KEY_HASH_PREFIX,   // tz1
		// SECP256K1_PUBLIC_KEY_HASH_PREFIX, // tz2
		// P256_PUBLIC_KEY_HASH_PREFIX,      // tz3
		// NOCURVE_PUBLIC_KEY_HASH_PREFIX,   // KT1
		// BLINDED_PUBLIC_KEY_HASH_PREFIX,   // btz1
		// BLS12_381_PUBLIC_KEY_HASH_PREFIX, // tz4
		// TX_ROLLUP_ADDRESS_PREFIX,         // txr1
		// SMART_ROLLUP_ADDRESS_PREFIX,      // sr1
		if strings.HasPrefix(s, typ.HashType.B58Prefix) {
			return true
		}
	}
	return false
}

func DetectAddressType(s string) AddressType {
	for _, typ := range addressTypes[1:] {
		if strings.HasPrefix(s, typ.HashType.B58Prefix) {
			return AddressType(typ.Id)
		}
	}
	return AddressTypeInvalid
}

// Address represents a typed tezos address
type Address [21]byte

func NewAddress(typ AddressType, hash []byte) (a Address) {
	isValid := typ.HashType().Len == len(hash)
	a[0] = typ.asByte() * byte(b2i(isValid))
	copy(a[1:], hash)
	return
}

func (a Address) Type() AddressType {
	return AddressType(a[0])
}

func (a Address) Hash() []byte {
	return a[1:]
}

func (a Address) KeyType() KeyType {
	return AddressType(a[0]).KeyType()
}

func (a Address) IsValid() bool {
	return a.Type() != AddressTypeInvalid
}

func (a Address) IsEOA() bool {
	return a.Type().KeyType().IsValid()
}

func (a Address) IsContract() bool {
	return a.Type() == AddressTypeContract
}

func (a Address) IsRollup() bool {
	return a.Type() == AddressTypeSmartRollup || a.Type() == AddressTypeTxRollup
}

func (a Address) Equal(b Address) bool {
	return a == b
}

func (a Address) Clone() (b Address) {
	copy(b[:], a[:])
	return
}

// String returns the string encoding of the address.
func (a Address) String() string {
	return EncodeAddress(a.Type(), a[1:])
}

func (a *Address) UnmarshalText(data []byte) error {
	if len(data) > MAX_ADDRESS_LEN {
		data = data[:MAX_ADDRESS_LEN]
	}
	astr, _, _ := strings.Cut(string(data), "%")
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
// func (a Address) Bytes() []byte {
func (a Address) Encode() []byte {
	var buf [22]byte
	switch a.Type() {
	case AddressTypeInvalid:
		return nil
	case AddressTypeContract:
		buf[0] = 1
		copy(buf[1:], a[1:])
	case AddressTypeTxRollup:
		buf[0] = 2
		copy(buf[1:], a[1:])
	case AddressTypeSmartRollup:
		buf[0] = 3
		copy(buf[1:], a[1:])
	default:
		// 21 byte version for implicit addresses
		buf[0] = a.Type().Tag()
		copy(buf[1:], a[1:])
		return buf[:21]
	}
	return buf[:]
}

// Bytes22 returns the 22 byte tagged and padded binary encoding for contracts
// and EOAs (tz1/2/3). In contrast to Bytes which outputs the 21 byte address for EOAs
// here we add a leading 0-byte.
// func (a Address) Bytes22() []byte {
func (a Address) EncodePadded() []byte {
	var buf [22]byte
	switch a.Type() {
	case AddressTypeInvalid:
		return nil
	case AddressTypeContract:
		buf[0] = 1
		copy(buf[1:], a[1:])
	case AddressTypeTxRollup:
		buf[0] = 2
		copy(buf[1:], a[1:])
	case AddressTypeSmartRollup:
		buf[0] = 3
		copy(buf[1:], a[1:])
	default:
		buf[1] = a.Type().Tag()
		copy(buf[2:], a[1:])
	}
	return buf[:]
}

// MarshalBinary outputs the 21 byte TzGo version of an address containing
// a one byte type tag and the 20 byte address hash.
func (a Address) MarshalBinary() ([]byte, error) {
	return a[:], nil
}

// UnmarshalBinary reads the 21 byte TzGo version of an address containing
// a one byte type tag and the 20 byte address hash.
func (a *Address) UnmarshalBinary(b []byte) error {
	if len(b) != 21 {
		return io.ErrShortBuffer
	}
	copy(a[:], b)
	return nil
}

// Decode reads a 21 byte or 22 byte address versions and is
// resilient to longer byte strings that contain extra padding or a suffix
// (e.g. an entrypoint suffix as found in smart contract data).
func (a *Address) Decode(b []byte) error {
	a[0] = 0
	switch {
	case len(b) >= 22 && b[0] <= 3:
		switch b[0] {
		case 0:
			a[0] = parseAddressTag(b[1])
			copy(a[1:], b[2:22])
		case 1:
			a[0] = AddressTypeContract.asByte()
			copy(a[1:], b[1:21])
		case 2:
			a[0] = AddressTypeTxRollup.asByte()
			copy(a[1:], b[1:21])
		case 3:
			a[0] = AddressTypeSmartRollup.asByte()
			copy(a[1:], b[1:21])
		default:
			return fmt.Errorf("tezos: invalid binary address prefix %x", b[0])
		}
	case len(b) >= 21:
		a[0] = parseAddressTag(b[0])
		copy(a[1:], b[1:21])
	default:
		return fmt.Errorf("tezos: invalid binary address length %d", len(b))
	}
	if !a.IsValid() {
		return ErrUnknownAddressType
	}
	return nil
}

// IsAddressBytes checks whether a buffer likely contains a binary encoded address.
func IsAddressBytes(b []byte) bool {
	if len(b) < 21 {
		return false
	}
	switch {
	case len(b) == 22 && (b[0] <= 3):
		return true
	case len(b) == 21:
		return parseAddressTag(b[0]) != AddressTypeInvalid.asByte()
	default:
		return false
	}
}

// ContractAddress returns the string encoding of the address when used
// as originated contract.
func (a Address) ContractAddress() string {
	return EncodeAddress(AddressTypeContract, a[1:])
}

// TxRollupAddress returns the string encoding of the address when used
// as rollup contract.
func (a Address) TxRollupAddress() string {
	return EncodeAddress(AddressTypeTxRollup, a[1:])
}

// SmartRollupAddress returns the string encoding of the address when used
// as smart rollup contract.
func (a Address) SmartRollupAddress() string {
	return EncodeAddress(AddressTypeSmartRollup, a[1:])
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (a *Address) Set(addr string) (err error) {
	*a, err = ParseAddress(addr)
	return
}

func MustParseAddress(addr string) Address {
	a, err := ParseAddress(addr)
	if err != nil {
		panic(err)
	}
	return a
}

func ParseAddress(addr string) (a Address, err error) {
	// accept empty strings, but return an invalid address
	if len(addr) == 0 {
		return
	}
	if len(addr) > MAX_ADDRESS_LEN {
		err = fmt.Errorf("tezos: invalid base58 address length")
		return
	}
	typ := DetectAddressType(addr)
	if !typ.IsValid() {
		err = fmt.Errorf("tezos: unknown address type for %q", addr)
		return
	}
	ht := typ.HashType()
	ibuf := bufPool32.Get()
	dec, _, err2 := base58.CheckDecode(addr, len(ht.Id), ibuf.([]byte))
	if err2 != nil {
		bufPool32.Put(ibuf)
		if err2 == base58.ErrChecksum {
			err = ErrChecksumMismatch
			return
		}
		err = fmt.Errorf("tezos: invalid %s address: %w", typ, err2)
		return
	}
	a[0] = typ.asByte()
	copy(a[1:], dec)
	bufPool32.Put(ibuf)
	return
}

func EncodeAddress(typ AddressType, hash []byte) string {
	if typ == AddressTypeInvalid {
		return ""
	}
	return base58.CheckEncode(hash, typ.HashType().Id)
}
