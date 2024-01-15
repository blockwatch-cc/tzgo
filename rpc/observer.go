// Copyright (c) 2020-2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package rpc

import (
	"context"
	"sync"
	"time"

	"blockwatch.cc/tzgo/tezos"
)

// WIP: interface may change
//
// TODO:
// - support AdressObserver with address subscription filter
// - disable events/polling when no subscriber exists
// - handle reorgs (inclusion may switch to a different block hash)

type ObserverCallback func(*BlockHeaderLogEntry, int64, int, int, bool) bool

type observerSubscription struct {
	id      int
	cb      ObserverCallback
	oh      tezos.OpHash
	matched bool
}

type Observer struct {
	subs     map[int]*observerSubscription
	watched  map[tezos.OpHash][]int
	recent   map[tezos.OpHash][3]int64
	seq      int
	once     sync.Once
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	c        *Client
	minDelay time.Duration
	head     *BlockHeaderLogEntry
}

func NewObserver() *Observer {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Observer{
		subs:     make(map[int]*observerSubscription),
		watched:  make(map[tezos.OpHash][]int),
		recent:   make(map[tezos.OpHash][3]int64),
		minDelay: tezos.DefaultParams.MinimalBlockDelay,
		ctx:      ctx,
		cancel:   cancel,
		head: &BlockHeaderLogEntry{
			Level: -1,
		},
	}
	return m
}

func (m *Observer) Head() *BlockHeaderLogEntry {
	return m.head
}

func (m *Observer) WithDelay(minDelay time.Duration) *Observer {
	m.minDelay = minDelay
	return m
}

func (m *Observer) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancel()
	m.subs = make(map[int]*observerSubscription)
	m.watched = make(map[tezos.OpHash][]int)
	m.recent = make(map[tezos.OpHash][3]int64)
}

func (m *Observer) Subscribe(oh tezos.OpHash, cb ObserverCallback) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	seq := m.seq
	m.subs[seq] = &observerSubscription{
		id: seq,
		cb: cb,
		oh: oh,
	}
	if pos, ok := m.recent[oh]; ok {
		match := m.subs[seq]
		m.c.Log.Debugf("monitor: %03d direct match %s", seq, oh)
		if remove := match.cb(m.head, pos[0], int(pos[1]), int(pos[2]), false); remove {
			delete(m.subs, match.id)
		}
	}
	m.c.Log.Debugf("monitor: %03d subscribed %s", seq, oh)
	m.watched[oh] = append(m.watched[oh], seq)
	return seq
}

func (m *Observer) Unsubscribe(id int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	req, ok := m.subs[id]
	if ok {
		m.removeWatcher(req.oh, id)
		delete(m.subs, id)
		m.c.Log.Debugf("monitor: %03d unsubscribed %s", id, req.oh)
	}
}

func (m *Observer) removeWatcher(oh tezos.OpHash, id int) {
	n := 0
	for _, v := range m.watched[oh] {
		if v != id {
			m.watched[oh][n] = v
			n++
		}
	}
	if n == 0 {
		delete(m.watched, oh)
	} else {
		m.watched[oh] = m.watched[oh][:n]
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
	var (
		mon       *BlockHeaderMonitor
		useEvents bool = true
		firstLoop bool = true
	)
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
				if ErrorStatus(err) == 404 {
					m.c.Log.Debug("monitor: event mode unsupported, falling back to poll mode.")
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

		var head *BlockHeaderLogEntry
		if mon != nil && useEvents && !firstLoop {
			// event mode: wait for next block message
			var err error
			head, err = mon.Recv(m.ctx)
			// reconnect on error unless context was cancelled
			if err != nil {
				mon.Close()
				mon = nil
				continue
			}
			// m.c.Log.Debugf("monitor: new head %s", head.Hash)
		} else {
			// poll mode: check every n sec
			h, err := m.c.GetTipHeader(m.ctx)
			if err != nil {
				// wait 5 sec, but also return on close
				select {
				case <-m.ctx.Done():
					return
				case <-time.After(5 * time.Second):
				}
				continue
			}
			head = h.LogEntry()
		}
		firstLoop = false

		// skip already processed blocks
		if head.Hash.Equal(m.head.Hash) && !useEvents {
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
		m.c.Log.Debugf("monitor: new block %d %s", head.Level, head.Hash)

		// TODO: check for reorg and gaps

		// handle block watchers
		m.mu.Lock()
		for _, id := range m.watched[tezos.ZeroOpHash] {
			sub, ok := m.subs[id]
			if !ok {
				m.removeWatcher(tezos.ZeroOpHash, id)
				continue
			}
			if remove := sub.cb(head, head.Level, -1, -1, false); remove {
				delete(m.subs, id)
				m.removeWatcher(tezos.ZeroOpHash, id)
			}
		}

		// callback for all previous matches who have not yet unregistered (i.e. waiting for
		// additional confirmations)
		for _, v := range m.subs {
			if v.matched {
				m.c.Log.Debugf("monitor: signal n-th match for %d %s", v.id, v.oh)
				if remove := v.cb(head, head.Level, -1, -1, false); remove {
					delete(m.subs, v.id)
					m.removeWatcher(v.oh, v.id)
				}
			}
		}
		// clear recent op hashes
		for n := range m.recent {
			delete(m.recent, n)
		}

		numSubs := len(m.subs)
		m.mu.Unlock()

		// pull block ops when subs exist
		var (
			ohs [][]tezos.OpHash
			err error
		)
		if numSubs > 0 {
			ohs, err = m.c.GetBlockOperationHashes(m.ctx, head.Hash)
			if err != nil {
				m.c.Log.Warnf("monitor: cannot fetch block ops: %v", err)
				continue
			}
		}

		// fan-out matches
		m.mu.Lock()
		for l, list := range ohs {
			for n, h := range list {
				// keep as recent
				m.recent[h] = [3]int64{head.Level, int64(l), int64(n)}

				// match op hash against subs
				ids, ok := m.watched[h]
				if !ok {
					m.c.Log.Debugf("monitor: --- !! %s", h)
					continue
				}

				// handle all subscriptions for this op hash
				var removed []*observerSubscription
				for _, id := range ids {
					sub, ok := m.subs[id]
					if !ok {
						m.c.Log.Debugf("monitor: --- !! %s", h)
						continue
					}

					m.c.Log.Debugf("monitor: matched %d %s", sub.id, sub.oh)

					// callback
					if remove := sub.cb(head, head.Level, l, n, false); remove {
						delete(m.subs, sub.id)
						removed = append(removed, sub)
					} else {
						sub.matched = true
					}
				}

				// remove deleted subs from watch list
				for _, sub := range removed {
					m.removeWatcher(sub.oh, sub.id)
				}
			}
		}

		// update monitor state
		m.head = head
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
