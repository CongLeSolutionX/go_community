// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"fmt"
	"go/constant"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/src"
)

// makeSrcPosBase translates from a *syntax.PosBase to a *src.PosBase.
func (p *noder2) makeSrcPosBase(b0 *syntax.PosBase) *src.PosBase {
	// fast path: most likely PosBase hasn't changed
	if p.basecache.last == b0 {
		return p.basecache.base
	}

	b1, ok := p.basemap[b0]
	if !ok {
		fn := b0.Filename()
		if b0.IsFileBase() {
			b1 = src.NewFileBase(fn, absFilename(fn))
		} else {
			// line directive base
			p0 := b0.Pos()
			p0b := p0.Base()
			if p0b == b0 {
				panic("infinite recursion in makeSrcPosBase")
			}
			p1 := src.MakePos(p.makeSrcPosBase(p0b), p0.Line(), p0.Col())
			b1 = src.NewLinePragmaBase(p1, fn, fileh(fn), b0.Line(), b0.Col())
		}
		p.basemap[b0] = b1
	}

	// update cache
	p.basecache.last = b0
	p.basecache.base = b1

	return b1
}

func (p *noder2) makeXPos(pos syntax.Pos) (_ src.XPos) {
	return Ctxt.PosTable.XPos(src.MakePos(p.makeSrcPosBase(pos.Base()), pos.Line(), pos.Col()))
}

func (p *noder2) yyerrorpos(pos syntax.Pos, format string, args ...interface{}) {
	yyerrorl(p.makeXPos(pos), format, args...)
}

// noder2 transforms package syntax's AST into a Node tree.
type noder2 struct {
	pkg  *types2.Package
	info *types2.Info

	basemap   map[*syntax.PosBase]*src.PosBase
	basecache struct {
		last *syntax.PosBase
		base *src.PosBase
	}

	file           *syntax.File
	linknames      []linkname
	pragcgobuf     [][]string
	err            chan syntax.Error
	scope          ScopeID
	importedUnsafe bool
	importedEmbed  bool

	// scopeVars is a stack tracking the number of variables declared in the
	// current function at the moment each open scope was opened.
	scopeVars []int

	lastCloseScopePos syntax.Pos
}

func (p *noder2) funcBody(fn *Node, block *syntax.BlockStmt) {
	oldScope := p.scope
	p.scope = 0
	funchdr(fn)

	if block != nil {
		body := p.stmts(block.List)
		if body == nil {
			body = []*Node{nod(OEMPTY, nil, nil)}
		}
		fn.Nbody.Set(body)

		lineno = p.makeXPos(block.Rbrace)
		fn.Func.Endlineno = lineno
	}

	funcbody()
	p.scope = oldScope
}

func (p *noder2) openScope(pos syntax.Pos) {
	types.Markdcl()

	if trackScopes {
		Curfn.Func.Parents = append(Curfn.Func.Parents, p.scope)
		p.scopeVars = append(p.scopeVars, len(Curfn.Func.Dcl))
		p.scope = ScopeID(len(Curfn.Func.Parents))

		p.markScope(pos)
	}
}

func (p *noder2) closeScope(pos syntax.Pos) {
	p.lastCloseScopePos = pos
	types.Popdcl()

	if trackScopes {
		scopeVars := p.scopeVars[len(p.scopeVars)-1]
		p.scopeVars = p.scopeVars[:len(p.scopeVars)-1]
		if scopeVars == len(Curfn.Func.Dcl) {
			// no variables were declared in this scope, so we can retract it.

			if int(p.scope) != len(Curfn.Func.Parents) {
				Fatalf("scope tracking inconsistency, no variables declared but scopes were not retracted")
			}

			p.scope = Curfn.Func.Parents[p.scope-1]
			Curfn.Func.Parents = Curfn.Func.Parents[:len(Curfn.Func.Parents)-1]

			nmarks := len(Curfn.Func.Marks)
			Curfn.Func.Marks[nmarks-1].Scope = p.scope
			prevScope := ScopeID(0)
			if nmarks >= 2 {
				prevScope = Curfn.Func.Marks[nmarks-2].Scope
			}
			if Curfn.Func.Marks[nmarks-1].Scope == prevScope {
				Curfn.Func.Marks = Curfn.Func.Marks[:nmarks-1]
			}
			return
		}

		p.scope = Curfn.Func.Parents[p.scope-1]

		p.markScope(pos)
	}
}

func (p *noder2) markScope(pos syntax.Pos) {
	xpos := p.makeXPos(pos)
	if i := len(Curfn.Func.Marks); i > 0 && Curfn.Func.Marks[i-1].Pos == xpos {
		Curfn.Func.Marks[i-1].Scope = p.scope
	} else {
		Curfn.Func.Marks = append(Curfn.Func.Marks, Mark{xpos, p.scope})
	}
}

// closeAnotherScope is like closeScope, but it reuses the same mark
// position as the last closeScope call. This is useful for "for" and
// "if" statements, as their implicit blocks always end at the same
// position as an explicit block.
func (p *noder2) closeAnotherScope() {
	p.closeScope(p.lastCloseScopePos)
}

func (p *noder2) node0() {
	types.Block = 1

	p.setlineno(p.file.PkgName)
	mkpackage(p.file.PkgName.Value)

	for _, decl := range p.file.DeclList {
		switch decl := decl.(type) {
		case *syntax.ImportDecl:
			if !decl.Path.Bad {
				val := p.basicLit(decl.Path)
				ipkg := importfile(&val)
				if ipkg == nil {
					Fatalf("wah: %v failed to import", decl.Path)
				}
			}
		case *syntax.TypeDecl:
			if decl.Alias {
				continue
			}
			n := p.declName(decl.Name)
			setTypeNode(n, types.New(TFORW))
			n.Type.Sym = n.Sym
			declare(n, dclcontext)
		}
	}
}

func (p *noder2) node1() {
	p.setlineno(p.file.PkgName)

	for _, decl := range p.file.DeclList {
		switch decl := decl.(type) {
		case *syntax.TypeDecl:
			xtop = append(xtop, p.typeDecl(decl, true))
		}
	}
}

func (p *noder2) node() {
	types.Block = 1
	p.importedUnsafe = false
	p.importedEmbed = false

	p.setlineno(p.file.PkgName)
	mkpackage(p.file.PkgName.Value)

	if pragma, ok := p.file.Pragma.(*Pragma); ok {
		pragma.Flag &^= GoBuildPragma
		p.checkUnused(pragma)
	}

	xtop = append(xtop, p.decls(p.file.DeclList, true)...)

	for _, n := range p.linknames {
		if !p.importedUnsafe {
			p.yyerrorpos(n.pos, "//go:linkname only allowed in Go files that import \"unsafe\"")
			continue
		}
		s := lookup(n.local)
		if n.remote != "" {
			s.Linkname = n.remote
		} else {
			// Use the default object symbol name if the
			// user didn't provide one.
			if myimportpath == "" {
				p.yyerrorpos(n.pos, "//go:linkname requires linkname argument or -p compiler flag")
			} else {
				s.Linkname = objabi.PathToPrefix(myimportpath) + "." + n.local
			}
		}
	}

	// The linker expects an ABI0 wrapper for all cgo-exported
	// functions.
	for _, prag := range p.pragcgobuf {
		switch prag[0] {
		case "cgo_export_static", "cgo_export_dynamic":
			if symabiRefs == nil {
				symabiRefs = make(map[string]obj.ABI)
			}
			symabiRefs[prag[1]] = obj.ABI0
		}
	}

	pragcgobuf = append(pragcgobuf, p.pragcgobuf...)
	lineno = src.NoXPos
	clearImports()
}

func (p *noder2) decls(decls []syntax.Decl, top bool) (l []*Node) {
	var cs constState

	for _, decl := range decls {
		p.setlineno(decl)
		switch decl := decl.(type) {
		case *syntax.ImportDecl:
			p.importDecl(decl)

		case *syntax.VarDecl:
			l = append(l, p.varDecl(decl)...)

		case *syntax.ConstDecl:
			l = append(l, p.constDecl(decl, &cs)...)

		case *syntax.TypeDecl:
			if !top {
				l = append(l, p.typeDecl(decl, top))
			}

		case *syntax.FuncDecl:
			l = append(l, p.funcDecl(decl))

		default:
			panic("unhandled Decl")
		}
	}

	return
}

func (p *noder2) importDecl(imp *syntax.ImportDecl) {
	if imp.Path.Bad {
		return // avoid follow-on errors if there was a syntax error
	}

	if pragma, ok := imp.Pragma.(*Pragma); ok {
		p.checkUnused(pragma)
	}

	val := p.basicLit(imp.Path)
	ipkg := importfile(&val)
	if ipkg == nil {
		if nerrors == 0 {
			Fatalf("phase error in import")
		}
		return
	}

	if ipkg == unsafepkg {
		p.importedUnsafe = true
	}
	if ipkg.Path == "embed" {
		p.importedEmbed = true
	}

	ipkg.Direct = true

	var my *types.Sym
	if imp.LocalPkgName != nil {
		my = p.name(imp.LocalPkgName)
	} else {
		my = lookup(ipkg.Name)
	}

	pack := p.nod(imp, OPACK, nil, nil)
	pack.Sym = my
	pack.Name.Pkg = ipkg

	switch my.Name {
	case ".":
		importdot(ipkg, pack)
		return
	case "init":
		yyerrorl(pack.Pos, "cannot import package as init - init must be a func")
		return
	case "_":
		return
	}
	if my.Def != nil {
		redeclare(pack.Pos, my, "as imported package name")
	}
	my.Def = asTypesNode(pack)
	my.Lastlineno = pack.Pos
	my.Block = 1 // at top level
}

func (p *noder2) varDecl(decl *syntax.VarDecl) []*Node {
	names := p.declNames(decl.NameList)
	typ := p.typeExprOrNil(decl.Type)

	var exprs []*Node
	if decl.Values != nil {
		exprs = p.exprList(decl.Values)
	}

	if pragma, ok := decl.Pragma.(*Pragma); ok {
		if len(pragma.Embeds) > 0 {
			if !p.importedEmbed {
				// This check can't be done when building the list pragma.Embeds
				// because that list is created before the noder starts walking over the file,
				// so at that point it hasn't seen the imports.
				// We're left to check now, just before applying the //go:embed lines.
				for _, e := range pragma.Embeds {
					p.yyerrorpos(e.Pos, "//go:embed only allowed in Go files that import \"embed\"")
				}
			} else {
				exprs = varEmbed2(p, names, typ, exprs, pragma.Embeds)
			}
			pragma.Embeds = nil
		}
		p.checkUnused(pragma)
	}

	p.setlineno(decl)
	return variter(names, typ, exprs)
}

func (p *noder2) constDecl(decl *syntax.ConstDecl, cs *constState) []*Node {
	if decl.Group == nil || decl.Group != cs.group {
		*cs = constState{
			group: decl.Group,
		}
	}

	if pragma, ok := decl.Pragma.(*Pragma); ok {
		p.checkUnused(pragma)
	}

	names := p.declNames(decl.NameList)
	typ := p.typeExprOrNil(decl.Type)

	var values []*Node
	if decl.Values != nil {
		values = p.exprList(decl.Values)
		cs.typ, cs.values = typ, values
	} else {
		if typ != nil {
			yyerror("const declaration cannot have type without expression")
		}
		typ, values = cs.typ, cs.values
	}

	nn := make([]*Node, 0, len(names))
	for i, n := range names {
		if i >= len(values) {
			yyerror("missing value in const declaration")
			break
		}

		if c, isConst := p.info.Defs[decl.NameList[i]].(*types2.Const); isConst {
			typ := c.Type()
			b, isBasic := typ.Underlying().(*types2.Basic)
			if !isBasic {
				Fatalf("WELL I WAS WRONG: %v, %v", c, b)
			}

			n.Op = OLITERAL
			n.Type = p.typeExpr2(typ)

			floatVal := func(v interface{}) *Mpflt {
				f := newMpflt()
				switch v := v.(type) {
				case *big.Float:
					f.Val.Set(v)
					// fmt.Printf("FloatSet: %v -> %v\n", v, f.Val)
				case *big.Rat:
					f.Val.SetRat(v)
					// fmt.Printf("SetRat: %v -> %v\n", v, f.Val)
				default:
					Fatalf("unexpected type: %v", v)
				}
				return f
			}

			var val Val
			if v := c.Val(); v.Kind() == constant.Complex {
				var c Mpcplx
				c.Real = *floatVal(constant.Real(v))
				c.Imag = *floatVal(constant.Imag(v))
				val.U = &c
			} else {
				switch v := constant.Val(v).(type) {
				case bool, string:
					val.U = v
				case int64:
					var i Mpint
					i.SetInt64(v)
					i.Rune = b.Kind() == types2.UntypedRune
					val.U = &i
				case *big.Int:
					var i Mpint
					i.Val.Set(v)
					i.Rune = b.Kind() == types2.UntypedRune
					val.U = &i
				default:
					val.U = floatVal(v)
				}
			}

			if n.Type.Etype == TIDEAL { // untyped number
				switch n.Type {
				case types.UntypedInt, types.UntypedRune:
					_ = val.U.(*Mpint) // assert type
				case types.UntypedFloat:
					val = toflt(val)
				case types.UntypedComplex:
					val = tocplx(val)
				default:
					Fatalf("unexpected ideal type: %v", n.Type)
				}
			} else {
				val = convertVal(val, n.Type, false)
			}

			n.SetVal(val)
			declare(n, dclcontext)

			nn = append(nn, p.nod(decl, ODCLCONST, n, nil))

			continue
		}

		Fatalf("doesn't happen")

		v := values[i]
		if decl.Values == nil {
			v = treecopy(v, n.Pos)
		}

		n.Op = OLITERAL
		declare(n, dclcontext)

		n.Name.Param.Ntype = typ
		n.Name.Defn = v
		n.SetIota(cs.iota)

		nn = append(nn, p.nod(decl, ODCLCONST, n, nil))
	}

	if len(values) > len(names) {
		yyerror("extra expression in const declaration")
	}

	cs.iota++

	return nn
}

func (p *noder2) typeExpr2(typ types2.Type) *types.Type {
	var TODO *types.Pkg

	fieldOf := func(v *types2.Var) *types.Field {
		if pkg := v.Pkg(); pkg != p.pkg {
			Fatalf("weird package: %v != %v", pkg, p.pkg)
		}

		vt := p.typeExpr2(v.Type())
		if vt == nil {
			return nil // TODO: Remove once impossible
		}

		f := types.NewField()
		f.Pos = p.pos(v)
		if name := v.Name(); name != "" {
			pkg := localpkg
			if v.Pkg() != p.pkg && !types.IsExported(name) {
				// pkg = types.NewPkg(v.Pkg().Path(), "")
				fmt.Printf("Did I need to use %q??\n", v.Pkg().Path())
			}
			f.Sym = pkg.Lookup(name)
		}
		f.Type = vt
		return f
	}

	switch typ := typ.(type) {
	case *types2.Named:
		switch obj := typ.Obj(); obj.Pkg() {
		case nil: /* universe */
			if obj.Name() == "error" {
				return types.Errortype
			}
			fmt.Println("TODO: universal named type: %v", obj)
			return nil

		case p.pkg: /* current package */
			def := oldname(lookup(obj.Name()))
			if def.Op != OTYPE {
				Fatalf("definition for %v is not a type: %v (%v)", obj, def, def.Op)
			}
			return def.Type

		default:
			pkg := types.NewPkg(obj.Pkg().Path(), "")
			def := asNode(pkg.Lookup(obj.Name()).Def)
			if def == nil {
				Fatalf("missing definition for %v", obj)
			}
			def = resolve(def)
			if def.Op != OTYPE {
				Fatalf("definition for %v is not a type: %v (%v)", obj, def, def.Op)
			}
			if got := asNode(def.Type.Nod); got != def {
				fmt.Printf("WAH WAH: %v (%p) != %v (%p)", got, got, def, def)
			}
			return def.Type
		}
	case *types2.Basic:
		if k := typ.Kind(); uint64(k) < uint64(len(etypes)) {
			if et := etypes[k]; et != 0 {
				return types.Types[et]
			}
		}
		switch typ.Kind() {
		case types2.UntypedInt:
			return types.UntypedInt
		case types2.UntypedRune:
			return types.UntypedRune
		case types2.UntypedFloat:
			return types.UntypedFloat
		case types2.UntypedComplex:
			return types.UntypedComplex
		case types2.UntypedBool:
			return types.UntypedBool
		case types2.UntypedString:
			return types.UntypedString
		}
		Fatalf("DIDN'T HANDLE THIS BASIC TYPE: %v\n", typ)
	case *types2.Array:
		if elem := p.typeExpr2(typ.Elem()); elem != nil {
			return types.NewArray(elem, typ.Len())
		}
	case *types2.Chan:
		if elem := p.typeExpr2(typ.Elem()); elem != nil {
			var dir types.ChanDir
			switch dir2 := typ.Dir(); dir2 {
			case types2.SendRecv:
				dir = types.Cboth
			case types2.SendOnly:
				dir = types.Csend
			case types2.RecvOnly:
				dir = types.Crecv
			default:
				Fatalf("unexpected dir2: %v", dir2)
			}
			return types.NewChan(elem, dir)
		}
	case *types2.Signature:
		var recv *types.Field
		if r := typ.Recv(); r != nil {
			switch r.Type().Underlying().(type) {
			case *types2.Interface: // N.B., can also be a named interface
				recv = fakeRecvField()
			default:
				Fatalf("STILL GOTTA DO THIS TOO: %v", r)
			}
		}

		fieldsOf := func(tup *types2.Tuple) []*types.Field {
			fields := make([]*types.Field, tup.Len())
			for i := range fields {
				fields[i] = fieldOf(tup.At(i))
				if fields[i] == nil {
					return nil
				}
			}
			return fields
		}

		if in := fieldsOf(typ.Params()); in != nil {
			if out := fieldsOf(typ.Results()); out != nil {
				if typ.Variadic() {
					in[len(in)-1].SetIsDDD(true)
				}
				t := functypefield(recv, in, out)
				t.SetPkg(TODO)
				return t
			}
		}
	case *types2.Map:
		if key := p.typeExpr2(typ.Key()); key != nil {
			if elem := p.typeExpr2(typ.Elem()); elem != nil {
				return types.NewMap(key, elem)
			}
		}
	case *types2.Pointer:
		if elem := p.typeExpr2(typ.Elem()); elem != nil {
			return types.NewPtr(elem)
		}
	case *types2.Slice:
		if elem := p.typeExpr2(typ.Elem()); elem != nil {
			return types.NewSlice(elem)
		}
	case *types2.Struct:
		t := types.New(TSTRUCT)

		fields := make([]*types.Field, typ.NumFields())
		for i := range fields {
			v, tag := typ.Field(i), typ.Tag(i)
			f := fieldOf(v)
			if f == nil {
				return nil // TODO: Remove once impossible
			}

			f.Note = tag
			if v.Embedded() {
				f.Embedded = 1
			}

			fields[i] = f
		}
		t.SetPkg(TODO)
		t.SetFields(fields)
		checkwidth(t)
		return t

	case *types2.Interface:
		embeddeds := make([]*types.Field, typ.NumEmbeddeds())
		for i := range embeddeds {
			ft := p.typeExpr2(typ.EmbeddedType(i))
			if ft == nil {
				return nil
			}

			f := types.NewField()
			// TODO(mdempsky): Set f.Pos
			f.Type = ft
			embeddeds[i] = f
		}

		methods := make([]*types.Field, typ.NumExplicitMethods())
		for i := range methods {
			fun := typ.ExplicitMethod(i)

			pos := p.pos(fun)
			sym := lookup(fun.Name())
			ft := p.typeExpr2(fun.Type())
			if ft == nil {
				return nil
			}

			f := types.NewField()
			f.Pos = pos
			f.Sym = sym
			f.Type = ft
			methods[i] = f
		}

		t := types.New(TINTER)
		t.SetPkg(TODO)
		t.SetInterface(append(embeddeds, methods...))

		// Ensure we expand the interface in the frontend (#25055).
		checkwidth(t)
		return t

	default:
		Fatalf("ANOTHER TYPE TO WORRY ABOUT: %T, %v\n", typ, typ)
	}
	return nil
}

var etypes = [...]types.EType{
	types2.Bool:          TBOOL,
	types2.Int:           TINT,
	types2.Int8:          TINT8,
	types2.Int16:         TINT16,
	types2.Int32:         TINT32,
	types2.Int64:         TINT64,
	types2.Uint:          TUINT,
	types2.Uint8:         TUINT8,
	types2.Uint16:        TUINT16,
	types2.Uint32:        TUINT32,
	types2.Uint64:        TUINT64,
	types2.Uintptr:       TUINTPTR,
	types2.Float32:       TFLOAT32,
	types2.Float64:       TFLOAT64,
	types2.Complex64:     TCOMPLEX64,
	types2.Complex128:    TCOMPLEX128,
	types2.String:        TSTRING,
	types2.UnsafePointer: TUNSAFEPTR,
}

func (p *noder2) typeDecl(decl *syntax.TypeDecl, top bool) *Node {
	lineno = p.pos(decl)

	var n *Node

	if top != (dclcontext == PEXTERN) {
		Fatalf("weird mismatch: %v != %v", top, dclcontext)
	}

	if !top || decl.Alias {
		n = p.declName(decl.Name)
		n.Op = OTYPE
		if !decl.Alias {
			setTypeNode(n, types.New(TFORW))
			n.Type.Sym = n.Sym
		}
		declare(n, dclcontext)
	} else {
		n = oldname(p.name(decl.Name))
		if n.Op != OTYPE || n.Type.Etype != TFORW {
			Fatalf("weird!! %v, %v", n.Op, n.Type)
		}
	}
	// fmt.Println("typeDecl", n, n.Op, top, decl.Alias, dclcontext, decl.Name, asNode(n.Sym.Def))
	// fmt.Printf("typeDecl conntinued: %p, %p; %v, %p\n", n, asNode(n.Sym.Def), n.Sym, n.Sym)

	param := n.Name.Param

	name := p.info.Defs[decl.Name].(*types2.TypeName)
	if decl.Alias != name.IsAlias() {
		Fatalf("wah, %v != %v", decl.Alias, name.IsAlias())
	}

	typ := name.Type()
	if !name.IsAlias() {
		typ = typ.Underlying()
	}

	typ2 := p.typeExpr2(typ)
	if typ2 == nil {
		Fatalf("typeExpr2 failed for %v", typ)
	}

	param.SetAlias(decl.Alias)
	if pragma, ok := decl.Pragma.(*Pragma); ok {
		if !decl.Alias {
			param.SetPragma(pragma.Flag & TypePragmas)
			pragma.Flag &^= TypePragmas
		}
		p.checkUnused(pragma)
	}

	if decl.Alias {
		n.Type = typ2
		fmt.Printf("typ2.Nod was %v (%p)\n", typ2.Nod, typ2.Nod)
		n2 := typenodl(p.pos(decl.Type), typ2)
		fmt.Printf("for alias: reassigning n.Sym.Def for %v (%p) = %v (%p)\n", n.Sym, n.Sym, n2, n2)
		n.Sym.Def = asTypesNode(n2)
	} else {
		n.SetTypecheck(1)
		setUnderlying(n.Type, typ2)
	}

	nod := p.nod(decl, ODCLTYPE, n, nil)
	if param.Alias() && !langSupported(1, 9, localpkg) {
		yyerrorl(nod.Pos, "type aliases only supported as of -lang=go1.9")
	}
	return nod
}

func (p *noder2) declNames(names []*syntax.Name) []*Node {
	nodes := make([]*Node, 0, len(names))
	for _, name := range names {
		nodes = append(nodes, p.declName(name))
	}
	return nodes
}

func (p *noder2) declName(name *syntax.Name) *Node {
	n := dclname(p.name(name))
	n.Pos = p.pos(name)
	return n
}

func (p *noder2) funcDecl(fun *syntax.FuncDecl) *Node {
	name := p.name(fun.Name)
	t := p.signature(fun.Recv, fun.Type)
	f := p.nod(fun, ODCLFUNC, nil, nil)

	if fun.Recv == nil {
		if name.Name == "init" {
			name = renameinit()
			if t.List.Len() > 0 || t.Rlist.Len() > 0 {
				yyerrorl(f.Pos, "func init must have no arguments and no return values")
			}
		}

		if localpkg.Name == "main" && name.Name == "main" {
			if t.List.Len() > 0 || t.Rlist.Len() > 0 {
				yyerrorl(f.Pos, "func main must have no arguments and no return values")
			}
		}
	} else {
		f.Func.Shortname = name
		name = nblank.Sym // filled in by typecheckfunc
	}

	f.Func.Nname = newfuncnamel(p.pos(fun.Name), name)
	f.Func.Nname.Name.Defn = f
	f.Func.Nname.Name.Param.Ntype = t

	if pragma, ok := fun.Pragma.(*Pragma); ok {
		f.Func.Pragma = pragma.Flag & FuncPragmas
		if pragma.Flag&Systemstack != 0 && pragma.Flag&Nosplit != 0 {
			yyerrorl(f.Pos, "go:nosplit and go:systemstack cannot be combined")
		}
		pragma.Flag &^= FuncPragmas
		p.checkUnused(pragma)
	}

	if fun.Recv == nil {
		declare(f.Func.Nname, PFUNC)
	}

	p.funcBody(f, fun.Body)

	if fun.Body != nil {
		if f.Func.Pragma&Noescape != 0 {
			yyerrorl(f.Pos, "can only use //go:noescape with external func implementations")
		}
	} else {
		if pure_go || strings.HasPrefix(f.funcname(), "init.") {
			// Linknamed functions are allowed to have no body. Hopefully
			// the linkname target has a body. See issue 23311.
			isLinknamed := false
			for _, n := range p.linknames {
				if f.funcname() == n.local {
					isLinknamed = true
					break
				}
			}
			if !isLinknamed {
				yyerrorl(f.Pos, "missing function body")
			}
		}
	}

	return f
}

func (p *noder2) signature(recv *syntax.Field, typ *syntax.FuncType) *Node {
	n := p.nod(typ, OTFUNC, nil, nil)
	if recv != nil {
		n.Left = p.param(recv, false, false)
	}
	n.List.Set(p.params(typ.ParamList, true))
	n.Rlist.Set(p.params(typ.ResultList, false))
	return n
}

func (p *noder2) params(params []*syntax.Field, dddOk bool) []*Node {
	nodes := make([]*Node, 0, len(params))
	for i, param := range params {
		p.setlineno(param)
		nodes = append(nodes, p.param(param, dddOk, i+1 == len(params)))
	}
	return nodes
}

func (p *noder2) param(param *syntax.Field, dddOk, final bool) *Node {
	var name *types.Sym
	if param.Name != nil {
		name = p.name(param.Name)
	}

	typ := p.typeExpr(param.Type)
	n := p.nodSym(param, ODCLFIELD, typ, name)

	// rewrite ...T parameter
	if typ.Op == ODDD {
		if !dddOk {
			// We mark these as syntax errors to get automatic elimination
			// of multiple such errors per line (see yyerrorl in subr.go).
			yyerror("syntax error: cannot use ... in receiver or result parameter list")
		} else if !final {
			if param.Name == nil {
				yyerror("syntax error: cannot use ... with non-final parameter")
			} else {
				p.yyerrorpos(param.Name.Pos(), "syntax error: cannot use ... with non-final parameter %s", param.Name.Value)
			}
		}
		typ.Op = OTARRAY
		typ.Right = typ.Left
		typ.Left = nil
		n.SetIsDDD(true)
		if n.Left != nil {
			n.Left.SetIsDDD(true)
		}
	}

	return n
}

func (p *noder2) exprList(expr syntax.Expr) []*Node {
	if list, ok := expr.(*syntax.ListExpr); ok {
		return p.exprs(list.ElemList)
	}
	return []*Node{p.expr(expr)}
}

func (p *noder2) exprs(exprs []syntax.Expr) []*Node {
	nodes := make([]*Node, 0, len(exprs))
	for _, expr := range exprs {
		nodes = append(nodes, p.expr(expr))
	}
	return nodes
}

func (p *noder2) expr(expr syntax.Expr) *Node {
	p.setlineno(expr)
	switch expr := expr.(type) {
	case nil, *syntax.BadExpr:
		return nil
	case *syntax.Name:
		return p.mkname(expr)
	case *syntax.BasicLit:
		n := nodlit(p.basicLit(expr))
		n.SetDiag(expr.Bad) // avoid follow-on errors if there was a syntax error
		return n
	case *syntax.CompositeLit:
		n := p.nod(expr, OCOMPLIT, nil, nil)
		if expr.Type != nil {
			n.Right = p.expr(expr.Type)
		}
		l := p.exprs(expr.ElemList)
		for i, e := range l {
			l[i] = p.wrapname(expr.ElemList[i], e)
		}
		n.List.Set(l)
		lineno = p.makeXPos(expr.Rbrace)
		return n
	case *syntax.KeyValueExpr:
		// use position of expr.Key rather than of expr (which has position of ':')
		return p.nod(expr.Key, OKEY, p.expr(expr.Key), p.wrapname(expr.Value, p.expr(expr.Value)))
	case *syntax.FuncLit:
		return p.funcLit(expr)
	case *syntax.ParenExpr:
		return p.nod(expr, OPAREN, p.expr(expr.X), nil)
	case *syntax.SelectorExpr:
		// parser.new_dotname
		obj := p.expr(expr.X)
		if obj.Op == OPACK {
			obj.Name.SetUsed(true)
			return importName(obj.Name.Pkg.Lookup(expr.Sel.Value))
		}
		n := nodSym(OXDOT, obj, p.name(expr.Sel))
		n.Pos = p.pos(expr) // lineno may have been changed by p.expr(expr.X)
		return n
	case *syntax.IndexExpr:
		return p.nod(expr, OINDEX, p.expr(expr.X), p.expr(expr.Index))
	case *syntax.SliceExpr:
		op := OSLICE
		if expr.Full {
			op = OSLICE3
		}
		n := p.nod(expr, op, p.expr(expr.X), nil)
		var index [3]*Node
		for i, x := range &expr.Index {
			if x != nil {
				index[i] = p.expr(x)
			}
		}
		n.SetSliceBounds(index[0], index[1], index[2])
		return n
	case *syntax.AssertExpr:
		return p.nod(expr, ODOTTYPE, p.expr(expr.X), p.typeExpr(expr.Type))
	case *syntax.Operation:
		if expr.Op == syntax.Add && expr.Y != nil {
			return p.sum(expr)
		}
		x := p.expr(expr.X)
		if expr.Y == nil {
			return p.nod(expr, p.unOp(expr.Op), x, nil)
		}
		return p.nod(expr, p.binOp(expr.Op), x, p.expr(expr.Y))
	case *syntax.CallExpr:
		n := p.nod(expr, OCALL, p.expr(expr.Fun), nil)
		n.List.Set(p.exprs(expr.ArgList))
		n.SetIsDDD(expr.HasDots)
		return n

	case *syntax.ArrayType:
		var len *Node
		if expr.Len != nil {
			len = p.expr(expr.Len)
		} else {
			len = p.nod(expr, ODDD, nil, nil)
		}
		return p.nod(expr, OTARRAY, len, p.typeExpr(expr.Elem))
	case *syntax.SliceType:
		return p.nod(expr, OTARRAY, nil, p.typeExpr(expr.Elem))
	case *syntax.DotsType:
		return p.nod(expr, ODDD, p.typeExpr(expr.Elem), nil)
	case *syntax.StructType:
		return p.structType(expr)
	case *syntax.InterfaceType:
		return p.interfaceType(expr)
	case *syntax.FuncType:
		return p.signature(nil, expr)
	case *syntax.MapType:
		return p.nod(expr, OTMAP, p.typeExpr(expr.Key), p.typeExpr(expr.Value))
	case *syntax.ChanType:
		n := p.nod(expr, OTCHAN, p.typeExpr(expr.Elem), nil)
		n.SetTChanDir(p.chanDir(expr.Dir))
		return n

	case *syntax.TypeSwitchGuard:
		n := p.nod(expr, OTYPESW, nil, p.expr(expr.X))
		if expr.Lhs != nil {
			n.Left = p.declName(expr.Lhs)
			if n.Left.isBlank() {
				yyerror("invalid variable name %v in type switch", n.Left)
			}
		}
		return n
	}
	panic("unhandled Expr")
}

// sum efficiently handles very large summation expressions (such as
// in issue #16394). In particular, it avoids left recursion and
// collapses string literals.
func (p *noder2) sum(x syntax.Expr) *Node {
	// While we need to handle long sums with asymptotic
	// efficiency, the vast majority of sums are very small: ~95%
	// have only 2 or 3 operands, and ~99% of string literals are
	// never concatenated.

	adds := make([]*syntax.Operation, 0, 2)
	for {
		add, ok := x.(*syntax.Operation)
		if !ok || add.Op != syntax.Add || add.Y == nil {
			break
		}
		adds = append(adds, add)
		x = add.X
	}

	// nstr is the current rightmost string literal in the
	// summation (if any), and chunks holds its accumulated
	// substrings.
	//
	// Consider the expression x + "a" + "b" + "c" + y. When we
	// reach the string literal "a", we assign nstr to point to
	// its corresponding Node and initialize chunks to {"a"}.
	// Visiting the subsequent string literals "b" and "c", we
	// simply append their values to chunks. Finally, when we
	// reach the non-constant operand y, we'll join chunks to form
	// "abc" and reassign the "a" string literal's value.
	//
	// N.B., we need to be careful about named string constants
	// (indicated by Sym != nil) because 1) we can't modify their
	// value, as doing so would affect other uses of the string
	// constant, and 2) they may have types, which we need to
	// handle correctly. For now, we avoid these problems by
	// treating named string constants the same as non-constant
	// operands.
	var nstr *Node
	chunks := make([]string, 0, 1)

	n := p.expr(x)
	if Isconst(n, CTSTR) && n.Sym == nil {
		nstr = n
		chunks = append(chunks, nstr.StringVal())
	}

	for i := len(adds) - 1; i >= 0; i-- {
		add := adds[i]

		r := p.expr(add.Y)
		if Isconst(r, CTSTR) && r.Sym == nil {
			if nstr != nil {
				// Collapse r into nstr instead of adding to n.
				chunks = append(chunks, r.StringVal())
				continue
			}

			nstr = r
			chunks = append(chunks, nstr.StringVal())
		} else {
			if len(chunks) > 1 {
				nstr.SetVal(Val{U: strings.Join(chunks, "")})
			}
			nstr = nil
			chunks = chunks[:0]
		}
		n = p.nod(add, OADD, n, r)
	}
	if len(chunks) > 1 {
		nstr.SetVal(Val{U: strings.Join(chunks, "")})
	}

	return n
}

func (p *noder2) typeExpr(typ syntax.Expr) *Node {
	// TODO(mdempsky): Be stricter? typecheck should handle errors anyway.
	return p.expr(typ)
}

func (p *noder2) typeExprOrNil(typ syntax.Expr) *Node {
	if typ != nil {
		return p.expr(typ)
	}
	return nil
}

func (p *noder2) chanDir(dir syntax.ChanDir) types.ChanDir {
	switch dir {
	case 0:
		return types.Cboth
	case syntax.SendOnly:
		return types.Csend
	case syntax.RecvOnly:
		return types.Crecv
	}
	panic("unhandled ChanDir")
}

func (p *noder2) structType(expr *syntax.StructType) *Node {
	l := make([]*Node, 0, len(expr.FieldList))
	for i, field := range expr.FieldList {
		p.setlineno(field)
		var n *Node
		if field.Name == nil {
			n = p.embedded(field.Type)
		} else {
			n = p.nodSym(field, ODCLFIELD, p.typeExpr(field.Type), p.name(field.Name))
		}
		if i < len(expr.TagList) && expr.TagList[i] != nil {
			n.SetVal(p.basicLit(expr.TagList[i]))
		}
		l = append(l, n)
	}

	p.setlineno(expr)
	n := p.nod(expr, OTSTRUCT, nil, nil)
	n.List.Set(l)
	return n
}

func (p *noder2) interfaceType(expr *syntax.InterfaceType) *Node {
	l := make([]*Node, 0, len(expr.MethodList))
	for _, method := range expr.MethodList {
		p.setlineno(method)
		var n *Node
		if method.Name == nil {
			n = p.nodSym(method, ODCLFIELD, importName(p.packname(method.Type)), nil)
		} else {
			mname := p.name(method.Name)
			sig := p.typeExpr(method.Type)
			sig.Left = fakeRecv()
			n = p.nodSym(method, ODCLFIELD, sig, mname)
			ifacedcl(n)
		}
		l = append(l, n)
	}

	n := p.nod(expr, OTINTER, nil, nil)
	n.List.Set(l)
	return n
}

func (p *noder2) packname(expr syntax.Expr) *types.Sym {
	switch expr := expr.(type) {
	case *syntax.Name:
		name := p.name(expr)
		if n := oldname(name); n.Name != nil && n.Name.Pack != nil {
			n.Name.Pack.Name.SetUsed(true)
		}
		return name
	case *syntax.SelectorExpr:
		name := p.name(expr.X.(*syntax.Name))
		def := asNode(name.Def)
		if def == nil {
			yyerror("undefined: %v", name)
			return name
		}
		var pkg *types.Pkg
		if def.Op != OPACK {
			yyerror("%v is not a package", name)
			pkg = localpkg
		} else {
			def.Name.SetUsed(true)
			pkg = def.Name.Pkg
		}
		return pkg.Lookup(expr.Sel.Value)
	}
	panic(fmt.Sprintf("unexpected packname: %#v", expr))
}

func (p *noder2) embedded(typ syntax.Expr) *Node {
	op, isStar := typ.(*syntax.Operation)
	if isStar {
		if op.Op != syntax.Mul || op.Y != nil {
			panic("unexpected Operation")
		}
		typ = op.X
	}

	sym := p.packname(typ)
	n := p.nodSym(typ, ODCLFIELD, importName(sym), lookup(sym.Name))
	n.SetEmbedded(true)

	if isStar {
		n.Left = p.nod(op, ODEREF, n.Left, nil)
	}
	return n
}

func (p *noder2) stmts(stmts []syntax.Stmt) []*Node {
	return p.stmtsFall(stmts, false)
}

func (p *noder2) stmtsFall(stmts []syntax.Stmt, fallOK bool) []*Node {
	var nodes []*Node
	for i, stmt := range stmts {
		s := p.stmtFall(stmt, fallOK && i+1 == len(stmts))
		if s == nil {
		} else if s.Op == OBLOCK && s.Ninit.Len() == 0 {
			nodes = append(nodes, s.List.Slice()...)
		} else {
			nodes = append(nodes, s)
		}
	}
	return nodes
}

func (p *noder2) stmt(stmt syntax.Stmt) *Node {
	return p.stmtFall(stmt, false)
}

func (p *noder2) stmtFall(stmt syntax.Stmt, fallOK bool) *Node {
	p.setlineno(stmt)
	switch stmt := stmt.(type) {
	case *syntax.EmptyStmt:
		return nil
	case *syntax.LabeledStmt:
		return p.labeledStmt(stmt, fallOK)
	case *syntax.BlockStmt:
		l := p.blockStmt(stmt)
		if len(l) == 0 {
			// TODO(mdempsky): Line number?
			return nod(OEMPTY, nil, nil)
		}
		return liststmt(l)
	case *syntax.ExprStmt:
		return p.wrapname(stmt, p.expr(stmt.X))
	case *syntax.SendStmt:
		return p.nod(stmt, OSEND, p.expr(stmt.Chan), p.expr(stmt.Value))
	case *syntax.DeclStmt:
		return liststmt(p.decls(stmt.DeclList, false))
	case *syntax.AssignStmt:
		if stmt.Op != 0 && stmt.Op != syntax.Def {
			n := p.nod(stmt, OASOP, p.expr(stmt.Lhs), p.expr(stmt.Rhs))
			n.SetImplicit(stmt.Rhs == syntax.ImplicitOne)
			n.SetSubOp(p.binOp(stmt.Op))
			return n
		}

		n := p.nod(stmt, OAS, nil, nil) // assume common case

		rhs := p.exprList(stmt.Rhs)
		lhs := p.assignList(stmt.Lhs, n, stmt.Op == syntax.Def)

		if len(lhs) == 1 && len(rhs) == 1 {
			// common case
			n.Left = lhs[0]
			n.Right = rhs[0]
		} else {
			n.Op = OAS2
			n.List.Set(lhs)
			n.Rlist.Set(rhs)
		}
		return n

	case *syntax.BranchStmt:
		var op Op
		switch stmt.Tok {
		case syntax.Break:
			op = OBREAK
		case syntax.Continue:
			op = OCONTINUE
		case syntax.Fallthrough:
			if !fallOK {
				yyerror("fallthrough statement out of place")
			}
			op = OFALL
		case syntax.Goto:
			op = OGOTO
		default:
			panic("unhandled BranchStmt")
		}
		n := p.nod(stmt, op, nil, nil)
		if stmt.Label != nil {
			n.Sym = p.name(stmt.Label)
		}
		return n
	case *syntax.CallStmt:
		var op Op
		switch stmt.Tok {
		case syntax.Defer:
			op = ODEFER
		case syntax.Go:
			op = OGO
		default:
			panic("unhandled CallStmt")
		}
		return p.nod(stmt, op, p.expr(stmt.Call), nil)
	case *syntax.ReturnStmt:
		var results []*Node
		if stmt.Results != nil {
			results = p.exprList(stmt.Results)
		}
		n := p.nod(stmt, ORETURN, nil, nil)
		n.List.Set(results)
		if n.List.Len() == 0 && Curfn != nil {
			for _, ln := range Curfn.Func.Dcl {
				if ln.Class() == PPARAM {
					continue
				}
				if ln.Class() != PPARAMOUT {
					break
				}
				if asNode(ln.Sym.Def) != ln {
					yyerror("%s is shadowed during return", ln.Sym.Name)
				}
			}
		}
		return n
	case *syntax.IfStmt:
		return p.ifStmt(stmt)
	case *syntax.ForStmt:
		return p.forStmt(stmt)
	case *syntax.SwitchStmt:
		return p.switchStmt(stmt)
	case *syntax.SelectStmt:
		return p.selectStmt(stmt)
	}
	panic("unhandled Stmt")
}

func (p *noder2) assignList(expr syntax.Expr, defn *Node, colas bool) []*Node {
	if !colas {
		return p.exprList(expr)
	}

	defn.SetColas(true)

	var exprs []syntax.Expr
	if list, ok := expr.(*syntax.ListExpr); ok {
		exprs = list.ElemList
	} else {
		exprs = []syntax.Expr{expr}
	}

	res := make([]*Node, len(exprs))
	seen := make(map[*types.Sym]bool, len(exprs))

	newOrErr := false
	for i, expr := range exprs {
		p.setlineno(expr)
		res[i] = nblank

		name, ok := expr.(*syntax.Name)
		if !ok {
			p.yyerrorpos(expr.Pos(), "non-name %v on left side of :=", p.expr(expr))
			newOrErr = true
			continue
		}

		sym := p.name(name)
		if sym.IsBlank() {
			continue
		}

		if seen[sym] {
			p.yyerrorpos(expr.Pos(), "%v repeated on left side of :=", sym)
			newOrErr = true
			continue
		}
		seen[sym] = true

		if sym.Block == types.Block {
			res[i] = oldname(sym)
			continue
		}

		newOrErr = true
		n := newname(sym)
		declare(n, dclcontext)
		n.Name.Defn = defn
		defn.Ninit.Append(nod(ODCL, n, nil))
		res[i] = n
	}

	if !newOrErr {
		yyerrorl(defn.Pos, "no new variables on left side of :=")
	}
	return res
}

func (p *noder2) blockStmt(stmt *syntax.BlockStmt) []*Node {
	p.openScope(stmt.Pos())
	nodes := p.stmts(stmt.List)
	p.closeScope(stmt.Rbrace)
	return nodes
}

func (p *noder2) ifStmt(stmt *syntax.IfStmt) *Node {
	p.openScope(stmt.Pos())
	n := p.nod(stmt, OIF, nil, nil)
	if stmt.Init != nil {
		n.Ninit.Set1(p.stmt(stmt.Init))
	}
	if stmt.Cond != nil {
		n.Left = p.expr(stmt.Cond)
	}
	n.Nbody.Set(p.blockStmt(stmt.Then))
	if stmt.Else != nil {
		e := p.stmt(stmt.Else)
		if e.Op == OBLOCK && e.Ninit.Len() == 0 {
			n.Rlist.Set(e.List.Slice())
		} else {
			n.Rlist.Set1(e)
		}
	}
	p.closeAnotherScope()
	return n
}

func (p *noder2) forStmt(stmt *syntax.ForStmt) *Node {
	p.openScope(stmt.Pos())
	var n *Node
	if r, ok := stmt.Init.(*syntax.RangeClause); ok {
		if stmt.Cond != nil || stmt.Post != nil {
			panic("unexpected RangeClause")
		}

		n = p.nod(r, ORANGE, nil, p.expr(r.X))
		if r.Lhs != nil {
			n.List.Set(p.assignList(r.Lhs, n, r.Def))
		}
	} else {
		n = p.nod(stmt, OFOR, nil, nil)
		if stmt.Init != nil {
			n.Ninit.Set1(p.stmt(stmt.Init))
		}
		if stmt.Cond != nil {
			n.Left = p.expr(stmt.Cond)
		}
		if stmt.Post != nil {
			n.Right = p.stmt(stmt.Post)
		}
	}
	n.Nbody.Set(p.blockStmt(stmt.Body))
	p.closeAnotherScope()
	return n
}

func (p *noder2) switchStmt(stmt *syntax.SwitchStmt) *Node {
	p.openScope(stmt.Pos())
	n := p.nod(stmt, OSWITCH, nil, nil)
	if stmt.Init != nil {
		n.Ninit.Set1(p.stmt(stmt.Init))
	}
	if stmt.Tag != nil {
		n.Left = p.expr(stmt.Tag)
	}

	tswitch := n.Left
	if tswitch != nil && tswitch.Op != OTYPESW {
		tswitch = nil
	}
	n.List.Set(p.caseClauses(stmt.Body, tswitch, stmt.Rbrace))

	p.closeScope(stmt.Rbrace)
	return n
}

func (p *noder2) caseClauses(clauses []*syntax.CaseClause, tswitch *Node, rbrace syntax.Pos) []*Node {
	nodes := make([]*Node, 0, len(clauses))
	for i, clause := range clauses {
		p.setlineno(clause)
		if i > 0 {
			p.closeScope(clause.Pos())
		}
		p.openScope(clause.Pos())

		n := p.nod(clause, OCASE, nil, nil)
		if clause.Cases != nil {
			n.List.Set(p.exprList(clause.Cases))
		}
		if tswitch != nil && tswitch.Left != nil {
			nn := newname(tswitch.Left.Sym)
			declare(nn, dclcontext)
			n.Rlist.Set1(nn)
			// keep track of the instances for reporting unused
			nn.Name.Defn = tswitch
		}

		// Trim trailing empty statements. We omit them from
		// the Node AST anyway, and it's easier to identify
		// out-of-place fallthrough statements without them.
		body := clause.Body
		for len(body) > 0 {
			if _, ok := body[len(body)-1].(*syntax.EmptyStmt); !ok {
				break
			}
			body = body[:len(body)-1]
		}

		n.Nbody.Set(p.stmtsFall(body, true))
		if l := n.Nbody.Len(); l > 0 && n.Nbody.Index(l-1).Op == OFALL {
			if tswitch != nil {
				yyerror("cannot fallthrough in type switch")
			}
			if i+1 == len(clauses) {
				yyerror("cannot fallthrough final case in switch")
			}
		}

		nodes = append(nodes, n)
	}
	if len(clauses) > 0 {
		p.closeScope(rbrace)
	}
	return nodes
}

func (p *noder2) selectStmt(stmt *syntax.SelectStmt) *Node {
	n := p.nod(stmt, OSELECT, nil, nil)
	n.List.Set(p.commClauses(stmt.Body, stmt.Rbrace))
	return n
}

func (p *noder2) commClauses(clauses []*syntax.CommClause, rbrace syntax.Pos) []*Node {
	nodes := make([]*Node, 0, len(clauses))
	for i, clause := range clauses {
		p.setlineno(clause)
		if i > 0 {
			p.closeScope(clause.Pos())
		}
		p.openScope(clause.Pos())

		n := p.nod(clause, OCASE, nil, nil)
		if clause.Comm != nil {
			n.List.Set1(p.stmt(clause.Comm))
		}
		n.Nbody.Set(p.stmts(clause.Body))
		nodes = append(nodes, n)
	}
	if len(clauses) > 0 {
		p.closeScope(rbrace)
	}
	return nodes
}

func (p *noder2) labeledStmt(label *syntax.LabeledStmt, fallOK bool) *Node {
	lhs := p.nodSym(label, OLABEL, nil, p.name(label.Label))

	var ls *Node
	if label.Stmt != nil { // TODO(mdempsky): Should always be present.
		ls = p.stmtFall(label.Stmt, fallOK)
	}

	lhs.Name.Defn = ls
	l := []*Node{lhs}
	if ls != nil {
		if ls.Op == OBLOCK && ls.Ninit.Len() == 0 {
			l = append(l, ls.List.Slice()...)
		} else {
			l = append(l, ls)
		}
	}
	return liststmt(l)
}

func (p *noder2) unOp(op syntax.Operator) Op {
	if uint64(op) >= uint64(len(unOps)) || unOps[op] == 0 {
		panic("invalid Operator")
	}
	return unOps[op]
}

func (p *noder2) binOp(op syntax.Operator) Op {
	if uint64(op) >= uint64(len(binOps)) || binOps[op] == 0 {
		panic("invalid Operator")
	}
	return binOps[op]
}

func (p *noder2) basicLit(lit *syntax.BasicLit) Val {
	// We don't use the errors of the conversion routines to determine
	// if a literal string is valid because the conversion routines may
	// accept a wider syntax than the language permits. Rely on lit.Bad
	// instead.
	switch s := lit.Value; lit.Kind {
	case syntax.IntLit:
		checkLangCompat(lit)
		x := new(Mpint)
		if !lit.Bad {
			x.SetString(s)
		}
		return Val{U: x}

	case syntax.FloatLit:
		checkLangCompat(lit)
		x := newMpflt()
		if !lit.Bad {
			x.SetString(s)
		}
		return Val{U: x}

	case syntax.ImagLit:
		checkLangCompat(lit)
		x := newMpcmplx()
		if !lit.Bad {
			x.Imag.SetString(strings.TrimSuffix(s, "i"))
		}
		return Val{U: x}

	case syntax.RuneLit:
		x := new(Mpint)
		x.Rune = true
		if !lit.Bad {
			u, _ := strconv.Unquote(s)
			var r rune
			if len(u) == 1 {
				r = rune(u[0])
			} else {
				r, _ = utf8.DecodeRuneInString(u)
			}
			x.SetInt64(int64(r))
		}
		return Val{U: x}

	case syntax.StringLit:
		var x string
		if !lit.Bad {
			if len(s) > 0 && s[0] == '`' {
				// strip carriage returns from raw string
				s = strings.Replace(s, "\r", "", -1)
			}
			x, _ = strconv.Unquote(s)
		}
		return Val{U: x}

	default:
		panic("unhandled BasicLit kind")
	}
}

func (p *noder2) name(name *syntax.Name) *types.Sym {
	return lookup(name.Value)
}

func (p *noder2) mkname(name *syntax.Name) *Node {
	// TODO(mdempsky): Set line number?
	return mkname(p.name(name))
}

func (p *noder2) wrapname(n syntax.Node, x *Node) *Node {
	// These nodes do not carry line numbers.
	// Introduce a wrapper node to give them the correct line.
	switch x.Op {
	case OTYPE, OLITERAL:
		if x.Sym == nil {
			break
		}
		fallthrough
	case ONAME, ONONAME, OPACK:
		x = p.nod(n, OPAREN, x, nil)
		x.SetImplicit(true)
	}
	return x
}

func (p *noder2) nod(orig syntax.Node, op Op, left, right *Node) *Node {
	return nodl(p.pos(orig), op, left, right)
}

func (p *noder2) nodSym(orig syntax.Node, op Op, left *Node, sym *types.Sym) *Node {
	n := nodSym(op, left, sym)
	n.Pos = p.pos(orig)
	return n
}

func (p *noder2) pos(n interface{ Pos() syntax.Pos }) src.XPos {
	// TODO(gri): orig.Pos() should always be known - fix package syntax
	xpos := lineno
	if pos := n.Pos(); pos.IsKnown() {
		xpos = p.makeXPos(pos)
	}
	return xpos
}

func (p *noder2) setlineno(n syntax.Node) {
	if n != nil {
		lineno = p.pos(n)
	}
}

// error is called concurrently if files are parsed concurrently.
func (p *noder2) error(err error) {
	p.err <- err.(syntax.Error)
}

func (p *noder2) checkUnused(pragma *Pragma) {
	for _, pos := range pragma.Pos {
		if pos.Flag&pragma.Flag != 0 {
			p.yyerrorpos(pos.Pos, "misplaced compiler directive")
		}
	}
	if len(pragma.Embeds) > 0 {
		for _, e := range pragma.Embeds {
			p.yyerrorpos(e.Pos, "misplaced go:embed directive")
		}
	}
}

func (p *noder2) checkUnusedDuringParse(pragma *Pragma) {
	for _, pos := range pragma.Pos {
		if pos.Flag&pragma.Flag != 0 {
			p.error(syntax.Error{Pos: pos.Pos, Msg: "misplaced compiler directive"})
		}
	}
	if len(pragma.Embeds) > 0 {
		for _, e := range pragma.Embeds {
			p.error(syntax.Error{Pos: e.Pos, Msg: "misplaced go:embed directive"})
		}
	}
}

// pragma is called concurrently if files are parsed concurrently.
func (p *noder2) pragma(pos syntax.Pos, blankLine bool, text string, old syntax.Pragma) syntax.Pragma {
	pragma, _ := old.(*Pragma)
	if pragma == nil {
		pragma = new(Pragma)
	}

	if text == "" {
		// unused pragma; only called with old != nil.
		p.checkUnusedDuringParse(pragma)
		return nil
	}

	if strings.HasPrefix(text, "line ") {
		// line directives are handled by syntax package
		panic("unreachable")
	}

	if !blankLine {
		// directive must be on line by itself
		p.error(syntax.Error{Pos: pos, Msg: "misplaced compiler directive"})
		return pragma
	}

	switch {
	case strings.HasPrefix(text, "go:linkname "):
		f := strings.Fields(text)
		if !(2 <= len(f) && len(f) <= 3) {
			p.error(syntax.Error{Pos: pos, Msg: "usage: //go:linkname localname [linkname]"})
			break
		}
		// The second argument is optional. If omitted, we use
		// the default object symbol name for this and
		// linkname only serves to mark this symbol as
		// something that may be referenced via the object
		// symbol name from another package.
		var target string
		if len(f) == 3 {
			target = f[2]
		}
		p.linknames = append(p.linknames, linkname{pos, f[1], target})

	case text == "go:embed", strings.HasPrefix(text, "go:embed "):
		args, err := parseGoEmbed(text[len("go:embed"):])
		if err != nil {
			p.error(syntax.Error{Pos: pos, Msg: err.Error()})
		}
		if len(args) == 0 {
			p.error(syntax.Error{Pos: pos, Msg: "usage: //go:embed pattern..."})
			break
		}
		pragma.Embeds = append(pragma.Embeds, PragmaEmbed{pos, args})

	case strings.HasPrefix(text, "go:cgo_import_dynamic "):
		// This is permitted for general use because Solaris
		// code relies on it in golang.org/x/sys/unix and others.
		fields := pragmaFields(text)
		if len(fields) >= 4 {
			lib := strings.Trim(fields[3], `"`)
			if lib != "" && !safeArg(lib) && !isCgoGeneratedFile(pos) {
				p.error(syntax.Error{Pos: pos, Msg: fmt.Sprintf("invalid library name %q in cgo_import_dynamic directive", lib)})
			}
			p.pragcgo(pos, text)
			pragma.Flag |= pragmaFlag("go:cgo_import_dynamic")
			break
		}
		fallthrough
	case strings.HasPrefix(text, "go:cgo_"):
		// For security, we disallow //go:cgo_* directives other
		// than cgo_import_dynamic outside cgo-generated files.
		// Exception: they are allowed in the standard library, for runtime and syscall.
		if !isCgoGeneratedFile(pos) && !compiling_std {
			p.error(syntax.Error{Pos: pos, Msg: fmt.Sprintf("//%s only allowed in cgo-generated code", text)})
		}
		p.pragcgo(pos, text)
		fallthrough // because of //go:cgo_unsafe_args
	default:
		verb := text
		if i := strings.Index(text, " "); i >= 0 {
			verb = verb[:i]
		}
		flag := pragmaFlag(verb)
		const runtimePragmas = Systemstack | Nowritebarrier | Nowritebarrierrec | Yeswritebarrierrec
		if !compiling_runtime && flag&runtimePragmas != 0 {
			p.error(syntax.Error{Pos: pos, Msg: fmt.Sprintf("//%s only allowed in runtime", verb)})
		}
		if flag == 0 && !allowedStdPragmas[verb] && compiling_std {
			p.error(syntax.Error{Pos: pos, Msg: fmt.Sprintf("//%s is not allowed in the standard library", verb)})
		}
		pragma.Flag |= flag
		pragma.Pos = append(pragma.Pos, PragmaPos{flag, pos})
	}

	return pragma
}
