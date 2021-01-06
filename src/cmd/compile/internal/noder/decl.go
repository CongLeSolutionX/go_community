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
	"go/constant"
	"go/token"
)

func (p *irgen) decls(top *noder, decls []syntax.Decl) []ir.Node {
	var res []ir.Node
	for _, decl := range decls {
		switch decl := decl.(type) {
		case *syntax.ImportDecl:
			p.importDecl(top, decl)
		case *syntax.ConstDecl:
			p.defs(decl.NameList)
		case *syntax.FuncDecl:
			p.funcDecl(decl)
		case *syntax.TypeDecl:
			p.def(decl.Name)
		case *syntax.VarDecl:
			p.varDecl(decl)
		}
	}
	return res
}

func (p *irgen) importDecl(top *noder, decl *syntax.ImportDecl) {
	ipkg := importfile(constant.MakeFromLiteral(decl.Path.Value, token.STRING, 0))

	if !ipkg.Direct {
		typecheck.Target.Imports = append(typecheck.Target.Imports, ipkg)
	}
	ipkg.Direct = true

	if ipkg == ir.Pkgs.Unsafe {
		top.importedUnsafe = true
	}
}

func (p *irgen) funcDecl(decl *syntax.FuncDecl) {
	fn := ir.NewFunc(p.pos(decl))
	fn.Nname = p.def(decl.Name)
	fn.Nname.Func = fn
	fn.Nname.Defn = fn

	if decl.Name.Value == "init" {
		typecheck.Target.Inits = append(typecheck.Target.Inits, fn)
	}

	p.funcBody(fn, decl.Body)
}

func (p *irgen) varDecl(decl *syntax.VarDecl) []ir.Node {
	pos := p.pos(decl)
	names := p.defs(decl.NameList)
	values := p.exprList(decl.Values)

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
	res = append(res, as)
	return res
}

func (p *irgen) def(name *syntax.Name) *ir.Name {
	if obj, ok := p.info.Defs[name]; ok {
		return p.obj(obj)
	}
	base.FatalfAt(p.pos(name), "unknown name %v", name)
	panic("unreachable")
}

func (p *irgen) defs(names []*syntax.Name) []*ir.Name {
	res := make([]*ir.Name, len(names))
	for i, name := range names {
		res[i] = p.def(name)
	}
	return res
}

func (p *irgen) use(name *syntax.Name) ir.Node {
	if obj, ok := p.info.Uses[name]; ok {
		// TODO(mdempsky): Create closure variables.
		return p.obj(obj)
	}
	base.FatalfAt(p.pos(name), "unknown name %v", name)
	panic("unreachable")
}

func (p *irgen) obj(obj types2.Object) *ir.Name {
	if obj.Pkg() == nil {
		return types.BuiltinPkg.Lookup(obj.Name()).Def.(*ir.Name)
	}
	if name, ok := p.objs[obj]; ok {
		return name
	}

	var name *ir.Name
	pos := p.pos(obj)
	top := obj.Parent() == obj.Pkg().Scope()

	do := func(op ir.Op, ctxt ir.Class, definedType bool) {
		name = ir.NewDeclNameAt(pos, op, p.sym(obj))
		if definedType {
			name.SetType(types.NewNamed(name))
		} else {
			name.SetType(p.typ(obj.Type()))
		}
		name.SetTypecheck(1)
		name.SetWalkdef(1)

		if top {
			if ctxt == ir.PFUNC && obj.Name() == "init" {
				name.SetSym(renameinit())
			}
			typecheck.Declare(name, ctxt)
		}
	}

	switch obj := obj.(type) {
	case *types2.Const:
		do(ir.OLITERAL, ir.PEXTERN, false)
		name.SetVal(obj.Val())
	case *types2.Func:
		do(ir.ONAME, ir.PFUNC, false)
	case *types2.TypeName:
		do(ir.OTYPE, ir.PEXTERN, !obj.IsAlias())
		p.todos = append(p.todos, todo{name, obj})
	case *types2.Var:
		do(ir.ONAME, ir.PEXTERN, false)
		if !top {
			// TODO(mdempsky): Validate scopes.
			name.Curfn = ir.CurFunc
			ir.CurFunc.Dcl = append(ir.CurFunc.Dcl, name)
		}
	default:
		base.FatalfAt(p.pos(obj), "unhandled object: %v (%T)", obj, obj)
	}

	if p.objs == nil {
		p.objs = make(map[types2.Object]*ir.Name)
	}
	p.objs[obj] = name
	return name
}
