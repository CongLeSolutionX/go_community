// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/base"
	"cmd/internal/obj"
	"cmd/internal/src"
	"fmt"
)

func (f *Func) getSerializedInlineTree(ctxt *obj.Link, xpos src.XPos) string {
	s := ";" + f.fe.Func().LSym.String()
	pos := ctxt.InnermostPos(xpos)
	ctxt.InlTree.AllParents(pos.Base().InliningIndex(), func(call obj.InlinedCall) {
		s = ";" + call.Func.Name + s
	})
	return s
}

// getPosBBMapping scans through the blocks of f and returns a map that summarize
// the ssa value(only counting IsStmt()==true ones) distribution:
// {inline tree, line number, column number, block} -> ssa value count
//
// The semantic could be shown in the example below:
//
//	{";f;b", 5, 6, bb1} -> 7 means:
//	bb1 contains 7 IsStmt ssa values that has an inline tree of ";f;b", line
//	number 5 and column number 6.
func getPosBBMapping(f *Func) map[string]map[int]map[int]map[*Block]int {
	inlTlicoToBlockCnt := make(map[string]map[int]map[int]map[*Block]int)
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Pos.IsStmt() != src.PosIsStmt {
				// This is a Compiler-inserted value
				continue
			}
			inlT := f.getSerializedInlineTree(f.Config.ctxt, v.Pos)
			line, col := getLico(f.Config.ctxt, v.Pos)
			if _, ok := inlTlicoToBlockCnt[inlT]; !ok {
				inlTlicoToBlockCnt[inlT] = make(map[int]map[int]map[*Block]int)
			}
			if _, ok := inlTlicoToBlockCnt[inlT][line]; !ok {
				inlTlicoToBlockCnt[inlT][line] = make(map[int]map[*Block]int)
			}
			if _, ok := inlTlicoToBlockCnt[inlT][line][col]; !ok {
				inlTlicoToBlockCnt[inlT][line][col] = make(map[*Block]int)
			}
			inlTlicoToBlockCnt[inlT][line][col][b]++
		}
	}
	return inlTlicoToBlockCnt
}

func getLico(ctxt *obj.Link, xpos src.XPos) (int, int) {
	pos := ctxt.InnermostPos(xpos)
	if !pos.IsKnown() {
		pos = src.Pos{}
	}
	return int(pos.RelLine()), int(pos.RelCol())
}

// fuzzMatch matches a location {inline tree, line number, column number} to its
// nearest entry in inlTlicoToBlockCnt.
// If an exact match is found, return it, other wise pick the nearest one by
// comparing line first them column.
func fuzzyMatchInlTLico(inlT string, line int, col int, inlTlicoToBlockCnt map[string]map[int]map[int]map[*Block]int) map[*Block]int {
	if m1, ok := inlTlicoToBlockCnt[inlT]; ok {
		if m2, ok := m1[line]; ok {
			if m3, ok := m2[col]; ok {
				return m3
			}
		}
	} else {
		return nil
	}
	minDist := uint32(0xFFFFFFFF)
	matchedM1 := make(map[int]map[*Block]int)
	for l, m1 := range inlTlicoToBlockCnt[inlT] {
		lDist := l - line
		if lDist < 0 {
			lDist = -lDist
		}
		if uint32(lDist) < minDist {
			minDist = uint32(lDist)
			matchedM1 = m1
		}
	}
	minDist = uint32(0xFFFFFFFF)
	matchedM2 := make(map[*Block]int)
	for c, m2 := range matchedM1 {
		cDist := c - col
		if cDist < 0 {
			cDist = -cDist
		}
		if uint32(cDist) < minDist {
			minDist = uint32(cDist)
			matchedM2 = m2
		}
	}
	return matchedM2
}

// computeBlockWeights assigns frequencies from raw profile data to blocks.
//
// If a sample in the raw data matches multiple in inlTlicoToBlockCnt,
// they will be distributed proportionally by their ssa value count of
// the matched blocks.
func computeBlockWeights(f *Func, inlTlicoToBlockCnt map[string]map[int]map[int]map[*Block]int) {
	if f.pass.debug > 2 {
		fmt.Printf("\nBlock weights for function %s\n", f.Name)
	}

	for _, b := range f.Blocks {
		b.BBFreq = 0
	}

	e := f.Frontend()
	if e.Func().LicoFreqMap != nil {
		for inlT, m1 := range e.Func().LicoFreqMap {
			for l, m2 := range m1 {
				for c, freq := range m2 {
					bs := fuzzyMatchInlTLico(inlT, l, c, inlTlicoToBlockCnt)
					// fmt.Printf("FuzzyMatched %v for %s,%d,%d:%d\n", bs, inlT, l, c, freq)
					totalValCnt := float64(0)
					for _, valCnt := range bs {
						totalValCnt += float64(valCnt)
					}
					for b, valCnt := range bs {
						prorated := int((float64(valCnt) / totalValCnt) * float64(freq))
						b.BBFreq += prorated
					}
				}
			}
		}
	}
}

// assignblockfreq assignes profile frequencies to basic blocks by using a
// heuristic match method from {inline tree, line number, column number} to frequency.
func assignblockfreq(f *Func) {
	if base.Debug.PGOBBReorder == 0 || f.fe.Func() == nil || f.fe.Func().LSym == nil {
		// For functions missing linker symbols, the sample mapping won't work.
		return
	}
	inlTlicoToBlockCnt := getPosBBMapping(f)
	computeBlockWeights(f, inlTlicoToBlockCnt)
}
