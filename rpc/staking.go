// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"fmt"

	"blockwatch.cc/tzgo/tezos"
)

type StakingParameters struct {
	Cycle int64 `json:"cycle"`
	Limit int64 `json:"limit_of_staking_over_baking_millionth"`
	Edge  int64 `json:"edge_of_baking_over_staking_billionth"`
}

// GetDelegateStakingParams returns a delegate's current staking setup
func (c *Client) GetDelegateStakingParams(ctx context.Context, addr tezos.Address, id BlockID) (*StakingParameters, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/delegates/%s/active_staking_parameters", id, addr)
	p := &StakingParameters{}
	if err := c.Get(ctx, u, p); err != nil {
		return nil, err
	}
	return p, nil
}

// GetDelegatePendingStakingParams returns a delegate's future staking setup
func (c *Client) GetDelegatePendingStakingParams(ctx context.Context, addr tezos.Address, id BlockID) ([]StakingParameters, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/delegates/%s/pending_staking_parameters", id, addr)
	list := make([]StakingParameters, 0, 5)
	if err := c.Get(ctx, u, &list); err != nil {
		return nil, err
	}
	return list, nil
}

type FrozenDeposit struct {
	Cycle   int64 `json:"cycle"`
	Deposit int64 `json:"deposit,string"`
}

// GetUnstakedFrozenDeposits returns a delegate's unstaked frozen deposits
func (c *Client) GetUnstakedFrozenDeposits(ctx context.Context, addr tezos.Address, id BlockID) ([]FrozenDeposit, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/delegates/%s/unstaked_frozen_deposits", id, addr)
	list := make([]FrozenDeposit, 0, 7)
	if err := c.Get(ctx, u, &list); err != nil {
		return nil, err
	}
	return list, nil
}
