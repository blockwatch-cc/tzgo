// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package codec

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

func asTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Errorf("Parsing time %q: %v", s, err))
	}
	return t
}

func asHex(s string) tezos.HexBytes {
	buf, err := hex.DecodeString(s)
	if err != nil {
		panic(fmt.Errorf("Parsing hex %q: %v", s, err))
	}
	return tezos.HexBytes(buf)
}

func asScript(s string) micheline.Script {
	var sc micheline.Script
	if err := json.Unmarshal([]byte(s), &sc); err != nil {
		panic(fmt.Errorf("Parsing script %q: %v", s, err))
	}
	return sc
}

func asIntPtr(i int64) *tezos.N {
	n := tezos.N(i)
	return &n
}

func TestOp(t *testing.T) {
	type testcase struct {
		name string
		data tezos.HexBytes
		op   Op
	}

	var cases = []testcase{
		// Tenderbake preendorsement
		{
			name: "Tenderbake preendorsement",
			data: asHex("2f50673bab6b20dfb0a88ca93b4a0c72a34c807af5dffbece2cba3d2b509835f14006000000002000000041f1ebb39759cc957216f88fb4d005abc206fb00a53f8d57ac01be00c084cba97"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BL57uk2FrPckCtzBQwaQV1bYtPPShcDCqMShArucaBSpqtmDdRn"),
				Contents: []Operation{
					&TenderbakePreendorsement{
						Slot:             96,
						Level:            2,
						Round:            4,
						BlockPayloadHash: tezos.MustParsePayloadHash("vh1uq2uMDFaJAZZcydX5QeW2dG3Mpc2y31tT621LuEppkxfy11SK"),
					},
				},
			},
		},
		// Tenderbake endorsement
		{
			name: "Tenderbake endorsement",
			data: asHex("fc81eee810737b04018acef4db74d056b79edc43e6be46cae7e4c217c22a82f01500120000518d0000000003e7ea1f67dbb0bb6cfa372cb092cd9cf786b4f1b5e5139da95b915fb95e698d"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BMdVJUZrmcLJBnXsxdJLJaTDFJYyqarwmst7hpPu53Z3xLPtnMF"),
				Contents: []Operation{
					&TenderbakeEndorsement{
						Slot:             18,
						Level:            20877,
						Round:            0,
						BlockPayloadHash: tezos.MustParsePayloadHash("vh1hqtJCryS2Uzb8KDU2PAp33U1nDCeUB4g9yWKTjgVhiy4x9pQA"),
					},
				},
			},
		},
		// Tenderbake double endorsement evidence
		{
			name: "Tenderbake double endorsement evidence",
			data: asHex("fc81eee810737b04018acef4db74d056b79edc43e6be46cae7e4c217c22a82f0020000008ba60703a9567bf69ec66b368c3d8562eba4cbf29278c2c10447a684e3aa1436851500120000518d0000000003e7ea1f67dbb0bb6cfa372cb092cd9cf786b4f1b5e5139da95b915fb95e698dd3a9e1467b32104921d4e2dd93265739c1a5faee7a7f8880842b096c0b6714200c43fd5872f82581dfe1cb3a76ccdadaa4d6361d72b4abee6884cb7ed87f0b040000008ba60703a9567bf69ec66b368c3d8562eba4cbf29278c2c10447a684e3aa1436851500120000518d0000000003e7ea1f67dbb0bb6cfa372cb092cd9cf786b4f1b5e5139da95b915fb95e698dd3a9e1467b32104921d4e2dd93265739c1a5faee7a7f8880842b096c0b6714200c43fd5872f82581dfe1cb3a76ccdadaa4d6361d72b4abee6884cb7ed87f0b04"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BMdVJUZrmcLJBnXsxdJLJaTDFJYyqarwmst7hpPu53Z3xLPtnMF"),
				Contents: []Operation{
					&TenderbakeDoubleEndorsementEvidence{
						Op1: TenderbakeInlinedEndorsement{
							Branch: tezos.MustParseBlockHash("BLyQHMFeNzZEKHmKgfD9imcowLm8hc4aUo16QtYZcS5yvx7RFqQ"),
							Endorsement: TenderbakeEndorsement{
								Slot:             18,
								Level:            20877,
								Round:            0,
								BlockPayloadHash: tezos.MustParsePayloadHash("vh1hqtJCryS2Uzb8KDU2PAp33U1nDCeUB4g9yWKTjgVhiy4x9pQA"),
							},
							Signature: tezos.MustParseSignature("sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"),
						},
						Op2: TenderbakeInlinedEndorsement{
							Branch: tezos.MustParseBlockHash("BLyQHMFeNzZEKHmKgfD9imcowLm8hc4aUo16QtYZcS5yvx7RFqQ"),
							Endorsement: TenderbakeEndorsement{
								Slot:             18,
								Level:            20877,
								Round:            0,
								BlockPayloadHash: tezos.MustParsePayloadHash("vh1hqtJCryS2Uzb8KDU2PAp33U1nDCeUB4g9yWKTjgVhiy4x9pQA"),
							},
							Signature: tezos.MustParseSignature("sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"),
						},
					},
				},
			},
		},
		// Tenderbake double preendorsement evidence
		{
			name: "Tenderbake double preendorsement evidence",
			data: asHex("fc81eee810737b04018acef4db74d056b79edc43e6be46cae7e4c217c22a82f0070000008ba60703a9567bf69ec66b368c3d8562eba4cbf29278c2c10447a684e3aa1436851400120000518d0000000003e7ea1f67dbb0bb6cfa372cb092cd9cf786b4f1b5e5139da95b915fb95e698dd3a9e1467b32104921d4e2dd93265739c1a5faee7a7f8880842b096c0b6714200c43fd5872f82581dfe1cb3a76ccdadaa4d6361d72b4abee6884cb7ed87f0b040000008ba60703a9567bf69ec66b368c3d8562eba4cbf29278c2c10447a684e3aa1436851400120000518d0000000003e7ea1f67dbb0bb6cfa372cb092cd9cf786b4f1b5e5139da95b915fb95e698dd3a9e1467b32104921d4e2dd93265739c1a5faee7a7f8880842b096c0b6714200c43fd5872f82581dfe1cb3a76ccdadaa4d6361d72b4abee6884cb7ed87f0b04"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BMdVJUZrmcLJBnXsxdJLJaTDFJYyqarwmst7hpPu53Z3xLPtnMF"),
				Contents: []Operation{
					&TenderbakeDoublePreendorsementEvidence{
						Op1: TenderbakeInlinedPreendorsement{
							Branch: tezos.MustParseBlockHash("BLyQHMFeNzZEKHmKgfD9imcowLm8hc4aUo16QtYZcS5yvx7RFqQ"),
							Endorsement: TenderbakePreendorsement{
								Slot:             18,
								Level:            20877,
								Round:            0,
								BlockPayloadHash: tezos.MustParsePayloadHash("vh1hqtJCryS2Uzb8KDU2PAp33U1nDCeUB4g9yWKTjgVhiy4x9pQA"),
							},
							Signature: tezos.MustParseSignature("sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"),
						},
						Op2: TenderbakeInlinedPreendorsement{
							Branch: tezos.MustParseBlockHash("BLyQHMFeNzZEKHmKgfD9imcowLm8hc4aUo16QtYZcS5yvx7RFqQ"),
							Endorsement: TenderbakePreendorsement{
								Slot:             18,
								Level:            20877,
								Round:            0,
								BlockPayloadHash: tezos.MustParsePayloadHash("vh1hqtJCryS2Uzb8KDU2PAp33U1nDCeUB4g9yWKTjgVhiy4x9pQA"),
							},
							Signature: tezos.MustParseSignature("sigqgQgW5qQCsuHP5HhMhAYR2HjcChUE7zAczsyCdF681rfZXpxnXFHu3E6ycmz4pQahjvu3VLfa7FMCxZXmiMiuZFQS4MHy"),
						},
					},
				},
			},
		},
		// Tenderbake double baking evidence
		{
			name: "Tenderbake double baking evidence",
			data: asHex("fc81eee810737b04018acef4db74d056b79edc43e6be46cae7e4c217c22a82f003000001010004ad5702d21acd0569ff8e03cd564fdc15baae8e436b141510f4ca966bdadfe092904359000000006242e26904cf318e718893b9efb0a426130f6d8fac752db1c47a98d0c3f89780ec8b1a4740000000210000000102000000040004ad570000000000000004ffffffff00000004000000016ae0589f63d96d15d6b41b4e9a9c6f5670ae7e4a3495ffdaf0fa651a10b9e25d9253ed831d88bb031de4f49e43d62977864806a7b0945e8877b030150f2ae63b00000001df2ea592260c01000000517c25c5845f9694eae582055b16ecd9805b318c627d1645f0a4dbf8bf51f4fa51bf5ed45b7e0e1bf64e9fced0ccb96125a22532214d3cbedc745f16b94e0e45000001010004ad5702d21acd0569ff8e03cd564fdc15baae8e436b141510f4ca966bdadfe092904359000000006242e26904cf318e718893b9efb0a426130f6d8fac752db1c47a98d0c3f89780ec8b1a4740000000210000000102000000040004ad570000000000000004ffffffff00000004000000016ae0589f63d96d15d6b41b4e9a9c6f5670ae7e4a3495ffdaf0fa651a10b9e25d9253ed831d88bb031de4f49e43d62977864806a7b0945e8877b030150f2ae63b00000001df2ea592260c01000000c5fa33a8748fe231310655dd03d2543473856ef5beec03bf10030fed2f7d86a7d79c71a3e0a5814da2337865f3bfd307d4b6e7f0e69e9546b341c4109fc342e9"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BMdVJUZrmcLJBnXsxdJLJaTDFJYyqarwmst7hpPu53Z3xLPtnMF"),
				Contents: []Operation{
					&DoubleBakingEvidence{
						Bh1: BlockHeader{
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
							Signature:        tezos.MustParseSignature("sigYec9pbutMj4sxHxGhmQeoU62K96Xbdr8MZJE4XG7PkcKmUGsQMKwpegwdubccUXshdCHukUxDodvaCjpQQaDjagW43YeW"),
						},
						Bh2: BlockHeader{
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
							Signature:        tezos.MustParseSignature("sigotZGfNkiFwpditQfPjQ6DpN5QnAo6gjFjAMTdU9ATCcoUyugBtw2p6dqJmvNSETzqN2hTaKJytZJh2abMJ7S49AhX8n13"),
						},
					},
				},
			},
		},
		// seed nonce
		{
			name: "seed nonce",
			data: asHex("2bf383405f10841ea2a1180af0190a8612916c2d12c01dbcf25c415ded192105010004e000fcf77031019bf38c6edd3e360154b6160258df5b21e158f79a03576d67a284f7"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BL3e1ZpSS6s65BMZDkGnP7kVFuCAA5qfVeSnUmQiDn9NFKGDgNd"),
				Contents: []Operation{
					&SeedNonceRevelation{
						Level: 319488,
						Nonce: asHex("fcf77031019bf38c6edd3e360154b6160258df5b21e158f79a03576d67a284f7"),
					},
				},
			},
		},

		// proposal
		{
			name: "proposal",
			data: asHex("2bf383405f10841ea2a1180af0190a8612916c2d12c01dbcf25c415ded19210505002cca28ad0529681a2cc52e360ff1b4c1d67d7e600000000f00000020ce5f061e34b5a21feab8dbdfe755ef17e70c9f565464f067ac5e7c02be830a48"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BL3e1ZpSS6s65BMZDkGnP7kVFuCAA5qfVeSnUmQiDn9NFKGDgNd"),
				Contents: []Operation{
					&Proposals{
						Source: tezos.MustParseAddress("tz1PirbogVqfmBT9XCuYJ1KnDx4bnMSYfGru"),
						Period: 15,
						Proposals: []tezos.ProtocolHash{
							tezos.MustParseProtocolHash("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx"),
						},
					},
				},
			},
		},

		// ballot
		{
			name: "ballot",
			data: asHex("2bf383405f10841ea2a1180af0190a8612916c2d12c01dbcf25c415ded19210506002cca28ad0529681a2cc52e360ff1b4c1d67d7e600000000fce5f061e34b5a21feab8dbdfe755ef17e70c9f565464f067ac5e7c02be830a4800"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BL3e1ZpSS6s65BMZDkGnP7kVFuCAA5qfVeSnUmQiDn9NFKGDgNd"),
				Contents: []Operation{
					&Ballot{
						Source:   tezos.MustParseAddress("tz1PirbogVqfmBT9XCuYJ1KnDx4bnMSYfGru"),
						Period:   15,
						Proposal: tezos.MustParseProtocolHash("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx"),
						Ballot:   tezos.BallotVoteYay,
					},
				},
			},
		},

		// activate
		{
			name: "activate",
			data: asHex("2bf383405f10841ea2a1180af0190a8612916c2d12c01dbcf25c415ded19210504c118171a1334d71e988f7cc63f0e2d0e3d025c622cd088062ac8665178c8379fac2d4ac52b9357a3"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BL3e1ZpSS6s65BMZDkGnP7kVFuCAA5qfVeSnUmQiDn9NFKGDgNd"),
				Contents: []Operation{
					&ActivateAccount{
						PublicKeyHash: tezos.MustParseAddress("tz1dF1xxjjb5FGuogUBb9ti8xqF3n3Jzd9uv"),
						Secret:        asHex("2cd088062ac8665178c8379fac2d4ac52b9357a3"),
					},
				},
			},
		},
		// reveal
		{
			name: "reveal",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6b005c7886828ec2a24f1814484de7dd53e559831c3fe807c197b001e8070000654b5b22880736d33865b4f30367e90feb81b17cc0ceb7ac951a0066142d5847"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&Reveal{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886593,
							GasLimit:     1000,
							StorageLimit: 0,
						},
						PublicKey: tezos.MustParseKey("edpkuQqN9HB3jY1FvDzt15WQDVSHR4vQGd1wv6iqJ73wkrKecRtnXh"),
					},
				},
			},
		},

		// transaction
		{
			name: "transaction",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6c005c7886828ec2a24f1814484de7dd53e559831c3fe807c297b0018c0b00c0843d00002cca28ad0529681a2cc52e360ff1b4c1d67d7e6000"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&Transaction{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886594,
							GasLimit:     1420,
							StorageLimit: 0,
						},
						Destination: tezos.MustParseAddress("tz1PirbogVqfmBT9XCuYJ1KnDx4bnMSYfGru"),
						Amount:      1000000,
					},
				},
			},
		},

		// origination
		{
			name: "origination",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6d005c7886828ec2a24f1814484de7dd53e559831c3fef0ac297b0018157c30280c2d72f000000001c02000000170500036805010368050202000000080316053d036d03420000000a010000000568656c6c6f"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&Origination{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1391,
							Counter:      2886594,
							GasLimit:     11137,
							StorageLimit: 323,
						},
						Balance: 100000000,
						Script:  asScript(`{"code": [{"args": [{"prim": "string"}],"prim": "parameter"},{"args": [{"prim": "string"}],"prim": "storage"},{"args": [[{"prim": "CAR"},{"args": [{"prim": "operation"}],"prim": "NIL"},{"prim": "PAIR"}]],"prim": "code"}],"storage": {"string": "hello"}}`),
					},
				},
			},
		},

		// delegation
		{
			name: "delegation",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6e005c7886828ec2a24f1814484de7dd53e559831c3fe807c297b001e80700ff002cca28ad0529681a2cc52e360ff1b4c1d67d7e60"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&Delegation{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886594,
							GasLimit:     1000,
							StorageLimit: 0,
						},
						Delegate: tezos.MustParseAddress("tz1PirbogVqfmBT9XCuYJ1KnDx4bnMSYfGru"),
					},
				},
			},
		},

		// delegation withdraw
		{
			name: "delegation withdraw",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6e005c7886828ec2a24f1814484de7dd53e559831c3fe807c297b001e8070000"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&Delegation{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886594,
							GasLimit:     1000,
							StorageLimit: 0,
						},
					},
				},
			},
		},

		// delegation baker registration
		{
			name: "delegation baker registration",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6e005c7886828ec2a24f1814484de7dd53e559831c3fe807c297b001e80700ff005c7886828ec2a24f1814484de7dd53e559831c3f"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&Delegation{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886594,
							GasLimit:     1000,
							StorageLimit: 0,
						},
						Delegate: tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
					},
				},
			},
		},

		// global constant
		{
			name: "global constant",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c6f005c7886828ec2a24f1814484de7dd53e559831c3fe807c297b001a08d066400000011010000000c68656c6c6f20776f726c6421"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&RegisterGlobalConstant{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886594,
							GasLimit:     100000,
							StorageLimit: 100,
						},
						Value: micheline.NewString("hello world!"),
					},
				},
			},
		},

		// set deposits limit
		{
			name: "set deposits limit",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c70005c7886828ec2a24f1814484de7dd53e559831c3fe807c297b001a08d0664ff64"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&SetDepositsLimit{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886594,
							GasLimit:     100000,
							StorageLimit: 100,
						},
						Limit: asIntPtr(100),
					},
				},
			},
		},

		// clear deposits limit
		{
			name: "clear deposits limit",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c70005c7886828ec2a24f1814484de7dd53e559831c3fe807c297b001a08d066400"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&SetDepositsLimit{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1U4yF2Bkd7hV2JHW2styAWPif12TUCyS2S"),
							Fee:          1000,
							Counter:      2886594,
							GasLimit:     100000,
							StorageLimit: 100,
						},
					},
				},
			},
		},

		// failing noop
		{
			name: "failing noop",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c110000000c48656c6c6f20576f726c6421"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&FailingNoop{
						Arbitrary: "Hello World!",
					},
				},
			},
		},

		// transfer_ticket
		{
			name: "transfer_ticket",
			data: asHex("09af86395fee09cfbede6b11339cd53216aeee93c38b9bf5cee4c791b814df8c9e00fdf904a319c1fb0f073cd2ebc7c0ab71466a1781c306f5b30f82225600000012010000000d74686972642d6465706f736974000000020368013f4a259911e55e00ad15e1b23cacc020dd853bcc0001013f4a259911e55e00ad15e1b23cacc020dd853bcc0000000003787878"),
			op: Op{
				Branch: tezos.MustParseBlockHash("BKnYk1T5a49bb8me4WfQeugyFnMEH9h8cm6jqvL3BxRwE23EVBJ"),
				Contents: []Operation{
					&TransferTicket{
						Manager: Manager{
							Source:       tezos.MustParseAddress("tz1inuxjXxKhd9e4b97N1Wgz7DwmZSxFcDpM"),
							Fee:          835,
							Counter:      252405,
							GasLimit:     4354,
							StorageLimit: 86,
						},
						Contents:    micheline.NewString("third-deposit"),
						Type:        micheline.NewPrim(micheline.T_STRING),
						Ticketer:    tezos.MustParseAddress("KT1EMQxfYVvhTJTqMiVs2ho2dqjbYfYKk6BY"),
						Amount:      tezos.NewN(1),
						Destination: tezos.MustParseAddress("KT1EMQxfYVvhTJTqMiVs2ho2dqjbYfYKk6BY"),
						Entrypoint:  "xxx",
					},
				},
			},
		},
	}

	for _, c := range cases {
		// binary decode
		o, err := DecodeOp(c.data)
		if err != nil {
			t.Errorf("%q: decode failed: %v", c.name, err)
		}

		// json encode
		j2, err := c.op.MarshalJSON()
		if err != nil {
			t.Errorf("%q: JSON marshal failed: %v", c.name, err)
		}

		// compare json encodings
		if o != nil {
			j1, err := o.MarshalJSON()
			if err != nil {
				t.Errorf("%q: JSON marshal from decoded op failed: %v", c.name, err)
			}

			if !bytes.Equal(j1, j2) {
				t.Errorf("%q: JSON mismatch:\n    1: %s\n    2: %s\n", c.name,
					string(j1), string(j2),
				)
			}
		}

		// binary encode
		// we're using DefaultParams here, to change use op.WithParams()
		buf := c.op.Bytes()
		if !bytes.Equal(buf, c.data.Bytes()) {
			t.Errorf("%q: encode failed:\n    have: %s\n    want: %s\n", c.name,
				tezos.HexBytes(buf), c.data,
			)
		}
	}
}
