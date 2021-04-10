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
	Contract        tezos.Address     `json:"contract"`
	BigMapId        int64             `json:"bigmap_id"`
	NUpdates        int64             `json:"n_updates"`
	NKeys           int64             `json:"n_keys"`
	AllocatedHeight int64             `json:"alloc_height"`
	AllocatedBlock  tezos.BlockHash   `json:"alloc_block"`
	AllocatedTime   time.Time         `json:"alloc_time"`
	UpdatedHeight   int64             `json:"update_height"`
	UpdatedBlock    tezos.BlockHash   `json:"update_block"`
	UpdatedTime     time.Time         `json:"update_time"`
	IsRemoved       bool              `json:"is_removed"`
	KeyType         micheline.Typedef `json:"key_type"`
	ValueType       micheline.Typedef `json:"value_type"`
}

type BigmapType struct {
	Contract      tezos.Address     `json:"contract"`
	BigMapId      int64             `json:"bigmap_id"`
	KeyEncoding   string            `json:"key_encoding"`
	KeyType       micheline.Typedef `json:"key_type"`
	ValueType     micheline.Typedef `json:"value_type"`
	KeyTypePrim   micheline.Prim    `json:"key_type_prim"`
	ValueTypePrim micheline.Prim    `json:"value_type_prim"`
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
	Key     MultiKey       `json:"key"`
	KeyHash tezos.ExprHash `json:"key_hash"`
	Meta    BigmapMeta     `json:"meta"`
	Prim    micheline.Prim `json:"prim"`
}

type MultiKey struct {
	named  map[string]interface{}
	anon   []interface{}
	single string
}

func (k MultiKey) Len() int {
	if len(k.single) > 0 {
		return 1
	}
	return len(k.named) + len(k.anon)
}

func (k MultiKey) String() string {
	switch true {
	case len(k.named) > 0:
		strs := make([]string, 0)
		for n, v := range k.named {
			strs = append(strs, fmt.Sprintf("%s=%s", n, v))
		}
		return strings.Join(strs, ",")
	case len(k.anon) > 0:
		strs := make([]string, 0)
		for _, v := range k.anon {
			strs = append(strs, ToString(v))
		}
		return strings.Join(strs, ",")
	default:
		return k.single
	}
}

func (k MultiKey) MarshalJSON() ([]byte, error) {
	switch true {
	case len(k.named) > 0:
		return json.Marshal(k.named)
	case len(k.anon) > 0:
		return json.Marshal(k.anon)
	default:
		return []byte(strconv.Quote(k.single)), nil
	}
}

func (k *MultiKey) UnmarshalJSON(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	switch buf[0] {
	case '{':
		m := make(map[string]interface{})
		if err := json.Unmarshal(buf, &m); err != nil {
			return err
		}
		k.named = m
	case '[':
		m := make([]interface{}, 0)
		if err := json.Unmarshal(buf, &m); err != nil {
			return err
		}
		k.anon = m
	case '"':
		s, _ := strconv.Unquote(string(buf))
		k.single = s
	default:
		k.single = string(buf)
	}
	return nil
}

func (k MultiKey) GetString(path string) (string, bool) {
	return getPathString(nonNil(k.named, k.anon, k.single), path)
}

func (k MultiKey) GetInt64(path string) (int64, bool) {
	return getPathInt64(nonNil(k.named, k.anon, k.single), path)
}

func (k MultiKey) GetBig(path string) (*big.Int, bool) {
	return getPathBig(nonNil(k.named, k.anon, k.single), path)
}

func (k MultiKey) GetTime(path string) (time.Time, bool) {
	return getPathTime(nonNil(k.named, k.anon, k.single), path)
}

func (k MultiKey) GetAddress(path string) (tezos.Address, bool) {
	return getPathAddress(nonNil(k.named, k.anon, k.single), path)
}

func (k MultiKey) GetValue(path string) (interface{}, bool) {
	return getPathValue(nonNil(k.named, k.anon, k.single), path)
}

func (k MultiKey) Walk(path string, fn ValueWalkerFunc) error {
	val := nonNil(k.named, k.anon, k.single)
	if len(path) > 0 {
		var ok bool
		val, ok = getPathValue(val, path)
		if !ok {
			return nil
		}
	}
	return walkValueMap(path, val, fn)
}

type BigmapValue struct {
	Key       MultiKey       `json:"key"`
	KeyHash   tezos.ExprHash `json:"key_hash"`
	Meta      BigmapMeta     `json:"meta"`
	Value     interface{}    `json:"value"`
	KeyPrim   micheline.Prim `json:"key_prim"`
	ValuePrim micheline.Prim `json:"value_prim"`
}

func (v BigmapValue) GetString(path string) (string, bool) {
	return getPathString(v.Value, path)
}

func (v BigmapValue) GetInt64(path string) (int64, bool) {
	return getPathInt64(v.Value, path)
}

func (v BigmapValue) GetBig(path string) (*big.Int, bool) {
	return getPathBig(v.Value, path)
}

func (v BigmapValue) GetTime(path string) (time.Time, bool) {
	return getPathTime(v.Value, path)
}

func (v BigmapValue) GetAddress(path string) (tezos.Address, bool) {
	return getPathAddress(v.Value, path)
}

func (v BigmapValue) GetValue(path string) (interface{}, bool) {
	return getPathValue(v.Value, path)
}

func (v BigmapValue) Walk(path string, fn ValueWalkerFunc) error {
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

type BigmapUpdate struct {
	BigmapValue
	Action        micheline.DiffAction `json:"action"`
	KeyType       micheline.Typedef    `json:"key_type"`
	ValueType     micheline.Typedef    `json:"value_type"`
	KeyTypePrim   micheline.Prim       `json:"key_type_prim"`
	ValueTypePrim micheline.Prim       `json:"value_type_prim"`
	SourceId      int64                `json:"source_big_map"`
	DestId        int64                `json:"destination_big_map"`
}

type BigmapRow struct {
	RowId       uint64               `json:"row_id"`
	PrevId      uint64               `json:"prev_id"`
	Address     tezos.Address        `json:"address"`
	AccountId   uint64               `json:"account_id"`
	ContractId  uint64               `json:"contract_id"`
	OpId        uint64               `json:"op_id"`
	Op          tezos.OpHash         `json:"op"`
	Height      int64                `json:"height"`
	Timestamp   time.Time            `json:"time"`
	BigMapId    int64                `json:"bigmap_id"`
	Action      micheline.DiffAction `json:"action"`
	KeyHash     tezos.ExprHash       `json:"key_hash,omitempty"`
	KeyType     string               `json:"key_type,omitempty"`
	KeyEncoding string               `json:"key_encoding,omitempty"`
	Key         string               `json:"key,omitempty"`
	Value       string               `json:"value,omitempty"`
	IsReplaced  bool                 `json:"is_replaced"`
	IsDeleted   bool                 `json:"is_deleted"`
	IsCopied    bool                 `json:"is_copied"`

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
			b.Action, err = micheline.ParseDiffAction(f.(string))
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
