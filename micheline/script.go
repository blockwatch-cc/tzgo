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
			Param:   Prim{Type: PrimSequence, Args: []Prim{Prim{Type: PrimUnary, OpCode: K_PARAMETER}}},
			Storage: Prim{Type: PrimSequence, Args: []Prim{Prim{Type: PrimUnary, OpCode: K_STORAGE}}},
			Code:    Prim{Type: PrimSequence, Args: []Prim{Prim{Type: PrimUnary, OpCode: K_CODE}}},
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
			s.Code.Param = c.Args[0]
			s.Code.Storage = c.Args[1]
			s.Code.Code = c.Args[2]
			if len(c.Args) > 3 {
				s.Code.View = c.Args[3]
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
					*p = c
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

// Returns a list of bigmaps referenced by a contracts current storage. Note that
// in rare cases when storage type uses a T_OR branch above its bigmap type definitions
// and the relevant branch is inactive/hidden the storage value lacks bigmap
// references and this function will return an empty list, even though bigmaps exist.
func (s Script) BigmapsById() []int64 {
	ids := make([]int64, 0)
	stack := NewStack(s.Storage)
	_ = s.Code.Storage.Walk(func(p Prim) error {
		val := stack.Pop()
		if p.OpCode == T_BIG_MAP && val.IsValid() && val.Type == PrimInt {
			ids = append(ids, val.Int.Int64())
			return PrimSkip
		}
		switch p.OpCode {
		case T_OR, T_PAIR, T_OPTION, K_STORAGE:
			// recurse
			if val.IsScalar() {
				stack.Push(val)
			} else {
				stack.Push(val.Args...)
			}
			return nil
		default:
			// ignore
			return PrimSkip
		}
	})
	return ids
}

// Returns a named map containing all bigmaps currently referenced by a contracts
// storage value. Names are derived from Michelson type annotations and if missing,
// a sequence number. Optionally appends a sequence number to prevent duplicate names.
func (s Script) BigmapsByName() map[string]int64 {
	ids := s.BigmapsById()
	named := make(map[string]int64)
	bigmaps, _ := s.Code.Storage.FindOpCodes(T_BIG_MAP)
	for i := 0; i < min(len(ids), len(bigmaps)); i++ {
		n := bigmaps[i].GetVarAnnoAny()
		if n == "" {
			n = strconv.Itoa(i)
		}
		if _, ok := named[n]; ok {
			n += "_" + strconv.Itoa(i)
		}
		named[n] = ids[i]
	}
	return named
}

// Returns a named map containing all bigmaps defined in contracts storgae spec.
// Names are derived from Michelson type annotations and if missing,
// a sequence number. Optionally appends a sequence number to prevent duplicate names.
func (s Script) BigmapTypesByName() map[string]Type {
	named := make(map[string]Type)
	bigmaps, _ := s.Code.Storage.FindOpCodes(T_BIG_MAP)
	for i := range bigmaps {
		n := bigmaps[i].GetVarAnnoAny()
		if n == "" {
			n = strconv.Itoa(i)
		}
		if _, ok := named[n]; ok {
			n += "_" + strconv.Itoa(i)
		}
		named[n] = NewType(bigmaps[i])
	}
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
	binary.BigEndian.PutUint32(res[:], uint32(len(res)-4))

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
		log.Warnf("micheline: unexpected program tag 0x%x", prim.Type)
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
