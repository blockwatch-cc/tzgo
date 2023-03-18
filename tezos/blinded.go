// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"fmt"

	"blockwatch.cc/tzgo/base58"
	"golang.org/x/crypto/blake2b"
)

// blinded pks
// def genesis_commitments(wallets, blind):
//     commitments = []
//     for pkh_b58, amount in wallets.iteritems():
//         # Public key hash corresponding to this Tezos address.
//         pkh = bitcoin.b58check_to_bin(pkh_b58)[2:]
//         # The redemption code is unique to the public key hash and deterministically
//         # constructed using a secret blinding value.
//         secret = secret_code(pkh, blind)
//         # The redemption code is used to blind the pkh
//         blinded_pkh = blake2b(pkh, 20, key=secret).digest()
//         commitment = {
//             'blinded_pkh': bitcoin.bin_to_b58check(blinded_pkh, magicbyte=16921055),
//             'amount': amount
//         }
//         commitments.append(commitment)
//     return commitments

func BlindHash(hash, secret []byte) ([]byte, error) {
	h, err := blake2b.New(20, secret)
	if err != nil {
		return nil, err
	}
	h.Write(hash)
	return h.Sum(nil), nil
}

func BlindAddress(a Address, secret []byte) (Address, error) {
	bh, err := BlindHash(a[1:], secret)
	if err != nil {
		return Address{}, err
	}
	return NewAddress(AddressTypeBlinded, bh), nil
}

// Checks if address a when blinded with secret equals blinded address b.
func MatchBlindedAddress(a, b Address, secret []byte) bool {
	bh, _ := BlindHash(a[1:], secret)
	return bytes.Equal(bh, b[1:])
}

func DecodeBlindedAddress(addr string) (a Address, err error) {
	ibuf := bufPool32.Get()
	dec, ver, err2 := base58.CheckDecode(addr, 4, ibuf.([]byte))
	if err2 != nil {
		if err == base58.ErrChecksum {
			err = ErrChecksumMismatch
			return
		}
		err = fmt.Errorf("tezos: decoded address is of unknown format: %w", err2)
		return
	}
	if len(dec) != 20 {
		err = fmt.Errorf("tezos: decoded address hash has invalid length %d", len(dec))
		return
	}
	if !bytes.Equal(ver, BLINDED_PUBLIC_KEY_HASH_ID) {
		err = fmt.Errorf("tezos: decoded address %s is of unknown type %x", addr, ver)
		return
	}
	a[0] = byte(AddressTypeBlinded)
	copy(a[1:], dec)
	bufPool32.Put(ibuf)
	return
}

func EncodeBlindedAddress(hash, secret []byte) (string, error) {
	bh, err := BlindHash(hash, secret)
	if err != nil {
		return "", err
	}
	return EncodeAddress(AddressTypeBlinded, bh), nil
}
