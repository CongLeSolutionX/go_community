// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"bytes"
	"io"

	"cmd/compile/internal/base"
	"cmd/compile/internal/inline"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
)

// unifiedIRLevel controls what -G level enables unified IR.
const unifiedIRLevel = 0

var haveLegacyImports = false

// useUnifiedIR reports whether the unified IR frontend should be
// used; and if so, uses it to construct the local package's IR.
func useUnifiedIR(noders []*noder) bool {
	if base.Flag.G < unifiedIRLevel {
		return false
	}

	inline.NewInline = InlineCall

	var sb bytes.Buffer // TODO(mdempsky): strings.Builder after #44505 is resolved
	writePkgBits(&sb, noders)
	data := sb.String()

	// TODO(mdempsky): At this point, we're done with types2. Run the
	// garbage collector and use finalizers or something to make sure we
	// release its memory.

	typecheck.TypecheckAllowed = true

	assert(types.LocalPkg.Path == "")
	pr0 := newPkgDecoder(types.LocalPkg.Path, data)
	readPkgBits(pr0, types.LocalPkg, typecheck.Target)

	return true
}

func writePkgBits(data io.Writer, noders []*noder) {
	m, pkg, info := checkFiles(noders)

	pw := newPkgWriter(m, pkg, info)

	publicRootWriter := pw.newWriter(relocMeta, syncPublic)
	privateRootWriter := pw.newWriter(relocMeta, syncPublic)

	assert(publicRootWriter.idx == publicRootIdx)
	assert(privateRootWriter.idx == privateRootIdx)

	{
		// TODO(mdempsky): Turn into a b-side of the public index.
		w := privateRootWriter
		w.top = true
		w.pack(noders)
		w.flush()
	}

	{
		w := publicRootWriter
		w.pkg(pkg)
		w.bool(false) // has init; XXX

		scope := pkg.Scope()
		names := scope.Names()
		w.len(len(names))
		for _, name := range scope.Names() {
			w.rawObj(scope.Lookup(name))
		}

		w.sync(syncEOF)
		w.flush()
	}

	pw.dump(data)
}
