// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"encoding"
	"encoding/binary"
	// "encoding/hex"
	"fmt"
	"testing"
)

type Marshallable interface {
	encoding.TextUnmarshaler
	encoding.BinaryUnmarshaler
	fmt.Stringer
}

func TestHash(t *testing.T) {
	type testcase struct {
		String string
		Bytes  []byte
		Type   HashType
		Val    Marshallable
	}

	cases := []testcase{
		// chain id
		{
			String: "NetXdQprcVkpaWU",
			Bytes:  MustDecodeString("7a06a770"),
			Type:   HashTypeChainId,
			Val:    &ChainIdHash{},
		},
		// block
		{
			String: "BKjS7rtCjysnMNWUuevZiF2a6NkUas9bnSsNQ3ibh5GfKNrQoGk",
			Bytes:  MustDecodeString("029d4ed3161d644bedccb8673f30c6682b6e0a11756a3f75d7a739dede1cf29e"),
			Type:   HashTypeBlock,
			Val:    &BlockHash{},
		},
		// protocol
		{
			String: "PtLimaPtLMwfNinJi9rCfDPWea8dFgTZ1MeJ9f1m2SRic6ayiwW",
			Bytes:  MustDecodeString("d57ed88be5a69815e39386a33f7dcad391f5f507e03b376e499272c86c6cf2a7"),
			Type:   HashTypeProtocol,
			Val:    &ProtocolHash{},
		},
		// op
		{
			String: "oogC8ju9tMDqeB6RiAXdch3hnt8u3Pbf2ZXyyhAmJAhjQ4q1wUS",
			Bytes:  MustDecodeString("88315c911f6b4c38b2d8ea27319cf91d3614e1dde486dc83d55cd47bfbc568b4"),
			Type:   HashTypeOperation,
			Val:    &OpHash{},
		},
		// op list list
		{
			String: "LLoZnUuxzhNESHg7HvXxoccUCvWPrVmAucjaJDfKBeea39LqyVKEP",
			Bytes:  MustDecodeString("3ccb42fba1b24ce6c4e99ed1c4674dc90542d9fbc5694b9220ed74feb0eb3507"),
			Type:   HashTypeOperationListList,
			Val:    &OpListListHash{},
		},
		// payload
		{
			String: "vh26kqZ6LeKwygSY9JYZbqSdhtRjphAcrSEJqJ6a8EgnQEMtQx8J",
			Bytes:  MustDecodeString("37eec128736b994b3ce44a36c81dd00b2eab68057c11900a8135eaa5bff606fa"),
			Type:   HashTypeBlockPayload,
			Val:    &PayloadHash{},
		},
		// expr
		{
			String: "expruPBWMccKybChcJmGF8oMo263Ri6HgbKbAJRS8j6GbmqZJPJfVG",
			Bytes:  MustDecodeString("72386e5b4dbe9cc415bb0d909fe63e3162bc4b662ea4bfc92c485ad12f6700e8"),
			Type:   HashTypeScriptExpr,
			Val:    &ExprHash{},
		},
		// nonce
		{
			String: "nceUcirB7QYmVgcNUYQvd1fTzqSHjyusoE8VmJX9SNgELeQ4ffdcr",
			Bytes:  MustDecodeString("33b55c290efcc31f235a3809198ff324b31c65088a18815b8c67bf5ed1567dcd"),
			Type:   HashTypeNonce,
			Val:    &NonceHash{},
		},
		// context
		{
			String: "CoWAJ6dKTDySvhrV5njZwHpckRSBgzb84vXZKrEd1AwyshxFb9vo",
			Bytes:  MustDecodeString("c7c8fad1f5d2c144edbc763dc4e009c1fca3637dd265ae2c0044168065986b2b"),
			Type:   HashTypeContext,
			Val:    &ContextHash{},
		},
	}

	for i, c := range cases {
		// base58 must parse
		if err := c.Val.UnmarshalText([]byte(c.String)); err != nil {
			t.Fatalf("Case %d - unmarshal %s hash %s: %v", i, c.Type, c.String, err)
		}

		// write binary
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, c.Val)
		if err != nil {
			t.Fatalf("Case %d - write binary for hash %s: %v", i, c.String, err)
		}

		// check binary
		if !bytes.Equal(buf.Bytes(), c.Bytes) {
			t.Errorf("Case %d - mismatched hash got=%x want=%x", i, buf.Bytes(), c.Bytes)
		}

		// unmarshal from bytes
		if err := c.Val.UnmarshalBinary(c.Bytes); err != nil {
			t.Fatalf("Case %d - unmarshal binary %s: %v", i, c.Bytes, err)
		}

		// marshal text
		s := c.Val.String()

		// check text
		if s != c.String {
			t.Errorf("Case %d - mismatched text encoding got=%x want=%x", i, s, c.String)
		}
	}
}

func TestInvalidHash(t *testing.T) {
	// invalid base58 string
	if _, err := ParseBlockHash("tz1KzpjBnunNJVABHBnzfG4iuLmphitExW2"); err == nil {
		t.Errorf("Expected error on invalid base58 string")
	}

	// decode from short buffer
	var b BlockHash
	err := b.UnmarshalBinary(MustDecodeString("000b78887fdd0cd3bfbe75a717655728e0205bb9"))
	if err == nil {
		t.Errorf("Expected unmarshal error from short buffer")
	}

	// decode from nil buffer is OK
	err = b.UnmarshalBinary(nil)
	if err != nil {
		t.Errorf("Expected no unmarshal error from nil buffer, got %v", err)
	}

	// decode from empty string is OK
	err = b.UnmarshalText(nil)
	if err != nil {
		t.Errorf("Expected no unmarshal error from null string, got %v", err)
	}
}

func BenchmarkHashDecode(b *testing.B) {
	b.SetBytes(32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ParseBlockHash("BKjS7rtCjysnMNWUuevZiF2a6NkUas9bnSsNQ3ibh5GfKNrQoGk")
	}
}

func BenchmarkHashEncode(b *testing.B) {
	b.SetBytes(32)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ZeroBlockHash.String()
	}
}
