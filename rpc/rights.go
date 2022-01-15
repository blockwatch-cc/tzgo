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
	Priority      int           `json:"priority"`
	EstimatedTime time.Time     `json:"estimated_time"`
}

func (r BakingRight) Address() tezos.Address {
	return r.Delegate
}

// EndorsingRight holds information about the right to endorse a specific Tezos block
type EndorsingRight struct {
	Delegate      tezos.Address `json:"delegate"`
	Level         int64         `json:"level"`
	EstimatedTime time.Time     `json:"estimated_time"`
	Slots         []int         `json:"slots"`
}

func (r EndorsingRight) Address() tezos.Address {
	return r.Delegate
}

type SnapshotIndex struct {
	LastRoll     []string `json:"last_roll"`
	Nonces       []string `json:"nonces"`
	RandomSeed   string   `json:"random_seed"`
	RollSnapshot int64    `json:"roll_snapshot"`
	Cycle        int64    `json:"cycle"`
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
func (c *Client) ListBakingRights(ctx context.Context, id BlockID) ([]BakingRight, error) {
	rights := make([]BakingRight, 0, 64)
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/baking_rights?all=true&max_priority=63", id)
	if err := c.Get(ctx, u, &rights); err != nil {
		return nil, err
	}
	return rights, nil
}

// ListBakingRightsCycle returns information about baking rights for an entire cycle
// as seen from block id. Note block and cycle must be no further than preserved cycles
// away.
func (c *Client) ListBakingRightsCycle(ctx context.Context, id BlockID, cycle int64) ([]BakingRight, error) {
	rights := make([]BakingRight, 0, 64*4096)
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/baking_rights?all=true&cycle=%d&max_priority=63", id, cycle)
	if err := c.Get(ctx, u, &rights); err != nil {
		return nil, err
	}
	return rights, nil
}

// ListEndorsingRights returns information about block endorsing rights.
func (c *Client) ListEndorsingRights(ctx context.Context, id BlockID) ([]EndorsingRight, error) {
	rights := make([]EndorsingRight, 0, 256)
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/endorsing_rights?all=true", id)
	if err := c.Get(ctx, u, &rights); err != nil {
		return nil, err
	}
	return rights, nil
}

// ListEndorsingRightsCycle returns information about endorsing rights for an entire cycle
// as seen from block id. Note block and cycle must be no further than preserved cycles
// away.
func (c *Client) ListEndorsingRightsCycle(ctx context.Context, id BlockID, cycle int64) ([]EndorsingRight, error) {
	rights := make([]EndorsingRight, 0, 256*8192)
	u := fmt.Sprintf("chains/main/blocks/%s/helpers/endorsing_rights?all=true&cycle=%d", id, cycle)
	if err := c.Get(ctx, u, &rights); err != nil {
		return nil, err
	}
	return rights, nil
}

// GetSnapshotIndexCycle returns information about a roll snapshot as seen from block id.
// Note block and cycle must be no further than preserved cycles away.
func (c *Client) GetSnapshotIndexCycle(ctx context.Context, id BlockID, cycle int64) (*SnapshotIndex, error) {
	idx := &SnapshotIndex{Cycle: cycle}
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/cycle/%d", id, cycle)
	if err := c.Get(ctx, u, idx); err != nil {
		return nil, err
	}
	if idx.RandomSeed == "" {
		return nil, fmt.Errorf("missing snapshot for cycle %d at block %s", cycle, id)
	}
	return idx, nil
}

// ListSnapshotRollOwners returns information about a roll snapshot ownership.
// Response is a nested array `[[roll_id, pubkey]]`.
func (c *Client) ListSnapshotRollOwners(ctx context.Context, id BlockID, cycle, index int64) (*SnapshotOwners, error) {
	owners := &SnapshotOwners{Cycle: cycle, Index: index}
	u := fmt.Sprintf("chains/main/blocks/%s/context/raw/json/rolls/owner/snapshot/%d/%d?depth=1", id, cycle, index)
	if err := c.Get(ctx, u, &owners.Rolls); err != nil {
		return nil, err
	}
	return owners, nil
}
