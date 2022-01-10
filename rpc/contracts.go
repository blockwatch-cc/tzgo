// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"fmt"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// Contracts holds a list of addresses
type Contracts []tezos.Address

// Contracts holds info about a Tezos account
type ContractInfo struct {
	Balance  int64         `json:"balance,string"`
	Delegate tezos.Address `json:"delegate"`
	Counter  int64         `json:"counter,string"`
}

// GetContract returns the full info about a contract at block id
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-contracts-contract-id
func (c *Client) GetContract(ctx context.Context, addr tezos.Address, id BlockID) (*ContractInfo, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", id, addr)
	var info ContractInfo
	err := c.Get(ctx, u, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// ListContracts returns a list of all known contracts at head
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-contracts
func (c *Client) ListContracts(ctx context.Context, id BlockID) (Contracts, error) {
	contracts := make(Contracts, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts", id)
	if err := c.Get(ctx, u, &contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// GetContractScript returns the originated contract script
func (c *Client) GetContractScript(ctx context.Context, addr tezos.Address) (*micheline.Script, error) {
	u := fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/script", addr)
	s := micheline.NewScript()
	err := c.Get(ctx, u, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetContractStorage returns the most recent version of the contract's storage
func (c *Client) GetContractStorage(ctx context.Context, addr tezos.Address, id BlockID) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/storage", id, addr)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetContractEntrypoints returns the contract's entrypoints
func (c *Client) GetContractEntrypoints(ctx context.Context, addr tezos.Address) (map[string]micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/entrypoints", addr)
	type eptype struct {
		Entrypoints map[string]micheline.Prim `json:"entrypoints"`
	}
	eps := &eptype{}
	err := c.Get(ctx, u, eps)
	if err != nil {
		return nil, err
	}
	return eps.Entrypoints, nil
}

// GetBigmapKeys returns all keys in the bigmap b at block id
func (c *Client) GetBigmapKeys(ctx context.Context, bigmap int64, id BlockID) ([]tezos.ExprHash, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/big_maps/index/%d/contents", id, bigmap)
	hashes := make([]tezos.ExprHash, 0)
	err := c.Get(ctx, u, &hashes)
	if err != nil {
		return nil, err
	}
	return hashes, nil
}

// GetActiveBigmapKeys returns all active keys in the bigmap
func (c *Client) GetActiveBigmapKeys(ctx context.Context, bigmap int64) ([]tezos.ExprHash, error) {
	return c.GetBigmapKeys(ctx, bigmap, Head)
}

// GetBigmapValue returns value at key hash from bigmap at block id
func (c *Client) GetBigmapValue(ctx context.Context, bigmap int64, hash tezos.ExprHash, id BlockID) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/big_maps/index/%d/contents/%s", id, bigmap, hash)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetActiveBigmapValue returns current active value at key hash from bigmap
func (c *Client) GetActiveBigmapValue(ctx context.Context, bigmap int64, hash tezos.ExprHash) (micheline.Prim, error) {
	return c.GetBigmapValue(ctx, bigmap, hash, Head)
}

type BigmapInfo struct {
	KeyType    micheline.Prim `json:"key_type"`
	ValueType  micheline.Prim `json:"value_type"`
	TotalBytes int64          `json:"total_bytes,string"`
}

// GetActiveBigmapInfo returns type and content info from bigmap at current head
func (c *Client) GetActiveBigmapInfo(ctx context.Context, bigmap int64) (*BigmapInfo, error) {
	return c.GetBigmapInfo(ctx, bigmap, Head)
}

// GetBigmapInfo returns type and content info from bigmap at block id
func (c *Client) GetBigmapInfo(ctx context.Context, bigmap int64, id BlockID) (*BigmapInfo, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/big_maps/index/%d", id, bigmap)
	info := &BigmapInfo{}
	err := c.Get(ctx, u, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
