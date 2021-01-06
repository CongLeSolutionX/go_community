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
	"cmd/internal/src"
)

func (g *irgen) decls(p *noder, decls []syntax.Decl) []ir.Node {
	var res []ir.Node
	for _, decl := range decls {
		switch decl := decl.(type) {
		case *syntax.ImportDecl:
			// already handled
		case *syntax.ConstDecl:
			g.defs(decl.NameList)
		case *syntax.FuncDecl:
			g.funcDecl(decl)
		case *syntax.TypeDecl:
			g.typeDecl(decl)
		case *syntax.VarDecl:
			res = append(res, g.varDecl(decl)...)
		default:
			g.unhandled("declaration", decl)
		}
	}

	if p != nil {
		// TODO(mdempsky): Handle this better.
		for _, n := range res {
			if n.Op() != ir.ODCL {
				typecheck.Target.Decls = append(typecheck.Target.Decls, typecheck.Stmt(n))
			}
		}
	}

	return res
}

func (g *irgen) pragma(pragma syntax.Pragma, allowed ir.PragmaFlag) ir.PragmaFlag {
	if pragma == nil {
		return 0
	}
	p := pragma.(*pragmas)
	present := p.Flag & allowed
	p.Flag &^= allowed
	g.checkUnused(p)
	return present
}

func (g *irgen) checkUnused(pragma *pragmas) {
	for _, pos := range pragma.Pos {
		if pos.Flag&pragma.Flag != 0 {
			base.ErrorfAt(g.pos0(pos.Pos), "misplaced compiler directive")
		}
	}
	if len(pragma.Embeds) > 0 {
		for _, e := range pragma.Embeds {
			base.ErrorfAt(g.pos0(e.Pos), "misplaced go:embed directive")
		}
	}
}

func (g *irgen) importDecl(p *noder, decl *syntax.ImportDecl) {
	// TODO(mdempsky): Merge into gcimports or remove altogether.

	path := constant.MakeFromLiteral(decl.Path.Value, token.STRING, 0)

	ipkg := importfile(path)
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

	fn.Pragma = g.pragma(decl.Pragma, funcPragmas)
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

func (g *irgen) typeDecl(decl *syntax.TypeDecl) {
	name := g.def(decl.Name)
	name.Type().Vargen = name.Vargen

	allowed := typePragmas
	if decl.Alias {
		allowed = 0
		if !types.AllowsGoVersion(types.LocalPkg, 1, 9) {
			base.ErrorfAt(g.pos(decl), "type aliases only supported as of -lang=go1.9")
		}
	}
	name.SetPragma(g.pragma(decl.Pragma, allowed))

	// TODO(mdempsky): I'm not convinced this is early enough to be safe. The
	// implementation of this flag should perhaps be reworked.
	if name.Pragma()&ir.NotInHeap != 0 {
		name.Type().SetNotInHeap(true)
	}
}

func (g *irgen) varDecl(decl *syntax.VarDecl) []ir.Node {
	pos := g.pos(decl)
	names := g.defs(decl.NameList)
	values := g.exprList(decl.Values)

	if pragma, ok := decl.Pragma.(*pragmas); ok {
		if embeds := pragma.Embeds; len(embeds) > 0 && len(names) == 1 && len(values) == 0 && typecheck.DeclContext == ir.PEXTERN {
			var list []ir.Embed
			for _, e := range embeds {
				list = append(list, ir.Embed{Pos: g.pos0(e.Pos), Patterns: e.Patterns})
			}
			names[0].Embed = &list
			typecheck.Target.Embeds = append(typecheck.Target.Embeds, names[0])
			pragma.Embeds = nil
		}
		g.checkUnused(pragma)
	}

	var as2 *ir.AssignListStmt
	if len(values) != 0 && len(names) != len(values) {
		as2 = ir.NewAssignListStmt(pos, ir.OAS2, make([]ir.Node, len(names)), values)
	}

	var res []ir.Node
	for i, name := range names {
		res = append(res, ir.NewDecl(pos, ir.ODCL, name))
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
			res = append(res, as)
		}
	}
	if as2 != nil {
		res = append(res, as2)
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

func (g *irgen) use(name *syntax.Name) *ir.Name {
	if obj, ok := g.info.Uses[name]; ok {
		return g.capture(ir.CurFunc, g.obj(obj))
	}
	base.FatalfAt(g.pos(name), "unknown name %v", name)
	panic("unreachable")
}

func (g *irgen) capture(fn *ir.Func, n *ir.Name) *ir.Name {
	if n.Op() != ir.ONAME || n.Curfn == nil || n.Curfn == fn {
		return n
	}

	c := n.Innermost
	if c != nil && c.Curfn == fn {
		return c
	}

	// Do not have a closure var for the active closure yet; make one.
	c = ir.NewNameAt(n.Pos(), n.Sym())
	c.Curfn = fn
	c.Class = ir.PAUTOHEAP
	c.SetIsClosureVar(true)
	c.Defn = n
	c.SetType(n.Type())
	c.SetTypecheck(1)

	// Link into list of active closure variables.
	// Popped from list in func funcLit.
	c.Outer = g.capture(fn.Outer, n)
	n.Innermost = c

	fn.ClosureVars = append(fn.ClosureVars, c)

	return c
}

func (g *irgen) obj(obj types2.Object) *ir.Name {
	return g.obj0(obj, g.objSym(obj))
}

func (g *irgen) obj0(obj types2.Object, sym *types.Sym) *ir.Name {
	if sym == nil {
		base.FatalfAt(g.pos(obj), "missing sym for %v", obj)
	}

	if obj.Pkg() != g.self {
		if sym.Def != nil {
			return sym.Def.(*ir.Name)
		}
		return typecheck.Resolve(ir.NewIdent(src.NoXPos, sym)).(*ir.Name)
	}

	if name, ok := g.objs[obj]; ok {
		return name
	}

	class := typecheck.DeclContext
	if obj.Pkg() != g.self {
		class = ir.PEXTERN
	}

	pos := g.pos(obj)

	var name *ir.Name
	do := func(op ir.Op, defined *types2.TypeName) {
		name = ir.NewDeclNameAt(pos, op, sym)
		g.objs[obj] = name
		if defined != nil {
			name.SetType(types.NewNamed(name))
			g.resolve(name, defined)
		} else {
			name.SetType(g.objType(obj))
		}

		// We already know name's type, but typecheck is
		// really eager to try to recompute it. This seems to
		// stop it from causing any trouble.
		name.Ntype = ir.TypeNode(name.Type())
		name.SetTypecheck(1)
		name.SetWalkdef(1)

		g.declare(name, class)
	}

	switch obj := obj.(type) {
	case *types2.Const:
		do(ir.OLITERAL, nil)
		name.SetVal(obj.Val())
	case *types2.Func:
		if obj.Name() == "init" && obj.Type().(*types2.Signature).Recv() == nil {
			sym = renameinit()
		}
		class = ir.PFUNC
		do(ir.ONAME, nil)
		if obj.Pkg() != g.self {
			// Stub ODCLFUNC for imported functions.
			name.Func = ir.NewFunc(name.Pos())
			name.Func.Nname = name
		}
	case *types2.TypeName:
		if obj.IsAlias() {
			do(ir.OTYPE, nil)
			// TODO(mdempsky): This matches how typecheckdef marks aliases for
			// export, but this won't generalize to function-scoped type
			// aliases. We should probably just use n.Alias() instead.
			if class == ir.PEXTERN {
				name.Sym().Def = ir.TypeNode(name.Type())
			}
		} else {
			do(ir.OTYPE, obj)
		}
	case *types2.Var:
		do(ir.ONAME, nil)
	default:
		g.unhandled("object", obj)
	}

	return name
}

func (g *irgen) objRenamed(obj types2.Object, prefix string, gen *int) *ir.Name {
	if obj == nil {
		base.Fatalf("missing obj")
	}

	sym := typecheck.LookupNum(prefix, *gen)
	*gen++

	return g.obj0(obj, sym)
}

func (g *irgen) objSym(obj types2.Object) *types.Sym {
	if obj, ok := obj.(*types2.Func); ok && obj.Name() != "_" {
		sig := obj.Type().(*types2.Signature)
		if recv := sig.Recv(); recv != nil {
			return ir.MethodSym(g.typ(recv.Type()), g.selector(obj))
		}
	}

	return g.sym(obj)
}

func (g *irgen) objType(obj types2.Object) *types.Type {
	if obj, ok := obj.(*types2.Func); ok {
		sig := obj.Type().(*types2.Signature)
		if recv := sig.Recv(); recv != nil {
			return g.signature(g.param(recv), sig)
		}
	}

	return g.typ(obj.Type())
}

// declare records that Name n declares symbol n.Sym in the specified
// declaration context.
func (g *irgen) declare(n *ir.Name, class ir.Class) {
	if ir.IsBlank(n) {
		return
	}

	sym := n.Sym()
	if sym.Pkg != types.LocalPkg && class != ir.PFUNC {
		class = ir.PEXTERN
	}

	n.Class = class
	if n.Class == ir.PFUNC {
		sym.SetFunc(true)
	}

	switch class {
	case ir.PEXTERN:
		typecheck.Target.Externs = append(typecheck.Target.Externs, n)
		fallthrough
	case ir.PFUNC:
		sym.Def = n

		if n.Class == ir.PFUNC && n.Type().Recv() != nil {
			break
		}
		if types.IsExported(sym.Name) {
			typecheck.Export(n)
		}
		if base.Flag.AsmHdr != "" && !n.Sym().Asm() {
			n.Sym().SetAsm(true)
			typecheck.Target.Asms = append(typecheck.Target.Asms, n)
		}

	default:
		n.Curfn = ir.CurFunc
		switch n.Op() {
		case ir.ONAME:
			ir.CurFunc.Dcl = append(ir.CurFunc.Dcl, n)
			vargen++
			n.Vargen = int32(vargen)
			n.SetFrameOffset(0) // TODO(mdempsky): Seems unnecessary.
		case ir.OTYPE:
			declare_typevargen++
			n.Vargen = int32(declare_typevargen)
		}
	}
}

var vargen, declare_typevargen int
