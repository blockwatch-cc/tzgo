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
	DefaultParams = (&Params{
		MinimalBlockDelay:            15 * time.Second,
		CostPerByte:                  250,
		OriginationSize:              257,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		HardStorageLimitPerOperation: 60000,
		MaxOperationDataLength:       32768,
		MaxOperationsTTL:             240,
	}).
		WithChainId(Mainnet).
		WithDeployment(Deployments[Mainnet].AtProtocol(ProtoV016_2))

	// GhostnetParams defines the blockchain configuration for Ghostnet testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	GhostnetParams = (&Params{
		MinimalBlockDelay:            8 * time.Second,
		CostPerByte:                  250,
		OriginationSize:              257,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		HardStorageLimitPerOperation: 60000,
		MaxOperationDataLength:       32768,
		MaxOperationsTTL:             240,
	}).
		WithChainId(Ghostnet).
		WithDeployment(Deployments[Ghostnet].AtProtocol(ProtoV016_2))

	// NairobinetParams defines the blockchain configuration for Nairobi testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	NairobinetParams = (&Params{
		MinimalBlockDelay:            8 * time.Second,
		CostPerByte:                  250,
		OriginationSize:              257,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		HardStorageLimitPerOperation: 60000,
		MaxOperationDataLength:       32768,
		MaxOperationsTTL:             240,
	}).
		WithChainId(Nairobinet).
		WithDeployment(Deployments[Nairobinet].AtProtocol(ProtoV017))

	// OxfordnetParams defines the blockchain configuration for Oxford testnet.
	// To produce compliant transactions, use these defaults in op.WithParams().
	OxfordnetParams = (&Params{
		MinimalBlockDelay:            8 * time.Second,
		CostPerByte:                  250,
		OriginationSize:              257,
		HardGasLimitPerOperation:     1040000,
		HardGasLimitPerBlock:         2600000,
		HardStorageLimitPerOperation: 60000,
		MaxOperationDataLength:       32768,
		MaxOperationsTTL:             240,
	}).
		WithChainId(Oxfordnet).
		WithDeployment(Deployments[Oxfordnet].AtProtocol(ProtoV018))
)

// Params contains a subset of protocol configuration settings that are relevant
// for dapps and most indexers. For additional protocol data, call rpc.GetCustomConstants()
// with a custom data struct.
type Params struct {
	// identity
	Network  string       `json:"network,omitempty"`
	ChainId  ChainIdHash  `json:"chain_id"`
	Protocol ProtocolHash `json:"protocol"`
	Version  int          `json:"version"`

	// timing
	MinimalBlockDelay time.Duration `json:"minimal_block_delay"`

	// costs
	CostPerByte     int64 `json:"cost_per_byte"`
	OriginationSize int64 `json:"origination_size"`

	// limits
	BlocksPerCycle               int64 `json:"blocks_per_cycle"`
	PreservedCycles              int64 `json:"preserved_cycles"`
	BlocksPerSnapshot            int64 `json:"blocks_per_snapshot"`
	HardGasLimitPerOperation     int64 `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock         int64 `json:"hard_gas_limit_per_block"`
	HardStorageLimitPerOperation int64 `json:"hard_storage_limit_per_operation"`
	MaxOperationDataLength       int   `json:"max_operation_data_length"`
	MaxOperationsTTL             int64 `json:"max_operations_ttl"`

	// extra features to follow protocol upgrades
	OperationTagsVersion int   `json:"operation_tags_version,omitempty"` // 1 after v005
	StartHeight          int64 `json:"start_height"`                     // protocol start (may be != cycle start!!)
	EndHeight            int64 `json:"end_height"`                       // protocol end (may be != cycle end!!)
	StartOffset          int64 `json:"start_offset"`                     // correction for cycle start
	StartCycle           int64 `json:"start_cycle"`                      // correction cycle length
}

func NewParams() *Params {
	return &Params{
		Network:     "unknown",
		StartHeight: 1<<63 - 1,
	}
}

func (p Params) Clone() *Params {
	np := p
	return &np
}

func (p *Params) WithChainId(id ChainIdHash) *Params {
	p.ChainId = id
	if p.Network == "unknown" || p.Network == "" {
		switch id {
		case Mainnet:
			p.Network = "Mainnet"
		case Ghostnet:
			p.Network = "Ghostnet"
		case Nairobinet:
			p.Network = "Nairobinet"
		case Oxfordnet:
			p.Network = "Oxfordnet"
		}
	}
	return p
}

func (p *Params) WithProtocol(h ProtocolHash) *Params {
	var ok bool
	p.Protocol = h
	p.Version, ok = Versions[h]
	if !ok {
		var max int
		for _, v := range Versions {
			if v < max {
				continue
			}
			max = v
		}
		p.Version = max + 1
		Versions[h] = p.Version
	}
	switch {
	case p.Version > 11:
		p.OperationTagsVersion = 2
	case p.Version > 4:
		p.OperationTagsVersion = 1
	}
	return p
}

func (p *Params) WithNetwork(n string) *Params {
	if p.Network == "unknown" || p.Network == "" {
		p.Network = n
	}
	return p
}

func (p *Params) WithDeployment(d Deployment) *Params {
	if d.BlocksPerCycle > 0 {
		p.WithProtocol(d.Protocol)
		p.StartOffset = d.StartOffset
		p.StartHeight = d.StartHeight
		p.EndHeight = d.EndHeight
		p.StartCycle = d.StartCycle
		p.PreservedCycles = d.PreservedCycles
		p.BlocksPerCycle = d.BlocksPerCycle
		p.BlocksPerSnapshot = d.BlocksPerSnapshot
	}
	return p
}

func (p *Params) WithBlock(height int64) *Params {
	d := Deployments[p.ChainId].AtBlock(height)
	p.StartOffset = d.StartOffset
	p.StartHeight = d.StartHeight
	p.EndHeight = d.EndHeight
	p.StartCycle = d.StartCycle
	return p
}

func (p *Params) AtBlock(height int64) *Params {
	if p.ContainsHeight(height) {
		return p
	}
	return p.Clone().WithDeployment(Deployments[p.ChainId].AtBlock(height))
}

func (p *Params) AtCycle(cycle int64) *Params {
	if p.ContainsCycle(cycle) {
		return p
	}
	return p.Clone().WithDeployment(Deployments[p.ChainId].AtCycle(cycle))
}

func (p Params) SnapshotBaseCycle(cycle int64) int64 {
	var offset int64 = 2
	if p.Version >= 12 {
		offset = 1
	}
	return cycle - (p.PreservedCycles + offset)
}

func (p Params) IsMainnet() bool {
	return p.ChainId.Equal(Mainnet)
}

// Note: functions below require StartHeight, EndHeight and/or StartCycle!
func (p Params) ContainsHeight(height int64) bool {
	// treat -1 as special height query that matches open interval params only
	return (height < 0 && p.EndHeight < 0) ||
		(p.StartHeight <= height && (p.EndHeight < 0 || p.EndHeight >= height))
}

func (p Params) ContainsCycle(c int64) bool {
	// FIX granada early start
	s := p.StartCycle
	if c == 387 && p.IsMainnet() {
		s--
	}
	return s <= c
}

func (p *Params) CycleFromHeight(height int64) int64 {
	// adjust to target height
	at := p.AtBlock(height)

	// FIX granada early start
	s := at.StartCycle
	if height == 1589248 && at.IsMainnet() {
		s--
	}
	return s + (height-(at.StartHeight-at.StartOffset))/at.BlocksPerCycle
}

func (p *Params) CycleStartHeight(c int64) int64 {
	// adjust to target cycle
	at := p.AtCycle(c)
	return at.StartHeight - at.StartOffset + (c-at.StartCycle)*at.BlocksPerCycle
}

func (p *Params) CycleEndHeight(c int64) int64 {
	// adjust to target cycle
	at := p.AtCycle(c)
	return at.CycleStartHeight(c) + at.BlocksPerCycle - 1
}

func (p *Params) CyclePosition(height int64) int64 {
	// adjust to target height
	at := p.AtBlock(height)
	pos := (height - (at.StartHeight - at.StartOffset)) % at.BlocksPerCycle
	if pos < 0 {
		pos += at.BlocksPerCycle
	}
	return pos
}

func (p *Params) IsCycleStart(height int64) bool {
	return height > 0 && (height == 1 || p.CyclePosition(height) == 0)
}

func (p *Params) IsCycleEnd(height int64) bool {
	// adjust to target height
	at := p.AtBlock(height)
	return at.CyclePosition(height)+1 == at.BlocksPerCycle
}

func (p *Params) IsSnapshotBlock(height int64) bool {
	// adjust to target height
	at := p.AtBlock(height)
	pos := at.CyclePosition(height) + 1
	return pos > 0 && pos%at.BlocksPerSnapshot == 0
}

func (p *Params) SnapshotBlock(cycle int64, index int) int64 {
	// adjust to target cycle
	at := p.AtCycle(cycle)
	base := at.SnapshotBaseCycle(cycle)
	if base < 0 {
		return 0
	}
	return at.CycleStartHeight(base) + int64(index+1)*at.BlocksPerSnapshot - 1
}

func (p *Params) SnapshotIndex(height int64) int {
	// FIX granada early start
	if height == 1589248 && p.IsMainnet() {
		return 15
	}
	// adjust to target height
	at := p.AtBlock(height)
	return int((at.CyclePosition(height)+1)/at.BlocksPerSnapshot) - 1
}
