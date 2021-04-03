// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"context"
	"strconv"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

type Tip struct {
	Name               string            `json:"name"`
	Network            string            `json:"network"`
	Symbol             string            `json:"symbol"`
	ChainId            tezos.ChainIdHash `json:"chain_id"`
	GenesisTime        time.Time         `json:"genesis_time"`
	BestHash           tezos.BlockHash   `json:"block_hash"`
	Timestamp          time.Time         `json:"timestamp"`
	Height             int64             `json:"height"`
	Cycle              int64             `json:"cycle"`
	TotalAccounts      int64             `json:"total_accounts"`
	FundedAccounts     int64             `json:"funded_accounts"`
	TotalOps           int64             `json:"total_ops"`
	Delegators         int64             `json:"delegators"`
	Delegates          int64             `json:"delegates"`
	Rolls              int64             `json:"rolls"`
	RollOwners         int64             `json:"roll_owners"`
	NewAccounts30d     int64             `json:"new_accounts_30d"`
	ClearedAccounts30d int64             `json:"cleared_accounts_30d"`
	FundedAccounts30d  int64             `json:"funded_accounts_30d"`
	Inflation1Y        float64           `json:"inflation_1y"`
	InflationRate1Y    float64           `json:"inflation_rate_1y"`
	Health             int               `json:"health"`
	Supply             *Supply           `json:"supply"`
	Deployments        []Deployment      `json:"deployments"`
	Status             Status            `json:"status"`
}

type Deployment struct {
	Protocol    string `json:"protocol"`
	Version     int    `json:"version"`      // protocol version sequence on indexed chain
	StartHeight int64  `json:"start_height"` // first block on indexed chain
	EndHeight   int64  `json:"end_height"`   // last block on indexed chain or -1
}

type Status struct {
	Status   string  `json:"status"` // loading, connecting, stopping, stopped, waiting, syncing, synced, failed
	Blocks   int64   `json:"blocks"`
	Indexed  int64   `json:"indexed"`
	Progress float64 `json:"progress"`
}

func (c *Client) GetStatus(ctx context.Context) (*Status, error) {
	s := &Status{}
	if err := c.get(ctx, "/explorer/status", nil, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Client) GetTip(ctx context.Context) (*Tip, error) {
	tip := &Tip{}
	if err := c.get(ctx, "/explorer/tip", nil, tip); err != nil {
		return nil, err
	}
	return tip, nil
}

type BlockchainConfig struct {
	Name                         string     `json:"name"`
	Network                      string     `json:"network"`
	Symbol                       string     `json:"symbol"`
	ChainId                      string     `json:"chain_id"`
	Version                      int        `json:"version"`
	Deployment                   int        `json:"deployment"`
	Protocol                     string     `json:"protocol"`
	StartHeight                  int64      `json:"start_height"`
	EndHeight                    int64      `json:"end_height"`
	NoRewardCycles               int64      `json:"no_reward_cycles"`
	SecurityDepositRampUpCycles  int64      `json:"security_deposit_ramp_up_cycles"`
	Decimals                     int        `json:"decimals"`
	Token                        int64      `json:"units"`
	BlockReward                  float64    `json:"block_rewards"`
	BlockSecurityDeposit         float64    `json:"block_security_deposit"`
	BlocksPerCommitment          int64      `json:"blocks_per_commitment"`
	BlocksPerCycle               int64      `json:"blocks_per_cycle"`
	BlocksPerRollSnapshot        int64      `json:"blocks_per_roll_snapshot"`
	BlocksPerVotingPeriod        int64      `json:"blocks_per_voting_period"`
	CostPerByte                  int64      `json:"cost_per_byte"`
	EndorsementReward            float64    `json:"endorsement_reward"`
	EndorsementSecurityDeposit   float64    `json:"endorsement_security_deposit"`
	EndorsersPerBlock            int        `json:"endorsers_per_block"`
	HardGasLimitPerBlock         int64      `json:"hard_gas_limit_per_block"`
	HardGasLimitPerOperation     int64      `json:"hard_gas_limit_per_operation"`
	HardStorageLimitPerOperation int64      `json:"hard_storage_limit_per_operation"`
	MaxOperationDataLength       int        `json:"max_operation_data_length"`
	MaxProposalsPerDelegate      int        `json:"max_proposals_per_delegate"`
	MaxRevelationsPerBlock       int        `json:"max_revelations_per_block"`
	MichelsonMaximumTypeSize     int        `json:"michelson_maximum_type_size"`
	NonceLength                  int        `json:"nonce_length"`
	OriginationBurn              float64    `json:"origination_burn"`
	OriginationSize              int64      `json:"origination_size"`
	PreservedCycles              int64      `json:"preserved_cycles"`
	ProofOfWorkNonceSize         int        `json:"proof_of_work_nonce_size"`
	ProofOfWorkThreshold         uint64     `json:"proof_of_work_threshold"`
	SeedNonceRevelationTip       float64    `json:"seed_nonce_revelation_tip"`
	TimeBetweenBlocks            [2]int     `json:"time_between_blocks"`
	TokensPerRoll                float64    `json:"tokens_per_roll"`
	TestChainDuration            int64      `json:"test_chain_duration"`
	MinProposalQuorum            int64      `json:"min_proposal_quorum"`
	QuorumMin                    int64      `json:"quorum_min"`
	QuorumMax                    int64      `json:"quorum_max"`
	BlockRewardV6                [2]float64 `json:"block_rewards_v6"`
	EndorsementRewardV6          [2]float64 `json:"endorsement_rewards_v6"`
	MaxAnonOpsPerBlock           int        `json:"max_anon_ops_per_block"`
	NumVotingPeriods             int        `json:"num_voting_periods"`
}

func (c *Client) GetConfig(ctx context.Context) (*BlockchainConfig, error) {
	config := &BlockchainConfig{}
	if err := c.get(ctx, "/explorer/config/head", nil, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Client) GetConfigHeight(ctx context.Context, height int64) (*BlockchainConfig, error) {
	config := &BlockchainConfig{}
	if err := c.get(ctx, "/explorer/config/"+strconv.FormatInt(height, 10), nil, config); err != nil {
		return nil, err
	}
	return config, nil
}

type Supply struct {
	RowId               uint64    `json:"row_id"`
	Height              int64     `json:"height"`
	Cycle               int64     `json:"cycle"`
	Timestamp           time.Time `json:"time"`
	Total               float64   `json:"total"`
	Activated           float64   `json:"activated"`
	Unclaimed           float64   `json:"unclaimed"`
	Vested              float64   `json:"vested"`
	Unvested            float64   `json:"unvested"`
	Circulating         float64   `json:"circulating"`
	Delegated           float64   `json:"delegated"`
	Staking             float64   `json:"staking"`
	ActiveDelegated     float64   `json:"active_delegated"`
	ActiveStaking       float64   `json:"active_staking"`
	InactiveDelegated   float64   `json:"inactive_delegated"`
	InactiveStaking     float64   `json:"inactive_staking"`
	Minted              float64   `json:"minted"`
	MintedBaking        float64   `json:"minted_baking"`
	MintedEndorsing     float64   `json:"minted_endorsing"`
	MintedSeeding       float64   `json:"minted_seeding"`
	MintedAirdrop       float64   `json:"minted_airdrop"`
	Burned              float64   `json:"burned"`
	BurnedDoubleBaking  float64   `json:"burned_double_baking"`
	BurnedDoubleEndorse float64   `json:"burned_double_endorse"`
	BurnedOrigination   float64   `json:"burned_origination"`
	BurnedImplicit      float64   `json:"burned_implicit"`
	BurnedSeedMiss      float64   `json:"burned_seed_miss"`
	Frozen              float64   `json:"frozen"`
	FrozenDeposits      float64   `json:"frozen_deposits"`
	FrozenRewards       float64   `json:"frozen_rewards"`
	FrozenFees          float64   `json:"frozen_fees"`
}
