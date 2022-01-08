// Copyright (c) 2020-2021 Blockwatch Data Inc.
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
		ForProtocol(ProtoV011_2).
		Mixin(&Params{
			OperationTagsVersion: 1,
			TimeBetweenBlocks: [2]time.Duration{
				60 * time.Second,
				40 * time.Second,
			},
			MaxOperationsTTL:             120,
			HardGasLimitPerOperation:     1040000,
			HardGasLimitPerBlock:         5200000,
			OriginationSize:              257,
			CostPerByte:                  250,
			HardStorageLimitPerOperation: 60000,
			MinimalBlockDelay:            30 * time.Second,
		})
)

type Params struct {
	// chain identity, not part of RPC
	Name        string       `json:"name"`
	Network     string       `json:"network"`
	Symbol      string       `json:"symbol"`
	Deployment  int          `json:"deployment"` // as activated on chain
	Version     int          `json:"version"`    // as implemented
	ChainId     ChainIdHash  `json:"chain_id"`
	Protocol    ProtocolHash `json:"protocol"`
	StartHeight int64        `json:"start_height"`
	EndHeight   int64        `json:"end_height"`
	Decimals    int          `json:"decimals"`
	Token       int64        `json:"units"` // atomic units per token

	NoRewardCycles               int64            `json:"no_reward_cycles"`                // from mainnet genesis
	SecurityDepositRampUpCycles  int64            `json:"security_deposit_ramp_up_cycles"` // increase 1/64th each cycle
	PreservedCycles              int64            `json:"preserved_cycles"`
	BlocksPerCycle               int64            `json:"blocks_per_cycle"`
	BlocksPerCommitment          int64            `json:"blocks_per_commitment"`
	BlocksPerRollSnapshot        int64            `json:"blocks_per_roll_snapshot"`
	BlocksPerVotingPeriod        int64            `json:"blocks_per_voting_period"`
	TimeBetweenBlocks            [2]time.Duration `json:"time_between_blocks"`
	EndorsersPerBlock            int              `json:"endorsers_per_block"`
	HardGasLimitPerOperation     int64            `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock         int64            `json:"hard_gas_limit_per_block"`
	ProofOfWorkThreshold         int64            `json:"proof_of_work_threshold"`
	ProofOfWorkNonceSize         int              `json:"proof_of_work_nonce_size"`
	TokensPerRoll                int64            `json:"tokens_per_roll"`
	MichelsonMaximumTypeSize     int              `json:"michelson_maximum_type_size"`
	SeedNonceRevelationTip       int64            `json:"seed_nonce_revelation_tip"`
	OriginationSize              int64            `json:"origination_size"`
	OriginationBurn              int64            `json:"origination_burn"`
	BlockSecurityDeposit         int64            `json:"block_security_deposit"`
	EndorsementSecurityDeposit   int64            `json:"endorsement_security_deposit"`
	BlockReward                  int64            `json:"block_reward"`
	EndorsementReward            int64            `json:"endorsement_reward"`
	BlockRewardV6                [2]int64         `json:"block_rewards_v6"`
	EndorsementRewardV6          [2]int64         `json:"endorsement_rewards_v6"`
	CostPerByte                  int64            `json:"cost_per_byte"`
	HardStorageLimitPerOperation int64            `json:"hard_storage_limit_per_operation"`
	TestChainDuration            int64            `json:"test_chain_duration"`
	MaxOperationDataLength       int              `json:"max_operation_data_length"`
	MaxProposalsPerDelegate      int              `json:"max_proposals_per_delegate"`
	MaxRevelationsPerBlock       int              `json:"max_revelations_per_block"`
	NonceLength                  int              `json:"nonce_length"`
	MaxOperationsTTL             int64            `json:"max_operations_ttl"`

	// New in Bablyon v005
	MinProposalQuorum int64 `json:"min_proposal_quorum"`
	QuorumMin         int64 `json:"quorum_min"`
	QuorumMax         int64 `json:"quorum_max"`

	// New in Delphi v007
	MaxAnonOpsPerBlock int `json:"max_anon_ops_per_block"` // was max_revelations_per_block

	// New in Granada v010
	LiquidityBakingEscapeEmaThreshold int64         `json:"liquidity_baking_escape_ema_threshold"`
	LiquidityBakingSubsidy            int64         `json:"liquidity_baking_subsidy"`
	LiquidityBakingSunsetLevel        int64         `json:"liquidity_baking_sunset_level"`
	MinimalBlockDelay                 time.Duration `json:"minimal_block_delay"`

	// hidden invoice feature
	Invoices map[string]int64 `json:"invoices,omitempty"`

	// extra features to follow protocol upgrades
	SilentSpendable      bool  `json:"silent_spendable"` // contracts are spendable/delegatable without flag set
	HasOriginationBug    bool  `json:"has_origination_bug"`
	ReactivateByTx       bool  `json:"reactivate_by_tx"`
	OperationTagsVersion int   `json:"operation_tags_version"`
	NumVotingPeriods     int   `json:"num_voting_periods"`
	StartBlockOffset     int64 `json:"start_block_offset"` // correct start/end cycle since Granada
	StartCycle           int64 `json:"start_cycle"`        // correction since Granada v10
	VoteBlockOffset      int64 `json:"vote_block_offset"`  // correction for Edo + Florence Mainnet-only +1 bug
}

func NewParams() *Params {
	return &Params{
		Name:             "Tezos",
		Network:          "",
		Symbol:           "XTZ",
		StartHeight:      -1,
		EndHeight:        -1,
		Decimals:         6,
		Token:            1000000,
		NumVotingPeriods: 4,
		MaxOperationsTTL: 60,
	}
}

func (p *Params) Mixin(src *Params) *Params {
	buf, _ := json.Marshal(src)
	_ = json.Unmarshal(buf, p)
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
	return (height-pp.StartBlockOffset)%pp.BlocksPerRollSnapshot == 0
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

func (p *Params) SnapshotBlock(cycle, index int64) int64 {
	if cycle < p.PreservedCycles+2 {
		return 0
	}
	baseCycle := cycle - (p.PreservedCycles + 2)
	pp := p
	if !p.ContainsCycle(baseCycle) {
		pp = p.ForCycle(baseCycle)
	}
	return pp.CycleStartHeight(baseCycle) + (index+1)*pp.BlocksPerRollSnapshot - 1
}

func (p *Params) SnapshotIndex(height int64) int64 {
	pp := p
	if !p.ContainsHeight(height) {
		pp = p.ForHeight(height)
	}
	if height == pp.StartBlockOffset {
		return pp.BlocksPerCycle/pp.BlocksPerRollSnapshot - 1
	}
	return ((height - pp.StartBlockOffset - pp.BlocksPerRollSnapshot) % pp.BlocksPerCycle) / pp.BlocksPerRollSnapshot
}

func (p *Params) MaxSnapshotIndex() int64 {
	return (p.BlocksPerCycle / p.BlocksPerRollSnapshot) - 1
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
	return p.BlockReward + p.EndorsementReward*int64(p.EndorsersPerBlock)
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
	if p.MinimalBlockDelay > 0 {
		return p.MinimalBlockDelay
	}
	return p.TimeBetweenBlocks[0]
}
