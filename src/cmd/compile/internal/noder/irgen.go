// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
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
		self: pkg,
		info: &info,
		objs: make(map[types2.Object]*ir.Name),
	}
	g.generate(noders)

	if base.Flag.G < 3 {
		os.Exit(0)
	}
}

type irgen struct {
	self *types2.Package
	info *types2.Info

	posMap
	objs   map[types2.Object]*ir.Name
	marker dwarfgen.ScopeMarker
}

func (g *irgen) generate(noders []*noder) {
	types.LocalPkg.Name = g.self.Name()
	typecheck.TypecheckAllowed = true

	for _, p := range noders {
		g.posMap.join(&p.posMap)
	}

	for _, p := range noders {
		g.pragmaFlags(p.file.Pragma, ir.GoBuildPragma)
		for _, decl := range p.file.DeclList {
			switch decl := decl.(type) {
			case *syntax.ImportDecl:
				g.importDecl(p, decl)
			}
		}
	}

	for _, p := range noders {
		for _, decl := range p.file.DeclList {
			switch decl := decl.(type) {
			case *syntax.TypeDecl:
				g.typeDecl(nil, decl)
			}
		}
	}

	for _, p := range noders {
		g.decls(p.file.DeclList)
	}

	types.LocalPkg.Height = myheight
	typecheck.DeclareUniverse()

	for i, n := range typecheck.Target.Externs {
		if n.Op() == ir.ONAME {
			typecheck.Target.Externs[i] = typecheck.Expr(n)
		}
	}

	for _, p := range noders {
		// Process linkname and cgo pragmas.
		p.processPragmas()

		// Double check for any type-checking inconsistencies.
		syntax.Walk(p.file, func(n syntax.Node) bool {
			g.validate(n)
			return false
		})
	}
}

func (g *irgen) unhandled(what string, p poser) {
	base.FatalfAt(g.pos(p), "unhandled %s: %T", what, p)
	panic("unreachable")
}
