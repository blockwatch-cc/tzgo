// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

// BakingRight holds information about the right to bake a specific Tezos block
type BakingRight struct {
	Delegate      tezos.Address `json:"delegate"`
	Level         int64         `json:"level"`
	Priority      int           `json:"priority"` // until v011
	Round         int           `json:"round"`    // v012+
	EstimatedTime time.Time     `json:"estimated_time"`
}

func (r BakingRight) Address() tezos.Address {
	return r.Delegate
}

// EndorsingRight holds information about the right to endorse a specific Tezos block
// valid for all protocols up to v001
type EndorsingRight struct {
	Delegate      tezos.Address `json:"delegate"`
	Level         int64         `json:"level"`
	EstimatedTime time.Time     `json:"estimated_time"`
	Slots         []int         `json:"slots,omitempty"` // until v011
	FirstSlot     int           `json:"first_slot"`      // v012+
	Power         int           `json:"endorsing_power"` // v012+
}

func (r EndorsingRight) Address() tezos.Address {
	return r.Delegate
}

type RollSnapshotInfo struct {
	LastRoll     []string `json:"last_roll"`
	Nonces       []string `json:"nonces"`
	RandomSeed   string   `json:"random_seed"`
	RollSnapshot int      `json:"roll_snapshot"`
}

type StakeInfo struct {
	ActiveStake int64         `json:"active_stake,string"`
	Baker       tezos.Address `json:"baker"`
}

// v012+
type StakingSnapshotInfo struct {
	Nonces           []string    `json:"nonces"`
	RandomSeed       string      `json:"random_seed"`
	BakerStake       []StakeInfo `json:"selected_stake_distribution,omitempty"`
	TotalActiveStake int64       `json:"total_active_stake,string"`
	// SlashedDeposits  []??       `json:"slashed_deposits"`
}

type SnapshotIndex struct {
	Cycle int64 // the requested cycle that contains rights from the snapshot
	Base  int64 // the cycle where the snapshot happened
	Index int   // the index inside base where snapshot happened
}

type SnapshotRoll struct {
	RollId   int64
	OwnerKey tezos.Key
}

func (r *SnapshotRoll) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if len(data) == 2 {
		return nil
	}
	if data[0] != '[' || data[len(data)-1] != ']' {
		return fmt.Errorf("SnapshotRoll: invalid json array '%s'", string(data))
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	unpacked := make([]interface{}, 0)
	err := dec.Decode(&unpacked)
	if err != nil {
		return err
	}
	return r.decode(unpacked)
}

func (r SnapshotRoll) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 2048)
	buf = append(buf, '[')
	buf = strconv.AppendInt(buf, r.RollId, 10)
	buf = append(buf, ',')
	buf = strconv.AppendQuote(buf, r.OwnerKey.String())
	buf = append(buf, ']')
	return buf, nil
}

func (r *SnapshotRoll) decode(unpacked []interface{}) error {
	if l := len(unpacked); l != 2 {
		return fmt.Errorf("SnapshotRoll: invalid json array len %d", l)
	}
	id, err := strconv.ParseInt(unpacked[0].(json.Number).String(), 10, 64)
	if err != nil {
		return fmt.Errorf("SnapshotRoll: invalid roll id: %w", err)
	}
	if err = r.OwnerKey.UnmarshalText([]byte(unpacked[1].(string))); err != nil {
		return err
	}
	r.RollId = id
	return nil
}

type SnapshotOwners struct {
	Cycle int64          `json:"cycle"`
	Index int64          `json:"index"`
	Rolls []SnapshotRoll `json:"rolls"`
}

// ListBakingRights returns information about baking rights at block id.
// Use max to set a max block priority (before Ithaca) or a max round (after Ithaca).
func (c *Client) ListBakingRights(ctx context.Context, id BlockID, max int) ([]BakingRight, error) {
	p, err := c.GetParams(ctx, id)
	if err != nil {
		return nil, err
	}
	maxSelector := "max_priority=%d"
	if p.Version >= 12 {
		maxSelector = "max_round=%d"
	}
	if p.Version < 6 {
		max++
	}
	rights := make([]BakingRight, 0)
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/baking_rights?all=true&"+maxSelector, id, max)
	if err := c.Get(ctx, u, &rights); err != nil {
		return nil, err
	}
	return rights, nil
}

// ListBakingRightsCycle returns information about baking rights for an entire cycle
// as seen from block id. Note block and cycle must be no further than preserved cycles
// away from each other. Use max to set a max block priority (before Ithaca) or a max
// round (after Ithaca).
func (c *Client) ListBakingRightsCycle(ctx context.Context, id BlockID, cycle int64, max int) ([]BakingRight, error) {
	p, err := c.GetParams(ctx, id)
	if err != nil {
		return nil, err
	}
	maxSelector := "max_priority=%d"
	if p.Version >= 12 {
		maxSelector = "max_round=%d"
	}
	if p.Version < 6 {
		max++
	}
	rights := make([]BakingRight, 0, (max+1)*int(p.BlocksPerCycle))
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/baking_rights?all=true&cycle=%d&"+maxSelector, id, cycle, max)
	if err := c.Get(ctx, u, &rights); err != nil {
		return nil, err
	}
	return rights, nil
}

// ListEndorsingRights returns information about block endorsing rights.
func (c *Client) ListEndorsingRights(ctx context.Context, id BlockID) ([]EndorsingRight, error) {
	p, err := c.GetParams(ctx, id)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/endorsing_rights?all=true", id)
	rights := make([]EndorsingRight, 0, (p.EndorsersPerBlock + p.ConsensusCommitteeSize))
	// Note: future cycles are seen from current protocol (!)
	if p.Version >= 12 && p.StartHeight <= id.Int64() {
		type V12Rights struct {
			Level         int64            `json:"level"`
			Delegates     []EndorsingRight `json:"delegates"`
			EstimatedTime time.Time        `json:"estimated_time"`
		}
		v12rights := make([]V12Rights, 0, 1)
		if err := c.Get(ctx, u, &v12rights); err != nil {
			return nil, err
		}
		for _, v := range v12rights {
			for _, r := range v.Delegates {
				r.Level = v.Level
				r.EstimatedTime = v.EstimatedTime
				rights = append(rights, r)
			}
		}

	} else {
		if err := c.Get(ctx, u, &rights); err != nil {
			return nil, err
		}
	}
	return rights, nil
}

// ListEndorsingRightsCycle returns information about endorsing rights for an entire cycle
// as seen from block id. Note block and cycle must be no further than preserved cycles
// away.
func (c *Client) ListEndorsingRightsCycle(ctx context.Context, id BlockID, cycle int64) ([]EndorsingRight, error) {
	p, err := c.GetParams(ctx, id)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/endorsing_rights?all=true&cycle=%d", id, cycle)
	rights := make([]EndorsingRight, 0, (p.EndorsersPerBlock+p.ConsensusCommitteeSize)*int(p.BlocksPerCycle))
	// Note: future cycles are seen from current protocol (!)
	if p.Version >= 12 && p.StartHeight <= id.Int64() {
		type V12Rights struct {
			Level         int64            `json:"level"`
			Delegates     []EndorsingRight `json:"delegates"`
			EstimatedTime time.Time        `json:"estimated_time"`
		}
		v12rights := make([]V12Rights, 0, p.BlocksPerCycle)
		if err := c.Get(ctx, u, &v12rights); err != nil {
			return nil, err
		}
		for _, v := range v12rights {
			for _, r := range v.Delegates {
				r.Level = v.Level
				r.EstimatedTime = v.EstimatedTime
				rights = append(rights, r)
			}
		}
	} else {
		if err := c.Get(ctx, u, &rights); err != nil {
			return nil, err
		}
	}
	return rights, nil
}

// GetRollSnapshotInfoCycle returns information about a roll snapshot as seen from block id.
// Note block and cycle must be no further than preserved cycles away.
func (c *Client) GetRollSnapshotInfoCycle(ctx context.Context, id BlockID, cycle int64) (*RollSnapshotInfo, error) {
	idx := &RollSnapshotInfo{}
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/cycle/%d", id, cycle)
	if err := c.Get(ctx, u, idx); err != nil {
		return nil, err
	}
	if idx.RandomSeed == "" {
		return nil, fmt.Errorf("missing snapshot for cycle %d at block %s", cycle, id)
	}
	return idx, nil
}

// GetStakingSnapshotInfoCycle returns information about a roll snapshot as seen from block id.
// Note block and cycle must be no further than preserved cycles away.
func (c *Client) GetStakingSnapshotInfoCycle(ctx context.Context, id BlockID, cycle int64) (*StakingSnapshotInfo, error) {
	idx := &StakingSnapshotInfo{}
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/cycle/%d", id, cycle)
	if err := c.Get(ctx, u, idx); err != nil {
		return nil, err
	}
	return idx, nil
}

// GetSnapshotIndexCycle returns information about a roll or staking snapshot that
// produced rights at cycle.
// Note block and cycle must be no further than preserved cycles away.
func (c *Client) GetSnapshotIndexCycle(ctx context.Context, cycle int64) (*SnapshotIndex, error) {
	var id BlockID
	if c.Params == nil {
		id = Head
	} else {
		id = BlockLevel(c.Params.CycleStartHeight(cycle))
	}
	p, err := c.GetParams(ctx, id)
	if err != nil {
		return nil, err
	}
	idx := &SnapshotIndex{}
	if p.Version >= 12 {
		// post-Ithaca we can no longer look ahead N cycles, i.e. query the snapshot
		// that just happened at N-1, instead we can at most query cycle
		// N-PRESERVED_CYCLES-1. This is a limitation of the new Tezos RPC which
		// allows no cycle argument.
		//
		// The second important change happened to the protocol. Now the distance
		// between snapshot and rights is no longer N-PRESEREVD-2, instead its -1!
		//
		// Tezos docu says
		// Add ``/chains/main/blocks/<block>/context/selected_snapshot`` RPC to
		// retrieve the snapshot index used to compute baking right for the
		// given block's cycle. Context entry at
		// ``/chains/main/blocks/<block>/context/raw/bytes/cycle/<cycle>/roll_snapshot``
		// are no longer accessible in Tenderbake.
		//
		// ref: https://gitlab.com/tezos/tezos/-/merge_requests/4479/diffs#a164b2628ca1dfd5974b47defa53b423b606f373_166_167
		u := fmt.Sprintf("chains/main/blocks/%s/context/selected_snapshot", id)
		if err := c.Get(ctx, u, &idx.Index); err != nil {
			return nil, err
		}
		idx.Cycle = cycle
		idx.Base = cycle
	} else {
		// pre-Ithaca we can at most look PRESERVED_CYCLES into the future since
		// the snapshot happened 2 cycles back from the block we're looking from.
		base := p.SnapshotBaseCycle(cycle)
		if base < 0 {
			base = cycle
		}
		id = BlockLevel(p.CycleStartHeight(base))
		var info RollSnapshotInfo
		u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/cycle/%d", id, cycle)
		if err := c.Get(ctx, u, &info); err != nil {
			return nil, err
		}
		if info.RandomSeed == "" {
			return nil, fmt.Errorf("missing snapshot for cycle %d at block %s", cycle, id)
		}
		idx.Cycle = cycle
		idx.Base = base
		idx.Index = info.RollSnapshot
	}
	return idx, nil
}

// ListSnapshotRollOwners returns information about a roll snapshot ownership.
// Response is a nested array `[[roll_id, pubkey]]`. Deprecated in Ithaca.
func (c *Client) ListSnapshotRollOwners(ctx context.Context, id BlockID, cycle, index int64) (*SnapshotOwners, error) {
	owners := &SnapshotOwners{Cycle: cycle, Index: index}
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/rolls/owner/snapshot/%d/%d?depth=1", id, cycle, index)
	if err := c.Get(ctx, u, &owners.Rolls); err != nil {
		return nil, err
	}
	return owners, nil
}
