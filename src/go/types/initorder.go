// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"container/heap"
	"fmt"
)

// initOrder computes the Info.InitOrder for package variables.
func (check *Checker) initOrder() {
	// An InitOrder may already have been computed if a package is
	// built from several calls to (*Checker).Files. Clear it.
	check.Info.InitOrder = check.Info.InitOrder[:0]

	// Compute the object dependency graph and initialize
	// a priority queue with the list of graph nodes.
	pq := nodeQueue(dependencyGraph(check.objMap))
	heap.Init(&pq)

	const debug = false
	if debug {
		fmt.Printf("Computing initialization order for %s\n\n", check.pkg)
		fmt.Println("Object dependency graph:")
		for obj, d := range check.objMap {
			if len(d.deps) > 0 {
				fmt.Printf("\t%s depends on\n", obj.Name())
				for dep := range d.deps {
					fmt.Printf("\t\t%s\n", dep.Name())
				}
			} else {
				fmt.Printf("\t%s has no dependencies\n", obj.Name())
			}
		}
		fmt.Println()

		fmt.Println("Transposed object dependency graph:")
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
			if i := valIndex(cycle); i >= 0 {
				check.reportCycle(cycle, i)
			}
			// ok to continue, but the variable initialization order
			// will be incorrect at this point since it assumes no
			// cycle errors
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
func findPath(objMap map[Object]*declInfo, from, to Object, visited map[Object]bool) []Object {
	if visited[from] {
		return nil // node already seen
	}
	visited[from] = true

	for d := range objMap[from].deps {
		if d == to {
			return []Object{d}
		}
		if P := findPath(objMap, d, to, visited); P != nil {
			return append(P, d)
		}
	}

	return nil
}

// valIndex returns the index of the first constant or variable in a,
// if any; or a value < 0.
// TODO(gri): If we eliminate all both Const and Var nodes from
// the dependency graph, then we can eliminate this function.
func valIndex(a []Object) int {
	for i, n := range a {
		switch n.(type) {
		case *Const, *Var:
			return i
		}
	}
	return -1
}

// reportCycle reports an error for the cycle starting at i.
func (check *Checker) reportCycle(cycle []Object, i int) {
	obj := cycle[i]
	check.errorf(obj.Pos(), "initialization cycle for %s", obj.Name())
	// print cycle
	for _ = range cycle {
		check.errorf(obj.Pos(), "\t%s refers to", obj.Name()) // secondary error, \t indented
		// go backward since cycle is in reverse order
		// TODO(gri) If we can get rid of valIndex, the, the start
		// index here will be 0 and we can simplify this code.
		if i == 0 {
			i = len(cycle)
		}
		i--
		obj = cycle[i]
	}
	check.errorf(obj.Pos(), "\t%s", obj.Name())
}

// ----------------------------------------------------------------------------
// Object dependency graph

// An graphNode represents a node in the object dependency graph.
// Each node p in n.pred represents an edge p->n, and each node
// s in n.succ represents an edge n->s; with a->b indicating that
// a depends on b.
type graphNode struct {
	obj        Object  // object represented by this node
	pred, succ nodeSet // consumers and dependencies of this node (lazily initialized)
	index      int     // node index in graph slice/priority queue
	ndeps      int     // number of outstanding dependencies before this object can be initialized
}

type nodeSet map[*graphNode]bool

func (n *graphNode) addPred(p *graphNode) {
	m := n.pred
	if m == nil {
		m = make(nodeSet)
		n.pred = m
	}
	m[p] = true
}

func (n *graphNode) addSucc(s *graphNode) {
	m := n.succ
	if m == nil {
		m = make(nodeSet)
		n.succ = m
	}
	m[s] = true
}

// dependencyGraph computes the object dependency graph from the given objMap,
// with any function nodes removed.
func dependencyGraph(objMap map[Object]*declInfo) []*graphNode {
	// M is the Object -> graphNode mapping
	M := make(map[Object]*graphNode, len(objMap))
	for obj := range objMap {
		M[obj] = &graphNode{obj: obj}
	}

	// build full graph
	for obj, n := range M {
		for d := range objMap[obj].deps {
			d := M[d]
			n.addSucc(d)
			d.addPred(n)
		}
	}

	// remove function nodes
	for obj, n := range M {
		if _, ok := obj.(*Func); ok {
			// connect each predecessor p of n with each successor s
			for p := range n.pred {
				// ignore self-cycles
				if p != n {
					// each successor s of n becomes a successor of p
					for s := range n.succ {
						if s != n {
							p.addSucc(s)
							s.addPred(p)
							delete(s.pred, n) // ?
						}
					}
					// remove edges to n
					delete(p.succ, n)
				}
			}
			delete(M, obj) // remove function
		}
	}

	// build new graph G without functions
	G := make([]*graphNode, len(M))
	i := 0
	for _, n := range M {
		n.ndeps = len(n.succ)
		n.index = i
		G[i] = n
		i++
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
