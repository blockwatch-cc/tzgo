// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

var (
	ProtoAlpha     = MustParseProtocolHash("ProtoALphaALphaALphaALphaALphaALphaALphaALphaDdp3zK")
	ProtoGenesis   = MustParseProtocolHash("PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i")
	ProtoBootstrap = MustParseProtocolHash("Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P")
	ProtoV001      = MustParseProtocolHash("PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY")
	ProtoV002      = MustParseProtocolHash("PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt")
	ProtoV003      = MustParseProtocolHash("PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP")
	ProtoV004      = MustParseProtocolHash("Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd")
	ProtoV005_2    = MustParseProtocolHash("PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS")
	ProtoV006_2    = MustParseProtocolHash("PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb")
	ProtoV007      = MustParseProtocolHash("PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo")
	ProtoV008_2    = MustParseProtocolHash("PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA")
	ProtoV009      = MustParseProtocolHash("PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i")
	ProtoV010      = MustParseProtocolHash("PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV")
	ProtoV011_2    = MustParseProtocolHash("PtHangz2aRngywmSRGGvrcTyMbbdpWdpFKuS4uMWxg2RaH9i1qx")
	ProtoV012_2    = MustParseProtocolHash("Psithaca2MLRFYargivpo7YvUr7wUDqyxrdhC5CQq78mRvimz6A")
	ProtoV013_2    = MustParseProtocolHash("PtJakart2xVj7pYXJBXrqHgd82rdkLey5ZeeGwDgPp9rhQUbSqY")
	ProtoV014      = MustParseProtocolHash("PtKathmankSpLLDALzWw7CGD2j2MtyveTwboEYokqUCP4a1LxMg")
	ProtoV015      = MustParseProtocolHash("PtLimaPtLMwfNinJi9rCfDPWea8dFgTZ1MeJ9f1m2SRic6ayiwW")
	ProtoV016      = MustParseProtocolHash("PtMumbaiiFFEGbew1rRjzSPyzRbA51Tm3RVZL5suHPxSZYDhCEc")
	ProtoV016_2    = MustParseProtocolHash("PtMumbai2TmsJHNGRkD8v8YDbtao7BLUC3wjASn1inAKLFCjaH1")

	// aliases
	PtAthens = ProtoV004
	PsBabyM1 = ProtoV005_2
	PsCARTHA = ProtoV006_2
	PsDELPH1 = ProtoV007
	PtEdo2Zk = ProtoV008_2
	PsFLoren = ProtoV009
	PtGRANAD = ProtoV010
	PtHangz2 = ProtoV011_2
	Psithaca = ProtoV012_2
	PtJakart = ProtoV013_2
	PtKathma = ProtoV014
	PtLimaPt = ProtoV015
	PtMumbai = ProtoV016_2

	Mainnet      = MustParseChainIdHash("NetXdQprcVkpaWU")
	Ghostnet     = MustParseChainIdHash("NetXnHfVqm9iesp")
	Jakartanet   = MustParseChainIdHash("NetXLH1uAxK7CCh")
	Kathmandunet = MustParseChainIdHash("NetXi2ZagzEsXbZ")
	Limanet      = MustParseChainIdHash("NetXizpkH94bocH")
	Mumbainet    = MustParseChainIdHash("NetXgbcrNtXD2yA")

	Versions = map[ProtocolHash]int{
		ProtoGenesis:   0,
		ProtoBootstrap: 0,
		ProtoV001:      1,
		ProtoV002:      2,
		ProtoV003:      3,
		ProtoV004:      4,
		ProtoV005_2:    5,
		ProtoV006_2:    6,
		ProtoV007:      7,
		ProtoV008_2:    8,
		ProtoV009:      9,
		ProtoV010:      10,
		ProtoV011_2:    11,
		ProtoV012_2:    12,
		ProtoV013_2:    13,
		ProtoV014:      14,
		ProtoV015:      15,
		ProtoV016:      16,
		ProtoV016_2:    16,
		ProtoAlpha:     17,
	}
	ProtocolVersions = map[uint32][]ProtocolHash{
		Mainnet.Uint32(): {
			ProtoGenesis,   // -1
			ProtoBootstrap, // 0
			ProtoV001,      // 1
			ProtoV002,      // 2
			ProtoV003,      // 3
			ProtoV004,      // 4
			ProtoV005_2,    // 5
			ProtoV006_2,    // 6
			ProtoV007,      // 7
			ProtoV008_2,    // 8
			ProtoV009,      // 9
			ProtoV010,      // 10
			ProtoV011_2,    // 11
			ProtoV012_2,    // 12
			ProtoV013_2,    // 13
			ProtoV014,      // 14
			ProtoV015,      // 15
			ProtoV016,
			ProtoV016_2,
			ProtoAlpha,
		},
		Jakartanet.Uint32(): {
			ProtoGenesis,   // -1
			ProtoBootstrap, // 0
			ProtoV012_2,    // 1
			ProtoV013_2,    // 2
		},
		Kathmandunet.Uint32(): {
			ProtoGenesis,   // -1
			ProtoBootstrap, // 0
			ProtoV013_2,    // 1
			ProtoV014,      // 2
		},
		Limanet.Uint32(): {
			ProtoGenesis,   // -1
			ProtoBootstrap, // 0
			ProtoV014,      // 1
			ProtoV015,      // 2
		},
		Ghostnet.Uint32(): {
			ProtoGenesis,   // -1
			ProtoBootstrap, // 0
			ProtoV011_2,    // 1
			ProtoV012_2,    // 2
			ProtoV013_2,    // 3
			ProtoV014,      // 4
			ProtoV015,      // 5
		},
	}
)

func (p *Params) ForNetwork(net ChainIdHash) *Params {
	pp := &Params{}
	*pp = *p
	pp.ChainId = net
	switch {
	case Mainnet.Equal(net):
		pp.Network = "Mainnet"
	case Ghostnet.Equal(net):
		pp.Network = "Ghostnet"
		pp.Version = 11 // starts at Hangzhou
	case Jakartanet.Equal(net):
		pp.Network = "Jakartanet"
		pp.Version = 12 // starts at Ithaca
	case Kathmandunet.Equal(net):
		pp.Network = "Kathmandunet"
		pp.Version = 13 // starts at Jakarta
	case Limanet.Equal(net):
		pp.Network = "Limanet"
		pp.Version = 14 // starts at Katmandu
	default:
		pp.Network = "Sandbox"
	}
	return pp
}

func (p *Params) ForProtocol(proto ProtocolHash) *Params {
	pp := &Params{}
	*pp = *p
	pp.Protocol = proto
	pp.NumVotingPeriods = 4
	pp.MaxOperationsTTL = 60
	switch {
	case ProtoGenesis.Equal(proto):
		// retain version set in ForNetwork()
		pp.StartHeight = 0
		pp.EndHeight = 0

	case ProtoBootstrap.Equal(proto):
		// retain version set in ForNetwork()
		pp.StartHeight = 1
		pp.EndHeight = 1

	case ProtoV001.Equal(proto):
		pp.Version = 1
		pp.StartHeight = 2
		pp.EndHeight = 28082

	case ProtoV002.Equal(proto):
		pp.Version = 2
		pp.StartHeight = 28083
		pp.EndHeight = 204761

	case ProtoV003.Equal(proto):
		pp.Version = 3
		pp.StartHeight = 204762
		pp.EndHeight = 458752

	case ProtoV004.Equal(proto): // Athens
		pp.Version = 4
		pp.StartHeight = 458753
		pp.EndHeight = 655360

	case PsBabyM1.Equal(proto): // Babylon
		pp.Version = 5
		pp.OperationTagsVersion = 1
		pp.StartHeight = 655361
		pp.EndHeight = 851968

	case PsCARTHA.Equal(proto): // Carthage
		pp.Version = 6
		pp.OperationTagsVersion = 1
		pp.StartHeight = 851969
		pp.EndHeight = 1212416

	case PsDELPH1.Equal(proto): // Delphi
		pp.Version = 7
		pp.OperationTagsVersion = 1
		// this is extremely hacky!
		pp.StartBlockOffset = 0
		pp.StartCycle = 0
		pp.BlocksPerCycle = 4096
		pp.BlocksPerCommitment = 32
		pp.BlocksPerSnapshot = 256
		pp.BlocksPerVotingPeriod = 32768
		pp.EndorsersPerBlock = 32
		pp.StartHeight = 1212417
		pp.EndHeight = 1343488

	case PtEdo2Zk.Equal(proto): // Edo
		pp.Version = 8
		pp.OperationTagsVersion = 1
		pp.NumVotingPeriods = 5

		// OK, bear with me. If you are from the future, forgive the Tezos
		// core devs.
		//
		// Apparently Tezos can switch cycle lengths which messes up all
		// kinds of calculations big-time. They have also introduced a bug
		// into vote block alignment during the switch from Delphi to Edo
		// (Granada is supposed to fix this). We use an extra offset for
		// tracking this bug. Note that it only starts to appear from
		// cycle end of the first Edo vote epoch.
		//
		// Edo has also added a 5th vote period and decreased period
		// durations from 32,768 blocks to 20,480 so we need a custom start
		// offset to keep our vote start/end calculations correct.
		//
		pp.StartBlockOffset = 1343488
		pp.StartCycle = 328
		pp.VoteBlockOffset = 1
		// this is extremely hacky!
		pp.BlocksPerCycle = 4096
		pp.BlocksPerCommitment = 32
		pp.BlocksPerSnapshot = 256
		pp.BlocksPerVotingPeriod = 20480
		pp.EndorsersPerBlock = 32
		pp.StartHeight = 1343489
		pp.EndHeight = 1466367

	case PsFLoren.Equal(proto): // Florence
		pp.Version = 9
		pp.OperationTagsVersion = 1
		pp.NumVotingPeriods = 5
		pp.StartBlockOffset = 1466368
		pp.StartCycle = 358
		pp.VoteBlockOffset = 1 // same as Edo (!!)
		// FIXME: this is extremely hacky!
		pp.BlocksPerCycle = 4096
		pp.BlocksPerCommitment = 32
		pp.BlocksPerSnapshot = 256
		pp.BlocksPerVotingPeriod = 20480
		pp.StartHeight = 1466368
		pp.EndHeight = 1589247

	case PtGRANAD.Equal(proto): // Granada
		pp.Version = 10
		pp.OperationTagsVersion = 1
		pp.NumVotingPeriods = 5
		pp.MaxOperationsTTL = 120

		// It gets more fun in Granada. Now the major cycle length has doubled
		// since block times have halved. This requires a second offset to track
		// start cycle in addition to start block offset.
		//
		// In an attempt to fix vote/cycle alignment Granada offsets
		// voting period by +1 on activation which apparently fails on
		// Granadanet or any other network that is not mainnet because the
		// problem did not exist there. Anyways, vote start and cycle start
		// should be the same again, but there is one problem:
		//
		// The last Florence vote epoch ends on 1,589,247 (due to the Edo bug),
		// one block short of cycle end like all epochs since Edo started.
		// Since Granada will activate at block 1,589,248 == at cycle end
		// of the last Florence cycle, this block will also become the first
		// voting block in Granada. A great start for re-alignement, isn't it!
		// The sane choice would have been to skip one block and let vote start
		// at cycle start 1,589,249. However, in tyical Tezos manner, things
		// have to be more complicated and instead we make the first voting epoch
		// in Granada 1 block longer (on top of that all RPC voting counters
		// are broken).
		// https://tezos.gitlab.io/protocols/010_granada.html#bogus-rpc-results
		//
		// TzGo will NOT implement this brainfuck exception and instead do
		// the following:
		//
		// Block      Proto     Cycle Start   Cycle End   Vote Start   Vote End
		// --------------------------------------------------------------------
		// 1,589,247  Florence                                            x
		// 1,589,248  Granada                    x                        x
		// 1,589,249  Granada       x                          x
		//
		pp.StartBlockOffset = 1589248
		pp.StartCycle = 388
		pp.VoteBlockOffset = 0
		// FIXME: this is extremely hacky!
		pp.BlocksPerCycle = 8192
		pp.BlocksPerCommitment = 64
		pp.BlocksPerSnapshot = 512
		pp.BlocksPerVotingPeriod = 40960
		pp.EndorsersPerBlock = 256
		pp.StartHeight = 1589248
		pp.EndHeight = 1916928

	case PtHangz2.Equal(proto): // Hangzhou
		pp.Version = 11
		pp.OperationTagsVersion = 1
		pp.NumVotingPeriods = 5
		pp.MaxOperationsTTL = 120
		if Mainnet.Equal(p.ChainId) {
			pp.StartBlockOffset = 1916928
			pp.StartCycle = 428
			pp.VoteBlockOffset = 0
			// FIXME: this is extremely hacky!
			pp.BlocksPerCycle = 8192
			pp.BlocksPerCommitment = 64
			pp.BlocksPerSnapshot = 512
			pp.BlocksPerVotingPeriod = 40960
			pp.EndorsersPerBlock = 256
			pp.StartHeight = 1916929
			pp.EndHeight = 2244608
		}
	case Psithaca.Equal(proto): // Ithaca
		pp.Version = 12
		pp.OperationTagsVersion = 2
		pp.NumVotingPeriods = 5
		pp.MaxOperationsTTL = 120
		switch {
		case Mainnet.Equal(p.ChainId):
			pp.StartBlockOffset = 2244608
			pp.StartCycle = 468
			pp.VoteBlockOffset = 0
			// FIXME: this is extremely hacky!
			pp.BlocksPerCycle = 8192
			pp.BlocksPerCommitment = 64
			pp.BlocksPerSnapshot = 512
			pp.BlocksPerVotingPeriod = 40960
			pp.EndorsersPerBlock = 0
			pp.StartHeight = 2244609
			pp.EndHeight = -1
		case Jakartanet.Equal(p.ChainId):
			pp.StartBlockOffset = 0
			pp.StartCycle = 0
			pp.StartHeight = 2
			pp.EndHeight = 8192
		case Ghostnet.Equal(p.ChainId):
			pp.StartBlockOffset = 8192
			pp.StartCycle = 2     // block 8192 is not in c2, but whatever
			pp.StartHeight = 8192 // off by one due to Tezos testnet fuckup !!
			pp.EndHeight = 765952
		}
	case PtJakart.Equal(proto): // Jakarta
		pp.Version = 13
		pp.OperationTagsVersion = 2
		pp.NumVotingPeriods = 5
		pp.MaxOperationsTTL = 120
		switch {
		case Mainnet.Equal(p.ChainId):
			pp.StartBlockOffset = 2490368
			pp.StartCycle = 498
			pp.VoteBlockOffset = 0
			// FIXME: this is extremely hacky!
			pp.BlocksPerCycle = 8192
			pp.BlocksPerCommitment = 64
			pp.BlocksPerSnapshot = 512
			pp.BlocksPerVotingPeriod = 40960
			pp.EndorsersPerBlock = 0
			pp.StartHeight = 2490369
			pp.EndHeight = 2736128
		case Kathmandunet.Equal(p.ChainId):
			pp.StartBlockOffset = 0
			pp.StartCycle = 0
			pp.StartHeight = 2
			pp.EndHeight = 8192
		case Jakartanet.Equal(p.ChainId):
			pp.StartBlockOffset = 8192
			pp.StartCycle = 2
			pp.StartHeight = 8193
			pp.EndHeight = -1
		case Ghostnet.Equal(p.ChainId):
			pp.StartBlockOffset = 765952
			pp.StartCycle = 187
			pp.StartHeight = 765953
			pp.EndHeight = 1191936
		}
	case PtKathma.Equal(proto): // Kathmandu
		pp.Version = 14
		pp.OperationTagsVersion = 2
		pp.NumVotingPeriods = 5
		pp.MaxOperationsTTL = 120
		switch {
		case Mainnet.Equal(p.ChainId):
			pp.StartBlockOffset = 2736128
			pp.StartCycle = 528
			pp.VoteBlockOffset = 0
			// FIXME: this is extremely hacky!
			pp.BlocksPerCycle = 8192
			pp.BlocksPerCommitment = 64
			pp.BlocksPerSnapshot = 512
			pp.BlocksPerVotingPeriod = 40960
			pp.EndorsersPerBlock = 0
			pp.StartHeight = 2736129
			pp.EndHeight = 2981888
		case Kathmandunet.Equal(p.ChainId):
			pp.StartBlockOffset = 8192
			pp.StartCycle = 2
			pp.StartHeight = 8193
			pp.EndHeight = -1
		case Limanet.Equal(p.ChainId):
			pp.StartBlockOffset = 0
			pp.StartCycle = 0
			pp.StartHeight = 2
			pp.EndHeight = 8192
		case Ghostnet.Equal(p.ChainId):
			pp.StartBlockOffset = 1191936
			pp.StartCycle = 291
			pp.StartHeight = 1191937
			pp.EndHeight = 1654784
		}
	case PtLimaPt.Equal(proto): // Lima
		pp.Version = 15
		pp.OperationTagsVersion = 2
		pp.NumVotingPeriods = 5
		pp.MaxOperationsTTL = 120
		switch {
		case Mainnet.Equal(p.ChainId):
			pp.StartBlockOffset = 2981888
			pp.StartCycle = 558
			pp.VoteBlockOffset = 0
			// FIXME: this is extremely hacky!
			pp.BlocksPerCycle = 8192
			pp.BlocksPerCommitment = 64
			pp.BlocksPerSnapshot = 512
			pp.BlocksPerVotingPeriod = 40960
			pp.EndorsersPerBlock = 0
			pp.StartHeight = 2981889
			pp.EndHeight = 3268608
		case Limanet.Equal(p.ChainId):
			pp.StartBlockOffset = 8192
			pp.StartCycle = 2
			pp.StartHeight = 8193
			pp.EndHeight = -1
		case Ghostnet.Equal(p.ChainId):
			pp.StartBlockOffset = 1654784
			pp.StartCycle = 404
			pp.StartHeight = 1654785
			pp.EndHeight = 3268608
		}
	case PtMumbai.Equal(proto): // Mumbai
		pp.Version = 16
		pp.OperationTagsVersion = 2
		pp.NumVotingPeriods = 5
		pp.MaxOperationsTTL = 120
		switch {
		case Mainnet.Equal(p.ChainId):
			pp.StartBlockOffset = 3268608
			pp.StartCycle = 593
			pp.VoteBlockOffset = 0
			// FIXME: this is extremely hacky!
			// Block time reduced to 15s now so we double the block per cycle
			pp.BlocksPerCycle = 16384
			pp.BlocksPerCommitment = 128
			pp.BlocksPerSnapshot = 1024
			pp.BlocksPerVotingPeriod = 81920
			pp.EndorsersPerBlock = 0
			pp.StartHeight = 3268609
			pp.EndHeight = -1
		case Limanet.Equal(p.ChainId), Ghostnet.Equal(p.ChainId), Mumbainet.Equal(p.ChainId):
			panic("not implemented")
		}
	}
	return pp
}

func (p Params) Clean() *Params {
	pp := p
	pp.OperationTagsVersion = 0
	pp.NumVotingPeriods = 0
	pp.StartBlockOffset = 0
	pp.StartCycle = 0
	pp.VoteBlockOffset = 0
	pp.StartHeight = -1
	pp.EndHeight = -1
	return &pp
}

func (p *Params) ForHeight(h int64) *Params {
	versions, ok := ProtocolVersions[p.ChainId.Uint32()]
	if !ok {
		return p
	}
	pp := p.Clean()
	for i := len(versions) - 1; i >= 0; i-- {
		pp = pp.Clean().ForNetwork(p.ChainId).ForProtocol(versions[i])
		if uint64(h-pp.StartHeight) < uint64(pp.EndHeight-pp.StartHeight+1) {
			return pp
		}
	}
	return p
}

func (p *Params) ForCycle(c int64) *Params {
	versions, ok := ProtocolVersions[p.ChainId.Uint32()]
	if !ok {
		return p
	}
	pp := p.Clean()
	for i := len(versions) - 1; i >= 0; i-- {
		pp = pp.Clean().ForNetwork(p.ChainId).ForProtocol(versions[i])
		if pp.StartCycle == 0 || pp.StartCycle <= c {
			return pp
		}
	}
	return p
}
