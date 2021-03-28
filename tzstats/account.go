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

type Account struct {
	RowId              uint64              `json:"row_id"`
	Address            tezos.Address       `json:"address"`
	AddressType        tezos.AddressType   `json:"address_type"`
	DelegateId         uint64              `json:"delegate_id"`
	Delegate           tezos.Address       `json:"delegate"`
	CreatorId          uint64              `json:"creator_id"`
	Creator            tezos.Address       `json:"creator"`
	Pubkey             tezos.Key           `json:"pubkey"`
	FirstIn            int64               `json:"first_in"`
	FirstOut           int64               `json:"first_out"`
	FirstSeen          int64               `json:"first_seen"`
	LastIn             int64               `json:"last_in"`
	LastOut            int64               `json:"last_out"`
	LastSeen           int64               `json:"last_seen"`
	DelegatedSince     int64               `json:"delegated_since"`
	DelegateSince      int64               `json:"delegate_since"`
	DelegateUntil      int64               `json:"delegate_until"`
	FirstSeenTime      time.Time           `json:"first_seen_time"`
	LastSeenTime       time.Time           `json:"last_seen_time"`
	FirstInTime        time.Time           `json:"first_in_time"`
	LastInTime         time.Time           `json:"last_in_time"`
	FirstOutTime       time.Time           `json:"first_out_time"`
	LastOutTime        time.Time           `json:"last_out_time"`
	DelegatedSinceTime time.Time           `json:"delegated_since_time"`
	DelegateSinceTime  time.Time           `json:"delegate_since_time"`
	DelegateUntilTime  time.Time           `json:"delegate_until_time"`
	TotalReceived      float64             `json:"total_received"`
	TotalSent          float64             `json:"total_sent"`
	TotalBurned        float64             `json:"total_burned"`
	TotalFeesPaid      float64             `json:"total_fees_paid"`
	TotalRewardsEarned float64             `json:"total_rewards_earned"`
	TotalFeesEarned    float64             `json:"total_fees_earned"`
	TotalLost          float64             `json:"total_lost"`
	FrozenDeposits     float64             `json:"frozen_deposits"`
	FrozenRewards      float64             `json:"frozen_rewards"`
	FrozenFees         float64             `json:"frozen_fees"`
	UnclaimedBalance   float64             `json:"unclaimed_balance"`
	SpendableBalance   float64             `json:"spendable_balance"`
	DelegatedBalance   float64             `json:"delegated_balance"`
	TotalDelegations   int64               `json:"total_delegations"`
	ActiveDelegations  int64               `json:"active_delegations"`
	IsFunded           bool                `json:"is_funded"`
	IsActivated        bool                `json:"is_activated"`
	IsDelegated        bool                `json:"is_delegated"`
	IsRevealed         bool                `json:"is_revealed"`
	IsDelegate         bool                `json:"is_delegate"`
	IsActiveDelegate   bool                `json:"is_active_delegate"`
	IsContract         bool                `json:"is_contract"`
	BlocksBaked        int                 `json:"blocks_baked"`
	BlocksMissed       int                 `json:"blocks_missed"`
	BlocksStolen       int                 `json:"blocks_stolen"`
	BlocksEndorsed     int                 `json:"blocks_endorsed"`
	SlotsEndorsed      int                 `json:"slots_endorsed"`
	SlotsMissed        int                 `json:"slots_missed"`
	NOps               int                 `json:"n_ops"`
	NOpsFailed         int                 `json:"n_ops_failed"`
	NTx                int                 `json:"n_tx"`
	NDelegation        int                 `json:"n_delegation"`
	NOrigination       int                 `json:"n_origination"`
	NProposal          int                 `json:"n_proposal"`
	NBallot            int                 `json:"n_ballot"`
	TokenGenMin        int64               `json:"token_gen_min"`
	TokenGenMax        int64               `json:"token_gen_max"`
	GracePeriod        int64               `json:"grace_period"`
	StakingBalance     float64             `json:"staking_balance"`
	StakingCapacity    float64             `json:"staking_capacity"`
	Rolls              int64               `json:"rolls"`
	LastBakeHeight     int64               `json:"last_bake_height"`
	LastBakeBlock      string              `json:"last_bake_block"`
	LastBakeTime       time.Time           `json:"last_bake_time"`
	LastEndorseHeight  int64               `json:"last_endorse_height"`
	LastEndorseBlock   string              `json:"last_endorse_block"`
	LastEndorseTime    time.Time           `json:"last_endorse_time"`
	NextBakeHeight     int64               `json:"next_bake_height"`
	NextBakePriority   int                 `json:"next_bake_priority"`
	NextBakeTime       time.Time           `json:"next_bake_time"`
	NextEndorseHeight  int64               `json:"next_endorse_height"`
	NextEndorseTime    time.Time           `json:"next_endorse_time"`
	AvgLuck64          int64               `json:"avg_luck_64"`
	AvgPerformance64   int64               `json:"avg_performance_64"`
	AvgContribution64  int64               `json:"avg_contribution_64"`
	BakerVersion       string              `json:"baker_version"`
	Metadata           map[string]Metadata `json:"metadata"`
	columns            []string            `json:"-"`
}

type AccountList struct {
	Accounts []*Account
	columns  []string
}

func (l *AccountList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if data[0] != '[' {
		return fmt.Errorf("AccountList: expected JSON array")
	}
	array := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	for _, v := range array {
		r := &Account{
			columns: l.columns,
		}
		if err := r.UnmarshalJSON(v); err != nil {
			return err
		}
		r.columns = nil
		l.Accounts = append(l.Accounts, r)
	}
	return nil
}

func (a *Account) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if len(data) == 2 {
		return nil
	}
	if data[0] == '[' {
		return a.UnmarshalJSONBrief(data)
	}
	type Alias *Account
	return json.Unmarshal(data, Alias(a))
}

func (a *Account) UnmarshalJSONBrief(data []byte) error {
	acc := Account{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	unpacked := make([]interface{}, 0)
	err := dec.Decode(&unpacked)
	if err != nil {
		return err
	}
	for i, v := range a.columns {
		f := unpacked[i]
		if f == nil {
			continue
		}
		switch v {
		case "row_id":
			acc.RowId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "address":
			acc.Address, err = tezos.ParseAddress(f.(string))
		case "address_type":
			acc.AddressType = tezos.ParseAddressType(f.(string))
		case "delegate_id":
			acc.DelegateId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "delegate":
			acc.Delegate, err = tezos.ParseAddress(f.(string))
		case "creator_id":
			acc.CreatorId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "creator":
			acc.Creator, err = tezos.ParseAddress(f.(string))
		case "pubkey":
			acc.Pubkey, err = tezos.ParseKey(f.(string))
		case "first_in":
			acc.FirstIn, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "first_out":
			acc.FirstOut, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "first_seen":
			acc.FirstSeen, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "last_in":
			acc.LastIn, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "last_out":
			acc.LastOut, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "last_seen":
			acc.LastSeen, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "delegated_since":
			acc.DelegatedSince, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "delegate_since":
			acc.DelegateSince, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "delegate_until":
			acc.DelegateUntil, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "total_received":
			acc.TotalReceived, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "total_sent":
			acc.TotalSent, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "total_burned":
			acc.TotalBurned, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "total_fees_paid":
			acc.TotalFeesPaid, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "total_rewards_earned":
			acc.TotalRewardsEarned, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "total_fees_earned":
			acc.TotalFeesEarned, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "total_lost":
			acc.TotalLost, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "frozen_deposits":
			acc.FrozenDeposits, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "frozen_rewards":
			acc.FrozenRewards, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "frozen_fees":
			acc.FrozenFees, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "unclaimed_balance":
			acc.UnclaimedBalance, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "spendable_balance":
			acc.SpendableBalance, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "delegated_balance":
			acc.DelegatedBalance, err = strconv.ParseFloat(f.(json.Number).String(), 64)
		case "total_delegations":
			acc.TotalDelegations, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "active_delegations":
			acc.ActiveDelegations, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "is_funded":
			acc.IsFunded, err = strconv.ParseBool(f.(json.Number).String())
		case "is_activated":
			acc.IsActivated, err = strconv.ParseBool(f.(json.Number).String())
		case "is_delegated":
			acc.IsDelegated, err = strconv.ParseBool(f.(json.Number).String())
		case "is_revealed":
			acc.IsRevealed, err = strconv.ParseBool(f.(json.Number).String())
		case "is_active_delegate":
			acc.IsActiveDelegate, err = strconv.ParseBool(f.(json.Number).String())
		case "is_delegate":
			acc.IsDelegate, err = strconv.ParseBool(f.(json.Number).String())
		case "is_contract":
			acc.IsContract, err = strconv.ParseBool(f.(json.Number).String())
		case "blocks_baked":
			acc.BlocksBaked, err = strconv.Atoi(f.(json.Number).String())
		case "blocks_missed":
			acc.BlocksMissed, err = strconv.Atoi(f.(json.Number).String())
		case "blocks_stolen":
			acc.BlocksStolen, err = strconv.Atoi(f.(json.Number).String())
		case "blocks_endorsed":
			acc.BlocksEndorsed, err = strconv.Atoi(f.(json.Number).String())
		case "slots_endorsed":
			acc.SlotsEndorsed, err = strconv.Atoi(f.(json.Number).String())
		case "slots_missed":
			acc.SlotsMissed, err = strconv.Atoi(f.(json.Number).String())
		case "n_ops":
			acc.NOps, err = strconv.Atoi(f.(json.Number).String())
		case "n_ops_failed":
			acc.NOpsFailed, err = strconv.Atoi(f.(json.Number).String())
		case "n_tx":
			acc.NTx, err = strconv.Atoi(f.(json.Number).String())
		case "n_delegation":
			acc.NDelegation, err = strconv.Atoi(f.(json.Number).String())
		case "n_origination":
			acc.NOrigination, err = strconv.Atoi(f.(json.Number).String())
		case "n_proposal":
			acc.NProposal, err = strconv.Atoi(f.(json.Number).String())
		case "n_ballot":
			acc.NBallot, err = strconv.Atoi(f.(json.Number).String())
		case "token_gen_min":
			acc.TokenGenMin, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "token_gen_max":
			acc.TokenGenMax, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "grace_period":
			acc.GracePeriod, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "first_seen_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.FirstSeenTime = time.Unix(0, ts*1000000).UTC()
			}
		case "last_seen_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.LastSeenTime = time.Unix(0, ts*1000000).UTC()
			}
		case "first_in_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.FirstInTime = time.Unix(0, ts*1000000).UTC()
			}
		case "last_in_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.LastInTime = time.Unix(0, ts*1000000).UTC()
			}
		case "first_out_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.FirstOutTime = time.Unix(0, ts*1000000).UTC()
			}
		case "last_out_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.LastOutTime = time.Unix(0, ts*1000000).UTC()
			}
		case "delegated_since_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.DelegatedSinceTime = time.Unix(0, ts*1000000).UTC()
			}
		case "delegate_since_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.DelegateSinceTime = time.Unix(0, ts*1000000).UTC()
			}
		case "delegate_until_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				acc.DelegateUntilTime = time.Unix(0, ts*1000000).UTC()
			}
		}
		if err != nil {
			return err
		}
	}
	*a = acc
	return nil
}

type AccountParams struct {
	Params
}

func NewAccountParams() AccountParams {
	return AccountParams{NewParams()}
}

func (p AccountParams) WithLimit(v uint) AccountParams {
	p.Query.Set("limit", strconv.Itoa(int(v)))
	return p
}

func (p AccountParams) WithOffset(v uint) AccountParams {
	p.Query.Set("offset", strconv.Itoa(int(v)))
	return p
}

func (p AccountParams) WithCursor(v uint64) AccountParams {
	p.Query.Set("cursor", strconv.FormatUint(v, 10))
	return p
}

func (p AccountParams) WithOrder(v OrderType) AccountParams {
	p.Query.Set("order", string(v))
	return p
}

func (p AccountParams) WithMeta() AccountParams {
	p.Query.Set("meta", "1")
	return p
}

type AccountQuery struct {
	TableQuery
}

func (c *Client) NewAccountQuery() AccountQuery {
	tinfo, err := GetTypeInfo(&Account{}, "")
	if err != nil {
		panic(err)
	}
	q := TableQuery{
		client:  c,
		Params:  c.params.Copy(),
		Table:   "account",
		Format:  FormatJSON,
		Limit:   DefaultLimit,
		Order:   OrderAsc,
		Columns: tinfo.Aliases(),
		Filter:  make(FilterList, 0),
	}
	return AccountQuery{q}
}

func (q AccountQuery) Run(ctx context.Context) (*AccountList, error) {
	result := &AccountList{
		columns: q.Columns,
	}
	if err := q.client.QueryTable(ctx, q.TableQuery, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) QueryAccounts(ctx context.Context, filter FilterList, cols []string) (*AccountList, error) {
	q := c.NewAccountQuery()
	if len(cols) > 0 {
		q.Columns = cols
	}
	if len(filter) > 0 {
		q.Filter = filter
	}
	return q.Run(ctx)
}

func (c *Client) GetAccount(ctx context.Context, addr string, params AccountParams) (*Account, error) {
	a := &Account{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/account/%s", addr))
	if err := c.get(ctx, u, nil, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (c *Client) GetAccountContracts(ctx context.Context, addr string, params AccountParams) ([]*Account, error) {
	cc := make([]*Account, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/account/%s/contracts", addr))
	if err := c.get(ctx, u, nil, &cc); err != nil {
		return nil, err
	}
	return cc, nil
}

func (c *Client) GetAccountOps(ctx context.Context, addr string, params OpParams) ([]*Op, error) {
	ops := make([]*Op, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/contract/%s/operations", addr))
	if err := c.get(ctx, u, nil, &ops); err != nil {
		return nil, err
	}
	return ops, nil
}
