// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"fmt"

	"github.com/legonian/tzgo/micheline"
	"github.com/legonian/tzgo/tezos"
)

// UnparsingMode defines the way types and values are represented in Micheline script
// and storage. This affects timestamps, keys, addresses, signatures and nested pairs.
// Optimized encodings use integers for timestamps and bytes instead of base58 encoded
// values. Legacy mode is supposed to output 2-ary pairs only, but is messed up on
// certain endpoints (e.g. /script/normalized), so there's no guarantee.
//
type UnparsingMode string

const (
	UnparsingModeInvalid   = ""
	UnparsingModeLegacy    = "Optimized_legacy"
	UnparsingModeOptimized = "Optimized"
	UnparsingModeReadable  = "Readable"
)

func (m UnparsingMode) String() string {
	return string(m)
}

// Contracts holds a list of addresses
type Contracts []tezos.Address

// Contracts holds info about a Tezos account
type ContractInfo struct {
	Balance  int64         `json:"balance,string"`
	Delegate tezos.Address `json:"delegate"`
	Counter  int64         `json:"counter,string"`
	Manager  string        `json:"manager"`
}

func (i ContractInfo) IsRevealed() bool {
	_, ok := tezos.ParseKeyType(i.Manager)
	return ok
}

func (i ContractInfo) ManagerKey() tezos.Key {
	key, _ := tezos.ParseKey(i.Manager)
	return key
}

// GetContract returns info about an account at block id.
func (c *Client) GetContract(ctx context.Context, addr tezos.Address, id BlockID) (*ContractInfo, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s", id, addr)
	var info ContractInfo
	err := c.Get(ctx, u, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetManagerKey returns the revealed public key of an account at block id.
func (c *Client) GetManagerKey(ctx context.Context, addr tezos.Address, id BlockID) (tezos.Key, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/manager_key", id, addr)
	var key tezos.Key
	err := c.Get(ctx, u, &key)
	return key, err
}

// GetContractExt returns info about an account at block id including its public key when revealed.
func (c *Client) GetContractExt(ctx context.Context, addr tezos.Address, id BlockID) (*ContractInfo, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/contracts/index/%s", id, addr)
	var info ContractInfo
	err := c.Get(ctx, u, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// ListContracts returns a list of all known contracts at head. This call may be very SLOW for
// large chains and there is no means to limit the result. Use with caution and consider
// calling an indexer API instead.
func (c *Client) ListContracts(ctx context.Context, id BlockID) (Contracts, error) {
	contracts := make(Contracts, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts", id)
	if err := c.Get(ctx, u, &contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// GetContractScript returns the originated contract script in default data mode.
func (c *Client) GetContractScript(ctx context.Context, addr tezos.Address) (*micheline.Script, error) {
	u := fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/script", addr)
	s := micheline.NewScript()
	err := c.Get(ctx, u, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetNormalizedScript returns the originated contract script with global constants
// expanded using given unparsing mode.
func (c *Client) GetNormalizedScript(ctx context.Context, addr tezos.Address, mode UnparsingMode) (*micheline.Script, error) {
	u := fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/script/normalized", addr)
	s := micheline.NewScript()
	if mode == "" {
		mode = UnparsingModeOptimized
	}
	postData := struct {
		Mode UnparsingMode `json:"unparsing_mode"`
	}{
		Mode: mode,
	}
	err := c.Post(ctx, u, &postData, s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// GetContractStorage returns the contract's storage at block id.
func (c *Client) GetContractStorage(ctx context.Context, addr tezos.Address, id BlockID) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/storage", id, addr)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetContractStorageNormalized returns contract's storage at block id using unparsing mode.
func (c *Client) GetContractStorageNormalized(ctx context.Context, addr tezos.Address, id BlockID, mode UnparsingMode) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/contracts/%s/storage/normalized", id, addr)
	if mode == "" {
		mode = UnparsingModeOptimized
	}
	postData := struct {
		Mode UnparsingMode `json:"unparsing_mode"`
	}{
		Mode: mode,
	}
	prim := micheline.Prim{}
	err := c.Post(ctx, u, &postData, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetContractEntrypoints returns the contract's entrypoints.
func (c *Client) GetContractEntrypoints(ctx context.Context, addr tezos.Address) (map[string]micheline.Type, error) {
	u := fmt.Sprintf("chains/main/blocks/head/context/contracts/%s/entrypoints", addr)
	type eptype struct {
		Entrypoints map[string]micheline.Type `json:"entrypoints"`
	}
	eps := &eptype{}
	err := c.Get(ctx, u, eps)
	if err != nil {
		return nil, err
	}
	return eps.Entrypoints, nil
}

// ListBigmapKeys returns all keys in the bigmap at block id. This call may be very SLOW for
// large bigmaps and there is no means to limit the result. Use with caution and consider
// calling an indexer API instead.
func (c *Client) ListBigmapKeys(ctx context.Context, bigmap int64, id BlockID) ([]tezos.ExprHash, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/big_maps/index/%d/contents", id, bigmap)
	hashes := make([]tezos.ExprHash, 0)
	err := c.Get(ctx, u, &hashes)
	if err != nil {
		return nil, err
	}
	return hashes, nil
}

// ListActiveBigmapKeys returns all active keys in the bigmap. This call may be very SLOW for
// large bigmaps and there is no means to limit the result. Use with caution and consider
// calling an indexer API instead.
func (c *Client) ListActiveBigmapKeys(ctx context.Context, bigmap int64) ([]tezos.ExprHash, error) {
	return c.ListBigmapKeys(ctx, bigmap, Head)
}

// GetBigmapValue returns value at key hash from bigmap at block id
func (c *Client) GetBigmapValue(ctx context.Context, bigmap int64, hash tezos.ExprHash, id BlockID) (micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/big_maps/%d/%s", id, bigmap, hash)
	prim := micheline.Prim{}
	err := c.Get(ctx, u, &prim)
	if err != nil {
		return micheline.InvalidPrim, err
	}
	return prim, nil
}

// GetActiveBigmapValue returns current active value at key hash from bigmap.
func (c *Client) GetActiveBigmapValue(ctx context.Context, bigmap int64, hash tezos.ExprHash) (micheline.Prim, error) {
	return c.GetBigmapValue(ctx, bigmap, hash, Head)
}

// ListBigmapValues returns all values from bigmap at block id. This call may be very SLOW for
// large bigmaps and there is no means to limit the result. Use with caution and consider
// calling an indexer API instead.
func (c *Client) ListBigmapValues(ctx context.Context, bigmap int64, id BlockID) ([]micheline.Prim, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/big_maps/%d", id, bigmap)
	vals := make([]micheline.Prim, 0)
	err := c.Get(ctx, u, &vals)
	if err != nil {
		return nil, err
	}
	return vals, nil
}

// ListActiveBigmapValues returns all values from bigmap at block id. This call may be very SLOW for
// large bigmaps and there is no means to limit the result. Use with caution and consider
// calling an indexer API instead.
func (c *Client) ListActiveBigmapValues(ctx context.Context, bigmap int64, id BlockID) ([]micheline.Prim, error) {
	return c.ListBigmapValues(ctx, bigmap, Head)
}

type BigmapInfo struct {
	KeyType    micheline.Prim `json:"key_type"`
	ValueType  micheline.Prim `json:"value_type"`
	TotalBytes int64          `json:"total_bytes,string"`
}

// GetActiveBigmapInfo returns type and content info from bigmap at current head.
func (c *Client) GetActiveBigmapInfo(ctx context.Context, bigmap int64) (*BigmapInfo, error) {
	return c.GetBigmapInfo(ctx, bigmap, Head)
}

// GetBigmapInfo returns type and content info from bigmap at block id.
func (c *Client) GetBigmapInfo(ctx context.Context, bigmap int64, id BlockID) (*BigmapInfo, error) {
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/big_maps/index/%d", id, bigmap)
	info := &BigmapInfo{}
	err := c.Get(ctx, u, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}
