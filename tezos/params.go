// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"time"
)

var (
	// DefaultParams defines the blockchain configuration for Mainnet under the latest
	// protocol. It is used to generate compliant transaction encodings. To change,
	// either overwrite this default or set custom params per operation using
	// op.WithParams().
	DefaultParams = &Params{
		Network:                      "Mainnet",
		ChainId:                      Mainnet,
		Protocol:                     ProtoV016_2,
		Version:                      16,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             240,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            15 * time.Second,
		PreservedCycles:              5,
	}

	// GhostnetParams defines the blockchain configuration for Ghostnet testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	GhostnetParams = &Params{
		Network:                      "Ghostnet",
		ChainId:                      Ghostnet,
		Protocol:                     ProtoV016_2,
		Version:                      16,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             240,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            15 * time.Second,
		PreservedCycles:              3,
	}

	// LimanetParams defines the blockchain configuration for Lima testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	LimanetParams = &Params{
		Network:                      "Limanet",
		ChainId:                      Limanet,
		Protocol:                     ProtoV015,
		Version:                      15,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             120,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         5200000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            15 * time.Second,
		PreservedCycles:              3,
	}

	// MumbainetParams defines the blockchain configuration for Mumbai testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	MumbainetParams = &Params{
		Network:                      "Mumbainet",
		ChainId:                      Mumbainet,
		Protocol:                     ProtoV016_2,
		Version:                      16,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             240,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            8 * time.Second,
		PreservedCycles:              3,
	}
)

func NewParams() *Params {
	return &Params{
		Name:             "Tezos",
		Network:          "",
		Symbol:           "XTZ",
		StartHeight:      -1,
		EndHeight:        -1,
		Decimals:         6,
		Token:            1000000, // initial, changed several times later
		NumVotingPeriods: 4,       // initial, changed once in v008
		MaxOperationsTTL: 60,      // initial, changed once in v011
	}
}

// Params contains a subset of protocol configuration settings that are relevant
// for dapps and most indexers. For additional protocol data, call rpc.GetCustomConstants()
// with a custom data struct.
type Params struct {
	// identity
	Name        string       `json:"name"`
	Network     string       `json:"network,omitempty"`
	Symbol      string       `json:"symbol"`
	Deployment  int          `json:"deployment"`
	Version     int          `json:"version"`
	ChainId     ChainIdHash  `json:"chain_id"`
	Protocol    ProtocolHash `json:"protocol"`
	StartHeight int64        `json:"start_height"`
	EndHeight   int64        `json:"end_height"`
	Decimals    int          `json:"decimals"`
	Token       int64        `json:"units"` // atomic units per token

	EndorsementReward int64 `json:"endorsement_reward"`

	// timing
	MinimalBlockDelay time.Duration `json:"minimal_block_delay"`

	// costs
	CostPerByte     int64 `json:"cost_per_byte"`
	OriginationSize int64 `json:"origination_size"`

	// limits
	BlockReward                  int64 `json:"block_reward"`
	BlocksPerCycle               int64 `json:"blocks_per_cycle"`
	EndorsersPerBlock            int   `json:"endorsers_per_block"`
	PreservedCycles              int64 `json:"preserved_cycles"`
	HardGasLimitPerOperation     int64 `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock         int64 `json:"hard_gas_limit_per_block"`
	HardStorageLimitPerOperation int64 `json:"hard_storage_limit_per_operation"`
	MaxOperationDataLength       int   `json:"max_operation_data_length"`
	MaxOperationsTTL             int64 `json:"max_operations_ttl"`
	BlocksPerCommitment          int64 `json:"blocks_per_commitment"`
	BlocksPerSnapshot            int64 `json:"blocks_per_snapshot"`
	ConsensusCommitteeSize       int   `json:"consensus_committee_size"`

	BlocksPerVotingPeriod int64 `json:"blocks_per_voting_period"`
	CyclesPerVotingPeriod int64 `json:"cycles_per_voting_period"`
	MinProposalQuorum     int64 `json:"min_proposal_quorum"`
	QuorumMin             int64 `json:"quorum_min"`
	QuorumMax             int64 `json:"quorum_max"`

	// extra features to follow protocol upgrades
	OperationTagsVersion int   `json:"operation_tags_version,omitempty"` // 1 after v005
	NumVotingPeriods     int   `json:"num_voting_periods,omitempty"`     // 5 after v008, 4 before
	StartBlockOffset     int64 `json:"start_block_offset,omitempty"`     // correct start/end cycle since Granada
	StartCycle           int64 `json:"start_cycle,omitempty"`            // correction since Granada v10
	VoteBlockOffset      int64 `json:"vote_block_offset,omitempty"`      // correction for Edo + Florence Mainnet-only +1 bug
}

func (p *Params) WithChainId(id ChainIdHash) *Params {
	p.ChainId = id
	return p
}

func (p *Params) WithProtocol(h ProtocolHash) *Params {
	p.Protocol = h
	p.Version = Versions[h]
	switch {
	case p.Version > 11:
		p.OperationTagsVersion = 2
	case p.Version > 4:
		p.OperationTagsVersion = 1
	}
	return p
}

func (p *Params) WithNetwork(n string) *Params {
	p.Network = n
	return p
}

func (p Params) SnapshotBaseCycle(cycle int64) int64 {
	var offset int64 = 2
	if p.Version >= 12 {
		offset = 1
	}
	return cycle - (p.PreservedCycles + offset)
}

func (p *Params) IsCycleStart(height int64) bool {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	return height > 0 && (height-pp.StartBlockOffset-1)%pp.BlocksPerCycle == 0
}

func (p *Params) IsCycleEnd(height int64) bool {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	return height > 0 && (height-pp.StartBlockOffset)%pp.BlocksPerCycle == 0
}

func (p *Params) IsSnapshotBlock(height int64) bool {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	return (height-pp.StartBlockOffset)%pp.SnapshotBlocks() == 0
}

func (p *Params) IsSeedRequired(height int64) bool {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	return (height-pp.StartBlockOffset)%pp.BlocksPerCommitment == 0
}

func (p *Params) CycleFromHeight(height int64) int64 {
	if height == 0 {
		return 0
	}
	correct := int64(0)
	if p.StartBlockOffset == height {
		correct = 1
	}
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	return pp.StartCycle + (height-pp.StartBlockOffset-1)/pp.BlocksPerCycle - correct
}

func (p *Params) CycleStartHeight(cycle int64) int64 {
	pp := p
	if !p.ContainsCycle(cycle) {
		pp = p.ForCycle(cycle)
	}
	return pp.StartBlockOffset + (cycle-pp.StartCycle)*pp.BlocksPerCycle + 1
}

func (p *Params) CycleEndHeight(cycle int64) int64 {
	pp := p
	if !p.ContainsCycle(cycle) {
		pp = p.ForCycle(cycle)
	}
	return pp.StartBlockOffset + (cycle-pp.StartCycle+1)*pp.BlocksPerCycle
}

func (p *Params) CyclePosition(height int64) int64 {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	return height - pp.CycleStartHeight(pp.CycleFromHeight(height))
}

func (p *Params) SnapshotBlock(cycle int64, index int) int64 {
	base := p.SnapshotBaseCycle(cycle)
	if base < 0 {
		return 0
	}
	pp := p
	if !p.ContainsCycle(base) {
		pp = p.ForCycle(base)
	}
	return pp.CycleStartHeight(base) + int64(index+1)*pp.SnapshotBlocks() - 1
}

func (p *Params) SnapshotIndex(height int64) int {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	if height == pp.StartBlockOffset {
		return int(pp.BlocksPerCycle/pp.SnapshotBlocks() - 1)
	}
	return int(((height - pp.StartBlockOffset - pp.SnapshotBlocks()) % pp.BlocksPerCycle) / pp.SnapshotBlocks())
}

func (p *Params) SnapshotBlocks() int64 {
	return p.BlocksPerSnapshot
}

func (p *Params) MaxSnapshotIndex() int64 {
	return (p.BlocksPerCycle / p.SnapshotBlocks()) - 1
}

func (p *Params) VotingStartCycleFromHeight(height int64) int64 {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	// Edo voting bug does not apply to first Edo block
	offs := pp.VoteBlockOffset
	if height == pp.StartBlockOffset+1 {
		offs = 0
	}
	currentCycle := pp.CycleFromHeight(height + offs)
	offset := (height + offs - pp.StartBlockOffset - 1) % pp.BlocksPerVotingPeriod
	return currentCycle - offset/pp.BlocksPerCycle
}

func (p *Params) IsVoteStart(height int64) bool {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	// Edo voting bug does not apply to first Edo block
	offs := pp.VoteBlockOffset
	if height == pp.StartBlockOffset+1 {
		offs = 0
	}
	return height > 0 && (height-pp.StartBlockOffset-1+offs)%pp.BlocksPerVotingPeriod == 0
}

func (p *Params) IsVoteEnd(height int64) bool {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	offs := pp.VoteBlockOffset
	return height > 0 && (height-pp.StartBlockOffset+offs)%pp.BlocksPerVotingPeriod == 0
}

func (p *Params) VoteStartHeight(height int64) int64 {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	// Edo voting bug does not apply to first Edo block
	offs := pp.VoteBlockOffset
	if height == pp.StartBlockOffset+1 {
		offs = 0
	}
	return pp.CycleStartHeight(pp.VotingStartCycleFromHeight(height)) - offs
}

func (p *Params) VoteEndHeight(height int64) int64 {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	return pp.VoteStartHeight(height) + pp.BlocksPerVotingPeriod - 1
}

func (p *Params) MaxBlockReward() int64 {
	if p.Version < 12 {
		return p.BlockReward + p.EndorsementReward*int64(p.EndorsersPerBlock)
	}
	return p.BlockReward + p.EndorsementReward*int64(p.ConsensusCommitteeSize)
}

func (p *Params) ContainsHeight(height int64) bool {
	// treat -1 as special height query that matches open interval params only
	return (height < 0 && p.EndHeight < 0) ||
		(p.StartHeight <= height && (p.EndHeight < 0 || p.EndHeight >= height))
}

func (p *Params) ContainsCycle(cycle int64) bool {
	return p.StartCycle == 0 || p.StartCycle <= cycle
}

func (p *Params) IsMainnet() bool {
	return p.ChainId.Equal(Mainnet)
}

func (p *Params) IsPostBabylon() bool {
	return p.IsMainnet() && p.Version >= 5
}

func (p *Params) IsPreBabylonHeight(height int64) bool {
	return p.IsMainnet() && height < 655360
}

func (p *Params) BlockTime() time.Duration {
	return p.MinimalBlockDelay
}

func (p *Params) NumEndorsers() int {
	return p.EndorsersPerBlock + p.ConsensusCommitteeSize
}
