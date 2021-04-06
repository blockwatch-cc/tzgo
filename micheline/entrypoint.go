// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"fmt"
	"strings"
)

type Entrypoint struct {
	Id      int       `json:"id"`
	Call    string    `json:"call"`
	Branch  string    `json:"branch"`
	Typedef []Typedef `json:"type"`
	Prim    *Prim     `json:"prim,omitempty"`
}

func (e Entrypoint) Type() Type {
	if e.Prim == nil {
		e.Prim = &Prim{} // invalid
	}
	return Type{*e.Prim}
}

type Entrypoints map[string]Entrypoint

func (e Entrypoints) FindBranch(branch string) (Entrypoint, bool) {
	if branch == "" {
		return Entrypoint{}, false
	}
	for _, v := range e {
		if v.Branch == branch {
			return v, true
		}
	}
	return Entrypoint{}, false
}

func (e Entrypoints) FindId(id int) (Entrypoint, bool) {
	for _, v := range e {
		if v.Id == id {
			return v, true
		}
	}
	return Entrypoint{}, false
}

func (s *Script) Entrypoints(withPrim bool) (Entrypoints, error) {
	e := make(Entrypoints)
	if s == nil || len(s.Code.Param.Args) == 0 {
		return e, nil
	}
	if err := listEntrypoints(e, "", s.Code.Param.Args[0]); err != nil {
		return nil, err
	}
	if !withPrim {
		for n, v := range e {
			v.Prim = nil
			e[n] = v
		}
	}
	return e, nil
}

// returns path to named entrypoint
func (s *Script) SearchEntrypointName(name string) string {
	if s == nil || len(s.Code.Param.Args) == 0 {
		return ""
	}
	return searchEntrypointName(name, "", s.Code.Param.Args[0])
}

func searchEntrypointName(name, branch string, node Prim) string {
	if node.GetVarAnnoAny() == name {
		return branch
	}
	if node.OpCode == T_OR && (len(branch) == 0 || !node.HasAnno()) {
		b := searchEntrypointName(name, branch+"/L", node.Args[0])
		if b != "" {
			return b
		}
		b = searchEntrypointName(name, branch+"/R", node.Args[1])
		if b != "" {
			return b
		}
	}
	return ""
}

// Explicit list of prefixes for detecting entrypoints.
//
// This is necessary to resolve ambiguities in contract designs that
// use T_OR as call parameter.

// - to handle conflicts between T_OR used for call params vs used for marking entrypoint
//   skip annotated T_OR branches (exclude the root T_OR and any branch called 'default')
var knownEntrypointPrefixes = []string{"_Liq_entry_"}

func isKnownEntrypointPrefix(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, v := range knownEntrypointPrefixes {
		if strings.HasPrefix(s, v) {
			return true
		}
	}
	return false
}

// walks T_OR expressions and stores each non-T_OR branch as entrypoint
func listEntrypoints(e Entrypoints, branch string, node Prim) error {
	// prefer % annotations
	name := node.GetVarAnnoAny()
	if node.OpCode == T_OR && !isKnownEntrypointPrefix(name) {
		if l := len(node.Args); l != 2 {
			return fmt.Errorf("micheline: expected 2 arguments for T_OR, got %d", l)
		}

		if err := listEntrypoints(e, branch+"/L", node.Args[0]); err != nil {
			return err
		}

		if err := listEntrypoints(e, branch+"/R", node.Args[1]); err != nil {
			return err
		}

		return nil
	}

	// need unique entrypoint name
	if name == "" {
		name = fmt.Sprintf("__entry_%02d__", len(e))
	}

	// process non-T_OR branches
	cp := node.Clone()
	ep := Entrypoint{
		Id:      len(e),
		Branch:  branch,
		Call:    name,
		Typedef: buildTypedef("", node).Args,
		Prim:    &cp,
	}
	// if cp.HasAnno() {
	// 	// drop entrypoint name annotation, keep any other annots (in case a single
	// 	// value entrypoint has another variable name)
	// 	ep.Prim.StripAnno(name)
	// }

	e[name] = ep
	return nil
}
