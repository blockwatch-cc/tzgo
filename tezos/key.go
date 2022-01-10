// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"

	"blockwatch.cc/tzgo/base58"

	"github.com/decred/dcrd/dcrec/secp256k1"
	"golang.org/x/crypto/blake2b"
)

var (
	// ErrUnknownKeyType describes an error where a type for a
	// public key is undefined.
	ErrUnknownKeyType = errors.New("tezos: unknown key type")

	// ErrPassphrase is returned when a required passphrase is missing
	ErrPassphrase = errors.New("tezos: passphrase required")

	InvalidKey = Key{Type: KeyTypeInvalid, Data: nil}

	// Digest is an alias for blake2b checksum algorithm
	Digest = blake2b.Sum256
)

// PassphraseFunc is a callback used to obtain a passphrase for decrypting a private key
type PassphraseFunc func() ([]byte, error)

// KeyType is a type that describes which cryptograhic curve is used by a public or
// private key
type KeyType byte

const (
	KeyTypeEd25519 KeyType = iota
	KeyTypeSecp256k1
	KeyTypeP256
	KeyTypeInvalid
)

func (t KeyType) IsValid() bool {
	return t >= 0 && t < KeyTypeInvalid
}

func (t KeyType) String() string {
	return t.PkPrefix()
}

func (t KeyType) Curve() elliptic.Curve {
	switch t {
	case KeyTypeSecp256k1:
		return secp256k1.S256()
	case KeyTypeP256:
		return elliptic.P256()
	default:
		return nil
	}
}

func (t KeyType) PkHashType() HashType {
	switch t {
	case KeyTypeEd25519:
		return HashTypePkEd25519
	case KeyTypeSecp256k1:
		return HashTypePkSecp256k1
	case KeyTypeP256:
		return HashTypePkP256
	default:
		return HashTypeInvalid
	}
}

func (t KeyType) SkHashType() HashType {
	switch t {
	case KeyTypeEd25519:
		return HashTypeSkEd25519
	case KeyTypeSecp256k1:
		return HashTypeSkSecp256k1
	case KeyTypeP256:
		return HashTypeSkP256
	default:
		return HashTypeInvalid
	}
}

func (t KeyType) AddressType() AddressType {
	switch t {
	case KeyTypeEd25519:
		return AddressTypeEd25519
	case KeyTypeSecp256k1:
		return AddressTypeSecp256k1
	case KeyTypeP256:
		return AddressTypeP256
	default:
		return AddressTypeInvalid
	}
}

func (t KeyType) PkPrefixBytes() []byte {
	switch t {
	case KeyTypeEd25519:
		return ED25519_PUBLIC_KEY_ID
	case KeyTypeSecp256k1:
		return SECP256K1_PUBLIC_KEY_ID
	case KeyTypeP256:
		return P256_PUBLIC_KEY_ID
	default:
		return nil
	}
}

func (t KeyType) PkPrefix() string {
	switch t {
	case KeyTypeEd25519:
		return ED25519_PUBLIC_KEY_PREFIX
	case KeyTypeSecp256k1:
		return SECP256K1_PUBLIC_KEY_PREFIX
	case KeyTypeP256:
		return P256_PUBLIC_KEY_PREFIX
	default:
		return ""
	}
}

func (t KeyType) SkPrefixBytes() []byte {
	switch t {
	case KeyTypeEd25519:
		return ED25519_SEED_ID
	case KeyTypeSecp256k1:
		return SECP256K1_SECRET_KEY_ID
	case KeyTypeP256:
		return P256_SECRET_KEY_ID
	default:
		return nil
	}
}

func (t KeyType) SkePrefixBytes() []byte {
	switch t {
	case KeyTypeEd25519:
		return ED25519_ENCRYPTED_SEED_ID
	case KeyTypeSecp256k1:
		return SECP256K1_ENCRYPTED_SECRET_KEY_ID
	case KeyTypeP256:
		return P256_ENCRYPTED_SECRET_KEY_ID
	default:
		return nil
	}
}

func (t KeyType) SkPrefix() string {
	switch t {
	case KeyTypeEd25519:
		return ED25519_SECRET_KEY_PREFIX
	case KeyTypeSecp256k1:
		return SECP256K1_SECRET_KEY_PREFIX
	case KeyTypeP256:
		return P256_SECRET_KEY_PREFIX
	default:
		return ""
	}
}

func (t KeyType) SkePrefix() string {
	switch t {
	case KeyTypeEd25519:
		return ED25519_ENCRYPTED_SEED_PREFIX
	case KeyTypeSecp256k1:
		return SECP256K1_ENCRYPTED_SECRET_KEY_PREFIX
	case KeyTypeP256:
		return P256_ENCRYPTED_SECRET_KEY_PREFIX
	default:
		return ""
	}
}

func (t KeyType) Tag() byte {
	switch t {
	case KeyTypeEd25519:
		return 0
	case KeyTypeSecp256k1:
		return 1
	case KeyTypeP256:
		return 2
	default:
		return 255
	}
}

func ParseKeyTag(b byte) KeyType {
	switch b {
	case 0:
		return KeyTypeEd25519
	case 1:
		return KeyTypeSecp256k1
	case 2:
		return KeyTypeP256
	default:
		return KeyTypeInvalid
	}
}

func ParseKeyType(s string) (KeyType, bool) {
	switch s {
	case ED25519_ENCRYPTED_SEED_PREFIX:
		return KeyTypeEd25519, true
	case SECP256K1_ENCRYPTED_SECRET_KEY_PREFIX:
		return KeyTypeSecp256k1, true
	case P256_ENCRYPTED_SECRET_KEY_PREFIX:
		return KeyTypeP256, true
	case ED25519_SEED_PREFIX: // same as	ED25519_SECRET_KEY_PREFIX
		return KeyTypeEd25519, false
	case SECP256K1_SECRET_KEY_PREFIX:
		return KeyTypeSecp256k1, false
	case P256_SECRET_KEY_PREFIX:
		return KeyTypeP256, false
	default:
		return KeyTypeInvalid, false
	}
}

func IsPublicKey(s string) bool {
	for _, prefix := range []string{
		ED25519_PUBLIC_KEY_PREFIX,
		SECP256K1_PUBLIC_KEY_PREFIX,
		P256_PUBLIC_KEY_PREFIX,
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func IsPrivateKey(s string) bool {
	for _, prefix := range []string{
		ED25519_SEED_PREFIX,
		ED25519_SECRET_KEY_PREFIX,
		SECP256K1_SECRET_KEY_PREFIX,
		P256_SECRET_KEY_PREFIX,
		ED25519_ENCRYPTED_SEED_PREFIX,
		SECP256K1_ENCRYPTED_SECRET_KEY_PREFIX,
		P256_ENCRYPTED_SECRET_KEY_PREFIX,
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func IsEncryptedKey(s string) bool {
	for _, prefix := range []string{
		ED25519_ENCRYPTED_SEED_PREFIX,
		SECP256K1_ENCRYPTED_SECRET_KEY_PREFIX,
		P256_ENCRYPTED_SECRET_KEY_PREFIX,
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func HasKeyPrefix(s string) bool {
	return IsPublicKey(s) || IsPrivateKey(s)
}

// Key represents a public key on the Tezos blockchain.
type Key struct {
	Type KeyType
	Data []byte
}

func NewKey(typ KeyType, data []byte) Key {
	return Key{
		Type: typ,
		Data: data,
	}
}

// Verify verifies the signature using the public key.
func (k Key) Verify(hash []byte, sig Signature) error {
	switch k.Type {
	case KeyTypeEd25519:
		pk := ed25519.PublicKey(k.Data)
		if ok := ed25519.Verify(pk, hash, sig.Data); !ok {
			return ErrSignature
		}
	case KeyTypeSecp256k1, KeyTypeP256:
		curve := k.Type.Curve()
		pk, err := ecUnmarshalCompressed(curve, k.Data)
		if err != nil {
			return err
		}
		if ok := ecVerifySignature(pk, hash, sig); !ok {
			return ErrSignature
		}
	}
	return nil
}

func (k Key) IsValid() bool {
	return k.Type.IsValid() && k.Type.PkHashType().Len() == len(k.Data)
}

func (k Key) IsEqual(k2 Key) bool {
	return k.Type == k2.Type && bytes.Compare(k.Data, k2.Data) == 0
}

func (k Key) Clone() Key {
	buf := make([]byte, len(k.Data))
	copy(buf, k.Data)
	return Key{
		Type: k.Type,
		Data: buf,
	}
}

func (k Key) Hash() []byte {
	h, _ := blake2b.New(20, nil)
	h.Write(k.Data)
	return h.Sum(nil)
}

func (k Key) Address() Address {
	return Address{
		Type: k.Type.AddressType(),
		Hash: k.Hash(),
	}
}

func (k Key) String() string {
	if !k.IsValid() {
		return ""
	}
	return base58.CheckEncode(k.Data, k.Type.PkPrefixBytes())
}

func (k Key) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *Key) UnmarshalText(data []byte) error {
	key, err := ParseKey(string(data))
	if err != nil {
		return err
	}
	*k = key
	return nil
}

func (k Key) MarshalBinary() ([]byte, error) {
	buf := k.Bytes()
	if buf == nil {
		return nil, ErrUnknownKeyType
	}
	return buf, nil
}

func (k Key) Bytes() []byte {
	if !k.Type.IsValid() {
		return nil
	}
	return append([]byte{k.Type.Tag()}, k.Data...)
}

func DecodeKey(buf []byte) (Key, error) {
	k := Key{}
	if len(buf) == 0 {
		return k, nil
	}
	if err := k.UnmarshalBinary(buf); err != nil {
		return k, err
	}
	return k, nil
}

func (k *Key) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < 33 {
		return fmt.Errorf("tezos: invalid binary key length %d", l)
	}
	if typ := ParseKeyTag(b[0]); !typ.IsValid() {
		return fmt.Errorf("tezos: invalid binary key type %x", b[0])
	} else {
		k.Type = typ
	}
	if cap(k.Data) < l-1 {
		k.Data = make([]byte, l-1)
	} else {
		k.Data = k.Data[:l-1]
	}
	copy(k.Data, b[1:])
	return nil
}

func (k *Key) EncodeBuffer(buf *bytes.Buffer) error {
	_, err := buf.Write(k.Bytes())
	return err
}

func (k *Key) DecodeBuffer(buf *bytes.Buffer) error {
	if l := buf.Len(); l < 33 {
		return fmt.Errorf("tezos: invalid binary key length %d", l)
	}
	tag := buf.Next(1)[0]
	if typ := ParseKeyTag(tag); !typ.IsValid() {
		return fmt.Errorf("tezos: invalid binary key type %x", tag)
	} else {
		k.Type = typ
	}
	l := k.Type.PkHashType().Len()
	k.Data = make([]byte, l)
	copy(k.Data, buf.Next(l))
	if !k.IsValid() {
		return fmt.Errorf("tezos: invalid binary key typ=%s len=%d", k.Type, len(k.Data))
	}
	return nil
}

func ParseKey(s string) (Key, error) {
	k := Key{}
	if len(s) == 0 {
		return k, nil
	}
	decoded, version, err := base58.CheckDecode(s, 4, nil)
	if err != nil {
		if err == base58.ErrChecksum {
			return k, ErrChecksumMismatch
		}
		return k, fmt.Errorf("tezos: unknown format for key %s: %w", s, err)
	}
	switch true {
	case bytes.Compare(version, ED25519_PUBLIC_KEY_ID) == 0:
		k.Type = KeyTypeEd25519
	case bytes.Compare(version, SECP256K1_PUBLIC_KEY_ID) == 0:
		k.Type = KeyTypeSecp256k1
	case bytes.Compare(version, P256_PUBLIC_KEY_ID) == 0:
		k.Type = KeyTypeP256
	default:
		return k, fmt.Errorf("tezos: unknown version %x for key %s", version, s)
	}
	if l := len(decoded); l != k.Type.PkHashType().Len() {
		return k, fmt.Errorf("tezos: invalid length %d for %s key data", l, k.Type.PkPrefix())
	}
	k.Data = decoded
	return k, nil
}

func MustParseKey(key string) Key {
	k, err := ParseKey(key)
	if err != nil {
		panic(err)
	}
	return k
}

// PrivateKey represents a typed private key used for signing messages.
type PrivateKey struct {
	Type KeyType
	Data []byte
}

func (k PrivateKey) IsValid() bool {
	return k.Type.IsValid() && k.Type.SkHashType().Len() == len(k.Data)
}

func (k PrivateKey) String() string {
	var buf []byte
	switch k.Type {
	case KeyTypeEd25519:
		buf = ed25519.PrivateKey(k.Data).Seed()
	case KeyTypeSecp256k1, KeyTypeP256:
		buf = k.Data
	default:
		return ""
	}
	return base58.CheckEncode(buf, k.Type.SkPrefixBytes())
}

func (k PrivateKey) Address() Address {
	return k.Public().Address()
}

func (k PrivateKey) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *PrivateKey) UnmarshalText(data []byte) error {
	key, err := ParsePrivateKey(string(data))
	if err != nil {
		return err
	}
	*k = key
	return nil
}

// GenerateKey creates a random private key.
func GenerateKey(typ KeyType) (PrivateKey, error) {
	key := PrivateKey{
		Type: typ,
	}
	switch typ {
	case KeyTypeEd25519:
		_, sk, err := ed25519.GenerateKey(nil)
		if err != nil {
			return key, err
		}
		key.Data = []byte(sk)
	case KeyTypeSecp256k1, KeyTypeP256:
		curve := typ.Curve()
		ecKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return key, err
		}
		key.Data = make([]byte, typ.SkHashType().Len())
		ecKey.D.FillBytes(key.Data)
	}
	return key, nil
}

// Public returns the public key associated with the private key.
func (k PrivateKey) Public() Key {
	pk := Key{
		Type: k.Type,
	}
	switch k.Type {
	case KeyTypeEd25519:
		pk.Data = []byte(ed25519.PrivateKey(k.Data).Public().(ed25519.PublicKey))
	case KeyTypeSecp256k1, KeyTypeP256:
		curve := k.Type.Curve()
		ecKey, err := ecPrivateKeyFromBytes(k.Data, curve)
		if err != nil {
			pk.Type = KeyTypeInvalid
			return pk
		}
		pk.Data = elliptic.MarshalCompressed(curve, ecKey.PublicKey.X, ecKey.PublicKey.Y)
	}
	return pk
}

// Encrypt encrypts the private key with a passphrase obtained from calling fn.
func (k PrivateKey) Encrypt(fn PassphraseFunc) (string, error) {
	var buf []byte
	switch k.Type {
	case KeyTypeEd25519:
		buf = []byte(ed25519.PrivateKey(k.Data).Seed())
	case KeyTypeSecp256k1, KeyTypeP256:
		buf = k.Data
	}
	enc, err := encryptPrivateKey(buf, fn)
	if err != nil {
		return "", err
	}
	return base58.CheckEncode(enc, k.Type.SkePrefixBytes()), nil
}

// Sign signs the digest (hash) of a message with the private key.
func (k PrivateKey) Sign(hash []byte) (Signature, error) {
	switch k.Type {
	case KeyTypeEd25519:
		return Signature{
			Type: SignatureTypeEd25519,
			Data: ed25519.Sign(ed25519.PrivateKey(k.Data), hash),
		}, nil
	case KeyTypeSecp256k1, KeyTypeP256:
		curve := k.Type.Curve()
		sig := Signature{
			Type: SignatureTypeSecp256k1,
		}
		if k.Type == KeyTypeP256 {
			sig.Type = SignatureTypeP256
		}
		ecKey, err := ecPrivateKeyFromBytes(k.Data, curve)
		if err != nil {
			return sig, err
		}
		sig.Data, err = ecSign(ecKey, hash)
		return sig, err
	default:
		return Signature{}, ErrUnknownKeyType
	}
}

// ParseEncryptedPrivateKey attempts to parse and optionally decrypt a
// Tezos private key. When an encrypted key is detected, fn is called
// and expected to return the decoding passphrase.
func ParseEncryptedPrivateKey(s string, fn PassphraseFunc) (k PrivateKey, err error) {
	var (
		prefixLen     int = 4
		shouldDecrypt bool
	)
	if IsEncryptedKey(s) {
		prefixLen = 5
		shouldDecrypt = true
	}

	// decode base58, version length differs between encrypted and non-encrypted keys
	decoded, version, err := base58.CheckDecode(s, prefixLen, nil)
	if err != nil {
		if err == base58.ErrChecksum {
			err = ErrChecksumMismatch
			return
		}
		err = fmt.Errorf("tezos: unknown format for private key %s: %w", s, err)
		return
	}

	// decrypt if necessary
	if shouldDecrypt {
		decoded, err = decryptPrivateKey(decoded, fn)
		if err != nil {
			return
		}
		switch true {
		case bytes.Compare(version, ED25519_ENCRYPTED_SEED_ID) == 0:
			version = ED25519_SEED_ID
		case bytes.Compare(version, SECP256K1_ENCRYPTED_SECRET_KEY_ID) == 0:
			version = SECP256K1_SECRET_KEY_ID
		case bytes.Compare(version, P256_ENCRYPTED_SECRET_KEY_ID) == 0:
			version = P256_SECRET_KEY_ID
		}
	}

	// detect type
	switch true {
	case bytes.Compare(version, ED25519_SEED_ID) == 0:
		if l := len(decoded); l != ed25519.SeedSize {
			return k, fmt.Errorf("tezos: invalid ed25519 seed length: %d", l)
		}
		k.Type = KeyTypeEd25519
		// convert seed to key
		decoded = []byte(ed25519.NewKeyFromSeed(decoded))
	case bytes.Compare(version, ED25519_SECRET_KEY_ID) == 0:
		k.Type = KeyTypeEd25519
	case bytes.Compare(version, SECP256K1_SECRET_KEY_ID) == 0:
		k.Type = KeyTypeSecp256k1
	case bytes.Compare(version, P256_SECRET_KEY_ID) == 0:
		k.Type = KeyTypeP256
	default:
		err = fmt.Errorf("tezos: unknown version %x for private key %s", version, s)
		return
	}
	if l := len(decoded); l != k.Type.SkHashType().Len() {
		return k, fmt.Errorf("tezos: invalid length %d for %s private key data", l, k.Type.SkPrefix())
	}
	k.Data = decoded
	return
}

func ParsePrivateKey(s string) (PrivateKey, error) {
	return ParseEncryptedPrivateKey(s, nil)
}
