// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"fmt"
	"strings"
)

// ToDot returns a Graphviz dot description of the network.
func (c *CNet) ToDot() string {
	return layersToDot(c.layers)
}

func layersToDot(layers []layer) string {
	var o strings.Builder
	fmt.Fprintf(&o, "digraph { ranksep=2;\n")
	// Draw layers from bottom up so that we don't have to pre-declare nodes.
	for i := len(layers) - 1; i >= 0; i-- {
		layer := layers[i]
		fmt.Fprintf(&o, "  subgraph cluster_%d {\n", i)
		label := fmt.Sprintf("layer %d shift %d span %s", i, layer.shift, layer.heapSpan)
		fmt.Fprintf(&o, "    label=%q;\n", label)
		for bufI, buf := range layer.buffers {
			fmt.Fprintf(&o, "    n%d_%d", i, bufI)
			var label string
			if i == 0 {
				label = fmt.Sprintf("P %d", bufI)
			} else {
				//label = heap.Range{heap.VAddr(buf.start), layer.heapSpan}.ShortString()
				label = fmt.Sprintf("[%s,%s)", buf.start, buf.start.Plus(layer.heapSpan))
			}
			fmt.Fprintf(&o, " [label=%q];\n", label)

			if i < len(layers)-1 {
				fmt.Fprintf(&o, "    n%d_%d", i, bufI)
				for j := range layer.fanOut {
					to := layer.topo[bufI] + j
					if j == 0 {
						o.WriteString(" ->")
					} else {
						o.WriteByte(',')
					}
					fmt.Fprintf(&o, " n%d_%d", i+1, to)
				}
				fmt.Fprintf(&o, ";\n")
			}
		}

		// For nodes to appear in left-to-right order. (This doesn't *entirely*
		// work and I have no idea why.)
		fmt.Fprintf(&o, "  { rank=same; rankdir=LR; edge [style=invis; weight=100];\n")
		for bufI := range layer.buffers {
			if bufI == 0 {
				o.WriteString("    ")
			} else {
				o.WriteString(" -> ")

			}
			fmt.Fprintf(&o, "n%d_%d", i, bufI)
		}
		o.WriteString(";\n")
		o.WriteString("  }\n")

		fmt.Fprintf(&o, "  }\n")
	}

	fmt.Fprintf(&o, "}\n")
	return o.String()
}
