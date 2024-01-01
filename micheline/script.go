// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"blockwatch.cc/tzgo/tezos"
)

type Script struct {
	Code    Code `json:"code"`    // code section, i.e. parameter & storage types, code
	Storage Prim `json:"storage"` // data section, i.e. initial contract storage
}

type Code struct {
	Param   Prim // call types
	Storage Prim // storage types
	Code    Prim // program code
	View    Prim // view code (i.e. list of views, may be empty)
	BadCode Prim // catch-all for ill-formed contracts
}

func NewScript() *Script {
	return &Script{
		Code: Code{
			Param:   Prim{Type: PrimSequence, Args: []Prim{{Type: PrimUnary, OpCode: K_PARAMETER}}},
			Storage: Prim{Type: PrimSequence, Args: []Prim{{Type: PrimUnary, OpCode: K_STORAGE}}},
			Code:    Prim{Type: PrimSequence, Args: []Prim{{Type: PrimUnary, OpCode: K_CODE}}},
			View:    Prim{Type: PrimSequence, Args: []Prim{}},
		},
		Storage: Prim{},
	}
}

func (s Script) IsValid() bool {
	return s.Code.Param.IsValid() && s.Code.Storage.IsValid()
}

func (s Script) StorageType() Type {
	return Type{s.Code.Storage.Args[0]}
}

func (s Script) ParamType() Type {
	return Type{s.Code.Param.Args[0]}
}

func (s Script) Entrypoints(withPrim bool) (Entrypoints, error) {
	return s.ParamType().Entrypoints(withPrim)
}

func (s Script) ResolveEntrypointPath(name string) string {
	return s.ParamType().ResolveEntrypointPath(name)
}

func (s Script) Views(withPrim, withCode bool) (Views, error) {
	views := make(Views, len(s.Code.View.Args))
	for _, v := range s.Code.View.Args {
		view := NewView(v)
		if !withPrim {
			view.Prim = InvalidPrim
		}
		if !withCode {
			view.Code = InvalidPrim
		}
		views[view.Name] = view
	}
	return views, nil
}

func (s Script) Constants() []tezos.ExprHash {
	c := make([]tezos.ExprHash, 0)
	for _, prim := range []Prim{
		s.Code.Param,
		s.Code.Storage,
		s.Code.Code,
		s.Code.View,
		s.Code.BadCode,
	} {
		prim.Walk(func(p Prim) error {
			if p.IsConstant() {
				if h, err := tezos.ParseExprHash(p.Args[0].String); err == nil {
					c = append(c, h)
				}
			}
			return nil
		})
	}
	return c
}

func (s *Script) ExpandConstants(dict ConstantDict) {
	// first check if the entire script is a constant
	if s.Code.BadCode.IsConstant() {
		if c, ok := dict.GetString(s.Code.BadCode.Args[0].String); ok {
			// replace entire code section from constant
			s.Code.Param = c.Args[0].Clone()
			s.Code.Storage = c.Args[1].Clone()
			s.Code.Code = c.Args[2].Clone()
			if len(c.Args) > 3 {
				for _, view := range c.Args[3:] {
					s.Code.View.Args = append(s.Code.View.Args, view.Clone())
				}
			}
		}
		s.Code.BadCode = Prim{}
	}
	// continue replacing nested constants
	for _, prim := range []*Prim{
		&s.Code.Param,
		&s.Code.Storage,
		&s.Code.Code,
		&s.Code.View,
	} {
		_ = prim.Visit(func(p *Prim) error {
			if p.IsConstant() {
				if c, ok := dict.GetString(p.Args[0].String); ok {
					*p = c.Clone()
				}
			}
			return nil
		})
	}
}

// Returns the first 4 bytes of the SHA256 hash from a binary encoded parameter type
// definition. This value is sufficiently unique to identify contracts with exactly
// the same entrypoints including annotations.
//
// To identify syntactically equal entrypoints with or without annotations use
// `IsEqual()`, `IsEqualWithAnno()` or `IsEqualPrim()`.
func (s Script) InterfaceHash() uint64 {
	return s.Code.Param.Hash64()
}

// Returns the first 4 bytes of the SHA256 hash from a binary encoded storage type
// definition. This value is sufficiently unique to identify contracts with exactly
// the same entrypoints including annotations.
func (s Script) StorageHash() uint64 {
	return s.Code.Storage.Hash64()
}

// Returns the first 4 bytes of the SHA256 hash from a binary encoded code section
// of a contract.
func (s Script) CodeHash() uint64 {
	return s.Code.Code.Hash64()
}

// Returns named bigmap ids from the script's storage type and current value.
func (s Script) Bigmaps() map[string]int64 {
	return DetectBigmaps(s.Code.Storage, s.Storage)
}

// Returns a map of named bigmap ids obtained from a storage type and a storage value.
// In the edge case where a T_OR branch hides an exsting bigmap behind a None value,
// the hidden bigmap is not detected.
func DetectBigmaps(typ, storage Prim) map[string]int64 {
	named := make(map[string]int64)
	uniqueName := func(n string) string {
		if _, ok := named[n]; !ok && n != "" {
			return n
		}
		if n == "" {
			n = "bigmap"
		}
		for i := 0; ; i++ {
			name := n + "_" + strconv.Itoa(i)
			if _, ok := named[name]; ok {
				continue
			}
			return name
		}
	}
	stack := NewStack(storage)
	_ = typ.Walk(func(p Prim) error {
		val := stack.Pop()
		switch p.OpCode {
		case T_BIG_MAP:
			if val.IsValid() && val.Type == PrimInt {
				named[uniqueName(p.GetVarAnnoAny())] = val.Int.Int64()
			}
			return PrimSkip

		case K_STORAGE:
			stack.Push(val)
			return nil

		case T_LAMBDA:
			// unsupported
			return PrimSkip

		case T_LIST, T_SET:
			for i, p := range val.Args {
				for n, v := range DetectBigmaps(typ.Args[0], p) {
					n = n + "_" + strconv.Itoa(i)
					named[uniqueName(n)] = v
				}
			}
			return PrimSkip

		case T_OR:
			branch := p.Args[0]
			if val.OpCode == D_RIGHT {
				branch = p.Args[1]
			}
			if len(val.Args) > 0 {
				for n, v := range DetectBigmaps(branch, val.Args[0]) {
					named[uniqueName(n)] = v
				}
			}
			return PrimSkip

		case T_OPTION:
			if val.OpCode == D_SOME {
				stack.Push(val.Args...)
				return nil
			}
			return PrimSkip

		case T_MAP:
			if p.Args[1].OpCode != T_BIG_MAP {
				return PrimSkip
			}
			for i, v := range val.Args {
				if v.OpCode != D_ELT || v.Args[1].Type != PrimInt {
					break
				}
				var name string
				switch v.Args[0].Type {
				case PrimString:
					name = v.Args[0].String
				case PrimBytes:
					buf := v.Args[0].Bytes
					if isASCIIBytes(buf) {
						name = string(buf)
					} else if tezos.IsAddressBytes(buf) {
						a := tezos.Address{}
						_ = a.Decode(buf)
						name = a.String()
					}
				}
				if name == "" {
					name = p.GetVarAnnoAny() + "_" + strconv.Itoa(i)
				}
				named[uniqueName(name)] = v.Args[1].Int.Int64()
			}
			return PrimSkip

		case T_PAIR:
			switch {
			case val.IsScalar() || val.LooksLikeContainer():
				stack.Push(val)
			default:
				stack.Push(val.Args...)
			}
			return nil

		default:
			return PrimSkip
		}
	})
	return named
}

// Returns a map of all known bigmap type definitions inside the scripts storage type.
// Unlabeled bigmaps are prefixed `bigmap_` followed by a unique sequence number.
// Duplicate names are prevented by adding by a unique sequence number as well.
func (s Script) BigmapTypes() map[string]Type {
	return DetectBigmapTypes(s.Code.Storage)
}

// Returns a map of all known bigmap type definitions inside a given prim. Keys are
// derived from type annotations. Unlabeled bigmaps are prefixed `bigmap_` followed
// by a unique sequence number. Duplicate names are prevented by adding by
// a unique sequence number as well.
func DetectBigmapTypes(typ Prim) map[string]Type {
	named := make(map[string]Type)
	uniqueName := func(n string) string {
		if _, ok := named[n]; !ok && n != "" {
			return n
		}
		if n == "" {
			n = "bigmap"
		}
		for i := 0; ; i++ {
			name := n + "_" + strconv.Itoa(i)
			if _, ok := named[name]; ok {
				continue
			}
			return name
		}
	}
	_ = typ.Walk(func(p Prim) error {
		switch p.OpCode {
		case T_BIG_MAP:
			named[uniqueName(p.GetVarAnnoAny())] = NewType(p)
			return PrimSkip
		case T_MAP:
			if p.Args[1].OpCode != T_BIG_MAP {
				return PrimSkip
			}
			name := p.GetVarAnnoAny()
			if n := p.Args[1].GetVarAnnoAny(); n != "" {
				name = n
			}
			named[uniqueName(name)] = NewType(p.Args[1])
			return PrimSkip
		case T_LIST:
			if p.Args[0].OpCode != T_BIG_MAP {
				return PrimSkip
			}
			name := p.GetVarAnnoAny()
			if n := p.Args[0].GetVarAnnoAny(); n != "" {
				name = n
			}
			named[uniqueName(name)] = NewType(p.Args[0])
			return PrimSkip

		case T_LAMBDA:
			return PrimSkip

		default:
			// return PrimSkip
			return nil
		}
	})

	return named
}

func (p Script) EncodeBuffer(buf *bytes.Buffer) error {
	// 1 write code segment
	code, err := p.Code.MarshalBinary()
	if err != nil {
		return err
	}

	// 2 write data segment
	data, err := p.Storage.MarshalBinary()
	if err != nil {
		return err
	}

	// append to output buffer
	buf.Write(code)

	// write data size
	binary.Write(buf, binary.BigEndian, uint32(len(data)))

	// append to output buffer
	buf.Write(data)

	return nil
}

func (p *Script) DecodeBuffer(buf *bytes.Buffer) error {
	// 1 Code
	if err := p.Code.DecodeBuffer(buf); err != nil {
		return err
	}

	// 2 Storage

	// check storage is present
	if buf.Len() < 4 {
		return io.ErrShortBuffer
	}

	// starts with BE uint32 total size
	size := int(binary.BigEndian.Uint32(buf.Next(4)))
	if buf.Len() < size {
		return io.ErrShortBuffer
	}

	// read primitive tree
	n := buf.Len()
	if err := p.Storage.DecodeBuffer(buf); err != nil {
		return err
	}

	// check we've read the defined amount of bytes
	read := n - buf.Len()
	if size != read {
		return fmt.Errorf("micheline: expected script size %d but read %d bytes", size, read)
	}

	return nil
}

func (p Script) MarshalJSON() ([]byte, error) {
	type alias Script
	return json.Marshal(alias(p))
}

func (p Script) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := p.EncodeBuffer(buf)
	return buf.Bytes(), err
}

func (p *Script) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	err := p.DecodeBuffer(buf)
	if err != nil {
		return err
	}
	if buf.Len() > 0 {
		return fmt.Errorf("micheline: %d unexpected extra trailer bytes", buf.Len())
	}
	return nil
}

func (c Code) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	// keep space for size
	binary.Write(buf, binary.BigEndian, uint32(0))

	// root element is a sequence
	root := Prim{
		Type: PrimSequence,
		Args: []Prim{c.Param, c.Storage, c.Code},
	}

	if len(c.View.Args) > 0 {
		root.Args = append(root.Args, c.View.Args...)
	}

	// store ill-formed contracts
	if c.BadCode.IsValid() {
		root = Prim{
			Type: PrimSequence,
			Args: []Prim{EmptyPrim, EmptyPrim, EmptyPrim, c.BadCode},
		}
	}

	if err := root.EncodeBuffer(buf); err != nil {
		return nil, err
	}

	// patch code size
	res := buf.Bytes()
	binary.BigEndian.PutUint32(res, uint32(len(res)-4))

	return res, nil
}

func (c *Code) UnmarshalBinary(data []byte) error {
	return c.DecodeBuffer(bytes.NewBuffer(data))
}

func (c *Code) DecodeBuffer(buf *bytes.Buffer) error {
	// starts with BE uint32 total size
	size := int(binary.BigEndian.Uint32(buf.Next(4)))
	if buf.Len() < size {
		return io.ErrShortBuffer
	}

	// read primitive tree
	var prim Prim
	if err := prim.DecodeBuffer(buf); err != nil {
		return err
	}

	// check for sequence tag
	if prim.Type != PrimSequence {
		return fmt.Errorf("micheline: unexpected program tag 0x%x", prim.Type)
	}

	// unpack keyed program parts
	for _, v := range prim.Args {
		switch v.OpCode {
		case K_PARAMETER:
			c.Param = v
		case K_STORAGE:
			c.Storage = v
		case K_CODE:
			c.Code = v
		case K_VIEW:
			// append to view list
			c.View.Args = append(c.View.Args, v)
		case 255:
			c.BadCode = v
		default:
			return fmt.Errorf("micheline: unexpected program key 0x%x", v.OpCode)
		}
	}
	return nil
}

// UnmarshalScriptType is an optimized binary unmarshaller which decodes type trees only.
// Use this to access smart contract types when script and storage are not required.
func UnmarshalScriptType(data []byte) (param Type, storage Type, err error) {
	buf := bytes.NewBuffer(data)

	// starts with BE uint32 total size
	size := int(binary.BigEndian.Uint32(buf.Next(4)))
	if buf.Len() < size {
		err = io.ErrShortBuffer
		return
	}

	// we expect a sequence
	b := buf.Next(1)
	if len(b) == 0 {
		err = io.ErrShortBuffer
		return
	}

	// check tag
	if PrimType(b[0]) != PrimSequence {
		err = fmt.Errorf("micheline: unexpected program tag 0x%x", b[0])
		return
	}

	// cross-check content size
	size = int(binary.BigEndian.Uint32(buf.Next(4)))
	if buf.Len() < size {
		err = io.ErrShortBuffer
		return
	}

	// decode K_PARAMETER primitive and unwrap outer prim
	var pPrim Prim
	if err = pPrim.DecodeBuffer(buf); err != nil {
		return
	}
	param.Prim = pPrim.Args[0]

	// decode K_STORAGE primitive and unwrap outer prim
	var sPrim Prim
	if err = sPrim.DecodeBuffer(buf); err != nil {
		return
	}
	storage.Prim = sPrim.Args[0]
	return
}

func (c Code) MarshalJSON() ([]byte, error) {
	root := Prim{
		Type: PrimSequence,
		Args: []Prim{c.Param, c.Storage, c.Code},
	}
	if len(c.View.Args) > 0 {
		root.Args = append(root.Args, c.View.Args...)
	}
	if c.BadCode.IsValid() {
		root = c.BadCode
	}
	return json.Marshal(root)
}

func (c *Code) UnmarshalJSON(data []byte) error {
	// read primitive tree
	var prim Prim
	if err := json.Unmarshal(data, &prim); err != nil {
		return err
	}

	// check for sequence tag
	if prim.Type != PrimSequence {
		c.BadCode = prim
		return nil
	}

	// unpack keyed program parts
	isBadCode := false
stopcode:
	for _, v := range prim.Args {
		switch v.OpCode {
		case K_PARAMETER:
			c.Param = v
		case K_STORAGE:
			c.Storage = v
		case K_CODE:
			c.Code = v
		case K_VIEW:
			c.View.Args = append(c.View.Args, v)
		default:
			isBadCode = true
			log.Warnf("micheline: unexpected program key 0x%x (%d)", byte(v.OpCode), v.OpCode)
			break stopcode
		}
	}
	if isBadCode {
		c.BadCode = prim
	}
	return nil
}
