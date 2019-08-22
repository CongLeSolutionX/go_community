// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sym

import "cmd/internal/dwarf"

// CompilationUnit is an abstraction used by DWARF to represent a chunk of
// debug-related data. We create a CompilationUnit per Object file in a
// library (so, one for all the Go code, one for each assembly file, etc.).
type CompilationUnit struct {
	Pkg            string        // our package
	Lib            *Library      // Our library
	Consts         *Symbol       // Package constants DIEs
	PCs            []dwarf.Range // PC ranges, relative to textp[0]
	DWInfo         *dwarf.DWDie  // CU root DIE
	FuncDIEs       []*Symbol     // Function DIE subtrees
	AbsFnDIEs      []*Symbol     // Abstract function DIE subtrees
	RangeSyms      []*Symbol     // symbols for debug_range
	Textp          []*Symbol     // text symbols in this CU
	DWARFFileTable []string      // The file table used to generate the .debug_lines
}
