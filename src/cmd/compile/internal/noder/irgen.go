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

	n := irgen{
		self:    pkg,
		info:    &info,
		basemap: make(map[*syntax.PosBase]*src.PosBase),
	}
	n.meow(noders)

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

	objs  map[types2.Object]*ir.Name
	todos []todo

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
func (p *irgen) makeSrcPosBase(b0 *syntax.PosBase) *src.PosBase {
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

func (p *irgen) pos0(pos syntax.Pos) src.XPos {
	return base.Ctxt.PosTable.XPos(src.MakePos(p.makeSrcPosBase(pos.Base()), pos.Line(), pos.Col()))
}

func (p *irgen) pos(n interface{ Pos() syntax.Pos }) src.XPos {
	return p.pos0(n.Pos())
}

func (p *irgen) meow(noders []*noder) {
	types.LocalPkg.Name = p.self.Name()
	typecheck.TypecheckAllowed = true

	for _, no := range noders {
		for k, v := range no.basemap {
			if p.basemap[k] != nil {
				base.Fatalf("duplicate basemap? %v", k)
			}
			p.basemap[k] = v
		}
	}

	for _, no := range noders {
		p.decls(no, no.file.DeclList)
	}
	p.resolve()

	types.LocalPkg.Height = myheight

	for _, n := range typecheck.Target.Decls {
		if n, ok := n.(*ir.Func); ok {
			typecheck.Func(n)
			typecheck.FuncBody(n)
		}
	}

	for _, p := range noders {
		p.processPragmas()
	}
}

func (p *irgen) resolve() {
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

func (p *irgen) pkg(pkg *types2.Package) *types.Pkg {
	switch pkg {
	case nil:
		return types.BuiltinPkg
	case p.self:
		return types.LocalPkg
	}
	return types.NewPkg(pkg.Path(), pkg.Name())
}
