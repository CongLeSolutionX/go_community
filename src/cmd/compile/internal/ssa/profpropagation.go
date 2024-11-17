// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/base"
	"fmt"
)

// edge represents a CFG edge, which has source and sink basic blocks.
// The Edge struct in file block.go does not carry two basic blocks.
type edge struct {
	source *Block
	sink   *Block
}

var sampleProfileMaxPropagateIterations = 100

// Checks whether the incoming edge is visited or not.
func visitEdge(e *edge, numUnknownEdges *int, unknownEdge **edge, visitedEdges map[string]bool, edgeWeights map[string]int) int {
	s := getHashString(e.source, e.sink)
	sRev := getHashString(e.sink, e.source)
	if visitedEdges[s] == false && visitedEdges[sRev] == false {
		*numUnknownEdges = *numUnknownEdges + 1
		*unknownEdge = e
		return 0
	} else {
		if visitedEdges[s] {
			edgeWeights[sRev] = edgeWeights[s]
		} else {
			edgeWeights[s] = edgeWeights[sRev]
		}
		visitedEdges[s] = true
		visitedEdges[sRev] = true
		return edgeWeights[s]
	}
}

// encodeUint32 returns the encoding string given the input value.
func encodeUint32(i uint32) string {
	// Max value is 4294967295.
	buf := make([]byte, 10)
	for b, d := buf, uint32(1000000000); d > 0; d /= 10 {
		b[0] = byte(i/d%10 + '0')
		if b[0] == '0' && len(b) == len(buf) && len(buf) > 1 {
			buf = buf[1:]
		}
		b = b[1:]
		i %= d
	}
	return string(buf)
}

// getHashString returns the concat string for the IDs of two basic blocks.
func getHashString(source *Block, sink *Block) string {
	return encodeUint32(uint32(source.ID)) + "#" + encodeUint32(uint32(sink.ID))
}

// propagateThroughEdges propagates weights through incoming/outgoing edges.
//
// If the weight of a basic block is known, and there is only one edge
// with an unknown weight, we can calculate the weight of that edge.
//
// Similarly, if all the edges have a known count, we can calculate the
// count of the basic block, if needed.
//
// param F  Function to process.
// param UpdateBlockCount  Whether we should update basic block counts that
// has already been annotated.
//
// returns  True if new weights were assigned to edges or blocks.
func propagateThroughEdges(f *Func, updateBlockCount bool, visitedEdges map[string]bool, edgeWeights map[string]int, visitedBlocks map[*Block]bool) bool {

	if f.pass.debug > 2 {
		fmt.Printf("\nPropagation through edges\n")
	}

	var changed = false
	for _, b := range f.Blocks {
		// Visit all the predecessor and successor edges to determine
		// which ones have a weight assigned already. Note that it doesn't
		// matter that we only keep track of a single unknown edge. The
		// only case we are interested in handling is when only a single
		// edge is unknown.
		for i := 0; i < 2; i++ {
			var totalWeight int
			var numUnknownEdges int
			var unknownEdge, selfReferentialEdge *edge
			var neibors []Edge
			if i == 0 {
				neibors = b.Preds
			} else {
				neibors = b.Succs
			}
			var e *edge
			for _, n := range neibors {
				e = &edge{n.b, b}
				totalWeight = totalWeight + visitEdge(e, &numUnknownEdges, &unknownEdge, visitedEdges, edgeWeights)
				if e.source == e.sink {
					selfReferentialEdge = e
				}
			}

			// After visiting all the edges, there are four cases that we
			// can handle immediately:
			//
			// - All the edge weights are known (i.e., NumUnknownEdges == 0),
			//   and the BB has been visited. In this case, depending on
			//   the UpdateBlockCount flag, we will either:
			//   UpdateBlockCount == true: adjust the block weight to match
			//                             the total edge weights, but only
			//							   if it's increasing the BB weight.
			//   UpdateBlockCount == false: adjust the edge weights to match
			//                             the block weight
			//
			// - Only one edge is unknown and BB has already been visited.
			//   In this case, we can compute the weight of the edge by
			//   subtracting the total block weight from all the known
			//   edge weights. If the edges weight more than BB, then the
			//   edge of the last remaining edge is set to zero.
			//
			// - There exists a self-referential edge and the weight of BB is
			//   known. In this case, this edge can be based on BB's weight.
			//   We add up all the other known edges and set the weight on
			//   the self-referential edge as we did in the previous case.
			//
			// - The BB weight is known and is 0, then all edges are set to 0.
			//
			// Additionally, at the end of the iteration if
			// UpdateBlockCount == true, we will update the unvisited BB weights
			// to the total edge weight, and mark it as visited.
			//
			// We will continue iterating. Eventually,
			// all edges will get a weight and stablize, or iteration will stop when
			// it reaches sampleProfileMaxPropagateIterations.

			if numUnknownEdges <= 1 {
				if numUnknownEdges == 0 {
					if totalWeight != b.BBFreq && visitedBlocks[b] {
						changed = true
						if updateBlockCount {
							if totalWeight > b.BBFreq {
								b.BBFreq = totalWeight
							}
						} else {
							fraction := float64(b.BBFreq) / float64(totalWeight)
							for _, n := range neibors {
								edge := getHashString(b, n.b)
								edgeWeights[edge] = int(fraction * float64(edgeWeights[edge]))
							}
						}
					}
				} else if numUnknownEdges == 1 {
					if _, ok := visitedBlocks[b]; ok {
						changed = true
						S := getHashString(unknownEdge.source, unknownEdge.sink)
						if b.BBFreq >= totalWeight {
							edgeWeights[S] = b.BBFreq - totalWeight
						} else {
							edgeWeights[S] = 0
						}
						var neibor *Block
						if i == 0 {
							neibor = unknownEdge.source
						} else {
							neibor = unknownEdge.sink
						}
						if _, ok := visitedBlocks[neibor]; ok {
							if edgeWeights[S] > neibor.BBFreq {
								edgeWeights[S] = neibor.BBFreq
							}
						}
						visitedEdges[S] = true
						if f.pass.debug > 2 {
							fmt.Printf("Set weight for edge: weight[b%d->b%d]: %v\n", unknownEdge.source.ID, unknownEdge.sink.ID, edgeWeights[S])
						}
					}
				}
			} else if selfReferentialEdge != nil && selfReferentialEdge.source != nil {
				if _, ok := visitedBlocks[b]; ok {
					S := getHashString(selfReferentialEdge.source, selfReferentialEdge.sink)
					oldEW := edgeWeights[S]
					oldVE := visitedEdges[S]
					if b.BBFreq >= totalWeight {
						edgeWeights[S] = b.BBFreq - totalWeight
					} else {
						edgeWeights[S] = 0
					}
					visitedEdges[S] = true
					changed = (oldEW != edgeWeights[S]) || (oldVE != true)
				}
			} else if b.BBFreq == 0 {
				if _, ok := visitedBlocks[b]; ok {
					for _, e := range neibors {
						var E *edge = &edge{e.b, b}
						S := getHashString(E.source, E.sink)
						oldEW := edgeWeights[S]
						oldVE := visitedEdges[S]
						edgeWeights[S] = 0
						visitedEdges[S] = true
						changed = changed || (oldEW != 0) || (oldVE != true)
					}
				}
			}

			if updateBlockCount && totalWeight > 0 {
				if _, ok := visitedBlocks[b]; !ok {
					b.BBFreq = totalWeight
					visitedBlocks[b] = true
					changed = true
					if f.pass.debug > 2 {
						fmt.Printf("Set weight for b%d: %v\n", b.ID, totalWeight)
					}
				}
			}
		}
	}
	return changed
}

// propagateWeights propagates weights into edges based on heuristics:
//
// It will first aggresively propagate weights by allowing unvisited block weights to be changed.
//
// Then a conservative propagation that constrains on visited blocks will be performed to tune
// the weights.
//
// Finally, another aggresive propagation.

func propagateWeights(f *Func, visitedBlocks map[*Block]bool) {
	edgeWeights := make(map[string]int)
	visitedEdges := make(map[string]bool)

	var changed bool = false

	for i := 0; i < sampleProfileMaxPropagateIterations; i++ {
		changed = propagateThroughEdges(f, true, visitedEdges, edgeWeights, visitedBlocks)
		if changed == false {
			break
		}
	}

	visitedEdges = make(map[string]bool)
	for i := 0; i < sampleProfileMaxPropagateIterations; i++ {
		changed = propagateThroughEdges(f, false, visitedEdges, edgeWeights, visitedBlocks)
		if changed == false {
			break
		}
	}

	for i := 0; i < sampleProfileMaxPropagateIterations; i++ {
		changed = propagateThroughEdges(f, true, visitedEdges, edgeWeights, visitedBlocks)
		if changed == false {
			break
		}
	}

	updateCfgEdgeWeights(f, edgeWeights)
}

// getEdgeWeight returns the computed edge weight if it exists.
func getEdgeWeight(source *Block, sink *Block, EdgeWeights map[string]int) int {
	if val, ok := EdgeWeights[getHashString(source, sink)]; ok {
		return val
	} else {
		return 0
	}
}

// updateCfgEdgeWeights updates the edge weight in the control flow graph according
// to the computation in the frequency propagation.
func updateCfgEdgeWeights(f *Func, EdgeWeights map[string]int) {
	for _, b := range f.Blocks {
		for i := 0; i < 2; i++ {
			if i == 0 {
				for i, e := range b.Preds {
					b.Preds[i].EdgeFreq = getEdgeWeight(e.b, b, EdgeWeights)
				}
			} else {
				for i, e := range b.Succs {
					b.Succs[i].EdgeFreq = getEdgeWeight(b, e.b, EdgeWeights)
				}
			}
		}
	}
}

// initialVisited marks blocks with frequency, additionally also
// mark exit blocks. Prof propagation will starts from these blocks.
func initialVisited(f *Func, visitedBlocks map[*Block]bool) bool {
	changed := false
	for _, b := range f.Blocks {
		if b.BBFreq != 0 {
			visitedBlocks[b] = true
			changed = true
		}
	}
	if changed {
		for _, b := range f.Blocks {
			if b.BBFreq == 0 &&
				(b.Kind == BlockExit || b.Kind == BlockRet || b.Kind == BlockRetJmp || b.Kind == BlockDefer) {
				visitedBlocks[b] = true
			}
		}
	}
	return changed
}

// freqPropagate propagates block weights into edges. This uses some simple
// propagation heuristics. The following rules are applied to every
// block BB in the CFG. See propagateThroughEdges comments for details.
// Block weights will be updated too.
//
// Since this propagation is not guaranteed to finalize for every CFG, we
// only allow it to proceed for a limited number of iterations.

func freqPropagate(f *Func) {
	if base.Debug.PGOBBReorder != 2 || f.fe.Func() == nil || f.fe.Func().LSym == nil {
		// For functions missing linker symbols, the sample mapping won't work.
		return
	}
	visitedBlocks := make(map[*Block]bool)
	var changed = initialVisited(f, visitedBlocks)
	if changed == true {
		propagateWeights(f, visitedBlocks)
	}
}
