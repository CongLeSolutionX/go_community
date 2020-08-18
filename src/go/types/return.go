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
	switch s := stmt.(type) {
	default:
		unreachable()

	case astBadStmt, astDeclStmt, astEmptyStmt, astSendStmt,
		astIncDecStmt, astAssignStmt, astGoStmt, astDeferStmt,
		astRangeStmt:
		// no chance

	case astLabeledStmt:
		return check.isTerminating(s.Stmt(), s.Label().IdentName())

	case astExprStmt:
		// calling the predeclared (possibly parenthesized) panic() function is terminating
		if call, _ := unparen(s.X()).(astCallExpr); call != nil && check.isPanic[call] {
			return true
		}

	case astReturnStmt:
		return true

	case astBranchStmt:
		if s.Tok() == token.GOTO || s.Tok() == token.FALLTHROUGH {
			return true
		}

	case astBlockStmt:
		return check.isTerminatingList(s.List(), "")

	case astIfStmt:
		if s.Else() != nil &&
			check.isTerminating(s.Body(), "") &&
			check.isTerminating(s.Else(), "") {
			return true
		}

	case astSwitchStmt:
		return check.isTerminatingSwitch(s.Body(), label)

	case astTypeSwitchStmt:
		return check.isTerminatingSwitch(s.Body(), label)

	case astSelectStmt:
		for si := 0; si < s.Body().List().Len(); si++ {
			s := s.Body().List().Stmt(si)
			cc := s.(astCommClause)
			if !check.isTerminatingList(cc.Body(), "") || hasBreakList(cc.Body(), label, true) {
				return false
			}

		}
		return true

	case astForStmt:
		if s.Cond() == nil && !hasBreak(s.Body(), label, true) {
			return true
		}
	}

	return false
}

func (check *Checker) isTerminatingList(list astStmtList, label string) bool {
	// trailing empty statements are permitted - skip them
	for i := list.Len() - 1; i >= 0; i-- {
		if empty, _ := list.Stmt(i).(astEmptyStmt); empty == nil {
			return check.isTerminating(list.Stmt(i), label)
		}
	}
	return false // all statements are empty
}

func (check *Checker) isTerminatingSwitch(body astBlockStmt, label string) bool {
	hasDefault := false
	for si := 0; si < body.List().Len(); si++ {
		cc := body.List().Stmt(si).(astCaseClause)
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
func hasBreak(stmt astStmt, label string, implicit bool) bool {
	switch s := stmt.(type) {
	default:
		unreachable()

	case astBadStmt, astDeclStmt, astEmptyStmt, astExprStmt,
		astSendStmt, astIncDecStmt, astAssignStmt, astGoStmt,
		astDeferStmt, astReturnStmt:
		// no chance

	case astLabeledStmt:
		return hasBreak(s.Stmt(), label, implicit)

	case astBranchStmt:
		if s.Tok() == token.BREAK {
			if s.Label() == nil {
				return implicit
			}
			if s.Label().IdentName() == label {
				return true
			}
		}

	case astBlockStmt:
		return hasBreakList(s.List(), label, implicit)

	case astIfStmt:
		if hasBreak(s.Body(), label, implicit) ||
			s.Else() != nil && hasBreak(s.Else(), label, implicit) {
			return true
		}

	case astCaseClause:
		return hasBreakList(s.Body(), label, implicit)

	case astSwitchStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case astTypeSwitchStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case astCommClause:
		return hasBreakList(s.Body(), label, implicit)

	case astSelectStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case astForStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
			return true
		}

	case astRangeStmt:
		if label != "" && hasBreak(s.Body(), label, false) {
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
