// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Parameters struct {
	Entrypoint string `json:"entrypoint"`
	Value      Prim   `json:"value"`
}

func (p Parameters) MarshalJSON() ([]byte, error) {
	if p.Entrypoint == "" || (p.Entrypoint == "default" && p.Value.OpCode == D_UNIT) {
		return json.Marshal(p.Value)
	}
	type alias Parameters
	return json.Marshal(alias(p))
}

func (p Parameters) MapEntrypoint(typ Type) (Entrypoint, Prim, error) {
	var ep Entrypoint
	var ok bool
	var prim Prim

	// get list of script entrypoints
	eps, _ := typ.Entrypoints(true)

	switch p.Entrypoint {
	case "default":
		// rebase branch by prepending the path to the named default entrypoint
		prefix := typ.ResolveEntrypointPath("default")
		// can be [LR]+ or empty when entrypoint is used
		branch := p.Branch(prefix, eps)
		ep, ok = eps.FindBranch(branch)
		if !ok {
			ep, _ = eps.FindId(0)
			prim = p.Value
		} else {
			prim = p.Unwrap(strings.TrimPrefix(ep.Branch, prefix))
		}

	case "root", "":
		// search unnamed naked entrypoint
		branch := p.Branch("", eps)
		ep, ok = eps.FindBranch(branch)
		if !ok {
			ep, _ = eps.FindId(0)
		}
		prim = p.Unwrap(ep.Branch)

	default:
		// search for named entrypoint
		ep, ok = eps[p.Entrypoint]
		if !ok {
			// entrypoint can be a combination of an annotated branch and more T_OR branches
			// inside parameters, so lets find the named branch
			prefix := typ.ResolveEntrypointPath(p.Entrypoint)
			if prefix == "" {
				// meh
				return ep, prim, fmt.Errorf("micheline: missing entrypoint '%s'", p.Entrypoint)
			}
			// otherwise rebase using the annotated branch as prefix
			branch := p.Branch(prefix, eps)
			ep, ok = eps.FindBranch(branch)
			if !ok {
				return ep, prim, fmt.Errorf("micheline: missing entrypoint '%s' + %s", p.Entrypoint, prefix)
			}
			// unwrap the suffix branch only
			prim = p.Unwrap(strings.TrimPrefix(ep.Branch, prefix))
		} else {
			prim = p.Value
		}
	}
	return ep, prim, nil
}

func (p Parameters) Branch(prefix string, eps Entrypoints) string {
	node := p.Value
	if !node.IsValid() {
		return ""
	}
	branch := prefix
done:
	for {
		switch node.OpCode {
		case D_LEFT:
			branch += "/L"
		case D_RIGHT:
			branch += "/R"
		default:
			break done
		}
		node = node.Args[0]
		if _, ok := eps.FindBranch(branch); ok {
			break done
		}
	}
	return branch
}

func (p Parameters) Unwrap(branch string) Prim {
	node := p.Value
	branch = strings.TrimPrefix(branch, "/")
	branch = strings.TrimSuffix(branch, "/")
	for _, v := range strings.Split(branch, "/") {
		if !node.IsValid() {
			break
		}
		switch v {
		case "L":
			node = node.Args[0]
		case "R":
			node = node.Args[0]
		}
	}
	return node
}

func (p Parameters) EncodeBuffer(buf *bytes.Buffer) error {
	// marshal value first to catch any error before writing to buffer
	val, err := p.Value.MarshalBinary()
	if err != nil {
		return err
	}

	switch p.Entrypoint {
	case "", "default":
		buf.WriteByte(0)
	case "root":
		buf.WriteByte(1)
	case "do":
		buf.WriteByte(2)
	case "set_delegate":
		buf.WriteByte(3)
	case "remove_delegate":
		buf.WriteByte(4)
	default:
		buf.WriteByte(255)
		buf.WriteByte(byte(len(p.Entrypoint)))
		buf.WriteString(p.Entrypoint)
	}
	binary.Write(buf, binary.BigEndian, uint32(len(val)))
	buf.Write(val)
	return nil
}

func (p Parameters) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := p.EncodeBuffer(buf)
	return buf.Bytes(), err
}

func (p *Parameters) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '[' {
		// non-entrypoint calling convention
		return json.Unmarshal(data, &p.Value)
	} else {
		// try entrypoint calling convention
		type alias *Parameters
		if err := json.Unmarshal(data, alias(p)); err != nil {
			return err
		}
		if p.Value.IsValid() {
			return nil
		}
		// try legacy calling convention for single prim values
		p.Entrypoint = "default"
		return json.Unmarshal(data, &p.Value)
	}
}

func (p *Parameters) DecodeBuffer(buf *bytes.Buffer) error {
	if buf.Len() < 1 {
		return io.ErrShortBuffer
	}
	tag := buf.Next(1)
	if len(tag) == 0 {
		return io.ErrShortBuffer
	}
	switch tag[0] {
	case 0:
		p.Entrypoint = "default"
	case 1:
		p.Entrypoint = "root"
	case 2:
		p.Entrypoint = "do"
	case 3:
		p.Entrypoint = "set_delegate"
	case 4:
		p.Entrypoint = "remove_delegate"
	default:
		sz := buf.Next(1)
		if len(sz) == 0 || buf.Len() < int(sz[0]) {
			return io.ErrShortBuffer
		}
		p.Entrypoint = string(buf.Next(int(sz[0])))
	}
	if buf.Len() == 0 {
		p.Value = Prim{Type: PrimNullary, OpCode: D_UNIT}
		return nil
	}
	size := int(binary.BigEndian.Uint32(buf.Next(4)))
	if buf.Len() < size {
		return io.ErrShortBuffer
	}
	prim := Prim{}
	if err := prim.DecodeBuffer(buf); err != nil {
		return err
	}
	p.Value = prim
	return nil
}

func (p *Parameters) UnmarshalBinary(data []byte) error {
	return p.DecodeBuffer(bytes.NewBuffer(data))
}
