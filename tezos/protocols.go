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
	ProtoV016_2    = MustParseProtocolHash("PtMumbai2TmsJHNGRkD8v8YDbtao7BLUC3wjASn1inAKLFCjaH1")
	ProtoV017      = MustParseProtocolHash("PtNairobiyssHuh87hEhfVBGCVrK3WnS8Z2FT4ymB5tAa4r1nQf")
	ProtoV018      = MustParseProtocolHash("ProxfordYmVfjWnRcgjWH36fW6PArwqykTFzotUxRs6gmTcZDuH")

	// aliases
	PtAthens  = ProtoV004
	PsBabyM1  = ProtoV005_2
	PsCARTHA  = ProtoV006_2
	PsDELPH1  = ProtoV007
	PtEdo2Zk  = ProtoV008_2
	PsFLoren  = ProtoV009
	PtGRANAD  = ProtoV010
	PtHangz2  = ProtoV011_2
	Psithaca  = ProtoV012_2
	PtJakart  = ProtoV013_2
	PtKathma  = ProtoV014
	PtLimaPt  = ProtoV015
	PtMumbai  = ProtoV016_2
	PtNairobi = ProtoV017
	Proxford  = ProtoV018

	Mainnet    = MustParseChainIdHash("NetXdQprcVkpaWU")
	Ghostnet   = MustParseChainIdHash("NetXnHfVqm9iesp")
	Nairobinet = MustParseChainIdHash("NetXyuzvDo2Ugzb")
	Oxfordnet  = MustParseChainIdHash("NetXxWsskGahzQB")

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
		ProtoV016_2:    16,
		ProtoV017:      17,
		ProtoV018:      18,
		ProtoAlpha:     19,
	}

	Deployments = map[ChainIdHash]ProtocolHistory{
		Mainnet: {
			{ProtoGenesis, 0, 0, 0, 0, 5, 4096, 256},              // 0
			{ProtoBootstrap, 0, 1, 1, 0, 5, 4096, 256},            // 0
			{ProtoV001, 2, 2, 28082, 0, 5, 4096, 256},             // v1
			{ProtoV002, 3507, 28083, 204761, 6, 5, 4096, 256},     // v2
			{ProtoV003, 4057, 204762, 458752, 49, 5, 4096, 256},   // v3
			{PtAthens, 0, 458753, 655360, 112, 5, 4096, 256},      // v4
			{PsBabyM1, 0, 655361, 851968, 160, 5, 4096, 256},      // v5
			{PsCARTHA, 0, 851969, 1212416, 208, 5, 4096, 256},     // v6
			{PsDELPH1, 0, 1212417, 1343488, 296, 5, 4096, 256},    // v7
			{PtEdo2Zk, 0, 1343489, 1466367, 328, 5, 4096, 256},    // v8
			{PsFLoren, 4095, 1466368, 1589247, 357, 5, 4096, 256}, // v9
			{PtGRANAD, -1, 1589248, 1916928, 388, 5, 8192, 512},   // v10
			{PtHangz2, 0, 1916929, 2244608, 428, 5, 8192, 512},    // v11
			{Psithaca, 0, 2244609, 2490368, 468, 5, 8192, 512},    // v12
			{PtJakart, 0, 2490369, 2736128, 498, 5, 8192, 512},    // v13
			{PtKathma, 0, 2736129, 2981888, 528, 5, 8192, 512},    // v14
			{PtLimaPt, 0, 2981889, 3268608, 558, 5, 8192, 512},    // v15
			{PtMumbai, 0, 3268609, 3760128, 593, 5, 16384, 1024},  // v16
			{PtNairobi, 0, 3760129, -1, 623, 5, 16384, 1024},      // v17
			{Proxford, 0, 5070849, -1, 703, 5, 16384, 1024},       // v18
		},
		Ghostnet: {
			{ProtoGenesis, 0, 0, 0, 0, 3, 4096, 256},           // 0
			{ProtoBootstrap, 0, 1, 1, 0, 3, 4096, 256},         // 0
			{PtHangz2, 2, 2, 8191, 0, 3, 4096, 256},            // v11
			{Psithaca, 0, 8192, 765952, 2, 3, 4096, 256},       // v12
			{PtJakart, 0, 765953, 1191936, 187, 3, 4096, 256},  // v13
			{PtKathma, 0, 1191937, 1654784, 291, 3, 4096, 256}, // v14
			{PtLimaPt, 0, 1654785, 2162688, 404, 3, 4096, 256}, // v15
			{PtMumbai, 0, 2162689, 2957312, 528, 3, 8192, 512}, // v16
			{PtNairobi, 0, 2957313, -1, 625, 3, 8192, 512},     // v17
			// {Proxford, 0, 2957313, -1, 625, 3, 8192, 512},     // v18
		},
		Nairobinet: {
			{ProtoGenesis, 0, 0, 0, 0, 3, 8192, 512},   // 0
			{ProtoBootstrap, 0, 1, 1, 0, 3, 8192, 512}, // 0
			{PtMumbai, 2, 2, 16384, 0, 3, 8192, 512},   // v16
			{PtNairobi, 0, 16385, -1, 2, 3, 8192, 512}, // v17
		},
		Oxfordnet: {
			{ProtoGenesis, 0, 0, 0, 0, 3, 8192, 512},   // 0
			{ProtoBootstrap, 0, 1, 1, 0, 3, 8192, 512}, // 0
			{PtNairobi, 2, 2, 16384, 0, 3, 8192, 512},  // v17
			{Proxford, 0, 16385, -1, 2, 3, 8192, 512},  // v18
		},
	}
)

type Deployment struct {
	Protocol          ProtocolHash
	StartOffset       int64
	StartHeight       int64
	EndHeight         int64
	StartCycle        int64
	PreservedCycles   int64
	BlocksPerCycle    int64
	BlocksPerSnapshot int64
}

type ProtocolHistory []Deployment

func (h ProtocolHistory) Clone() ProtocolHistory {
	clone := make(ProtocolHistory, len(h))
	copy(clone, h)
	return clone
}

func (h ProtocolHistory) AtBlock(height int64) (d Deployment) {
	d = h.Last()
	for i := len(h) - 1; i >= 0; i-- {
		if h[i].StartHeight <= height {
			d = h[i]
			break
		}
	}
	return
}

func (h ProtocolHistory) AtCycle(cycle int64) (d Deployment) {
	d = h.Last()
	for i := len(h) - 1; i >= 0; i-- {
		if h[i].StartCycle <= cycle {
			d = h[i]
			break
		}
	}
	return
}

func (h ProtocolHistory) AtProtocol(proto ProtocolHash) (d Deployment) {
	d = h.Last()
	for _, v := range h {
		if v.Protocol == proto {
			d = v
			break
		}
	}
	return
}

func (h *ProtocolHistory) Add(d Deployment) {
	(*h) = append((*h), d)
}

func (h ProtocolHistory) Last() (d Deployment) {
	if l := len(h); l > 0 {
		d = h[l-1]
	}
	return
}
