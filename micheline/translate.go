// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

//go:build !trace
// +build !trace

package micheline

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func walkTree(m map[string]interface{}, label string, typ Type, stack *Stack, lvl int) error {
	// abort infinite type recursions
	if lvl > 99 {
		return fmt.Errorf("micheline: max nesting level reached")
	}

	// take next value from stack
	val := stack.Pop()

	// unfold unexpected pairs
	if !val.WasPacked && val.IsPair() && !typ.IsPair() {
		unfolded := val.UnfoldPair(typ)
		stack.Push(unfolded...)
		val = stack.Pop()
	}

	// detect type for unpacked values
	if val.WasPacked && (!val.IsScalar() || typ.OpCode == T_BYTES) {
		labels := typ.Anno
		typ = val.BuildType()
		typ.WasPacked = true
		typ.Anno = labels
	}

	// make sure value + type we're going to process actually match up
	// accept any kind of pairs/seq which will be unfolded again below
	if !typ.IsPair() && !val.IsSequence() && !val.matchOpCode(typ.OpCode) {
		return fmt.Errorf("micheline: type mismatch: type[%s]=%s value[%s/%d]=%s",
			typ.OpCode, typ.DumpLimit(512), val.Type, val.OpCode, val.DumpLimit(512))
	}

	// get the label from our type tree
	typeLabel := typ.Label()
	haveTypeLabel := len(typeLabel) > 0
	haveKeyLabel := label != EMPTY_LABEL && len(label) > 0
	if label == EMPTY_LABEL {
		if haveTypeLabel {
			// overwrite struct field label from type annotation
			label = typeLabel
		} else {
			// or use sequence number when type annotation is empty
			label = strconv.Itoa(len(m))
		}
	}

	// attach sub-records and array elements based on type code
	switch typ.OpCode {
	case T_SET:
		// set <comparable type>
		if len(typ.Args) == 0 {
			return fmt.Errorf("micheline: broken T_SET type prim")
		}
		arr := make([]interface{}, 0, len(val.Args))
		for _, v := range val.Args {
			if v.IsScalar() && !v.IsSequence() {
				// array of scalar types
				arr = append(arr, v.Value(typ.Args[0].OpCode))
			} else {
				// array of complex types
				mm := make(map[string]interface{})
				if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[0]}, NewStack(v), lvl+1); err != nil {
					return err
				}
				arr = append(arr, mm)
			}
		}
		m[label] = arr

	case T_LIST:
		// list <type>
		// dbg("List: lvl=%d", lvl)
		// dbg("List: typ=%s %s", typ.OpCode, label)
		// dbg("List: val=%s/%s", val.Type, val.OpCode)
		// dbg("-----------------------")

		// fix for pair(list, x) - we wrongly prevent a nested list from being unpacked,
		// this is to compensate for CanUnfold() and another case where list/pair unfold
		// does not work
		//
		// Conflicting cases
		// Jakartanet: oorcMSVaYBH3rcsDJ3n8EvpU4e8h38WFjJJfYUu2wXyDN4N7NMX
		// Mainnet: KT1K4jn23GonEmZot3pMGth7unnzZ6EaMVjY
		// Mainnet: ooxcyrwLVfC7kcJvLvYTGXKsAvdrotzKci95au8tBwdjhMMjFTU
		// Mainnet: ooQuRnwv2Bo1VVPMxmFvUZrDB7t34H3eCty2DAZW2Ps6LLyWoH6
		//
		// if len(typ.Args) > 0 && !typ.Args[0].IsList() && len(val.Args) > 1 && !val.LooksLikeContainer() && val.Args[0].IsSequence() { //&& !val.Args[0].IsConvertedComb() {
		if len(typ.Args) > 0 && !typ.Args[0].IsList() && len(val.Args) > 1 && !val.LooksLikeContainer() && val.Args[0].IsSequence() && !val.Args[0].IsConvertedComb() {
			stack.Push(val.Args...)
			val = stack.Pop()
		}

		arr := make([]interface{}, 0, len(val.Args))
		for i, v := range val.Args {
			// lists may contain different types, i.e. when unpack+detect is used
			valType := typ.Args[0]
			if len(typ.Args) > i {
				valType = typ.Args[i]
			}
			// unpack into map
			mm := make(map[string]interface{})
			if err := walkTree(mm, EMPTY_LABEL, Type{valType}, NewStack(v), lvl+1); err != nil {
				return err
			}
			// lift scalar nested list and simple element
			unwrapped := false
			if len(mm) == 1 {
				if mval, ok := mm["0"]; ok {
					if marr, ok := mval.([]interface{}); ok {
						arr = append(arr, marr)
					} else {
						arr = append(arr, mval)
					}
					unwrapped = true
				}
			}
			if !unwrapped {
				arr = append(arr, mm)
			}
		}
		m[label] = arr

	case T_LAMBDA:
		// LAMBDA <type> <type> { <instruction> ... }
		// dbg("LAMBDA: lvl=%d", lvl)
		// dbg("LAMBDA: typ=%s %s", typ.OpCode, label)
		// dbg("LAMBDA: val=%s/%s", val.Type, val.OpCode)
		// dbg("-----------------------")
		m[label] = val

	case T_MAP, T_BIG_MAP:
		// map <comparable type> <type>
		// big_map <comparable type> <type>
		// sequence of Elt (key/value) pairs
		// dbg("MAP: lvl=%d", lvl)
		// dbg("MAP: typ=%s %s", typ.OpCode, label)
		// dbg("MAP: val=%s/%s", val.Type, val.OpCode)
		// dbg("-----------------------")

		// render bigmap reference
		if typ.OpCode == T_BIG_MAP && (len(val.Args) == 0 || !val.Args[0].IsElt()) {
			switch val.Type {
			case PrimInt:
				// Babylon bigmaps contain a reference here
				m[label] = val.Int.Int64()
			case PrimSequence:
				if len(val.Args) == 0 {
					// pre-babylon there's only an empty sequence
					// FIXME: we could insert the bigmap id, but this is unknown at ths point
					m[label] = nil
				} else {
					if val.Args[0].IsSequence() {
						if err := walkTree(m, label, typ, NewStack(val.Args[0]), lvl); err != nil {
							return err
						}
					} else {
						m[label] = val.Args[0].Int.Int64()
					}
					stack.Push(val.Args[1:]...)
				}
			}
			return nil
		}

		if len(typ.Args) == 0 {
			return fmt.Errorf("micheline: broken T_BIG_MAP type prim")
		}

		switch val.Type {
		case PrimBinary: // single ELT
			keyType := Type{typ.Args[0]}
			valType := Type{typ.Args[1]}

			// build type info if prim was packed
			if val.Args[0].WasPacked {
				keyType = val.Args[0].BuildType()
			}

			// build type info if prim was packed
			if val.Args[1].WasPacked {
				valType = val.Args[1].BuildType()
			}

			// prepare key
			key, err := NewKey(keyType, val.Args[0])
			if err != nil {
				return err
			}
			mm := make(map[string]interface{})
			if err := walkTree(mm, key.String(), valType, NewStack(val.Args[1]), lvl+1); err != nil {
				return err
			}
			m[label] = mm

		case PrimSequence: // sequence of ELTs
			mm := make(map[string]interface{})
			for _, v := range val.Args {
				if v.OpCode != D_ELT {
					return fmt.Errorf("micheline: unexpected type %s [%s] for %s Elt item", v.Type, v.OpCode, typ.OpCode)
				}

				keyType := Type{typ.Args[0]}
				valType := Type{typ.Args[1]}

				// build type info if prim was packed
				if v.Args[0].WasPacked {
					keyType = v.Args[0].BuildType()
				}

				// build type info if prim was packed
				if v.Args[1].WasPacked {
					valType = v.Args[1].BuildType()
				}

				key, err := NewKey(keyType, v.Args[0])
				if err != nil {
					return err
				}
				if err := walkTree(mm, key.String(), valType, NewStack(v.Args[1]), lvl+1); err != nil {
					return err
				}
			}
			m[label] = mm

		default:
			buf, _ := json.Marshal(val)
			return fmt.Errorf("%*s> micheline: unexpected type %s [%s] for %s Elt sequence: %s",
				lvl, "", val.Type, val.OpCode, typ.OpCode, buf)
		}

	case T_PAIR:
		// pair <type> <type> or COMB
		mm := m
		if haveTypeLabel || haveKeyLabel {
			mm = make(map[string]interface{})
		}
		// dbg("PAIR: lvl=%d", lvl)

		// Try unfolding value (again) when type is T_PAIR,
		// reuse the existing stack and push unfolded values
		switch {
		case val.IsPair() && !typ.IsPair():
			// unfold regular pair
			// dbg("Unfold1: lvl=%d", lvl)
			// dbg("Unfold1: typ=%s", typ.Dump())
			// dbg("Unfold1: val=%s", val.Dump())
			// dbg("-----------------------")
			unfolded := val.UnfoldPair(typ)
			stack.Push(unfolded...)
		case val.CanUnfold(typ):
			// comb pair
			// dbg("Unfold2: lvl=%d", lvl)
			// dbg("Unfold2: typ=%s %s", typ.OpCode, label)
			// dbg("Unfold2: val=%s/%s", val.Type, val.OpCode)
			// dbg("-----------------------")
			stack.Push(val.Args...)
		default:
			// push value back on stack
			// dbg("Unfold3: lvl=%d", lvl)
			// dbg("Unfold3: typ=%s %s", typ.OpCode, label)
			// dbg("Unfold3: val=%s/%s", val.Type, val.OpCode)
			// dbg("-----------------------")
			stack.Push(val)
		}

		for _, t := range typ.Args {
			if err := walkTree(mm, EMPTY_LABEL, Type{t}, stack, lvl+1); err != nil {
				return err
			}
		}

		if haveTypeLabel || haveKeyLabel {
			m[label] = mm
		}

	case T_OPTION:
		// option <type>
		switch val.OpCode {
		case D_NONE:
			// add empty option values as null
			m[label] = nil
		case D_SOME:
			// skip if broken
			if len(val.Args) == 0 {
				return fmt.Errorf("micheline: broken T_OPTION value prim")
			}

			// detect nested type when missing, this can happen with option types
			// inside containers when the first element (used to detect the type
			// for all elements) has option None.
			if len(typ.Args) == 0 {
				typ = val.BuildType()
				// skip if broken
				if len(typ.Args) == 0 {
					return fmt.Errorf("micheline: broken T_OPTION type/value prim")
				}
			}

			// with annots (name) use it for scalar or complex render
			// when next level annot equals this option annot, skip this annot
			if val.IsScalar() || label == typ.Args[0].GetVarAnnoAny() {
				if anno := typ.Args[0].GetVarAnnoAny(); anno != "" {
					label = anno
				}
				if err := walkTree(m, label, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
			} else {
				mm := make(map[string]interface{})
				if anno := typ.Args[0].GetVarAnnoAny(); anno != "" {
					label = anno
				}
				if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
				m[label] = mm
			}
		default:
			return fmt.Errorf("micheline: unexpected T_OPTION code %s [%s]: %s", val.OpCode, val.OpCode, val.Dump())
		}

	case T_OR:
		// or <type> <type>
		// skip if broken
		if len(typ.Args) == 0 {
			return fmt.Errorf("micheline: broken T_OR type prim")
		}
		if len(val.Args) == 0 {
			return fmt.Errorf("micheline: broken T_OR value prim")
		}

		// use map to capture nested names
		mm := make(map[string]interface{})
		switch val.OpCode {
		case D_LEFT:
			if !(haveTypeLabel || haveKeyLabel) {
				mmm := make(map[string]interface{})
				if err := walkTree(mmm, EMPTY_LABEL, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
				// lift named content
				if len(mmm) == 1 {
					for n, v := range mmm {
						switch n {
						case "0":
							mm["@or_0"] = v
						default:
							mm[n] = v
						}
					}
				} else {
					mm["@or_0"] = mmm
				}
			} else {
				if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
			}
		case D_RIGHT:
			if !(haveTypeLabel || haveKeyLabel) {
				mmm := make(map[string]interface{})
				if err := walkTree(mmm, EMPTY_LABEL, Type{typ.Args[1]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
				// lift named content
				if len(mmm) == 1 {
					for n, v := range mmm {
						switch n {
						case "0":
							mm["@or_1"] = v
						default:
							mm[n] = v
						}
					}
				} else {
					mm["@or_1"] = mmm
				}
			} else {
				if err := walkTree(mm, EMPTY_LABEL, Type{typ.Args[1]}, NewStack(val.Args[0]), lvl+1); err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("micheline: unexpected T_OR branch with value %s", val.Dump())
		}

		// lift anon content
		if v, ok := mm["0"]; ok && len(mm) == 1 {
			m[label] = v
		} else {
			m[label] = mm
		}

	case T_TICKET:
		if len(typ.Args) == 0 {
			return fmt.Errorf("micheline: broken T_TICKET type prim")
		}
		// always Pair( ticketer:address, Pair( original_type, int ))
		stack.Push(val)
		if err := walkTree(m, label, TicketType(typ.Args[0]), stack, lvl+1); err != nil {
			return err
		}

	case T_SAPLING_STATE:
		if len(typ.Args) == 0 {
			return fmt.Errorf("micheline: broken T_SAPLING_STATE value prim")
		}
		mm := make(map[string]interface{})
		if err := walkTree(mm, "memo_size", Type{NewPrim(T_INT)}, NewStack(typ.Args[0]), lvl+1); err != nil {
			return err
		}
		if err := walkTree(mm, "content", val.BuildType(), NewStack(val), lvl+1); err != nil {
			return err
		}
		m[label] = mm

	default:
		// int
		// nat
		// string
		// bytes
		// mutez
		// bool
		// key_hash
		// timestamp
		// address
		// contract
		// key
		// unit
		// signature
		// operation
		// contract <type> (??)
		// chain_id
		// never
		// chest_key
		// chest
		// l2 address
		// append scalar or other complex value

		// dbg("Other: lvl=%d", lvl)
		// dbg("Other: typ=%s %s", typ.OpCode, label)
		// dbg("Other: val=%s/%s", val.Type, val.OpCode)
		// dbg("-----------------------")

		// comb-pair records might have slipped through in LooksLikeContainer()
		// so if we detect any unpacked comb part (i.e. sequence) we unpack it here
		if val.IsSequence() {
			stack.Push(val.Args...)
			val = stack.Pop()
		}

		// safety check: skip invalid values (only happens if type detect was wrong)
		if !val.IsValid() {
			break
		}

		if val.IsScalar() {
			m[label] = val.Value(typ.OpCode)
		} else {
			mm := make(map[string]interface{})
			if err := walkTree(mm, EMPTY_LABEL, typ, NewStack(val), lvl+1); err != nil {
				return err
			}
			m[label] = mm
		}
	}
	return nil
}
