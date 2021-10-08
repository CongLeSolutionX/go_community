// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"cmd/compile/internal/noder"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"fmt"
)

// importRuntimeCoveragePackage forces an import of the
// runtime/coverage package, so that we can refer to routines in it.
//
// FIXME: this seems less than ideal (since it requires reaching into
// the noder package to expose noder.ReadImportFile); perhaps there is
// a cleaner way to handle this. One possibility would be to
// predeclare the various routines (via the typecheck "builtin"
// mechanism) and go that route instead.
func importRuntimeCoveragePackage() *types.Pkg {
	path := "runtime/coverage"
	pkg, _, err := noder.ReadImportFile(path, typecheck.Target, nil, nil)
	if pkg == nil && err == nil {
		err = fmt.Errorf("noder.ReadImportFile(%s) returned nil but no error", path)
	}
	if err != nil {
		panic(fmt.Sprintf("importing runtime/coverage: %v", err))
	}
	return pkg
}
