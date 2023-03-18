// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"bytes"
	"encoding/binary"
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

type BigmapEvents []BigmapEvent

func (l BigmapEvents) Filter(id int64) BigmapEvents {
	var res BigmapEvents
	for _, v := range l {
		if v.Id == id {
			res = append(res, v)
		}
	}
	return res
}

type BigmapEvent struct {
	Action    DiffAction     `json:"action"`
	Id        int64          `json:"big_map,string"`
	KeyHash   tezos.ExprHash `json:"key_hash"`                   // update/remove
	Key       Prim           `json:"key"`                        // update/remove
	Value     Prim           `json:"value"`                      // update
	KeyType   Prim           `json:"key_type"`                   // alloc
	ValueType Prim           `json:"value_type"`                 // alloc
	SourceId  int64          `json:"source_big_map,string"`      // copy
	DestId    int64          `json:"destination_big_map,string"` // copy
}

func (e BigmapEvent) Encoding() PrimType {
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

func (e BigmapEvent) GetKey(typ Type) Key {
	k, err := NewKey(typ, e.Key)
	if err != nil {
		log.Error(err)
	}
	return k
}

func (e BigmapEvent) GetKeyPtr(typ Type) *Key {
	k, err := NewKey(typ, e.Key)
	if err != nil {
		log.Error(err)
	}
	return &k
}

func (e *BigmapEvent) UnmarshalJSON(data []byte) error {
	type alias BigmapEvent
	err := json.Unmarshal(data, (*alias)(e))
	if err != nil {
		return err
	}
	// translate update with empty value to remove
	if e.Action == DiffActionUpdate && !e.Value.IsValid() {
		e.Action = DiffActionRemove
	}
	return nil
}

func (e BigmapEvent) MarshalJSON() ([]byte, error) {
	var res interface{}
	switch e.Action {
	case DiffActionUpdate, DiffActionRemove:
		// set key, keyhash, value
		val := struct {
			Id      int64           `json:"big_map,string"`
			Action  DiffAction      `json:"action"`
			Key     *Prim           `json:"key,omitempty"`
			KeyHash *tezos.ExprHash `json:"key_hash,omitempty"`
			Value   *Prim           `json:"value,omitempty"`
		}{
			Id:     e.Id,
			Action: e.Action,
		}
		if e.KeyHash.IsValid() {
			val.KeyHash = &e.KeyHash
		}
		if e.Key.IsValid() {
			val.Key = &e.Key
		}
		if e.Value.IsValid() {
			val.Value = &e.Value
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

func (b BigmapEvents) MarshalBinary() ([]byte, error) {
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
					{
						Type:  PrimBytes,
						Bytes: v.KeyHash[:],
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
					{
						Type: PrimInt,
						Int:  big.NewInt(v.SourceId),
					},
					{
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

func (b *BigmapEvents) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	for buf.Len() > 0 {
		id := int32(binary.BigEndian.Uint32(buf.Next(4)))
		elem := BigmapEvent{
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
