// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lockgraph provides a representation of a lock class graph
// and algorithms for working with it.
//
// Cycles in a lock class graph indicate potential deadlocks.
package lockgraph

import (
	"strings"

	"github.com/aclements/go-moremath/graph"
	"github.com/aclements/go-moremath/graph/graphalg"
)

// Graph is a graph of lock classes. Each node represents a lock class
// and each edge indicates that the target lock class was acquired
// while the source lock class was held.
//
// If there are cycles in the lock graph, this represents the
// potential for deadlock.
//
// Each node is labeled with its lock class name. Each edge stores the
// set of stacks at which the source and target lock classes were
// acquired (in this sense, this is a multi-graph).
//
// Graph satisfies the graph.Graph interface.
//
// Graph can be serialized to and from JSON.
type Graph struct {
	Labels []string // Node ID -> node label
	Edges  [][]Edge // Node ID -> Edge number -> Edge info
	To     [][]int  // Node ID -> Edge number -> Target node ID

	StackTable *StackTable
}

// Edge records extra information about a single edge in the lock
// graph.
type Edge struct {
	// Stacks records all acquire/acquire stack pairs on this
	// edge.
	Stacks []StackPair
}

// A LockEdge represents a single edge between two lock classes in a
// lock graph.
type StackPair struct {
	// S1 and S2 are the stacks at which the first and second
	// locks on this edge were acquired.
	S1, S2 Stack

	// Count is the number of times this edge has been observed.
	Count int
}

func (g *Graph) NumNodes() int {
	return len(g.Labels)
}

func (g *Graph) Out(i int) []int {
	return g.To[i]
}

func (g *Graph) Label(i int) string {
	const strip = "runtime."
	label := g.Labels[i]
	if strings.HasPrefix(label, strip) {
		label = label[len(strip):]
	}
	return label
}

// Cycles returns the nodes and edges involved in cycles of the graph.
func Cycles(g graph.Graph) (nodes []int, edges []graph.Edge) {
	// Nodes involved in cycles are those that are in non-trivial
	// connected components.
	scc := graphalg.SCC(g, graphalg.SCCSubnodeComponent)
	marks := graphalg.NewNodeMarks()
	numNodes := 0
	for cid := 0; cid < scc.NumNodes(); cid++ {
		nids := scc.Subnodes(cid)
		if len(nids) <= 1 {
			continue
		}
		numNodes += len(nids)
		for _, nid := range nids {
			marks.Mark(nid)
		}
	}

	// Create nodes slice and compute a close upper-bound on the
	// number of edges.
	nodes = make([]int, 0, numNodes)
	numEdges := 0
	for nid := marks.Next(-1); nid >= 0; nid = marks.Next(nid) {
		nodes = append(nodes, nid)
		for n2id := range g.Out(nid) {
			if marks.Test(n2id) {
				numEdges++
			}
		}
	}

	// Create edges.
	edges = make([]graph.Edge, 0, numEdges)
	for _, nid := range nodes {
		cid := scc.SubnodeComponent(nid)
		for eid, n2id := range g.Out(nid) {
			c2id := scc.SubnodeComponent(n2id)
			if cid == c2id {
				edges = append(edges, graph.Edge{nid, eid})
			}
		}
	}

	return
}

// Filter filters g to a subset of nodes and edges. It panics if any
// included edge points to an excluded node.
func (g *Graph) Filter(cNodes []int, cEdges []graph.Edge) *Graph {
	// Create new nodes and old-to-new ID mapping.
	labels := make([]string, len(cNodes))
	oldToNew := make(map[int]int, len(cNodes))
	for newID, oldID := range cNodes {
		labels[newID] = g.Labels[oldID]
		oldToNew[oldID] = newID
	}

	// Create new edges.
	edges := make([][]Edge, len(cNodes))
	out := make([][]int, len(cNodes))
	for _, oldEdge := range cEdges {
		newNode, ok := oldToNew[oldEdge.Node]
		if !ok {
			panic("cannot include edge from excluded node")
		}
		newNode2, ok := oldToNew[g.To[oldEdge.Node][oldEdge.Edge]]
		if !ok {
			panic("cannot include edge to excluded node")
		}
		out[newNode] = append(out[newNode], newNode2)
		edges[newNode] = append(edges[newNode], g.Edges[oldEdge.Node][oldEdge.Edge])
	}

	return &Graph{labels, edges, out, g.StackTable}
}
