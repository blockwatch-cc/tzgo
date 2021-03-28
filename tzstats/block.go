// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

type Block struct {
	RowId               uint64                 `json:"row_id"`
	ParentId            uint64                 `json:"parent_id"`
	ParentHash          tezos.BlockHash        `json:"predecessor"`
	FollowerHash        tezos.BlockHash        `json:"successor"`
	Hash                tezos.BlockHash        `json:"hash"`
	IsOrphan            bool                   `json:"is_orphan"`
	Height              int64                  `json:"height"`
	Cycle               int64                  `json:"cycle"`
	IsCycleSnapshot     bool                   `json:"is_cycle_snapshot"`
	Timestamp           time.Time              `json:"time"`
	Solvetime           int                    `json:"solvetime"`
	Version             int                    `json:"version"`
	Validation          int                    `json:"validation_pass"`
	Fitness             uint64                 `json:"fitness"`
	Priority            int                    `json:"priority"`
	Nonce               uint64                 `json:"nonce"`
	VotingPeriodKind    tezos.VotingPeriodKind `json:"voting_period_kind"`
	BakerId             uint64                 `json:"baker_id"`
	Baker               tezos.Address          `json:"baker"`
	SlotsEndorsed       uint32                 `json:"endorsed_slots"`
	NSlotsEndorsed      int                    `json:"n_endorsed_slots"`
	NOps                int                    `json:"n_ops"`
	NOpsFailed          int                    `json:"n_ops_failed"`
	NOpsContract        int                    `json:"n_ops_contract"`
	NOpsImplicit        int                    `json:"n_ops_implicit"`
	NTx                 int                    `json:"n_tx"`
	NActivation         int                    `json:"n_activation"`
	NSeedNonce          int                    `json:"n_seed_nonce_revelation"`
	N2Baking            int                    `json:"n_double_baking_evidence"`
	N2Endorsement       int                    `json:"n_double_endorsement_evidence"`
	NEndorsement        int                    `json:"n_endorsement"`
	NDelegation         int                    `json:"n_delegation"`
	NReveal             int                    `json:"n_reveal"`
	NOrigination        int                    `json:"n_origination"`
	NProposal           int                    `json:"n_proposal"`
	NBallot             int                    `json:"n_ballot"`
	Volume              int64                  `json:"volume"`
	Fee                 int64                  `json:"fee"`
	Reward              int64                  `json:"reward"`
	Deposit             int64                  `json:"deposit"`
	UnfrozenFees        int64                  `json:"unfrozen_fees"`
	UnfrozenRewards     int64                  `json:"unfrozen_rewards"`
	UnfrozenDeposits    int64                  `json:"unfrozen_deposits"`
	ActivatedSupply     int64                  `json:"activated_supply"`
	BurnedSupply        int64                  `json:"burned_supply"`
	SeenAccounts        int                    `json:"n_accounts"`
	NewAccounts         int                    `json:"n_new_accounts"`
	NewImplicitAccounts int                    `json:"n_new_implicit"`
	NewManagedAccounts  int                    `json:"n_new_managed"`
	NewContracts        int                    `json:"n_new_contracts"`
	ClearedAccounts     int                    `json:"n_cleared_accounts"`
	FundedAccounts      int                    `json:"n_funded_accounts"`
	GasLimit            int64                  `json:"gas_limit"`
	GasUsed             int64                  `json:"gas_used"`
	GasPrice            float64                `json:"gas_price"`
	StorageSize         int64                  `json:"storage_size"`
	TDD                 float64                `json:"days_destroyed"`
	PctAccountReuse     float64                `json:"pct_account_reuse"`
	Metadata            map[string]Metadata    `json:"metadata"`
	Rights              []Right                `json:"rights"`
	Ops                 []*Op                  `json:"ops"`
	columns             []string               `json:"-"`
}

type BlockList struct {
	Blocks  []*Block
	columns []string
}

func (l *BlockList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if data[0] != '[' {
		return fmt.Errorf("BlockList: expected JSON array")
	}
	array := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	for _, v := range array {
		r := &Block{
			columns: l.columns,
		}
		if err := r.UnmarshalJSON(v); err != nil {
			return err
		}
		r.columns = nil
		l.Blocks = append(l.Blocks, r)
	}
	return nil
}

func (b *Block) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if len(data) == 2 {
		return nil
	}
	if data[0] == '[' {
		return b.UnmarshalJSONBrief(data)
	}
	type Alias *Block
	return json.Unmarshal(data, Alias(b))
}

func (b *Block) UnmarshalJSONBrief(data []byte) error {
	block := Block{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	unpacked := make([]interface{}, 0)
	err := dec.Decode(&unpacked)
	if err != nil {
		return err
	}
	for i, v := range b.columns {
		f := unpacked[i]
		if f == nil {
			continue
		}
		switch v {
		case "row_id":
			block.RowId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "parent_id":
			block.ParentId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "hash":
			block.Hash, err = tezos.ParseBlockHash(f.(string))
		case "is_orphan":
			block.IsOrphan, err = strconv.ParseBool(f.(json.Number).String())
		case "height":
			block.Height, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "cycle":
			block.Cycle, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "is_cycle_snapshot":
			block.IsCycleSnapshot, err = strconv.ParseBool(f.(json.Number).String())
		case "time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				block.Timestamp = time.Unix(0, ts*1000000).UTC()
			}
		case "solvetime":
			block.Solvetime, err = strconv.Atoi(f.(json.Number).String())
		case "version":
			block.Version, err = strconv.Atoi(f.(json.Number).String())
		case "validation_pass":
			block.Validation, err = strconv.Atoi(f.(json.Number).String())
		case "fitness":
			block.Fitness, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "priority":
			block.Priority, err = strconv.Atoi(f.(json.Number).String())
		case "nonce":
			block.Nonce, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "voting_period_kind":
			block.VotingPeriodKind = tezos.ParseVotingPeriod(f.(string))
		case "baker_id":
			block.BakerId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "endorsed_slots":
			var i uint64
			i, err = strconv.ParseUint(f.(json.Number).String(), 10, 32)
			block.SlotsEndorsed = uint32(i)
		case "n_endorsed_slots":
			block.NSlotsEndorsed, err = strconv.Atoi(f.(json.Number).String())
		case "n_ops":
			block.NOps, err = strconv.Atoi(f.(json.Number).String())
		case "n_ops_failed":
			block.NOpsFailed, err = strconv.Atoi(f.(json.Number).String())
		case "n_ops_contract":
			block.NOpsContract, err = strconv.Atoi(f.(json.Number).String())
		case "n_ops_implicit":
			block.NOpsImplicit, err = strconv.Atoi(f.(json.Number).String())
		case "n_tx":
			block.NTx, err = strconv.Atoi(f.(json.Number).String())
		case "n_activation":
			block.NActivation, err = strconv.Atoi(f.(json.Number).String())
		case "n_seed_nonce_revelation":
			block.NSeedNonce, err = strconv.Atoi(f.(json.Number).String())
		case "n_double_baking_evidence":
			block.N2Baking, err = strconv.Atoi(f.(json.Number).String())
		case "n_double_endorsement_evidence":
			block.N2Endorsement, err = strconv.Atoi(f.(json.Number).String())
		case "n_endorsement":
			block.NEndorsement, err = strconv.Atoi(f.(json.Number).String())
		case "n_delegation":
			block.NDelegation, err = strconv.Atoi(f.(json.Number).String())
		case "n_reveal":
			block.NReveal, err = strconv.Atoi(f.(json.Number).String())
		case "n_origination":
			block.NOrigination, err = strconv.Atoi(f.(json.Number).String())
		case "n_proposal":
			block.NProposal, err = strconv.Atoi(f.(json.Number).String())
		case "n_ballot":
			block.NBallot, err = strconv.Atoi(f.(json.Number).String())
		case "volume":
			block.Volume, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "fee":
			block.Fee, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "reward":
			block.Reward, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "deposit":
			block.Deposit, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "unfrozen_fees":
			block.UnfrozenFees, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "unfrozen_rewards":
			block.UnfrozenRewards, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "unfrozen_deposits":
			block.UnfrozenDeposits, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "activated_supply":
			block.ActivatedSupply, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "burned_supply":
			block.BurnedSupply, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "n_accounts":
			block.SeenAccounts, err = strconv.Atoi(f.(json.Number).String())
		case "n_new_accounts":
			block.NewAccounts, err = strconv.Atoi(f.(json.Number).String())
		case "n_new_implicit":
			block.NewImplicitAccounts, err = strconv.Atoi(f.(json.Number).String())
		case "n_new_managed":
			block.NewManagedAccounts, err = strconv.Atoi(f.(json.Number).String())
		case "n_new_contracts":
			block.NewContracts, err = strconv.Atoi(f.(json.Number).String())
		case "n_cleared_accounts":
			block.ClearedAccounts, err = strconv.Atoi(f.(json.Number).String())
		case "n_funded_accounts":
			block.FundedAccounts, err = strconv.Atoi(f.(json.Number).String())
		case "gas_limit":
			block.GasLimit, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "gas_used":
			block.GasUsed, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "gas_price":
			block.GasPrice, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "storage_size":
			block.StorageSize, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "days_destroyed":
			block.TDD, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "pct_account_reuse":
			block.PctAccountReuse, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "baker":
			block.Baker, err = tezos.ParseAddress(f.(string))
		}
		if err != nil {
			return err
		}
	}
	*b = block
	return nil
}

type BlockQuery struct {
	TableQuery
}

func (c *Client) NewBlockQuery() BlockQuery {
	tinfo, err := GetTypeInfo(&Block{}, "")
	if err != nil {
		panic(err)
	}
	q := TableQuery{
		client:  c,
		Params:  c.params.Copy(),
		Table:   "block",
		Format:  FormatJSON,
		Limit:   DefaultLimit,
		Columns: tinfo.Aliases(),
		Order:   OrderAsc,
		Filter:  make(FilterList, 0),
	}
	return BlockQuery{q}
}

func (q BlockQuery) Run(ctx context.Context) (*BlockList, error) {
	result := &BlockList{
		columns: q.Columns,
	}
	if err := q.client.QueryTable(ctx, q.TableQuery, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) QueryBlocks(ctx context.Context, filter FilterList, cols []string) (*BlockList, error) {
	q := c.NewBlockQuery()
	if len(cols) > 0 {
		q.Columns = cols
	}
	if len(filter) > 0 {
		q.Filter = filter
	}
	return q.Run(ctx)
}

type BlockParams struct {
	Params
}

func NewBlockParams() BlockParams {
	return BlockParams{NewParams()}
}

func (p BlockParams) WithLimit(v uint) BlockParams {
	p.Query.Set("limit", strconv.Itoa(int(v)))
	return p
}

func (p BlockParams) WithOffset(v uint) BlockParams {
	p.Query.Set("offset", strconv.Itoa(int(v)))
	return p
}

func (p BlockParams) WithCursor(v uint64) BlockParams {
	p.Query.Set("cursor", strconv.FormatUint(v, 10))
	return p
}

func (p BlockParams) WithOrder(v OrderType) BlockParams {
	p.Query.Set("order", string(v))
	return p
}

func (p BlockParams) WithMeta() BlockParams {
	p.Query.Set("meta", "1")
	return p
}

func (p BlockParams) WithRights() BlockParams {
	p.Query.Set("rights", "1")
	return p
}

func (c *Client) GetBlock(ctx context.Context, hash string, params BlockParams) (*Block, error) {
	b := &Block{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/block/%s", hash))
	if err := c.get(ctx, u, nil, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) GetBlockWithOps(ctx context.Context, hash string, params BlockParams) (*Block, error) {
	b := &Block{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/block/%s/op", hash))
	if err := c.get(ctx, u, nil, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) GetBlockOps(ctx context.Context, hash string, params OpParams) ([]*Op, error) {
	ops := make([]*Op, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/block/%s/operations", hash))
	if err := c.get(ctx, u, nil, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}
