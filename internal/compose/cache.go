// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var (
	cacheFullPath string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("cannot read $HOME: %v", err))
	}
	cacheFullPath = filepath.Join(home, ".cache/tzcompose")
}

type PipelineCache struct {
	hash uint64
	last int
}

func NewCache() *PipelineCache {
	return &PipelineCache{
		hash: 0,
		last: 0,
	}
}

func (c *PipelineCache) Update(idx int) error {
	c.last = idx
	return writeFile(c.hash, []byte(strconv.Itoa(c.last)))
}

func (c PipelineCache) Get() int {
	return c.last
}

func (c *PipelineCache) Load(hash uint64, reset bool) error {
	c.hash = hash
	if reset {
		return c.Update(0)
	}
	buf, err := readFile(hash)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		c.last = -1

	} else {
		num, err := strconv.Atoi(string(buf))
		if err != nil {
			return fmt.Errorf("parsing cache file %016x: %v", hash, err)
		}
		c.last = num
	}
	return nil
}

func writeFile(hash uint64, buf []byte) error {
	_, err := os.Stat(cacheFullPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		err = os.MkdirAll(cacheFullPath, 0755)
		if err != nil {
			return err
		}
	}
	filename := filepath.Join(cacheFullPath, strconv.FormatUint(hash, 16))
	return os.WriteFile(filename, buf, 0644)
}

func readFile(hash uint64) ([]byte, error) {
	filename := filepath.Join(cacheFullPath, strconv.FormatUint(hash, 16))
	return os.ReadFile(filename)
}
