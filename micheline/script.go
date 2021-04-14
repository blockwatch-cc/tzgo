// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

type Script struct {
	Code    Code `json:"code"`    // code section, i.e. parameter & storage types, code
	Storage Prim `json:"storage"` // data section, i.e. initial contract storage
}

type Code struct {
	Param   Prim  // call types
	Storage Prim  // storage types
	Code    Prim  // program code
	BadCode *Prim // catch-all for ill-formed contracts
}

func NewScript() *Script {
	return &Script{
		Code: Code{
			Param:   Prim{Type: PrimSequence, Args: []Prim{Prim{Type: PrimUnary, OpCode: K_PARAMETER}}},
			Storage: Prim{Type: PrimSequence, Args: []Prim{Prim{Type: PrimUnary, OpCode: K_STORAGE}}},
			Code:    Prim{Type: PrimSequence, Args: []Prim{Prim{Type: PrimUnary, OpCode: K_CODE}}},
		},
		Storage: Prim{},
	}
}

func (s *Script) StorageType() Type {
	return Type{s.Code.Storage.Args[0]}
}

// Returns the first 4 bytes of the SHA256 hash from a binary encoded parameter type
// definition. This value is sufficiently unique to identify contracts with exactly
// the same entrypoints including annotations.
//
// To identify syntactically equal entrypoints with or without annotations use
// `IsEqual()`, `IsEqualWithAnno()` or `IsEqualPrim()`.
func (s *Script) InterfaceHash() []byte {
	buf, _ := s.Code.Param.MarshalBinary()
	h := sha256.Sum256(buf)
	return h[:4]
}

// Returns the first 4 bytes of the SHA256 hash from a binary encoded storage type
// definition. This value is sufficiently unique to identify contracts with exactly
// the same entrypoints including annotations.
func (s *Script) StorageHash() []byte {
	buf, _ := s.Code.Storage.MarshalBinary()
	h := sha256.Sum256(buf)
	return h[:4]
}

// Returns the first 4 bytes of the SHA256 hash from a binary encoded code section
// of a contract.
func (s *Script) CodeHash() []byte {
	buf, _ := s.Code.Code.MarshalBinary()
	h := sha256.Sum256(buf)
	return h[:4]
}

// Returns a list of bigmaps referenced by a contracts current storage. Note that
// in rare cases when storage type uses a T_OR branch above its bigmap type definitions
// and the relevant branch is inactive/hidden the storage value lacks bigmap
// references and this function will return an empty list, even though bigmaps exist.
func (s *Script) BigmapsById() []int64 {
	ids := make([]int64, 0)
	stack := NewStack(s.Storage)
	_ = s.Code.Storage.Walk(func(p Prim) error {
		val := stack.Pop()
		// fmt.Printf("WALK typ=%s\n", p.Dump())
		// fmt.Printf("WALK val=%s\n", val.Dump())
		if p.OpCode == T_BIG_MAP && val.IsValid() && val.Type == PrimInt {
			ids = append(ids, val.Int.Int64())
			return PrimSkip
		}
		switch p.OpCode {
		case T_OR, T_PAIR, T_OPTION, K_STORAGE:
			// recurse
			// fmt.Printf("WALK recurse\n")
			if val.IsScalar() {
				stack.Push(val)
			} else {
				stack.Push(val.Args...)
			}
			return nil
		default:
			// ignore
			// fmt.Printf("WALK skip\n")
			return PrimSkip
		}
	})
	return ids
}

// Returns a named map containing all bigmaps currently referenced by a contracts
// storage value. Names are derived from Michelson type annotations and if missing,
// a sequence number.
func (s *Script) BigmapsByName() map[string]int64 {
	ids := s.BigmapsById()
	named := make(map[string]int64)
	bigmaps, _ := s.Code.Storage.FindOpCodes(T_BIG_MAP)
	for i := 0; i < min(len(ids), len(bigmaps)); i++ {
		n := bigmaps[i].GetVarAnnoAny()
		if n == "" {
			n = strconv.Itoa(i)
		}
		named[n] = ids[i]
	}
	return named
}

func (p Script) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	// 1 write code segment
	code, err := p.Code.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// append to output buffer
	buf.Write(code)

	// 2 write data segment
	data, err := p.Storage.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// write data size
	binary.Write(buf, binary.BigEndian, uint32(len(data)))

	// append to output buffer
	buf.Write(data)

	return buf.Bytes(), nil
}

func (p *Script) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)

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
	if err := p.Storage.DecodeBuffer(buf); err != nil {
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

	// store ill-formed contracts
	if c.BadCode != nil {
		root = Prim{
			Type: PrimSequence,
			Args: []Prim{EmptyPrim, EmptyPrim, EmptyPrim, *c.BadCode},
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
		case 255:
			c.BadCode = &v
		default:
			return fmt.Errorf("micheline: unexpected program key 0x%x", v.OpCode)
		}
	}
	return nil
}

func (c Code) MarshalJSON() ([]byte, error) {
	root := Prim{
		Type: PrimSequence,
		Args: []Prim{c.Param, c.Storage, c.Code},
	}
	if c.BadCode != nil {
		root = *c.BadCode
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
		c.BadCode = &prim
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
		default:
			isBadCode = true
			log.Warnf("micheline: unexpected program key 0x%x (%d)", byte(v.OpCode), v.OpCode)
			break stopcode
		}
	}
	if isBadCode {
		c.BadCode = &prim
	}
	return nil
}
