// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
)

func Lookup(name string) *types.Sym {
	return types.LocalPkg.Lookup(name)
}

// loadsys loads the definitions for the low-level runtime functions,
// so that the compiler can generate calls to them,
// but does not make them visible to user code.
func LoadSys() {
	types.Block = 1

	inimport = true
	TypecheckAllowed = true

	typs := runtimeTypes()
	for _, d := range &runtimeDecls {
		sym := types.Pkgs.Runtime.Lookup(d.name)
		typ := typs[d.typ]
		switch d.tag {
		case funcTag:
			importfunc(types.Pkgs.Runtime, src.NoXPos, sym, typ)
		case varTag:
			importvar(types.Pkgs.Runtime, src.NoXPos, sym, typ)
		default:
			base.Fatalf("unhandled declaration tag %v", d.tag)
		}
	}

	TypecheckAllowed = false
	inimport = false
}

// sysfunc looks up Go function name in package runtime. This function
// must follow the internal calling convention.
func LookupRuntimeFunc(name string) *obj.LSym {
	s := types.Pkgs.Runtime.Lookup(name)
	s.SetFunc(true)
	return s.Linksym()
}

// sysvar looks up a variable (or assembly function) name in package
// runtime. If this is a function, it may have a special calling
// convention.
func LookupRuntimeVar(name string) *obj.LSym {
	return types.Pkgs.Runtime.Lookup(name).Linksym()
}
