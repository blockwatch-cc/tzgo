// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package alpha

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"blockwatch.cc/tzgo/internal/compose"
	"blockwatch.cc/tzgo/micheline"
	"blockwatch.cc/tzgo/tezos"
	"gopkg.in/yaml.v3"
)

type Spec struct {
	Version   string            `yaml:"version"`
	Accounts  []Account         `yaml:"accounts,omitempty"`
	Variables map[string]string `yaml:"variables,omitempty"`
	Pipelines PipelineList      `yaml:"pipelines"`
}

func (s Spec) Validate(ctx compose.Context) error {
	if s.Version == "" {
		return compose.ErrNoVersion
	}
	if len(s.Pipelines) == 0 {
		return compose.ErrNoPipeline
	}
	for _, v := range s.Pipelines {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("pipeline %s: %v", v.Name, err)
		}
	}
	return nil
}

type Account struct {
	Name string `yaml:"name"`
	Id   uint   `yaml:"id,omitempty"`
}

type Pipeline struct {
	Name  string `yaml:"-"`
	Tasks []Task `yaml:",inline"`
}

type PipelineList []Pipeline

func (l *PipelineList) UnmarshalYAML(node *yaml.Node) error {
	// decode named map into list
	*l = make([]Pipeline, len(node.Content)/2)
	for i, v := range node.Content {
		switch v.Kind {
		case yaml.ScalarNode:
			(*l)[i/2].Name = v.Value
		case yaml.SequenceNode:
			if err := v.Decode(&(*l)[i/2].Tasks); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected yaml node kind=%s value=%s", v.Tag[2:], v.Value)
		}
	}

	return nil
}

func (l PipelineList) MarshalYAML() (any, error) {
	// manualy create a list of named map nodes
	node := &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: []*yaml.Node{},
	}
	for _, v := range l {
		var child yaml.Node
		if err := child.Encode(v.Tasks); err != nil {
			return nil, err
		}
		node.Content = append(node.Content,
			// key
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: v.Name,
			},
			// value
			&child,
		)
	}
	return node, nil
}

func (p Pipeline) Hash64() uint64 {
	h := fnv.New64()
	enc := json.NewEncoder(h)
	enc.Encode(p)
	return h.Sum64()
}

func (p Pipeline) Len() int {
	return len(p.Tasks)
}

func (p Pipeline) Validate(ctx compose.Context) error {
	if p.Name == "" {
		return compose.ErrNoPipelineName
	}
	for i, v := range p.Tasks {
		if err := v.Validate(ctx); err != nil {
			return fmt.Errorf("task %d (%s): %v", i, v.Type, err)
		}
	}
	return nil
}

type Task struct {
	Type        string         `yaml:"task"`
	Alias       string         `yaml:"alias,omitempty"`
	Skip        bool           `yaml:"skip,omitempty"`
	Amount      uint64         `yaml:"amount,omitempty"`
	Script      *Script        `yaml:"script,omitempty"` // deploy only
	Params      *Params        `yaml:"params,omitempty"` // call only
	Source      string         `yaml:"source,omitempty"`
	Destination string         `yaml:"destination,omitempty"`
	Args        map[string]any `yaml:"args,omitempty"`     // token_* only
	Contents    []Task         `yaml:"contents,omitempty"` // batch only
	WaitMode    WaitMode       `yaml:"for,omitempty"`      // wait only
	Value       string         `yaml:"value,omitempty"`    // wait only
	Log         string         `yaml:"log,omitempty"`      // log level override
	OnError     ErrorMode      `yaml:"on_error,omitempty"` // how to handle errors: fail|warn|ignore
}

func (t Task) Validate(ctx compose.Context) error {
	if t.Type == "" {
		return compose.ErrNoTaskType
	}
	if t.Params != nil {
		if err := t.Params.Validate(ctx); err != nil {
			return fmt.Errorf("params: %v", err)
		}
	}
	if _, ok := ctx.Variables[t.Alias]; !ok && t.Alias != "" {
		ctx.AddVariable(t.Alias, tezos.ZeroAddress.String())
	}
	if t.Script != nil {
		if err := t.Script.Validate(ctx); err != nil {
			return fmt.Errorf("script: %v", err)
		}
	}
	if t.Type == "batch" {
		if len(t.Contents) == 0 {
			return fmt.Errorf("empty contents")
		}
		for _, v := range t.Contents {
			if v.Type == "batch" {
				return fmt.Errorf("nested batch tasks not allowed")
			}
			if v.Contents != nil {
				return fmt.Errorf("nested contents not allowed")
			}
			if v.Source != "" && v.Source != t.Source {
				return fmt.Errorf("switching source is not allowed in batch contents")
			}
			if err := v.Validate(ctx); err != nil {
				return fmt.Errorf("contents: %v", err)
			}
		}
	} else if t.Contents != nil {
		return fmt.Errorf("contents only allowed in batch tasks")
	}
	return nil
}

type ErrorMode byte

const (
	ErrorModeFail ErrorMode = iota
	ErrorModeWarn
	ErrorModeIgnore
)

func (m *ErrorMode) UnmarshalYAML(node *yaml.Node) error {
	return m.UnmarshalText([]byte(node.Value))
}

func (m *ErrorMode) UnmarshalText(buf []byte) error {
	switch string(buf) {
	case "fail":
		*m = ErrorModeFail
	case "warn":
		*m = ErrorModeWarn
	case "ignore":
		*m = ErrorModeIgnore
	default:
		return fmt.Errorf("invalid error mode %q", string(buf))
	}
	return nil
}

type WaitMode byte

const (
	WaitModeInvalid WaitMode = iota
	WaitModeCycle
	WaitModeBlock
	WaitModeTime
)

func (m *WaitMode) UnmarshalYAML(node *yaml.Node) error {
	return m.UnmarshalText([]byte(node.Value))
}

func (m *WaitMode) UnmarshalText(buf []byte) error {
	switch string(buf) {
	case "cycle":
		*m = WaitModeCycle
	case "block":
		*m = WaitModeBlock
	case "time":
		*m = WaitModeTime
	case "":
		*m = WaitModeInvalid
	default:
		return fmt.Errorf("invalid wait mode %q", string(buf))
	}
	return nil
}

type ValueSource struct {
	File  string `yaml:"file,omitempty"`
	Url   string `yaml:"url,omitempty"`
	Value string `yaml:"value,omitempty"`
}

func (v ValueSource) IsUsed() bool {
	return len(v.Url)+len(v.File)+len(v.Value) > 0
}

func (v ValueSource) Validate(ctx compose.Context) error {
	if !v.IsUsed() {
		return fmt.Errorf("required file, url or value")
	}
	switch {
	case v.Url != "":
		if u, err := url.Parse(v.Url); err != nil {
			return fmt.Errorf("url: %v", err)
		} else if u.Host == "" {
			return fmt.Errorf("missing url host")
		}
	case v.File != "":
		fname, _, _ := strings.Cut(v.File, "#")
		fname = filepath.Join(ctx.Filepath(), fname)
		if _, err := os.Stat(fname); err != nil {
			return fmt.Errorf("file: %v", err)
		}
	case v.Value != "":
		if !isHex(v.Value) {
			var x any
			if err := json.Unmarshal([]byte(v.Value), &x); err != nil {
				return fmt.Errorf("json value: %v", err)
			}
		} else {
			buf, err := hex.DecodeString(v.Value)
			if err != nil {
				return fmt.Errorf("bin value: %v", err)
			}
			var prim micheline.Prim
			if err := prim.UnmarshalBinary(buf); err != nil {
				return fmt.Errorf("bin value: %v", err)
			}
		}
	}
	return nil
}

type Script struct {
	ValueSource `yaml:",inline"`
	Code        *Code    `yaml:"code,omitempty"`
	Storage     *Storage `yaml:"storage,omitempty"`
}

type Code struct {
	ValueSource `yaml:",inline"`
	// Patch       []Patch `yaml:"patch,omitempty"`
}

type Storage struct {
	ValueSource `yaml:",inline"`
	Args        any     `yaml:"args,omitempty"`
	Patch       []Patch `yaml:"patch,omitempty"`
}

func (s Script) Validate(ctx compose.Context) error {
	// either script-level source or both storage+code sources are required
	if s.ValueSource.IsUsed() {
		if err := s.ValueSource.Validate(ctx); err != nil {
			return err
		}
	} else {
		if s.Code == nil {
			return fmt.Errorf("missing code section")
		}
		if err := s.Code.ValueSource.Validate(ctx); err != nil {
			return err
		}
		if s.Storage == nil {
			return fmt.Errorf("missing storage section")
		}
		if s.Storage.Args == nil {
			// without args, storage must come from a valid source
			if err := s.Storage.ValueSource.Validate(ctx); err != nil {
				return fmt.Errorf("storage: %v", err)
			}
		}
	}
	// if s.Code != nil {
	// 	for _, v := range s.Code.Patch {
	// 		if err := v.Validate(ctx); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	if s.Storage != nil {
		for _, v := range s.Storage.Patch {
			if err := v.Validate(ctx); err != nil {
				return err
			}
		}
		if err := checkArgs(ctx, s.Storage.Args); err != nil {
			return err
		}
	}
	return nil
}

func checkArgs(ctx compose.Context, args any) error {
	if args == nil {
		return nil
	}
	switch v := args.(type) {
	case map[string]any:
		for n, v := range v {
			if err := checkArgs(ctx, v); err != nil {
				return fmt.Errorf("arg %s: %v", n, err)
			}
		}
	case []any:
		for i, v := range v {
			if err := checkArgs(ctx, v); err != nil {
				return fmt.Errorf("arg %d: %v", i, err)
			}
		}
	case string:
		if _, err := ctx.ResolveString(v); err != nil {
			return err
		}
	case int:
		// skip
	case bool:
		// skip
	default:
		return fmt.Errorf("unchecked arg type %T", args)
	}
	return nil
}

type Params struct {
	ValueSource `yaml:",inline"`
	Entrypoint  string  `yaml:"entrypoint"`
	Args        any     `yaml:"args,omitempty"`
	Patch       []Patch `yaml:"patch,omitempty"`
}

func (p Params) Validate(ctx compose.Context) error {
	if p.Args == nil {
		if err := p.ValueSource.Validate(ctx); err != nil && p.Args == nil {
			return err
		}
	} else {
		if err := checkArgs(ctx, p.Args); err != nil {
			return err
		}
	}
	if p.Entrypoint == "" {
		return fmt.Errorf("missing entrypoint")
	}
	for _, v := range p.Patch {
		if err := v.Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Patch struct {
	Type      string  `yaml:"type"`
	Key       *string `yaml:"key,omitempty"`
	Path      *string `yaml:"path,omitempty"`
	Value     *string `yaml:"value"`
	Optimized bool    `yaml:"optimized"`
}

func (p Patch) Validate(ctx compose.Context) error {
	// accept empty string
	if p.Key == nil && p.Path == nil {
		return fmt.Errorf("patch: required key or path")
	}
	// accept empty string
	if p.Value == nil {
		return fmt.Errorf("patch: required value")
	}
	// type must be correct
	oc, err := micheline.ParseOpCode(p.Type)
	if err != nil {
		return fmt.Errorf("patch: %v", err)
	}
	if !oc.IsTypeCode() {
		return fmt.Errorf("patch: %s is not a valid type code", p.Type)
	}
	// value must resolve
	val, err := ctx.ResolveString(*p.Value)
	if err != nil {
		return fmt.Errorf("patch: %v", err)
	}
	// value must parse against type code
	if _, err := compose.ParseValue(oc, val); err != nil {
		return fmt.Errorf("patch: %v", err)
	}
	return nil
}
