// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"go/constant"
	"go/token"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
)

func (g *irgen) decls(p *noder, decls []syntax.Decl) []ir.Node {
	var res []ir.Node
	for _, decl := range decls {
		switch decl := decl.(type) {
		case *syntax.ImportDecl:
			g.importDecl(p, decl)
		case *syntax.ConstDecl:
			g.defs(decl.NameList)
		case *syntax.FuncDecl:
			g.funcDecl(decl)
		case *syntax.TypeDecl:
			g.def(decl.Name)
		case *syntax.VarDecl:
			res = append(res, g.varDecl(decl)...)
		default:
			g.unhandled("declaration", decl)
		}
	}
	return res
}

func (g *irgen) importDecl(p *noder, decl *syntax.ImportDecl) {
	// TODO(mdempsky): Merge into gcimports or remove altogether.

	ipkg := importfile(constant.MakeFromLiteral(decl.Path.Value, token.STRING, 0))

	if !ipkg.Direct {
		typecheck.Target.Imports = append(typecheck.Target.Imports, ipkg)
		ipkg.Direct = true
	}

	if ipkg == ir.Pkgs.Unsafe {
		p.importedUnsafe = true
	}
}

func (g *irgen) funcDecl(decl *syntax.FuncDecl) {
	fn := ir.NewFunc(g.pos(decl))
	fn.Nname = g.def(decl.Name)
	fn.Nname.Func = fn
	fn.Nname.Defn = fn

	if decl.Name.Value == "init" {
		typecheck.Target.Inits = append(typecheck.Target.Inits, fn)
	}

	g.funcBody(fn, decl.Body)
}

func (g *irgen) varDecl(decl *syntax.VarDecl) []ir.Node {
	pos := g.pos(decl)
	names := g.defs(decl.NameList)
	values := g.exprList(decl.Values)

	// TODO(mdempsky): To avoid typecheck clobbering things later,
	// I'm not setting name.Defn. That is, I'm basically
	// representing "var x T = v" as separate "var x T; x = v"
	// statements. This works correctly, but it will interfere with
	// frontend optimizations that recognize never-reassigned
	// variables.

	var as *ir.AssignListStmt
	if len(values) != 0 && len(names) != len(values) {
		as = ir.NewAssignListStmt(pos, ir.OAS2, make([]ir.Node, len(names)), values)
	}

	var res []ir.Node
	for i, name := range names {
		res = append(res, ir.NewDecl(pos, ir.ODCL, name), ir.NewAssignStmt(pos, name, nil))
		if as != nil {
			as.Lhs[i] = name
		} else if len(values) != 0 {
			res = append(res, ir.NewAssignStmt(pos, name, values[i]))
		}
	}
	if as != nil {
		res = append(res, as)
	}
	return res
}

func (g *irgen) def(name *syntax.Name) *ir.Name {
	if obj, ok := g.info.Defs[name]; ok {
		return g.obj(obj)
	}
	base.FatalfAt(g.pos(name), "unknown name %v", name)
	panic("unreachable")
}

func (g *irgen) defs(names []*syntax.Name) []*ir.Name {
	res := make([]*ir.Name, len(names))
	for i, name := range names {
		res[i] = g.def(name)
	}
	return res
}

func (g *irgen) use(name *syntax.Name) ir.Node {
	if obj, ok := g.info.Uses[name]; ok {
		// TODO(mdempsky): Create closure variables.
		return g.obj(obj)
	}
	base.FatalfAt(g.pos(name), "unknown name %v", name)
	panic("unreachable")
}

func (g *irgen) obj(obj types2.Object) *ir.Name {
	if obj.Pkg() == nil || obj.Pkg() == types2.Unsafe {
		return g.pkg(obj.Pkg()).Lookup(obj.Name()).Def.(*ir.Name)
	}
	if name, ok := g.objs[obj]; ok {
		return name
	}

	pos, sym := g.pos(obj), g.sym(obj)
	top := obj.Parent() == obj.Pkg().Scope()

	var name *ir.Name
	do := func(op ir.Op, ctxt ir.Class, definedType bool) {
		name = ir.NewDeclNameAt(pos, op, sym)
		if definedType {
			name.SetType(types.NewNamed(name))
		} else {
			name.SetType(g.typ(obj.Type()))
		}
		name.SetTypecheck(1)
		name.SetWalkdef(1)

		if top {
			typecheck.Declare(name, ctxt)
		}
	}

	switch obj := obj.(type) {
	case *types2.Const:
		do(ir.OLITERAL, ir.PEXTERN, false)
		name.SetVal(obj.Val())
	case *types2.Func:
		if obj.Name() == "init" {
			sym = renameinit()
		}
		do(ir.ONAME, ir.PFUNC, false)
	case *types2.TypeName:
		if obj.IsAlias() {
			do(ir.OTYPE, ir.PEXTERN, false)
		} else {
			// TODO(mdempsky): This is a little clumsy.
			do(ir.OTYPE, ir.PEXTERN, true)
			g.todos = append(g.todos, todo{name, obj})
		}
	case *types2.Var:
		if sym == nil || sym.Name == "_" {
			// TODO(mdempsky): Only rename parameters, and
			// use the normal ~r and ~b names, instead of
			// abusing renameinit.
			sym = renameinit()
		}
		do(ir.ONAME, ir.PEXTERN, false)

		// TODO(mdempsky): Why is top sometimes not the same
		// as ir.CurFunc != nil?
		if !top && ir.CurFunc != nil {
			// TODO(mdempsky): Validate scopes to make
			// sure we're handling function literals
			// correctly.
			name.Curfn = ir.CurFunc
			ir.CurFunc.Dcl = append(ir.CurFunc.Dcl, name)
		}
	default:
		g.unhandled("object", obj)
	}

	if g.objs == nil {
		g.objs = make(map[types2.Object]*ir.Name)
	}
	g.objs[obj] = name
	return name
}
