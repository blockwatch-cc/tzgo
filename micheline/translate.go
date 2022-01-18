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
        m[label] = val

    case T_MAP, T_BIG_MAP:
        // map <comparable type> <type>
        // big_map <comparable type> <type>
        // sequence of Elt (key/value) pairs

        // render bigmap reference
        if typ.OpCode == T_BIG_MAP && (len(val.Args) == 0 || !val.Args[0].IsElt()) {
            switch val.Type {
            case PrimInt:
                // Babylon bigmaps contain a reference here
                m[label] = val.Value(T_INT)
            case PrimSequence:
                if len(val.Args) == 0 {
                    // pre-babylon there's only an empty sequence
                    // FIXME: we could insert the bigmap id, but this is unknown at ths point
                    m[label] = nil
                } else {
                    m[label] = val.Args[0].Value(T_INT)
                    stack.Push(val.Args[1:]...)
                }
            }
            return nil
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

        // Try unfolding value (again) when type is T_PAIR,
        // reuse the existing stack and push unfolded values
        if val.IsPair() && !typ.IsPair() {
            // unfold regular pair
            unfolded := val.UnfoldPair(typ)
            stack.Push(unfolded...)
        } else if val.CanUnfold(typ) {
            // comb pair
            stack.Push(val.Args...)
        } else {
            // push value back on stack
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
            // with annots (name) use it for scalar or complex render
            // when next level annot equals this option annot, skip this annot
            if val.IsScalar() || label == typ.Args[0].GetVarAnnoAny() {
                if err := walkTree(m, label, Type{typ.Args[0]}, NewStack(val.Args[0]), lvl+1); err != nil {
                    return err
                }
            } else {
                mm := make(map[string]interface{})
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
            return fmt.Errorf("micheline: unexpected T_OR branch with value opcode %s", val.OpCode)
        }

        // lift anon content
        if v, ok := mm["0"]; ok && len(mm) == 1 {
            m[label] = v
        } else {
            m[label] = mm
        }

    case T_TICKET:
        // always Pair( ticketer:address, Pair( original_type, int ))
        stack.Push(val)
        if err := walkTree(m, label, TicketType(typ.Args[0]), stack, lvl+1); err != nil {
            return err
        }

    case T_SAPLING_STATE:
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
        // key
        // unit
        // signature
        // operation
        // contract <type> (??)
        // chain_id
        // never
        // chest_key
        // chest
        // append scalar or other complex value

        // comb-pair records might have slipped through in LooksLikeContainer()
        // so if we detect any unpacked comb part (i.e. sequence) we unpack it here
        if val.IsSequence() {
            stack.Push(val.Args...)
            val = stack.Pop()
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
