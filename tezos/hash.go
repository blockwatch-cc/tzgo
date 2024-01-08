// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"blockwatch.cc/tzgo/base58"
)

var (
	// ErrUnknownHashType describes an error where a hash can not
	// decoded as a specific hash type because the string encoding
	// starts with an unknown identifier.
	ErrUnknownHashType = errors.New("tezos: unknown hash type")

	// Zero hashes
	ZeroChainIdHash           = NewChainIdHash(nil)
	ZeroBlockHash             = NewBlockHash(nil)
	ZeroProtocolHash          = NewProtocolHash(nil)
	ZeroOpHash                = NewOpHash(nil)
	ZeroOpListListHash        = NewOpListListHash(nil)
	ZeroPayloadHash           = NewPayloadHash(nil)
	ZeroExprHash              = NewExprHash(nil)
	ZeroNonceHash             = NewNonceHash(nil)
	ZeroContextHash           = NewContextHash(nil)
	ZeroSmartRollupStateHash  = NewSmartRollupStateHash(nil)
	ZeroSmartRollupCommitHash = NewSmartRollupCommitHash(nil)
)

type HashType struct {
	Id        []byte
	Len       int
	B58Prefix string
	B58Len    int
}

var (
	HashTypeInvalid              = HashType{nil, 0, "", 0}
	HashTypeChainId              = HashType{CHAIN_ID, 4, CHAIN_ID_PREFIX, 15}
	HashTypeId                   = HashType{ID_HASH_ID, 16, ID_HASH_PREFIX, 36}
	HashTypePkhEd25519           = HashType{ED25519_PUBLIC_KEY_HASH_ID, 20, ED25519_PUBLIC_KEY_HASH_PREFIX, 36}
	HashTypePkhSecp256k1         = HashType{SECP256K1_PUBLIC_KEY_HASH_ID, 20, SECP256K1_PUBLIC_KEY_HASH_PREFIX, 36}
	HashTypePkhP256              = HashType{P256_PUBLIC_KEY_HASH_ID, 20, P256_PUBLIC_KEY_HASH_PREFIX, 36}
	HashTypePkhNocurve           = HashType{NOCURVE_PUBLIC_KEY_HASH_ID, 20, NOCURVE_PUBLIC_KEY_HASH_PREFIX, 36}
	HashTypePkhBlinded           = HashType{BLINDED_PUBLIC_KEY_HASH_ID, 20, BLINDED_PUBLIC_KEY_HASH_PREFIX, 37}
	HashTypeBlock                = HashType{BLOCK_HASH_ID, 32, BLOCK_HASH_PREFIX, 51}
	HashTypeOperation            = HashType{OPERATION_HASH_ID, 32, OPERATION_HASH_PREFIX, 51}
	HashTypeOperationList        = HashType{OPERATION_LIST_HASH_ID, 32, OPERATION_LIST_HASH_PREFIX, 52}
	HashTypeOperationListList    = HashType{OPERATION_LIST_LIST_HASH_ID, 32, OPERATION_LIST_LIST_HASH_PREFIX, 53}
	HashTypeProtocol             = HashType{PROTOCOL_HASH_ID, 32, PROTOCOL_HASH_PREFIX, 51}
	HashTypeContext              = HashType{CONTEXT_HASH_ID, 32, CONTEXT_HASH_PREFIX, 52}
	HashTypeNonce                = HashType{NONCE_HASH_ID, 32, NONCE_HASH_PREFIX, 53}
	HashTypeSeedEd25519          = HashType{ED25519_SEED_ID, 32, ED25519_SEED_PREFIX, 54}
	HashTypePkEd25519            = HashType{ED25519_PUBLIC_KEY_ID, 32, ED25519_PUBLIC_KEY_PREFIX, 54}
	HashTypeSkEd25519            = HashType{ED25519_SECRET_KEY_ID, 64, ED25519_SECRET_KEY_PREFIX, 98}
	HashTypePkSecp256k1          = HashType{SECP256K1_PUBLIC_KEY_ID, 33, SECP256K1_PUBLIC_KEY_PREFIX, 55}
	HashTypeSkSecp256k1          = HashType{SECP256K1_SECRET_KEY_ID, 32, SECP256K1_SECRET_KEY_PREFIX, 54}
	HashTypePkP256               = HashType{P256_PUBLIC_KEY_ID, 33, P256_PUBLIC_KEY_PREFIX, 55}
	HashTypeSkP256               = HashType{P256_SECRET_KEY_ID, 32, P256_SECRET_KEY_PREFIX, 54}
	HashTypeScalarSecp256k1      = HashType{SECP256K1_SCALAR_ID, 33, SECP256K1_SCALAR_PREFIX, 53}
	HashTypeElementSecp256k1     = HashType{SECP256K1_ELEMENT_ID, 33, SECP256K1_ELEMENT_PREFIX, 54}
	HashTypeScriptExpr           = HashType{SCRIPT_EXPR_HASH_ID, 32, SCRIPT_EXPR_HASH_PREFIX, 54}
	HashTypeEncryptedSeedEd25519 = HashType{ED25519_ENCRYPTED_SEED_ID, 56, ED25519_ENCRYPTED_SEED_PREFIX, 88}
	HashTypeEncryptedSkSecp256k1 = HashType{SECP256K1_ENCRYPTED_SECRET_KEY_ID, 56, SECP256K1_ENCRYPTED_SECRET_KEY_PREFIX, 88}
	HashTypeEncryptedSkP256      = HashType{P256_ENCRYPTED_SECRET_KEY_ID, 56, P256_ENCRYPTED_SECRET_KEY_PREFIX, 88}
	HashTypeSigEd25519           = HashType{ED25519_SIGNATURE_ID, 64, ED25519_SIGNATURE_PREFIX, 99}
	HashTypeSigSecp256k1         = HashType{SECP256K1_SIGNATURE_ID, 64, SECP256K1_SIGNATURE_PREFIX, 99}
	HashTypeSigP256              = HashType{P256_SIGNATURE_ID, 64, P256_SIGNATURE_PREFIX, 98}
	HashTypeSigGeneric           = HashType{GENERIC_SIGNATURE_ID, 64, GENERIC_SIGNATURE_PREFIX, 96}

	HashTypeBlockPayload              = HashType{BLOCK_PAYLOAD_HASH_ID, 32, BLOCK_PAYLOAD_HASH_PREFIX, 52}
	HashTypeBlockMetadata             = HashType{BLOCK_METADATA_HASH_ID, 32, BLOCK_METADATA_HASH_PREFIX, 52}
	HashTypeOperationMetadata         = HashType{OPERATION_METADATA_HASH_ID, 32, OPERATION_METADATA_HASH_PREFIX, 51}
	HashTypeOperationMetadataList     = HashType{OPERATION_METADATA_LIST_HASH_ID, 32, OPERATION_METADATA_LIST_HASH_PREFIX, 52}
	HashTypeOperationMetadataListList = HashType{OPERATION_METADATA_LIST_LIST_HASH_ID, 32, OPERATION_METADATA_LIST_LIST_HASH_PREFIX, 53}
	HashTypeEncryptedSecp256k1Scalar  = HashType{SECP256K1_ENCRYPTED_SCALAR_ID, 60, SECP256K1_ENCRYPTED_SCALAR_PREFIX, 93}
	HashTypeSaplingSpendingKey        = HashType{SAPLING_SPENDING_KEY_ID, 169, SAPLING_SPENDING_KEY_PREFIX, 241}
	HashTypeSaplingAddress            = HashType{SAPLING_ADDRESS_ID, 43, SAPLING_ADDRESS_PREFIX, 69}

	HashTypePkhBls12_381              = HashType{BLS12_381_PUBLIC_KEY_HASH_ID, 20, BLS12_381_PUBLIC_KEY_HASH_PREFIX, 36}
	HashTypeSigGenericAggregate       = HashType{GENERIC_AGGREGATE_SIGNATURE_ID, 96, GENERIC_AGGREGATE_SIGNATURE_PREFIX, 141}
	HashTypeSigBls12_381              = HashType{BLS12_381_SIGNATURE_ID, 96, BLS12_381_SIGNATURE_PREFIX, 142}
	HashTypePkBls12_381               = HashType{BLS12_381_PUBLIC_KEY_ID, 48, BLS12_381_PUBLIC_KEY_PREFIX, 76}
	HashTypeSkBls12_381               = HashType{BLS12_381_SECRET_KEY_ID, 32, BLS12_381_SECRET_KEY_PREFIX, 54}
	HashTypeEncryptedSkBls12_381      = HashType{BLS12_381_ENCRYPTED_SECRET_KEY_ID, 58, BLS12_381_ENCRYPTED_SECRET_KEY_PREFIX, 88}
	HashTypeTxRollupAddress           = HashType{TX_ROLLUP_ADDRESS_ID, 20, TX_ROLLUP_ADDRESS_PREFIX, 37}
	HashTypeTxRollupInbox             = HashType{TX_ROLLUP_INBOX_HASH_ID, 32, TX_ROLLUP_INBOX_HASH_PREFIX, 53}
	HashTypeTxRollupMessage           = HashType{TX_ROLLUP_MESSAGE_HASH_ID, 32, TX_ROLLUP_MESSAGE_HASH_PREFIX, 53}
	HashTypeTxRollupCommitment        = HashType{TX_ROLLUP_COMMITMENT_HASH_ID, 32, TX_ROLLUP_COMMITMENT_HASH_PREFIX, 53}
	HashTypeTxRollupMessageResult     = HashType{TX_ROLLUP_MESSAGE_RESULT_HASH_ID, 32, TX_ROLLUP_MESSAGE_RESULT_HASH_PREFIX, 54}
	HashTypeTxRollupMessageResultList = HashType{TX_ROLLUP_MESSAGE_RESULT_LIST_HASH_ID, 32, TX_ROLLUP_MESSAGE_RESULT_LIST_HASH_PREFIX, 53}
	HashTypeTxRollupWithdrawList      = HashType{TX_ROLLUP_WITHDRAW_LIST_HASH_ID, 32, TX_ROLLUP_WITHDRAW_LIST_HASH_PREFIX, 53}
	HashTypeSmartRollupAddress        = HashType{SMART_ROLLUP_ADDRESS_ID, 20, SMART_ROLLUP_ADDRESS_PREFIX, 36}
	HashTypeSmartRollupStateHash      = HashType{SMART_ROLLUP_STATE_HASH_ID, 32, SMART_ROLLUP_STATE_HASH_PREFIX, 54}
	HashTypeSmartRollupCommitHash     = HashType{SMART_ROLLUP_COMMITMENT_HASH_ID, 32, SMART_ROLLUP_COMMITMENT_HASH_PREFIX, 54}
	HashTypeSmartRollupRevealHash     = HashType{SMART_ROLLUP_REVEAL_HASH_ID, 32, SMART_ROLLUP_REVEAL_HASH_PREFIX, 56}
)

func (t HashType) IsValid() bool {
	return len(t.Id) > 0
}

func (t HashType) String() string {
	return t.B58Prefix
}

func (t HashType) Equal(x HashType) bool {
	return t.B58Prefix == x.B58Prefix
}

// ChainIdHash
type ChainIdHash [4]byte

func NewChainIdHash(buf []byte) (h ChainIdHash) {
	copy(h[:], buf)
	return
}

func (h ChainIdHash) IsValid() bool {
	return !h.Equal(ZeroChainIdHash)
}

func (h ChainIdHash) Equal(h2 ChainIdHash) bool {
	return h == h2
}

func (h ChainIdHash) Clone() ChainIdHash {
	return NewChainIdHash(h[:])
}

func (h ChainIdHash) String() string {
	return base58.CheckEncode(h[:], HashTypeChainId.Id)
}

func (h ChainIdHash) Bytes() []byte {
	return h[:]
}

func (h ChainIdHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *ChainIdHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeChainId, h[:])
}

func (h ChainIdHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *ChainIdHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeChainId.Len {
		return fmt.Errorf("tezos: short chain_id hash")
	}
	copy(h[:], buf)
	return nil
}

func (h ChainIdHash) Uint32() uint32 {
	return binary.BigEndian.Uint32(h[:])
}

func ParseChainIdHash(s string) (h ChainIdHash, err error) {
	err = decodeHashString(s, HashTypeChainId, h[:])
	return
}

func MustParseChainIdHash(s string) ChainIdHash {
	h, err := ParseChainIdHash(s)
	panicOnError(err)
	return h
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ChainIdHash) Set(s string) (err error) {
	*h, err = ParseChainIdHash(s)
	return
}

// BlockHash
type BlockHash [32]byte

func NewBlockHash(buf []byte) (h BlockHash) {
	copy(h[:], buf)
	return
}

func (h BlockHash) IsValid() bool {
	return !h.Equal(ZeroBlockHash)
}

func (h BlockHash) Equal(h2 BlockHash) bool {
	return h == h2
}

func (h BlockHash) Clone() BlockHash {
	return NewBlockHash(h[:])
}

func (h BlockHash) String() string {
	return base58.CheckEncode(h[:], HashTypeBlock.Id)
}

func (h BlockHash) Bytes() []byte {
	return h[:]
}

func (h BlockHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *BlockHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeBlock, h[:])
}

func (h BlockHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *BlockHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeBlock.Len {
		return fmt.Errorf("tezos: short block hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseBlockHash(s string) (h BlockHash, err error) {
	err = decodeHashString(s, HashTypeBlock, h[:])
	return
}

func MustParseBlockHash(s string) BlockHash {
	h, err := ParseBlockHash(s)
	panicOnError(err)
	return h
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *BlockHash) Set(s string) (err error) {
	*h, err = ParseBlockHash(s)
	return
}

// Int64 ensures interface compatibility with the RPC packages' BlockID type
func (h BlockHash) Int64() int64 {
	return -1
}

// ProtocolHash
type ProtocolHash [32]byte

func NewProtocolHash(buf []byte) (h ProtocolHash) {
	copy(h[:], buf)
	return
}

func (h ProtocolHash) IsValid() bool {
	return !h.Equal(ZeroProtocolHash)
}

func (h ProtocolHash) Equal(h2 ProtocolHash) bool {
	return h == h2
}

func (h ProtocolHash) Clone() ProtocolHash {
	return NewProtocolHash(h[:])
}

func (h ProtocolHash) String() string {
	return base58.CheckEncode(h[:], HashTypeProtocol.Id)
}

func (h ProtocolHash) Bytes() []byte {
	return h[:]
}

func (h ProtocolHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *ProtocolHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeProtocol, h[:])
}

func (h ProtocolHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *ProtocolHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeProtocol.Len {
		return fmt.Errorf("tezos: short protocol hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseProtocolHash(s string) (h ProtocolHash, err error) {
	err = decodeHashString(s, HashTypeProtocol, h[:])
	return
}

func MustParseProtocolHash(s string) ProtocolHash {
	h, err := ParseProtocolHash(s)
	panicOnError(err)
	return h
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ProtocolHash) Set(s string) (err error) {
	*h, err = ParseProtocolHash(s)
	return
}

// OpHash
type OpHash [32]byte

func NewOpHash(buf []byte) (h OpHash) {
	copy(h[:], buf)
	return
}

func (h OpHash) IsValid() bool {
	return !h.Equal(ZeroOpHash)
}

func (h OpHash) Equal(h2 OpHash) bool {
	return h == h2
}

func (h OpHash) Clone() OpHash {
	return NewOpHash(h[:])
}

func (h OpHash) String() string {
	return base58.CheckEncode(h[:], HashTypeOperation.Id)
}

func (h OpHash) Bytes() []byte {
	return h[:]
}

func (h OpHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *OpHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeOperation, h[:])
}

func (h OpHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *OpHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeOperation.Len {
		return fmt.Errorf("tezos: short operation hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseOpHash(s string) (h OpHash, err error) {
	err = decodeHashString(s, HashTypeOperation, h[:])
	return
}

func MustParseOpHash(s string) OpHash {
	h, err := ParseOpHash(s)
	panicOnError(err)
	return h
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *OpHash) Set(s string) (err error) {
	*h, err = ParseOpHash(s)
	return
}

// OpListListHash
type OpListListHash [32]byte

func NewOpListListHash(buf []byte) (h OpListListHash) {
	copy(h[:], buf)
	return
}

func (h OpListListHash) IsValid() bool {
	return !h.Equal(ZeroOpListListHash)
}

func (h OpListListHash) Equal(h2 OpListListHash) bool {
	return h == h2
}

func (h OpListListHash) Clone() OpListListHash {
	return NewOpListListHash(h[:])
}

func (h OpListListHash) String() string {
	return base58.CheckEncode(h[:], HashTypeOperationListList.Id)
}

func (h OpListListHash) Bytes() []byte {
	return h[:]
}

func (h OpListListHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *OpListListHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeOperationListList, h[:])
}

func (h OpListListHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *OpListListHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeOperationListList.Len {
		return fmt.Errorf("tezos: short operation list list hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseOpListListHash(s string) (h OpListListHash, err error) {
	err = decodeHashString(s, HashTypeOperationListList, h[:])
	return
}

func MustParseOpListListHash(s string) OpListListHash {
	b, err := ParseOpListListHash(s)
	panicOnError(err)
	return b
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *OpListListHash) Set(s string) (err error) {
	*h, err = ParseOpListListHash(s)
	return
}

// PayloadHash
type PayloadHash [32]byte

func NewPayloadHash(buf []byte) (h PayloadHash) {
	copy(h[:], buf)
	return
}

func (h PayloadHash) IsValid() bool {
	return !h.Equal(ZeroPayloadHash)
}

func (h PayloadHash) Equal(h2 PayloadHash) bool {
	return h == h2
}

func (h PayloadHash) Clone() PayloadHash {
	return NewPayloadHash(h[:])
}

func (h PayloadHash) String() string {
	return base58.CheckEncode(h[:], HashTypeBlockPayload.Id)
}

func (h PayloadHash) Bytes() []byte {
	return h[:]
}

func (h PayloadHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *PayloadHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeBlockPayload, h[:])
}

func (h PayloadHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *PayloadHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeBlockPayload.Len {
		return fmt.Errorf("teshortd for payload hash")
	}
	copy(h[:], buf)
	return nil
}

func ParsePayloadHash(s string) (h PayloadHash, err error) {
	err = decodeHashString(s, HashTypeBlockPayload, h[:])
	return
}

func MustParsePayloadHash(s string) PayloadHash {
	b, err := ParsePayloadHash(s)
	panicOnError(err)
	return b
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *PayloadHash) Set(hash string) (err error) {
	*h, err = ParsePayloadHash(hash)
	return
}

// ExprHash
type ExprHash [32]byte

func NewExprHash(buf []byte) (h ExprHash) {
	copy(h[:], buf)
	return
}

func (h ExprHash) IsValid() bool {
	return !h.Equal(ZeroExprHash)
}

func (h ExprHash) Equal(h2 ExprHash) bool {
	return h == h2
}

func (h ExprHash) Clone() ExprHash {
	return NewExprHash(h[:])
}

func (h ExprHash) String() string {
	return base58.CheckEncode(h[:], HashTypeScriptExpr.Id)
}

func (h ExprHash) Bytes() []byte {
	return h[:]
}

func (h ExprHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *ExprHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeScriptExpr, h[:])
}

func (h ExprHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *ExprHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeScriptExpr.Len {
		return fmt.Errorf("tezoshortfor script expression hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseExprHash(s string) (h ExprHash, err error) {
	err = decodeHashString(s, HashTypeScriptExpr, h[:])
	return
}

func MustParseExprHash(s string) ExprHash {
	b, err := ParseExprHash(s)
	panicOnError(err)
	return b
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ExprHash) Set(hash string) (err error) {
	*h, err = ParseExprHash(hash)
	return
}

// NonceHash
type NonceHash [32]byte

func NewNonceHash(buf []byte) (h NonceHash) {
	copy(h[:], buf)
	return
}

func (h NonceHash) IsValid() bool {
	return !h.Equal(ZeroNonceHash)
}

func (h NonceHash) Equal(h2 NonceHash) bool {
	return h == h2
}

func (h NonceHash) Clone() NonceHash {
	return NewNonceHash(h[:])
}

func (h NonceHash) String() string {
	return base58.CheckEncode(h[:], HashTypeNonce.Id)
}

func (h NonceHash) Bytes() []byte {
	return h[:]
}

func (h NonceHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *NonceHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeNonce, h[:])
}

func (h NonceHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *NonceHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeNonce.Len {
		return fmt.Errorf("tezos: short nonce")
	}
	copy(h[:], buf)
	return nil
}

func ParseNonceHash(s string) (h NonceHash, err error) {
	err = decodeHashString(s, HashTypeNonce, h[:])
	return
}

func MustParseNonceHash(s string) NonceHash {
	b, err := ParseNonceHash(s)
	panicOnError(err)
	return b
}

func ParseNonceHashSafe(s string) (h NonceHash) {
	_ = decodeHashString(s, HashTypeNonce, h[:])
	return
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *NonceHash) Set(hash string) (err error) {
	*h, err = ParseNonceHash(hash)
	return
}

// ContextHash
type ContextHash [32]byte

func NewContextHash(buf []byte) (h ContextHash) {
	copy(h[:], buf)
	return
}

func (h ContextHash) IsValid() bool {
	return !h.Equal(ZeroContextHash)
}

func (h ContextHash) Equal(h2 ContextHash) bool {
	return h == h2
}

func (h ContextHash) Clone() ContextHash {
	return NewContextHash(h[:])
}

func (h ContextHash) String() string {
	return base58.CheckEncode(h[:], HashTypeContext.Id)
}

func (h ContextHash) Bytes() []byte {
	return h[:]
}

func (h ContextHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *ContextHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeContext, h[:])
}

func (h ContextHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *ContextHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeContext.Len {
		return fmt.Errorf("tezos: short context hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseContextHash(s string) (h ContextHash, err error) {
	err = decodeHashString(s, HashTypeContext, h[:])
	return
}

func MustParseContextHash(s string) ContextHash {
	b, err := ParseContextHash(s)
	panicOnError(err)
	return b
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *ContextHash) Set(hash string) (err error) {
	*h, err = ParseContextHash(hash)
	return
}

// SmartRollupCommitHash
type SmartRollupCommitHash [32]byte

func NewSmartRollupCommitHash(buf []byte) (h SmartRollupCommitHash) {
	copy(h[:], buf)
	return
}

func (h SmartRollupCommitHash) IsValid() bool {
	return !h.Equal(ZeroSmartRollupCommitHash)
}

func (h SmartRollupCommitHash) Equal(h2 SmartRollupCommitHash) bool {
	return h == h2
}

func (h SmartRollupCommitHash) Clone() SmartRollupCommitHash {
	return NewSmartRollupCommitHash(h[:])
}

func (h SmartRollupCommitHash) String() string {
	return base58.CheckEncode(h[:], HashTypeSmartRollupCommitHash.Id)
}

func (h SmartRollupCommitHash) Bytes() []byte {
	return h[:]
}

func (h SmartRollupCommitHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *SmartRollupCommitHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeSmartRollupCommitHash, h[:])
}

func (h SmartRollupCommitHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *SmartRollupCommitHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeSmartRollupCommitHash.Len {
		return fmt.Errorf("tezos: short smart rollup commit hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseSmartRollupCommitHash(s string) (h SmartRollupCommitHash, err error) {
	err = decodeHashString(s, HashTypeSmartRollupCommitHash, h[:])
	return
}

func MustParseSmartRollupCommitHash(s string) SmartRollupCommitHash {
	b, err := ParseSmartRollupCommitHash(s)
	panicOnError(err)
	return b
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *SmartRollupCommitHash) Set(hash string) (err error) {
	*h, err = ParseSmartRollupCommitHash(hash)
	return
}

// SmartRollupStateHash
type SmartRollupStateHash [32]byte

func NewSmartRollupStateHash(buf []byte) (h SmartRollupStateHash) {
	copy(h[:], buf)
	return
}

func (h SmartRollupStateHash) IsValid() bool {
	return !h.Equal(ZeroSmartRollupStateHash)
}

func (h SmartRollupStateHash) Equal(h2 SmartRollupStateHash) bool {
	return h == h2
}

func (h SmartRollupStateHash) Clone() SmartRollupStateHash {
	return NewSmartRollupStateHash(h[:])
}

func (h SmartRollupStateHash) String() string {
	return base58.CheckEncode(h[:], HashTypeSmartRollupStateHash.Id)
}

func (h SmartRollupStateHash) Bytes() []byte {
	return h[:]
}

func (h SmartRollupStateHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *SmartRollupStateHash) UnmarshalText(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	return decodeHash(buf, HashTypeSmartRollupStateHash, h[:])
}

func (h SmartRollupStateHash) MarshalBinary() ([]byte, error) {
	return h[:], nil
}

func (h *SmartRollupStateHash) UnmarshalBinary(buf []byte) error {
	if l := len(buf); l > 0 && l != HashTypeSmartRollupStateHash.Len {
		return fmt.Errorf("tezos: short smart rollup commit hash")
	}
	copy(h[:], buf)
	return nil
}

func ParseSmartRollupStateHash(s string) (h SmartRollupStateHash, err error) {
	err = decodeHashString(s, HashTypeSmartRollupStateHash, h[:])
	return
}

func MustParseSmartRollupStateHash(s string) SmartRollupStateHash {
	b, err := ParseSmartRollupStateHash(s)
	panicOnError(err)
	return b
}

// Set implements the flags.Value interface for use in command line argument parsing.
func (h *SmartRollupStateHash) Set(hash string) (err error) {
	*h, err = ParseSmartRollupStateHash(hash)
	return
}

// internal decoders
func decodeHash(src []byte, typ HashType, dst []byte) error {
	return decodeHashString(string(src), typ, dst)
}

func decodeHashString(src string, typ HashType, dst []byte) error {
	if len(src) == 0 {
		return nil
	}
	ibuf := bufPool32.Get()
	dec, ver, err := base58.CheckDecode(src, len(typ.Id), ibuf.([]byte))
	if err != nil {
		bufPool32.Put(ibuf)
		if err == base58.ErrChecksum {
			return ErrChecksumMismatch
		}
		return fmt.Errorf("tezos: unknown hash format: %w", err)
	}
	if !bytes.Equal(ver, typ.Id) {
		bufPool32.Put(ibuf)
		return fmt.Errorf("tezos: invalid prefix '%x' for decoded hash type '%s'", ver, typ)
	}
	copy(dst, dec)
	bufPool32.Put(ibuf)
	return nil
}
