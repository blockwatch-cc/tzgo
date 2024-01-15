// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"testing"

	"blockwatch.cc/tzgo/tezos"
)

func TestBlock(t *testing.T) {
	type testcase struct {
		data  tezos.HexBytes
		head  BlockHeader
		chain tezos.ChainIdHash
		hash  tezos.BlockHash
		key   tezos.PrivateKey
		sig   tezos.Signature
	}

	var cases = []testcase{
		// Tenderbake block
		{
			data: asHex("0004ad5702d21acd0569ff8e03cd564fdc15baae8e436b141510f4ca966bdadfe092904359000000006242e26904cf318e718893b9efb0a426130f6d8fac752db1c47a98d0c3f89780ec8b1a4740000000210000000102000000040004ad570000000000000004ffffffff00000004000000016ae0589f63d96d15d6b41b4e9a9c6f5670ae7e4a3495ffdaf0fa651a10b9e25d9253ed831d88bb031de4f49e43d62977864806a7b0945e8877b030150f2ae63b00000001df2ea592260c01000000"),
			head: BlockHeader{
				Level:          306519,
				Proto:          2,
				Predecessor:    tezos.MustParseBlockHash("BMJpBGs6rDpEGki8vLVd6VAcLrnEnAhxAwpGjExRcT8qDCmwQQm"),
				Timestamp:      asTime("2022-03-29T10:41:45Z"),
				ValidationPass: 4,
				OperationsHash: tezos.MustParseOpListListHash("LLoatxVfWnkjHGYBuXL6ELtKqkX1EvLr5ffHSd5pySMY88AY2MeMr"),
				Fitness: []tezos.HexBytes{
					asHex("02"),
					asHex("0004ad57"),
					asHex(""),
					asHex("ffffffff"),
					asHex("00000001"),
				},
				Context:          tezos.MustParseContextHash("CoVTNsN2t3DU6m6nL2HL3qEzCA2jhiUEAnWZde7dPmzcdm61EYBp"),
				PayloadHash:      tezos.MustParsePayloadHash("vh2nZrxixzv4ZjAJn7PRj79GumUMAJzxuEYMjo496TYSaWhXYjZM"),
				PayloadRound:     1,
				ProofOfWorkNonce: asHex("df2ea592260c0100"),
				LbVote:           tezos.FeatureVoteOn,
				AiVote:           tezos.FeatureVoteOn,
			},
		},
		{
			data: asHex("00000533010e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8000000005e9dcbb00242e9bc4583d4f9fa6ba422733f45d3a44397141a953d2237bf8df62e5046eef700000011000000010100000008000000000000000a4c7319284b55068bb7c4e0b9f8585729db7fb27ab4ca9cff2038a1fc324f650c000000000000000000000000000000000000000000000000000000000000000000000000101895ca00000000ff043691f53c02ca1ac6f1a0c1586bf77973e04c2d9b618a8309e79651daf0d55800"),
			head: BlockHeader{
				Level:          1331,
				Proto:          1,
				Predecessor:    tezos.MustParseBlockHash("BKpbfCvh777DQHnXjU2sqHvVUNZ7dBAdqEfKkdw8EGSkD9LSYXb"),
				Timestamp:      asTime("2020-04-20T16:20:00Z"),
				ValidationPass: 2,
				OperationsHash: tezos.MustParseOpListListHash("LLoZqBDX1E2ADRXbmwYo8VtMNeHG6Ygzmm4Zqv97i91UPBQHy9Vq3"),
				Fitness: []tezos.HexBytes{
					asHex("01"),
					asHex("000000000000000a"),
				},
				Context:          tezos.MustParseContextHash("CoVDyf9y9gHfAkPWofBJffo4X4bWjmehH2LeVonDcCKKzyQYwqdk"),
				PayloadHash:      tezos.MustParsePayloadHash("vh1g87ZG6scSYxKhspAUzprQVuLAyoa5qMBKcUfjgnQGnFb3dJcG"),
				PayloadRound:     0,
				ProofOfWorkNonce: asHex("101895ca00000000"),
				SeedNonceHash:    tezos.MustParseNonceHash("nceUFoeQDgkJCmzdMWh19ZjBYqQD3N9fe6bXQ1ZsUKKvMn7iun5Z3"),
				LbVote:           tezos.FeatureVoteOn,
				AiVote:           tezos.FeatureVoteOn,
			},
		},
		{
			data: asHex("0000518e0118425847ac255b6d7c30ce8fec23b8eaf13b741de7d18509ac2ef83c741209630000000061947af504805682ea5d089837764b3efcc90b91db24294ff9ddb66019f332ccba17cc4741000000210000000102000000040000518e0000000000000004ffffffff0000000400000000eb1320a71e8bf8b0162a3ec315461e9153a38b70d00d5dde2df85eb92748f8d068d776e356683a9e23c186ccfb72ddc6c9857bb1704487972922e7c89a7121f800000000a8e1dd3c000000000000"),
			head: BlockHeader{
				Level:          20878,
				Proto:          1,
				Predecessor:    tezos.MustParseBlockHash("BKty19HXfE15jjeLFCTxpEZRXRVkQKGBcArzn4eAgMYTrdaf6xc"),
				Timestamp:      asTime("2021-11-17T03:45:57Z"),
				ValidationPass: 4,
				OperationsHash: tezos.MustParseOpListListHash("LLoaJEEVU5t92V3PEFG9SZ6JrgG3AAwLhKXkXxHjfiZFxLZeqaRcg"),
				Fitness: []tezos.HexBytes{
					asHex("02"),
					asHex("0000518e"),
					asHex(""),
					asHex("ffffffff"),
					asHex("00000000"),
				},
				Context:          tezos.MustParseContextHash("CoWRqXN1hCqPoLNF5K53DkcqHSHA9638oXnyhg5nBBsK1gNVAQdZ"),
				PayloadHash:      tezos.MustParsePayloadHash("vh2UJ9qvkLHcFbiotR462Ni84QU7xJ83fNwspoo9kq7spoNeSMkH"),
				PayloadRound:     0,
				ProofOfWorkNonce: asHex("a8e1dd3c00000000"),
				LbVote:           tezos.FeatureVoteOn,
				AiVote:           tezos.FeatureVoteOn,
			},
		},
		{
			data:  asHex("0000004c013ce881bd760605e86f07bfd8b462eae0f25151ac056978a9abb392e17cba833f0000000065a3e6f304ebbfbd7dd321a66af71edbb7ae270e4f84d4c50f56a412c53f9efdaaefa11516000000210000000102000000040000004c0000000000000004ffffffff0000000400000000081a1078f1729b2ac0da78480aade344e5f5ccfdcd2035bc59563c1ff80897d05677d1d20902581194154052443de6d0b3c5ffffd62bb8e771e2c1ec62217a25000000007769d51b04000000ff650c5a8c733f89717032992c43ac7f3e121a8dd141184353b5e8c76f25706ea90a"),
			chain: tezos.MustParseChainIdHash("NetXdQprcVkpaWU"),
			head: BlockHeader{
				Level:          76,
				Proto:          1,
				Predecessor:    tezos.MustParseBlockHash("BLB79vHaoWiyzYjc68zXWCQFB2snCY28reHR3w6bpvKwZqkZDTE"),
				Timestamp:      asTime("2024-01-14T13:51:47Z"),
				ValidationPass: 4,
				OperationsHash: tezos.MustParseOpListListHash("LLob7XuR6DGQ2jQPurB7AgBGNFi19WukXyuHd1ncjyXGF13qaAZFc"),
				Fitness: []tezos.HexBytes{
					asHex("02"),
					asHex("0000004c"),
					asHex(""),
					asHex("ffffffff"),
					asHex("00000000"),
				},
				Context:          tezos.MustParseContextHash("CoUhsoi3yZqpNGCW1pgu4f7eX2kzbkgKdoekLCny4WtGYyUiH96s"),
				PayloadHash:      tezos.MustParsePayloadHash("vh2LCpkG49XP71LxG7kVc1ob1erR3FnD3jfHjGJa8caN2N7Jn9nx"),
				PayloadRound:     0,
				ProofOfWorkNonce: asHex("7769d51b04000000"),
				SeedNonceHash:    tezos.MustParseNonceHash("nceUzTAhmkcBo1CsZhgPKmt9daAaEWcHspXeeHQqSikEMG7dryrMR"),
				LbVote:           tezos.FeatureVotePass,
				AiVote:           tezos.FeatureVotePass,
			},
			hash: tezos.MustParseBlockHash("BKsWjip13q21S4cRDmrPG85pYJePST41BrFFrRskWoPCecd88fJ"),
			key:  tezos.MustParsePrivateKey("edsk2uqQB9AY4FvioK2YMdfmyMrer5R8mGFyuaLLFfSRo8EoyNdht3"),
			sig:  tezos.MustParseSignature("sigqKNyR7Xuo8TzuMSKA5HaL9XRVmozGM1brMm2ekUSpj14HCTE9zPszEvE6Vy1WEFHhpc4m1wsff4MGkXJQcNmhbALJa7bt"),
		},
	}

	for i, c := range cases {
		// binary decode
		var bh BlockHeader
		if err := bh.UnmarshalBinary(c.data); err != nil {
			t.Errorf("Case %d: decode failed: %v", i, err)
		}

		// json encode
		j2, err := c.head.MarshalJSON()
		if err != nil {
			t.Errorf("Case %d: JSON marshal failed: %v", i, err)
		}

		// compare json encodings
		j1, err := bh.MarshalJSON()
		if err != nil {
			t.Errorf("Case %d: JSON marshal from decoded block failed: %v", i, err)
		}

		if !bytes.Equal(j1, j2) {
			t.Errorf("Case %d: JSON mismatch:\n    1: %s\n    2: %s\n", i,
				string(j1), string(j2),
			)
		}

		// binary encode
		// we're using DefaultParams here, to change use op.WithParams()
		buf := c.head.Bytes()
		if !bytes.Equal(buf, c.data.Bytes()) {
			t.Errorf("Case %d: encode failed:\n    have: %s\n    want: %s\n", i,
				tezos.HexBytes(buf), c.data,
			)
		}

		// check sig
		if c.sig.IsValid() && c.key.IsValid() {
			if err := bh.WithChainId(c.chain).Sign(c.key); err != nil {
				t.Errorf("Case %d: JSON marshal failed: %v", i, err)
			}
			if !bh.Signature.Equal(c.sig) {
				t.Errorf("Case %d: signature mismatch:\n    have: %s\n    want: %s\n", i,
					bh.Signature, c.sig,
				)
			}

			// check hash (needs a valid signature)
			if c.hash.IsValid() {
				if bh.WithChainId(c.chain).Hash() != c.hash {
					t.Errorf("Case %d: hash failed:\n    have: %s\n    want: %s\n", i,
						bh.Hash(), c.hash,
					)
				}
			}
		}
	}
}
