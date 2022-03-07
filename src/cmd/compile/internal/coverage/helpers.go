// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/noder"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"fmt"
	"internal/coverage"
	"strconv"
	"strings"
)

func fixupMetaAndCounterVariables() (*ir.Name, *ir.Name, *ir.Func, coverage.CounterMode) {
	metaVarName := base.Flag.Cfg.CoverageInfo["metavar"]
	pkgIdVarName := base.Flag.Cfg.CoverageInfo["pkgidvar"]
	counterMode := base.Flag.Cfg.CoverageInfo["countermode"]
	counterPrefix := base.Flag.Cfg.CoverageInfo["counterprefix"]
	var metavar *ir.Name
	var pkgidvar *ir.Name
	var initfn *ir.Func

	ckTypSanity := func(nm *ir.Name, tag string) {
		if nm.Type() == nil || nm.Type().HasPointers() {
			base.Fatalf("unsuitable %s %q mentioned in coveragecfg, improper type", tag, nm.Sym().Name)
		}
	}

	for _, n := range typecheck.Target.Decls {
		if fn, ok := n.(*ir.Func); ok && ir.FuncName(fn) == "init" {
			if initfn != nil {
				panic("unexpected")
			}
			initfn = fn
			continue
		}
		as, ok := n.(*ir.AssignStmt)
		if !ok {
			continue
		}
		nm, ok := as.X.(*ir.Name)
		if !ok {
			continue
		}
		s := nm.Sym()
		switch s.Name {
		case metaVarName:
			metavar = nm
			ckTypSanity(nm, "metavar")
			nm.MarkReadonly()
			continue
		case pkgIdVarName:
			pkgidvar = nm
			ckTypSanity(nm, "pkgidvar")
			continue
		}
		if strings.HasPrefix(s.Name, counterPrefix) {
			ckTypSanity(nm, "countervar")
			nm.SetCoverageCounter(true)
		}
	}
	var cm coverage.CounterMode
	switch counterMode {
	case "set":
		cm = coverage.CtrModeSet
	case "count":
		cm = coverage.CtrModeCounter
	case "atomic":
		cm = coverage.CtrModeAtomic
	default:
		base.Fatalf("bad setting %q for covermode in coveragecfg:",
			counterMode)
	}
	return metavar, pkgidvar, initfn, cm
}

func metaHashAndLen() ([16]byte, int) {

	// Read meta-data hash from config entry.
	mhash := base.Flag.Cfg.CoverageInfo["metahash"]
	if len(mhash) != 32 {
		base.Fatalf("unexpected: got metahash length %d want 32", len(mhash))
	}
	var hv [16]byte
	for i := 0; i < 16; i++ {
		nib := string(mhash[i*2 : i*2+2])
		x, err := strconv.ParseInt(nib, 16, 32)
		if err != nil {
			base.Fatalf("metahash bad byte %q", nib)
		}
		hv[i] = byte(x)
	}

	// Collect meta-data length similarly.
	mls := base.Flag.Cfg.CoverageInfo["metalen"]
	hl, err := strconv.ParseInt(mls, 10, 64)
	if err != nil {
		base.Fatalf("metalen bad value %s: %v", mls, err)
	}

	return hv, int(hl)
}

func Fixup() {
	metavar, pkgIdVar, initfn, covermode := fixupMetaAndCounterVariables()
	hashv, len := metaHashAndLen()
	registerMeta(metavar, initfn, hashv, len, pkgIdVar, covermode)
}

func registerMeta(mdname *ir.Name, initfn *ir.Func, hash [16]byte, mdlen int, pkgIdVar *ir.Name, cmode coverage.CounterMode) {

	// Materialize expression for hash (an array literal)
	pos := initfn.Pos()
	elist := make([]ir.Node, 0, 16)
	for i := 0; i < 16; i++ {
		elem := ir.NewInt(int64(hash[i]))
		elist = append(elist, elem)
	}
	ht := types.NewArray(types.Types[types.TUINT8], 16)
	hashx := ir.NewCompLitExpr(pos, ir.OCOMPLIT, ir.TypeNode(ht), elist)

	// Materalize expression corresponding to address of the meta-data symbol.
	mdax := typecheck.NodAddr(mdname)
	mdauspx := typecheck.ConvNop(mdax, types.Types[types.TUNSAFEPTR])

	// Materialize expression for length.
	lenx := ir.NewInt(int64(mdlen)) // untyped

	// Generate a call to runtime.addcovmeta, e.g.
	//
	//   pkgIdVar = runtime.addcovmeta(&sym, len, hash, pkgpath, pkid, cmode)
	//
	fn := typecheck.LookupRuntime("addcovmeta")
	pkid := coverage.HardCodedPkgId(base.Ctxt.Pkgpath)
	pkIdNode := ir.NewInt(int64(pkid))
	cmodeNode := ir.NewInt(int64(cmode))
	pkPathNode := ir.NewString(base.Ctxt.Pkgpath)
	callx := typecheck.Call(pos, fn, []ir.Node{mdauspx, lenx, hashx,
		pkPathNode, pkIdNode, cmodeNode}, false)
	assign := callx
	if pkid == coverage.NotHardCoded {
		assign = typecheck.Stmt(ir.NewAssignStmt(pos, pkgIdVar, callx))
	}

	// Tack the call onto the start of our init function. We do this
	// early in the init since it's possible that instrumented function
	// bodies (with counter updates) might be inlined into init.
	initfn.Body.Prepend(assign)

	// If we are building the "main" package, then make a call into
	// the runtime to register the function 'runtime/coverage.onExitHook'
	// as a function that needs to be invoked when os.Exit() is called, e.g.
	//
	//     runOnNonZeroExit := true
	//     runtime.addExitHook(coverage.onExitHook, runOnNonZeroExit)
	//     coverage.initHook()
	//
	if base.Ctxt.Pkgpath == "main" {
		pk := importRuntimeCoveragePackage()
		typecheck.InitCoverage(pk)
		if !base.Flag.DisableCovHooks {
			regf := typecheck.LookupRuntime("addExitHook")
			hookf := typecheck.LookupCoverage(pk, "onExitHook")
			initf := typecheck.LookupCoverage(pk, "initHook")
			args := []ir.Node{hookf, ir.NewBool(true)}
			callx := typecheck.Call(pos, regf, args, false)
			initfn.Body.Append(callx)
			args = []ir.Node{}
			callx = typecheck.Call(pos, initf, args, false)
			initfn.Body.Append(callx)
		}
	}
}

// importRuntimeCoveragePackage forces an import of the
// runtime/coverage package, so that we can refer to routines in it.
//
// FIXME: this seems less than ideal (since it requires reaching into
// the noder package to expose noder.ReadImportFile); perhaps there is
// a cleaner way to handle this. One possibility would be to
// predeclare the various routines (via the typecheck "builtin"
// mechanism) and go that route instead.  My attempt at this resulted
// in problems, however (".onExitHook·f: relocation target .onExitHook
// not defined").
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
