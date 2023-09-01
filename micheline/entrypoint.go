// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"fmt"
	"strings"
)

type Entrypoint struct {
	Id      int       `json:"id"`
	Name    string    `json:"name"`
	Branch  string    `json:"branch"`
	Typedef []Typedef `json:"type"`
	Prim    *Prim     `json:"prim,omitempty"`
}

func (e Entrypoint) Type() Type {
	if e.Prim == nil {
		e.Prim = &Prim{} // invalid
	}
	typ := NewType(*e.Prim)
	if !typ.HasLabel() {
		typ.Prim.Anno = []string{"@" + e.Name}
	}
	return typ
}

func (e Entrypoint) IsCallback() bool {
	if e.Prim == nil {
		return false
	}
	return e.Prim.IsPair() && e.Prim.Args.Last().OpCode == T_CONTRACT
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

func (t Type) Entrypoints(withPrim bool) (Entrypoints, error) {
	e := make(Entrypoints)
	if !t.IsValid() {
		return e, nil
	}
	if err := listEntrypoints(e, "", t.Prim); err != nil {
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
func (t Type) ResolveEntrypointPath(name string) string {
	if !t.IsValid() {
		return ""
	}
	return resolveEntrypointPath(name, "", t.Prim)
}

func resolveEntrypointPath(name, branch string, node Prim) string {
	if node.GetVarAnnoAny() == name {
		return branch
	}
	if node.OpCode == T_OR {
		if b := resolveEntrypointPath(name, branch+"/L", node.Args[0]); b != "" {
			return b
		}
		if b := resolveEntrypointPath(name, branch+"/R", node.Args[1]); b != "" {
			return b
		}
	}
	return ""
}

// Explicit list of prefixes for detecting entrypoints.
//
// This is necessary to resolve ambiguities in contract designs that
// use T_OR as call parameter.

//   - to handle conflicts between T_OR used for call params vs used for marking entrypoint
//     skip annotated T_OR branches (exclude the root T_OR and any branch called 'default')
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
		if len(e) == 0 {
			name = DEFAULT
		} else {
			name = fmt.Sprintf("%s_%d", CONST_ENTRYPOINT, len(e))
		}
	}

	// process non-T_OR branches
	cp := node.Clone()
	ep := Entrypoint{
		Id:     len(e),
		Branch: branch,
		Name:   name,
		Prim:   &cp,
	}
	if node.IsScalarType() || node.IsContainerType() {
		ep.Typedef = []Typedef{buildTypedef("", node, []int{})}
		ep.Typedef[0].Name = ""
	} else {
		td := buildTypedef("", node, []int{})
		switch len(td.Args) {
		case 0:
			ep.Typedef = []Typedef{td}
		case 1:
			td.Name = ""
			ep.Typedef = []Typedef{td}
		default:
			ep.Typedef = td.Args
		}
	}

	e[name] = ep
	return nil
}
