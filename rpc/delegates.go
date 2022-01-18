// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"fmt"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

// Delegate holds information about an active delegate
type Delegate struct {
	Delegate             tezos.Address   `json:"-"`
	Block                string          `json:"-"`
	Deactivated          bool            `json:"deactivated"`
	Balance              int64           `json:"balance,string"`
	DelegatedContracts   []tezos.Address `json:"delegated_contracts"`
	FrozenBalance        int64           `json:"frozen_balance,string"`
	FrozenBalanceByCycle []CycleBalance  `json:"frozen_balance_by_cycle"`
	GracePeriod          int64           `json:"grace_period"`
	StakingBalance       int64           `json:"staking_balance,string"`
}

type CycleBalance struct {
	Cycle   int64 `json:"cycle"`
	Deposit int64 `json:"deposit,string"`
	Fees    int64 `json:"fees,string"`
	Rewards int64 `json:"rewards,string"`
}

// DelegateList contains a list of delegates
type DelegateList []tezos.Address

// ListActiveDelegates returns information about all active delegates at a block.
func (c *Client) ListActiveDelegates(ctx context.Context, id BlockID) (DelegateList, error) {
	delegates := make(DelegateList, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/context/delegates?active=true", id)
	if err := c.Get(ctx, u, &delegates); err != nil {
		return nil, err
	}
	return delegates, nil
}

// ListActiveDelegatesWithRolls returns information about all active delegates at a block
// who have at least one roll.
func (c *Client) ListActiveDelegatesWithRolls(ctx context.Context, id BlockID) (DelegateList, error) {
	delegates := make(DelegateList, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/active_delegates_with_rolls", id)
	if err := c.Get(ctx, u, &delegates); err != nil {
		return nil, err
	}
	return delegates, nil
}

// GetDelegate returns information about a delegate at a specific height.
func (c *Client) GetDelegate(ctx context.Context, addr tezos.Address, id BlockID) (*Delegate, error) {
	delegate := &Delegate{
		Delegate: addr,
		Block:    id.String(),
	}
	u := fmt.Sprintf("chains/main/blocks/%s/context/delegates/%s", id, addr)
	if err := c.Get(ctx, u, &delegate); err != nil {
		return nil, err
	}
	return delegate, nil
}

// GetDelegateBalance returns a delegate's balance
func (c *Client) GetDelegateBalance(ctx context.Context, addr tezos.Address, id BlockID) (int64, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/delegates/%s/balance", id, addr)
	var bal string
	err := c.Get(ctx, u, &bal)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(bal, 10, 64)
}
