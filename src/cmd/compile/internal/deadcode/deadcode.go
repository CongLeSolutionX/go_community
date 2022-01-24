// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deadcode

import (
	"go/constant"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
)

func Func(fn *ir.Func) {
	stmts(&fn.Body)

	if len(fn.Body) == 0 {
		return
	}

	for _, n := range fn.Body {
		if len(n.Init()) > 0 {
			return
		}
		switch n.Op() {
		case ir.OIF:
			n := n.(*ir.IfStmt)
			if !ir.IsConst(n.Cond, constant.Bool) || len(n.Body) > 0 || len(n.Else) > 0 {
				return
			}
		case ir.OFOR:
			n := n.(*ir.ForStmt)
			if !ir.IsConst(n.Cond, constant.Bool) || ir.BoolVal(n.Cond) {
				return
			}
		default:
			return
		}
	}

	ir.VisitList(fn.Body, markHiddenClosureDead)
	fn.Body = []ir.Node{ir.NewBlockStmt(base.Pos, nil)}
}

func stmts(nn *ir.Nodes) {
	var lastLabel = -1
	for i, n := range *nn {
		if n != nil && n.Op() == ir.OLABEL {
			lastLabel = i
		}
	}
	for i, n := range *nn {
		// Cut is set to true when all nodes after i'th position
		// should be removed.
		// In other words, it marks whole slice "tail" as dead.
		cut := false
		if n == nil {
			continue
		}
		if n.Op() == ir.OIF {
			n := n.(*ir.IfStmt)
			n.Cond = expr(n.Cond)
			if ir.IsConst(n.Cond, constant.Bool) {
				var body ir.Nodes
				if ir.BoolVal(n.Cond) {
					ir.VisitList(n.Else, markHiddenClosureDead)
					n.Else = ir.Nodes{}
					body = n.Body
				} else {
					ir.VisitList(n.Body, markHiddenClosureDead)
					n.Body = ir.Nodes{}
					body = n.Else
				}
				// If "then" or "else" branch ends with panic or return statement,
				// it is safe to remove all statements after this node.
				// isterminating is not used to avoid goto-related complications.
				// We must be careful not to deadcode-remove labels, as they
				// might be the target of a goto. See issue 28616.
				if body := body; len(body) != 0 {
					switch body[(len(body) - 1)].Op() {
					case ir.ORETURN, ir.OTAILCALL, ir.OPANIC:
						if i > lastLabel {
							cut = true
						}
					}
				}
			}
		}

		if len(n.Init()) != 0 {
			stmts(n.(ir.InitNode).PtrInit())
		}
		switch n.Op() {
		case ir.OBLOCK:
			n := n.(*ir.BlockStmt)
			stmts(&n.List)
		case ir.OFOR:
			n := n.(*ir.ForStmt)
			stmts(&n.Body)
		case ir.OIF:
			n := n.(*ir.IfStmt)
			stmts(&n.Body)
			stmts(&n.Else)
		case ir.ORANGE:
			n := n.(*ir.RangeStmt)
			stmts(&n.Body)
		case ir.OSELECT:
			n := n.(*ir.SelectStmt)
			for _, cas := range n.Cases {
				stmts(&cas.Body)
			}
		case ir.OSWITCH:
			n := n.(*ir.SwitchStmt)
			switchStmt(n)
		}

		if cut {
			ir.VisitList((*nn)[i+1:len(*nn)], markHiddenClosureDead)
			*nn = (*nn)[:i+1]
			break
		}
	}
}

func expr(n ir.Node) ir.Node {
	// Perform dead-code elimination on short-circuited boolean
	// expressions involving constants with the intent of
	// producing a constant 'if' condition.
	switch n.Op() {
	case ir.OANDAND:
		n := n.(*ir.LogicalExpr)
		n.X = expr(n.X)
		n.Y = expr(n.Y)
		if ir.IsConst(n.X, constant.Bool) {
			if ir.BoolVal(n.X) {
				return n.Y // true && x => x
			} else {
				return n.X // false && x => false
			}
		}
	case ir.OOROR:
		n := n.(*ir.LogicalExpr)
		n.X = expr(n.X)
		n.Y = expr(n.Y)
		if ir.IsConst(n.X, constant.Bool) {
			if ir.BoolVal(n.X) {
				return n.X // true || x => true
			} else {
				return n.Y // false || x => x
			}
		}
	}
	return n
}

func markHiddenClosureDead(n ir.Node) {
	if n.Op() != ir.OCLOSURE {
		return
	}
	clo := n.(*ir.ClosureExpr)
	if clo.Func.IsHiddenClosure() {
		clo.Func.SetIsDeadcodeClosure(true)
	}
	ir.VisitList(clo.Func.Body, markHiddenClosureDead)
}

// switchStmt performs deadcode elimination for an ir.SwitchStmt node.
//
// If switch statement has non-constant tag, it calls stmts on each ir.CaseClause's body.
//
// If switch statement has constant tag:
//
//  + Removing any const/non-equal cases
//  + Removing all cases if they are const and non-equal
func switchStmt(n *ir.SwitchStmt) {
	firstCasePos := -1
	if ir.ConstType(n.Tag) != constant.Unknown {
		tv := ir.ConstValue(n.Tag)
		defaultCasePos := -1
		allCasesAreConstant := true
		for casePos, ncase := range n.Cases {
			if len(ncase.List) == 0 {
				defaultCasePos = casePos
				continue
			}
			caseList := ncase.List[:0]
			for _, cv := range ncase.List {
				cvTyp := ir.ConstType(cv)
				if cvTyp != constant.Unknown && ir.ConstValue(cv) != tv {
					// This is a const/non-equal case, remove it.
					continue
				}
				caseList = append(caseList, cv)
				if cvTyp == constant.Unknown {
					allCasesAreConstant = false
				}
				if cvTyp != constant.Unknown && ir.ConstValue(cv) == tv {
					firstCasePos = casePos
				}
			}
			for i := len(caseList); i < len(ncase.List); i++ {
				ncase.List[i] = nil // allow gc-ed
			}
			ncase.List = caseList
		}

		// If all cases are const/non-equal, match default case.
		if firstCasePos == -1 && allCasesAreConstant {
			firstCasePos = defaultCasePos
		}
		// If all cases are const/non-equal, and there's no default case, simply removing all cases.
		if firstCasePos == -1 && allCasesAreConstant {
			n.Cases = nil
			return
		}
	}

	// Tracking what cases need to be kept.
	casesKept := make(map[int]bool)
	constCaseFolded := firstCasePos > -1
	if constCaseFolded {
		casesKept[firstCasePos] = true
		for i := firstCasePos; i < len(n.Cases); i++ {
			endsInfallthrough, _ := ir.EndsInFallthrough(n.Cases[i].Body)
			if !endsInfallthrough {
				break
			}
			firstCasePos++
			casesKept[firstCasePos] = true
		}
	}

	for i, cas := range n.Cases {
		if constCaseFolded && !casesKept[i] {
			ir.VisitList(cas.Body, markHiddenClosureDead)
		} else {
			stmts(&cas.Body)
		}
	}
	if constCaseFolded {
		cases := make([]*ir.CaseClause, 0, len(casesKept))
		for i := range n.Cases {
			if casesKept[i] {
				cases = append(cases, n.Cases[i])
			}
		}
		n.Cases = cases
	}
}
