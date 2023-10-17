// Copyright (c) 2023 Blockwatch Data Inc.
// Author: alex@blockwatch.cc, abdul@blockwatch.cc

package compose

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type RunMode byte

const (
	RunModeValidate RunMode = iota
	RunModeSimulate
	RunModeExecute
)

type Spec struct {
	Version  string `yaml:"version"`
	Filename string `yaml:"-"`
}

func Run(ctx Context, fpath string, mode RunMode) error {
	ctx.Log.Debug("Initializing context...")
	ctx.WithMode(mode)
	if err := ctx.Init(); err != nil {
		return err
	}
	visitRecursive := strings.HasSuffix(fpath, "...")
	fpath = strings.TrimSuffix(fpath, "...")
	finfo, err := os.Stat(fpath)
	if err != nil {
		return err
	}
	if finfo.IsDir() {
		isFirst := true
		return filepath.WalkDir(fpath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if visitRecursive || isFirst {
					isFirst = false
					return nil
				} else {
					return filepath.SkipDir
				}
			}
			if _, ok := yamlExts[filepath.Ext(path)]; !ok {
				return nil
			}
			return handleFile(ctx, path, mode)
		})
	} else {
		return handleFile(ctx, fpath, mode)
	}
}

func handleFile(ctx Context, fpath string, mode RunMode) error {
	ctx.Log.Debugf("Reading file %s", fpath)
	s, err := ParseFile[Spec](fpath)
	if err != nil {
		return err
	}
	s.Filename = fpath

	ver := s.Version
	if ver == "" {
		ver = lastVersion
	}
	if !HasVersion(ver) {
		return fmt.Errorf("%s: %v", fpath, ErrInvalidVersion)
	}
	eng := New(ver)
	switch mode {
	case RunModeValidate:
		ctx.Log.Debugf("Validating file %s", s.Filename)
		err = eng.Validate(ctx, s.Filename)
	case RunModeSimulate:
		ctx.Log.Debugf("Simulating file %s", s.Filename)
		err = eng.Run(ctx, s.Filename)
	case RunModeExecute:
		ctx.Log.Debugf("Running file %s", s.Filename)
		err = eng.Run(ctx, s.Filename)
	}
	if err != nil {
		return fmt.Errorf("%s: %v", fpath, err)
	} else {
		ctx.Log.Infof("%s OK", s.Filename)
	}
	return nil
}
