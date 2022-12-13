// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"blockwatch.cc/tzgo/base58"
)

var (
	// ErrUnknownHashType describes an error where a hash can not
	// decoded as a specific hash type because the string encoding
	// starts with an unknown identifier.
	ErrUnknownHashType = errors.New("tezos: unknown hash type")

	// InvalidHash represents an empty invalid hash type
	InvalidHash = Hash{Type: HashTypeInvalid, Hash: nil}

	// ZeroHash
	ZeroOpHash    = NewOpHash(make([]byte, HashTypeOperation.Len()))
	ZeroBlockHash = NewBlockHash(make([]byte, HashTypeBlock.Len()))
	EmptyExprHash = MustParseExprHash("expru5X1yxJG6ezR2uHMotwMLNmSzQyh5t1vUnhjx4cS6Pv9qE1Sdo")
)

type HashType byte

const (
	HashTypeInvalid HashType = iota
	HashTypeChainId
	HashTypeId
	HashTypePkhEd25519
	HashTypePkhSecp256k1
	HashTypePkhP256
	HashTypePkhNocurve
	HashTypePkhBlinded
	HashTypeBlock
	HashTypeOperation
	HashTypeOperationList
	HashTypeOperationListList
	HashTypeProtocol
	HashTypeContext
	HashTypeNonce
	HashTypeSeedEd25519
	HashTypePkEd25519
	HashTypeSkEd25519
	HashTypePkSecp256k1
	HashTypeSkSecp256k1
	HashTypePkP256
	HashTypeSkP256
	HashTypeScalarSecp256k1
	HashTypeElementSecp256k1
	HashTypeScriptExpr
	HashTypeEncryptedSeedEd25519
	HashTypeEncryptedSkSecp256k1
	HashTypeEncryptedSkP256
	HashTypeSigEd25519
	HashTypeSigSecp256k1
	HashTypeSigP256
	HashTypeSigGeneric

	HashTypeBlockPayload
	HashTypeBlockMetadata
	HashTypeOperationMetadata
	HashTypeOperationMetadataList
	HashTypeOperationMetadataListList
	HashTypeEncryptedSecp256k1Scalar
	HashTypeSaplingSpendingKey
	HashTypeSaplingAddress

	HashTypePkhBls12_381
	HashTypeSigGenericAggregate
	HashTypeSigBls12_381
	HashTypePkBls12_381
	HashTypeSkBls12_381
	HashTypeEncryptedSkBls12_381
	HashTypeToruAddress
	HashTypeToruInbox
	HashTypeToruMessage
	HashTypeToruCommitment
	HashTypeToruMessageResult
	HashTypeToruMessageResultList
	HashTypeToruWithdrawList
	HashTypeScruAddress

	HashTypeDekuContract
)

func ParseHashType(s string) HashType {
	switch len(s) {
	case 15:
		if strings.HasPrefix(s, CHAIN_ID_PREFIX) {
			return HashTypeChainId
		}
	case 30:
		if strings.HasPrefix(s, ID_HASH_PREFIX) {
			return HashTypeId
		}
	case 36:
		switch {
		case strings.HasPrefix(s, ED25519_PUBLIC_KEY_HASH_PREFIX):
			return HashTypePkhEd25519
		case strings.HasPrefix(s, SECP256K1_PUBLIC_KEY_HASH_PREFIX):
			return HashTypePkhSecp256k1
		case strings.HasPrefix(s, P256_PUBLIC_KEY_HASH_PREFIX):
			return HashTypePkhP256
		case strings.HasPrefix(s, NOCURVE_PUBLIC_KEY_HASH_PREFIX):
			return HashTypePkhNocurve
		case strings.HasPrefix(s, BLINDED_PUBLIC_KEY_HASH_PREFIX):
			return HashTypePkhBlinded
		case strings.HasPrefix(s, BLS12_381_PUBLIC_KEY_HASH_PREFIX):
			return HashTypePkhBls12_381
		case strings.HasPrefix(s, DEKU_CONTRACT_HASH_PREFIX):
			return HashTypeDekuContract
		}
	case 37:
		switch {
		case strings.HasPrefix(s, TORU_ADDRESS_PREFIX):
			return HashTypeToruAddress
		case strings.HasPrefix(s, SCRU_ADDRESS_PREFIX):
			return HashTypeScruAddress
		}
	case 43:
		if strings.HasPrefix(s, SAPLING_ADDRESS_PREFIX) {
			return HashTypeSaplingAddress
		}
	case 51:
		switch {
		case strings.HasPrefix(s, BLOCK_HASH_PREFIX):
			return HashTypeBlock
		case strings.HasPrefix(s, OPERATION_HASH_PREFIX):
			return HashTypeOperation
		case strings.HasPrefix(s, PROTOCOL_HASH_PREFIX):
			return HashTypeProtocol
		case strings.HasPrefix(s, OPERATION_METADATA_HASH_PREFIX):
			return HashTypeOperationMetadata
		}
	case 52:
		switch {
		case strings.HasPrefix(s, OPERATION_LIST_HASH_PREFIX):
			return HashTypeOperationList
		case strings.HasPrefix(s, CONTEXT_HASH_PREFIX):
			return HashTypeContext
		case strings.HasPrefix(s, BLOCK_PAYLOAD_HASH_PREFIX):
			return HashTypeBlockPayload
		case strings.HasPrefix(s, BLOCK_METADATA_HASH_PREFIX):
			return HashTypeBlockMetadata
		case strings.HasPrefix(s, OPERATION_METADATA_LIST_HASH_PREFIX):
			return HashTypeOperationMetadataList
		}
	case 53:
		switch {
		case strings.HasPrefix(s, OPERATION_LIST_LIST_HASH_PREFIX):
			return HashTypeOperationListList
		case strings.HasPrefix(s, SECP256K1_SCALAR_PREFIX):
			return HashTypeScalarSecp256k1
		case strings.HasPrefix(s, NONCE_HASH_PREFIX):
			return HashTypeNonce
		case strings.HasPrefix(s, OPERATION_METADATA_LIST_LIST_HASH_PREFIX):
			return HashTypeOperationMetadataListList
		case strings.HasPrefix(s, TORU_INBOX_HASH_PREFIX):
			return HashTypeToruInbox
		case strings.HasPrefix(s, TORU_MESSAGE_HASH_PREFIX):
			return HashTypeToruMessage
		case strings.HasPrefix(s, TORU_COMMITMENT_HASH_PREFIX):
			return HashTypeToruCommitment
		case strings.HasPrefix(s, TORU_MESSAGE_RESULT_LIST_HASH_PREFIX):
			return HashTypeToruMessageResultList
		case strings.HasPrefix(s, TORU_WITHDRAW_LIST_HASH_PREFIX):
			return HashTypeToruWithdrawList
		}
	case 54:
		switch {
		case strings.HasPrefix(s, ED25519_SEED_PREFIX):
			return HashTypeSeedEd25519
		case strings.HasPrefix(s, ED25519_PUBLIC_KEY_PREFIX):
			return HashTypePkEd25519
		case strings.HasPrefix(s, SECP256K1_SECRET_KEY_PREFIX):
			return HashTypeSkSecp256k1
		case strings.HasPrefix(s, P256_SECRET_KEY_PREFIX):
			return HashTypeSkP256
		case strings.HasPrefix(s, SECP256K1_ELEMENT_PREFIX):
			return HashTypeElementSecp256k1
		case strings.HasPrefix(s, SCRIPT_EXPR_HASH_PREFIX):
			return HashTypeScriptExpr
		case strings.HasPrefix(s, BLS12_381_SECRET_KEY_PREFIX):
			return HashTypeSkBls12_381
		case strings.HasPrefix(s, TORU_MESSAGE_RESULT_HASH_PREFIX):
			return HashTypeToruMessageResult
		}
	case 55:
		switch {
		case strings.HasPrefix(s, SECP256K1_PUBLIC_KEY_PREFIX):
			return HashTypePkSecp256k1
		case strings.HasPrefix(s, P256_PUBLIC_KEY_PREFIX):
			return HashTypePkP256
		}
	case 76:
		if strings.HasPrefix(s, BLS12_381_PUBLIC_KEY_PREFIX) {
			return HashTypePkBls12_381
		}
	case 88:
		switch {
		case strings.HasPrefix(s, ED25519_ENCRYPTED_SEED_PREFIX):
			return HashTypeEncryptedSeedEd25519
		case strings.HasPrefix(s, SECP256K1_ENCRYPTED_SECRET_KEY_PREFIX):
			return HashTypeEncryptedSkSecp256k1
		case strings.HasPrefix(s, P256_ENCRYPTED_SECRET_KEY_PREFIX):
			return HashTypeEncryptedSkP256
		case strings.HasPrefix(s, BLS12_381_ENCRYPTED_SECRET_KEY_PREFIX):
			return HashTypeEncryptedSkBls12_381
		}
	case 93:
		if strings.HasPrefix(s, SECP256K1_ENCRYPTED_SCALAR_PREFIX) {
			return HashTypeEncryptedSecp256k1Scalar
		}
	case 96:
		if strings.HasPrefix(s, GENERIC_SIGNATURE_PREFIX) {
			return HashTypeSigGeneric
		}
	case 98:
		switch {
		case strings.HasPrefix(s, ED25519_SECRET_KEY_PREFIX):
			return HashTypeSkEd25519
		case strings.HasPrefix(s, P256_SIGNATURE_PREFIX):
			return HashTypeSigP256
		}
	case 99:
		switch {
		case strings.HasPrefix(s, ED25519_SIGNATURE_PREFIX):
			return HashTypeSigEd25519
		case strings.HasPrefix(s, SECP256K1_SIGNATURE_PREFIX):
			return HashTypeSigSecp256k1
		}
	case 141:
		if strings.HasPrefix(s, GENERIC_AGGREGATE_SIGNATURE_PREFIX) {
			return HashTypeSigGenericAggregate
		}
	case 142:
		if strings.HasPrefix(s, BLS12_381_SIGNATURE_PREFIX) {
			return HashTypeSigBls12_381
		}
	case 169:
		if strings.HasPrefix(s, SAPLING_SPENDING_KEY_PREFIX) {
			return HashTypeSaplingSpendingKey
		}
	}
	return HashTypeInvalid
}

func (t HashType) IsValid() bool {
	return t != HashTypeInvalid
}

func (t HashType) String() string {
	return t.Prefix()
}

func (t HashType) MatchPrefix(s string) bool {
	return strings.HasPrefix(s, t.Prefix())
}

func (t HashType) Prefix() string {
	switch t {
	case HashTypeChainId:
		return CHAIN_ID_PREFIX
	case HashTypeId:
		return ID_HASH_PREFIX
	case HashTypePkhEd25519:
		return ED25519_PUBLIC_KEY_HASH_PREFIX
	case HashTypePkhSecp256k1:
		return SECP256K1_PUBLIC_KEY_HASH_PREFIX
	case HashTypePkhP256:
		return P256_PUBLIC_KEY_HASH_PREFIX
	case HashTypePkhNocurve:
		return NOCURVE_PUBLIC_KEY_HASH_PREFIX
	case HashTypePkhBlinded:
		return BLINDED_PUBLIC_KEY_HASH_PREFIX
	case HashTypeBlock:
		return BLOCK_HASH_PREFIX
	case HashTypeOperation:
		return OPERATION_HASH_PREFIX
	case HashTypeOperationList:
		return OPERATION_LIST_HASH_PREFIX
	case HashTypeOperationListList:
		return OPERATION_LIST_LIST_HASH_PREFIX
	case HashTypeProtocol:
		return PROTOCOL_HASH_PREFIX
	case HashTypeContext:
		return CONTEXT_HASH_PREFIX
	case HashTypeNonce:
		return NONCE_HASH_PREFIX
	case HashTypeSeedEd25519:
		return ED25519_SEED_PREFIX
	case HashTypePkEd25519:
		return ED25519_PUBLIC_KEY_PREFIX
	case HashTypeSkEd25519:
		return ED25519_SECRET_KEY_PREFIX
	case HashTypePkSecp256k1:
		return SECP256K1_PUBLIC_KEY_PREFIX
	case HashTypeSkSecp256k1:
		return SECP256K1_SECRET_KEY_PREFIX
	case HashTypePkP256:
		return P256_PUBLIC_KEY_PREFIX
	case HashTypeSkP256:
		return P256_SECRET_KEY_PREFIX
	case HashTypeScalarSecp256k1:
		return SECP256K1_SCALAR_PREFIX
	case HashTypeElementSecp256k1:
		return SECP256K1_ELEMENT_PREFIX
	case HashTypeScriptExpr:
		return SCRIPT_EXPR_HASH_PREFIX
	case HashTypeEncryptedSeedEd25519:
		return ED25519_ENCRYPTED_SEED_PREFIX
	case HashTypeEncryptedSkSecp256k1:
		return SECP256K1_ENCRYPTED_SECRET_KEY_PREFIX
	case HashTypeEncryptedSkP256:
		return P256_ENCRYPTED_SECRET_KEY_PREFIX
	case HashTypeSigEd25519:
		return ED25519_SIGNATURE_PREFIX
	case HashTypeSigSecp256k1:
		return SECP256K1_SIGNATURE_PREFIX
	case HashTypeSigP256:
		return P256_SIGNATURE_PREFIX
	case HashTypeSigGeneric:
		return GENERIC_SIGNATURE_PREFIX
	case HashTypeBlockPayload:
		return BLOCK_PAYLOAD_HASH_PREFIX
	case HashTypeBlockMetadata:
		return BLOCK_METADATA_HASH_PREFIX
	case HashTypeOperationMetadata:
		return OPERATION_METADATA_HASH_PREFIX
	case HashTypeOperationMetadataList:
		return OPERATION_METADATA_LIST_HASH_PREFIX
	case HashTypeOperationMetadataListList:
		return OPERATION_METADATA_LIST_LIST_HASH_PREFIX
	case HashTypeEncryptedSecp256k1Scalar:
		return SECP256K1_ENCRYPTED_SCALAR_PREFIX
	case HashTypeSaplingSpendingKey:
		return SAPLING_SPENDING_KEY_PREFIX
	case HashTypeSaplingAddress:
		return SAPLING_ADDRESS_PREFIX
	case HashTypePkhBls12_381:
		return BLS12_381_PUBLIC_KEY_HASH_PREFIX
	case HashTypeSigGenericAggregate:
		return GENERIC_AGGREGATE_SIGNATURE_PREFIX
	case HashTypeSigBls12_381:
		return BLS12_381_SIGNATURE_PREFIX
	case HashTypePkBls12_381:
		return BLS12_381_PUBLIC_KEY_PREFIX
	case HashTypeSkBls12_381:
		return BLS12_381_SECRET_KEY_PREFIX
	case HashTypeEncryptedSkBls12_381:
		return BLS12_381_ENCRYPTED_SECRET_KEY_PREFIX
	case HashTypeToruAddress:
		return TORU_ADDRESS_PREFIX
	case HashTypeToruInbox:
		return TORU_INBOX_HASH_PREFIX
	case HashTypeToruMessage:
		return TORU_MESSAGE_HASH_PREFIX
	case HashTypeToruCommitment:
		return TORU_COMMITMENT_HASH_PREFIX
	case HashTypeToruMessageResult:
		return TORU_MESSAGE_RESULT_HASH_PREFIX
	case HashTypeToruMessageResultList:
		return TORU_MESSAGE_RESULT_LIST_HASH_PREFIX
	case HashTypeToruWithdrawList:
		return TORU_WITHDRAW_LIST_HASH_PREFIX
	case HashTypeScruAddress:
		return SCRU_ADDRESS_PREFIX
	case HashTypeDekuContract:
		return DEKU_CONTRACT_HASH_PREFIX
	default:
		return ""
	}
}

func (t HashType) PrefixBytes() []byte {
	switch t {
	case HashTypeChainId:
		return CHAIN_ID
	case HashTypeId:
		return ID_HASH_ID
	case HashTypePkhEd25519:
		return ED25519_PUBLIC_KEY_HASH_ID
	case HashTypePkhSecp256k1:
		return SECP256K1_PUBLIC_KEY_HASH_ID
	case HashTypePkhP256:
		return P256_PUBLIC_KEY_HASH_ID
	case HashTypePkhNocurve:
		return NOCURVE_PUBLIC_KEY_HASH_ID
	case HashTypePkhBlinded:
		return BLINDED_PUBLIC_KEY_HASH_ID
	case HashTypeBlock:
		return BLOCK_HASH_ID
	case HashTypeOperation:
		return OPERATION_HASH_ID
	case HashTypeOperationList:
		return OPERATION_LIST_HASH_ID
	case HashTypeOperationListList:
		return OPERATION_LIST_LIST_HASH_ID
	case HashTypeProtocol:
		return PROTOCOL_HASH_ID
	case HashTypeContext:
		return CONTEXT_HASH_ID
	case HashTypeNonce:
		return NONCE_HASH_ID
	case HashTypeSeedEd25519:
		return ED25519_SEED_ID
	case HashTypePkEd25519:
		return ED25519_PUBLIC_KEY_ID
	case HashTypeSkEd25519:
		return ED25519_SECRET_KEY_ID
	case HashTypePkSecp256k1:
		return SECP256K1_PUBLIC_KEY_ID
	case HashTypeSkSecp256k1:
		return SECP256K1_SECRET_KEY_ID
	case HashTypePkP256:
		return P256_PUBLIC_KEY_ID
	case HashTypeSkP256:
		return P256_SECRET_KEY_ID
	case HashTypeScalarSecp256k1:
		return SECP256K1_SCALAR_ID
	case HashTypeElementSecp256k1:
		return SECP256K1_ELEMENT_ID
	case HashTypeScriptExpr:
		return SCRIPT_EXPR_HASH_ID
	case HashTypeEncryptedSeedEd25519:
		return ED25519_ENCRYPTED_SEED_ID
	case HashTypeEncryptedSkSecp256k1:
		return SECP256K1_ENCRYPTED_SECRET_KEY_ID
	case HashTypeEncryptedSkP256:
		return P256_ENCRYPTED_SECRET_KEY_ID
	case HashTypeSigEd25519:
		return ED25519_SIGNATURE_ID
	case HashTypeSigSecp256k1:
		return SECP256K1_SIGNATURE_ID
	case HashTypeSigP256:
		return P256_SIGNATURE_ID
	case HashTypeSigGeneric:
		return GENERIC_SIGNATURE_ID
	case HashTypeBlockPayload:
		return BLOCK_PAYLOAD_HASH_ID
	case HashTypeBlockMetadata:
		return BLOCK_METADATA_HASH_ID
	case HashTypeOperationMetadata:
		return OPERATION_METADATA_HASH_ID
	case HashTypeOperationMetadataList:
		return OPERATION_METADATA_LIST_HASH_ID
	case HashTypeOperationMetadataListList:
		return OPERATION_METADATA_LIST_LIST_HASH_ID
	case HashTypeEncryptedSecp256k1Scalar:
		return SECP256K1_ENCRYPTED_SCALAR_ID
	case HashTypeSaplingSpendingKey:
		return SAPLING_SPENDING_KEY_ID
	case HashTypeSaplingAddress:
		return SAPLING_ADDRESS_ID
	case HashTypePkhBls12_381:
		return BLS12_381_PUBLIC_KEY_HASH_ID
	case HashTypeSigGenericAggregate:
		return GENERIC_AGGREGATE_SIGNATURE_ID
	case HashTypeSigBls12_381:
		return BLS12_381_SIGNATURE_ID
	case HashTypePkBls12_381:
		return BLS12_381_PUBLIC_KEY_ID
	case HashTypeSkBls12_381:
		return BLS12_381_SECRET_KEY_ID
	case HashTypeEncryptedSkBls12_381:
		return BLS12_381_ENCRYPTED_SECRET_KEY_ID
	case HashTypeToruAddress:
		return TORU_ADDRESS_ID
	case HashTypeToruInbox:
		return TORU_INBOX_HASH_ID
	case HashTypeToruMessage:
		return TORU_MESSAGE_HASH_ID
	case HashTypeToruCommitment:
		return TORU_COMMITMENT_HASH_ID
	case HashTypeToruMessageResult:
		return TORU_MESSAGE_RESULT_HASH_ID
	case HashTypeToruMessageResultList:
		return TORU_MESSAGE_RESULT_LIST_HASH_ID
	case HashTypeToruWithdrawList:
		return TORU_WITHDRAW_LIST_HASH_ID
	case HashTypeDekuContract:
		return DEKU_CONTRACT_HASH_ID
	case HashTypeScruAddress:
		return SCRU_ADDRESS_ID
	default:
		return nil
	}
}

func (t HashType) Len() int {
	switch t {
	case HashTypeChainId:
		return 4
	case HashTypeId:
		return 16
	case HashTypePkhEd25519,
		HashTypePkhSecp256k1,
		HashTypePkhP256,
		HashTypePkhNocurve,
		HashTypePkhBlinded,
		HashTypePkhBls12_381,
		HashTypeToruAddress,
		HashTypeScruAddress,
		HashTypeDekuContract:
		return 20
	case HashTypeBlock,
		HashTypeOperation,
		HashTypeOperationList,
		HashTypeOperationListList,
		HashTypeProtocol,
		HashTypeContext,
		HashTypeNonce,
		HashTypeSeedEd25519,
		HashTypePkEd25519,
		HashTypeSkSecp256k1,
		HashTypeSkP256,
		HashTypeScriptExpr,
		HashTypeBlockPayload,
		HashTypeBlockMetadata,
		HashTypeOperationMetadata,
		HashTypeOperationMetadataList,
		HashTypeOperationMetadataListList,
		HashTypeToruInbox,
		HashTypeToruMessage,
		HashTypeToruCommitment,
		HashTypeToruMessageResultList,
		HashTypeToruWithdrawList,
		HashTypeToruMessageResult,
		HashTypeSkBls12_381:
		return 32
	case HashTypePkSecp256k1,
		HashTypePkP256,
		HashTypeScalarSecp256k1,
		HashTypeElementSecp256k1:
		return 33
	case HashTypeSaplingAddress:
		return 43
	case HashTypePkBls12_381:
		return 48
	case HashTypeEncryptedSeedEd25519,
		HashTypeEncryptedSkSecp256k1,
		HashTypeEncryptedSkP256:
		return 56
	case HashTypeEncryptedSkBls12_381:
		return 58
	case HashTypeEncryptedSecp256k1Scalar:
		return 60
	case HashTypeSkEd25519,
		HashTypeSigEd25519,
		HashTypeSigSecp256k1,
		HashTypeSigP256,
		HashTypeSigGeneric:
		return 64
	case HashTypeSigGenericAggregate,
		HashTypeSigBls12_381:
		return 96
	case HashTypeSaplingSpendingKey:
		return 169
	default:
		return 0
	}
}

func (t HashType) Base58Len() int {
	switch t {
	case HashTypeChainId:
		return 15
	case HashTypeId,
		HashTypePkhEd25519,
		HashTypePkhSecp256k1,
		HashTypePkhP256,
		HashTypePkhNocurve,
		HashTypePkhBls12_381,
		HashTypeDekuContract:
		return 36
	case HashTypePkhBlinded,
		HashTypeToruAddress,
		HashTypeScruAddress:
		return 37
	case HashTypeBlock,
		HashTypeOperation,
		HashTypeProtocol,
		HashTypeOperationMetadata:
		return 51
	case HashTypeOperationList,
		HashTypeContext,
		HashTypeBlockPayload,
		HashTypeBlockMetadata,
		HashTypeOperationMetadataList:
		return 52
	case HashTypeOperationListList,
		HashTypeNonce,
		HashTypeScalarSecp256k1,
		HashTypeOperationMetadataListList,
		HashTypeToruInbox,
		HashTypeToruMessage,
		HashTypeToruCommitment,
		HashTypeToruMessageResultList,
		HashTypeToruWithdrawList:
		return 53
	case HashTypeSeedEd25519,
		HashTypePkEd25519,
		HashTypeSkSecp256k1,
		HashTypeSkP256,
		HashTypeElementSecp256k1,
		HashTypeScriptExpr,
		HashTypeToruMessageResult,
		HashTypeSkBls12_381:
		return 54
	case HashTypePkSecp256k1,
		HashTypePkP256:
		return 55
	case HashTypeSaplingAddress:
		return 69
	case HashTypePkBls12_381:
		return 76
	case HashTypeEncryptedSeedEd25519,
		HashTypeEncryptedSkSecp256k1,
		HashTypeEncryptedSkP256,
		HashTypeEncryptedSkBls12_381:
		return 88
	case HashTypeEncryptedSecp256k1Scalar:
		return 93
	case HashTypeSigGeneric:
		return 96
	case HashTypeSkEd25519,
		HashTypeSigP256:
		return 98
	case HashTypeSigEd25519,
		HashTypeSigSecp256k1:
		return 99
	case HashTypeSigGenericAggregate:
		return 141
	case HashTypeSigBls12_381:
		return 142
	case HashTypeSaplingSpendingKey:
		return 241
	default:
		return 0
	}
}

type Hash struct {
	Type HashType
	Hash []byte
}

func NewHash(typ HashType, hash []byte) Hash {
	return Hash{
		Type: typ,
		Hash: hash,
	}
}

func (h Hash) IsValid() bool {
	return h.Type != HashTypeInvalid && len(h.Hash) == h.Type.Len()
}

func (h Hash) IsEmpty() bool {
	return len(h.Hash) == 0
}

func (h Hash) IsZero() bool {
	zero := make([]byte, h.Type.Len())
	return len(h.Hash) == 0 || bytes.Equal(h.Hash, zero)
}

func (h Hash) Equal(h2 Hash) bool {
	return h.Type == h2.Type && bytes.Equal(h.Hash, h2.Hash)
}

func (h Hash) Clone() Hash {
	buf := make([]byte, len(h.Hash))
	copy(buf, h.Hash)
	return Hash{
		Type: h.Type,
		Hash: buf,
	}
}

func (h *Hash) Reset() {
	h.Type = HashTypeInvalid
	h.Hash = nil
}

// String returns the string encoding of the hash.
func (h Hash) String() string {
	s, _ := encodeHash(h.Type, h.Hash)
	return s
}

// Int64 ensures interface compatibility with the RPC packages' BlockID type
func (h Hash) Int64() int64 {
	return -1
}

// Bytes returns the raw byte representation of the hash without type info.
func (h Hash) Bytes() []byte {
	return h.Hash
}

func (h Hash) Short() string {
	s := h.String()
	if len(s) < 12 {
		return s
	}
	return s[:8] + "..." + s[len(s)-4:]
}

func ParseHash(s string) (Hash, error) {
	return decodeHash(s)
}

func (h *Hash) UnmarshalText(data []byte) error {
	x, err := decodeHash(string(data))
	if err != nil {
		return err
	}
	*h = x
	return nil
}

func (h Hash) MarshalText() ([]byte, error) {
	if h.IsValid() {
		return []byte(h.String()), nil
	}
	return nil, nil
}

func (h Hash) MarshalBinary() ([]byte, error) {
	return h.Hash, nil
}

// ChainIdHash
type ChainIdHash struct {
	Hash
}

func NewChainIdHash(buf []byte) ChainIdHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return ChainIdHash{Hash: NewHash(HashTypeChainId, b)}
}

func (h ChainIdHash) Equal(h2 ChainIdHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h ChainIdHash) Clone() ChainIdHash {
	return ChainIdHash{h.Hash.Clone()}
}

func (h *ChainIdHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), CHAIN_ID_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for chain id hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeChainId {
		return fmt.Errorf("tezos: invalid type %s for chain id hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for chain id hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *ChainIdHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeChainId.Len() {
		return fmt.Errorf("tezos: invalid len %d for chain id hash", len(data))
	}
	h.Type = HashTypeChainId
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func (h ChainIdHash) Uint32() uint32 {
	return binary.BigEndian.Uint32(h.Hash.Hash)
}

func MustParseChainIdHash(s string) ChainIdHash {
	h, err := ParseChainIdHash(s)
	if err != nil {
		panic(err)
	}
	return h
}

func ParseChainIdHash(s string) (ChainIdHash, error) {
	var h ChainIdHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ChainIdHash) Set(hash string) (err error) {
	*h, err = ParseChainIdHash(hash)
	return
}

// BlockHash
type BlockHash struct {
	Hash
}

func NewBlockHash(buf []byte) BlockHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return BlockHash{Hash: NewHash(HashTypeBlock, b)}
}

func (h BlockHash) Clone() BlockHash {
	return BlockHash{h.Hash.Clone()}
}

func (h BlockHash) Equal(h2 BlockHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *BlockHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), BLOCK_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for block hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeBlock {
		return fmt.Errorf("tezos: invalid type %s for block hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for block hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *BlockHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeBlock.Len() {
		return fmt.Errorf("tezos: invalid len %d for block hash", len(data))
	}
	h.Type = HashTypeBlock
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func MustParseBlockHash(s string) BlockHash {
	h, err := ParseBlockHash(s)
	if err != nil {
		panic(err)
	}
	return h
}

func ParseBlockHash(s string) (BlockHash, error) {
	var h BlockHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *BlockHash) Set(hash string) (err error) {
	*h, err = ParseBlockHash(hash)
	return
}

// ProtocolHash
type ProtocolHash struct {
	Hash
}

func NewProtocolHash(buf []byte) ProtocolHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return ProtocolHash{Hash: NewHash(HashTypeProtocol, b)}
}

func (h ProtocolHash) Clone() ProtocolHash {
	return ProtocolHash{h.Hash.Clone()}
}

func (h ProtocolHash) Equal(h2 ProtocolHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *ProtocolHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), PROTOCOL_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for protocol hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeProtocol {
		return fmt.Errorf("tezos: invalid type %s for protocol hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for protocol hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *ProtocolHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeProtocol.Len() {
		return fmt.Errorf("tezos: invalid len %d for protocol hash", len(data))
	}
	h.Type = HashTypeProtocol
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func ParseProtocolHash(s string) (ProtocolHash, error) {
	var h ProtocolHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

func MustParseProtocolHash(s string) ProtocolHash {
	b, err := ParseProtocolHash(s)
	if err != nil {
		panic(err)
	}
	return b
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ProtocolHash) Set(hash string) (err error) {
	*h, err = ParseProtocolHash(hash)
	return
}

// OpHash
type OpHash struct {
	Hash
}

func NewOpHash(buf []byte) OpHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return OpHash{Hash: NewHash(HashTypeOperation, b)}
}

func (h OpHash) Clone() OpHash {
	return OpHash{h.Hash.Clone()}
}

func (h OpHash) Equal(h2 OpHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *OpHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), OPERATION_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for operation hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeOperation {
		return fmt.Errorf("tezos: invalid type %s for operation hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for operation hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *OpHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeOperation.Len() {
		return fmt.Errorf("tezos: invalid len %d for operation hash", len(data))
	}
	h.Type = HashTypeOperation
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func MustParseOpHash(s string) OpHash {
	b, err := ParseOpHash(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ParseOpHash(s string) (OpHash, error) {
	var h OpHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *OpHash) Set(hash string) (err error) {
	*h, err = ParseOpHash(hash)
	return
}

// OpListListHash
type OpListListHash struct {
	Hash
}

func NewOpListListHash(buf []byte) OpListListHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return OpListListHash{Hash: NewHash(HashTypeOperationListList, b)}
}

func (h OpListListHash) Clone() OpListListHash {
	return OpListListHash{h.Hash.Clone()}
}

func (h OpListListHash) Equal(h2 OpListListHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *OpListListHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), OPERATION_LIST_LIST_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for operation list list hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeOperationListList {
		return fmt.Errorf("tezos: invalid type %s for operation list list hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for operation list list hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *OpListListHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeOperationListList.Len() {
		return fmt.Errorf("tezos: invalid len %d for operation list list hash", len(data))
	}
	h.Type = HashTypeOperationListList
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func MustParseOpListListHash(s string) OpListListHash {
	b, err := ParseOpListListHash(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ParseOpListListHash(s string) (OpListListHash, error) {
	var h OpListListHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *OpListListHash) Set(hash string) (err error) {
	*h, err = ParseOpListListHash(hash)
	return
}

// PayloadHash
type PayloadHash struct {
	Hash
}

func NewPayloadHash(buf []byte) PayloadHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return PayloadHash{Hash: NewHash(HashTypeBlockPayload, b)}
}

func (h PayloadHash) Clone() PayloadHash {
	return PayloadHash{h.Hash.Clone()}
}

func (h PayloadHash) Equal(h2 PayloadHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *PayloadHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), BLOCK_PAYLOAD_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for payload hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeBlockPayload {
		return fmt.Errorf("tezos: invalid type %s for payload hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for payload hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *PayloadHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeBlockPayload.Len() {
		return fmt.Errorf("tezos: invalid len %d for payload hash", len(data))
	}
	h.Type = HashTypeBlockPayload
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func MustParsePayloadHash(s string) PayloadHash {
	b, err := ParsePayloadHash(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ParsePayloadHash(s string) (PayloadHash, error) {
	var h PayloadHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *PayloadHash) Set(hash string) (err error) {
	*h, err = ParsePayloadHash(hash)
	return
}

// ExprHash
type ExprHash struct {
	Hash
}

func NewExprHash(buf []byte) ExprHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return ExprHash{Hash: NewHash(HashTypeScriptExpr, b)}
}

func (h ExprHash) Clone() ExprHash {
	return ExprHash{h.Hash.Clone()}
}

func (h ExprHash) Equal(h2 ExprHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *ExprHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), SCRIPT_EXPR_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for script expression hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeScriptExpr {
		return fmt.Errorf("tezos: invalid type %s for script expression hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for script expression hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *ExprHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeScriptExpr.Len() {
		return fmt.Errorf("tezos: invalid len %d for script expression hash", len(data))
	}
	h.Type = HashTypeScriptExpr
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func MustParseExprHash(s string) ExprHash {
	b, err := ParseExprHash(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ParseExprHash(s string) (ExprHash, error) {
	var h ExprHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ExprHash) Set(hash string) (err error) {
	*h, err = ParseExprHash(hash)
	return
}

// NonceHash
type NonceHash struct {
	Hash
}

func NewNonceHash(buf []byte) NonceHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return NonceHash{Hash: NewHash(HashTypeNonce, b)}
}

func (h NonceHash) Clone() NonceHash {
	return NonceHash{h.Hash.Clone()}
}

func (h NonceHash) Equal(h2 NonceHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *NonceHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), NONCE_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for nonce hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeNonce {
		return fmt.Errorf("tezos: invalid type %s for nonce hash '%s'", h.Type.Prefix(), string(data))
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for nonce hash '%s'", len(h.Hash.Hash), string(data))
	}
	return nil
}

func (h *NonceHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeNonce.Len() {
		return fmt.Errorf("tezos: invalid len %d for nonce hash '%s'", len(data), string(data))
	}
	h.Type = HashTypeNonce
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func ParseNonceHash(s string) (NonceHash, error) {
	var h NonceHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

func MustParseNonceHash(s string) NonceHash {
	b, err := ParseNonceHash(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ParseNonceHashSafe(s string) NonceHash {
	var h NonceHash
	h.UnmarshalText([]byte(s))
	return h
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *NonceHash) Set(hash string) (err error) {
	*h, err = ParseNonceHash(hash)
	return
}

// ContextHash
type ContextHash struct {
	Hash
}

func NewContextHash(buf []byte) ContextHash {
	b := make([]byte, len(buf))
	copy(b, buf)
	return ContextHash{Hash: NewHash(HashTypeContext, b)}
}

func (h ContextHash) Clone() ContextHash {
	return ContextHash{h.Hash.Clone()}
}

func (h ContextHash) Equal(h2 ContextHash) bool {
	return h.Hash.Equal(h2.Hash)
}

func (h *ContextHash) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if !strings.HasPrefix(string(data), CONTEXT_HASH_PREFIX) {
		return fmt.Errorf("tezos: invalid prefix for context hash '%s'", string(data))
	}
	if err := h.Hash.UnmarshalText(data); err != nil {
		return err
	}
	if h.Type != HashTypeContext {
		return fmt.Errorf("tezos: invalid type %s for context hash", h.Type.Prefix())
	}
	if len(h.Hash.Hash) != h.Type.Len() {
		return fmt.Errorf("tezos: invalid len %d for context hash", len(h.Hash.Hash))
	}
	return nil
}

func (h *ContextHash) UnmarshalBinary(data []byte) error {
	if l := len(data); l > 0 && l != HashTypeContext.Len() {
		return fmt.Errorf("tezos: invalid len %d for context hash", len(data))
	}
	h.Type = HashTypeContext
	h.Hash.Hash = make([]byte, h.Type.Len())
	copy(h.Hash.Hash, data)
	return nil
}

func MustParseContextHash(s string) ContextHash {
	b, err := ParseContextHash(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ParseContextHash(s string) (ContextHash, error) {
	var h ContextHash
	if err := h.UnmarshalText([]byte(s)); err != nil {
		return h, err
	}
	return h, nil
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ContextHash) Set(hash string) (err error) {
	*h, err = ParseContextHash(hash)
	return
}

// internal decoders
func decodeHash(hstr string) (Hash, error) {
	typ := ParseHashType(hstr)
	if typ == HashTypeInvalid {
		return Hash{}, ErrUnknownHashType
	}
	decoded, version, err := base58.CheckDecode(hstr, len(typ.PrefixBytes()), nil)
	if err != nil {
		if err == base58.ErrChecksum {
			return Hash{}, ErrChecksumMismatch
		}
		return Hash{}, fmt.Errorf("tezos: unknown hash format: %w", err)
	}
	if !bytes.Equal(version, typ.PrefixBytes()) {
		return Hash{}, fmt.Errorf("tezos: invalid prefix '%x' for decoded hash type '%s'", version, typ)
	}
	if have, want := len(decoded), typ.Len(); have != want {
		return Hash{}, fmt.Errorf("tezos: invalid length for decoded hash have=%d want=%d", have, want)
	}
	return Hash{
		Type: typ,
		Hash: decoded,
	}, nil
}

func encodeHash(typ HashType, h []byte) (string, error) {
	if typ == HashTypeInvalid {
		return "", ErrUnknownHashType
	}
	if have, want := len(h), typ.Len(); have != want {
		return "", fmt.Errorf("tezos: invalid hash length have=%d want=%d", have, want)
	}
	return base58.CheckEncode(h, typ.PrefixBytes()), nil
}
