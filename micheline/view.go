// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package micheline

import (
	"encoding/json"
)

type View struct {
	Name   string
	Param  Type
	Retval Type
	Code   Prim
	Prim   Prim
}

type Views map[string]View

func NewView(p Prim) View {
	vp := p.Clone()
	return View{
		Name:   vp.Args[0].String,
		Param:  Type{vp.Args[1]},
		Retval: Type{vp.Args[2]},
		Code:   vp.Args[3],
		Prim:   vp,
	}
}

func NewViewPtr(p Prim) *View {
	v := NewView(p)
	return &v
}

func (v View) IsValid() bool {
	return v.Name != "" && v.Param.IsValid() && v.Retval.IsValid()
}

func (v View) IsEqual(v2 View) bool {
	return IsEqualPrim(v.Param.Prim, v2.Param.Prim, false) && IsEqualPrim(v.Retval.Prim, v2.Retval.Prim, false)
}

func (v View) IsEqualWithAnno(v2 View) bool {
	return IsEqualPrim(v.Param.Prim, v2.Param.Prim, true) && IsEqualPrim(v.Retval.Prim, v2.Retval.Prim, true)
}

func (v View) IsEqualWithCode(v2 View) bool {
	return v.IsEqual(v2) && IsEqualPrim(v.Code, v2.Code, false)
}

// TODO if needed
// func (v *View) UnmarshalBinary(buf []byte) error {
//  return t.Prim.UnmarshalBinary(buf)
// }

func (v View) MarshalJSON() ([]byte, error) {
	if !v.IsValid() {
		return []byte("{}"), nil
	}
	type view struct {
		Type Typedef `json:"type"`
		Prim *Prim   `json:"prim,omitempty"`
		Code *Prim   `json:"code,omitempty"`
	}
	val := view{
		Type: v.Typedef(),
	}
	if v.Code.IsValid() {
		val.Code = &v.Code
	}
	if v.Prim.IsValid() {
		val.Prim = &v.Prim
	}
	return json.Marshal(val)
}

func (v View) Clone() View {
	return View{
		Name:   v.Name,
		Param:  v.Param.Clone(),
		Retval: v.Retval.Clone(),
		Code:   v.Code.Clone(),
	}
}

func (v View) Typedef() Typedef {
	return Typedef{
		Name: v.Name,
		Type: K_VIEW.String(),
		Args: []Typedef{
			buildTypedef(CONST_PARAM, v.Param.Prim),
			buildTypedef(CONST_RETURN, v.Retval.Prim),
		},
	}
}

func (v View) TypedefPtr(name string) *Typedef {
	td := v.Typedef()
	return &td
}
