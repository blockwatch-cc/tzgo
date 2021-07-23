// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func (p Prim) Dump() string {
	buf, _ := p.MarshalJSON()
	return string(buf)
}

func (p Prim) DumpLimit(n int) string {
	buf, _ := p.MarshalJSON()
	return limit(string(buf), n)
}

func (e Value) Dump() string {
	buf := bytes.NewBuffer(nil)
	e.DumpTo(buf)
	return string(buf.Bytes())
}

func (e Value) DumpLimit(n int) string {
	return limit(e.Dump(), n)
}

func (e Value) DumpTo(w io.Writer) {
	dumpTree(w, "", e.Type, e.Value)
}

func (s Stack) Dump() string {
	return s.DumpIdent(0)
}

func (s Stack) DumpIdent(indent int) string {
	idnt := strings.Repeat(" ", indent)
	lines := make([]string, s.Len())
	for i := range s {
		n := len(s) - i - 1
		lines[n] = idnt + fmt.Sprintf("%02d  ", n+1) + s[i].Dump()
	}
	return strings.Join(lines, "\n")
}

// TODO: improve tree output
func dumpTree(w io.Writer, path string, typ Type, val Prim) {
	if s, err := dump(path, typ, val); err != nil {
		io.WriteString(w, err.Error())
	} else {
		io.WriteString(w, s)
	}
	switch val.Type {
	case PrimSequence:
		// keep the type
		for i, v := range val.Args {
			p := path + "." + strconv.Itoa(i)
			dumpTree(w, p, typ, v)
		}
	default:
		// advance type as well
		for i, v := range val.Args {
			t := Type{}
			if len(typ.Args) > i {
				t = Type{typ.Args[i]}
			}
			p := path + "." + strconv.Itoa(i)
			dumpTree(w, p, t, v)
		}
	}
}

func dump(path string, typ Type, val Prim) (string, error) {
	if !val.matchOpCode(typ.OpCode) {
		return "", fmt.Errorf("Type mismatch val_type=%s type_code=%s", val.Type, typ.OpCode)
	}

	vtyp := "-"
	switch val.Type {
	case PrimSequence, PrimBytes, PrimInt, PrimString:
	default:
		vtyp = val.OpCode.String()
	}

	return fmt.Sprintf("path=%-20s val_prim=%-8s val_type=%-8s val_val=%-10s type_prim=%-8s type_code=%-8s type_name=%-8s\n",
		path, val.Type, vtyp, val.DumpLimit(512), typ.Type, typ.OpCode, typ.Label(),
	), nil
}
