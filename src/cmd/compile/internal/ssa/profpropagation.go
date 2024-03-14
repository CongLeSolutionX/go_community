// Copyright 2016 The Go Authors. All rights reserved.
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

type parentBlock struct {
	b      *Block
	parent ID // parent in dominator tree.  0 = no parent (entry or unreachable)
}

var SampleProfileMaxPropagateIterations = 100

// Checks whether the incoming edge is visited or not.
func visitEdge(E *edge, NumUnknownEdges *int, UnknownEdge **edge, VisitedEdges map[string]*edge, EdgeWeights map[string]int64) int64 {
	var S string
	S = getHashString(E.source, E.sink)
	if _, ok := VisitedEdges[S]; !ok {
		*NumUnknownEdges = *NumUnknownEdges + 1
		*UnknownEdge = E
		return 0
	} else {
		return EdgeWeights[S]
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
func propagateThroughEdges(f *Func, UpdateBlockCount bool, VisitedEdges map[string]*edge, EdgeWeights map[string]int64, VisitedBlocks map[*Block]bool, blocks []parentBlock) bool {

	if f.pass.debug > 2 {
		fmt.Printf("\nPropagation through edges\n")
	}

	var Changed = false
	for _, b := range f.Blocks {
		pid := blocks[b.ID].parent
		if pid != 0 && b.BBFreq.RawCount != 0 && blocks[pid].b.BBFreq.RawCount == 0 {
			blocks[pid].b.BBFreq.RawCount = b.BBFreq.RawCount
			Changed = true
		}

		// Visit all the predecessor and successor edges to determine
		// which ones have a weight assigned already. Note that it doesn't
		// matter that we only keep track of a single unknown edge. The
		// only case we are interested in handling is when only a single
		// edge is unknown.
		for i := 0; i < 2; i++ {
			var TotalWeight int64 = 0
			var NumUnknownEdges = 0
			var NumTotalEdges = 0
			var UnknownEdge, SelfReferentialEdge *edge
			var SingleEdge string
			var Edge [2]string
			SelfReferentialEdge = nil
			Edge[0] = ""
			Edge[1] = ""
			if i == 0 {
				NumTotalEdges = len(b.Preds)
				var E *edge
				for _, e := range b.Preds {
					E = &edge{e.b, b}
					TotalWeight = TotalWeight + visitEdge(E, &NumUnknownEdges, &UnknownEdge, VisitedEdges, EdgeWeights)
					if E.source == E.sink {
						SelfReferentialEdge = E
					}
				}
				if NumTotalEdges == 1 {
					var S string
					S = getHashString(E.source, E.sink)
					SingleEdge = S
				} else if NumTotalEdges == 2 && SelfReferentialEdge == nil {
					var E *edge
					cnt := 0
					for _, e := range b.Preds {
						E = &edge{e.b, b}
						var S string
						S = getHashString(E.source, E.sink)
						Edge[cnt] = S
						cnt = cnt + 1
					}
				}
			} else {
				NumTotalEdges = len(b.Succs)
				var E *edge
				for _, e := range b.Succs {
					E = &edge{b, e.b}
					TotalWeight = TotalWeight + visitEdge(E, &NumUnknownEdges, &UnknownEdge, VisitedEdges, EdgeWeights)
				}
				if NumTotalEdges == 1 {
					var S string
					S = getHashString(E.source, E.sink)
					SingleEdge = S
				} else if NumTotalEdges == 2 && SelfReferentialEdge == nil {
					var E *edge
					cnt := 0
					for _, e := range b.Succs {
						E = &edge{b, e.b}
						var S string
						S = getHashString(E.source, E.sink)
						Edge[cnt] = S
						cnt = cnt + 1
					}
				}
			}

			// After visiting all the edges, there are three cases that we
			// can handle immediately:
			//
			// - All the edge weights are known (i.e., NumUnknownEdges == 0).
			//   In this case, we simply check that the sum of all the edges
			//   is the same as BB's weight. If not, we change BB's weight
			//   to match. Additionally, if BB had not been visited before,
			//   we mark it visited.
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
			// In any other case, we must continue iterating. Eventually,
			// all edges will get a weight, or iteration will stop when
			// it reaches SampleProfileMaxPropagateIterations.

			if NumUnknownEdges <= 1 {
				if NumUnknownEdges == 0 {
					if TotalWeight > b.BBFreq.RawCount {
						b.BBFreq.RawCount = TotalWeight
						Changed = true
					} else if NumTotalEdges == 1 && EdgeWeights[SingleEdge] < b.BBFreq.RawCount {
						EdgeWeights[SingleEdge] = b.BBFreq.RawCount
						Changed = true
					} else if NumTotalEdges == 2 && Edge[0] != "" && Edge[1] != "" {
						if EdgeWeights[Edge[0]] == 0 && b.BBFreq.RawCount > EdgeWeights[Edge[1]] {
							EdgeWeights[Edge[1]] = b.BBFreq.RawCount
							Changed = true
						}
						if EdgeWeights[Edge[1]] == 0 && b.BBFreq.RawCount > EdgeWeights[Edge[0]] {
							EdgeWeights[Edge[0]] = b.BBFreq.RawCount
							Changed = true
						}
					}
				} else if NumUnknownEdges == 1 {
					if _, ok := VisitedBlocks[b]; ok {
						var S string
						S = getHashString(UnknownEdge.source, UnknownEdge.sink)
						if b.BBFreq.RawCount >= TotalWeight {
							EdgeWeights[S] = b.BBFreq.RawCount - TotalWeight
						} else {
							EdgeWeights[S] = 0
						}
						var OtherEC *Block
						if i == 0 {
							OtherEC = UnknownEdge.source
						} else {
							OtherEC = UnknownEdge.sink
						}
						if _, ok := VisitedBlocks[OtherEC]; ok {
							if EdgeWeights[S] > OtherEC.BBFreq.RawCount {
								EdgeWeights[S] = OtherEC.BBFreq.RawCount
							}
						}
						VisitedEdges[S] = UnknownEdge
						Changed = true
						for I := 0; I < len(UnknownEdge.source.Succs); I++ {
							if UnknownEdge.source.Succs[I].b == UnknownEdge.sink {
								UnknownEdge.source.Succs[I].EdgeFreq.RawCount = EdgeWeights[S]
								break
							}
						}
						for I := 0; I < len(UnknownEdge.sink.Preds); I++ {
							if UnknownEdge.sink.Preds[I].b == UnknownEdge.source {
								UnknownEdge.sink.Preds[I].EdgeFreq.RawCount = EdgeWeights[S]
								break
							}
						}
						if f.pass.debug > 2 {
							fmt.Printf("Set weight for edge: weight[b%d->b%d]: %v\n", UnknownEdge.source.ID, UnknownEdge.sink.ID, EdgeWeights[S])
						}
					}
				}
			} else if b.BBFreq.RawCount == 0 {
				if _, ok := VisitedBlocks[b]; ok {
					if i == 0 {
						for _, e := range b.Preds {
							var E *edge = &edge{e.b, b}
							var S string
							S = getHashString(E.source, E.sink)
							EdgeWeights[S] = 0
							VisitedEdges[S] = E
						}
					} else {
						for _, e := range b.Succs {
							var E *edge = &edge{e.b, b}
							var S string
							S = getHashString(E.source, E.sink)
							EdgeWeights[S] = 0
							VisitedEdges[S] = E
						}
					}
				}
			} else if SelfReferentialEdge != nil && SelfReferentialEdge.source != nil {
				if _, ok := VisitedBlocks[b]; ok {
					var S string
					S = getHashString(SelfReferentialEdge.source, SelfReferentialEdge.sink)
					if b.BBFreq.RawCount >= TotalWeight {
						EdgeWeights[S] = b.BBFreq.RawCount - TotalWeight
					} else {
						EdgeWeights[S] = 0
					}
					VisitedEdges[S] = SelfReferentialEdge
					Changed = true
				}
			}
			if UpdateBlockCount && TotalWeight > 0 {
				if _, ok := VisitedBlocks[b]; !ok {
					b.BBFreq.RawCount = TotalWeight
					VisitedBlocks[b] = true
					Changed = true
					if f.pass.debug > 2 {
						fmt.Printf("Set weight for b%d: %v\n", b.ID, TotalWeight)
					}
				}
			}
		}
	}
	return Changed

}

// propagateWeights propagates weights into edges
//
// The following rules are applied to every block BB in the CFG:
//
// - If BB has a single predecessor/successor, then the weight
//   of that edge is the weight of the block.
//
// - If all incoming or outgoing edges are known except one, and the
//   weight of the block is already known, the weight of the unknown
//   edge will be the weight of the block minus the sum of all the known
//   edges. If the sum of all the known edges is larger than BB's weight,
//   we set the unknown edge weight to zero.
//
// - If there is a self-referential edge, and the weight of the block is
//   known, the weight for that edge is set to the weight of the block
//   minus the weight of the other incoming edges to that block (if
//   known).

func propagateWeights(f *Func, VisitedBlocks map[*Block]bool) {
	var EdgeWeights map[string]int64
	if EdgeWeights == nil {
		EdgeWeights = make(map[string]int64)
	}
	for k := range EdgeWeights {
		delete(EdgeWeights, k)
	}
	var Changed bool = false
	var VisitedEdges map[string]*edge
	if VisitedEdges == nil {
		VisitedEdges = make(map[string]*edge)
	}
	for k := range VisitedEdges {
		delete(VisitedEdges, k)
	}
	dom := f.Idom()

	// Build tree
	blocks := make([]parentBlock, f.NumBlocks())
	for _, b := range f.Blocks {
		blocks[b.ID].b = b
		if dom[b.ID] == nil {
			continue // entry or unreachable
		}
		parent := dom[b.ID].ID
		blocks[b.ID].parent = parent
	}

	for i := 0; i < SampleProfileMaxPropagateIterations; i++ {
		Changed = propagateThroughEdges(f, true, VisitedEdges, EdgeWeights, VisitedBlocks, blocks)
		if Changed == false {
			break
		}
	}

	for k := range VisitedEdges {
		delete(VisitedEdges, k)
	}
	Changed = true
	for i := 0; i < SampleProfileMaxPropagateIterations; i++ {
		Changed = propagateThroughEdges(f, false, VisitedEdges, EdgeWeights, VisitedBlocks, blocks)
		if Changed == false {
			break
		}
	}

	Changed = true
	for i := 0; i < SampleProfileMaxPropagateIterations; i++ {
		Changed = propagateThroughEdges(f, true, VisitedEdges, EdgeWeights, VisitedBlocks, blocks)
		if Changed == false {
			break
		}
	}

	updateCfgEdgeWeights(f, EdgeWeights)

}

// getEdgeWeight returns the computed edge weight if it exists.
func getEdgeWeight(source *Block, sink *Block, EdgeWeights map[string]int64) int64 {
	var E *edge
	E = &edge{source, sink}
	var S string
	S = getHashString(E.source, E.sink)
	if val, ok := EdgeWeights[S]; ok {
		return val
	} else {
		return 0
	}
}

// updateCfgEdgeWeights updates the edge weight in the control flow graph according
// to the computation in the frequency propagation.
func updateCfgEdgeWeights(f *Func, EdgeWeights map[string]int64) {
	for _, b := range f.Blocks {

		for i := 0; i < 2; i++ {
			if i == 0 {
				for _, e := range b.Preds {
					e.EdgeFreq.RawCount = getEdgeWeight(e.b, b, EdgeWeights)
				}
			} else {
				for _, e := range b.Succs {
					e.EdgeFreq.RawCount = getEdgeWeight(b, e.b, EdgeWeights)
				}
			}
		}
	}
}

// computeBlockWeights computes and stores the weights of every basic block.
func computeBlockWeights(f *Func, VisitedBlocks map[*Block]bool) bool {
	if f.pass.debug > 2 {
		fmt.Printf("\nBlock weights for function %s\n", f.Name)
	}

	var Changed bool = false

	e := f.Frontend()
	for _, b := range f.Blocks {
		if e.Func().BBFreqMap != nil {
			if bbFreq, ok := e.Func().BBFreqMap[int64(b.ID)]; ok {
				b.BBFreq.RawCount = int64(bbFreq)
			}
		}

		if b.BBFreq.RawCount != 0 {
			VisitedBlocks[b] = true
			Changed = true
			if f.pass.debug > 2 {
				fmt.Printf("weight[b%d]: %v\n", b.ID, b.BBFreq.RawCount)
			}
		}
	}
	return Changed
}

// setExitCount adds an exit count to the function using the samples gathered at the
// function exit. By default the value 1 is set.
func setExitCount(f *Func, VisitedBlocks map[*Block]bool) {
	for _, b := range f.Blocks {
		if b.BBFreq.RawCount == 0 &&
			(b.Kind == BlockExit || b.Kind == BlockRet || b.Kind == BlockRetJmp || b.Kind == BlockDefer) {
			VisitedBlocks[b] = true
			b.BBFreq.RawCount = 0
		}
	}

}

// freqPropagate propagates block weights into edges. This uses a simple
//    propagation heuristic. The following rules are applied to every
//    block BB in the CFG:
//
//    - If BB has a single predecessor/successor, then the weight
//      of that edge is the weight of the block.
//
//    - If all the edges are known except one, and the weight of the
//      block is already known, the weight of the unknown edge will
//      be the weight of the block minus the sum of all the known
//      edges. If the sum of all the known edges is larger than BB's weight,
//      we set the unknown edge weight to zero.
//
//    - If there is a self-referential edge, and the weight of the block is
//      known, the weight for that edge is set to the weight of the block
//      minus the weight of the other incoming edges to that block (if
//      known).
//
// Since this propagation is not guaranteed to finalize for every CFG, we
// only allow it to proceed for a limited number of iterations.

func freqPropagate(f *Func) {
	if base.Flag.PgoBb == 0 {
		return
	}
	var VisitedBlocks map[*Block]bool
	if VisitedBlocks == nil {
		VisitedBlocks = make(map[*Block]bool)
	}
	var Changed = computeBlockWeights(f, VisitedBlocks)
	if Changed == true {
		setExitCount(f, VisitedBlocks)
		propagateWeights(f, VisitedBlocks)
	}
}
