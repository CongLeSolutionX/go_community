// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements isTerminating.

package types

import (
	"go/token"
)

// isTerminating reports if s is a terminating statement.
// If s is labeled, label is the label name; otherwise s
// is "".
func (check *Checker) isTerminating(stmt astStmt, label string) bool {
	switch kindOfStmt(stmt) {
	default:
		unreachable()

	case badStmtKind, declStmtKind, emptyStmtKind, sendStmtKind,
		incDecStmtKind, assignStmtKind, goStmtKind, deferStmtKind,
		rangeStmtKind:
		// no chance

	case labeledStmtKind:
		s := stmt.LabeledStmt()
		return check.isTerminating(s.Stmt(), s.Label().Name())

	case exprStmtKind:
		s := stmt.ExprStmt()
		// calling the predeclared (possibly parenthesized) panic() function is terminating
		if call := unparen(s.X()).CallExpr(); call != nil && check.isPanic[call] {
			return true
		}

	case returnStmtKind:
		return true

	case branchStmtKind:
		s := stmt.BranchStmt()
		if s.Tok() == token.GOTO || s.Tok() == token.FALLTHROUGH {
			return true
		}

	case blockStmtKind:
		return check.isTerminatingList(stmt.BlockStmt().List(), "")

	case ifStmtKind:
		s := stmt.IfStmt()
		if s.Else() != nil &&
			check.isTerminating(s.Body(), "") &&
			check.isTerminating(s.Else(), "") {
			return true
		}

	case switchStmtKind:
		return check.isTerminatingSwitch(stmt.SwitchStmt().Body(), label)

	case typeSwitchStmtKind:
		return check.isTerminatingSwitch(stmt.TypeSwitchStmt().Body(), label)

	case selectStmtKind:
		s := stmt.SelectStmt()
		for si := 0; si < s.Body().List().Len(); si++ {
			s := s.Body().List().Stmt(si)
			cc := s.CommClause()
			if !check.isTerminatingList(cc.Body(), "") || hasBreakList(cc.Body(), label, true) {
				return false
			}

		}
		return true

	case forStmtKind:
		s := stmt.ForStmt()
		if s.Cond() == nil && !hasBreak(s.Body(), label, true) {
			return true
		}
	}

	return false
}

func (check *Checker) isTerminatingList(list astStmtList, label string) bool {
	// trailing empty statements are permitted - skip them
	for i := list.Len() - 1; i >= 0; i-- {
		if empty := list.Stmt(i).EmptyStmt(); empty == nil {
			return check.isTerminating(list.Stmt(i), label)
		}
	}
	return false // all statements are empty
}

func (check *Checker) isTerminatingSwitch(body astBlockStmt, label string) bool {
	hasDefault := false
	for si := 0; si < body.List().Len(); si++ {
		cc := body.List().Stmt(si).CaseClause()
		// TODO: check for more nilness checks against lists
		if cc.List().Len() == 0 {
			hasDefault = true
		}
		if !check.isTerminatingList(cc.Body(), "") || hasBreakList(cc.Body(), label, true) {
			return false
		}
	}
	return hasDefault
}

// TODO(gri) For nested breakable statements, the current implementation of hasBreak
//	     will traverse the same subtree repeatedly, once for each label. Replace
//           with a single-pass label/break matching phase.

// hasBreak reports if s is or contains a break statement
// referring to the label-ed statement or implicit-ly the
// closest outer breakable statement.
func hasBreak(stmt astStmt, label string, implicit bool) bool {
	switch stmt.Kind() {
	default:
		unreachable()

	case badStmtKind, declStmtKind, emptyStmtKind, exprStmtKind,
		sendStmtKind, incDecStmtKind, assignStmtKind, goStmtKind,
		deferStmtKind, returnStmtKind:
		// no chance

	case labeledStmtKind:
		return hasBreak(stmt.LabeledStmt().Stmt(), label, implicit)

	case branchStmtKind:
		s := stmt.BranchStmt()
		if s.Tok() == token.BREAK {
			if s.Label() == nil {
				return implicit
			}
			if s.Label().Name() == label {
				return true
			}
		}

	case blockStmtKind:
		return hasBreakList(stmt.BlockStmt().List(), label, implicit)

	case ifStmtKind:
		s := stmt.IfStmt()
		if hasBreak(s.Body(), label, implicit) ||
			s.Else() != nil && hasBreak(s.Else(), label, implicit) {
			return true
		}

	case caseClauseKind:
		return hasBreakList(stmt.CaseClause().Body(), label, implicit)

	case switchStmtKind:
		if label != "" && hasBreak(stmt.SwitchStmt().Body(), label, false) {
			return true
		}

	case typeSwitchStmtKind:
		if label != "" && hasBreak(stmt.TypeSwitchStmt().Body(), label, false) {
			return true
		}

	case commClauseKind:
		return hasBreakList(stmt.CommClause().Body(), label, implicit)

	case selectStmtKind:
		if label != "" && hasBreak(stmt.SelectStmt().Body(), label, false) {
			return true
		}

	case forStmtKind:
		if label != "" && hasBreak(stmt.ForStmt().Body(), label, false) {
			return true
		}

	case rangeStmtKind:
		if label != "" && hasBreak(stmt.RangeStmt().Body(), label, false) {
			return true
		}
	}

	return false
}

func hasBreakList(list astStmtList, label string, implicit bool) bool {
	for si := 0; si < list.Len(); si++ {
		s := list.Stmt(si)
		if hasBreak(s, label, implicit) {
			return true
		}
	}
	return false
}
