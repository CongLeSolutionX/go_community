// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements isTerminating.

package types

import (
	"go/token"
	itypes "internal/types"
)

// isTerminating reports if s is a terminating statement.
// If s is labeled, label is the label name; otherwise s
// is "".
func (check *Checker) isTerminating(stmt itypes.Stmt, label string) bool {
	switch s := stmt.(type) {
	default:
		unreachable()

	case itypes.BadStmt, itypes.DeclStmt, itypes.EmptyStmt, itypes.SendStmt,
		itypes.IncDecStmt, itypes.AssignStmt, itypes.GoStmt, itypes.DeferStmt,
		itypes.RangeStmt:
		// no chance

	case itypes.LabeledStmt:
		return check.isTerminating(s.Stmt(), s.Label().Name())

	case itypes.ExprStmt:
		// calling the predeclared (possibly parenthesized) panic() function is terminating
		if call, _ := unparen(s.X()).(itypes.CallExpr); call != nil && check.isPanic[call] {
			return true
		}

	case itypes.ReturnStmt:
		return true

	case itypes.BranchStmt:
		if s.Tok() == token.GOTO || s.Tok() == token.FALLTHROUGH {
			return true
		}

	case itypes.BlockStmt:
		return check.isTerminatingList(s.List(), "")

	case itypes.IfStmt:
		if s.Else() != nil &&
			check.isTerminating(s.Body(), "") &&
			check.isTerminating(s.Else(), "") {
			return true
		}

	case itypes.SwitchStmt:
		return check.isTerminatingSwitch(s.Body(), label)

	case itypes.TypeSwitchStmt:
		return check.isTerminatingSwitch(s.Body(), label)

	case itypes.SelectStmt:
		for si := 0; si < s.Body().List().Len(); si++ {
			s := s.Body().List().Stmt(si)
			cc := s.(itypes.CommClause)
			if !check.isTerminatingList(cc.Body(), "") || hasBreakList(cc.Body(), label, true) {
				return false
			}

		}
		return true

	case itypes.ForStmt:
		if s.Cond() == nil && !hasBreak(s.Body(), label, true) {
			return true
		}
	}

	return false
}

func (check *Checker) isTerminatingList(list itypes.StmtList, label string) bool {
	// trailing empty statements are permitted - skip them
	for i := list.Len() - 1; i >= 0; i-- {
		if empty, _ := list.Stmt(i).(itypes.EmptyStmt); empty == nil {
			return check.isTerminating(list.Stmt(i), label)
		}
	}
	return false // all statements are empty
}

func (check *Checker) isTerminatingSwitch(body itypes.BlockStmt, label string) bool {
	hasDefault := false
	for si := 0; si < body.List().Len(); si++ {
		cc := body.List().Stmt(si).(itypes.CaseClause)
		// TODO: check for more nilness checks against lists
		if cc.ListLen() == 0 {
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
func hasBreak(stmt itypes.Stmt, label string, implicit bool) bool {
	switch s := stmt.(type) {
	default:
		unreachable()

	case itypes.BadStmt, itypes.DeclStmt, itypes.EmptyStmt, itypes.ExprStmt,
		itypes.SendStmt, itypes.IncDecStmt, itypes.AssignStmt, itypes.GoStmt,
		itypes.DeferStmt, itypes.ReturnStmt:
		// no chance

	case itypes.LabeledStmt:
		return hasBreak(s.Stmt(), label, implicit)

	case itypes.BranchStmt:
		if s.Tok() == token.BREAK {
			if s.Label() == nil {
				return implicit
			}
			if s.Label().Name() == label {
				return true
			}
		}

	case itypes.BlockStmt:
		return hasBreakList(s.List(), label, implicit)

	case itypes.IfStmt:
		if hasBreak(s.Body(), label, implicit) ||
			s.Else() != nil && hasBreak(s.Else(), label, implicit) {
			return true
		}

	case itypes.CaseClause:
		return hasBreakList(s.Body(), label, implicit)

	case itypes.SwitchStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case itypes.TypeSwitchStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case itypes.CommClause:
		return hasBreakList(s.Body(), label, implicit)

	case itypes.SelectStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case itypes.ForStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case itypes.RangeStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}
	}

	return false
}

func hasBreakList(list itypes.StmtList, label string, implicit bool) bool {
	for si := 0; si < list.Len(); si++ {
		s := list.Stmt(si)
		if hasBreak(s, label, implicit) {
			return true
		}
	}
	return false
}
