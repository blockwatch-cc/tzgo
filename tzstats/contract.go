// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

type Contract struct {
	RowId         uint64              `json:"row_id"`
	AccountId     uint64              `json:"account_id"`
	Address       tezos.Address       `json:"address"`
	CreatorId     uint64              `json:"creator_id"`
	Creator       tezos.Address       `json:"creator"`
	DelegateId    uint64              `json:"delegate_id"`
	Delegate      tezos.Address       `json:"delegate"`
	FirstSeen     int64               `json:"first_seen"`
	LastSeen      int64               `json:"last_seen"`
	FirstSeenTime time.Time           `json:"first_seen_time"`
	LastSeenTime  time.Time           `json:"last_seen_time"`
	StorageSize   int64               `json:"storage_size"`
	StoragePaid   int64               `json:"storage_paid"`
	Script        *micheline.Script   `json:"script"`
	Storage       *micheline.Prim     `json:"storage"`
	InterfaceHash string              `json:"iface_hash"`
	CodeHash      string              `json:"code_hash"`
	Features      []string            `json:"features"`
	Interfaces    []string            `json:"interfaces"`
	CallStats     map[string]int      `json:"call_stats"`
	BigMaps       map[string]int64    `json:"bigmaps"`
	NOps          int                 `json:"n_ops"`
	NOpsFailed    int                 `json:"n_ops_failed"`
	Metadata      map[string]Metadata `json:"metadata"`

	columns []string `json:"-"`
}

type ContractList struct {
	Contracts []*Contract
	columns   []string
}

func (l *ContractList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if data[0] != '[' {
		return fmt.Errorf("ContractList: expected JSON array")
	}
	array := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	for _, v := range array {
		r := &Contract{
			columns: l.columns,
		}
		if err := r.UnmarshalJSON(v); err != nil {
			return err
		}
		r.columns = nil
		l.Contracts = append(l.Contracts, r)
	}
	return nil
}

func (a *Contract) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if len(data) == 2 {
		return nil
	}
	if data[0] == '[' {
		return a.UnmarshalJSONBrief(data)
	}
	type Alias *Contract
	return json.Unmarshal(data, Alias(a))
}

func (c *Contract) UnmarshalJSONBrief(data []byte) error {
	cc := Contract{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	unpacked := make([]interface{}, 0)
	err := dec.Decode(&unpacked)
	if err != nil {
		return err
	}
	for i, v := range c.columns {
		// var t int64
		f := unpacked[i]
		if f == nil {
			continue
		}
		switch v {
		case "row_id":
			cc.RowId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "account_id":
			cc.AccountId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "address":
			cc.Address, err = tezos.ParseAddress(f.(string))
		case "creator_id":
			cc.CreatorId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "creator":
			cc.Creator, err = tezos.ParseAddress(f.(string))
		case "first_seen":
			cc.FirstSeen, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "last_seen":
			cc.LastSeen, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "first_seen_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				cc.FirstSeenTime = time.Unix(0, ts*1000000).UTC()
			}
		case "last_seen_time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				cc.LastSeenTime = time.Unix(0, ts*1000000).UTC()
			}
		case "storage_size":
			cc.StorageSize, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "storage_paid":
			cc.StoragePaid, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "script":
			var buf []byte
			buf, err = hex.DecodeString(f.(string))
			if err == nil {
				cc.Script = &micheline.Script{}
				err = cc.Script.UnmarshalBinary(buf)
			}
		case "storage":
			var buf []byte
			buf, err = hex.DecodeString(f.(string))
			if err == nil {
				cc.Storage = &micheline.Prim{}
				err = cc.Storage.UnmarshalBinary(buf)
			}
		case "iface_hash":
			cc.InterfaceHash = f.(string)
		case "code_hash":
			cc.CodeHash = f.(string)
		case "call_stats":
			var buf []byte
			buf, err = hex.DecodeString(f.(string))
			if err == nil {
				cc.CallStats = make(map[string]int)
				if cc.Script != nil {
					var eps micheline.Entrypoints
					eps, err = cc.Script.Entrypoints(false)
					for _, ep := range eps {
						if len(buf) < ep.Id*4+4 {
							continue
						}
						cc.CallStats[ep.Call] = int(binary.BigEndian.Uint32(buf[ep.Id*4:]))
					}
				} else {
					for i := 0; i < len(buf); i += 4 {
						cc.CallStats[strconv.Itoa(i/4)] = int(binary.BigEndian.Uint32(buf[i:]))
					}
				}
			}
		case "features":
			cc.Features = strings.Split(f.(string), ",")
		case "interfaces":
			cc.Interfaces = strings.Split(f.(string), ",")
		}
		if err != nil {
			return err
		}
	}
	*c = cc
	return nil
}

type ContractMeta struct {
	Address string    `json:"contract"`
	Time    time.Time `json:"time"`
	Height  int64     `json:"height"`
	Block   string    `json:"block"`
}

type ContractParameters struct {
	Entrypoint string `json:"entrypoint"`
	Call       string `json:"call"`
	Branch     string `json:"branch"`
	Id         int    `json:"id"`
	ContractValue
}

type ContractScript struct {
	StorageType micheline.Typedef     `json:"storage_type"`
	Entrypoints micheline.Entrypoints `json:"entrypoints"`
	Script      *micheline.Script     `json:"script"`
}

type ContractStorage struct {
	Meta ContractMeta `json:"meta"`
	ContractValue
}

type ContractValue struct {
	Value interface{}    `json:"value"`
	Prim  micheline.Prim `json:"prim"`
}

func (v ContractValue) GetString(path string) (string, bool) {
	return getPathString(v.Value, path)
}

func (v ContractValue) GetInt64(path string) (int64, bool) {
	return getPathInt64(v.Value, path)
}

func (v ContractValue) GetBig(path string) (*big.Int, bool) {
	return getPathBig(v.Value, path)
}

func (v ContractValue) GetTime(path string) (time.Time, bool) {
	return getPathTime(v.Value, path)
}

func (v ContractValue) GetAddress(path string) (tezos.Address, bool) {
	return getPathAddress(v.Value, path)
}

func (v ContractValue) GetValue(path string) (interface{}, bool) {
	return getPathValue(v.Value, path)
}

func (v ContractValue) Walk(path string, fn ValueWalkerFunc) error {
	val := v.Value
	if len(path) > 0 {
		var ok bool
		val, ok = getPathValue(val, path)
		if !ok {
			return nil
		}
	}
	return walkValueMap(path, val, fn)
}

type ContractParams struct {
	Params
}

func NewContractParams() ContractParams {
	return ContractParams{NewParams()}
}

func (p ContractParams) WithLimit(v uint) ContractParams {
	p.Query.Set("limit", strconv.Itoa(int(v)))
	return p
}

func (p ContractParams) WithOffset(v uint) ContractParams {
	p.Query.Set("offset", strconv.Itoa(int(v)))
	return p
}

func (p ContractParams) WithCursor(v uint64) ContractParams {
	p.Query.Set("cursor", strconv.FormatUint(v, 10))
	return p
}

func (p ContractParams) WithOrder(v OrderType) ContractParams {
	p.Query.Set("order", string(v))
	return p
}

func (p ContractParams) WithBlock(v string) ContractParams {
	p.Query.Set("block", v)
	return p
}

func (p ContractParams) WithSince(v string) ContractParams {
	p.Query.Set("since", v)
	return p
}

func (p ContractParams) WithUnpack() ContractParams {
	p.Query.Set("unpack", "1")
	return p
}

func (p ContractParams) WithPrim() ContractParams {
	p.Query.Set("prim", "1")
	return p
}

func (p ContractParams) WithMeta() ContractParams {
	p.Query.Set("meta", "1")
	return p
}

func (p ContractParams) WithCollapse() ContractParams {
	p.Query.Set("collapse", "1")
	return p
}

type ContractQuery struct {
	TableQuery
}

func (c *Client) NewContractQuery() ContractQuery {
	tinfo, err := GetTypeInfo(&Contract{}, "")
	if err != nil {
		panic(err)
	}
	q := TableQuery{
		client:  c,
		Params:  c.params.Copy(),
		Table:   "contract",
		Format:  FormatJSON,
		Limit:   DefaultLimit,
		Order:   OrderAsc,
		Columns: tinfo.Aliases(),
		Filter:  make(FilterList, 0),
	}
	return ContractQuery{q}
}

func (q ContractQuery) Run(ctx context.Context) (*ContractList, error) {
	result := &ContractList{
		columns: q.Columns,
	}
	if err := q.client.QueryTable(ctx, q.TableQuery, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) QueryContracts(ctx context.Context, filter FilterList, cols []string) (*ContractList, error) {
	q := c.NewContractQuery()
	if len(cols) > 0 {
		q.Columns = cols
	}
	if len(filter) > 0 {
		q.Filter = filter
	}
	return q.Run(ctx)
}

func (c *Client) GetContract(ctx context.Context, addr string, params ContractParams) (*Contract, error) {
	cc := &Contract{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/contract/%s", addr))
	if err := c.get(ctx, u, nil, cc); err != nil {
		return nil, err
	}
	return cc, nil
}

func (c *Client) GetContractScript(ctx context.Context, addr string, params ContractParams) (*ContractScript, error) {
	cc := &ContractScript{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/contract/%s/script", addr))
	if err := c.get(ctx, u, nil, cc); err != nil {
		return nil, err
	}
	return cc, nil
}

func (c *Client) GetContractStorage(ctx context.Context, addr string, params ContractParams) (*ContractStorage, error) {
	cc := &ContractStorage{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/contract/%s/storage", addr))
	if err := c.get(ctx, u, nil, cc); err != nil {
		return nil, err
	}
	return cc, nil
}

func (c *Client) GetContractCalls(ctx context.Context, addr string, params ContractParams) ([]*Op, error) {
	calls := make([]*Op, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/contract/%s/calls", addr))
	if err := c.get(ctx, u, nil, &calls); err != nil {
		return nil, err
	}
	return calls, nil
}
