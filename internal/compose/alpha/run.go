// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package alpha

import (
	"fmt"
	"path/filepath"

	"blockwatch.cc/tzgo/internal/compose"
)

// 1 load yaml file
// 2 create accounts
// 3 process pipeline

func (e *Engine) Run(ctx compose.Context, fname string) error {
	spec, err := compose.ParseFile[Spec](fname)
	if err != nil {
		return err
	}
	ctx.WithPath(filepath.Dir(fname))
	for _, a := range spec.Accounts {
		if acc, err := ctx.MakeAccount(int(a.Id), a.Name); err != nil {
			return err
		} else {
			ctx.Log.Infof("Account %s %s", a.Name, acc.Address)
		}
	}
	for n, v := range spec.Variables {
		ctx.AddVariable(n, v)
	}
	if err := spec.Validate(ctx); err != nil {
		return err
	}

	// rewire loggers
	defer ctx.RestoreLogger()
	for _, p := range spec.Pipelines {
		if err := ctx.Cache().Load(p.Hash64(), !ctx.ShouldResume()); err != nil {
			return err
		}
		ctx.Log.Infof("Running pipeline %s", p.Name)
		for i, task := range p.Tasks[ctx.Cache().Get():] {
			ctx.SwitchLogger(fmt.Sprintf("%s[%d/%d]", p.Name, i+1, p.Len()), task.Log)
			if task.Skip {
				if err := ctx.Cache().Update(i); err != nil {
					return err
				}
				continue
			}
			t, err := NewTask(task.Type)
			if err != nil {
				return fmt.Errorf("%s[%d] (%s): %v", p.Name, i, task.Type, err)
			}

			ctx.Log.Debugf("%s %s", task.Type, task.Destination)

			// Build
			op, opts, err := t.Build(ctx, task)
			if err != nil {
				if err == compose.ErrSkip {
					if err := ctx.Cache().Update(i); err != nil {
						return err
					}
					continue
				}
				return fmt.Errorf("%s[%d] (%s): %v", p.Name, i, task.Type, err)
			}

			// send
			opts.Confirmations = 0
			rcpt, err := ctx.Send(op, opts)
			if err != nil {
				switch task.OnError {
				case ErrorModeFail:
					return fmt.Errorf("%s[%d] (%s): %v", p.Name, i, task.Type, err)
				case ErrorModeWarn:
					ctx.Log.Warnf("%s[%d] (%s): %v", p.Name, i, task.Type, err)
				}
			} else {
				// log receipt
				ctx.Log.Infof("%s SUCCESS block=%d hash=%s", t.Type(), rcpt.Height, rcpt.Op.Hash)
			}

			// handle receipt
			if task.Alias != "" {
				addr, _ := rcpt.OriginatedContract()
				ctx.AddVariable(task.Alias, addr.String())
				ctx.Log.Infof("NEW contract %s %s", task.Alias, addr)
			}

			// update pipeline cache
			if err := ctx.Cache().Update(i); err != nil {
				return err
			}
		}
		ctx.RestoreLogger()
	}
	return nil
}

func (e *Engine) Validate(ctx compose.Context, fname string) error {
	spec, err := compose.ParseFile[Spec](fname)
	if err != nil {
		return err
	}
	ctx.WithPath(filepath.Dir(fname))
	for _, a := range spec.Accounts {
		if _, err := ctx.MakeAccount(int(a.Id), a.Name); err != nil {
			return err
		}
	}
	for n, v := range spec.Variables {
		ctx.AddVariable(n, v)
	}
	if err := spec.Validate(ctx); err != nil {
		return err
	}
	for _, p := range spec.Pipelines {
		ctx.Log.Debugf("Validating pipeline %s", p.Name)
		for i, task := range p.Tasks {
			t, err := NewTask(task.Type)
			if err != nil {
				return fmt.Errorf("%s[%d] (%s): %v", p.Name, i, task.Type, err)
			}
			if err := t.Validate(ctx, task); err != nil {
				return fmt.Errorf("%s[%d] (%s): %v", p.Name, i, task.Type, err)
			}
			if task.Type == "deploy" && task.Alias != "" {
				script, err := ParseScript(ctx, task)
				if err != nil {
					return fmt.Errorf("parse script: %v", err)
				}
				acc, _ := ctx.MakeAccount(-2, task.Alias)
				ctx.AddVariable(task.Alias, acc.Address.String())
				ctx.Contracts[acc.Address] = script
			}
		}
	}
	return nil
}
