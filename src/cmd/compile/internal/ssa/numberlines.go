// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/internal/src"
)

func numberLines(f *Func) {
	po := f.Postorder()
	endlines := make(map[ID]int)

	// Visit in reverse post order so that all non-loop predecessors come first.
	for j := len(po) - 1; j >= 0; j-- {
		b := po[j]
		// Find the first interesting position and check to see if it differs from any predecessor
		firstPos := src.NoXPos
		firstPosIndex := -1
		for i, v := range b.Values {
			if v.Pos.IsStmt() != src.LicoNotStmt {
				// Heuristic: OpAddr & OpStrutSelect usually go away, so default to following ssa node if possible.
				if (v.Op == OpAddr || v.Op == OpStructSelect) && i < len(b.Values)-1 &&
					b.Values[i+1].Pos.IsStmt() != src.LicoNotStmt && b.Values[i+1].Pos.Line() == v.Pos.Line() {
					continue
				}
				firstPosIndex = i
				firstPos = v.Pos
				v.Pos = firstPos.WithDefaultStmt() // default to default
				break
			}
		}

		if firstPosIndex == -1 { // Effectively empty block, consider preds.
			line := -1
			for _, p := range b.Preds {
				pbi := p.Block().ID
				if endlines[pbi] != line {
					if line == -1 {
						line = endlines[pbi]
						continue
					} else {
						line = -1
						break
					}

				}
			}
			endlines[b.ID] = line
			continue
		}
		// check predecessors for any difference
		if len(b.Preds) == 0 { // Don't forget the entry block
			b.Values[firstPosIndex].Pos = firstPos.WithIsStmt()
		} else {
			for _, p := range b.Preds {
				pbi := p.Block().ID
				if endlines[pbi] != int(firstPos.Line()) {
					b.Values[firstPosIndex].Pos = firstPos.WithIsStmt()
					break
				}

				// if pb.Control != nil && pb.Pos.IsStmt() != src.LicoNotStmt && pb.Pos.Line() != firstPos.Line() {
				// 	b.Values[firstPosIndex].Pos = firstPos.WithIsStmt()
				// 	break outer
				// }
				// for i := len(pb.Values) - 1; i >= 0; i-- {
				// 	pp := pb.Values[i].Pos
				// 	if pp.IsStmt() == src.LicoNotStmt {
				// 		continue
				// 	}
				// 	if pp.Line() != firstPos.Line() {
				// 		b.Values[firstPosIndex].Pos = firstPos.WithIsStmt()
				// 		break outer
				// 	}
				// }
			}
		}

		// iterate forward setting each new (interesting) position as a statement boundary.
		for i := firstPosIndex + 1; i < len(b.Values); i++ {
			v := b.Values[i]
			if v.Pos.IsStmt() == src.LicoNotStmt {
				continue
			}
			// Heuristic: OpAddr & OpStrutSelect usually go away, so default to following ssa node if possible.
			if (v.Op == OpAddr || v.Op == OpStructSelect) && i < len(b.Values)-1 &&
				b.Values[i+1].Pos.IsStmt() != src.LicoNotStmt && b.Values[i+1].Pos.Line() == v.Pos.Line() {
				continue
			}
			if v.Pos.Line() != firstPos.Line() {
				firstPos = v.Pos
				v.Pos = v.Pos.WithIsStmt()
			} else {
				v.Pos = v.Pos.WithDefaultStmt()
			}
		}
		endlines[b.ID] = int(firstPos.Line())
		if b.Control != nil && b.Pos.IsStmt() != src.LicoNotStmt && b.Pos.Line() != firstPos.Line() {
			b.Pos = b.Pos.WithIsStmt()
			// endlines[b.ID] = b.Pos.Line()
		}
	}
}
