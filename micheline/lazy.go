// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"bytes"
	"encoding/json"
	"fmt"

	"blockwatch.cc/tzgo/tezos"
)

type LazyKind string

const (
	LazyKindInvalid LazyKind = ""
	LazyKindBigmap  LazyKind = "big_map"
	LazyKindSapling LazyKind = "sapling_state"
)

func (k LazyKind) String() string {
	switch k {
	case LazyKindBigmap:
		return "big_map"
	case LazyKindSapling:
		return "sapling_state"
	}
	return ""
}

func ParseLazyKind(data string) LazyKind {
	switch data {
	case "big_map":
		return LazyKindBigmap
	case "sapling_state":
		return LazyKindSapling
	default:
		return LazyKindInvalid
	}
}

func (k LazyKind) IsValid() bool {
	return k != LazyKindInvalid
}

func (k LazyKind) MarshalText() ([]byte, error) {
	return []byte(k.String()), nil
}

func (k *LazyKind) UnmarshalText(data []byte) error {
	lk := ParseLazyKind(string(data))
	if !lk.IsValid() {
		return fmt.Errorf("micheline: invalid lazy kind %q", string(data))
	}
	*k = lk
	return nil
}

type LazyEvent interface {
	Kind() LazyKind
	Id() int64
}

type GenericEvent struct {
	EventKind LazyKind `json:"kind"`
	ObjectId  int64    `json:"id,string"`
}

func (d *GenericEvent) Kind() LazyKind {
	return d.EventKind
}

func (d *GenericEvent) Id() int64 {
	return d.ObjectId
}

type LazyEvents []LazyEvent

func (d LazyEvents) BigmapEvents() BigmapEvents {
	if len(d) == 0 {
		return nil
	}

	// count number of updates before allocating a slice
	var count int
	for _, v := range d {
		if v.Kind() != LazyKindBigmap {
			continue
		}
		ev := v.(*LazyBigmapEvent)
		switch ev.Diff.Action {
		case DiffActionAlloc:
			count += 1 + len(ev.Diff.Updates)
		case DiffActionUpdate:
			count += len(ev.Diff.Updates)
		default:
			count++
		}
	}
	if count == 0 {
		return nil
	}

	// translate updates
	events := make(BigmapEvents, 0, count)
	for _, v := range d {
		if v.Kind() != LazyKindBigmap {
			continue
		}
		ev := v.(*LazyBigmapEvent)
		switch ev.Diff.Action {
		case DiffActionUpdate:
			// upsert or remove
			for _, vv := range ev.Diff.Updates {
				event := BigmapEvent{
					Action:  DiffActionUpdate,
					Id:      ev.BigmapId,
					KeyHash: vv.KeyHash,
					Key:     vv.Key,
					Value:   vv.Value,
				}
				if !vv.Value.IsValid() {
					event.Action = DiffActionRemove
				}
				events = append(events, event)
			}
		case DiffActionRemove:
			// key remove or bigmap remove
			for _, vv := range ev.Diff.Updates {
				event := BigmapEvent{
					Action:  DiffActionRemove,
					Id:      ev.BigmapId,
					KeyHash: vv.KeyHash,
					Key:     vv.Key,
				}
				if !vv.Key.IsValid() && !vv.KeyHash.IsValid() {
					event.Key = Prim{
						Type:   PrimNullary,
						OpCode: I_EMPTY_BIG_MAP,
					}
				}
				events = append(events, event)
			}

		case DiffActionAlloc:
			// add an alloc event
			events = append(events, BigmapEvent{
				Action:    DiffActionAlloc,
				Id:        ev.BigmapId,
				KeyType:   ev.Diff.KeyType,
				ValueType: ev.Diff.ValueType,
			})
			// may contain upserts
			for _, vv := range ev.Diff.Updates {
				events = append(events, BigmapEvent{
					Action:  DiffActionUpdate,
					Id:      ev.BigmapId,
					KeyHash: vv.KeyHash,
					Key:     vv.Key,
					Value:   vv.Value,
				})
			}
		case DiffActionCopy:
			events = append(events, BigmapEvent{
				Action:   DiffActionCopy,
				SourceId: ev.Diff.SourceId,
				DestId:   ev.BigmapId,
			})
		}
	}
	return events
}

func (d *LazyEvents) UnmarshalJSON(data []byte) error {
	if len(data) <= 2 {
		return nil
	}

	if data[0] != '[' {
		return fmt.Errorf("micheline: expected lazy event array")
	}

	// fmt.Printf("Decoding ops: %s\n", string(data))
	dec := json.NewDecoder(bytes.NewReader(data))

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return fmt.Errorf("micheline: %v", err)
	}

	for dec.More() {
		// peek into `{"kind":"...",` field
		start := int(dec.InputOffset()) + 9
		// after first JSON object, decoder pos is at `,`
		if data[start] == '"' {
			start += 1
		}
		end := start + bytes.IndexByte(data[start:], '"')
		kind := ParseLazyKind(string(data[start:end]))
		var ev LazyEvent
		switch kind {
		case LazyKindBigmap:
			ev = &LazyBigmapEvent{}
		case LazyKindSapling:
			ev = &LazySaplingEvent{}
		default:
			log.Warnf("micheline: unsupported lazy diff kind %q", string(data[start:end]))
			ev = &GenericEvent{}
		}

		if err := dec.Decode(ev); err != nil {
			return fmt.Errorf("micheline: lazy kind %s: %v", kind, err)
		}
		(*d) = append(*d, ev)
	}

	return nil
}

type LazyBigmapEvent struct {
	BigmapId int64 `json:"id,string"`
	Diff     struct {
		Action  DiffAction `json:"action"`
		Updates []struct {
			KeyHash tezos.ExprHash `json:"key_hash"` // update/remove
			Key     Prim           `json:"key"`      // update/remove
			Value   Prim           `json:"value"`    // update
		} `json:"updates"` // update
		KeyType   Prim  `json:"key_type"`      // alloc
		ValueType Prim  `json:"value_type"`    // alloc
		SourceId  int64 `json:"source,string"` // copy
	} `json:"diff"`
}

func (d *LazyBigmapEvent) Kind() LazyKind {
	return LazyKindBigmap
}

func (d *LazyBigmapEvent) Id() int64 {
	return d.BigmapId
}

type LazySaplingEvent struct {
	PoolId int64           `json:"id,string"`
	Diff   SaplingDiffElem `json:"diff"`
}

func (d *LazySaplingEvent) Kind() LazyKind {
	return LazyKindSapling
}

func (d *LazySaplingEvent) Id() int64 {
	return d.PoolId
}
