// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/token"
	itypes "internal/types"
)

// labels checks correct label use in body.
func (check *Checker) labels(body itypes.BlockStmt) {
	// set of all labels in this body
	all := newScope(nil, body.Pos(), body.End(), "label")

	fwdJumps := check.blockBranches(all, nil, nil, body.List())

	// If there are any forward jumps left, no label was found for
	// the corresponding goto statements. Either those labels were
	// never defined, or they are inside blocks and not reachable
	// for the respective gotos.
	for _, jmp := range fwdJumps {
		var msg string
		name := jmp.Label().Name()
		if alt := all.Lookup(name); alt != nil {
			msg = "goto %s jumps into block"
			alt.(*Label).used = true // avoid another error
		} else {
			msg = "label %s not declared"
		}
		check.errorf(jmp.Label().Pos(), msg, name)
	}

	// spec: "It is illegal to define a label that is never used."
	for _, obj := range all.elems {
		if lbl := obj.(*Label); !lbl.used {
			check.softErrorf(lbl.pos, "label %s declared but not used", lbl.name)
		}
	}
}

// A block tracks label declarations in a block and its enclosing blocks.
type block struct {
	parent *block                        // enclosing block
	lstmt  itypes.LabeledStmt            // labeled statement to which this block belongs, or nil
	labels map[string]itypes.LabeledStmt // allocated lazily
}

// insert records a new label declaration for the current block.
// The label must not have been declared before in any block.
func (b *block) insert(s itypes.LabeledStmt) {
	name := s.Label().Name()
	if debug {
		assert(b.gotoTarget(name) == nil)
	}
	labels := b.labels
	if labels == nil {
		labels = make(map[string]itypes.LabeledStmt)
		b.labels = labels
	}
	labels[name] = s
}

// gotoTarget returns the labeled statement in the current
// or an enclosing block with the given label name, or nil.
func (b *block) gotoTarget(name string) itypes.LabeledStmt {
	for s := b; s != nil; s = s.parent {
		if t := s.labels[name]; t != nil {
			return t
		}
	}
	return nil
}

// enclosingTarget returns the innermost enclosing labeled
// statement with the given label name, or nil.
func (b *block) enclosingTarget(name string) itypes.LabeledStmt {
	for s := b; s != nil; s = s.parent {
		if t := s.lstmt; t != nil && t.Label().Name() == name {
			return t
		}
	}
	return nil
}

// blockBranches processes a block's statement list and returns the set of outgoing forward jumps.
// all is the scope of all declared labels, parent the set of labels declared in the immediately
// enclosing block, and lstmt is the labeled statement this block is associated with (or nil).
func (check *Checker) blockBranches(all *Scope, parent *block, lstmt itypes.LabeledStmt, list itypes.StmtList) []itypes.BranchStmt {
	b := &block{parent: parent, lstmt: lstmt}

	var (
		// TODO: fix this placeholder.
		varDeclPos         itypes.Pos = token.NoPos
		fwdJumps, badJumps []itypes.BranchStmt
	)

	// All forward jumps jumping over a variable declaration are possibly
	// invalid (they may still jump out of the block and be ok).
	// recordVarDecl records them for the given position.
	recordVarDecl := func(pos itypes.Pos) {
		varDeclPos = pos
		badJumps = append(badJumps[:0], fwdJumps...) // copy fwdJumps to badJumps
	}

	jumpsOverVarDecl := func(jmp itypes.BranchStmt) bool {
		if varDeclPos.IsValid() {
			for _, bad := range badJumps {
				if jmp == bad {
					return true
				}
			}
		}
		return false
	}

	blockBranches := func(lstmt itypes.LabeledStmt, list itypes.StmtList) {
		// Unresolved forward jumps inside the nested block
		// become forward jumps in the current block.
		fwdJumps = append(fwdJumps, check.blockBranches(all, b, lstmt, list)...)
	}

	var stmtBranches func(itypes.Stmt)
	stmtBranches = func(stmt itypes.Stmt) {
		switch s := stmt.(type) {
		case itypes.DeclStmt:
			if d, _ := s.Decl().(itypes.GenDecl); d != nil && d.Tok() == token.VAR {
				recordVarDecl(d.Pos())
			}

		case itypes.LabeledStmt:
			// declare non-blank label
			if name := s.Label().Name(); name != "_" {
				lbl := newLabel(s.Label().Pos(), check.pkg, name)
				if alt := all.Insert(lbl); alt != nil {
					check.softErrorf(lbl.pos, "label %s already declared", name)
					check.reportAltDecl(alt)
					// ok to continue
				} else {
					b.insert(s)
					check.recordDef(s.Label(), lbl)
				}
				// resolve matching forward jumps and remove them from fwdJumps
				i := 0
				for _, jmp := range fwdJumps {
					if jmp.Label().Name() == name {
						// match
						lbl.used = true
						check.recordUse(jmp.Label(), lbl)
						if jumpsOverVarDecl(jmp) {
							check.softErrorf(
								jmp.Label().Pos(),
								"goto %s jumps over variable declaration at line %d",
								name,
								// TODO: eliminate this type assertion
								check.fset.Position(varDeclPos.(token.Pos)).Line,
							)
							// ok to continue
						}
					} else {
						// no match - record new forward jump
						fwdJumps[i] = jmp
						i++
					}
				}
				fwdJumps = fwdJumps[:i]
				lstmt = s
			}
			stmtBranches(s.Stmt())

		case itypes.BranchStmt:
			sLabel := s.Label()
			if sLabel == nil {
				return // checked in 1st pass (check.stmt)
			}

			// determine and validate target
			name := sLabel.Name()
			switch s.Tok() {
			case token.BREAK:
				// spec: "If there is a label, it must be that of an enclosing
				// "for", "switch", or "select" statement, and that is the one
				// whose execution terminates."
				valid := false
				if t := b.enclosingTarget(name); t != nil {
					switch t.Stmt().(type) {
					case itypes.SwitchStmt, itypes.TypeSwitchStmt, itypes.SelectStmt, itypes.ForStmt, itypes.RangeStmt:
						valid = true
					}
				}
				if !valid {
					check.errorf(s.Label().Pos(), "invalid break label %s", name)
					return
				}

			case token.CONTINUE:
				// spec: "If there is a label, it must be that of an enclosing
				// "for" statement, and that is the one whose execution advances."
				valid := false
				if t := b.enclosingTarget(name); t != nil {
					switch t.Stmt().(type) {
					case itypes.ForStmt, itypes.RangeStmt:
						valid = true
					}
				}
				if !valid {
					check.errorf(s.Label().Pos(), "invalid continue label %s", name)
					return
				}

			case token.GOTO:
				if b.gotoTarget(name) == nil {
					// label may be declared later - add branch to forward jumps
					fwdJumps = append(fwdJumps, s)
					return
				}

			default:
				check.invalidAST(s.Pos(), "branch statement: %s %s", s.Tok, name)
				return
			}

			// record label use
			obj := all.Lookup(name)
			obj.(*Label).used = true
			check.recordUse(s.Label(), obj)

		case itypes.AssignStmt:
			if s.Tok() == token.DEFINE {
				recordVarDecl(s.Pos())
			}

		case itypes.BlockStmt:
			blockBranches(lstmt, s.List())

		case itypes.IfStmt:
			stmtBranches(s.Body())
			if s.Else() != nil {
				stmtBranches(s.Else())
			}

		case itypes.CaseClause:
			blockBranches(nil, s.Body())

		case itypes.SwitchStmt:
			stmtBranches(s.Body())

		case itypes.TypeSwitchStmt:
			stmtBranches(s.Body())

		case itypes.CommClause:
			blockBranches(nil, s.Body())

		case itypes.SelectStmt:
			stmtBranches(s.Body())

		case itypes.ForStmt:
			stmtBranches(s.Body())

		case itypes.RangeStmt:
			stmtBranches(s.Body())
		}
	}

	for si := 0; si < list.Len(); si++ {
		s := list.Stmt(si)
		stmtBranches(s)
	}

	return fwdJumps
}
