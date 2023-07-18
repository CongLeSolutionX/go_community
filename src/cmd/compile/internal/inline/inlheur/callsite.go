// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/pgo"
	"cmd/internal/src"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// TODO:
// - add "in init function" heuristic
// - fill in panic path stuff
// - in separate CL: skeleton framework for callsites,
//   dump mechanism, and unit tests.

type CallSite struct {
	Callee *ir.Func
	Call   ir.Node
	Flags  CSPropBits
	Score  int
	Id     uint
}

// CallSiteTab is a table of call sites, keyed by call ir.Node.
// Ideally it would be nice to key the table by src.XPos, but
// this results in collisions for calls on very long lines (the
// front end saturates column numbers at 255). We also wind up
// with many calls that share the same auto-generated pos.
type CallSiteTab map[ir.Node]CallSite

// Package-level table of callsites.
var cstab = CallSiteTab{}

type CSPropBits uint32

const (
	CallSiteInLoop CSPropBits = 1 << iota
	CallSiteOnPanicPath
	CallSiteInInitFunc
)

type callSiteAnalyzer struct {
	cstab    CallSiteTab
	ptab     map[ir.Node]pstate
	nstack   []ir.Node
	loopNest int
	isInit   bool
}

// encodedCallSiteTab is a table keyed by "encoded" callsite (stringified
// src.XPos plus call site ID) mapping to a value of call property bits.
type encodedCallSiteTab map[string]CSPropBits

func makeCallSiteAnalyzer(fn *ir.Func, ptab map[ir.Node]pstate) *callSiteAnalyzer {
	isInit := fn.IsPackageInit() || strings.HasPrefix(fn.Sym().Name, "init.")
	return &callSiteAnalyzer{
		cstab:  make(CallSiteTab),
		ptab:   ptab,
		isInit: isInit,
	}
}

func (cst CallSiteTab) merge(other CallSiteTab) error {
	for k, v := range other {
		if prev, ok := cst[k]; ok {
			return fmt.Errorf("internal error: collision during call site table merge, fn=%s callsite=%s", prev.Callee.Sym().Name, fmtFullPos(prev.Call.Pos()))
		}
		cst[k] = v
	}
	return nil
}

func computeCallSiteTable(fn *ir.Func, ptab map[ir.Node]pstate) CallSiteTab {
	if debugTrace != 0 {
		fmt.Fprintf(os.Stderr, "=-= making callsite table for func %v:\n",
			fn.Sym().Name)
	}
	csa := makeCallSiteAnalyzer(fn, ptab)
	var doNode func(ir.Node) bool
	doNode = func(n ir.Node) bool {
		csa.nodeVisitPre(n)
		ir.DoChildren(n, doNode)
		csa.nodeVisitPost(n)
		return false
	}
	doNode(fn)
	return csa.cstab
}

func (csa *callSiteAnalyzer) flagsForNode(call ir.Node) CSPropBits {
	var r CSPropBits

	if debugTrace&debugTraceCalls != 0 {
		fmt.Fprintf(os.Stderr, "=-= analyzing call at %s\n",
			fmtFullPos(call.Pos()))
	}

	// Set a bit if this call is within a loop.
	if csa.loopNest > 0 {
		r |= CallSiteInLoop
	}

	// Set a bit if the call is within an init function (either
	// compiler-generated or user-written).
	if csa.isInit {
		r |= CallSiteInInitFunc
	}

	// Try to determine if this call is on a panic path. Do this by
	// walking back up the node stack to see if we can find either A)
	// an enclosing panic, or B) a statement node that we've
	// determined leads to a panic/exit.
	for ri := range csa.nstack[:len(csa.nstack)-1] {
		i := len(csa.nstack) - ri - 1
		n := csa.nstack[i]
		_, isCallExpr := n.(*ir.CallExpr)
		_, isStmt := n.(ir.Stmt)
		if isCallExpr {
			isStmt = false
		}

		if debugTrace&debugTraceCalls != 0 {
			ps, inps := csa.ptab[n]
			fmt.Fprintf(os.Stderr, "=-= callpar %d op=%s ps=%s inptab=%v stmt=%v\n", i, n.Op().String(), ps.String(), inps, isStmt)
		}

		if n.Op() == ir.OPANIC {
			r |= CallSiteOnPanicPath
			break
		}
		if v, ok := csa.ptab[n]; ok {
			if v == psCallsPanic {
				r |= CallSiteOnPanicPath
				break
			}
			if isStmt {
				break
			}
		}
	}

	return r
}

func fmtFullPos(p src.XPos) string {
	var sb strings.Builder
	sep := ""
	base.Ctxt.AllPos(p, func(pos src.Pos) {
		fmt.Fprintf(&sb, sep)
		sep = "|"
		file := filepath.Base(pos.Filename())
		fmt.Fprintf(&sb, "%s:%d:%d", file, pos.Line(), pos.Col())
	})
	return sb.String()
}

func (csa *callSiteAnalyzer) addCallSite(callee *ir.Func, call ir.Node) {
	cs := CallSite{
		Call:   call,
		Callee: callee,
		Flags:  csa.flagsForNode(call),
		Id:     uint(len(csa.cstab)),
	}
	if _, ok := csa.cstab[call]; ok {
		fmt.Fprintf(os.Stderr, "*** cstab duplicate entry at: %s\n",
			fmtFullPos(call.Pos()))
		fmt.Fprintf(os.Stderr, "*** call: %+v\n", call)
		panic("bad")
	}
	csa.cstab[call] = cs
}

func (csa *callSiteAnalyzer) nodeVisitPre(n ir.Node) {
	csa.nstack = append(csa.nstack, n)
	switch n.Op() {
	case ir.ORANGE, ir.OFOR:
		csa.loopNest++
	case ir.OCALLFUNC:
		ce := n.(*ir.CallExpr)
		callee := pgo.DirectCallee(ce.X)
		if callee != nil && callee.Inl != nil {
			csa.addCallSite(callee, n)
		}
	}
}

func (csa *callSiteAnalyzer) nodeVisitPost(n ir.Node) {
	csa.nstack = csa.nstack[:len(csa.nstack)-1]
	switch n.Op() {
	case ir.ORANGE, ir.OFOR:
		csa.loopNest--
	}
}

func encodeCallSiteKey(cs *CallSite) string {
	var sb strings.Builder
	// FIXME: rewrite line offsets relative to function start
	sb.WriteString(fmtFullPos(cs.Call.Pos()))
	fmt.Fprintf(&sb, "|%d", cs.Id)
	return sb.String()
}

func buildEncodedCallSiteTab(tab CallSiteTab) encodedCallSiteTab {
	r := make(encodedCallSiteTab)
	for _, cs := range tab {
		k := encodeCallSiteKey(&cs)
		r[k] = cs.Flags
	}
	return r
}

func dumpCallSiteComments(w io.Writer, tab CallSiteTab, ecst encodedCallSiteTab) {
	if ecst == nil {
		ecst = buildEncodedCallSiteTab(tab)
	}
	tags := make([]string, 0, len(ecst))
	for k := range ecst {
		tags = append(tags, k)
	}
	sort.Strings(tags)
	for _, s := range tags {
		v := ecst[s]
		fmt.Fprintf(w, "// callsite: %s %q %d\n", s, v.String(), v)
	}
	fmt.Fprintf(w, "// %s\n", csDelimiter)
}
