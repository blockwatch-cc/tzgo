// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"time"
)

var (
	// DefaultParams defines the blockchain configuration for Mainnet under the latest
	// protocol. It is used to generate compliant transaction encodings. To change,
	// either overwrite this default or set custom params per operation using
	// op.WithParams().
	DefaultParams = &Params{
		Network:                      "Mainnet",
		ChainId:                      Mainnet,
		Protocol:                     ProtoV016_2,
		Version:                      16,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             240,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            15 * time.Second,
		PreservedCycles:              5,
	}

	// GhostnetParams defines the blockchain configuration for Ghostnet testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	GhostnetParams = &Params{
		Network:                      "Ghostnet",
		ChainId:                      Ghostnet,
		Protocol:                     ProtoV016_2,
		Version:                      16,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             240,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            15 * time.Second,
		PreservedCycles:              3,
	}

	// LimanetParams defines the blockchain configuration for Lima testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	LimanetParams = &Params{
		Network:                      "Limanet",
		ChainId:                      Limanet,
		Protocol:                     ProtoV015,
		Version:                      15,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             120,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         5200000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            15 * time.Second,
		PreservedCycles:              3,
	}

	// MumbainetParams defines the blockchain configuration for Mumbai testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	MumbainetParams = &Params{
		Network:                      "Mumbainet",
		ChainId:                      Mumbainet,
		Protocol:                     ProtoV016_2,
		Version:                      16,
		OperationTagsVersion:         2,
		MaxOperationsTTL:             240,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		OriginationSize:              257,
		CostPerByte:                  250,
		HardStorageLimitPerOperation: 60000,
		MinimalBlockDelay:            8 * time.Second,
		PreservedCycles:              3,
	}
)

// Params contains a subset of protocol configuration settings that are relevant
// for dapps and most indexers. For additional protocol data, call rpc.GetCustomConstants()
// with a custom data struct.
type Params struct {
	// identity
	Network  string       `json:"network,omitempty"`
	Version  int          `json:"version"`
	ChainId  ChainIdHash  `json:"chain_id"`
	Protocol ProtocolHash `json:"protocol"`

	// timing
	MinimalBlockDelay time.Duration `json:"minimal_block_delay"`

	// costs
	CostPerByte     int64 `json:"cost_per_byte"`
	OriginationSize int64 `json:"origination_size"`

	// limits
	BlocksPerCycle               int64 `json:"blocks_per_cycle"`
	PreservedCycles              int64 `json:"preserved_cycles"`
	HardGasLimitPerOperation     int64 `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock         int64 `json:"hard_gas_limit_per_block"`
	HardStorageLimitPerOperation int64 `json:"hard_storage_limit_per_operation"`
	MaxOperationDataLength       int   `json:"max_operation_data_length"`
	MaxOperationsTTL             int64 `json:"max_operations_ttl"`

	// extra features to follow protocol upgrades
	OperationTagsVersion int `json:"operation_tags_version,omitempty"` // 1 after v005
}

func (p *Params) WithChainId(id ChainIdHash) *Params {
	p.ChainId = id
	return p
}

func (p *Params) WithProtocol(h ProtocolHash) *Params {
	p.Protocol = h
	return p
}

func (p *Params) WithVersion(v int) *Params {
	p.Version = v
	switch {
	case v > 11:
		p.OperationTagsVersion = 2
	case v > 4:
		p.OperationTagsVersion = 1
	}
	return p
}

func (p *Params) WithNetwork(n string) *Params {
	p.Network = n
	return p
}

func (p Params) SnapshotBaseCycle(cycle int64) int64 {
	var offset int64 = 2
	if p.Version >= 12 {
		offset = 1
	}
	return cycle - (p.PreservedCycles + offset)
}
