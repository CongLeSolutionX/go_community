// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"cmd/compile/internal/base"
)

// ReassignOracle is a helper object designed to support queries about
// reassignment of ir.Names defined by short declarations (e.g.
// https://go.dev/ref/spec#Short_variable_declarations), but in a way
// that avoids walking the body of the enclosing function (as with
// ir.StaticValue and ir.Reassigned). The intended usage model is to
// create a ReassignOracle, initialize it for a given function, then
// make a series of queries using it (with the understanding that
// changing/mutating the func body IR can invalidate the info cached
// in the oracle).
type ReassignOracle struct {
	fn *Func
	ok map[*Name]Node
}

// Init initializes the oracle based on the IR in function fn, laying
// the groundwork for future calls to the StaticValue and Reassigned
// methods.
func (ro *ReassignOracle) Init(fn *Func) {
	ro.fn = fn

	// Collect candidate map. Start by adding function parameters
	// explicitly.
	ro.ok = make(map[*Name]Node)
	sig := fn.Type()
	numParams := sig.NumRecvs() + sig.NumParams()
	for _, pm := range fn.Dcl[:numParams] {
		if IsBlank(pm) {
			continue
		}
		// For params, use func itself as defining node.
		ro.ok[pm] = fn
	}

	// For locals defined via :=, walk the function body to discover.
	var findLocals func(n Node) bool
	findLocals = func(n Node) bool {
		if nn, ok := n.(*Name); ok {
			if nn.Defn != nil && !nn.Addrtaken() && nn.Class == PAUTO {
				ro.ok[nn] = nn.Defn
			}
		} else if nn, ok := n.(*ClosureExpr); ok {
			Any(nn.Func, findLocals)
		}
		return false
	}
	Any(fn, findLocals)

	outerName := func(x Node) *Name {
		if x == nil {
			return nil
		}
		n, ok := OuterValue(x).(*Name)
		if ok {
			return n.Canonical()
		}
		return nil
	}

	pruneIfNeeded := func(nn Node, asn Node) {
		oname := outerName(nn)
		if oname == nil {
			return
		}
		defn, ok := ro.ok[oname]
		if !ok {
			return
		}
		// any assignment to a param invalidates the entry.
		paramAssigned := oname.Class == PPARAM
		// assignment to local ok iff assignment is its orig def.
		localAssigned := (oname.Class == PAUTO && asn != defn)
		if paramAssigned || localAssigned {
			// We found an assignment to name N that doesn't
			// correspond to its original definition; remove
			// from candidates.
			delete(ro.ok, oname)
		}
	}

	// Prune away anything that looks assigned. This code modeled after
	// similar code in ir.Reassigned.
	var do func(n Node) bool
	do = func(n Node) bool {
		switch n.Op() {
		case OAS:
			asn := n.(*AssignStmt)
			pruneIfNeeded(asn.X, n)
		case OAS2, OAS2FUNC, OAS2MAPR, OAS2DOTTYPE, OAS2RECV, OSELRECV2:
			asn := n.(*AssignListStmt)
			for _, p := range asn.Lhs {
				pruneIfNeeded(p, n)
			}
		case OASOP:
			asn := n.(*AssignOpStmt)
			pruneIfNeeded(asn.X, n)
		case ORANGE:
			rs := n.(*RangeStmt)
			pruneIfNeeded(rs.Key, n)
			pruneIfNeeded(rs.Value, n)
		case OCLOSURE:
			n := n.(*ClosureExpr)
			Any(n.Func, do)
		}
		return false
	}
	Any(fn, do)
}

// StaticValue method has the same semantics as the ir package function
// of the same name; see comments on StaticValue for more info.
func (ro *ReassignOracle) StaticValue(n Node) Node {
	arg := n
	for {
		if n.Op() == OCONVNOP {
			n = n.(*ConvExpr).X
			continue
		}

		if n.Op() == OINLCALL {
			n = n.(*InlinedCallExpr).SingleResult()
			continue
		}

		n1 := ro.staticValue1(n)
		if n1 == nil {
			checkStaticValueResult(arg, n)
			return n
		}
		n = n1
	}
}

func (ro *ReassignOracle) staticValue1(nn Node) Node {
	if nn.Op() != ONAME {
		return nil
	}
	n := nn.(*Name).Canonical()
	if n.Class != PAUTO {
		return nil
	}

	defn := n.Defn
	if defn == nil {
		return nil
	}

	var rhs Node
FindRHS:
	switch defn.Op() {
	case OAS:
		defn := defn.(*AssignStmt)
		rhs = defn.Y
	case OAS2:
		defn := defn.(*AssignListStmt)
		for i, lhs := range defn.Lhs {
			if lhs == n {
				rhs = defn.Rhs[i]
				break FindRHS
			}
		}
		base.Fatalf("%v missing from LHS of %v", n, defn)
	default:
		return nil
	}
	if rhs == nil {
		base.Fatalf("RHS is nil: %v", defn)
	}

	if _, ok := ro.ok[n]; !ok {
		return nil
	}

	return rhs
}

// Reassigned method has the same semantics as the ir package function
// of the same name; see comments on Reassigned for more info.
func (ro *ReassignOracle) Reassigned(n *Name) bool {
	_, ok := ro.ok[n]
	result := !ok
	checkReassignedResult(n, result)
	return result
}
