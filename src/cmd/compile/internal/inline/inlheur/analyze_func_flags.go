// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"fmt"
	"os"
)

// funcFlagsAnalyzer computes the "Flags" value for the FuncProps
// object we're computing. The main item of interest here is "nstate",
// which stores the disposition of a given ir Node with respect to the
// flags/properties we're trying to compute.
type funcFlagsAnalyzer struct {
	fn     *ir.Func
	nstate map[ir.Node]pstate
	noInfo bool // set if we see something inscrutable/un-analyzable
}

// pstate keeps track of the disposition of a given node and its
// children with respect to panic/exit calls.
type pstate int

const (
	psNoInfo     pstate = iota // nothing interesting about this node
	psCallsPanic               // node causes call to panic or os.Exit
	psMayReturn                // executing node may trigger a "return" stmt
	psTop                      // dataflow lattice "top" element
)

func makeFuncFlagsAnalyzer(fn *ir.Func) *funcFlagsAnalyzer {
	return &funcFlagsAnalyzer{
		fn:     fn,
		nstate: make(map[ir.Node]pstate),
	}
}

// results returns the final results from the analysis in the form of
// a value of type FuncPropBits.
func (ffa *funcFlagsAnalyzer) results() FuncPropBits {
	var rv FuncPropBits
	if !ffa.noInfo && ffa.dolist(ffa.fn.Body) == psCallsPanic {
		rv = FuncPropNeverReturns
	}
	// This is slightly hacky, but include a special case for main.main,
	// which often ends in a call to os.Exit.
	if isMainMain(ffa.fn) {
		rv &^= FuncPropNeverReturns
	}
	return rv
}

func (ffa *funcFlagsAnalyzer) getstate(n ir.Node) pstate {
	val, ok := ffa.nstate[n]
	if !ok {
		base.Fatalf("funcFlagsAnalyzer: fn %q node %s line %s: internal error, no setting for node:\n%+v\n", ffa.fn.Sym().Name, n.Op().String(), ir.Line(n), n)
	}
	return val
}

func (ffa *funcFlagsAnalyzer) setstate(n ir.Node, st pstate) {
	if _, ok := ffa.nstate[n]; ok {
		base.Fatalf("funcFlagsAnalyzer: fn %q internal error, existing setting for node:\n%+v\n", ffa.fn.Sym().Name, n)
	} else {
		ffa.nstate[n] = st
	}
}

func (ffa *funcFlagsAnalyzer) setstateSoft(n ir.Node, st pstate) {
	ffa.nstate[n] = st
}

// blockCombine merges together states as part of a linear sequence of
// statements, where 'succ' is a statement that appears immediately
// after 'pred'. Examples:
//
//	case 1:             case 2:
//	    panic("foo")      if q { return x }        <-pred
//	    return x          panic("boo")             <-succ
func blockCombine(succ, pred pstate) pstate {
	switch succ {
	case psTop:
		return pred
	case psMayReturn:
		if pred == psCallsPanic {
			return psCallsPanic
		}
		return psMayReturn
	case psNoInfo:
		return pred
	case psCallsPanic:
		if pred == psMayReturn {
			return psMayReturn
		}
		return psCallsPanic
	}
	panic("should never execute")
}

// branchCombine combines two states at a control flow branch point where
// either p1 or p2 executes (as in an "if" statement).
func branchCombine(p1, p2 pstate) pstate {
	if p1 == psCallsPanic && p2 == psCallsPanic {
		return psCallsPanic
	}
	if p1 == psMayReturn || p2 == psMayReturn {
		return psMayReturn
	}
	return psNoInfo
}

// dolist walks through a list of statements and computes the
// state/diposition for the entire list as a whole.
func (ffa *funcFlagsAnalyzer) dolist(list ir.Nodes) pstate {
	st := psTop
	ll := len(list)
	for k := range list {
		i := ll - k - 1
		n := list[i]
		psi := ffa.getstate(n)
		if debugTrace&debugTraceFuncFlags != 0 {
			fmt.Fprintf(os.Stderr, "=-= %v: dolist n=%s ps=%s\n",
				ir.Line(n), n.Op().String(), psi.String())
		}
		st = blockCombine(st, psi)
	}
	if st == psTop {
		st = psNoInfo
	}
	return st
}

func isMainMain(fn *ir.Func) bool {
	s := fn.Sym()
	if s == nil {
		return false
	}
	return (s.Pkg.Name == "main" && s.Name == "main")
}

func isWellKnownFunc(s *types.Sym, pkg, name string) bool {
	return (s.Pkg.Path == pkg ||
		(s.Pkg == types.LocalPkg && types.LocalPkg.Path == pkg)) &&
		s.Name == name
}

// isFuncName returns the *ir.Name for the func or method
// corresponding to node 'n', along with a boolean indicating success,
// and another boolean indicating whether the func is closure.
func isFuncName(n ir.Node) (*ir.Name, bool, bool) {
	sv := ir.StaticValue(n)
	if sv.Op() == ir.ONAME {
		name := sv.(*ir.Name)
		if name.Sym() != nil && name.Class == ir.PFUNC {
			return name, true, false
		}
	}
	if sv.Op() == ir.OCLOSURE {
		cloex := sv.(*ir.ClosureExpr)
		return cloex.Func.Nname, true, true
	}
	if sv.Op() == ir.OMETHEXPR {
		if mn := ir.MethodExprName(sv); mn != nil {
			return mn, true, false
		}
	}
	return nil, false, false
}

// isPanicLike returns TRUE if the node itself is an unconditional
// call to os.Exit(), a panic, or a function that does likewise.
func isPanicLike(n ir.Node) bool {
	if n.Op() != ir.OCALLFUNC {
		return false
	}
	cx := n.(*ir.CallExpr)
	name, isFunc, _ := isFuncName(cx.X)
	if !isFunc {
		return false
	}
	s := name.Sym()
	if isWellKnownFunc(s, "os", "Exit") ||
		isWellKnownFunc(s, "runtime", "throw") {
		return true
	}
	// FIXME: consult results of flags computation for
	// previously analyzer Go functions, including props
	// read from export data for functions in other packages.
	return false
}

// pessimize is called to record the fact that we saw something in the
// function that renders it entirely impossible to analyze.
func (ffa *funcFlagsAnalyzer) pessimize() {
	ffa.noInfo = true
}

// shouldVisit returns TRUE if this is an interesting node from the
// perspective of computing function flags. NB: due to the fact that
// ir.CallExpr implements the Stmt interface, we wind up visiting
// a lot of nodes that we don't really need to, but these can
// simply be screened out as part of the visit.
func shouldVisit(n ir.Node) bool {
	_, isStmt := n.(ir.Stmt)
	return n.Op() != ir.ODCL &&
		(isStmt || n.Op() == ir.OCALLFUNC || n.Op() == ir.OPANIC)
}

// nodeVisitPost helps implement the propAnalyzer interface; when
// called on a given node, it decides the disposition of that node
// based on the state(s) of the node's children.
func (ffa *funcFlagsAnalyzer) nodeVisitPost(n ir.Node) {
	if debugTrace&debugTraceFuncFlags != 0 {
		fmt.Fprintf(os.Stderr, "=+= nodevis %v %s should=%v\n",
			ir.Line(n), n.Op().String(), shouldVisit(n))
	}
	if !shouldVisit(n) {
		// invoke soft set, since node may be shared (e.g. ONAME)
		ffa.setstateSoft(n, psNoInfo)
		return
	}
	var st pstate
	switch n.Op() {
	case ir.OCALLFUNC:
		if isPanicLike(n) {
			st = psCallsPanic
		}
	case ir.OPANIC:
		st = psCallsPanic
	case ir.OBREAK, ir.ORETURN, ir.OCONTINUE:
		st = psMayReturn
	case ir.OBLOCK:
		blst := n.(*ir.BlockStmt)
		st = ffa.dolist(blst.List)
	case ir.OCASE:
		if ccst, ok := n.(*ir.CaseClause); ok {
			st = ffa.dolist(ccst.Body)
		} else if ccst, ok := n.(*ir.CommClause); ok {
			st = ffa.dolist(ccst.Body)
		} else {
			panic("unexpected")
		}
	case ir.OIF:
		ifst := n.(*ir.IfStmt)
		bst := ffa.dolist(ifst.Body)
		if bst == psMayReturn {
			st = psMayReturn
		} else if len(ifst.Else) == 0 {
			st = psNoInfo
		} else {
			est := ffa.dolist(ifst.Else)
			if est == psCallsPanic && bst == psCallsPanic {
				st = psCallsPanic
			}
		}
	case ir.OFOR:
		// Treat for { XXX } like a block.
		// Treat for <cond> { XXX } like an if statement with no else.
		forst := n.(*ir.ForStmt)
		bst := ffa.dolist(forst.Body)
		if forst.Cond == nil {
			st = bst
		} else {
			if bst == psMayReturn {
				st = psMayReturn
			}
		}
	case ir.ORANGE:
		// Treat for { XXX } like a block.
		// Treat for <cond> { XXX } like an if statement with no else.
		rst := n.(*ir.RangeStmt)
		if ffa.dolist(rst.Body) == psMayReturn {
			st = psMayReturn
		}
	case ir.OGOTO:
		// punt if we see even one goto. if we built a control
		// flow graph we could do more, but this is just a tree walk.
		ffa.pessimize()
	case ir.OSELECT:
		// process selects for "may return" but not "always panics",
		// the latter case seems very improbable.
		sst := n.(*ir.SelectStmt)
		for _, c := range sst.Cases {
			if ffa.dolist(c.Body) == psMayReturn {
				st = psMayReturn
				break
			}
		}
	case ir.OSWITCH:
		sst := n.(*ir.SwitchStmt)
		if len(sst.Cases) != 0 {
			st = psTop
			for _, c := range sst.Cases {
				st = branchCombine(ffa.dolist(c.Body), st)
			}
		}
	case ir.OFALL:
		// Not important.
	case ir.ODCLFUNC, ir.ORECOVER, ir.OAS, ir.OAS2, ir.OAS2FUNC, ir.OASOP,
		ir.OPRINTN, ir.OPRINT, ir.OLABEL, ir.OCALLINTER, ir.ODEFER,
		ir.OSEND, ir.ORECV, ir.OSELRECV2, ir.OGO, ir.OAPPEND, ir.OAS2DOTTYPE,
		ir.OAS2MAPR, ir.OGETG, ir.ODELETE, ir.OINLMARK, ir.OAS2RECV,
		ir.OMIN, ir.OMAX, ir.OMAKE, ir.ORECOVERFP, ir.OGETCALLERSP:
		// these should all be benign/uninteresting
	case ir.OTAILCALL, ir.OJUMPTABLE, ir.OTYPESW:
		// don't expect to see these at all.
		base.Fatalf("unexpected op %s in func %s",
			n.Op().String(), ir.FuncName(ffa.fn))
	default:
		base.Fatalf("%v: unhandled op %s in func %v",
			ir.Line(n), n.Op().String(), ir.FuncName(ffa.fn))
	}
	if debugTrace&debugTraceFuncFlags != 0 {
		fmt.Fprintf(os.Stderr, "=-= %v: visit n=%s returns %s\n",
			ir.Line(n), n.Op().String(), st.String())
	}
	ffa.setstate(n, st)
}

func (ffa *funcFlagsAnalyzer) nodeVisitPre(n ir.Node) {
}
