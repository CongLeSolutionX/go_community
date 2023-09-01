// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sym

import "cmd/internal/dwarf"

// An ID encapsulates a global symbol index, used to identify a specific
// Go symbol. The 0-valued ID is corresponds to an invalid symbol.
type ID uint32

// A CompilationUnit represents a set of source files that are compiled
// together. Since all Go sources in a Go package are compiled together,
// there's one CompilationUnit per package that represents all Go sources in
// that package, plus one for each assembly file.
//
// Equivalently, there's one CompilationUnit per object file in each Library
// loaded by the linker.
//
// These are used for both DWARF and pclntab generation.
type CompilationUnit struct {
	Lib       *Library      // Our library
	PclnIndex int           // Index of this CU in pclntab
	PCs       []dwarf.Range // PC ranges, relative to Textp[0]
	DWInfo    *dwarf.DWDie  // CU root DIE
	FileTable []string      // The file table used in this compilation unit.

	Consts    ID   // Package constants DIEs
	FuncDIEs  []ID // Function DIE subtrees
	VarDIEs   []ID // Global variable DIEs
	AbsFnDIEs []ID // Abstract function DIE subtrees
	RangeSyms []ID // Symbols for debug_range
	Textp     []ID // Text symbols in this CU
}
