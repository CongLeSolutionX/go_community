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
)

type callSiteAnalyzer struct {
	cstab CallSiteTab
}

// encodedCallSiteTab is a table keyed by "encoded" callsite (stringified
// src.XPos plus call site ID) mapping to a value of call property bits.
type encodedCallSiteTab map[string]CSPropBits

func makeCallSiteAnalyzer(fn *ir.Func) *callSiteAnalyzer {
	return &callSiteAnalyzer{
		cstab: make(CallSiteTab),
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

func computeCallSiteTable(fn *ir.Func) CallSiteTab {
	if debugTrace != 0 {
		fmt.Fprintf(os.Stderr, "=-= making callsite table for func %v:\n",
			fn.Sym().Name)
	}
	csa := makeCallSiteAnalyzer(fn)
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
	return 0
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
	switch n.Op() {
	case ir.OCALLFUNC:
		ce := n.(*ir.CallExpr)
		callee := pgo.DirectCallee(ce.X)
		if callee != nil {
			csa.addCallSite(callee, n)
		}
	}
}

func (csa *callSiteAnalyzer) nodeVisitPost(n ir.Node) {
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
