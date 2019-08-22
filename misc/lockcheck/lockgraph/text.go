// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lockgraph

import (
	"fmt"
	"io"
	"strings"

	"github.com/aclements/go-moremath/graph/graphalg"
)

// ReportText writes a user-readable text representation of lg to w.
func ReportText(w io.Writer, lg *Graph) {
	// Walk the edges in rough depth-first order so that cycles
	// naturally appear together in the report.
	scc := graphalg.SCC(lg, 0)

	// SCC components are in reverse topo order. Walk them in topo
	// order.
	visited := graphalg.NewNodeMarks()
	var visit func(node int)
	visit = func(node int) {
		visited.Mark(node)
		for i, succ := range lg.Out(node) {
			reportTextEdge(w, lg, node, succ, lg.Edges[node][i])
			if !visited.Test(succ) {
				visit(succ)
			}
		}
	}
	inComponent := make([]int, lg.NumNodes()) // Max CID of parent nodes
	for cid := scc.NumNodes() - 1; cid >= 0; cid-- {
		subGraph := scc.Subnodes(cid)
		// Find node in this subgraph with the maximal
		// in-component as a good starting point in this
		// component and update in-components.
		maxNode, maxCID := 0, -1
		for _, nid := range subGraph {
			if inComponent[nid] > maxCID {
				maxNode, maxCID = nid, cid
			}
			for _, nid2 := range lg.Out(nid) {
				if cid > inComponent[nid2] {
					inComponent[nid2] = cid
				}
			}
		}
		// Walk this subgraph starting from maxNode.
		visit(maxNode)
	}
}

func reportTextEdge(w io.Writer, lg *Graph, n1, n2 int, edge Edge) {
	fmt.Fprintf(w, "%s -> %s\n", lg.Label(n1), lg.Label(n2))
	for i, pair := range edge.Stacks {
		if i >= 5 {
			fmt.Fprintf(w, "  .. %d more stack pairs ..\n", len(edge.Stacks)-i)
			break
		}

		fmt.Fprintf(w, "  acquired %s %d times at:\n", lg.Label(n1), pair.Count)
		fmt.Fprintf(w, "%s\n", indent("    ", lg.StackTable.StringStack(pair.S1)))
		fmt.Fprintf(w, "  then %s at:\n", lg.Label(n2))
		fmt.Fprintf(w, "%s\n", indent("    ", lg.StackTable.StringStack(pair.S2)))
	}
	fmt.Fprintf(w, "\n")
}

func indent(by string, s string) string {
	return by + strings.Replace(s, "\n", "\n"+by, -1)
}
