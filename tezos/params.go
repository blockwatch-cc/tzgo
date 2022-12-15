// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"encoding/json"
	"time"
)

var (
	// DefaultParams defines the blockchain configuration for Mainnet under the latest
	// protocol. It is used to generate compliant transaction encodings. To change,
	// either overwrite this default or set custom params per operation using
	// op.WithParams().
	DefaultParams = NewParams().
			ForNetwork(Mainnet).
			ForProtocol(ProtoV012_2).
			SetBaseParams(BaseParams{
			OperationTagsVersion:         2,
			MaxOperationsTTL:             120,
			HardGasLimitPerOperation:     1040000,
			HardGasLimitPerBlock:         5200000,
			OriginationSize:              257,
			CostPerByte:                  250,
			HardStorageLimitPerOperation: 60000,
			MinimalBlockDelay:            30 * time.Second,
		})

	// GhostnetParams defines the blockchain configuration for Ghostnet testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	GhostnetParams = NewParams().
			ForNetwork(Ghostnet).
			ForProtocol(ProtoV013_2).
			SetBaseParams(BaseParams{
			OperationTagsVersion:         2,
			MaxOperationsTTL:             120,
			HardGasLimitPerOperation:     1040000,
			HardGasLimitPerBlock:         5200000,
			OriginationSize:              257,
			CostPerByte:                  250,
			HardStorageLimitPerOperation: 60000,
			MinimalBlockDelay:            15 * time.Second,
		})

	// JakartanetParams defines the blockchain configuration for Ithaca testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	JakartanetParams = NewParams().
				ForNetwork(Jakartanet).
				ForProtocol(ProtoV013_2).
				SetBaseParams(BaseParams{
			OperationTagsVersion:         2,
			MaxOperationsTTL:             120,
			HardGasLimitPerOperation:     1040000,
			HardGasLimitPerBlock:         5200000,
			OriginationSize:              257,
			CostPerByte:                  250,
			HardStorageLimitPerOperation: 60000,
			MinimalBlockDelay:            15 * time.Second,
		})

	// KathmandunetParams defines the blockchain configuration for Kathmandu testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	KathmandunetParams = NewParams().
				ForNetwork(Kathmandunet).
				ForProtocol(ProtoV014).
				SetBaseParams(BaseParams{
			OperationTagsVersion:         2,
			MaxOperationsTTL:             120,
			HardGasLimitPerOperation:     1040000,
			HardGasLimitPerBlock:         5200000,
			OriginationSize:              257,
			CostPerByte:                  250,
			HardStorageLimitPerOperation: 60000,
			MinimalBlockDelay:            15 * time.Second,
		})

	// LimanetParams defines the blockchain configuration for Kathmandu testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	LimanetParams = NewParams().
			ForNetwork(Limanet).
			ForProtocol(ProtoV015).
			SetBaseParams(BaseParams{
			OperationTagsVersion:         2,
			MaxOperationsTTL:             120,
			HardGasLimitPerOperation:     1040000,
			HardGasLimitPerBlock:         5200000,
			OriginationSize:              257,
			CostPerByte:                  250,
			HardStorageLimitPerOperation: 60000,
			MinimalBlockDelay:            15 * time.Second,
		})
)

type BaseParams struct {
	// timing
	MinimalBlockDelay time.Duration `json:"minimal_block_delay"`

	// costs
	CostPerByte     int64 `json:"cost_per_byte"`
	OriginationSize int64 `json:"origination_size"`

	// limits
	HardGasLimitPerOperation     int64 `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock         int64 `json:"hard_gas_limit_per_block"`
	HardStorageLimitPerOperation int64 `json:"hard_storage_limit_per_operation"`
	MaxOperationsTTL             int64 `json:"max_operations_ttl"`
	// extra features to follow protocol upgrades
	OperationTagsVersion int `json:"operation_tags_version,omitempty"` // 1 after v005
}

// Params contains a subset of protocol configuration settings that are relevant
// for dapps and most indexers. For additional protocol data, call rpc.GetCustomConstants()
// with a custom data struct.
type Params struct {
	BaseParams

	// chain identity, not part of RPC
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

	// sizes
	MinimalStake        int64 `json:"minimal_stake"`
	PreservedCycles     int64 `json:"preserved_cycles"`
	BlocksPerCycle      int64 `json:"blocks_per_cycle"`
	BlocksPerCommitment int64 `json:"blocks_per_commitment"`
	BlocksPerSnapshot   int64 `json:"blocks_per_snapshot"`

	// timing
	DelayIncrementPerRound time.Duration `json:"delay_increment_per_round"`

	// rewards
	SeedNonceRevelationTip   int64    `json:"seed_nonce_revelation_tip"`
	BlockReward              int64    `json:"block_reward"`
	EndorsementReward        int64    `json:"endorsement_reward"`
	BlockRewardV6            [2]int64 `json:"block_rewards_v6"`
	EndorsementRewardV6      [2]int64 `json:"endorsement_rewards_v6"`
	BakingRewardFixedPortion int64    `json:"baking_reward_fixed_portion"`
	BakingRewardBonusPerSlot int64    `json:"baking_reward_bonus_per_slot"`
	EndorsingRewardPerSlot   int64    `json:"endorsing_reward_per_slot"`

	// costs
	OriginationBurn            int64 `json:"origination_burn"`
	BlockSecurityDeposit       int64 `json:"block_security_deposit"`
	EndorsementSecurityDeposit int64 `json:"endorsement_security_deposit"`
	FrozenDepositsPercentage   int   `json:"frozen_deposits_percentage"`

	// limits
	MichelsonMaximumTypeSize int `json:"michelson_maximum_type_size"`
	EndorsersPerBlock        int `json:"endorsers_per_block"`
	MaxOperationDataLength   int `json:"max_operation_data_length"`
	ConsensusCommitteeSize   int `json:"consensus_committee_size"`
	ConsensusThreshold       int `json:"consensus_threshold"`

	// voting
	BlocksPerVotingPeriod int64 `json:"blocks_per_voting_period"`
	CyclesPerVotingPeriod int64 `json:"cycles_per_voting_period"`
	MinProposalQuorum     int64 `json:"min_proposal_quorum"`
	QuorumMin             int64 `json:"quorum_min"`
	QuorumMax             int64 `json:"quorum_max"`

	// extra features to follow protocol upgrades
	NumVotingPeriods int   `json:"num_voting_periods,omitempty"` // 5 after v008, 4 before
	StartBlockOffset int64 `json:"start_block_offset,omitempty"` // correct start/end cycle since Granada
	StartCycle       int64 `json:"start_cycle,omitempty"`        // correction since Granada v10
	VoteBlockOffset  int64 `json:"vote_block_offset,omitempty"`  // correction for Edo + Florence Mainnet-only +1 bug
}

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
		BaseParams: BaseParams{
			MaxOperationsTTL: 60, // initial, changed once in v011
		},
	}
}

func (p *Params) Mixin(src *Params) *Params {
	buf, _ := json.Marshal(src)
	_ = json.Unmarshal(buf, p)
	return p
}

func (p *Params) SetBaseParams(src BaseParams) *Params {
	p.BaseParams = src
	return p
}

// convertAmount converts a floating point number, which may or may not be representable
// as an integer, to an integer type by rounding to the nearest integer.
// This is performed consistent with the General Decimal Arithmetic spec
// and according to IEEE 754-2008 roundTiesToEven
func (p *Params) ConvertAmount(value float64) int64 {
	sign := int64(1)
	if value < 0 {
		sign = -1
	}
	f := value * float64(p.Token)
	i := int64(f)
	rem := (f - float64(i)) * float64(sign)
	if rem > 0.5 || rem == 0.5 && i*sign%2 == 1 {
		i += sign
	}
	return i
}

func (p *Params) ConvertValue(amount int64) float64 {
	return float64(amount) / float64(p.Token)
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

func (p *Params) SnapshotBaseCycle(cycle int64) int64 {
	var offset int64 = 2
	if p.Version >= 12 {
		offset = 1
	}
	return cycle - (p.PreservedCycles + offset)
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
