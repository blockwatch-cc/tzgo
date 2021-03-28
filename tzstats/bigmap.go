// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tzstats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
)

type Bigmap struct {
	Contract        tezos.Address   `json:"contract"`
	BigMapId        int64           `json:"bigmap_id"`
	NUpdates        int64           `json:"n_updates"`
	NKeys           int64           `json:"n_keys"`
	AllocatedHeight int64           `json:"alloc_height"`
	AllocatedBlock  tezos.BlockHash `json:"alloc_block"`
	AllocatedTime   time.Time       `json:"alloc_time"`
	UpdatedHeight   int64           `json:"update_height"`
	UpdatedBlock    tezos.BlockHash `json:"update_block"`
	UpdatedTime     time.Time       `json:"update_time"`
	IsRemoved       bool            `json:"is_removed"`
	Type            BigmapType      `json:"type"`
}

type BigmapType struct {
	Contract    tezos.Address  `json:"contract"`
	BigMapId    int64          `json:"bigmap_id"`
	KeyEncoding string         `json:"key_encoding"`
	KeyType     interface{}    `json:"key_type"`
	ValueType   interface{}    `json:"value_type"`
	Prim        BigmapTypePrim `json:"prim"`
}

type BigmapTypePrim struct {
	KeyPrim   *micheline.Prim `json:"key_type"`
	ValuePrim *micheline.Prim `json:"value_type"`
}

type BigmapMeta struct {
	Contract     tezos.Address   `json:"contract"`
	BigMapId     int64           `json:"bigmap_id"`
	UpdateTime   time.Time       `json:"time"`
	UpdateHeight int64           `json:"height"`
	UpdateBlock  tezos.BlockHash `json:"block"`
	IsReplaced   bool            `json:"is_replaced"`
	IsRemoved    bool            `json:"is_removed"`
}

type BigmapKey struct {
	Keys         MultiKey        `json:"key"`
	KeyHash      tezos.ExprHash  `json:"key_hash"`
	KeyBinary    string          `json:"key_binary"`
	KeysUnpacked MultiKey        `json:"key_unpacked"`
	KeyPretty    string          `json:"key_pretty"`
	Meta         BigmapMeta      `json:"meta"`
	Prim         *micheline.Prim `json:"prim"`
}

type MultiKey []string

func (k MultiKey) String() string {
	return strings.Join([]string(k), ",")
}

func (k *MultiKey) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	key := make([]string, 0)
	switch buf[0] {
	case '{':
		m := make(map[string]interface{})
		if err := json.Unmarshal(buf, &m); err != nil {
			return err
		}
		for _, v := range m {
			key = append(key, ToString(v))
		}
	case '[':
		m := make([]interface{}, 0)
		if err := json.Unmarshal(buf, &m); err != nil {
			return err
		}
		for _, v := range m {
			key = append(key, ToString(v))
		}
	case '"':
		s, _ := strconv.Unquote(string(buf))
		key = []string{s}
	default:
		key = []string{string(buf)}
	}
	*k = key
	return nil
}

type BigmapValue struct {
	Keys         MultiKey        `json:"key"`
	KeyHash      tezos.ExprHash  `json:"key_hash"`
	KeyBinary    string          `json:"key_binary"`
	KeysUnpacked MultiKey        `json:"key_unpacked"`
	KeyPretty    string          `json:"key_pretty"`
	Meta         BigmapMeta      `json:"meta"`
	Value        interface{}     `json:"value"`
	Unpacked     interface{}     `json:"value_unpacked"`
	Prim         BigmapValuePrim `json:"prim"`
}

type BigmapValuePrim struct {
	KeyPrim   *micheline.Prim `json:"key"`
	ValuePrim *micheline.Prim `json:"value"`
}

func (v BigmapValue) GetString(path string) (string, bool) {
	if v.Unpacked != nil {
		if vv, ok := getPathString(v.Unpacked, path); ok {
			return vv, ok
		}
	}
	return getPathString(v.Value, path)
}

func (v BigmapValue) GetInt(path string) (int64, bool) {
	if v.Unpacked != nil {
		if vv, ok := getPathInt(v.Unpacked, path); ok {
			return vv, ok
		}
	}
	return getPathInt(v.Value, path)
}

func (v BigmapValue) GetBig(path string) (*big.Int, bool) {
	if v.Unpacked != nil {
		if vv, ok := getPathBig(v.Unpacked, path); ok {
			return vv, ok
		}
	}
	return getPathBig(v.Value, path)
}

func (v BigmapValue) GetTime(path string) (time.Time, bool) {
	if v.Unpacked != nil {
		if vv, ok := getPathTime(v.Unpacked, path); ok {
			return vv, ok
		}
	}
	return getPathTime(v.Value, path)
}

func (v BigmapValue) GetAddress(path string) (tezos.Address, bool) {
	if v.Unpacked != nil {
		if vv, ok := getPathAddress(v.Unpacked, path); ok {
			return vv, ok
		}
	}
	return getPathAddress(v.Value, path)
}

func (v BigmapValue) GetValue(path string) (interface{}, bool) {
	if v.Unpacked != nil {
		if vv, ok := getPathValue(v.Unpacked, path); ok {
			return vv, ok
		}
	}
	return getPathValue(v.Value, path)
}

func (v BigmapValue) Walk(path string, fn ValueWalkerFunc) error {
	val, ok := v.Value, v.Value != nil
	if !ok {
		val, ok = v.Unpacked, v.Unpacked != nil
	}
	if len(path) > 0 {
		val, ok = getPathValue(v.Value, path)
	}
	return walkValueMap(path, val, fn)
}

type BigmapUpdate struct {
	BigmapValue
	Action      string      `json:"action"`
	KeyEncoding string      `json:"key_encoding"`
	KeyType     interface{} `json:"key_type"`
	ValueType   interface{} `json:"value_type"`
	SourceId    int64       `json:"source_big_map"`
	DestId      int64       `json:"destination_big_map"`
}

type BigmapRow struct {
	RowId       uint64         `json:"row_id"`
	PrevId      uint64         `json:"prev_id"`
	Address     tezos.Address  `json:"address"`
	AccountId   uint64         `json:"account_id"`
	ContractId  uint64         `json:"contract_id"`
	OpId        uint64         `json:"op_id"`
	Op          tezos.OpHash   `json:"op"`
	Height      int64          `json:"height"`
	Timestamp   time.Time      `json:"time"`
	BigMapId    int64          `json:"bigmap_id"`
	Action      string         `json:"action"`
	KeyHash     tezos.ExprHash `json:"key_hash,omitempty"`
	KeyType     string         `json:"key_type,omitempty"`
	KeyEncoding string         `json:"key_encoding,omitempty"`
	Key         string         `json:"key,omitempty"`
	Value       string         `json:"value,omitempty"`
	IsReplaced  bool           `json:"is_replaced"`
	IsDeleted   bool           `json:"is_deleted"`
	IsCopied    bool           `json:"is_copied"`

	columns []string `json:"-"`
}

type BigmapRowList struct {
	Rows    []*BigmapRow
	columns []string
}

func (l *BigmapRowList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if data[0] != '[' {
		return fmt.Errorf("BigmapRowList: expected JSON array")
	}
	// log.Debugf("decode rights list from %d bytes", len(data))
	array := make([]json.RawMessage, 0)
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	for _, v := range array {
		b := &BigmapRow{
			columns: l.columns,
		}
		if err := b.UnmarshalJSON(v); err != nil {
			return err
		}
		b.columns = nil
		l.Rows = append(l.Rows, b)
	}
	return nil
}

func (b *BigmapRow) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || bytes.Compare(data, []byte("null")) == 0 {
		return nil
	}
	if len(data) == 2 {
		return nil
	}
	if data[0] == '[' {
		return b.UnmarshalJSONBrief(data)
	}
	type Alias *BigmapRow
	return json.Unmarshal(data, Alias(b))
}

func (b *BigmapRow) UnmarshalJSONBrief(data []byte) error {
	br := BigmapRow{}
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
			br.RowId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "prev_id":
			br.PrevId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "address":
			br.Address, err = tezos.ParseAddress(f.(string))
		case "account_id":
			br.AccountId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "contract_id":
			br.ContractId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "op_id":
			br.OpId, err = strconv.ParseUint(f.(json.Number).String(), 10, 64)
		case "op":
			br.Op, err = tezos.ParseOpHash(f.(string))
		case "height":
			br.Height, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "time":
			var ts int64
			ts, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
			if err == nil {
				br.Timestamp = time.Unix(0, ts*1000000).UTC()
			}
		case "bigmap_id":
			br.BigMapId, err = strconv.ParseInt(f.(json.Number).String(), 10, 64)
		case "action":
			b.Action = f.(string)
		case "key_hash":
			b.KeyHash, err = tezos.ParseExprHash(f.(string))
		case "key_type":
			b.KeyType = f.(string)
		case "key_encoding":
			b.KeyEncoding = f.(string)
		case "key":
			b.Key = f.(string)
		case "value":
			b.Value = f.(string)
		case "is_replaced":
			br.IsReplaced, err = strconv.ParseBool(f.(json.Number).String())
		case "is_deleted":
			br.IsDeleted, err = strconv.ParseBool(f.(json.Number).String())
		case "is_copied":
			br.IsCopied, err = strconv.ParseBool(f.(json.Number).String())
		}
		if err != nil {
			return err
		}
	}
	*b = br
	return nil
}

type BigmapQuery struct {
	TableQuery
}

func (c *Client) NewBigmapQuery() BigmapQuery {
	tinfo, err := GetTypeInfo(&BigmapRow{}, "")
	if err != nil {
		panic(err)
	}
	q := TableQuery{
		client:  c,
		Params:  c.params.Copy(),
		Table:   "bigmap",
		Format:  FormatJSON,
		Limit:   DefaultLimit,
		Order:   OrderAsc,
		Columns: tinfo.Aliases(),
		Filter:  make(FilterList, 0),
	}
	return BigmapQuery{q}
}

func (q BigmapQuery) Run(ctx context.Context) (*BigmapRowList, error) {
	result := &BigmapRowList{
		columns: q.Columns,
	}
	if err := q.client.QueryTable(ctx, q.TableQuery, result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) QueryBigmap(ctx context.Context, filter FilterList, cols []string) (*BigmapRowList, error) {
	q := c.NewBigmapQuery()
	if len(cols) > 0 {
		q.Columns = cols
	}
	if len(filter) > 0 {
		q.Filter = filter
	}
	return q.Run(ctx)
}

func (c *Client) GetBigmap(ctx context.Context, id int64, params ContractParams) (*Bigmap, error) {
	b := &Bigmap{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/bigmap/%d", id))
	if err := c.get(ctx, u, nil, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) GetBigmapType(ctx context.Context, id int64, params ContractParams) (*BigmapType, error) {
	b := &BigmapType{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/bigmap/%d/type", id))
	if err := c.get(ctx, u, nil, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) GetBigmapKeys(ctx context.Context, id int64, params ContractParams) ([]BigmapKey, error) {
	keys := make([]BigmapKey, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/bigmap/%d/keys", id))
	if err := c.get(ctx, u, nil, &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *Client) GetBigmapValues(ctx context.Context, id int64, params ContractParams) ([]BigmapValue, error) {
	vals := make([]BigmapValue, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/bigmap/%d/values", id))
	if err := c.get(ctx, u, nil, &vals); err != nil {
		return nil, err
	}
	return vals, nil
}

func (c *Client) GetBigmapUpdates(ctx context.Context, id int64, params ContractParams) ([]BigmapUpdate, error) {
	upd := make([]BigmapUpdate, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/bigmap/%d/updates", id))
	if err := c.get(ctx, u, nil, &upd); err != nil {
		return nil, err
	}
	return upd, nil
}

func (c *Client) GetBigmapKey(ctx context.Context, id int64, key string, params ContractParams) (*BigmapKey, error) {
	k := &BigmapKey{}
	u := params.AppendQuery(fmt.Sprintf("/explorer/bigmap/%d/%s", id, key))
	if err := c.get(ctx, u, nil, k); err != nil {
		return nil, err
	}
	return k, nil
}

func (c *Client) GetBigmapKeyUpdates(ctx context.Context, id int64, key string, params ContractParams) ([]BigmapUpdate, error) {
	upd := make([]BigmapUpdate, 0)
	u := params.AppendQuery(fmt.Sprintf("/explorer/bigmap/%d/%s/updates", id, key))
	if err := c.get(ctx, u, nil, &upd); err != nil {
		return nil, err
	}
	return upd, nil
}
