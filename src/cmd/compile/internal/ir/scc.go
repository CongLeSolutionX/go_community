// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
)

// Strongly connected components.
//
// Run analysis on minimal sets of mutually recursive functions
// or single non-recursive functions, bottom up.
//
// Finding these sets is finding strongly connected components
// by reverse topological order in the static call graph.
// The algorithm (known as Tarjan's algorithm) for doing that is taken from
// Sedgewick, Algorithms, Second Edition, p. 482, with two adaptations.
//
// First, a hidden closure function (n.Func.IsHiddenClosure()) cannot be the
// root of a connected component. Refusing to use it as a root
// forces it into the component of the function in which it appears.
// This is more convenient for escape analysis.
//
// Second, each function becomes two virtual nodes in the graph,
// with numbers n and n+1. We record the function's node number as n
// but search from node n+1. If the search tells us that the component
// number (min) is n+1, we know that this is a trivial component: one function
// plus its closures. If the search tells us that the component number is
// n, then there was a path from node n+1 back to node n, meaning that
// the function set is mutually recursive. The escape analysis can be
// more precise when analyzing a single non-recursive function than
// when analyzing a set of mutually recursive functions.

type bottomUpVisitor struct {
	analyze  func([]*Func, bool)
	visitgen uint32
	nodeID   map[*Func]uint32
	stack    []*Func
	typeID   map[*types.Type]uint32
	types    []*types.Type
}

// VisitFuncsBottomUp invokes analyze on the ODCLFUNC nodes listed in list.
// It calls analyze with successive groups of functions, working from
// the bottom of the call graph upward. Each time analyze is called with
// a list of functions, every function on that list only calls other functions
// on the list or functions that have been passed in previous invocations of
// analyze. Closures appear in the same list as their outer functions.
// The lists are as short as possible while preserving those requirements.
// (In a typical program, many invocations of analyze will be passed just
// a single function.) The boolean argument 'recursive' passed to analyze
// specifies whether the functions on the list are mutually recursive.
// If recursive is false, the list consists of only a single function and its closures.
// If recursive is true, the list may still contain only a single function,
// if that function is itself recursive.
func VisitFuncsBottomUp(list []*Func, analyze func(list []*Func, recursive bool)) {
	visitFuncsBottomUp(list, analyze, false)
}

// VisitFuncsBottomUpForEscapeAnalysis is like VisitFuncsBottomUp, but
// it also considers conversions of values from concrete type to
// interface type as an implicit call of all methods in that concrete
// type's method set.
func VisitFuncsBottomUpForEscapeAnalysis(list []*Func, analyze func(list []*Func, recursive bool)) {
	visitFuncsBottomUp(list, analyze, true)
}

func visitFuncsBottomUp(list []*Func, analyze func(list []*Func, recursive bool), forEscapeAnalysis bool) {
	var v bottomUpVisitor
	v.analyze = analyze
	v.nodeID = make(map[*Func]uint32)
	if forEscapeAnalysis {
		v.typeID = make(map[*types.Type]uint32)
	}
	for _, n := range list {
		if !n.IsHiddenClosure() {
			v.visitFunc(n)
		}
	}

	if len(v.stack) != 0 || len(v.types) != 0 {
		base.Fatalf("unexpected residuals: %v, %v", v.stack, v.types)
	}
}

func (v *bottomUpVisitor) visitFunc(n *Func) uint32 {
	if id := v.nodeID[n]; id > 0 {
		// already visited
		return id
	}

	v.visitgen++
	id := v.visitgen
	v.nodeID[n] = id
	v.visitgen++
	min := v.visitgen
	v.stack = append(v.stack, n)

	do := func(defn Node) {
		if defn != nil {
			if m := v.visitFunc(defn.(*Func)); m < min {
				min = m
			}
		}
	}

	Visit(n, func(n Node) {
		switch n.Op() {
		case ONAME:
			if n := n.(*Name); n.Class == PFUNC {
				do(n.Defn)
			}
		case ODOTMETH, OMETHVALUE, OMETHEXPR:
			if fn := MethodExprName(n); fn != nil {
				do(fn.Defn)
			}
		case OCLOSURE:
			n := n.(*ClosureExpr)
			do(n.Func)
		case OCONVIFACE:
			if v.typeID != nil {
				n := n.(*ConvExpr)
				if m := v.visitType(n.X.Type()); m < min {
					min = m
				}
			}
		}
	})

	if (min == id || min == id+1) && !n.IsHiddenClosure() {
		// This node is the root of a strongly connected component.

		// The original min was id+1. If the bottomUpVisitor found its way
		// back to id, then this block is a set of mutually recursive functions.
		// Otherwise, it's just a lone function that does not recurse.
		recursive := min == id

		var i int

		// Pop types whose ID is *greater* than ours, and reset their ID
		// to a large number. A greater ID means the type was found within
		// our component.
		for i = len(v.types); i > 0; i-- {
			typ := v.types[i-1]
			if v.typeID[typ] <= id {
				break
			}
			v.typeID[typ] = ^uint32(0)
		}
		v.types = v.types[:i]

		// Remove connected component from stack and mark v.nodeID so that future
		// visits return a large number, which will not affect the caller's min.
		for i = len(v.stack) - 1; i >= 0; i-- {
			x := v.stack[i]
			v.nodeID[x] = ^uint32(0)
			if x == n {
				break
			}
		}
		block := v.stack[i:]
		// Call analyze on this set of functions.
		v.stack = v.stack[:i]
		v.analyze(block, recursive)
	}

	return min
}

func (v *bottomUpVisitor) visitType(typ *types.Type) uint32 {
	mt := types.ReceiverBaseType(typ)
	if mt == nil {
		return ^uint32(0) // no possible methods
	}

	if id := v.typeID[typ]; id > 0 {
		return id // previously visited
	}

	// Note: We don't reserve a unique ID for typ itself, so it will end
	// up with the same ID as the *next* newly discovered function.
	min := v.visitgen
	v.typeID[typ] = min
	v.types = append(v.types, typ)

	types.CalcMethods(mt)
	for _, method := range mt.AllMethods() {
		// We want to visit all concrete methods in typ's method set,
		// which may be called through an interface.
		if method.Nointerface() || types.IsInterfaceMethod(method.Type) || !types.IsMethodApplicable(typ, method) {
			continue
		}
		if fn := method.Nname.(*Name); fn.Defn != nil {
			if m := v.visitFunc(fn.Defn.(*Func)); m < min {
				min = m
			}
		}
	}

	return min
}
