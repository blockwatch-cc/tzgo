// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"testing"
)

func TestSig(t *testing.T) {
	type testcase struct {
		Name  string
		Sig   string
		Type  SignatureType
		Bytes string
	}

	cases := []testcase{
		// edsig
		{
			Name:  "edsig",
			Sig:   "edsigtzWvLTwvEqaZy1BMzQoeFTCxALJ94aDx5YyDh6qhYNQowHfAb7k23doKazVMGvGnT6bCeTG9qbJfBqRqeL64zpEFLJyp9C",
			Type:  SignatureTypeEd25519,
			Bytes: "00cc2854776cd0ece57f46a03fde2cce2b95798409f344f379a7a7a0f8835b3fdc76a7fa6827c72f52eeebf2508d35e7b8ac63c90372fa3cb0aaa98c06f96c5e05",
		},
		// spsig
		{
			Name:  "spsig",
			Sig:   "spsig1Joen9pavJsoyjRQ6xEHFEDvB3WUDxX9LxbJKstvXss8b5tJpRqZBMPvWqvQ5xqavjcvpDe4TyvcfJjuYxKBeqGtSUAkBr",
			Type:  SignatureTypeSecp256k1,
			Bytes: "01635310a0b871b948e38dc476edcb9f135a8f148609243a60d9ddccf8b15a350769c2f80fa30fbd1504e84ab88277f081bde79adb2416a898b1e88d0b2e692ceb",
		},
		// p2sig
		{
			Name:  "p2sig",
			Sig:   "p2sigSDjb3CfSRVEYt9X75HAFe29Tf3hqSrQRsnyxvD8k7NDka9SY7oxDmZQGngJcD5E7E4AmVD2iiSveWtWYwT3mMurSMZicA",
			Type:  SignatureTypeP256,
			Bytes: "0225967eb0e75b6a4cab2d764ab0fce1304c6dbca43e7e5f562a8685294226f05f062b4ea9188ac5c5332ccdebdaf465d05c317a8c82a06d649c1ca0f9ac2fb5ac",
		},
		// BLsig
		{
			Name:  "BLsig",
			Sig:   "BLsigAqfbS14US8aPsoe6xu6VbQ3ukXZGbhx7X3WVmk2UpTvkZW4bkEctwvZ8S8ajprdDUfArjc6m4JqWRpffpK6jHKc23hToq8LtCs1fqXB3nfPeAQqiqo5Fe6DoomuJi9NXMMxLQ8N8k",
			Type:  SignatureTypeBls12_381,
			Bytes: "03a25136c9169bf72d62260b49e55c02371e827a5c58b1d5f92b3caee7b90f3abaa4fd384ecbe0489dd06cb2a2f64dcbfe0080508b5593ccdb15c13ab904bab34082db468fae469881568b5afcc70979d964412e1342e3b85f6a9a9b70fbfac021",
		},
		// sig
		{
			Name:  "sig",
			Sig:   "sigphTDQj1xuPXupC4rNxTJ9vSPWDD5wMMhsDUxkqiY2PrdTia6DpvxaZQYPsZBT6S1Z7aiwUnzyLtT3PdzBk6omA9TpiPrq",
			Type:  SignatureTypeGeneric,
			Bytes: "cc2854776cd0ece57f46a03fde2cce2b95798409f344f379a7a7a0f8835b3fdc76a7fa6827c72f52eeebf2508d35e7b8ac63c90372fa3cb0aaa98c06f96c5e05",
		},
		// asig
		// {
		//     Name:  "asig",
		//     Sig:   "",
		//     Type:  SignatureTypeGenericAggregate,
		//     Bytes: "",
		// },
	}

	for _, c := range cases {
		buf := MustDecodeString(c.Bytes)

		// base58 must parse
		sig, err := ParseSignature(c.Sig)
		if err != nil {
			t.Fatalf("%s: parsing signature %s: %v", c.Name, c.Sig, err)
		}

		// check type
		if got, want := sig.Type, c.Type; got != want {
			t.Errorf("%s: mismatched type got=%s want=%s", c.Name, got, want)
		}

		// check hash
		if !bytes.Equal(sig.Bytes(), buf) {
			t.Errorf("%s: mismatched hash got=%x want=%x", c.Name, sig.Bytes(), buf)
		}

		// marshal text
		out, err := sig.MarshalText()
		if err != nil {
			t.Errorf("%s: marshal text unexpected error: %v", c.Name, err)
		}

		if got, want := string(out), c.Sig; got != want {
			t.Errorf("%s: mismatched text encoding got=%s want=%s", c.Name, got, want)
		}

		// unmarshal from bytes
		var s2 Signature
		err = s2.UnmarshalBinary(buf)
		if err != nil {
			t.Fatalf("%s: unmarshal binary %s: %v", c.Name, c.Bytes, err)
		}

		if !s2.Equal(sig) {
			t.Errorf("%s: mismatched signature got=%s want=%s", c.Name, s2, sig)
		}

		// unmarshal text
		err = s2.UnmarshalText([]byte(c.Sig))
		if err != nil {
			t.Fatalf("%s: unmarshal text %s: %v", c.Name, c.Sig, err)
		}

		if !s2.Equal(sig) {
			t.Errorf("%s: mismatched signature got=%s want=%s", c.Name, s2, sig)
		}
	}
}

func TestInvalidSig(t *testing.T) {
	// invalid base58 string
	if _, err := ParseAddress("sigVZ3WLNWdhN7VmZxjVycgkkjmGJqSZyAmWctqueGX6NGgdtxCkBMXmyYmREiVpJ3zSMLZFvbksANKnKCs3soZrsVi999tX"); err == nil {
		t.Errorf("Expected error on invalid base58 string")
	}

	// init from invalid short hash
	hash := MustDecodeString("0b78887fdd0cd3bfbe75a717655728e0205bb9")
	s := NewSignature(SignatureTypeEd25519, hash)
	if s.IsValid() {
		t.Errorf("Expected invalid signature from short hash")
	}

	// init from invalid empty bytes
	s = NewSignature(SignatureTypeEd25519, nil)
	if s.IsValid() {
		t.Errorf("Expected invalid signature from nil hash")
	}

	// decode from short buffer
	err := s.UnmarshalBinary(MustDecodeString("000b78887fdd0cd3bfbe75a717655728e0205bb9"))
	if err == nil || s.IsValid() {
		t.Errorf("Expected unmarshal error from short buffer")
	}

	// decode from nil buffer
	err = s.UnmarshalBinary(nil)
	if err == nil || s.IsValid() {
		t.Errorf("Expected unmarshal error from short buffer")
	}

	// decode from invalid buffer (wrong type)
	// err = s.UnmarshalBinary(MustDecodeString(""))
	// if err == nil || s.IsValid() {
	//     t.Errorf("Expected unmarshal error from invalid buffer")
	// }
}
