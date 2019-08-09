// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sym

import "cmd/internal/dwarf"

// CompilationUnit is per-compilation unit (equivalently, per-package)
// debug-related data.
type CompilationUnit struct {
	Pkg       string        // our package
	Lib       *Library      // Our library
	Consts    *Symbol       // Package constants DIEs
	PCs       []dwarf.Range // PC ranges, relative to textp[0]
	DWInfo    *dwarf.DWDie  // CU root DIE
	FuncDIEs  []*Symbol     // Function DIE subtrees
	AbsFnDIEs []*Symbol     // Abstract function DIE subtrees
	RangeSyms []*Symbol     // symbols for debug_range
	TextP     []*Symbol     // text symbols in this CU
}
