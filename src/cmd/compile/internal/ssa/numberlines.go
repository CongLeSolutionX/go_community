// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/internal/src"
)

// isPoorStatementStart determines whether a statement
// is likely to disappear in a future rewrite and hence
// would be a poor place to "start" that statement.
func isPoorStatementStart(v *Value, i int, b *Block) bool {
	// If the value is the last one in the block, too bad, it will have to do
	// (this assumes that the value ordering vaguely corresponds to the source program execution order, which tends to be true directly after "build")
	if i >= len(b.Values)-1 {
		return false
	}
	// These are the likely-ephemeral/fragile opcodes expected to vanish in a rewrite.
	switch v.Op {
	case OpAddr, OpOffPtr, OpStructSelect, OpConst8, OpConst16, OpConst32, OpConst64, OpConst32F, OpConst64F:
	default:
		return false
	}

	// Look ahead to see what the line number is on the next thing that could be a boundary.
	for j := i + 1; j < len(b.Values); j++ {
		if b.Values[j].Pos.IsStmt() == src.PosNotStmt { // ignore non-statements
			continue
		}
		if b.Values[j].Pos.Line() == v.Pos.Line() {
			return true
		}
		return false
	}
	return false
}

// notStmtBoundary indicates which value opcodes can never be a statement
// boundary because they don't correespond to a user's understand of a
// statement boundary.  Called from *Value.reset(), located here to keep
// all the statement boundary heuristics in one place.
func notStmtBoundary(op Op) bool {
	switch op {
	case OpPhi, OpVarKill, OpVarDef:
		// not OpCopy; rewrite overwrites ops w/ OpCopy, but their Pos can be a statement boundary that needs to be preserved by movement to another Value.
		return true
	}
	return false
}

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
			if v.Pos.IsStmt() != src.PosNotStmt {
				if isPoorStatementStart(v, i, b) {
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
			}
		}
		// iterate forward setting each new (interesting) position as a statement boundary.
		for i := firstPosIndex + 1; i < len(b.Values); i++ {
			v := b.Values[i]
			if v.Pos.IsStmt() == src.PosNotStmt {
				continue
			}
			if isPoorStatementStart(v, i, b) {
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
		if b.Control != nil && b.Pos.IsStmt() != src.PosNotStmt && b.Pos.Line() != firstPos.Line() {
			b.Pos = b.Pos.WithIsStmt()
			// endlines[b.ID] = b.Pos.Line()
		}
	}
}
