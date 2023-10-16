// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package alpha

import (
	"bytes"

	"blockwatch.cc/tzgo/internal/compose"
	"gopkg.in/yaml.v3"
)

func (e *Engine) Clone(ctx compose.Context, ops []compose.Op, cfg compose.CloneConfig) ([]byte, error) {
	spec := Spec{
		Version: VERSION,
		Pipelines: []Pipeline{{
			Name:  cfg.Name,
			Tasks: make([]Task, 0, len(ops)),
		}},
	}

	for _, op := range ops {
		switch op.Type {
		case "origination":
			s := &Script{
				Code: &Code{},
				Storage: &Storage{
					Args: op.Args,
				},
			}
			switch cfg.Mode {
			case compose.CloneModeFile:
				s.Code.File = op.Url + "#code"
				if s.Storage.Args == nil {
					s.Storage.File = op.Url + "#storage"
				}
			case compose.CloneModeJson, compose.CloneModeBinary:
				s.Code.Value = op.PackedCode
				if s.Storage.Args == nil {
					s.Storage.Value = op.PackedStorage
				}
			case compose.CloneModeUrl:
				s.Code.Url = op.Url + "#code"
				if s.Storage.Args == nil {
					s.Storage.Url = op.Url + "#storage"
				}
			}
			spec.Pipelines[0].Tasks = append(spec.Pipelines[0].Tasks, Task{
				Type:   "deploy",
				Alias:  cfg.Name,
				Source: op.Sender,
				Amount: uint64(op.Amount * 1_000_000),
				Script: s,
			})
		case "transaction":
			typ := "transfer"
			var p *Params
			if op.Params != nil {
				typ = "call"
				p = &Params{
					Entrypoint: op.Params.Entrypoint,
					Args:       op.Args,
				}
				if p.Args == nil {
					switch cfg.Mode {
					case compose.CloneModeFile:
						p.File = op.Url
					case compose.CloneModeJson, compose.CloneModeBinary:
						p.Value = op.PackedParams
					case compose.CloneModeUrl:
						p.Url = op.Url
					}
				}
			}
			spec.Pipelines[0].Tasks = append(spec.Pipelines[0].Tasks, Task{
				Type:        typ,
				Source:      op.Sender,
				Destination: "$" + cfg.Name,
				Amount:      uint64(op.Amount * 1_000_000),
				Params:      p,
			})
		}
	}

	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	err := enc.Encode(spec)
	return buf.Bytes(), err
}
