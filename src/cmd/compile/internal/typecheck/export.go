// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typecheck

import (
	"fmt"
	"go/constant"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

// importalias declares symbol s as an imported type alias with type t.
// ipkg is the package being imported
func importalias(ipkg *types.Pkg, pos src.XPos, s *types.Sym, t *types.Type) {
	n := importobj(ipkg, pos, s, ir.OTYPE, ir.PEXTERN, t)
	if n == nil {
		return
	}

	if base.Flag.E != 0 {
		fmt.Printf("import type %v = %L\n", s, t)
	}
}

// importconst declares symbol s as an imported constant with type t and value val.
// ipkg is the package being imported
func importconst(ipkg *types.Pkg, pos src.XPos, s *types.Sym, t *types.Type, val constant.Value) {
	n := importobj(ipkg, pos, s, ir.OLITERAL, ir.PEXTERN, t)
	if n == nil { // TODO: Check that value matches.
		return
	}

	n.SetVal(val)

	if base.Flag.E != 0 {
		fmt.Printf("import const %v %L = %v\n", s, t, val)
	}
}

// importfunc declares symbol s as an imported function with type t.
// ipkg is the package being imported
func importfunc(ipkg *types.Pkg, pos src.XPos, s *types.Sym, t *types.Type) {
	n := importobj(ipkg, pos, s, ir.ONAME, ir.PFUNC, t)
	if n == nil {
		return
	}
	name := n.(*ir.Name)

	fn := ir.NewFunc(pos)
	fn.SetType(t)
	name.SetFunc(fn)
	fn.Nname = name

	if base.Flag.E != 0 {
		fmt.Printf("import func %v%S\n", s, t)
	}
}

// importobj declares symbol s as an imported object representable by op.
// ipkg is the package being imported
func importobj(ipkg *types.Pkg, pos src.XPos, s *types.Sym, op ir.Op, ctxt ir.Class, t *types.Type) ir.Node {
	n := importsym(ipkg, s, op)
	if n.Op() != ir.ONONAME {
		if n.Op() == op && (op == ir.ONAME && n.Class_ != ctxt || !types.Identical(n.Type(), t)) {
			Redeclared(base.Pos, s, fmt.Sprintf("during import %q", ipkg.Path))
		}
		return nil
	}

	n.SetOp(op)
	n.SetPos(pos)
	n.Class_ = ctxt
	if ctxt == ir.PFUNC {
		n.Sym().SetFunc(true)
	}
	n.SetType(t)
	return n
}

func importsym(ipkg *types.Pkg, s *types.Sym, op ir.Op) *ir.Name {
	n := ir.AsNode(s.PkgDef())
	if n == nil {
		// iimport should have created a stub ONONAME
		// declaration for all imported symbols. The exception
		// is declarations for Runtimepkg, which are populated
		// by loadsys instead.
		if s.Pkg != types.Pkgs.Runtime {
			base.Fatalf("missing ONONAME for %v\n", s)
		}

		n = ir.NewDeclNameAt(src.NoXPos, s)
		s.SetPkgDef(n)
		s.Importdef = ipkg
	}
	if n.Op() != ir.ONONAME && n.Op() != op {
		Redeclared(base.Pos, s, fmt.Sprintf("during import %q", ipkg.Path))
	}
	return n.(*ir.Name)
}

// importtype returns the named type declared by symbol s.
// If no such type has been declared yet, a forward declaration is returned.
// ipkg is the package being imported
func importtype(ipkg *types.Pkg, pos src.XPos, s *types.Sym) *types.Type {
	n := importsym(ipkg, s, ir.OTYPE)
	if n.Op() != ir.OTYPE {
		t := types.NewNamed(n)
		n.SetOp(ir.OTYPE)
		n.SetPos(pos)
		n.SetType(t)
		n.Class_ = ir.PEXTERN
	}

	t := n.Type()
	if t == nil {
		base.Fatalf("importtype %v", s)
	}
	return t
}

// importvar declares symbol s as an imported variable with type t.
// ipkg is the package being imported
func importvar(ipkg *types.Pkg, pos src.XPos, s *types.Sym, t *types.Type) {
	n := importobj(ipkg, pos, s, ir.ONAME, ir.PEXTERN, t)
	if n == nil {
		return
	}

	if base.Flag.E != 0 {
		fmt.Printf("import var %v %L\n", s, t)
	}
}
