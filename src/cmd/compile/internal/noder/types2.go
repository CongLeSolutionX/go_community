// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"fmt"
	"go/constant"
	"go/token"
	"io"
	"os"

	"cmd/compile/internal/base"
	"cmd/compile/internal/importer"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
	"cmd/internal/src"
)

func check2(noders []*noder) {
	// generic noding phase (using new typechecker)

	if base.SyntaxErrors() != 0 {
		base.ErrorExit()
	}

	// setup and syntax error reporting
	nodersmap := make(map[string]*noder)
	files := make([]*syntax.File, len(noders))
	for i, p := range noders {
		nodersmap[p.file.Pos().RelFilename()] = p
		files[i] = p.file
		p.file = nil
	}

	// typechecking
	conf := types2.Config{
		InferFromConstraints:  true,
		IgnoreBranches:        true, // parser already checked via syntax.CheckBranches mode
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
			p := nodersmap[terr.Pos.RelFilename()]
			base.ErrorfAt(p.makeXPos(terr.Pos), "%s", terr.Msg)
		},
		Importer: &gcimports{
			packages: make(map[string]*types2.Package),
			lookup: func(path string) (io.ReadCloser, error) {
				file, ok := findpkg(path)
				if !ok {
					return nil, fmt.Errorf("can't find import: %q", path)
				}
				return os.Open(file)
			},
		},
	}
	info := types2.Info{
		Types:      make(map[syntax.Expr]types2.TypeAndValue),
		Defs:       make(map[*syntax.Name]types2.Object),
		Uses:       make(map[*syntax.Name]types2.Object),
		Selections: make(map[*syntax.SelectorExpr]*types2.Selection),
		Implicits:  make(map[syntax.Node]types2.Object),
		// expand as needed
	}
	pkg, err := conf.Check(base.Ctxt.Pkgpath, files, &info)
	if err != nil {
		base.FatalfAt(src.NoXPos, "conf.Check error: %v", err)
	}
	base.ExitIfErrors()
	if base.Flag.G < 2 {
		os.Exit(0)
	}

	typecheck.TypecheckAllowed = true

	n := nyan{
		self:    pkg,
		info:    &info,
		basemap: make(map[*syntax.PosBase]*src.PosBase),
	}
	for _, p := range noders {
		for k, v := range p.basemap {
			if n.basemap[k] != nil {
				base.Fatalf("duplicate basemap? %v", k)
			}
			n.basemap[k] = v
		}
	}

	n.meow(files)

	for _, p := range noders {
		p.importedUnsafe = true // XXX
		p.processPragmas()
	}

	if base.Flag.G < 3 {
		os.Exit(0)
	}

	// Typecheck.
	types.LocalPkg.Height = myheight
}

// Temporary import helper to get type2-based type-checking going.
type gcimports struct {
	packages map[string]*types2.Package
	lookup   func(path string) (io.ReadCloser, error)
}

func (m *gcimports) Import(path string) (*types2.Package, error) {
	return m.ImportFrom(path, "" /* no vendoring */, 0)
}

func (m *gcimports) ImportFrom(path, srcDir string, mode types2.ImportMode) (*types2.Package, error) {
	if mode != 0 {
		panic("mode must be 0")
	}
	return importer.Import(m.packages, path, srcDir, m.lookup)
}

type nyan struct {
	self *types2.Package
	info *types2.Info

	objs  map[types2.Object]*ir.Name
	todos []todo

	basemap   map[*syntax.PosBase]*src.PosBase
	basecache struct {
		last *syntax.PosBase
		base *src.PosBase
	}
}

// makeSrcPosBase translates from a *syntax.PosBase to a *src.PosBase.
func (p *nyan) makeSrcPosBase(b0 *syntax.PosBase) *src.PosBase {
	if b0 == nil {
		return nil
	}

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

func (p *nyan) pos0(pos syntax.Pos) src.XPos {
	return base.Ctxt.PosTable.XPos(src.MakePos(p.makeSrcPosBase(pos.Base()), pos.Line(), pos.Col()))
}

type todo struct {
	name *ir.Name
	obj  *types2.TypeName
}

func (p *nyan) meow(files []*syntax.File) {
	types.LocalPkg.Name = p.self.Name()

	for _, file := range files {
		for _, decl := range file.DeclList {
			switch decl := decl.(type) {
			case *syntax.ImportDecl:
				ipkg := importfile(constant.MakeFromLiteral(decl.Path.Value, token.STRING, 0))

				if !ipkg.Direct {
					typecheck.Target.Imports = append(typecheck.Target.Imports, ipkg)
				}
				ipkg.Direct = true
			}
		}
	}

	scope := p.self.Scope()
	for _, name := range scope.Names() {
		p.obj(scope.Lookup(name))
		p.resolve()
	}

	for _, file := range files {
		for _, decl := range file.DeclList {
			switch decl := decl.(type) {
			case *syntax.FuncDecl:
				p.funk(decl)
				p.resolve()
			}
		}
	}

	for _, n := range typecheck.Target.Decls {
		if n, ok := n.(*ir.Func); ok {
			typecheck.Func(n)
		}
	}
	for _, n := range typecheck.Target.Decls {
		if n, ok := n.(*ir.Func); ok {
			typecheck.FuncBody(n)
		}
	}
}

func (p *nyan) pos(n interface{ Pos() syntax.Pos }) src.XPos { return p.pos0(n.Pos()) }

func (p *nyan) resolve() {
	for len(p.todos) > 0 {
		i := len(p.todos) - 1
		next := p.todos[i]
		p.todos = p.todos[:i]

		n1, n2 := next.name.Type(), next.obj.Type().(*types2.Named)
		n1.SetUnderlying(p.typ(n2.Underlying()))

		if !n1.IsInterface() {
			methods := make([]*types.Field, n2.NumMethods())
			for i := range methods {
				m := n2.Method(i)
				sig := m.Type().(*types2.Signature)
				mtyp := p.signature(p.param(sig.Recv()), sig)
				methods[i] = types.NewField(p.pos(m), p.selector(m), mtyp)
			}
			n1.Methods().Set(methods)
		}
	}
}

func (p *nyan) obj(obj types2.Object) *ir.Name {
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

func (p *nyan) sym(obj types2.Object) *types.Sym {
	if name := obj.Name(); name != "" {
		return p.pkg(obj.Pkg()).Lookup(obj.Name())
	}
	return nil
}

func (p *nyan) selector(obj types2.Object) *types.Sym {
	pkg, name := p.pkg(obj.Pkg()), obj.Name()
	if types.IsExported(name) {
		pkg = types.LocalPkg
	}
	return pkg.Lookup(name)
}

func (p *nyan) pkg(pkg *types2.Package) *types.Pkg {
	switch pkg {
	case nil:
		return types.BuiltinPkg
	case p.self:
		return types.LocalPkg
	}
	return types.NewPkg(pkg.Path(), pkg.Name())
}

func (p *nyan) typ(typ types2.Type) *types.Type {
	switch typ := typ.(type) {
	case *types2.Basic:
		return p.basic(typ)
	case *types2.Named:
		obj := p.obj(typ.Obj())
		if obj.Op() != ir.OTYPE {
			base.FatalfAt(obj.Pos(), "expected type: %L", obj)
		}
		return obj.Type()

	case *types2.Array:
		return types.NewArray(p.typ(typ.Elem()), typ.Len())
	case *types2.Chan:
		return types.NewChan(p.typ(typ.Elem()), dirs[typ.Dir()])
	case *types2.Map:
		return types.NewMap(p.typ(typ.Key()), p.typ(typ.Elem()))
	case *types2.Pointer:
		return types.NewPtr(p.typ(typ.Elem()))
	case *types2.Signature:
		return p.signature(nil, typ)
	case *types2.Slice:
		return types.NewSlice(p.typ(typ.Elem()))

	case *types2.Struct:
		fields := make([]*types.Field, typ.NumFields())
		for i := range fields {
			v := typ.Field(i)
			f := p.field(v)
			f.Note = typ.Tag(i)
			if v.Embedded() {
				f.Embedded = 1
			}
			fields[i] = f
		}
		return types.NewStruct(p.tpkg(typ), fields)

	case *types2.Interface:
		embeddeds := make([]*types.Field, typ.NumEmbeddeds())
		for i := range embeddeds {
			// TODO(mdempsky): Get embedding position.
			e := typ.EmbeddedType(i)
			embeddeds[i] = types.NewField(src.NoXPos, nil, p.typ(e))
		}

		methods := make([]*types.Field, typ.NumExplicitMethods())
		for i := range methods {
			m := typ.ExplicitMethod(i)
			mtyp := p.signature(typecheck.FakeRecvField(), m.Type().(*types2.Signature))
			methods[i] = types.NewField(p.pos(m), p.selector(m), mtyp)
		}

		return types.NewInterface(p.tpkg(typ), append(embeddeds, methods...))
	}

	base.FatalfAt(src.NoXPos, "unhandled type: %v (%T)", typ, typ)
	panic("unreachable")
}

func (p *nyan) field(v *types2.Var) *types.Field {
	return types.NewField(p.pos(v), p.selector(v), p.typ(v.Type()))
}

func (p *nyan) param(v *types2.Var) *types.Field {
	return types.NewField(p.pos(v), p.sym(v), p.typ(v.Type()))
}

func (p *nyan) signature(recv *types.Field, sig *types2.Signature) *types.Type {
	do := func(typ *types2.Tuple) []*types.Field {
		fields := make([]*types.Field, typ.Len())
		for i := range fields {
			fields[i] = p.param(typ.At(i))
		}
		return fields
	}

	params := do(sig.Params())
	results := do(sig.Results())
	if sig.Variadic() {
		params[len(params)-1].SetIsDDD(true)
	}

	return types.NewSignature(p.tpkg(sig), recv, params, results)
}

func (p *nyan) tpkg(typ types2.Type) *types.Pkg {
	// TODO(mdempsky): Return appropriate package for imported
	// types. This isn't urgent though: it's only needed so go/types can
	// return the correct Pkg for struct fields, signature parameters,
	// and interface methods.
	return types.LocalPkg
}

func (p *nyan) basic(typ *types2.Basic) *types.Type {
	switch typ.Name() {
	case "byte":
		return types.ByteType
	case "rune":
		return types.RuneType
	}
	return *basics[typ.Kind()]
}

var basics = [...]**types.Type{
	types2.Invalid:        new(*types.Type),
	types2.Bool:           &types.Types[types.TBOOL],
	types2.Int:            &types.Types[types.TINT],
	types2.Int8:           &types.Types[types.TINT8],
	types2.Int16:          &types.Types[types.TINT16],
	types2.Int32:          &types.Types[types.TINT32],
	types2.Int64:          &types.Types[types.TINT64],
	types2.Uint:           &types.Types[types.TUINT],
	types2.Uint8:          &types.Types[types.TUINT8],
	types2.Uint16:         &types.Types[types.TUINT16],
	types2.Uint32:         &types.Types[types.TUINT32],
	types2.Uint64:         &types.Types[types.TUINT64],
	types2.Uintptr:        &types.Types[types.TUINTPTR],
	types2.Float32:        &types.Types[types.TFLOAT32],
	types2.Float64:        &types.Types[types.TFLOAT64],
	types2.Complex64:      &types.Types[types.TCOMPLEX64],
	types2.Complex128:     &types.Types[types.TCOMPLEX128],
	types2.String:         &types.Types[types.TSTRING],
	types2.UnsafePointer:  &types.Types[types.TUNSAFEPTR],
	types2.UntypedBool:    &types.UntypedBool,
	types2.UntypedInt:     &types.UntypedInt,
	types2.UntypedRune:    &types.UntypedRune,
	types2.UntypedFloat:   &types.UntypedFloat,
	types2.UntypedComplex: &types.UntypedComplex,
	types2.UntypedString:  &types.UntypedString,
	types2.UntypedNil:     &types.Types[types.TNIL],
}

var dirs = [...]types.ChanDir{
	types2.SendRecv: types.Cboth,
	types2.SendOnly: types.Csend,
	types2.RecvOnly: types.Crecv,
}
