// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"fmt"
	"os"

	"cmd/compile/internal/base"
	"cmd/compile/internal/dwarfgen"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
	"cmd/internal/src"
)

// check2 type checks a Go package using types2, and then generates IR
// using the results.
func check2(noders []*noder) {
	if base.SyntaxErrors() != 0 {
		base.ErrorExit()
	}

	// setup and syntax error reporting
	var m posMap
	files := make([]*syntax.File, len(noders))
	for i, p := range noders {
		m.join(&p.posMap)
		files[i] = p.file
	}

	// typechecking
	conf := types2.Config{
		InferFromConstraints:  true,
		IgnoreLabels:          true, // parser already checked via syntax.CheckBranches mode
		CompilerErrorMessages: true, // use error strings matching existing compiler errors
		Error: func(err error) {
			terr := err.(types2.Error)
			if len(terr.Msg) > 0 && terr.Msg[0] == '\t' {
				// types2 reports error clarifications via separate
				// error messages which are indented with a tab.
				// Ignore them to satisfy tools and tests that expect
				// only one error in such cases.
				// TODO(gri) Need to adjust error reporting in types2.
				return
			}
			base.ErrorfAt(m.makeXPos(terr.Pos), "%s", terr.Msg)
		},
		Importer: &gcimports{
			packages: make(map[string]*types2.Package),
		},
		Sizes: &gcSizes{},
	}
	info := types2.Info{
		Types:      make(map[syntax.Expr]types2.TypeAndValue),
		Defs:       make(map[*syntax.Name]types2.Object),
		Uses:       make(map[*syntax.Name]types2.Object),
		Selections: make(map[*syntax.SelectorExpr]*types2.Selection),
		Implicits:  make(map[syntax.Node]types2.Object),
		Scopes:     make(map[syntax.Node]*types2.Scope),
		Inferred:   make(map[syntax.Expr]types2.Inferred),
		// expand as needed
	}
	pkg, err := conf.Check(base.Ctxt.Pkgpath, files, &info)
	files = nil
	if err != nil {
		base.FatalfAt(src.NoXPos, "conf.Check error: %v", err)
	}
	base.ExitIfErrors()
	if base.Flag.G < 2 {
		os.Exit(0)
	}

	g := irgen{
		target: typecheck.Target,
		self:   pkg,
		info:   &info,
		posMap: m,
		objs:   make(map[types2.Object]*ir.Name),
		typs:   make(map[types2.Type]*types.Type),
	}
	g.generate(noders)

	if base.Flag.G < 3 {
		os.Exit(0)
	}
}

type irgen struct {
	target *ir.Package
	self   *types2.Package
	info   *types2.Info

	posMap
	objs   map[types2.Object]*ir.Name
	typs   map[types2.Type]*types.Type
	marker dwarfgen.ScopeMarker
}

func (g *irgen) generate(noders []*noder) {
	types.LocalPkg.Name = g.self.Name()
	typecheck.TypecheckAllowed = true

	// Prevent size calculations until we set the underlying type
	// for all package-block defined types.
	types.DeferCheckSize()

	// At this point, types2 has already handled name resolution and
	// type checking. We just need to map from its object and type
	// representations to those currently used by the rest of the
	// compiler. This happens mostly in 3 passes.

	// 1. Process all import declarations. We use the compiler's own
	// importer for this, rather than types2's gcimporter-derived one,
	// to handle extensions and inline function bodies correctly.
	//
	// Also, we need to do this in a separate pass, because mappings are
	// instantiated on demand. If we interleaved processing import
	// declarations with other declarations, it's likely we'd end up
	// wanting to map an object/type from another source file, but not
	// yet have the import data it relies on.
	declLists := make([][]syntax.Decl, len(noders))
Outer:
	for i, p := range noders {
		g.pragmaFlags(p.file.Pragma, ir.GoBuildPragma)
		for j, decl := range p.file.DeclList {
			switch decl := decl.(type) {
			case *syntax.ImportDecl:
				g.importDecl(p, decl)
			default:
				declLists[i] = p.file.DeclList[j:]
				continue Outer // no more ImportDecls
			}
		}
	}
	types.LocalPkg.Height = myheight

	// 2. Process all package-block type declarations. As with imports,
	// we need to make sure all types are properly instantiated before
	// trying to map any expressions that utilize them. In particular,
	// we need to make sure type pragmas are already known (see comment
	// in irgen.typeDecl).
	//
	// We could perhaps instead defer processing of package-block
	// variable initializers and function bodies, like noder does, but
	// special-casing just package-block type declarations minimizes the
	// differences between processing package-block and function-scoped
	// declarations.
	for _, declList := range declLists {
		for _, decl := range declList {
			switch decl := decl.(type) {
			case *syntax.TypeDecl:
				g.typeDecl((*ir.Nodes)(&g.target.Decls), decl)
			}
		}
	}
	types.ResumeCheckSize()

	// 3. Process all remaining declarations.
	for _, declList := range declLists {
		g.target.Decls = append(g.target.Decls, g.decls(declList)...)
	}

	if base.Flag.W > 1 {
		for _, n := range g.target.Decls {
			s := fmt.Sprintf("\nafter noder2 %v", n)
			ir.Dump(s, n)
		}
	}

	typecheck.DeclareUniverse()

	for _, p := range noders {
		// Process linkname and cgo pragmas.
		p.processPragmas()

		// Double check for any type-checking inconsistencies. This can be
		// removed once we're confident in IR generation results.
		syntax.Walk(p.file, func(n syntax.Node) bool {
			g.validate(n)
			return false
		})
	}

	// Create any needed stencils of generic functions
	g.stencil()

	// For now, remove all generic functions from g.target.Decl, since they
	// don't compile. TODO: We will eventually export any exportable generic
	// functions.
	j := 0
	for i, decl := range g.target.Decls {
		if decl.Op() != ir.ODCLFUNC || decl.Type().NumTParams() == 0 {
			g.target.Decls[j] = g.target.Decls[i]
			j++
		}
	}
	g.target.Decls = g.target.Decls[:j]
}

func (g *irgen) unhandled(what string, p poser) {
	base.FatalfAt(g.pos(p), "unhandled %s: %T", what, p)
	panic("unreachable")
}

// stencil scans functions for instantiated generic function calls and
// creates the required stencils for simple generic functions.
func (g *irgen) stencil() {
	g.target.Stencils = make(map[string]*ir.Func)
	for _, decl := range g.target.Decls {
		if decl.Op() != ir.ODCLFUNC {
			continue
		}
		if decl.Type().NumTParams() > 0 {
			// Skip generic functions
			continue
		}
		f := decl.(*ir.Func)
		modified := false
		ir.VisitList(f.Body, func(n ir.Node) {
			if n.Op() != ir.OCALLFUNC || n.(*ir.CallExpr).X.Op() != ir.OFUNCINST {
				return
			}
			c := n.(*ir.CallExpr)
			inst := c.X.(*ir.InstExpr)
			s := makeInstName(inst)
			//fmt.Printf("Found generic func call in %v to %v\n", f, s)
			var st *ir.Func
			var ok bool
			if st, ok = g.target.Stencils[s]; !ok {
				// If stencil doesn't exist yet, create it and add
				// to the list of decls.
				st = genericSubst(s, inst)
				g.target.Stencils[s] = st
				g.target.Decls = append(g.target.Decls, st)
				if base.Flag.W > 1 {
					ir.Dump(fmt.Sprintf("\nstenciled %v", st), st)
				}
			}
			// Replace the OFUNCINST with a direct reference to the
			// new stenciled function
			c.X = st.Nname
			modified = true
		})
		if base.Flag.W > 1 && modified {
			ir.Dump(fmt.Sprintf("\nmodified %v", decl), decl)
		}
	}

}

// makeInstName makes the unique name for a stenciled generic function, based on
// the name of the function and the types of the type params.
func makeInstName(inst *ir.InstExpr) string {
	s := inst.X.(*ir.Name).Name().Sym().Name
	for _, targ := range inst.Targs {
		s += "_" + targ.Name().Sym().Name
	}
	return s
}

type gensubst struct {
	newf    *ir.Func // Func node for the new stenciled function
	tparams *types.Fields
	targs   []ir.Node
	// The substitution map from name nodes in the generic function to the
	// name nodes in the new stenciled function.
	vars map[*ir.Name]*ir.Name
}

// generciSubst returns a new function which is a stencil (instantiation) of a
// generic function with type params, as specified by inst.
func genericSubst(instName string, inst *ir.InstExpr) *ir.Func {
	// Similar to noder.go: funcDecl
	sym := typecheck.Lookup(instName)
	name := inst.X.(*ir.Name)
	gf := name.Func
	newf := ir.NewFunc(inst.Pos())
	newf.Nname = ir.NewNameAt(inst.Pos(), sym)
	newf.Nname.Func = newf
	newf.Nname.Defn = newf

	subst := &gensubst{
		newf:    newf,
		tparams: name.Type().TParams().Fields(),
		targs:   inst.Targs,
		vars:    make(map[*ir.Name]*ir.Name),
	}

	newf.Dcl = make([]*ir.Name, len(gf.Dcl))
	for i, n := range gf.Dcl {
		newf.Dcl[i] = subst.node(n).(*ir.Name)
	}
	newf.Body = subst.list(gf.Body)

	// Ugly: we have to insert the Name nodes of the parameters/results into
	// the function type. The current function type has no Nname fields set,
	// because it came via conversion from the types2 type.
	oldt := inst.Type()
	newt := types.NewSignature(oldt.Pkg(), nil, nil, subst.fields(ir.PPARAM, oldt.Params(), newf.Dcl),
		subst.fields(ir.PPARAMOUT, oldt.Results(), newf.Dcl))

	newf.Nname.Ntype = ir.TypeNode(newt)
	newf.Nname.SetType(newt)
	ir.MarkFunc(newf.Nname)
	newf.SetTypecheck(1)
	newf.Nname.SetTypecheck(1)
	// TODO(danscales) - remove later, but avoid confusion for now.
	newf.Pragma = ir.Noinline
	return newf
}

// node is like DeepCopy(), but creates distinct ONAME nodes, and also descends
// into closures. It substitutes type arguments for type parameters in all the new
// nodes.
func (subst *gensubst) node(n ir.Node) ir.Node {
	var edit func(ir.Node) ir.Node
	edit = func(x ir.Node) ir.Node {
		switch x.Op() {
		case ir.ONAME:
			name := x.(*ir.Name)
			if v := subst.vars[name]; v != nil {
				return v
			}
			m := ir.NewNameAt(name.Pos(), name.Sym())
			t := x.Type()
			newt := subst.typ(t)
			m.SetType(newt)
			m.Curfn = subst.newf
			m.Class = name.Class
			subst.vars[name] = m
			m.SetTypecheck(1)
			return m
		case ir.OLITERAL, ir.ONIL:
			if x.Sym() != nil {
				return x
			}
		}
		m := ir.Copy(x)
		if _, isExpr := m.(ir.Expr); isExpr {
			m.SetType(subst.typ(x.Type()))
		}
		ir.EditChildren(m, edit)
		if x.Op() == ir.OCLOSURE {
			x := x.(*ir.ClosureExpr)
			// Need to save/duplicate x.Func.Nname,
			// x.Func.Nname.Ntype, x.Func.Dcl, x.Func.ClosureVars, and
			// x.Func.Body.
			oldfn := x.Func
			newfn := ir.NewFunc(oldfn.Pos())
			if oldfn.ClosureCalled() {
				newfn.SetClosureCalled(true)
			}
			m.(*ir.ClosureExpr).Func = newfn
			newfn.Nname = ir.NewNameAt(oldfn.Nname.Pos(), oldfn.Nname.Sym())
			newfn.Nname.SetType(oldfn.Nname.Type())
			newfn.Nname.Ntype = subst.node(oldfn.Nname.Ntype).(ir.Ntype)
			newfn.Body = subst.list(oldfn.Body)
			// Make shallow copy of the Dcl and ClosureVar slices
			newfn.Dcl = append([]*ir.Name(nil), oldfn.Dcl...)
			newfn.ClosureVars = append([]*ir.Name(nil), oldfn.ClosureVars...)
		}
		return m
	}
	return edit(n)
}

func (subst *gensubst) list(ll []ir.Node) []ir.Node {
	s := make([]ir.Node, len(ll))
	for i, n := range ll {
		s[i] = subst.node(n)
	}
	return s
}

// typ substitutes any type parameter found with the corresponding type argument.
func (subst *gensubst) typ(t *types.Type) *types.Type {
	for i, tp := range subst.tparams.Slice() {
		if tp.Type == t {
			return subst.targs[i].Type()
		}
	}
	switch t.Kind() {
	case types.TARRAY:
		elem := t.Elem()
		newelem := subst.typ(elem)
		if subst.typ(elem) != elem {
			return types.NewArray(newelem, t.NumElem())
		}

	case types.TPTR:
		elem := t.Elem()
		newelem := subst.typ(elem)
		if subst.typ(elem) != elem {
			return types.NewPtr(newelem)
		}

	case types.TSLICE:
		elem := t.Elem()
		newelem := subst.typ(elem)
		if subst.typ(elem) != elem {
			return types.NewSlice(newelem)
		}

	case types.TSTRUCT:
		newfields := make([]*types.Field, t.NumFields())
		change := false
		for i, f := range t.Fields().Slice() {
			t2 := subst.typ(f.Type)
			if t2 != f.Type {
				change = true
			}
			newfields[i] = types.NewField(f.Pos, f.Sym, t2)
		}
		if change {
			return types.NewStruct(t.Pkg(), newfields)
		}

		// TODO: case TFUNC
		// TODO: case TCHAN
		// TODO: case TMAP
		// TODO: case TINTER
	}
	return t
}

// fields sets the Nname field for the Field nodes inside a type signature, based
// on the corresponding in/out parameters in dcl. It depends on the in and out
// parameters being in order in dcl.
func (subst *gensubst) fields(cl ir.Class, oldt *types.Type, dcl []*ir.Name) []*types.Field {
	oldfields := oldt.FieldSlice()
	newfields := make([]*types.Field, len(oldfields))
	var i int
	for i = range dcl {
		if dcl[i].Class == cl {
			break
		}
	}
	for j := range oldfields {
		newfields[j] = oldfields[j].Copy()
		newfields[j].Type = subst.typ(oldfields[j].Type)
		newfields[j].Nname = dcl[i]
		i++
	}
	return newfields
}
