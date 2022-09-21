// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements the (temporary) plumbing to get importing to work.

package types2_test

import (
	gcimporter "cmd/compile/internal/importer"
	"cmd/compile/internal/types2"
	"internal/buildinternal"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func defaultImporter() types2.Importer {
	stdlibPkgs, err := buildinternal.StdlibPkgfileMap()
	if err != nil {
		panic(err)
	}

	lookup := func(path string) (io.ReadCloser, error) {
		p, ok := stdlibPkgs[path]
		if !ok {
			p, ok = stdlibPkgs["vendor/"+path]
		}
		return os.Open(p)
	}

	lookupVendorCmd := func(path string) (io.ReadCloser, error) {
		p, ok := stdlibPkgs[path]
		if !ok {
			p, ok = stdlibPkgs["cmd/vendor/"+path]
		}
		return os.Open(p)
	}

	return &gcimports{
		packages:        make(map[string]*types2.Package),
		lookupVendorStd: lookup,
		lookupVendorCmd: lookupVendorCmd,
	}
}

type gcimports struct {
	packages        map[string]*types2.Package
	lookupVendorStd func(path string) (io.ReadCloser, error)
	lookupVendorCmd func(path string) (io.ReadCloser, error)
}

func (m *gcimports) Import(path string) (*types2.Package, error) {
	return m.ImportFrom(path, "" /* no vendoring */, 0)
}

func (m *gcimports) ImportFrom(path, srcDir string, mode types2.ImportMode) (*types2.Package, error) {
	if mode != 0 {
		panic("mode must be 0")
	}
	lookup := m.lookupVendorStd
	if strings.Contains(srcDir, filepath.Join("src", "cmd")) {
		lookup = m.lookupVendorCmd
	}
	return gcimporter.Import(m.packages, path, srcDir, lookup)
}
