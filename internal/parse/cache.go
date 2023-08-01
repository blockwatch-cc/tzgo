package parse

import (
	"encoding/json"
	"hash/fnv"

	"blockwatch.cc/tzgo/contract/ast"
)

type Cache struct {
	data map[uint64]*ast.Struct
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[uint64]*ast.Struct),
	}
}

func (c *Cache) CacheStruct(newStruct *ast.Struct) error {
	id, err := c.hash(newStruct)
	if err != nil {
		return err
	}
	if _, ok := c.data[id]; !ok {
		c.data[id] = newStruct
	}
	return nil
}

func (c *Cache) IsCached(newStruct *ast.Struct) (*ast.Struct, bool) {
	id, err := c.hash(newStruct)
	if err != nil {
		return nil, false
	}
	s, ok := c.data[id]
	return s, ok
}

func (c *Cache) hash(newStruct *ast.Struct) (uint64, error) {
	h := fnv.New64()
	b, err := json.Marshal(newStruct)
	if err != nil {
		return 0, err
	}
	h.Write(b)
	return h.Sum64(), nil
}
