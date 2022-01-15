// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package wallet

import (
    "context"
    "sync"
    "time"

    "fmt"

    "blockwatch.cc/tzgo/hash"
    "blockwatch.cc/tzgo/rpc"
    "blockwatch.cc/tzgo/tezos"
)

// WIP: interface may change
//
// TODO:
// - support op hashes in the past (i.e. in head block)
// - support address subscription filters
// - disable events/polling when no subscriber exists
// - cache head block op hashes to avoid race conditions with late subsribers
// - allow multiple subscription funcs for the same op hash
// - handle reorgs (inclusion may switch to a different block hash)

type MonitorCallback func(tezos.BlockHash, int, int, bool) bool

type monitorSubscription struct {
    id      int
    cb      MonitorCallback
    oh      tezos.OpHash
    matched bool
}

type Monitor struct {
    registry   map[int]*monitorSubscription
    hashmap    map[uint64]int
    seq        int
    once       sync.Once
    mu         sync.Mutex
    done       chan struct{}
    ctx        context.Context
    cancel     context.CancelFunc
    c          *rpc.Client
    minDelay   time.Duration
    bestHash   tezos.BlockHash
    bestHeight int64
}

func NewMonitor() *Monitor {
    ctx, cancel := context.WithCancel(context.Background())
    m := &Monitor{
        registry: make(map[int]*monitorSubscription),
        hashmap:  make(map[uint64]int),
        minDelay: 30 * time.Second,
        ctx:      ctx,
        cancel:   cancel,
    }
    return m
}

func (m *Monitor) WithDelay(minDelay time.Duration) *Monitor {
    m.minDelay = minDelay
    return m
}

func (m *Monitor) Close() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.cancel()
    close(m.done)
}

func (m *Monitor) Subscribe(oh tezos.OpHash, cb MonitorCallback) int {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.seq++
    seq := m.seq
    m.registry[seq] = &monitorSubscription{
        id: seq,
        cb: cb,
        oh: oh,
    }
    hashval := hash.NewInlineFNV64a()
    hashval.Write(oh.Hash.Hash)
    m.hashmap[hashval.Sum64()] = seq
    return seq
}

func (m *Monitor) Unsubscribe(id int) {
    m.mu.Lock()
    defer m.mu.Unlock()
    req, ok := m.registry[id]
    if ok {
        hashval := hash.NewInlineFNV64a()
        hashval.Write(req.oh.Hash.Hash)
        delete(m.hashmap, hashval.Sum64())
        delete(m.registry, id)
    }
}

func (m *Monitor) Listen(c *rpc.Client) {
    m.once.Do(func() {
        m.c = c
        go m.listenBlocks()
    })
}

func (m *Monitor) listenBlocks() {
    var mon *rpc.BlockHeaderMonitor
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
            mon = rpc.NewBlockHeaderMonitor()
            if err := m.c.MonitorBlockHeader(m.ctx, mon); err != nil {
                mon.Close()
                mon = nil
                if e, ok := err.(rpc.HTTPStatus); ok && e.StatusCode() == 404 {
                    fmt.Println("monitor: event mode unsupported, falling back to poll mode.")
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
            fmt.Println("monitor: new head", head.Hash)
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
        fmt.Printf("Monitor: new block %d %s\n", headHeight, headBlock)

        // TODO: check for reorg and gaps

        // callback for all previous matches
        m.mu.Lock()
        for _, v := range m.registry {
            if v.matched {
                fmt.Printf("Monitor: signal n-th match for %d %s\n", v.id, v.oh)
                if remove := v.cb(headBlock, -1, -1, false); remove {
                    delete(m.registry, v.id)
                }
            }
        }
        m.mu.Unlock()

        // pull block ops and fan-out matches
        ohs, err := m.c.GetBlockOperationHashes(m.ctx, headBlock)
        if err != nil {
            continue
        }
        hashval := hash.NewInlineFNV64a()
        m.mu.Lock()
        for l, list := range ohs {
            for n, h := range list {
                // match op hash against registry
                hashval.Write(h.Hash.Hash)
                id, ok := m.hashmap[hashval.Sum64()]
                match, ok := m.registry[id]
                hashval.Reset()
                if !ok {
                    continue
                }

                // cross check hash to guard against hash collisions
                if !match.oh.Equal(h) {
                    continue
                }

                fmt.Printf("Monitor: matched %d %s\n", match.id, match.oh)

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
