// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"fmt"
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
	if base.SyntaxErrors() != 0 {
		base.ErrorExit()
	}

	// setup and syntax error reporting
	nodersmap := make(map[string]*noder)
	files := make([]*syntax.File, len(noders))
	for i, p := range noders {
		nodersmap[p.file.Pos().RelFilename()] = p
		files[i] = p.file
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
	files = nil
	if err != nil {
		base.FatalfAt(src.NoXPos, "conf.Check error: %v", err)
	}
	base.ExitIfErrors()
	if base.Flag.G < 2 {
		os.Exit(0)
	}

	g := irgen{
		self:    pkg,
		info:    &info,
		basemap: make(map[*syntax.PosBase]*src.PosBase),
		objs:    make(map[types2.Object]*ir.Name),
	}
	g.generate(noders)

	if base.Flag.G < 3 {
		os.Exit(0)
	}
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

type irgen struct {
	self *types2.Package
	info *types2.Info

	objs      map[types2.Object]*ir.Name
	todos     []todo
	resolveOK bool

	basemap   map[*syntax.PosBase]*src.PosBase
	basecache struct {
		last *syntax.PosBase
		base *src.PosBase
	}
}

type todo struct {
	name *ir.Name
	obj  *types2.TypeName
}

// makeSrcPosBase translates from a *syntax.PosBase to a *src.PosBase.
func (g *irgen) makeSrcPosBase(b0 *syntax.PosBase) *src.PosBase {
	// TODO(mdempsky): Deduplicate this logic with noder's.

	if b0 == nil {
		// TODO(mdempsky): Why/when does this happen? It
		// wasn't needed in noder.makeSrcPosBase.
		return nil
	}

	// fast path: most likely PosBase hasn't changed
	if g.basecache.last == b0 {
		return g.basecache.base
	}

	b1, ok := g.basemap[b0]
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
			p1 := src.MakePos(g.makeSrcPosBase(p0b), p0.Line(), p0.Col())
			b1 = src.NewLinePragmaBase(p1, fn, fileh(fn), b0.Line(), b0.Col())
		}
		g.basemap[b0] = b1
	}

	// update cache
	g.basecache.last = b0
	g.basecache.base = b1

	return b1
}

// poser is implemented by both syntax.Node and types.Object.
type poser interface {
	Pos() syntax.Pos
}

// TODO(mdempsky): It would be really nice to ensure this function
// gets inlined so we can avoid runtime interface-to-interface
// conversions from syntax.Node/types2.Object to this anonymous
// "poser" interface type.
func (g *irgen) pos(p poser) src.XPos {
	return g.pos0(p.Pos())
}

func (g *irgen) unhandled(what string, p poser) {
	base.FatalfAt(g.pos(p), "unhandled %s: %T", what, p)
}

func (g *irgen) pos0(pos syntax.Pos) src.XPos {
	posBase := g.makeSrcPosBase(pos.Base())
	return base.Ctxt.PosTable.XPos(src.MakePos(posBase, pos.Line(), pos.Col()))
}

func (g *irgen) generate(noders []*noder) {
	types.LocalPkg.Name = g.self.Name()
	typecheck.TypecheckAllowed = true

	for _, p := range noders {
		for k, v := range p.basemap {
			if g.basemap[k] != nil {
				base.Fatalf("duplicate basemap? %v", k)
			}
			g.basemap[k] = v
		}
	}

	// Forward declare all objects and resolve their types.
	for _, p := range noders {
		for _, decl := range p.file.DeclList {
			switch decl := decl.(type) {
			case *syntax.ImportDecl:
				g.importDecl(p, decl)
			case *syntax.ConstDecl:
				g.defs(decl.NameList)
			case *syntax.FuncDecl:
				g.def(decl.Name)
			case *syntax.TypeDecl:
				g.def(decl.Name)
			case *syntax.VarDecl:
				g.defs(decl.NameList)
			}
		}
	}
	g.flushResolve()

	for _, p := range noders {
		p.processPragmas()
		g.decls(p, p.file.DeclList)
	}

	types.LocalPkg.Height = myheight
	typecheck.DeclareUniverse()

	for i, n := range typecheck.Target.Externs {
		if n.Op() == ir.ONAME {
			typecheck.Target.Externs[i] = typecheck.Expr(n)
		}
	}
}

func (g *irgen) flushResolve() {
	g.resolveOK = false

	for len(g.todos) > 0 {
		i := len(g.todos) - 1
		next := g.todos[i]
		g.todos = g.todos[:i]

		n1, n2 := next.name.Type(), next.obj.Type()
		n1.SetUnderlying(g.typ(n2.Underlying()))

		if n2, ok := n2.(*types2.Named); ok && !n1.IsInterface() {
			methods := make([]*types.Field, n2.NumMethods())
			for i := range methods {
				m := n2.Method(i)
				meth := g.obj(m)
				methods[i] = types.NewField(meth.Pos(), g.selector(m), meth.Type())
				methods[i].Nname = meth
			}
			n1.Methods().Set(methods)
		}
	}

	g.resolveOK = true
}

func (g *irgen) resolve(name *ir.Name, defined *types2.TypeName) {
	g.todos = append(g.todos, todo{name, defined})
	if g.resolveOK {
		g.flushResolve()
	}
}

func (g *irgen) pkg(pkg *types2.Package) *types.Pkg {
	switch pkg {
	case nil:
		return types.BuiltinPkg
	case g.self:
		return types.LocalPkg
	case types2.Unsafe:
		return ir.Pkgs.Unsafe
	}
	return types.NewPkg(pkg.Path(), pkg.Name())
}
