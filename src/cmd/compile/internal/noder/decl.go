// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
)

func (g *irgen) decls(decls []syntax.Decl) []ir.Node {
	var res ir.Nodes
	for _, decl := range decls {
		switch decl := decl.(type) {
		case *syntax.ImportDecl:
			// already handled
		case *syntax.ConstDecl:
			g.constDecl(&res, decl)
		case *syntax.FuncDecl:
			g.funcDecl(&res, decl)
		case *syntax.TypeDecl:
			if ir.CurFunc == nil {
				continue // already handled
			}
			g.typeDecl(&res, decl)
		case *syntax.VarDecl:
			g.varDecl(&res, decl)
		default:
			g.unhandled("declaration", decl)
		}
	}

	// TODO(mdempsky): Remove this clumsy hack.
	if ir.CurFunc == nil {
		for _, n := range res {
			typecheck.Stmt(n)
		}
		typecheck.Target.Decls = append(typecheck.Target.Decls, res...)
	}

	return res
}

func (g *irgen) importDecl(p *noder, decl *syntax.ImportDecl) {
	// TODO(mdempsky): Merge with gcimports so we don't have to import
	// packages twice.

	g.pragmaFlags(decl.Pragma, 0)

	ipkg := importfile(decl)
	if ipkg == ir.Pkgs.Unsafe {
		p.importedUnsafe = true
	}
}

func (g *irgen) constDecl(out *ir.Nodes, decl *syntax.ConstDecl) {
	g.pragmaFlags(decl.Pragma, 0)

	for _, name := range decl.NameList {
		name, obj := g.def(name)
		name.SetVal(obj.(*types2.Const).Val())
	}
}

func (g *irgen) funcDecl(out *ir.Nodes, decl *syntax.FuncDecl) {
	fn := ir.NewFunc(g.pos(decl))
	fn.Nname, _ = g.def(decl.Name)
	fn.Nname.Func = fn
	fn.Nname.Defn = fn

	fn.Pragma = g.pragmaFlags(decl.Pragma, funcPragmas)
	if fn.Pragma&ir.Systemstack != 0 && fn.Pragma&ir.Nosplit != 0 {
		base.ErrorfAt(fn.Pos(), "go:nosplit and go:systemstack cannot be combined")
	}

	if decl.Name.Value == "init" && decl.Recv == nil {
		typecheck.Target.Inits = append(typecheck.Target.Inits, fn)
	}

	typecheck.Func(fn)

	g.funcBody(fn, decl.Recv, decl.Type, decl.Body)

	typecheck.DeclContext = ir.PEXTERN
}

func (g *irgen) typeDecl(out *ir.Nodes, decl *syntax.TypeDecl) {
	if decl.Alias {
		if !types.AllowsGoVersion(types.LocalPkg, 1, 9) {
			base.ErrorfAt(g.pos(decl), "type aliases only supported as of -lang=go1.9")
		}

		name, _ := g.def(decl.Name)
		g.pragmaFlags(decl.Pragma, 0)

		// TODO(mdempsky): This matches how typecheckdef marks aliases for
		// export, but this won't generalize to exporting function-scoped
		// type aliases. We should maybe just use n.Alias() instead.
		if ir.CurFunc == nil {
			name.Sym().Def = ir.TypeNode(name.Type())
		}

		return
	}

	name, obj := g.def(decl.Name)
	ntyp, otyp := name.Type(), obj.Type()
	ntyp.Vargen = name.Vargen

	pragmas := g.pragmaFlags(decl.Pragma, typePragmas)
	name.SetPragma(pragmas) // TODO(mdempsky): Is this still needed?

	if pragmas&ir.NotInHeap != 0 {
		ntyp.SetNotInHeap(true)
	}

	// Note: We need to use g.typeExpr(decl.Type) here rather than
	// g.typ(otyp.Underlying()) to make sure we handle //go:notinheap
	// correctly for chained defined types.
	ntyp.SetUnderlying(g.typeExpr(decl.Type))

	if otyp, ok := otyp.(*types2.Named); ok && otyp.NumMethods() != 0 {
		methods := make([]*types.Field, otyp.NumMethods())
		for i := range methods {
			m := otyp.Method(i)
			meth := g.obj(m)
			methods[i] = types.NewField(meth.Pos(), g.selector(m), meth.Type())
			methods[i].Nname = meth
		}
		ntyp.Methods().Set(methods)
	}
}

func (g *irgen) varDecl(out *ir.Nodes, decl *syntax.VarDecl) {
	pos := g.pos(decl)
	names := g.defs(decl.NameList)
	values := g.exprList(decl.Values)

	if decl.Pragma != nil {
		pragma := decl.Pragma.(*pragmas)
		if err := varEmbed(g.makeXPos, names[0], decl, pragma); err != nil {
			base.ErrorfAt(g.pos(decl), "%s", err.Error())
		}
		g.reportUnused(pragma)
	}

	var as2 *ir.AssignListStmt
	if len(values) != 0 && len(names) != len(values) {
		as2 = ir.NewAssignListStmt(pos, ir.OAS2, make([]ir.Node, len(names)), values)
	}

	for i, name := range names {
		if ir.CurFunc != nil {
			out.Append(ir.NewDecl(pos, ir.ODCL, name))
		}
		if as2 != nil {
			as2.Lhs[i] = name
			name.Defn = as2
		} else {
			as := ir.NewAssignStmt(pos, name, nil)
			if len(values) != 0 {
				as.Y = values[i]
				name.Defn = as
			} else if ir.CurFunc == nil {
				name.Defn = as
			}
			out.Append(as)
		}
	}
	if as2 != nil {
		out.Append(as2)
	}
}

// pragmaFlags returns any specified pragma flags included in allowed,
// and reports errors about any other, unexpected pragmas.
func (g *irgen) pragmaFlags(pragma syntax.Pragma, allowed ir.PragmaFlag) ir.PragmaFlag {
	if pragma == nil {
		return 0
	}
	p := pragma.(*pragmas)
	present := p.Flag & allowed
	p.Flag &^= allowed
	g.reportUnused(p)
	return present
}

// reportUnused reports errors about any unused pragmas.
func (g *irgen) reportUnused(pragma *pragmas) {
	for _, pos := range pragma.Pos {
		if pos.Flag&pragma.Flag != 0 {
			base.ErrorfAt(g.makeXPos(pos.Pos), "misplaced compiler directive")
		}
	}
	if len(pragma.Embeds) > 0 {
		for _, e := range pragma.Embeds {
			base.ErrorfAt(g.makeXPos(e.Pos), "misplaced go:embed directive")
		}
	}
}
