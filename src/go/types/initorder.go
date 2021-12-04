// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"container/heap"
	"fmt"
	"sort"
)

// initOrder computes the Info.InitOrder for package variables.
func (check *Checker) initOrder() {
	// An InitOrder may already have been computed if a package is
	// built from several calls to (*Checker).Files. Clear it.
	check.Info.InitOrder = check.Info.InitOrder[:0]

	// Compute the object dependency graph and initialize
	// a priority queue with the list of graph nodes.
	pq := nodeQueue(dependencyGraph2(check.objMap))
	heap.Init(&pq)

	const debug = false
	if debug {
		fmt.Printf("Computing initialization order for %s\n\n", check.pkg)
		fmt.Println("Object dependency graph:")
		for obj, d := range check.objMap {
			// only print objects that may appear in the dependency graph
			if obj, _ := obj.(dependency); obj != nil {
				if len(d.deps) > 0 {
					fmt.Printf("\t%s depends on\n", obj.Name())
					for dep := range d.deps {
						fmt.Printf("\t\t%s\n", dep.Name())
					}
				} else {
					fmt.Printf("\t%s has no dependencies\n", obj.Name())
				}
			}
		}
		fmt.Println()

		fmt.Println("Transposed object dependency graph (functions eliminated):")
		for _, n := range pq {
			fmt.Printf("\t%s depends on %d nodes\n", n.obj.Name(), n.ndeps)
			for p := range n.pred {
				fmt.Printf("\t\t%s is dependent\n", p.obj.Name())
			}
		}
		fmt.Println()

		fmt.Println("Processing nodes:")
	}

	// Determine initialization order by removing the highest priority node
	// (the one with the fewest dependencies) and its edges from the graph,
	// repeatedly, until there are no nodes left.
	// In a valid Go program, those nodes always have zero dependencies (after
	// removing all incoming dependencies), otherwise there are initialization
	// cycles.
	emitted := make(map[*declInfo]bool)
	for len(pq) > 0 {
		// get the next node
		n := heap.Pop(&pq).(*graphNode)

		if debug {
			fmt.Printf("\t%s (src pos %d) depends on %d nodes now\n",
				n.obj.Name(), n.obj.order(), n.ndeps)
		}

		// if n still depends on other nodes, we have a cycle
		if n.ndeps > 0 {
			cycle := findPath(check.objMap, n.obj, n.obj, make(map[Object]bool))
			// If n.obj is not part of the cycle (e.g., n.obj->b->c->d->c),
			// cycle will be nil. Don't report anything in that case since
			// the cycle is reported when the algorithm gets to an object
			// in the cycle.
			// Furthermore, once an object in the cycle is encountered,
			// the cycle will be broken (dependency count will be reduced
			// below), and so the remaining nodes in the cycle don't trigger
			// another error (unless they are part of multiple cycles).
			if cycle != nil {
				check.reportCycle(cycle)
			}
			// Ok to continue, but the variable initialization order
			// will be incorrect at this point since it assumes no
			// cycle errors.
		}

		// reduce dependency count of all dependent nodes
		// and update priority queue
		for p := range n.pred {
			p.ndeps--
			heap.Fix(&pq, p.index)
		}

		// record the init order for variables with initializers only
		v, _ := n.obj.(*Var)
		info := check.objMap[v]
		if v == nil || !info.hasInitializer() {
			continue
		}

		// n:1 variable declarations such as: a, b = f()
		// introduce a node for each lhs variable (here: a, b);
		// but they all have the same initializer - emit only
		// one, for the first variable seen
		if emitted[info] {
			continue // initializer already emitted, if any
		}
		emitted[info] = true

		infoLhs := info.lhs // possibly nil (see declInfo.lhs field comment)
		if infoLhs == nil {
			infoLhs = []*Var{v}
		}
		init := &Initializer{infoLhs, info.init}
		check.Info.InitOrder = append(check.Info.InitOrder, init)
	}

	if debug {
		fmt.Println()
		fmt.Println("Initialization order:")
		for _, init := range check.Info.InitOrder {
			fmt.Printf("\t%s\n", init)
		}
		fmt.Println()
	}
}

// findPath returns the (reversed) list of objects []Object{to, ... from}
// such that there is a path of object dependencies from 'from' to 'to'.
// If there is no such path, the result is nil.
func findPath(objMap map[Object]*declInfo, from, to Object, seen map[Object]bool) []Object {
	if seen[from] {
		return nil
	}
	seen[from] = true

	for d := range objMap[from].deps {
		if d == to {
			return []Object{d}
		}
		if P := findPath(objMap, d, to, seen); P != nil {
			return append(P, d)
		}
	}

	return nil
}

// reportCycle reports an error for the given cycle.
func (check *Checker) reportCycle(cycle []Object) {
	obj := cycle[0]
	check.errorf(obj, _InvalidInitCycle, "initialization cycle for %s", obj.Name())
	// subtle loop: print cycle[i] for i = 0, n-1, n-2, ... 1 for len(cycle) = n
	for i := len(cycle) - 1; i >= 0; i-- {
		check.errorf(obj, _InvalidInitCycle, "\t%s refers to", obj.Name()) // secondary error, \t indented
		obj = cycle[i]
	}
	// print cycle[0] again to close the cycle
	check.errorf(obj, _InvalidInitCycle, "\t%s", obj.Name())
}

// ----------------------------------------------------------------------------
// Object dependency graph

// A dependency is an object that may be a dependency in an initialization
// expression. Only constants, variables, and functions can be dependencies.
// Constants are here because constant expression cycles are reported during
// initialization order computation.
type dependency interface {
	Object
	isDependency()
}

// A graphNode represents a node in the object dependency graph.
// Each node p in n.pred represents an edge p->n, and each node
// s in n.succ represents an edge n->s; with a->b indicating that
// a depends on b.
type graphNode struct {
	obj        dependency // object represented by this node
	pred, succ nodeSet    // consumers and dependencies of this node (lazily initialized)
	index      int        // node index in graph slice/priority queue
	ndeps      int        // number of outstanding dependencies before this object can be initialized
}

// cost returns the cost of contracting this edge, which involves copying each
// predecessor to each successor (and vice-versa).
func (n *graphNode) cost() int {
	return len(n.pred) * len(n.succ)
}

type nodeSet map[*graphNode]bool

func (s *nodeSet) add(p *graphNode) {
	if *s == nil {
		*s = make(nodeSet)
	}
	(*s)[p] = true
}

func (s nodeSet) String() string {
	var buf bytes.Buffer
	buf.WriteRune('{')
	var keys []string
	for n := range s {
		keys = append(keys, n.obj.Name())
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(k)
	}
	buf.WriteRune('}')
	return buf.String()
}

// dependencyGraph2 is like dependencyGraph, but sorts functions by cost before
// contracting them in the graph. As contracted function edges get recursively
// re-written, it is significantly faster to start with functions that have few
// predecessors or successors.
//
// TODO(rfindley): remove dependencyGraph in a follow-up CL.
func dependencyGraph2(objMap map[Object]*declInfo) []*graphNode {
	M := make(map[dependency]*graphNode)

	var G, funcG []*graphNode // separate non-functions and functions

	for obj := range objMap {
		// only consider nodes that may be an initialization dependency
		if obj, _ := obj.(dependency); obj != nil {
			n := &graphNode{obj: obj}
			M[obj] = n
			if _, ok := obj.(*Func); ok {
				funcG = append(funcG, n)
			} else {
				G = append(G, n)
			}
		}
	}

	// Compute edges for graph G. We need to include all nodes, even isolated
	// ones, because they still need to be scheduled for initialization in
	// correct order relative to other nodes.
	for _, n := range G {
		for d := range objMap[n.obj].deps {
			if d, _ := d.(dependency); d != nil {
				d := M[d]
				d.pred.add(n)
				n.succ.add(d)
			}
		}
	}

	// Repeat for functions, ignoring self-cycles as they would need to be
	// contracted below anyway.
	for _, n := range funcG {
		for d := range objMap[n.obj].deps {
			if d, _ := d.(dependency); d != nil && d != n.obj { // ignore self-cycles
				d := M[d]
				d.pred.add(n)
				n.succ.add(d)
			}
		}
	}

	// Contract function edges. Note that because we recursively copy
	// predecessors and successors throughout the function graph, the cost of
	// contracting a function at position X is proportional to cost *
	// (len(funcG)-X). Therefore, we should contract high-cost functions last.
	sort.Slice(funcG, func(i, j int) bool {
		return funcG[i].cost() < funcG[j].cost()
	})
	for _, n := range funcG {
		for d := range n.succ {
			for p := range n.pred {
				p.succ.add(d)
				d.pred.add(p)
			}
			delete(d.pred, n) // remove edge to n
		}
		for p := range n.pred {
			delete(p.succ, n) // remove edge to n
		}
	}

	// fill in index and ndeps fields
	for i, n := range G {
		n.index = i
		n.ndeps = len(n.succ)
	}

	if debug {
		// Compare with the original algorithm for correctness.
		orig := dependencyGraph(objMap)
		if len(orig) != len(G) {
			panic(fmt.Sprintf("len(dependencyGraph) = %d, len(dependencyGraph2) = %d", len(orig), len(G)))
		}
		for _, g1 := range orig {
			g2 := M[g1.obj]
			if g2 == nil {
				panic(fmt.Sprintf("missing %s", g1.obj.Name()))
			}
			if g1.ndeps != g2.ndeps {
				panic(fmt.Sprintf("g1[%s].succ = %s, g2[%s].succ = %s", g1.obj.Name(), g1.succ, g2.obj.Name(), g2.succ))
			}
		}
	}

	return G
}

// dependencyGraph computes the object dependency graph from the given objMap,
// with any function nodes removed. The resulting graph contains only constants
// and variables.
func dependencyGraph(objMap map[Object]*declInfo) []*graphNode {
	// M is the dependency (Object) -> graphNode mapping
	M := make(map[dependency]*graphNode)
	for obj := range objMap {
		// only consider nodes that may be an initialization dependency
		if obj, _ := obj.(dependency); obj != nil {
			M[obj] = &graphNode{obj: obj}
		}
	}

	// compute edges for graph M
	// (We need to include all nodes, even isolated ones, because they still need
	// to be scheduled for initialization in correct order relative to other nodes.)
	for obj, n := range M {
		// for each dependency obj -> d (= deps[i]), create graph edges n->s and s->n
		for d := range objMap[obj].deps {
			// only consider nodes that may be an initialization dependency
			if d, _ := d.(dependency); d != nil {
				d := M[d]
				n.succ.add(d)
				d.pred.add(n)
			}
		}
	}

	// remove function nodes and collect remaining graph nodes in G
	// (Mutually recursive functions may introduce cycles among themselves
	// which are permitted. Yet such cycles may incorrectly inflate the dependency
	// count for variables which in turn may not get scheduled for initialization
	// in correct order.)
	var G []*graphNode
	for obj, n := range M {
		if _, ok := obj.(*Func); ok {
			// connect each predecessor p of n with each successor s
			// and drop the function node (don't collect it in G)
			for p := range n.pred {
				// ignore self-cycles
				if p != n {
					// Each successor s of n becomes a successor of p, and
					// each predecessor p of n becomes a predecessor of s.
					for s := range n.succ {
						// ignore self-cycles
						if s != n {
							p.succ.add(s)
							s.pred.add(p)
							delete(s.pred, n) // remove edge to n
						}
					}
					delete(p.succ, n) // remove edge to n
				}
			}
		} else {
			// collect non-function nodes
			G = append(G, n)
		}
	}

	// fill in index and ndeps fields
	for i, n := range G {
		n.index = i
		n.ndeps = len(n.succ)
	}

	return G
}

// ----------------------------------------------------------------------------
// Priority queue

// nodeQueue implements the container/heap interface;
// a nodeQueue may be used as a priority queue.
type nodeQueue []*graphNode

func (a nodeQueue) Len() int { return len(a) }

func (a nodeQueue) Swap(i, j int) {
	x, y := a[i], a[j]
	a[i], a[j] = y, x
	x.index, y.index = j, i
}

func (a nodeQueue) Less(i, j int) bool {
	x, y := a[i], a[j]
	// nodes are prioritized by number of incoming dependencies (1st key)
	// and source order (2nd key)
	return x.ndeps < y.ndeps || x.ndeps == y.ndeps && x.obj.order() < y.obj.order()
}

func (a *nodeQueue) Push(x interface{}) {
	panic("unreachable")
}

func (a *nodeQueue) Pop() interface{} {
	n := len(*a)
	x := (*a)[n-1]
	x.index = -1 // for safety
	*a = (*a)[:n-1]
	return x
}
