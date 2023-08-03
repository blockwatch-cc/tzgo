package parse

import (
	"encoding/json"
	"fmt"
	"strconv"

	"blockwatch.cc/tzgo/contract/ast"
	"blockwatch.cc/tzgo/micheline"
	"github.com/pkg/errors"
)

func Parse(raw []byte, name string) (*ast.Contract, []*ast.Struct, error) {
	return newParser(raw).parse(name)
}

type parser struct {
	script   *micheline.Script
	raw      []byte
	contract *ast.Contract
	structs  []*ast.Struct
	cache    *Cache
}

func newParser(raw []byte) *parser {
	return &parser{
		script:   micheline.NewScript(),
		raw:      raw,
		contract: new(ast.Contract),
		cache:    NewCache(),
	}
}

func (p *parser) parse(name string) (*ast.Contract, []*ast.Struct, error) {
	err := json.Unmarshal(p.raw, &p.script)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal micheline code")
	}
	// Remove storage
	p.script.Storage = micheline.Prim{}
	p.raw, err = json.Marshal(p.script)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to re-marshall script")
	}
	p.contract.Name = name
	p.contract.Micheline = string(p.raw)
	if err = p.parseStorage(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse storage")
	}
	if err = p.parseEntrypoints(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse entrypoints")
	}
	return p.contract, p.nameStructs(), nil
}

func (p *parser) parseStorage() (err error) {
	p.contract.Storage, err = p.buildTypeStructs(p.script.StorageType().TypedefPtr("Storage"))
	if err != nil {
		return err
	}
	if p.contract.Storage.MichelineType == "struct" {
		p.contract.Storage.Flat = true
	}
	return nil
}

func (p *parser) nameStructs() []*ast.Struct {
	structs := []*ast.Struct{}
	for i, s := range p.structs {
		if s.Name == "" {
			s.Name = fmt.Sprintf("%s_record_%d", p.contract.Name, i)
		} else {
			newName := fmt.Sprintf("%s_%s", p.contract.Name, s.Name)
			if p.structNameExists(newName) {
				newName += strconv.Itoa(i)
			}
			s.Name = newName
		}
		structs = append(structs, s)
	}

	return structs
}

func (p *parser) structNameExists(name string) bool {
	for _, s := range p.structs {
		if s.Name == name {
			return true
		}
	}
	return false
}
