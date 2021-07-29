// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

var (
	ProtoGenesis   = ParseProtocolHashSafe("PrihK96nBAFSxVL1GLJTVhu9YnzkMFiBeuJRPA8NwuZVZCE1L6i")
	ProtoBootstrap = ParseProtocolHashSafe("PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex")
	ProtoV000      = ParseProtocolHashSafe("Ps9mPmXaRzmzk35gbAYNCAw6UXdE2qoABTHbN2oEEc1qM7CwT9P")
	ProtoV001      = ParseProtocolHashSafe("PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY")
	ProtoV002      = ParseProtocolHashSafe("PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt")
	ProtoV003      = ParseProtocolHashSafe("PsddFKi32cMJ2qPjf43Qv5GDWLDPZb3T3bF6fLKiF5HtvHNU7aP")
	ProtoV004      = ParseProtocolHashSafe("Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd")
	ProtoV005_1    = ParseProtocolHashSafe("PsBABY5HQTSkA4297zNHfsZNKtxULfL18y95qb3m53QJiXGmrbU")
	ProtoV005_2    = ParseProtocolHashSafe("PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS")
	ProtoV006_1    = ParseProtocolHashSafe("PtCarthavAMoXqbjBPVgDCRd5LgT7qqKWUPXnYii3xCaHRBMfHH")
	ProtoV006_2    = ParseProtocolHashSafe("PsCARTHAGazKbHtnKfLzQg3kms52kSRpgnDY982a9oYsSXRLQEb")
	ProtoV007      = ParseProtocolHashSafe("PsDELPH1Kxsxt8f9eWbxQeRxkjfbxoqM52jvs5Y5fBxWWh4ifpo")
	ProtoV008_1    = ParseProtocolHashSafe("PtEdoTezd3RHSC31mpxxo1npxFjoWWcFgQtxapi51Z8TLu6v6Uq")
	ProtoV008_2    = ParseProtocolHashSafe("PtEdo2ZkT9oKpimTah6x2embF25oss54njMuPzkJTEi5RqfdZFA")
	ProtoV009      = ParseProtocolHashSafe("PsFLorenaUUuikDWvMDr6fGBRG8kt3e3D3fHoXK1j1BFRxeSH4i")
	ProtoV010      = ParseProtocolHashSafe("PtGRANADsDU8R9daYKAgWnQYAJ64omN1o3KMGVCykShA97vQbvV")

	Mainnet     = MustParseChainIdHash("NetXdQprcVkpaWU")
	Alphanet    = MustParseChainIdHash("NetXgtSLGNJvNye")
	Zeronet     = MustParseChainIdHash("NetXKakFj1A7ouL")
	Babylonnet  = MustParseChainIdHash("NetXUdfLh6Gm88t")
	Carthagenet = MustParseChainIdHash("NetXjD3HPJJjmcd")
	Delphinet   = MustParseChainIdHash("NetXm8tYqnMWky1")
	Edonet      = MustParseChainIdHash("NetXSp4gfdanies")
	Edonet2     = MustParseChainIdHash("NetXSgo1ZT2DRUG")
	Florencenet = MustParseChainIdHash("NetXxkAx4woPLyu")
	Granadanet  = MustParseChainIdHash("NetXz969SFaFn8k")

	// maximum depth of branches for ops to be included on chain, also
	// defines max depth of a possible reorg and max block priorities
	MaxBranchDepth int64 = 64

	// Order of deployed protocols on different networks
	// required to lookup correct block/vote/cycle offsets
	ProtocolVersions = map[string][]ProtocolHash{
		Mainnet.String(): []ProtocolHash{
			ProtoGenesis, // -1
			ProtoV000,    // 0
			ProtoV001,    // 1
			ProtoV002,    // 2
			ProtoV003,    // 3
			ProtoV004,    // 4
			ProtoV005_2,  // 5
			ProtoV006_2,  // 6
			ProtoV007,    // 7
			ProtoV008_2,  // 8
			ProtoV009,    // 9
			ProtoV010,    // 10
		},
		Granadanet.String(): []ProtocolHash{
			ProtoGenesis,   // -1
			ProtoBootstrap, // 0
			ProtoV009,      // 1
			ProtoV010,      // 2
		},
	}
)

func (p *Params) ForNetwork(net ChainIdHash) *Params {
	pp := &Params{}
	*pp = *p
	pp.ChainId = net
	switch true {
	case Mainnet.Equal(net):
		pp.Network = "Mainnet"
		pp.SecurityDepositRampUpCycles = 64
	case Alphanet.Equal(net):
		pp.Network = "Alphanet"
	case Zeronet.Equal(net):
		pp.Network = "Zeronet"
	case Babylonnet.Equal(net):
		pp.Network = "Babylonnet"
	case Carthagenet.Equal(net):
		pp.Network = "Carthagenet"
	case Delphinet.Equal(net):
		pp.Network = "Delphinet"
	case Edonet.Equal(net):
		pp.Network = "Edonet"
	case Edonet2.Equal(net):
		pp.Network = "Edonet2"
	case Florencenet.Equal(net):
		pp.Network = "Florencenet"
	case Granadanet.Equal(net):
		pp.Network = "Granadanet"
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
	switch true {
	case ProtoV000.Equal(proto):
		pp.Version = 0
		pp.ReactivateByTx = true
		pp.HasOriginationBug = true
		pp.SilentSpendable = true
	case ProtoV001.Equal(proto):
		pp.Version = 1
		pp.ReactivateByTx = true
		pp.HasOriginationBug = true
		pp.SilentSpendable = true
	case ProtoV002.Equal(proto):
		pp.Version = 2
		pp.ReactivateByTx = true
		pp.SilentSpendable = true
	case ProtoV003.Equal(proto):
		pp.Version = 3
		pp.ReactivateByTx = true
		pp.SilentSpendable = true
	case ProtoV004.Equal(proto): // Athens
		pp.Version = 4
		pp.SilentSpendable = true
		pp.Invoices = map[string]int64{
			"tz1iSQEcaGpUn6EW5uAy3XhPiNg7BHMnRSXi": 100 * 1000000,
		}
	case ProtoV005_1.Equal(proto) || ProtoV005_2.Equal(proto): // Babylon
		// changed binary operation format
		pp.Version = 5
		pp.Invoices = map[string]int64{
			"KT1DUfaMfTRZZkvZAYQT5b3byXnvqoAykc43": 500 * 1000000,
		}
		pp.OperationTagsVersion = 1

	case ProtoV006_1.Equal(proto) || ProtoV006_2.Equal(proto): // Carthage
		// no invoice
		pp.Version = 6
		pp.OperationTagsVersion = 1

	case ProtoV007.Equal(proto): // Delphi
		pp.Version = 7
		pp.OperationTagsVersion = 1
		// this is extremely hacky!
		pp.StartBlockOffset = 0
		pp.StartCycle = 0
		pp.BlocksPerCycle = 4096
		pp.BlocksPerCommitment = 32
		pp.BlocksPerRollSnapshot = 256
		pp.BlocksPerVotingPeriod = 32768

	case ProtoV008_2.Equal(proto) || ProtoV008_1.Equal(proto): // Edo
		// no invoice
		pp.Version = 8
		pp.OperationTagsVersion = 1
		pp.NumVotingPeriods = 5
		if Mainnet.Equal(p.ChainId) {
			// OK, bear with me. If you are from the future, forgive the Tezos
			// core devs.
			//
			// Apparently Tezos can switch cycle lengths which messes up all
			// kinds of calculations big-time. They have also introduced a bug
			// into vote block alignment during the switch from Delphi to Edo
			// (Granada is supposed to fix this). We use an extra offset for
			// tracking this bug.
			//
			// Edo has also added a 5th vote period and decreased period
			// durations from 32,768 blocks to 20,480 so we need a custom start
			// offset to keep our vote start/end calculations correct.
			pp.StartBlockOffset = 1343488
			pp.StartCycle = 328
			// this is extremely hacky!
			pp.BlocksPerCycle = 4096
			pp.BlocksPerCommitment = 32
			pp.BlocksPerRollSnapshot = 256
			pp.BlocksPerVotingPeriod = 20480
		}

	case ProtoV009.Equal(proto): // Florence
		// invoice (will be applied as balance update from v009+)
		pp.Version = 9
		pp.OperationTagsVersion = 1
		pp.NumVotingPeriods = 5
		if Mainnet.Equal(p.ChainId) {
			pp.StartBlockOffset = 1466368
			pp.StartCycle = 358
			// FIXME: this is extremely hacky!
			pp.BlocksPerCycle = 4096
			pp.BlocksPerCommitment = 32
			pp.BlocksPerRollSnapshot = 256
			pp.BlocksPerVotingPeriod = 20480
		} else if Granadanet.Equal(p.ChainId) {
			pp.StartBlockOffset = 0
			pp.StartCycle = 0
			// FIXME: this is extremely hacky!
			pp.BlocksPerCycle = 2048
			pp.BlocksPerCommitment = 16
			pp.BlocksPerRollSnapshot = 128
			pp.BlocksPerVotingPeriod = 10240
			pp.EndorsersPerBlock = 32
		} else {
			pp.BlocksPerCycle = 2048
			pp.BlocksPerCommitment = 16
			pp.BlocksPerRollSnapshot = 128
			pp.BlocksPerVotingPeriod = 10240
			pp.EndorsersPerBlock = 32
		}

	case ProtoV010.Equal(proto): // Granada
		pp.Version = 10
		pp.OperationTagsVersion = 1
		pp.NumVotingPeriods = 5

		// It gets more fun in Granada. Now the major cycle length has doubled
		// since block times have halved. This requires a second offset to track
		// start cycle.
		//
		// In an attempt to fix vote/cycle alignment Granada also offsets
		// voting period by +1 on activation which apparently fails on
		// Granadanet or any other network that is not mainnet because the
		// problem did not exist there.
		if Mainnet.Equal(p.ChainId) {
			pp.StartBlockOffset = 1589248
			pp.StartCycle = 389
		} else if Granadanet.Equal(p.ChainId) {
			pp.StartBlockOffset = 4096
			pp.StartCycle = 2
		}
	}
	return pp
}

func (p *Params) Clean() *Params {
	pp := &Params{}
	*pp = *p
	pp.Invoices = nil
	pp.SilentSpendable = false
	pp.HasOriginationBug = false
	pp.ReactivateByTx = false
	pp.OperationTagsVersion = 0
	pp.NumVotingPeriods = 0
	pp.StartBlockOffset = 0
	pp.StartCycle = 0
	return pp
}

func (p *Params) ForHeight(h int64) *Params {
	versions, ok := ProtocolVersions[p.ChainId.String()]
	if !ok {
		return p
	}
	pp := p.Clean()
	for i := len(versions) - 1; i >= 0; i-- {
		pp = pp.Clean().ForNetwork(p.ChainId).ForProtocol(versions[i])
		if pp.StartBlockOffset == 0 || pp.StartBlockOffset < h {
			return pp
		}
	}
	return p
}

func (p *Params) ForCycle(c int64) *Params {
	versions, ok := ProtocolVersions[p.ChainId.String()]
	if !ok {
		return p
	}
	pp := p.Clean()
	for i := len(versions) - 1; i >= 0; i-- {
		pp = pp.Clean().ForNetwork(p.ChainId).ForProtocol(versions[i])
		if pp.StartCycle == 0 || pp.StartCycle < c {
			return pp
		}
	}
	return p
}
