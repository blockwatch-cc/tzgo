// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

// Constants represents only a limited subset of Tezos chain configuration params
// which are required by TzGo. Users must define custom structs to read other
// constants as needed.
type Constants struct {
	PreservedCycles              int64    `json:"preserved_cycles"`
	BlocksPerCycle               int64    `json:"blocks_per_cycle"`
	BlocksPerRollSnapshot        int64    `json:"blocks_per_roll_snapshot"`
	TimeBetweenBlocks            []string `json:"time_between_blocks"`
	HardGasLimitPerOperation     int64    `json:"hard_gas_limit_per_operation,string"`
	HardGasLimitPerBlock         int64    `json:"hard_gas_limit_per_block,string"`
	MichelsonMaximumTypeSize     int      `json:"michelson_maximum_type_size"`
	OriginationSize              int64    `json:"origination_size"`
	OriginationBurn              int64    `json:"origination_burn,string"`
	CostPerByte                  int64    `json:"cost_per_byte,string"`
	HardStorageLimitPerOperation int64    `json:"hard_storage_limit_per_operation,string"`
	MaxOperationDataLength       int      `json:"max_operation_data_length"`

	// New in v10
	MinimalBlockDelay int `json:"minimal_block_delay,string"`

	// New in v12
	MaxOperationsTimeToLive int64 `json:"max_operations_time_to_live"`
	BlocksPerStakeSnapshot  int64 `json:"blocks_per_stake_snapshot"`
}

// GetConstants returns chain configuration constants at block id
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-constants
func (c *Client) GetConstants(ctx context.Context, id BlockID) (con Constants, err error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/constants", id)
	err = c.Get(ctx, u, &con)
	return
}

// GetCustomConstants returns chain configuration constants at block id
// marshaled into a user-defined structure.
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-constants
func (c *Client) GetCustomConstants(ctx context.Context, id BlockID, resp any) error {
	u := fmt.Sprintf("chains/main/blocks/%s/context/constants", id)
	return c.Get(ctx, u, resp)
}

// GetParams returns a translated parameters structure for the current
// network at block id.
func (c *Client) GetParams(ctx context.Context, id BlockID) (*tezos.Params, error) {
	if !c.ChainId.IsValid() {
		id, err := c.GetChainId(ctx)
		if err != nil {
			return nil, err
		}
		c.ChainId = id
	}
	meta, err := c.GetBlockMetadata(ctx, id)
	if err != nil {
		return nil, err
	}
	con, err := c.GetConstants(ctx, id)
	if err != nil {
		return nil, err
	}
	ver, err := c.GetVersionInfo(ctx)
	if err != nil {
		return nil, err
	}
	p := con.MapToChainParams().
		WithChainId(c.ChainId).
		WithProtocol(meta.Protocol).
		WithNetwork(ver.NetworkVersion.ChainName).
		WithBlock(meta.GetLevel())
	return p, nil
}

func (c Constants) MapToChainParams() *tezos.Params {
	p := &tezos.Params{
		BlocksPerCycle:               c.BlocksPerCycle,
		PreservedCycles:              c.PreservedCycles,
		BlocksPerSnapshot:            c.BlocksPerRollSnapshot + c.BlocksPerStakeSnapshot,
		OriginationSize:              c.OriginationSize + c.OriginationBurn,
		CostPerByte:                  c.CostPerByte,
		HardGasLimitPerOperation:     c.HardGasLimitPerOperation,
		HardGasLimitPerBlock:         c.HardGasLimitPerBlock,
		HardStorageLimitPerOperation: c.HardStorageLimitPerOperation,
		MaxOperationDataLength:       c.MaxOperationDataLength,
		MaxOperationsTTL:             c.MaxOperationsTimeToLive,
		MinimalBlockDelay:            time.Duration(c.MinimalBlockDelay) * time.Second,
	}

	// default for old protocols
	if p.MaxOperationsTTL == 0 {
		p.MaxOperationsTTL = 120
	}

	// timing on old protocols
	if len(c.TimeBetweenBlocks) > 0 {
		if val, err := strconv.ParseInt(c.TimeBetweenBlocks[0], 10, 64); err == nil {
			p.MinimalBlockDelay = time.Duration(val) * time.Second
		}
	}

	return p
}
