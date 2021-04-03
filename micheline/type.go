// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Type struct {
	Prim
}

type KeyValueType struct {
	KeyType   Type `json:"key_type"`
	ValueType Type `json:"value_type"`
}

func NewType() Type {
	return Type{Prim{}}
}

func (t *Type) UnmarshalBinary(buf []byte) error {
	return t.Prim.UnmarshalBinary(buf)
}

func (t Type) Clone() Type {
	return Type{t.Prim.Clone()}
}

func (t Type) Label() string {
	return t.GetVarAnnoAny()
}

func (t Type) HasLabel() bool {
	return t.HasAnno()
}

func (t Type) IsEqual(t2 Type) bool {
	return IsEqualPrim(t.Prim, t2.Prim, false)
}

func (t Type) IsEqualWithAnno(t2 Type) bool {
	return IsEqualPrim(t.Prim, t2.Prim, true)
}

func (t Type) Left() Type {
	if len(t.Args) > 0 {
		return Type{t.Args[0]}
	}
	return Type{}
}

func (t Type) Right() Type {
	if len(t.Args) > 1 {
		return Type{t.Args[1]}
	}
	return Type{}
}

// FIXME: rework type listing
func (e Type) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, 1024)
	// always normalize comb pairs
	typ := e
	// if typ.IsPair() {
	// 	typ = Type{tcomb(typ.FlattenComb(typ)...)}
	// }
	if err := walkTypeTree(m, "", typ, false); err != nil {
		return nil, err
	}
	// lift embedded scalars unless they are named or container types
	// FIXME: is this still necessary
	if len(m) == 1 {
		for n, v := range m {
			fields := strings.Split(n, "@")
			oc, err := ParseOpCode(fields[len(fields)-1])
			if err == nil || strings.HasPrefix(n, "0") {
				switch oc {
				case T_LIST, T_MAP, T_SET, T_LAMBDA, T_BIG_MAP, T_OR, T_OPTION, T_PAIR:
				default:
					return json.Marshal(v)
				}
			}
		}
	}
	return json.Marshal(m)
}

// FIXME: do we need ArgType in addition to Type (just for not lifting scalars)
type ArgType struct {
	Prim
}

func (e ArgType) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, 1024)
	// always normalize comb pairs
	typ := Type{e.Prim}
	// if typ.IsPair() {
	// 	typ = Type{tcomb(typ.FlattenComb(typ)...)}
	// }
	if err := walkTypeTree(m, "", typ, true); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func (t *ArgType) StripAnno(name string) {
	for i := 0; i < len(t.Anno); i++ {
		if t.Anno[i][1:] == name {
			t.Anno = append(t.Anno[:i], t.Anno[i+1:]...)
			i--
		}
	}
}

func walkTypeTree(m map[string]interface{}, path string, typ Type, withSequenceNum bool) error {
	// use annot as name when exists; even though this is a type tree we need to produce
	// variable names that match the storage spec so our frontends can match between
	// type desc and values
	haveLabel := typ.HasLabel()
	if len(path) == 0 {
		if haveLabel {
			if withSequenceNum {
				path = strconv.Itoa(len(m)) + "@" + typ.Label()
			} else {
				path = typ.Label()
			}
		} else {
			path = strconv.Itoa(len(m))
		}
	}

	switch typ.OpCode {
	case T_LIST, T_SET:
		// list <type>
		// set <comparable type>
		if haveLabel {
			path = path + "@" + typ.OpCode.String()
		} else {
			path = strconv.Itoa(len(m)) + "@" + typ.OpCode.String()
		}

		// put list elements into own map, never collapse to avoid dropping list name/type
		mm := make(map[string]interface{})
		if err := walkTypeTree(mm, "", Type{typ.Args[0]}, withSequenceNum); err != nil {
			return err
		}
		m[path] = mm

	case T_MAP, T_BIG_MAP:
		// map <comparable type> <type>
		// big_map <comparable type> <type>
		mm := make(map[string]interface{})

		// key and value type
		for i, v := range typ.Args {
			mmm := make(map[string]interface{})
			if err := walkTypeTree(mmm, "", Type{v}, withSequenceNum); err != nil {
				return err
			}
			// lift scalar
			var lifted bool
			if len(mmm) == 1 {
				for n, v := range mmm {
					fields := strings.Split(n, "@")
					liftedName := strconv.Itoa(i)
					lifted = true
					if fields[0] == "0" {
						if len(fields) > 1 {
							liftedName += "@" + strings.Join(fields[1:], "@")
						}
					} else {
						liftedName += "@" + strings.Join(fields, "@")
					}
					mm[liftedName] = v
				}
			}
			if !lifted {
				mm[strconv.Itoa(i)] = mmm
			}
		}

		if haveLabel {
			path = path + "@" + typ.OpCode.String()
		} else {
			path = strconv.Itoa(len(m)) + "@" + typ.OpCode.String()
		}

		m[path] = mm

	case T_LAMBDA:
		// LAMBDA <type> <type> { <instruction> ... }
		mm := make(map[string]interface{})
		for _, v := range typ.Args {
			if err := walkTypeTree(mm, "", Type{v}, withSequenceNum); err != nil {
				return err
			}
		}

		if haveLabel {
			path = path + "@" + typ.OpCode.String()
		} else {
			path = strconv.Itoa(len(m)) + "@" + typ.OpCode.String()
		}

		m[path] = mm

	case T_PAIR:
		// pair <type> <type>
		if !haveLabel {
			// collapse pairs when annots are empty
			for _, v := range typ.Args {
				if err := walkTypeTree(m, "", Type{v}, withSequenceNum); err != nil {
					return err
				}
			}
		} else {
			// when annots are NOT empty, create a new sub-map unless value is scalar
			mm := make(map[string]interface{})
			// 	log.Debugf("marshal sub pair map %p %s into map %p", mm, path, m)
			for _, v := range typ.Args {
				if err := walkTypeTree(mm, "", Type{v}, withSequenceNum); err != nil {
					return err
				}
			}
			m[path] = mm
		}

	case T_OR, T_OPTION:
		// option <type>
		// or <type> <type>
		if len(typ.Args) == 0 {
			p := path
			if haveLabel {
				p = p + "@" + typ.OpCode.String()
			}
			for _, v := range typ.Args {
				if err := walkTypeTree(m, p, Type{v}, withSequenceNum); err != nil {
					return err
				}
			}
		} else {
			mm := m
			p := path
			if haveLabel {
				mm = make(map[string]interface{})
				p = p + "@" + typ.OpCode.String()
			}
			for _, v := range typ.Args {
				if err := walkTypeTree(mm, "", Type{v}, withSequenceNum); err != nil {
					return err
				}
			}
			if haveLabel {
				m[p] = mm
			}
		}
	case T_TICKET:
		return walkTypeTree(m, path, TicketType(typ.Args[0]), withSequenceNum)

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
		// contract <type> (??) is this an entrypoint?
		// chain_id
		m[path] = typ.OpCode.String()
	}
	return nil
}
