// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/internal/objabi"
)

// possiblyStripInstrumentation tries to strip out coverage
// instrumentation from a specific set of very hot small functions in
// the runtime package, such as the method runtime.heapBits.bits.
// Inserting counter updates into such functions can cause
// pathological performance problems in Go applications running on
// machines with high core counts, since you can get into situations
// (ex: garbage collection) where contention for counter-variable
// cache lines can almost bring things to a halt.
//
// Notes:
//
// - this is currently invoked prior to inlining (something to keep in
//
//	mind), meaning that we can strip all instrumentation from
//	function A, but if the inlined decides to inline instrumented
//	function B into the body of A, we'll wind up with some counter
//	updates regardless.
//
// - not sure if this is the best solution moving forward. other
//
//	alternatives might include cloning of counter variables during
//	inlining or introducing a "//go:nocoverage" pragma
func possiblyStripInstrumentation() {
	if base.Ctxt.Pkgpath != "runtime" {
		return
	}
	noInstrumentList := map[string]bool{
		"(*guintptr).cas": true,
		"acquireLockRank": true,
		"casgstatus":      true,
		"heapBits.bits":   true,
		"lock2":           true,
		"key32":           true,
		"nanotime":        true,
		"releaseLockRank": true,
		"runqget":         true,
		"runqput":         true,
		"scanobject":      true,
		"typedmemmove":    true,
	}
	for _, n := range typecheck.Target.Decls {
		if fn, ok := n.(*ir.Func); ok {
			if noInstrumentList[ir.FuncName(fn)] {
				stripInstrumentationFromFunc(fn)
			}
		}
	}
}

// anyChildIsStmt returns true if node "n" has any statement children
// (as opposed to no children or only expression children).
func anyChildIsStmt(n ir.Node) bool {
	hasChildStmt := false
	do := func(x ir.Node) bool {
		if _, ok := x.(ir.Stmt); ok {
			hasChildStmt = true
			return true
		}
		return false
	}
	ir.DoChildren(n, do)
	return hasChildStmt
}

// stripInstrumentationFromFunc removes code coverage instrumentation
// (updates to coverage counters) from the specified function. Any
// assignment that targets a coverage counter is replaced by an empty
// block statement. Note that this is somewhat dependent on the code
// shape produced by cmd/cover; if cmd/cover is updated to use a
// different style of instrumentation (for example, calling a builtin
// as opposed to making a direct assignment), this code will need to
// change.
func stripInstrumentationFromFunc(fn *ir.Func) {
	if len(fn.Body) == 0 {
		return
	}
	var edit func(ir.Node) ir.Node
	edit = func(n ir.Node) ir.Node {
		if anyChildIsStmt(n) {
			ir.EditChildren(n, edit)
		}
		switch n.Op() {
		case ir.OAS:
			// FIXME(thanm): common this up with the similar
			// code in the inliner.
			n := n.(*ir.AssignStmt)
			if n.X.Op() == ir.OINDEX {
				n := n.X.(*ir.IndexExpr)
				if n.X.Op() == ir.ONAME && n.X.Type().IsArray() {
					n := n.X.(*ir.Name)
					if n.Linksym().Type == objabi.SCOVERAGE_COUNTER {
						return ir.NewBlockStmt(n.Pos(), nil)
					}
				}
			}
		}
		return n
	}
	ir.EditChildren(fn, edit)
}
