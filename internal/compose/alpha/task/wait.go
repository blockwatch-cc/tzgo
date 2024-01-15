// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package task

import (
	"fmt"
	"strconv"
	"time"

	"blockwatch.cc/tzgo/codec"
	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/internal/compose/alpha"
	"blockwatch.cc/tzgo/rpc"

	"github.com/pkg/errors"
)

var _ alpha.TaskBuilder = (*WaitTask)(nil)

func init() {
	alpha.RegisterTask("wait", NewWaitTask)
}

type WaitTask struct {
	Mode     alpha.WaitMode // cycle, block, time
	Relative bool           // relative/absolute
	Value    int64          // parsed config value
	Start    int64          // wait start (for relative)
}

func NewWaitTask() alpha.TaskBuilder {
	return &WaitTask{}
}

func (t *WaitTask) Type() string {
	return "wait"
}

func (t *WaitTask) Build(ctx compose.Context, task alpha.Task) (*codec.Op, *rpc.CallOptions, error) {
	if err := t.parse(ctx, task); err != nil {
		return nil, nil, errors.Wrap(err, "parse")
	}

	done := make(chan struct{})
	id, err := ctx.SubscribeBlocks(func(h *rpc.BlockHeaderLogEntry, height int64, _ int, _ int, _ bool) bool {
		isDone := false
		var val int64
		p := ctx.Params()
		switch t.Mode {
		case alpha.WaitModeCycle:
			val = p.CycleFromHeight(height)
			if t.Start == 0 {
				t.Start = val
				if t.Relative {
					diff := time.Duration((p.CycleStartHeight(val+t.Value) - height)) * p.MinimalBlockDelay
					ctx.Log.Infof("waiting for %d cycles, approx %s", t.Value, diff)
				} else {
					diff := time.Duration((p.CycleStartHeight(t.Value) - height)) * p.MinimalBlockDelay
					ctx.Log.Infof("waiting until cycle %d, approx %s", t.Value, diff)
				}
			}
			ctx.Log.Debugf("block %d cycle=%d", height, val)
		case alpha.WaitModeBlock:
			val = height
			if t.Start == 0 {
				t.Start = val
				if t.Relative {
					diff := time.Duration(val+t.Value-height) * p.MinimalBlockDelay
					ctx.Log.Infof("waiting for %d blocks, approx %s", t.Value, diff)
				} else {
					diff := time.Duration(t.Value-height) * p.MinimalBlockDelay
					ctx.Log.Infof("waiting until block %d, approx %s", t.Value, diff)
				}
			}
			ctx.Log.Debugf("block %d", height)
		case alpha.WaitModeTime:
			val = time.Now().UTC().Unix()
			if t.Start == 0 {
				t.Start = val
				if t.Relative {
					ctx.Log.Infof("waiting for %s", time.Duration(t.Value)*time.Second)
				} else {
					diff := time.Unix(t.Value, 0).Sub(time.Unix(val, 0))
					ctx.Log.Infof("waiting until %s, approx %s", time.Unix(t.Value, 0), diff)
				}
			}
			ctx.Log.Debugf("block %d time=%d", height, val)
		}
		if t.Relative {
			isDone = val >= t.Start+t.Value
		} else {
			isDone = val >= t.Value
		}
		if isDone {
			ctx.Log.Debug("wait done")
			close(done)
			return true
		}
		return false
	})
	if err != nil {
		return nil, nil, err
	}
	defer ctx.UnsubscribeBlocks(id)

	// wait
	select {
	case <-done:
		return nil, nil, compose.ErrSkip
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
}

func (t *WaitTask) Validate(ctx compose.Context, task alpha.Task) error {
	return t.parse(ctx, task)
}

func (t *WaitTask) parse(ctx compose.Context, task alpha.Task) error {
	if task.WaitMode == alpha.WaitModeInvalid {
		return fmt.Errorf("missing wait mode")
	}
	t.Mode = task.WaitMode
	val, err := ctx.ResolveString(task.Value)
	if err != nil {
		return errors.Wrap(err, "value")
	}
	if val == "" {
		return fmt.Errorf("missing value")
	}
	t.Relative = val[0] == '+'
	if t.Relative {
		val = val[1:]
	}
	switch t.Mode {
	case alpha.WaitModeCycle, alpha.WaitModeBlock:
		var u uint64
		u, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		t.Value = int64(u)
	case alpha.WaitModeTime:
		if t.Relative {
			var d time.Duration
			d, err = time.ParseDuration(val)
			if err != nil {
				return err
			}
			t.Value = int64(d / time.Second)
		} else {
			// accept time expressions ($now+dur) and timestamps as text or unix seconds
			val, err = compose.ConvertTime(val)
			if err != nil {
				return err
			}
			var tm time.Time
			tm, err = time.Parse(time.RFC3339, val)
			if err != nil {
				return err
			}
			t.Value = tm.Unix()
		}
	}
	return nil
}
