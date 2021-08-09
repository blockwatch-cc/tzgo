// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"blockwatch.cc/tzgo/tezos"
)

var BigmapRefType = Prim{
	Type:   PrimNullary,
	OpCode: T_INT,
}

func NewBigmapRefType(anno string) Prim {
	r := Prim{
		Type:   PrimNullary,
		OpCode: T_INT,
	}
	if anno != "" {
		r.Anno = []string{"@" + anno}
	}
	return r
}

func NewBigmapRef(id int64) Prim {
	return Prim{
		Type: PrimInt,
		Int:  big.NewInt(id),
	}
}

type BigmapDiff []BigmapDiffElem

type BigmapDiffElem struct {
	Action    DiffAction
	Id        int64
	SourceId  int64 // used on copy
	DestId    int64 // used on copy
	KeyHash   tezos.ExprHash
	Key       Prim // works with any type
	Value     Prim
	KeyType   Prim // used on alloc/copy, uses Prim for native marshalling
	ValueType Prim // used on alloc/copy, uses Prim for native marshalling
}

func (e BigmapDiffElem) Encoding() PrimType {
	switch e.Action {
	case DiffActionRemove, DiffActionUpdate:
		return e.Key.OpCode.PrimType()
	case DiffActionAlloc, DiffActionCopy:
		return e.KeyType.OpCode.PrimType()
	default:
		// invalid
		return PrimBytes
	}
}

func (e BigmapDiffElem) GetKey(typ Type) Key {
	k, err := NewKey(typ, e.Key)
	if err != nil {
		log.Error(err)
	}
	return k
}

func (e BigmapDiffElem) GetKeyPtr(typ Type) *Key {
	k, err := NewKey(typ, e.Key)
	if err != nil {
		log.Error(err)
	}
	return &k
}

// TODO: lazy_storage_diff updates
func (e *BigmapDiffElem) UnmarshalJSON(data []byte) error {
	var val struct {
		Id        int64          `json:"big_map,string"`
		Action    DiffAction     `json:"action"`
		KeyType   Prim           `json:"key_type"`                   // alloc
		ValueType Prim           `json:"value_type"`                 // alloc
		Key       interface{}    `json:"key"`                        // update/remove
		KeyHash   tezos.ExprHash `json:"key_hash"`                   // update/remove
		Value     Prim           `json:"value"`                      // update
		SourceId  int64          `json:"source_big_map,string"`      // copy
		DestId    int64          `json:"destination_big_map,string"` // copy
	}
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}

	switch val.Action {
	case DiffActionUpdate, DiffActionRemove:
		// Note: on Edo key can be list or object
		switch key := val.Key.(type) {
		case map[string]interface{}:
			switch len(key) {
			case 0:
				// EMPTY_BIG_MAP opcode emits a remove action without key
				e.Key = Prim{
					Type:   PrimNullary,
					OpCode: I_EMPTY_BIG_MAP,
				}
			case 1:
				for n, v := range key {
					vv, ok := v.(string)
					if !ok {
						return fmt.Errorf("micheline: decoding bigmap key '%v': unexpected type %T", v, v)
					}
					switch n {
					case "int":
						p := Prim{
							Type: PrimInt,
							Int:  big.NewInt(0),
						}
						if err := p.Int.UnmarshalText([]byte(vv)); err != nil {
							return fmt.Errorf("micheline: decoding bigmap int key '%s': %w", v, err)
						}
						e.Key = p
					case "bytes":
						p := Prim{
							Type: PrimBytes,
						}
						p.Bytes, err = hex.DecodeString(vv)
						if err != nil {
							return fmt.Errorf("micheline: decoding bigmap bytes key '%s': %w", v, err)
						}
						e.Key = p
					case "string":
						e.Key = Prim{
							Type:   PrimString,
							String: vv,
						}
					case "prim":
						p := Prim{}
						if err := p.UnpackPrimitive(key); err != nil {
							return fmt.Errorf("micheline: decoding bigmap prim key: %w", err)
						}
						e.Key = p
					default:
						return fmt.Errorf("micheline: unsupported bigmap key type %s", n)
					}
				}
			default:
				p := Prim{}
				if err := p.UnpackPrimitive(key); err != nil {
					return fmt.Errorf("micheline: decoding bigmap pair key: %w", err)
				}
				e.Key = p
			}
		case []interface{}:
			p := Prim{}
			if err := p.UnpackSequence(key); err != nil {
				return fmt.Errorf("micheline: decoding bigmap list key: %w", err)
			}
			e.Key = p
		default:
			// EMPTY_BIG_MAP opcode emits a remove action without key
			if val.Key == nil {
				e.Key = Prim{
					Type:   PrimNullary,
					OpCode: I_EMPTY_BIG_MAP,
				}
			}
		}

		e.KeyHash = val.KeyHash
		e.Value = val.Value

	case DiffActionAlloc:
		if !val.KeyType.OpCode.IsValid() {
			return fmt.Errorf("micheline: unsupported bigmap key type (opcode) %s [%d]",
				e.KeyType.OpCode, e.KeyType.OpCode)
		}
		e.KeyType = val.KeyType
		e.ValueType = val.ValueType

	case DiffActionCopy:
		e.SourceId = val.SourceId
		e.DestId = val.DestId
	}

	// assign remaining values
	e.Id = val.Id
	e.Action = val.Action

	// pre-v005: set correct action on value deletion (missing value in JSON)
	if !e.Value.IsValid() && e.Action == DiffActionUpdate {
		e.Action = DiffActionRemove
	}

	return nil
}

func (e BigmapDiffElem) MarshalJSON() ([]byte, error) {
	var res interface{}
	switch e.Action {
	case DiffActionUpdate, DiffActionRemove:
		// set key, keyhash, value
		val := struct {
			Id      int64                  `json:"big_map,string"`
			Action  DiffAction             `json:"action"`
			Key     map[string]interface{} `json:"key,omitempty"`
			KeyHash *tezos.ExprHash        `json:"key_hash,omitempty"`
			Value   *Prim                  `json:"value,omitempty"`
		}{
			Id:     e.Id,
			Action: e.Action,
			Value:  &e.Value,
		}
		switch e.Key.Type {
		case PrimNullary:
			// no key on empty bigmap
		case PrimInt:
			val.Key = make(map[string]interface{})
			val.Key["int"] = e.Key.Int.Text(10)
			val.KeyHash = &e.KeyHash
		case PrimBytes:
			val.Key = make(map[string]interface{})
			val.Key["bytes"] = hex.EncodeToString(e.Key.Bytes)
			val.KeyHash = &e.KeyHash
		case PrimString:
			val.Key = make(map[string]interface{})
			val.Key["string"] = e.Key.String
			val.KeyHash = &e.KeyHash
		case PrimBinary:
			val.Key = make(map[string]interface{})
			val.KeyHash = &e.KeyHash
			buf, err := e.Key.MarshalJSON()
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(buf, &val.Key); err != nil {
				return nil, err
			}
		}

		// be API compatible with Babylon
		if e.Action == DiffActionRemove {
			val.Value = nil
		}
		res = val

	case DiffActionAlloc:
		res = struct {
			Id        int64      `json:"big_map,string"`
			Action    DiffAction `json:"action"`
			KeyType   Prim       `json:"key_type"`   // alloc, copy only; native Prim!
			ValueType Prim       `json:"value_type"` // alloc, copy only; native Prim!
		}{
			Id:        e.Id,
			Action:    e.Action,
			KeyType:   e.KeyType,
			ValueType: e.ValueType,
		}

	case DiffActionCopy:
		res = struct {
			Action   DiffAction `json:"action"`
			SourceId int64      `json:"source_big_map,string"`      // copy
			DestId   int64      `json:"destination_big_map,string"` // copy
		}{
			Action:   e.Action,
			SourceId: e.SourceId,
			DestId:   e.DestId,
		}
	}
	return json.Marshal(res)
}

func (b BigmapDiff) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	for _, v := range b {
		// prefix with id (4 byte) and action (1 byte)
		// temp bigmaps have negative numbers
		binary.Write(buf, binary.BigEndian, uint32(v.Id))
		buf.WriteByte(byte(v.Action))

		// encoding depends on action
		switch v.Action {
		case DiffActionUpdate, DiffActionRemove:
			// pair(pair(key_type, hash), value)
			key := Prim{
				Type:   PrimBinary,
				OpCode: T_PAIR,
				Args: []Prim{
					v.Key,
					Prim{
						Type:  PrimBytes,
						Bytes: v.KeyHash.Hash.Hash,
					},
				},
			}
			val := v.Value
			if !val.IsValid() {
				// DiffActionRemove
				val = Prim{
					Type:   PrimNullary,
					OpCode: D_NONE,
				}
			}
			kvpair := Prim{
				Type:   PrimBinary,
				OpCode: T_PAIR,
				Args:   []Prim{key, val},
			}
			if err := kvpair.EncodeBuffer(buf); err != nil {
				return nil, err
			}

		case DiffActionAlloc:
			// pair(key_type, value_type)
			kvpair := Prim{
				Type:   PrimBinary,
				OpCode: T_PAIR,
				Args:   []Prim{v.KeyType, v.ValueType},
			}
			if err := kvpair.EncodeBuffer(buf); err != nil {
				return nil, err
			}

		case DiffActionCopy:
			// pair(src, dest)
			kvpair := Prim{
				Type:   PrimBinary,
				OpCode: T_PAIR,
				Args: []Prim{
					Prim{
						Type: PrimInt,
						Int:  big.NewInt(v.SourceId),
					},
					Prim{
						Type: PrimInt,
						Int:  big.NewInt(v.DestId),
					},
				},
			}
			if err := kvpair.EncodeBuffer(buf); err != nil {
				return nil, err
			}
		}
	}
	return buf.Bytes(), nil
}

func (b *BigmapDiff) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	for buf.Len() > 0 {
		id := int32(binary.BigEndian.Uint32(buf.Next(4)))
		elem := BigmapDiffElem{
			Id:     int64(id),
			Action: DiffAction(buf.Next(1)[0]),
		}
		prim := Prim{}
		if err := prim.DecodeBuffer(buf); err != nil {
			return err
		}
		if prim.Type != PrimBinary && prim.OpCode != T_PAIR {
			return fmt.Errorf("micheline: unexpected big_map_diff keypair type %s opcode %s", prim.Type, prim.OpCode)
		}
		if l := len(prim.Args); l != 2 {
			return fmt.Errorf("micheline: unexpected big_map_diff keypair len %d", l)
		}

		switch elem.Action {
		case DiffActionUpdate, DiffActionRemove:
			// encoded as pair(pair(key,hash), val)
			if prim.Args[0].Args[1].Type != PrimBytes {
				return fmt.Errorf("micheline: unexpected big_map_diff keyhash type %s", prim.Args[0].Args[1].Type)
			}
			if err := elem.KeyHash.UnmarshalBinary(prim.Args[0].Args[1].Bytes); err != nil {
				return err
			}
			elem.Key = prim.Args[0].Args[0]
			elem.Value = prim.Args[1]
			if elem.Action == DiffActionRemove {
				elem.Value = Prim{}
			}
		case DiffActionAlloc:
			// encoded as pair(key_type, value_type)
			elem.KeyType = prim.Args[0]
			elem.ValueType = prim.Args[1]
			if !elem.KeyType.OpCode.IsValid() {
				return fmt.Errorf("micheline: invalid big_map_diff key type opcode %s [%d]",
					prim.Args[0].OpCode, prim.Args[0].OpCode)
			}
		case DiffActionCopy:
			// encoded as pair(src_id, dest_id)
			elem.SourceId = prim.Args[0].Int.Int64()
			elem.DestId = prim.Args[1].Int.Int64()
		}
		*b = append(*b, elem)
	}
	return nil
}
