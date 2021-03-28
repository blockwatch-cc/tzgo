// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"context"

	"blockwatch.cc/tzgo/tezos"
)

type Metadata struct {
	Address     tezos.Address `json:"address"`
	Name        string        `json:"name"`
	Category    string        `json:"category,omitempty"`
	Status      string        `json:"status,omitempty"`
	Country     string        `json:"country,omitempty"`
	City        string        `json:"city,omitempty"`
	Twitter     string        `json:"twitter,omitempty"`
	HasLogo     bool          `json:"logo,omitempty"`
	IsSponsored bool          `json:"sponsored,omitempty"`

	// baker info
	Fee            float64         `json:"fee,omitempty"`
	From           []tezos.Address `json:"from,omitempty"`
	PayoutDelay    bool            `json:"payout_delay,omitempty"`
	MinPayout      float64         `json:"min_payout,omitempty"`
	MinDelegation  float64         `json:"min_delegation,omitempty"`
	NonDelegatable bool            `json:"non_delegatable,omitempty"`

	// token info, type is custom, multisig, harbinger, dexter, fa12, tzbtc, staker
	Standard string `json:"standard,omitempty"`
	Code     string `json:"code,omitempty"`
	Decimals int    `json:"decimals,omitempty"`
}

func (c *Client) ListMetadata(ctx context.Context) ([]Metadata, error) {
	resp := make([]Metadata, 0)
	if err := c.get(ctx, "/explorer/metadata", nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetMetadata(ctx context.Context, addr string) (Metadata, error) {
	var resp Metadata
	if err := c.get(ctx, "/explorer/metadata/"+addr, nil, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) CreateMetadata(ctx context.Context, metadata []Metadata) error {
	return c.post(ctx, "/explorer/metadata", nil, &metadata, nil)
}

func (c *Client) UpdateMetadata(ctx context.Context, alias Metadata) (Metadata, error) {
	var resp Metadata
	if err := c.put(ctx, "/explorer/metadata/"+alias.Address.String(), nil, &alias, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func (c *Client) RemoveMetadata(ctx context.Context, addr string) error {
	return c.delete(ctx, "/explorer/metadata/"+addr, nil)
}

func (c *Client) PurgeMetadata(ctx context.Context) error {
	return c.delete(ctx, "/explorer/metadata", nil)
}
