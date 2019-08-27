// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lockgraph

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/aclements/go-moremath/graph/graphout"
)

var missingPos = SrcPos{Error: "missing stack"}

// EdgeToDot constructs a dot description of the stacks on a single
// edge of the lock graph.
func (g *Graph) EdgeToDot(w io.Writer, e Edge) {
	fmt.Fprintf(w, "digraph \"\" {\n")

	// Accumulate all of the stacks to put in the call graph.
	goexit := g.StackTable.Funcs.Add("runtime.goexit")
	tweakStack := func(s Stack) Stack {
		// Strip runtime.goexit.
		if len(s) > 0 && s[len(s)-1].Func == goexit {
			s = s[:len(s)-1]
		}
		// If the stack is empty, replace with one "missing" frame.
		if len(s) == 0 {
			s = Stack{missingPos}
		}
		return s
	}
	var stacks []Stack
	for _, pair := range e.Stacks {
		stacks = append(stacks, tweakStack(pair.S1))
		stacks = append(stacks, tweakStack(pair.S2))
	}

	// Emit call graph.
	noder := funcNoder{g.StackTable}
	posNodes := dotCallGraph(w, stacks, "n", noder)

	// Accumulate acquire/acquire edges on just the inner-most
	// source position.
	type edgeIDs struct{ n1, n2 string }
	weights := map[edgeIDs]int{}
	for _, pair := range e.Stacks {
		p1 := noder.key(tweakStack(pair.S1)[0])
		p2 := noder.key(tweakStack(pair.S2)[0])
		weights[edgeIDs{posNodes[p1], posNodes[p2]}] += pair.Count
	}

	// Emit acquire/acquire edges.
	const maxWidth = 8
	const minWidth = 2
	maxWeight := 0
	for _, weight := range weights {
		if weight > maxWeight {
			maxWeight = weight
		}
	}
	for edge, weight := range weights {
		// TODO: Maybe add a node on each side that shows
		// which lock is being acquired and at exactly what
		// line? Right now the edge graph doesn't show you
		// that, so it's easy to forget what the locks are.
		width := float64(weight)/float64(maxWeight)*(maxWidth-minWidth) + minWidth
		fmt.Fprintf(w, "  %s -> %s [label=%d, color=red, penwidth=%f];\n", edge.n1, edge.n2, weight, width)
	}

	fmt.Fprintf(w, "}\n")
}

// A srcPosNoder maps SrcPoses to nodes for a call graph.
type srcPosNoder interface {
	// key returns the node key for pos.
	key(pos SrcPos) interface{}
	// label returns the node label for pos.
	label(pos SrcPos) string
}

// lineNoder creates nodes for every unique source line.
type lineNoder struct {
	stackTable *StackTable
}

func (f lineNoder) key(pos SrcPos) interface{} {
	return pos
}

func (f lineNoder) label(pos SrcPos) string {
	if pos.Error != "" {
		return pos.Error
	}
	fun := f.stackTable.Funcs.Get(pos.Func)
	file := f.stackTable.Files.Get(pos.File)
	return fmt.Sprintf("%s\n%s:%d", fun, filepath.Base(file), pos.Line)
}

// funcNoder creates nodes only for unique functions.
type funcNoder struct {
	stackTable *StackTable
}

func (f funcNoder) key(pos SrcPos) interface{} {
	return pos.Func
}

func (f funcNoder) label(pos SrcPos) string {
	if pos.Error != "" {
		return pos.Error
	}
	return f.stackTable.Funcs.Get(pos.Func)
}

// dotCallGraph emits dot code for the call graph in stks.
//
// All node IDs will be prefixed with nodePrefix. The noder argument
// determines how positions are mapped to nodes and how nodes are
// labeled. It returns a map from noder's keys to node IDs.
func dotCallGraph(w io.Writer, stks []Stack, nodePrefix string, noder srcPosNoder) map[interface{}]string {
	nodes := map[interface{}]string{}
	nid := 0

	// Emit nodes.
	for _, stk := range stks {
		for _, pos := range stk {
			key := noder.key(pos)
			if _, ok := nodes[key]; ok {
				continue
			}
			node := fmt.Sprintf("%s%d", nodePrefix, nid)
			label := noder.label(pos)
			fmt.Fprintf(w, "  %s [label=%s];\n", node, graphout.DotString(label))
			nodes[key] = node
			nid++
		}
	}
	// Emit edges.
	type edgeIDs struct{ n1, n2 interface{} }
	haveEdges := map[edgeIDs]bool{}
	for _, stk := range stks {
		for i := 0; i < len(stk)-1; i++ {
			n1, n2 := noder.key(stk[i+1]), noder.key(stk[i])
			if haveEdges[edgeIDs{n1, n2}] {
				continue
			}
			haveEdges[edgeIDs{n1, n2}] = true
			fmt.Fprintf(w, "  %s -> %s;\n", nodes[n1], nodes[n2])
		}
	}

	return nodes
}
