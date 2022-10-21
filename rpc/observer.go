// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"sync"
	"time"

	"blockwatch.cc/tzgo/hash"
	"blockwatch.cc/tzgo/tezos"
)

// WIP: interface may change
//
// TODO:
// - support op hashes in the past (i.e. in head block)
// - support multiple subscriptions (funcs) for the same op hash
// - support block subscriptions (to connect a BlockObserver for full blocks + reorgs)
// - support AdressObserver with address subscription filter
// - disable events/polling when no subscriber exists
// - cache head block op hashes to avoid race conditions with late subsribers
// - handle reorgs (inclusion may switch to a different block hash)

type ObserverCallback func(tezos.BlockHash, int, int, bool) bool

type observerSubscription struct {
	id      int
	cb      ObserverCallback
	oh      tezos.OpHash
	matched bool
}

type Observer struct {
	registry   map[int]*observerSubscription
	hashmap    map[uint64]int
	seq        int
	once       sync.Once
	mu         sync.Mutex
	done       chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	c          *Client
	minDelay   time.Duration
	bestHash   tezos.BlockHash
	bestHeight int64
}

func NewObserver() *Observer {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Observer{
		registry: make(map[int]*observerSubscription),
		hashmap:  make(map[uint64]int),
		minDelay: tezos.DefaultParams.MinimalBlockDelay,
		ctx:      ctx,
		cancel:   cancel,
	}
	return m
}

func (m *Observer) WithDelay(minDelay time.Duration) *Observer {
	m.minDelay = minDelay
	return m
}

func (m *Observer) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancel()
	close(m.done)
	m.registry = make(map[int]*observerSubscription)
	m.hashmap = make(map[uint64]int)
}

func (m *Observer) Subscribe(oh tezos.OpHash, cb ObserverCallback) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	seq := m.seq
	m.registry[seq] = &observerSubscription{
		id: seq,
		cb: cb,
		oh: oh,
	}
	hashval := hash.NewInlineFNV64a()
	hashval.Write(oh.Hash.Hash)
	m.hashmap[hashval.Sum64()] = seq
	log.Debugf("monitor: %03d subscribed %s", seq, oh)
	return seq
}

func (m *Observer) Unsubscribe(id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	req, ok := m.registry[id]
	if ok {
		hashval := hash.NewInlineFNV64a()
		hashval.Write(req.oh.Hash.Hash)
		delete(m.hashmap, hashval.Sum64())
		delete(m.registry, id)
		log.Debugf("monitor: %03d unsubscribed %s", id, req.oh)
	}
}

func (m *Observer) Listen(cli *Client) {
	m.once.Do(func() {
		m.c = cli
		if m.c.Params != nil {
			m.minDelay = m.c.Params.MinimalBlockDelay
		}
		go m.listenBlocks()
	})
}

func (m *Observer) ListenMempool(cli *Client) {
	m.once.Do(func() {
		m.c = cli
		if m.c.Params != nil {
			m.minDelay = m.c.Params.MinimalBlockDelay
		}
		go m.listenMempool()
	})
}

func (m *Observer) listenMempool() {
	// TODO
}

func (m *Observer) listenBlocks() {
	var mon *BlockHeaderMonitor
	useEvents := true
	defer func() {
		if mon != nil {
			mon.Close()
		}
	}()

	for {
		// handle close request
		select {
		case <-m.ctx.Done():
			return
		default:
		}

		// (re)connect
		if mon == nil && useEvents {
			mon = NewBlockHeaderMonitor()
			if err := m.c.MonitorBlockHeader(m.ctx, mon); err != nil {
				mon.Close()
				mon = nil
				if e, ok := err.(HTTPStatus); ok && e.StatusCode() == 404 {
					log.Debug("monitor: event mode unsupported, falling back to poll mode.")
					useEvents = false
				} else {
					// wait 5 sec, but also return on close
					select {
					case <-m.ctx.Done():
						return
					case <-time.After(5 * time.Second):
					}
				}
				continue
			}
		}

		var (
			headBlock  tezos.BlockHash
			headHeight int64
		)
		if mon != nil && useEvents {
			// event mode: wait for next block message
			head, err := mon.Recv(m.ctx)
			// reconnect on error unless context was cancelled
			if err != nil {
				mon.Close()
				mon = nil
				continue
			}
			log.Debugf("monitor: new head %s", head.Hash)
			headBlock = head.Hash
			headHeight = head.Level
		} else {
			// poll mode: check every 30sec
			head, err := m.c.GetTipHeader(m.ctx)
			if err != nil {
				// wait 5 sec, but also return on close
				select {
				case <-m.ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
				continue
			}
			headBlock = head.Hash.Clone()
			headHeight = head.Level
		}

		// skip already processed blocks
		if headBlock.Equal(m.bestHash) && !useEvents {
			// wait minDelay/2 for late blocks
			if !useEvents {
				select {
				case <-m.ctx.Done():
					return
				case <-time.After(m.minDelay / 2):
				}
			}
			continue
		}
		log.Debugf("monitor: new block %d %s", headHeight, headBlock)

		// TODO: check for reorg and gaps

		// callback for all previous matches
		m.mu.Lock()
		for _, v := range m.registry {
			if v.matched {
				log.Debugf("monitor: signal n-th match for %d %s", v.id, v.oh)
				if remove := v.cb(headBlock, -1, -1, false); remove {
					delete(m.registry, v.id)
				}
			}
		}
		m.mu.Unlock()

		// pull block ops and fan-out matches
		ohs, err := m.c.GetBlockOperationHashes(m.ctx, headBlock)
		if err != nil {
			log.Warnf("monitor: cannot fetch block ops: %v", err)
			continue
		}
		hashval := hash.NewInlineFNV64a()
		m.mu.Lock()
		for l, list := range ohs {
			for n, h := range list {
				// match op hash against registry
				hashval.Write(h.Hash.Hash)
				id, ok := m.hashmap[hashval.Sum64()]
				hashval.Reset()
				if !ok {
					log.Debugf("monitor: --- !! %s", h)
					continue
				}
				match, ok := m.registry[id]
				if !ok {
					log.Debugf("monitor: --- !! %s", h)
					continue
				}

				// cross check hash to guard against hash collisions
				if !match.oh.Equal(h) {
					log.Debugf("monitor: %03d != %s", id, h)
					continue
				}

				log.Debugf("monitor: matched %d %s", match.id, match.oh)

				// callback
				if remove := match.cb(headBlock, l, n, false); remove {
					delete(m.registry, match.id)
				} else {
					match.matched = true
				}
			}
		}

		// update monitor state
		m.bestHash = headBlock
		m.bestHeight = headHeight
		m.mu.Unlock()

		// wait in poll mode
		if !useEvents {
			select {
			case <-m.ctx.Done():
				return
			case <-time.After(m.minDelay):
			}
		}
	}
}
