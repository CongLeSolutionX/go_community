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

// capture returns a Name suitable for referring to n from within
// fn. If n is a variable declared in a function that encloses fn,
// then capture returns a closure pseudo-variable denoting
// n. Otherwise, it returns n itself.
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

// obj returns the Name that represents the given object. If no such
// Name exists yet, it will be implicitly created.
//
// For objects declared at function scope, ir.CurFunc must already be
// set to the respective function when the Name is created.
func (g *irgen) obj(obj types2.Object) *ir.Name {
	// For imported objects, we use iimport directly instead of mapping
	// the types2 representation.
	if obj.Pkg() != g.self {
		sym := g.sym(obj)
		if sym.Def != nil {
			return sym.Def.(*ir.Name)
		}
		n := typecheck.Resolve(ir.NewIdent(src.NoXPos, sym))
		if n, ok := n.(*ir.Name); ok {
			return n
		}
		base.FatalfAt(g.pos(obj), "failed to resolve %v", obj)
	}

	if name, ok := g.objs[obj]; ok {
		return name // previously mapped
	}

	var name *ir.Name
	pos := g.pos(obj)

	class := typecheck.DeclContext
	if obj.Parent() == g.self.Scope() {
		class = ir.PEXTERN // forward reference to package-block declaration
	}

	// "You are in a maze of twisting little passages, all different."
	switch obj := obj.(type) {
	case *types2.Const:
		name = g.objCommon(pos, ir.OLITERAL, g.sym(obj), class, g.typ(obj.Type()))

	case *types2.Func:
		sig := obj.Type().(*types2.Signature)
		var sym *types.Sym
		var typ *types.Type
		if recv := sig.Recv(); recv == nil {
			if obj.Name() == "init" {
				sym = renameinit()
			} else {
				sym = g.sym(obj)
			}
			typ = g.typ(sig)
		} else {
			sym = ir.MethodSym(g.typ(recv.Type()), g.selector(obj))
			typ = g.signature(g.param(recv), sig)
		}
		name = g.objCommon(pos, ir.ONAME, sym, ir.PFUNC, typ)

	case *types2.TypeName:
		if obj.IsAlias() {
			name = g.objCommon(pos, ir.OTYPE, g.sym(obj), class, g.typ(obj.Type()))
		} else {
			name = ir.NewDeclNameAt(pos, ir.OTYPE, g.sym(obj))
			g.objFinish(name, class, types.NewNamed(name))
		}

	case *types2.Var:
		var sym *types.Sym
		if class == ir.PPARAMOUT {
			num := ir.CurFunc.Type().NumParams() + len(ir.CurFunc.Dcl)
			switch obj.Name() {
			case "":
				sym = typecheck.LookupNum("~r", num)
			case "_":
				sym = typecheck.LookupNum("~b", num)
			}
		}
		if sym == nil {
			sym = g.sym(obj)
		}
		name = g.objCommon(pos, ir.ONAME, sym, class, g.typ(obj.Type()))

	default:
		g.unhandled("object", obj)
	}

	g.objs[obj] = name
	return name
}

func (g *irgen) objCommon(pos src.XPos, op ir.Op, sym *types.Sym, class ir.Class, typ *types.Type) *ir.Name {
	name := ir.NewDeclNameAt(pos, op, sym)
	g.objFinish(name, class, typ)
	return name
}

func (g *irgen) objFinish(name *ir.Name, class ir.Class, typ *types.Type) {
	sym := name.Sym()

	name.SetType(typ)
	name.Class = class
	if name.Class == ir.PFUNC {
		sym.SetFunc(true)
	}

	// We already know name's type, but typecheck is really eager to try
	// recomputing it later. This appears to prevent that at least.
	name.Ntype = ir.TypeNode(typ)
	name.SetTypecheck(1)
	name.SetWalkdef(1)

	if ir.IsBlank(name) {
		return
	}

	switch class {
	case ir.PEXTERN:
		typecheck.Target.Externs = append(typecheck.Target.Externs, name)
		fallthrough
	case ir.PFUNC:
		sym.Def = name
		if name.Class == ir.PFUNC && name.Type().Recv() != nil {
			break // methods are exported with their receiver type
		}
		if types.IsExported(sym.Name) {
			typecheck.Export(name)
		}
		if base.Flag.AsmHdr != "" && !name.Sym().Asm() {
			name.Sym().SetAsm(true)
			typecheck.Target.Asms = append(typecheck.Target.Asms, name)
		}

	default:
		// Function-scoped declaration.
		name.Curfn = ir.CurFunc
		switch name.Op() {
		case ir.ONAME:
			ir.CurFunc.Dcl = append(ir.CurFunc.Dcl, name)
			name.Vargen = int32(len(ir.CurFunc.Dcl))
		case ir.OTYPE:
			declare_typevargen++
			name.Vargen = int32(declare_typevargen)
		}
	}
}

var declare_typevargen int
