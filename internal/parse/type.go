package parse

import (
	"blockwatch.cc/tzgo/contract/ast"
	m "blockwatch.cc/tzgo/micheline"
)

func (p *parser) buildTypeStructs(t *m.Typedef) (*ast.Struct, error) {
	// Unwrap optional
	if t.Optional {
		typ, err := p.buildTypeStructs(&m.Typedef{Name: t.Name, Type: t.Type, Args: t.Args})
		if err != nil {
			return nil, err
		}
		return typ, nil
	}
	// Builtin types
	if op, err := m.ParseOpCode(t.Type); err == nil {
		opstr := op.String()
		switch op {
		case m.T_NAT,
			m.T_INT,
			m.T_STRING,
			m.T_BOOL,
			m.T_BYTES,
			m.T_UNIT,
			m.T_TIMESTAMP,
			m.T_ADDRESS,
			m.T_MUTEZ,
			m.T_KEY,
			m.T_KEY_HASH,
			m.T_SIGNATURE,
			m.T_CHAIN_ID,
			m.T_OPERATION,
			m.T_CONTRACT:
			return &ast.Struct{
				Name:          t.Name,
				MichelineType: opstr,
			}, nil
		case m.T_BIG_MAP,
			m.T_MAP,
			m.T_LAMBDA:
			type1, err := p.buildTypeStructs(&t.Args[0])
			if err != nil {
				return nil, err
			}
			type2, err := p.buildTypeStructs(&t.Args[1])
			if err != nil {
				return nil, err
			}
			switch op {
			case m.T_BIG_MAP, m.T_MAP:
				return &ast.Struct{
					Name:          t.Name,
					MichelineType: opstr,
					Key:           type1,
					Value:         type2,
				}, nil
			case m.T_LAMBDA:
				return &ast.Struct{
					Name:          t.Name,
					MichelineType: opstr,
					ParamType:     type1,
					ReturnType:    type2,
				}, nil
			}
		}
	}
	// container type
	type1, err := p.buildTypeStructs(&t.Args[0])
	if err != nil {
		return nil, err
	}
	switch t.Type {
	case m.TypeStruct:
		return p.buildStruct(t)
	case m.TypeUnion:
		type2, err := p.buildTypeStructs(&t.Args[1])
		if err != nil {
			return nil, err
		}
		return &ast.Struct{
			MichelineType: "union",
			LeftType:      type1,
			RightType:     type2,
		}, nil
	case "list":
		return &ast.Struct{
			MichelineType: "list",
			Type:          type1,
		}, nil
	case "set":
		return &ast.Struct{
			MichelineType: "set",
			Type:          type1,
		}, nil
	}
	return nil, nil
}

func (p *parser) buildStruct(t *m.Typedef) (*ast.Struct, error) {
	fieldTypes := make([]*ast.Struct, 0, len(t.Args))
	path := make([][]int, 0, len(t.Args))
	for _, a := range t.Args {
		typ, err := p.buildTypeStructs(&a)
		if err != nil {
			return nil, err
		}
		name := a.Name
		if startsWithInt(name) {
			name = "field" + name
		}
		fieldTypes = append(fieldTypes, &ast.Struct{Name: name, Type: typ})
		path = append(path, a.Path)
	}
	st := &ast.Struct{
		MichelineType: "struct",
		Fields:        fieldTypes,
		Path:          path,
	}
	// Without annotation, structs gets a
	// @-prefixed auto generated name.
	// We want to remove it, so we get our auto-generated name.
	if len(t.Name) > 0 && t.Name[0] != '@' {
		st.Name = t.Name
	}
	cachedStruct, err := p.registerStruct(st)
	if err != nil {
		return nil, err
	}
	if cachedStruct != nil {
		return cachedStruct, nil
	}
	return st, nil
}

func (p *parser) registerStruct(newStruct *ast.Struct) (*ast.Struct, error) {
	if found, ok := p.cache.IsCached(newStruct); ok {
		return found, nil
	}
	p.structs = append(p.structs, newStruct)
	return nil, p.cache.CacheStruct(newStruct)
}
