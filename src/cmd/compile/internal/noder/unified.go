// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"bytes"
	"fmt"
	"runtime"

	"cmd/compile/internal/base"
	"cmd/compile/internal/inline"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
	"cmd/internal/src"
)

// unifiedIRLevel controls what -G level enables unified IR.
const unifiedIRLevel = 4

// localPkgReader holds the package reader used for reading the local
// package. It exists so the unified IR linker can refer back to it
// later.
var localPkgReader *pkgReader

// useUnifiedIR reports whether the unified IR frontend should be
// used; and if so, uses it to construct the local package's IR.
func useUnifiedIR(noders []*noder) bool {
	if base.Flag.G < unifiedIRLevel {
		return false
	}

	inline.NewInline = InlineCall

	data := writePkgStub(noders)

	assert(types.LocalPkg.Path == "")
	types.LocalPkg.Height = 0 // reset so pkgReader.pkgIdx doesn't complain

	target := typecheck.Target

	typecheck.TypecheckAllowed = true

	localPkgReader = newPkgReader(newPkgDecoder(types.LocalPkg.Path, data))
	readPackage(localPkgReader, types.LocalPkg)

	r := localPkgReader.newReader(relocMeta, privateRootIdx, syncPublic)
	r.ext = r
	r.pkgInit(types.LocalPkg, target)

	// Don't use range--bodyIdx can add closures to todoBodies.
	for len(todoBodies) > 0 {
		// The order we expand bodies doesn't matter, so pop from the end
		// to reduce todoBodies reallocations if it grows further.
		fn := todoBodies[len(todoBodies)-1]
		todoBodies = todoBodies[:len(todoBodies)-1]

		pri, ok := bodyReader[fn]
		assert(ok)
		pri.funcBody(fn)

		// Instantiated generic function: add to Decls for typechecking
		// and compilation.
		if len(pri.implicits) != 0 && fn.OClosure == nil {
			target.Decls = append(target.Decls, fn)
		}
	}
	todoBodies = nil

	// Don't use range--typecheck can add closures to Target.Decls.
	for i := 0; i < len(target.Decls); i++ {
		target.Decls[i] = typecheck.Stmt(target.Decls[i])
	}

	// Don't use range--typecheck can add closures to Target.Decls.
	for i := 0; i < len(target.Decls); i++ {
		if fn, ok := target.Decls[i].(*ir.Func); ok {
			if base.Flag.W > 1 {
				s := fmt.Sprintf("\nbefore typecheck %v", fn)
				ir.Dump(s, fn)
			}
			ir.CurFunc = fn
			typecheck.Stmts(fn.Body)
			if base.Flag.W > 1 {
				s := fmt.Sprintf("\nafter typecheck %v", fn)
				ir.Dump(s, fn)
			}
		}
	}

	base.ExitIfErrors() // just in case

	return true
}

// writePkgStub type checks the given parsed source files and then
// returns
func writePkgStub(noders []*noder) string {
	m, pkg, info := checkFiles(noders)

	pw := newPkgWriter(m, pkg, info)

	pw.collectDecls(noders)

	publicRootWriter := pw.newWriter(relocMeta, syncPublic)
	privateRootWriter := pw.newWriter(relocMeta, syncPublic)

	assert(publicRootWriter.idx == publicRootIdx)
	assert(privateRootWriter.idx == privateRootIdx)

	{
		w := publicRootWriter
		w.pkg(pkg)
		w.bool(false) // has init; XXX

		scope := pkg.Scope()
		names := scope.Names()
		w.len(len(names))
		for _, name := range scope.Names() {
			w.obj(scope.Lookup(name), nil)
		}

		w.sync(syncEOF)
		w.flush()
	}

	{
		w := privateRootWriter
		w.ext = w
		w.pkgInit(noders)
		w.flush()
	}

	var sb bytes.Buffer // TODO(mdempsky): strings.Builder after #44505 is resolved
	pw.dump(&sb)

	// TODO(mdempsky): At this point, types2 should be available to
	// garbage collect. Figure out how it's being kept alive still and
	// then re-enable this code.
	if false {
		done := make(chan struct{})
		runtime.SetFinalizer(pkg, func(*types2.Package) { close(done) })

		runtime.GC()
		runtime.GC()
		runtime.GC()

		select {
		case <-done:
			// ok
		default:
			base.Fatalf("types2 still alive")
		}
	}

	return sb.String()
}

func readPackage(pr *pkgReader, importpkg *types.Pkg) {
	r := pr.newReader(relocMeta, publicRootIdx, syncPublic)

	pkg := r.pkg()
	assert(pkg == importpkg)

	if r.bool() {
		sym := pkg.Lookup(".inittask")
		task := ir.NewNameAt(src.NoXPos, sym)
		task.Class = ir.PEXTERN
		sym.Def = task
	}

	for i, n := 0, r.len(); i < n; i++ {
		r.sync(syncObject)
		idx := r.reloc(relocObj)
		assert(r.len() == 0)

		path, name, code, _ := r.p.peekObj(idx)
		if code != objStub {
			objReader[types.NewPkg(path, "").Lookup(name)] = pkgReaderIndex{pr, idx, nil}
		}
	}
}
