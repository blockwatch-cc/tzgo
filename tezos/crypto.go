// Copyright (c) 2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
    "fmt"
    "math/big"

    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rand"
    "crypto/sha512"

    "github.com/decred/dcrd/dcrec/secp256k1"
    "golang.org/x/crypto/nacl/secretbox"
    "golang.org/x/crypto/pbkdf2"
)

// ecNormalizeSignature ensures strict compliance with the EC spec by returning
// S mod n for the appropriate keys curve.
//
// Details:
//   Step #6 of the ECDSA algorithm [x] defines an `S` value mod n[0],
//   but most signers (OpenSSL, SoftHSM, YubiHSM) don't return a strict modulo.
//   This variability was exploited with transaction malleability in Bitcoin,
//   leading to BIP#62.  BIP#62 Rule #5[1] requires that signatures return a
//   strict S = ... mod n which this function forces implemented in btcd here [2]
//   [0]: https://en.wikipedia.org/wiki/Elliptic_Curve_Digital_Signature_Algorithm
//   [1]: https://github.com/bitcoin/bips/blob/master/bip-0062.mediawiki#new-rules
//   [2]: https://github.com/btcsuite/btcd/blob/master/btcec/signature.go#L49
//
// See also Ecadlabs Signatory:
// https://github.com/ecadlabs/signatory/blob/f57871c2300cb5a53236ea5fcb4f203012b4fe41/pkg/cryptoutils/crypto.go#L17
func ecNormalizeSignature(r, s *big.Int, c elliptic.Curve) (*big.Int, *big.Int) {
    r = new(big.Int).Set(r)
    s = new(big.Int).Set(s)

    order := c.Params().N
    quo := new(big.Int).Quo(order, new(big.Int).SetInt64(2))
    if s.Cmp(quo) > 0 {
        s = s.Sub(order, s)
    }
    return r, s
}

func ecSign(sk *ecdsa.PrivateKey, hash []byte) ([]byte, error) {
    r, s, err := ecdsa.Sign(rand.Reader, sk, hash)
    if err != nil {
        return nil, err
    }
    // normalize
    r, s = ecNormalizeSignature(r, s, sk.Curve)
    // serialize
    buf := make([]byte, 64)
    r.FillBytes(buf[:32])
    s.FillBytes(buf[32:])
    return buf, nil
}

func ecVerifySignature(pk *ecdsa.PublicKey, hash []byte, sig Signature) bool {
    r := new(big.Int).SetBytes(sig.Data[:32])
    s := new(big.Int).SetBytes(sig.Data[32:])
    return ecdsa.Verify(pk, hash, r, s)
}

func ecPrivateKeyFromBytes(b []byte, curve elliptic.Curve) (key *ecdsa.PrivateKey, err error) {
    k := new(big.Int).SetBytes(b)
    curveOrder := curve.Params().N
    if k.Cmp(curveOrder) >= 0 {
        return nil, fmt.Errorf("tezos: invalid private key for curve %s", curve.Params().Name)
    }

    priv := &ecdsa.PrivateKey{
        PublicKey: ecdsa.PublicKey{
            Curve: curve,
        },
        D: k,
    }

    // https://cs.opensource.google/go/go/+/refs/tags/go1.17.5:src/crypto/ecdsa/ecdsa.go;l=149
    priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())
    return priv, nil
}

// See https://github.com/golang/go/blob/master/src/crypto/elliptic/elliptic.go
func ecUnmarshalCompressed(curve elliptic.Curve, data []byte) (pk *ecdsa.PublicKey, err error) {
    byteLen := (curve.Params().BitSize + 7) / 8
    if len(data) != 1+byteLen {
        return nil, fmt.Errorf("tezos: (%s) invalid public key length: %d", curve.Params().Name, len(data))
    }
    if data[0] != 2 && data[0] != 3 { // compressed form
        return nil, fmt.Errorf("tezos: (%s) invalid public key compression", curve.Params().Name)
    }
    p := curve.Params().P
    x := new(big.Int).SetBytes(data[1:])
    if x.Cmp(p) >= 0 {
        return nil, fmt.Errorf("tezos: (%s) invalid public key", curve.Params().Name)
    }

    // secp256k1 polynomial: x³ + b
    // P-* polynomial: x³ - 3x + b
    y := new(big.Int).Mul(x, x)
    y.Mul(y, x)
    if curve != secp256k1.S256() {
        x1 := new(big.Int).Lsh(x, 1)
        x1.Add(x1, x)
        y.Sub(y, x1)
    }
    y.Add(y, curve.Params().B)
    y.Mod(y, curve.Params().P)
    y.ModSqrt(y, p)

    if y == nil {
        return nil, fmt.Errorf("tezos: (%s) invalid public key", curve.Params().Name)
    }
    if byte(y.Bit(0)) != data[0]&1 {
        y.Neg(y).Mod(y, p)
    }
    if !curve.IsOnCurve(x, y) {
        return nil, fmt.Errorf("tezos: (%s) invalid public key", curve.Params().Name)
    }

    pk = &ecdsa.PublicKey{
        Curve: curve,
        X:     x,
        Y:     y,
    }
    return
}

const (
    encIterations = 32768
    encKeyLen     = 32
)

func decryptPrivateKey(enc []byte, fn PassphraseFunc) ([]byte, error) {
    if fn == nil {
        return nil, ErrPassphrase
    }
    passphrase, err := fn()
    if err != nil {
        return nil, err
    }
    if len(passphrase) == 0 {
        return nil, ErrPassphrase
    }

    salt, box := enc[:8], enc[8:]
    secretboxKey := pbkdf2.Key(passphrase, salt, encIterations, encKeyLen, sha512.New)

    var (
        tmp   [32]byte
        nonce [24]byte // implicitly 0x00..
    )
    copy(tmp[:], secretboxKey)
    dec, ok := secretbox.Open(nil, box, &nonce, &tmp)
    if !ok {
        return nil, fmt.Errorf("tezos: private key decrypt failed")
    }
    return dec, nil
}

func encryptPrivateKey(key []byte, fn PassphraseFunc) ([]byte, error) {
    if fn == nil {
        return nil, ErrPassphrase
    }
    passphrase, err := fn()
    if err != nil {
        return nil, err
    }
    if len(passphrase) == 0 {
        return nil, ErrPassphrase
    }

    salt := make([]byte, 8)
    _, err = rand.Read(salt)
    if err != nil {
        return nil, err
    }
    secretboxKey := pbkdf2.Key(passphrase, salt, encIterations, encKeyLen, sha512.New)

    var (
        tmp   [32]byte
        nonce [24]byte // implicitly 0x00..
    )
    copy(tmp[:], secretboxKey)
    enc := secretbox.Seal(nil, key, &nonce, &tmp)
    return append(salt, enc...), nil
}
