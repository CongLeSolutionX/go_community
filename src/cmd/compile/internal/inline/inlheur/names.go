// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/ir"
	"go/constant"
	"go/token"
)

type nameFinder struct {
	ro *ir.ReassignOracle
}

func makeNameFinder(fn *ir.Func) *nameFinder {
	var ro *ir.ReassignOracle
	if fn != nil {
		ro = &ir.ReassignOracle{}
		ro.Init(fn)
	}
	return &nameFinder{ro: ro}
}

// isFuncName returns the *ir.Name for the func or method
// corresponding to node 'n', along with a boolean indicating success,
// and another boolean indicating whether the func is a closure.
func (nf *nameFinder) isFuncName(n ir.Node) (*ir.Name, bool, bool) {
	sv := n
	if nf.ro != nil {
		sv = nf.ro.StaticValue(n)
	}
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

// isAllocatedMem returns true if node n corresponds to a memory
// allocation expression (make, new, or equivalent).
func (nf *nameFinder) isAllocatedMem(n ir.Node) bool {
	sv := n
	if nf.ro != nil {
		sv = nf.ro.StaticValue(n)
	}
	switch sv.Op() {
	case ir.OMAKESLICE, ir.ONEW, ir.OPTRLIT, ir.OSLICELIT:
		return true
	}
	return false
}

// isLiterate returns whether n is a constant value or nil (or singly
// assigned local containing nil/literal). First return is the literal value
// (or nil if value is nil) and second value is TRUE if the first
// value is valid.
func (nf *nameFinder) isLiteral(n ir.Node) (constant.Value, bool) {
	sv := n
	if nf.ro != nil {
		sv = nf.ro.StaticValue(n)
	}
	switch sv.Op() {
	case ir.ONIL:
		return nil, true
	case ir.OLITERAL:
		return sv.Val(), true
	}
	return nil, false
}

func (nf *nameFinder) staticValue(n ir.Node) ir.Node {
	if nf.ro == nil {
		return n
	}
	return nf.ro.StaticValue(n)
}

func (nf *nameFinder) reassigned(n *ir.Name) bool {
	if nf.ro == nil {
		return true
	}
	return nf.ro.Reassigned(n)
}

func (nf *nameFinder) isConcreteConvIface(n ir.Node) bool {
	sv := n
	if nf.ro != nil {
		sv = nf.ro.StaticValue(n)
	}
	if sv.Op() != ir.OCONVIFACE {
		return false
	}
	return !sv.(*ir.ConvExpr).X.Type().IsInterface()
}

// isSameLiteral checks to see if 'v1' and 'v2' correspond to the same
// literal value, or if they are both nil.
func isSameLiteral(v1, v2 constant.Value) bool {
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil || v2 == nil {
		return false
	}
	return constant.Compare(v1, token.EQL, v2)
}

func isSameFuncName(v1, v2 *ir.Name) bool {
	// NB: there are a few corner cases where pointer equality
	// doesn't work here, but this should be good enough for
	// our purposes here.
	return v1 == v2
}
