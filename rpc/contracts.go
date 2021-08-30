// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"fmt"
	"strconv"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

// Contracts holds a list of addresses
type Contracts []tezos.Address

// GetContracts returns a list of all known contracts at head
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-contracts
func (c *Client) GetContracts(ctx context.Context) (Contracts, error) {
	contracts := make(Contracts, 0)
	u := fmt.Sprintf("chains/%s/blocks/head/context/contracts", c.ChainID)
	if err := c.Get(ctx, u, &contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// GetContractsHeight returns a list of all known contracts at height
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-contracts
func (c *Client) GetContractsHeight(ctx context.Context, height int64) (Contracts, error) {
	u := fmt.Sprintf("chains/%s/blocks/%d/context/contracts", c.ChainID, height)
	contracts := make(Contracts, 0)
	if err := c.Get(ctx, u, &contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// GetContractBalance returns the current balance of a contract at head
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-contracts-contract-id-balance
func (c *Client) GetContractBalance(ctx context.Context, addr tezos.Address) (int64, error) {
	u := fmt.Sprintf("chains/%s/blocks/head/context/contracts/%s/balance", c.ChainID, addr)
	var bal string
	err := c.Get(ctx, u, &bal)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(bal, 10, 64)
}

// GetContractBalanceHeight returns the current balance of a contract at height
// https://tezos.gitlab.io/tezos/api/rpc.html#get-block-id-context-contracts-contract-id-balance
func (c *Client) GetContractBalanceHeight(ctx context.Context, addr tezos.Address, height int64) (int64, error) {
	u := fmt.Sprintf("chains/%s/blocks/%d/context/contracts/%s/balance", c.ChainID, height, addr)
	var bal string
	err := c.Get(ctx, u, &bal)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(bal, 10, 64)
}

// GetContractScript returns the originated contract script
func (c *Client) GetContractScript(ctx context.Context, addr tezos.Address) (*micheline.Script, error) {
	u := fmt.Sprintf("chains/%s/blocks/head/context/contracts/%s/script", c.ChainID, addr)
	s := micheline.NewScript()
	err := c.Get(ctx, u, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetContractStorage returns the most recent version of the contract's storage
func (c *Client) GetContractStorage(ctx context.Context, addr tezos.Address) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/%s/blocks/head/context/contracts/%s/storage", c.ChainID, addr)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetContractStorage returns the contract's storage at height
func (c *Client) GetContractStorageHeight(ctx context.Context, addr tezos.Address, height int64) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/%s/blocks/%d/context/contracts/%s/storage", c.ChainID, height, addr)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetContractEntrypoints returns the contract's entrypoints 
func (c *Client) GetContractEntrypoints(ctx context.Context, addr tezos.Address) (map[string]micheline.Prim, error) {
	u := fmt.Sprintf("chains/%s/blocks/head/context/contracts/%s/entrypoints", c.ChainID, addr)
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

// GetBigmapKeys returns all active keys in the bigmap id
func (c *Client) GetBigmapKeys(ctx context.Context, id int64) ([]tezos.ExprHash, error) {
	u := fmt.Sprintf("chains/%s/blocks/head/context/raw/json/big_maps/index/%d/contents", c.ChainID, id)
	hashes := make([]tezos.ExprHash, 0)
	err := c.Get(ctx, u, &hashes)
	if err != nil {
		return nil, err
	}
	return hashes, nil
}

// GetBigmapValue returns current active value at key hash from bigmap id
func (c *Client) GetBigmapValue(ctx context.Context, id int64, hash tezos.ExprHash) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/%s/blocks/head/context/raw/json/big_maps/index/%d/contents/%s", c.ChainID, id, hash)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetBigmapValueHeight returns a value from bigmap id at key hash that was active at height
func (c *Client) GetBigmapValueHeight(ctx context.Context, id int64, hash tezos.ExprHash, height int64) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/%s/blocks/%d/context/raw/json/big_maps/index/%d/contents/%s", c.ChainID, height, id, hash)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

type BigmapInfo struct {
	KeyType    micheline.Prim `json:"key_type"`
	ValueType  micheline.Prim `json:"value_type"`
	TotalBytes int64          `json:"total_bytes,string"`
}

// GetBigmapInfo returns type and content info from bigmap id
func (c *Client) GetBigmapInfo(ctx context.Context, id int64) (*BigmapInfo, error) {
	u := fmt.Sprintf("chains/%s/blocks/head/context/raw/json/big_maps/index/%d", c.ChainID, id)
	info := &BigmapInfo{}
	err := c.Get(ctx, u, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
