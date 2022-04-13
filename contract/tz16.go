// Copyright (c) 2020-2022 Blockwatch Data Inc.
// Author: alex@blockwatch.cc
package contract

import (
	"github.com/legonian/tzgo/micheline"
	// "github.com/legonian/tzgo/rpc"
	// "github.com/legonian/tzgo/tezos"
)

// Represents Tzip16 contract metadata
type Tz16 struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Version     string       `json:"version,omitempty"`
	License     *Tz16License `json:"license,omitempty"`
	Authors     []string     `json:"authors,omitempty"`
	Homepage    string       `json:"homepage,omitempty"`
	Source      *Tz16Source  `json:"source,omitempty"`
	Interfaces  []string     `json:"interfaces,omitempty"`
	Errors      []Tz16Error  `json:"errors,omitempty"`
	Views       []Tz16View   `json:"views,omitempty"`
}

type Tz16License struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

type Tz16Source struct {
	Tools    []string `json:"tools"`
	Location string   `json:"location,omitempty"`
}

type Tz16Error struct {
	Error     *micheline.Prim `json:"error,omitempty"`
	Expansion *micheline.Prim `json:"expansion,omitempty"`
	Languages []string        `json:"languages,omitempty"`
	View      string          `json:"view,omitempty"`
}

type Tz16View struct {
	Name            string         `json:"name"`
	Description     string         `json:"description,omitempty"`
	Pure            bool           `json:"pure,omitempty"`
	Implementations []Tz16ViewImpl `json:"implementations,omitempty"`
}

type Tz16ViewImpl struct {
	Storage *Tz16StorageView `json:"michelsonStorageView,omitempty"`
	Rest    *Tz16RestView    `json:"restApiQuery,omitempty"`
}

type Tz16StorageView struct {
	ParamType   *micheline.Prim      `json:"parameter,omitempty"`
	ReturnType  micheline.Prim       `json:"returnType"`
	Code        micheline.Prim       `json:"code"`
	Annotations []Tz16CodeAnnotation `json:"annotations,omitempty"`
	Version     string               `json:"version,omitempty"`
}

type Tz16CodeAnnotation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Tz16RestView struct {
	SpecUri string `json:"specificationUri"`
	BaseUri string `json:"baseUri"`
	Path    string `json:"path"`
	Method  string `json:"method"`
}

func (t Tz16) Validate() []error {
	// TODO: json schema validator
	return nil
}

func (t Tz16) HasView(name string) bool {
	for _, v := range t.Views {
		if v.Name == name {
			return true
		}
	}
	return false
}
