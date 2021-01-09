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
	"cmd/internal/src"
)

func (g *irgen) defs(names []*syntax.Name) []*ir.Name {
	res := make([]*ir.Name, len(names))
	for i, name := range names {
		res[i], _ = g.def(name)
	}
	return res
}

func (g *irgen) def(name *syntax.Name) (*ir.Name, types2.Object) {
	obj, ok := g.info.Defs[name]
	if !ok {
		base.FatalfAt(g.pos(name), "unknown name %v", name)
	}
	return g.obj(obj), obj
}

func (g *irgen) use(name *syntax.Name) *ir.Name {
	obj, ok := g.info.Uses[name]
	if !ok {
		base.FatalfAt(g.pos(name), "unknown name %v", name)
	}
	return g.capture(ir.CurFunc, g.obj(obj))
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

	fn.ClosureVars = append(fn.ClosureVars, c)

	// Add to stack of active closure variables.
	// Popped at the end of funcBody.
	c.Outer = n.Innermost
	n.Innermost = c

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
		n := typecheck.Resolve(ir.NewIdent(src.NoXPos, sym))
		if n, ok := n.(*ir.Name); ok {
			return n
		}
		base.FatalfAt(g.pos(obj), "why didn't we resolve %v successfully? got %+v", obj, n)
	}

	if name, ok := g.objs[obj]; ok {
		return name
	}

	op, class, defined := ir.ONAME, typecheck.DeclContext, false

	if obj.Parent() == g.self.Scope() {
		class = ir.PEXTERN // forward reference to package-block declaration
	}

	switch obj := obj.(type) {
	case *types2.Const:
		op = ir.OLITERAL
	case *types2.Func:
		if obj.Name() == "init" && obj.Type().(*types2.Signature).Recv() == nil {
			sym = renameinit()
		}
		class = ir.PFUNC
	case *types2.TypeName:
		op, defined = ir.OTYPE, !obj.IsAlias()
	case *types2.Var:
		// ok
	default:
		g.unhandled("object", obj)
	}

	name := ir.NewDeclNameAt(g.pos(obj), op, sym)
	g.objs[obj] = name
	if defined {
		name.SetType(types.NewNamed(name))
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

	return name
}

func (g *irgen) objRenamed(obj types2.Object, prefix string) *ir.Name {
	if obj == nil {
		base.Fatalf("missing obj")
	}

	want := ""
	switch obj.Name() {
	case "":
		want = "~r"
	case "_":
		want = "~b"
	default:
		base.FatalfAt(g.pos(obj), "unexpected renamed object: %v", obj)
	}
	if want != prefix {
		base.FatalfAt(g.pos(obj), "inconsistent prefixes: %q != %q", want, prefix)
	}

	sym := typecheck.LookupNum(prefix, len(ir.CurFunc.Dcl))

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
		fn := ir.CurFunc
		n.Curfn = fn
		switch n.Op() {
		case ir.ONAME:
			fn.Dcl = append(fn.Dcl, n)
			n.Vargen = int32(len(ir.CurFunc.Dcl))
			n.SetFrameOffset(0) // TODO(mdempsky): Seems unnecessary.
		case ir.OTYPE:
			declare_typevargen++
			n.Vargen = int32(declare_typevargen)
		}
	}
}

var declare_typevargen int
