// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/rpc"
	"blockwatch.cc/tzgo/tezos"
	"github.com/echa/log"
)

type Engine interface {
	Clone(Context, []Op, CloneConfig) ([]byte, error)
	Validate(Context, string) error
	Run(Context, string) error
}

type Account struct {
	Id         int
	Address    tezos.Address
	PrivateKey tezos.PrivateKey
}

type Context struct {
	context.Context
	BaseAccount  Account
	MaxId        int
	Accounts     map[tezos.Address]Account
	Contracts    map[tezos.Address]*micheline.Script
	Variables    map[string]string
	Log          log.Logger
	client       *rpc.Client // RPC client
	url          string      // RPC service URL
	apiKey       string      // RPC service API key
	path         string      // current compose file path
	resume       bool        // continue pipeline execution were we left off
	mode         RunMode     // selected engine run mode
	cache        *PipelineCache
	savedLoggers [2]log.Logger
}

func NewContext(ctx context.Context) Context {
	return Context{
		Context:   ctx,
		Accounts:  make(map[tezos.Address]Account),
		Contracts: make(map[tezos.Address]*micheline.Script),
		MaxId:     -1,
		Variables: make(map[string]string),
		Log:       log.Disabled,
		cache:     NewCache(),
	}
}

func (c *Context) WithLogger(l log.Logger) *Context {
	c.Log = l
	return c
}

func (c *Context) WithUrl(u string) *Context {
	if !strings.HasPrefix(u, "http") {
		u = "http://" + u
	}
	c.url = u
	return c
}

func (c *Context) WithApiKey(k string) *Context {
	c.apiKey = k
	return c
}

func (c *Context) WithResume(b bool) *Context {
	c.resume = b
	return c
}

func (c *Context) WithMode(m RunMode) *Context {
	c.mode = m
	return c
}

func (c Context) ShouldResume() bool {
	return c.resume
}

func (c Context) Filepath() string {
	return c.path
}

func (c *Context) WithPath(p string) *Context {
	c.path = p
	return c
}

func (c *Context) WithBase(k string) *Context {
	if k == "" {
		return c
	}
	sk, err := tezos.ParsePrivateKey(k)
	if err != nil {
		c.Log.Errorf("base key: %v", err)
		return c
	}
	c.BaseAccount.Id = -1
	c.BaseAccount.PrivateKey = sk
	c.BaseAccount.Address = sk.Address()
	c.AddVariable("base", c.BaseAccount.Address.String())
	c.AddAccount(c.BaseAccount)
	return c
}

func (c *Context) Cache() *PipelineCache {
	return c.cache
}

func (c *Context) Init() (err error) {
	if !c.BaseAccount.PrivateKey.IsValid() {
		err = ErrNoBaseKey
		return
	}
	c.Log.Infof("Using base account %s", c.BaseAccount.Address)
	c.AddVariable("zero", tezos.ZeroAddress.String())
	c.AddVariable("burn", tezos.BurnAddress.String())
	c.client, err = rpc.NewClient(c.url, nil)
	if err != nil {
		return
	}
	c.client.ApiKey = c.apiKey
	c.client.CloseConns = true // fix node EOF
	err = c.client.Init(c.Context)
	if err == nil {
		c.Log.Infof("Using chain %s (%s) with %s blocks",
			c.client.ChainId, c.client.Params.Network, c.client.Params.MinimalBlockDelay)
	}
	return
}

func (c *Context) AddVariable(key, val string) {
	if key == "" {
		return
	}
	c.Log.Debugf("Add var %s=%s", key, val)
	c.Variables[CreateVariable(key)] = val
}

func (c *Context) AddAccount(acc Account) {
	c.Log.Debugf("Add account %d=%s key=%s", acc.Id, acc.Address, acc.PrivateKey)
	c.Accounts[acc.Address] = acc
}

func (c *Context) ResolveString(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	v, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid type %T, expected string", val)
	}
	switch {
	case IsTimeExpression(v):
		if conv, err := ConvertTime(v); err != nil {
			return "", err
		} else {
			return conv, nil
		}
	case IsVariable(v):
		if val, ok := c.Variables[v]; ok {
			return val, nil
		} else {
			return "", fmt.Errorf("undefined variable %s", v)
		}
	case IsFile(v):
		fname := filepath.Join(c.path, v[1:])
		val, err := ReadJsonFile[string](fname)
		if err != nil {
			return "", err
		}
		return *val, nil
	default:
		return v, nil
	}
}

func (c *Context) ResolveAddress(val any) (a tezos.Address, err error) {
	if val == nil {
		err = fmt.Errorf("missing value")
		return
	}
	v, err := c.ResolveString(val)
	if err != nil {
		return
	}
	if v == "" {
		err = fmt.Errorf("missing value")
		return
	}
	a, err = tezos.ParseAddress(v)
	return
}

func (c *Context) ResolveInt64(val any) (i int64, err error) {
	if val == nil {
		err = fmt.Errorf("missing value")
		return
	}
	switch v := val.(type) {
	case int:
		i = int64(v)
	case int64:
		i = v
	case uint:
		i = int64(v)
	case uint64:
		i = int64(v)
	case string:
		i, err = strconv.ParseInt(v, 10, 64)
	default:
		err = fmt.Errorf("invalid type %T for int64 value", val)
	}
	return
}

func (c *Context) ResolveZ(val any) (z tezos.Z, err error) {
	if val == nil {
		err = fmt.Errorf("missing value")
		return
	}
	switch v := val.(type) {
	case int:
		z = tezos.NewZ(int64(v))
	case int64:
		z = tezos.NewZ(v)
	case uint:
		z = tezos.NewZ(int64(v))
	case uint64:
		z = tezos.NewZ(int64(v))
	case string:
		v, err = c.ResolveString(v)
		if err != nil {
			return
		}
		z, err = tezos.ParseZ(v)
	default:
		err = fmt.Errorf("invalid type %T for bigint value", val)
	}
	return
}

func (c *Context) ResolvePrivateKey(val any) (sk tezos.PrivateKey, err error) {
	v, ok := val.(string)
	if !ok {
		err = fmt.Errorf("invalid type %T, expected string", val)
		return
	}
	if v == "" {
		return c.BaseAccount.PrivateKey, nil
	}
	var addr tezos.Address
	addr, err = c.ResolveAddress(val)
	if err != nil {
		return
	}
	acc, ok := c.Accounts[addr]
	if !ok {
		err = ErrNoAccount
		return
	}
	sk = acc.PrivateKey
	return
}

func (c *Context) ResolveScript(addr tezos.Address) (*micheline.Script, error) {
	if s, ok := c.Contracts[addr]; ok {
		return s, nil
	}
	script, err := c.client.GetNormalizedScript(c.Context, addr, "")
	if err != nil {
		return nil, err
	}
	script.Code.Code = micheline.InvalidPrim
	script.Code.View = micheline.InvalidPrim
	c.Contracts[addr] = script
	return script, nil
}

func (c *Context) Send(op *codec.Op, opts *rpc.CallOptions) (*rpc.Receipt, error) {
	if c.mode == RunModeSimulate {
		key, err := opts.Signer.GetKey(c.Context, op.Source)
		if err != nil {
			return nil, err
		}
		if err := c.client.Complete(c.Context, op, key); err != nil {
			return nil, err
		}
		return c.client.Simulate(c.Context, op, opts)
	}
	rcpt, err := c.client.Send(c.Context, op, opts)
	if err != nil {
		return nil, err
	}
	if !rcpt.IsSuccess() {
		return rcpt, rcpt.Error()
	}
	return rcpt, nil
}

func (c *Context) SubscribeBlocks(cb rpc.ObserverCallback) (int, error) {
	c.client.Listen()
	id := c.client.BlockObserver.Subscribe(tezos.ZeroOpHash, cb)
	return id, nil
}

func (c *Context) UnsubscribeBlocks(id int) error {
	c.client.BlockObserver.Unsubscribe(id)
	return nil
}

func (c *Context) Params() *tezos.Params {
	return c.client.Params
}

func (c *Context) HeadBlock() *rpc.BlockHeaderLogEntry {
	return c.client.BlockObserver.Head()
}

func (c *Context) SwitchLogger(tag, lvl string) {
	if c.savedLoggers[0] == nil {
		c.savedLoggers[0] = c.Log
		c.savedLoggers[1] = c.client.Log
	}
	c.Log = c.savedLoggers[0].Clone().WithTag(tag).SetLevelString(lvl)
	c.client.Log = c.savedLoggers[1].Clone().WithTag(tag).SetLevelString(lvl)
}

func (c *Context) RestoreLogger() {
	if c.savedLoggers[0] != nil {
		c.Log = c.savedLoggers[0]
		c.client.Log = c.savedLoggers[1]
	}
	c.savedLoggers[0] = nil
	c.savedLoggers[1] = nil
}

func (c *Context) WaitNumBlocks(n int) error {
	if n == 0 {
		return nil
	}
	done := make(chan struct{})
	_, err := c.SubscribeBlocks(func(_ *rpc.BlockHeaderLogEntry, _ int64, _ int, _ int, _ bool) bool {
		n--
		if n <= 0 {
			close(done)
			return true
		}
		return false
	})
	if err != nil {
		return err
	}
	select {
	case <-done:
		return nil
	case <-c.Done():
		return c.Err()
	}
}

func (c Context) Fetch(u string, v any) error {
	if !strings.HasPrefix(u, "http") {
		u = c.url + u
	}
	raw, err := Fetch[json.RawMessage](c, u)
	if err != nil {
		return err
	}
	return json.Unmarshal(*raw, v)
}
