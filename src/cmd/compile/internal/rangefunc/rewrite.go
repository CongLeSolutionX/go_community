// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package rangefunc rewrites range-over-func to code that doesn't use range-over-funcs.
Rewriting the construct in the front end, before noder, means the functions generated during
the rewrite are available in a noder-generated representation for inlining by the back end.

# Theory of Operation

The basic idea is to rewrite

	for x := range f {
		...
	}

into

	f(func(x T) bool {
		...
	})

But it's not usually that easy.

# Range variables

For a range not using :=, the assigned variables cannot be function arguments
in the generated body function. Instead, we allocate fake parameters and
start the body with an assignment. For example:

	for bigExpr1, bigExpr2 = range f {
		...
	}

becomes

	f(func(#p1 #t1, #p2 #t2) bool {
		bigExpr1, bigExpr2 = #p1, #p2
		...
	})

(All the generated variables have a # at the start to signal that they
are internal variables when looking at the generated code in a
debugger. Because variables have all been resolved to the specific
objects they represent, there is no danger of using plain "p1" and
colliding with a Go variable named "p1"; the # is just nice to have,
not for correctness.)

It can also happen that there are fewer range variables than function
arguments, in which case we end up with something like

	f(func(x T, _ T2) bool {
		...
	})

or

	f(func(#p1 #t1, #p2 #t2, _ #t3) bool {
		bigExpr1, bigExpr2 = #p1, #p2
		...
	})

# Return

If the body contains a "break", that break turns into "return false",
to tell f to stop. And if the body contains a "continue", that turns
into "return true", to tell f to proceed with the next value.
Those are the easy cases.

If the body contains a return or a break/continue/goto L, then we need
to rewrite that into code that breaks out of the loop and then
triggers that control flow. In general we rewrite

	for x := range f {
		...
	}

into

	{
		var #next int
		f(func(x T) bool {
			...
			return true
		})
		... check #next ...
	}

The variable #next is an integer code that says what to do when f
returns. Each difficult statement sets #next and then returns false to
stop f.

A plain "return" rewrites to {#next = -1; return false}.
The return false breaks the loop. Then when f returns, the "check
#next" section includes

	if #next == -1 { return }

which causes the return we want.

Return with arguments is more involved. We need somewhere to store the
arguments while we break out of f, so we add them to the var
declaration, like:

	{
		var (
			#next int
			#r1 type1
			#r2 type2
		)
		f(func(x T) bool {
			...
			{
				// return a, b
				#r1, #r2 = a, b
				#next = -2
				return false
			}
			...
			return true
		})
		if #next == -2 { return #r1, #r2 }
	}

# Nested Loops

So far we've only considered a single loop. If a function contains a
sequence of loops, each can be translated individually. But loops can
be nested. It would work to translate the innermost loop and then
translate the loop around it, and so on, except that there'd be a lot
of rewriting of rewritten code and the overall traversals could end up
taking time quadratic in the depth of the nesting. To avoid all that,
we use a single rewriting pass that handles a top-most range-over-func
loop and all the range-over-func loops it contains at the same time.

If we need to return from inside a doubly-nested loop, the rewrites
above stay the same, but the check after the inner loop only says

	if #next < 0 { return false }

to stop the outer loop so it can do the actual return. That is,

	for range f {
		for range g {
			...
			return a, b
			...
		}
	}

becomes

	{
		var (
			#next int
			#r1 type1
			#r2 type2
		)
		f(func() {
			g(func() {
				...
				{
					// return a, b
					#r1, #r2 = a, b
					#next = -2
					return false
				}
				...
				return true
			})
			if #next < 0 {
				return false
			}
			return true
		})
		if #next == -2 {
			return #r1, #r2
		}
	}

Note that the #next < 0 after the inner loop handles both kinds of
return with a single check.

# Labeled break/continue of range-over-func loops

For a labeled break or continue of an outer range-over-func, we
use positive #next values. Any such labeled break or continue
really means "do N breaks" or "do N breaks and 1 continue".
We encode that as 2*N or 2*N+1 respectively.
Loops that might need to propagate a labeled break or continue
add one or both of these to the #next checks:

	if #next >= 2 {
		#next -= 2
		return false
	}

	if #next >= 1 {
		#next -= 1
		return true
	}

For example

	F: for range f {
		for range g {
			for range h {
				...
				break F
				...
				...
				continue F
				...
			}
		}
		...
	}

becomes

	{
		var #next int
		f(func() {
			g(func() {
				h(func() {
					...
					{
						// break F
						#next = 4
						return false
					}
					...
					{
						// continue F
						#next = 3
						return false
					}
					...
					return true
				})
				if #next >= 2 {
					#next -= 2
					return false
				}
				return true
			})
			if #next >= 2 {
				#next -= 2
				return false
			}
			if #next >= 1 {
				#next -= 1
				return true
			}
			...
			return true
		})
	}

Note that the post-h checks only consider a break,
since no generated code tries to continue g.

# Gotos and other labeled break/continue

The final control flow translations are goto and break/continue of a
non-range-over-func statement. In both cases, we may need to break out
of one or more range-over-func loops before we can do the actual
control flow statement. Each such break/continue/goto L statement is
assigned a unique negative #next value (below -2, since -1 and -2 are
for the two kinds of return). Then the post-checks for a given loop
test for the specific codes that refer to labels directly targetable
from that block. Otherwise, the generic

	if #next < 0 { return false }

check handles stopping the next loop to get one step closer to the label.

For example

	Top: print("start\n")
	for range f {
		for range g {
			for range h {
				...
				goto Top
				...
			}
		}
	}


becomes

	Top: print("start\n")
	{
		var #next int
		f(func() {
			g(func() {
				h(func() {
					...
					{
						// goto Top
						#next = -3
						return false
					}
					...
					return true
				})
				if #next < 0 {
					return false
				}
				return true
			})
			if #next < 0 {
				return false
			}
			return true
		})
		if #next == -3 {
			#next = 0
			goto Top
		}
	}

Labeled break/continue to non-range-over-funcs are handled the same
way as goto.

# Defers

The last wrinkle is handling defer statements. If we have

	for range f {
		defer print("A")
	}

we cannot rewrite that into

	f(func() {
		defer print("A")
	})

because the deferred code will run at the end of the iteration, not
the end of the containing function. To fix that, the runtime provides
a special hook that lets us obtain a defer "token" representing the
outer function and then use it in a later defer to attach the deferred
code to that outer function.

Normally,

	defer print("A")

compiles to

	runtime.deferproc(func() { print("A") })

This changes in a range-over-func. For example:

	for range f {
		defer print("A")
	}

compiles to

	var #defers = runtime.deferrangefunc()
	f(func() {
		runtime.deferprocat(func() { print("A") }, #defers)
	})

For this rewriting phase, we insert the explicit initialization of
#defers and then attach the #defers variable to the CallStmt
representing the defer. That variable will be propagated to the
backend and will cause the backend to compile the defer using
deferprocat instead of an ordinary deferproc.
*/
package rangefunc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"fmt"
	"go/constant"
	"os"
)

// nopos is the zero syntax.Pos.
var nopos syntax.Pos

// A rewriter implements rewriting the range-over-funcs in a given function.
type rewriter struct {
	pkg   *types2.Package
	info  *types2.Info
	outer *syntax.FuncType
	body  *syntax.BlockStmt

	// References to important types and values.
	anyName   *syntax.Name
	anyObj    types2.Object
	boolName  *syntax.Name
	boolObj   types2.Object
	intName   *syntax.Name
	intObj    types2.Object
	trueName  *syntax.Name
	falseName *syntax.Name

	// Branch numbering, computed as needed.
	branchNext map[branch]int             // branch -> #next value
	labelLoop  map[string]*syntax.ForStmt // label -> rangefunc loop it is declared inside (nil for no loop)

	// Stack of nodes being visited.
	stack    []syntax.Node // all nodes
	forStack []*forLoop    // range-over-func loops

	rewritten map[*syntax.ForStmt]syntax.Stmt

	// Declared variables in generated code for outermost loop.
	declStmt *syntax.DeclStmt
	nextName *syntax.Name
	retNames []*syntax.Name
	defers   *syntax.Name
}

// A branch is a single labeled branch.
type branch struct {
	tok   syntax.Token
	label string
}

// A forLoop describes a single range-over-func loop being processed.
type forLoop struct {
	nfor *syntax.ForStmt // actual syntax

	checkRet      bool     // add check for "return" after loop
	checkRetArgs  bool     // add check for "return args" after loop
	checkBreak    bool     // add check for "break" after loop
	checkContinue bool     // add check for "continue" after loop
	checkBranch   []branch // add check for labeled branch after loop
}

// Rewrite rewrites all the range-over-funcs in the files.
func Rewrite(pkg *types2.Package, info *types2.Info, files []*syntax.File) {
	for _, file := range files {
		syntax.Inspect(file, func(n syntax.Node) bool {
			switch n := n.(type) {
			case *syntax.FuncDecl:
				rewriteFunc(pkg, info, n.Type, n.Body)
				return false
			case *syntax.FuncLit:
				rewriteFunc(pkg, info, n.Type, n.Body)
				return false
			}
			return true
		})
	}
}

// rewriteFunc rewrites all the range-over-funcs in a single function (a top-level func or a func literal).
// The typ and body are the function's type and body.
func rewriteFunc(pkg *types2.Package, info *types2.Info, typ *syntax.FuncType, body *syntax.BlockStmt) {
	if body == nil {
		return
	}
	r := &rewriter{
		pkg:   pkg,
		info:  info,
		outer: typ,
		body:  body,
	}
	syntax.Inspect(body, r.inspect)
	if (base.Flag.W != 0) && r.forStack != nil {
		syntax.Fdump(os.Stderr, body)
	}
}

// inspect is a callback for syntax.Inspect that drives the actual rewriting.
// If it sees a func literal, it kicks off a separate rewrite for that literal.
// Otherwise, it maintains a stack of range-over-func loops and
// converts each in turn.
func (r *rewriter) inspect(n syntax.Node) bool {
	switch n := n.(type) {
	case *syntax.FuncLit:
		rewriteFunc(r.pkg, r.info, n.Type, n.Body)
		return false

	default:
		// Push n onto stack.
		r.stack = append(r.stack, n)
		if nfor, ok := forRangeFunc(n); ok {
			loop := &forLoop{nfor: nfor}
			r.forStack = append(r.forStack, loop)
			r.startLoop(loop)
		}

	case nil:
		// n == nil signals that we are done visiting
		// the top-of-stack node's children. Find it.
		n = r.stack[len(r.stack)-1]

		// If we are inside a range-over-func,
		// take this moment to replace any break/continue/goto/return
		// statements directly contained in this node.
		// Also replace any converted for statements
		// with the rewritten block.
		switch n := n.(type) {
		case *syntax.BlockStmt:
			for i, s := range n.List {
				n.List[i] = r.editStmt(s)
			}
		case *syntax.CaseClause:
			for i, s := range n.Body {
				n.Body[i] = r.editStmt(s)
			}
		case *syntax.CommClause:
			for i, s := range n.Body {
				n.Body[i] = r.editStmt(s)
			}
		case *syntax.LabeledStmt:
			n.Stmt = r.editStmt(n.Stmt)
		}

		// Pop n.
		if len(r.forStack) > 0 && r.stack[len(r.stack)-1] == r.forStack[len(r.forStack)-1].nfor {
			r.endLoop(r.forStack[len(r.forStack)-1])
			r.forStack = r.forStack[:len(r.forStack)-1]
		}
		r.stack = r.stack[:len(r.stack)-1]
	}
	return true
}

// startLoop sets up for converting a range-over-func loop.
func (r *rewriter) startLoop(loop *forLoop) {
	// For first loop in function, allocate syntax for any, bool, int, true, and false.
	if r.anyName == nil {
		pos := loop.nfor.Pos()
		r.anyName, r.anyObj = r.builtinType(pos, "any")
		r.boolName, r.boolObj = r.builtinType(pos, "bool")
		r.intName, r.intObj = r.builtinType(pos, "int")
		r.trueName, _ = r.builtinConst(pos, "true")
		r.falseName, _ = r.builtinConst(pos, "false")
		r.rewritten = make(map[*syntax.ForStmt]syntax.Stmt)
	}
}

// editStmt returns the replacement for the statement x,
// or x itself if it should be left alone.
// This includes the for loops we are converting,
// as left in x.rewritten by r.endLoop.
func (r *rewriter) editStmt(x syntax.Stmt) syntax.Stmt {
	if x, ok := x.(*syntax.ForStmt); ok {
		if s := r.rewritten[x]; s != nil {
			return s
		}
	}

	if len(r.forStack) > 0 {
		switch x := x.(type) {
		case *syntax.BranchStmt:
			return r.editBranch(x)
		case *syntax.CallStmt:
			if x.Tok == syntax.Defer {
				return r.editDefer(x)
			}
		case *syntax.ReturnStmt:
			return r.editReturn(x)
		}
	}

	return x
}

// editDefer returns the replacement for the defer statement x.
// See the "Defers" section in the package doc comment above for more context.
func (r *rewriter) editDefer(x *syntax.CallStmt) syntax.Stmt {
	if r.defers == nil {
		// Declare and initialize the #defers token.
		init := &syntax.CallExpr{
			Fun: runtimeSym(r.info, "deferrangefunc"),
		}
		stv := syntax.TypeAndValue{Type: r.anyObj.Type()}
		stv.SetIsValue()
		init.SetTypeInfo(stv)
		r.defers, _ = r.declVar("#defers", r.anyName, init)
	}

	// Attach the token as an "extra" argument to the defer.
	x.DeferAt = r.useVar(r.defers)
	setPos(x.DeferAt, x.Pos())
	return x
}

// editReturn returns the replacement for the return statement x.
// See the "Return" section in the package doc comment above for more context.
func (r *rewriter) editReturn(x *syntax.ReturnStmt) syntax.Stmt {
	// #next = -1 is return with no arguments; -2 is return with arguments.
	var next int
	if x.Results == nil {
		next = -1
		r.forStack[0].checkRet = true
	} else {
		next = -2
		r.forStack[0].checkRetArgs = true
	}

	// Tell the loops along the way to check for a return.
	for _, loop := range r.forStack[1:] {
		loop.checkRet = true
	}

	// Assign results, set #next, and return false.
	bl := &syntax.BlockStmt{}
	if x.Results != nil {
		if r.retNames == nil {
			for i, a := range r.outer.ResultList {
				n, _ := r.declVar(fmt.Sprintf("#r%d", i+1), a.Type, nil)
				r.retNames = append(r.retNames, n)
			}
		}
		bl.List = append(bl.List, &syntax.AssignStmt{Lhs: r.useList(r.retNames), Rhs: x.Results})
	}
	bl.List = append(bl.List, &syntax.AssignStmt{Lhs: r.useVar(r.next()), Rhs: r.intConst(next)})
	bl.List = append(bl.List, &syntax.ReturnStmt{Results: r.useVar(r.falseName)})
	setPos(bl, x.Pos())
	return bl
}

// editBranch returns the replacement for the branch statement x,
// or x itself if it should be left alone.
// See the package doc comment above for more context.
func (r *rewriter) editBranch(x *syntax.BranchStmt) syntax.Stmt {
	if x.Tok == syntax.Fallthrough {
		// Fallthrough is unaffected by the rewrite.
		return x
	}

	// Find target of break/continue/goto in r.forStack.
	// (The target may not be in r.forStack at all.)
	targ := x.Target
	i := len(r.forStack) - 1
	for i >= 0 && r.forStack[i].nfor != targ {
		i--
	}
	if i < 0 && x.Label == nil {
		// Unlabeled break or continue that's not nfor must be inside nfor. Leave alone.
		return x
	}

	// Compute the value to assign to #next and the specific return to use.
	var next int
	var ret *syntax.ReturnStmt
	if x.Tok == syntax.Goto || i < 0 {
		// goto Label
		// or break/continue of labeled non-range-over-func loop.
		// We may be able to leave it alone, or we may have to break
		// out of one or more nested loops and then use #next to signal
		// to complete the break/continue/goto.
		// Figure out which range-over-func loop contains the label.
		r.computeBranchNext()
		nfor := r.forStack[len(r.forStack)-1].nfor
		label := x.Label.Value
		targ := r.labelLoop[label]
		if nfor == targ {
			// Label is in the current func literal; use it directly.
			return x
		}

		// Set #next to the code meaning break/continue/goto label.
		next = r.branchNext[branch{x.Tok, label}]

		// Break out of nested loops up to targ.
		i := len(r.forStack) - 1
		for i >= 0 && r.forStack[i].nfor != targ {
			i--
		}
		i++ // skip loop containing targ or, when i=-1, outermost func body

		// Mark loop we exit to get to targ to check for that branch.
		top := r.forStack[i]
		top.checkBranch = append(top.checkBranch, branch{x.Tok, label})

		// Mark loops along the way to check for a plain return, so they break.
		for i++; i < len(r.forStack); i++ {
			r.forStack[i].checkRet = true
		}

		// In the innermost loop, use a plain "return false".
		ret = &syntax.ReturnStmt{Results: r.useVar(r.falseName)}
	} else {
		// break/continue of labeled non-range-over-func loop.
		depth := len(r.forStack) - 1 - i

		// For continue of innermost loop, use "return true".
		// Otherwise we are breaking the innermost loop, so "return false".
		retVal := r.falseName
		if depth == 0 && x.Tok == syntax.Continue {
			retVal = r.trueName
		}
		ret = &syntax.ReturnStmt{Results: r.useVar(retVal)}

		// If we're only operating on the innermost loop, the return is all we need.
		if depth == 0 {
			setPos(ret, x.Pos())
			return ret
		}

		// The loop inside the one we are break/continue-ing
		// needs to make that happen when we break out of it.
		if x.Tok == syntax.Continue {
			r.forStack[i+1].checkContinue = true
		} else {
			r.forStack[i+1].checkBreak = true
		}

		// The loops along the way just need to break.
		for j := i + 2; j < len(r.forStack); j++ {
			r.forStack[j].checkBreak = true
		}

		// Set next to break the appropriate number of times;
		// the final time may be a continue, not a break.
		next = 2 * depth
		if x.Tok == syntax.Continue {
			next--
		}
	}

	// Assign #next = next and do the return.
	as := &syntax.AssignStmt{Lhs: r.useVar(r.next()), Rhs: r.intConst(next)}
	bl := &syntax.BlockStmt{
		List: []syntax.Stmt{as, ret},
	}
	setPos(bl, x.Pos())
	return bl
}

// computeBranchNext computes the branchNext numbering
// and determines which labels end up inside which range-over-func loop bodies.
func (r *rewriter) computeBranchNext() {
	if r.labelLoop != nil {
		return
	}

	r.labelLoop = make(map[string]*syntax.ForStmt)
	r.branchNext = make(map[branch]int)

	var labels []string
	var stack []syntax.Node
	var forStack []*syntax.ForStmt
	forStack = append(forStack, nil)
	syntax.Inspect(r.body, func(n syntax.Node) bool {
		if n != nil {
			stack = append(stack, n)
			if nfor, ok := forRangeFunc(n); ok {
				forStack = append(forStack, nfor)
			}
			if n, ok := n.(*syntax.LabeledStmt); ok {
				l := n.Label.Value
				labels = append(labels, l)
				f := forStack[len(forStack)-1]
				r.labelLoop[l] = f
			}
		} else {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if n == forStack[len(forStack)-1] {
				forStack = forStack[:len(forStack)-1]
			}
		}
		return true
	})

	// Assign numbers to all the labels we observed.
	used := -2
	for _, l := range labels {
		used -= 3
		r.branchNext[branch{syntax.Break, l}] = used
		r.branchNext[branch{syntax.Continue, l}] = used + 1
		r.branchNext[branch{syntax.Goto, l}] = used + 2
	}
}

// endLoop finishes the conversion of a range-over-func loop.
// We have inspected and rewritten the body of the loop and can now
// construct the body function and rewrite the for loop into a call
// bracketed by any declarations and checks it requires.
func (r *rewriter) endLoop(loop *forLoop) {
	// Pick apart for range X { ... }
	nfor := loop.nfor
	start, end := nfor.Pos(), nfor.Body.Rbrace // start, end position of for loop
	rclause := nfor.Init.(*syntax.RangeClause)
	rfunc := types2.CoreType(rclause.X.GetTypeInfo().Type).(*types2.Signature) // type of X - func(func(...)bool)
	ftyp := rfunc.Params().At(0).Type().(*types2.Signature)                    // func(...) bool

	// Build X(bodyFunc)
	call := &syntax.ExprStmt{
		X: &syntax.CallExpr{
			Fun: rclause.X,
			ArgList: []syntax.Expr{
				r.bodyFunc(nfor.Body.List, rclause.Lhs, rclause.Def, ftyp, start, end),
			},
		},
	}
	setPos(call, start)

	// Build checks based on #next after X(bodyFunc)
	checks := r.checks(loop, end)

	// Rewrite for vars := range X { ... } to
	//
	//	{
	//		r.declStmt
	//		call
	//		checks
	//	}
	//
	// The r.declStmt can be added to by this loop or any inner loop
	// during the creation of r.bodyFunc; it is only emitted in the outermost
	// converted range loop.
	block := &syntax.BlockStmt{Rbrace: end}
	setPos(block, start)
	if len(r.forStack) == 1 && r.declStmt != nil {
		setPos(r.declStmt, start)
		block.List = append(block.List, r.declStmt)
	}
	block.List = append(block.List, call)
	block.List = append(block.List, checks...)

	if len(r.forStack) == 1 { // ending an outermost loop
		r.declStmt = nil
		r.nextName = nil
		r.retNames = nil
		r.defers = nil
	}

	r.rewritten[nfor] = block
}

// bodyFunc converts the loop body (control flow has already been updated)
// to a func literal that can be passed to the range function.
//
// vars is the range variables from the range statement.
// def indicates whether this is a := range statement.
// ftyp is the type of the function we are creating
// start and end are the syntax positions to use for new nodes
// that should be at the start or end of the loop.
func (r *rewriter) bodyFunc(body []syntax.Stmt, vars syntax.Expr, def bool, ftyp *types2.Signature, start, end syntax.Pos) *syntax.FuncLit {
	// Starting X(bodyFunc); build up bodyFunc first.
	var params, results []*types2.Var
	results = append(results, types2.NewVar(start, nil, "", r.boolObj.Type()))
	bodyFunc := &syntax.FuncLit{
		Type: &syntax.FuncType{
			ParamList:  []*syntax.Field{},
			ResultList: []*syntax.Field{{Type: r.boolName}},
		},
		Body: &syntax.BlockStmt{
			List:   []syntax.Stmt{},
			Rbrace: end,
		},
	}
	setPos(bodyFunc, start)

	// Arguments to function are exactly the range variables
	// with := and initialize the range variables otherwise.
	addParam := func(x syntax.Expr, i int) {
		typ := ftyp.Params().At(i).Type()
		typName := declType(start, fmt.Sprintf("#t%d", i+1), typ)

		var paramName *syntax.Name
		var paramVar *types2.Var
		if def && x != nil {
			// Reuse range variable as parameter.
			paramName = x.(*syntax.Name)
			paramVar = r.info.Defs[paramName].(*types2.Var)
		} else {
			// Declare new parameter and assign it to range expression.
			paramName, paramVar = r.newVar(start, fmt.Sprintf("#p%d", 1+len(bodyFunc.Type.ParamList)), typ)
			if x != nil {
				as := &syntax.AssignStmt{Lhs: x, Rhs: r.useVar(paramName)}
				as.SetPos(x.Pos())
				setPos(as.Rhs, x.Pos())
				bodyFunc.Body.List = append(bodyFunc.Body.List, as)
			}
		}
		f := &syntax.Field{
			Name: paramName,
			Type: typName,
		}
		f.SetPos(start)
		bodyFunc.Type.ParamList = append(bodyFunc.Type.ParamList, f)
		params = append(params, paramVar)
	}
	if vars != nil {
		if list, ok := vars.(*syntax.ListExpr); ok {
			for i, x := range list.ElemList {
				addParam(x, i)
			}
		} else {
			addParam(vars, 0)
		}
	}
	for len(params) < ftyp.Params().Len() {
		addParam(nil, len(params))
	}

	stv := syntax.TypeAndValue{
		Type: types2.NewSignatureType(nil, nil, nil,
			types2.NewTuple(params...),
			types2.NewTuple(results...),
			false),
	}
	stv.SetIsValue()
	bodyFunc.SetTypeInfo(stv)

	// Original loop body (already rewritten by editStmt during inspect).
	bodyFunc.Body.List = append(bodyFunc.Body.List, body...)

	// return true to continue at end of loop body
	ret := &syntax.ReturnStmt{Results: r.trueName}
	ret.SetPos(end)
	bodyFunc.Body.List = append(bodyFunc.Body.List, ret)

	return bodyFunc
}

// checks returns the post-call checks that need to be done for the given loop.
func (r *rewriter) checks(loop *forLoop, pos syntax.Pos) []syntax.Stmt {
	var list []syntax.Stmt
	if len(loop.checkBranch) > 0 {
		did := make(map[branch]bool)
		for _, br := range loop.checkBranch {
			if did[br] {
				continue
			}
			did[br] = true
			doBranch := &syntax.BranchStmt{Tok: br.tok, Label: &syntax.Name{Value: br.label}}
			list = append(list, r.ifNext(syntax.Eql, r.branchNext[br], doBranch))
		}
	}
	if len(r.forStack) == 1 {
		if loop.checkRetArgs {
			list = append(list, r.ifNext(syntax.Eql, -2, retStmt(r.useList(r.retNames))))
		}
		if loop.checkRet {
			list = append(list, r.ifNext(syntax.Eql, -1, retStmt(nil)))
		}
	} else {
		if loop.checkRetArgs || loop.checkRet {
			// Note: next < 0 also handles gotos handled by outer loops.
			// We set checkRet in that case to trigger this check.
			list = append(list, r.ifNext(syntax.Lss, 0, retStmt(r.useVar(r.falseName))))
		}
		if loop.checkBreak {
			list = append(list, r.ifNext(syntax.Geq, 2, retStmt(r.useVar(r.falseName))))
		}
		if loop.checkContinue {
			list = append(list, r.ifNext(syntax.Eql, 1, retStmt(r.useVar(r.trueName))))
		}
	}

	for _, j := range list {
		setPos(j, pos)
	}
	return list
}

// retStmt returns a return statement returning the given return values.
func retStmt(results syntax.Expr) *syntax.ReturnStmt {
	return &syntax.ReturnStmt{Results: results}
}

// ifNext returns the statement:
//
//	if #next op c { adjust; then }
//
// When op is >=, adjust is #next -= c.
// When op is == and c is not -1 or -2, adjust is #next = 0.
// Otherwise adjust is omitted.
func (r *rewriter) ifNext(op syntax.Operator, c int, then syntax.Stmt) syntax.Stmt {
	nif := &syntax.IfStmt{
		Cond: &syntax.Operation{Op: op, X: r.useVar(r.next()), Y: r.intConst(c)},
		Then: &syntax.BlockStmt{
			List: []syntax.Stmt{then},
		},
	}
	stv := syntax.TypeAndValue{Type: r.boolObj.Type()}
	stv.SetIsValue()
	nif.Cond.SetTypeInfo(stv)

	if op == syntax.Geq {
		sub := &syntax.AssignStmt{
			Op:  syntax.Sub,
			Lhs: r.useVar(r.next()),
			Rhs: r.intConst(c),
		}
		nif.Then.List = []syntax.Stmt{sub, then}
	}
	if op == syntax.Eql && c != -1 && c != -2 {
		clr := &syntax.AssignStmt{
			Lhs: r.useVar(r.next()),
			Rhs: r.intConst(0),
		}
		nif.Then.List = []syntax.Stmt{clr, then}
	}

	return nif
}

// next returns a reference to the #next variable.
func (r *rewriter) next() *syntax.Name {
	if r.nextName == nil {
		r.nextName, _ = r.declVar("#next", r.intName, nil)
	}
	return r.useVar(r.nextName)
}

// forRangeFunc checks whether n is a range-over-func.
// If so, it returns n.(*syntax.ForStmt), true.
// Otherwise it returns nil, false.
func forRangeFunc(n syntax.Node) (*syntax.ForStmt, bool) {
	nfor, ok := n.(*syntax.ForStmt)
	if !ok {
		return nil, false
	}
	nrange, ok := nfor.Init.(*syntax.RangeClause)
	if !ok {
		return nil, false
	}
	_, ok = types2.CoreType(nrange.X.GetTypeInfo().Type).(*types2.Signature)
	if !ok {
		return nil, false
	}
	return nfor, true
}

// builtinType returns references to the builtin type with the given name.
func (r *rewriter) builtinType(pos syntax.Pos, name string) (*syntax.Name, types2.Object) {
	obj := types2.Universe.Lookup(name)
	n := syntax.NewName(pos, name)
	stv := syntax.TypeAndValue{Type: obj.Type()}
	n.SetTypeInfo(stv)
	r.info.Uses[n] = obj
	return n, obj
}

// builtinConst returns references to the builtin constant with the given name.
func (r *rewriter) builtinConst(pos syntax.Pos, name string) (*syntax.Name, types2.Object) {
	obj := types2.Universe.Lookup(name)
	n := syntax.NewName(pos, name)
	stv := syntax.TypeAndValue{Type: obj.Type(), Value: obj.(*types2.Const).Val()}
	stv.SetIsValue()
	n.SetTypeInfo(stv)
	r.info.Uses[n] = obj
	return n, obj
}

// intConst returns syntax for an integer literal with the given value.
func (r *rewriter) intConst(c int) *syntax.BasicLit {
	lit := &syntax.BasicLit{
		Value: fmt.Sprint(c),
		Kind:  syntax.IntLit,
	}
	stv := syntax.TypeAndValue{Type: r.intObj.Type(), Value: constant.MakeInt64(int64(c))}
	stv.SetIsValue()
	lit.SetTypeInfo(stv)
	return lit
}

// useVar returns syntax for a reference to decl, which should be its declaration.
func (r *rewriter) useVar(decl *syntax.Name) *syntax.Name {
	obj := r.info.Uses[decl]
	n := syntax.NewName(nopos, decl.Value)
	stv := syntax.TypeAndValue{Type: obj.Type()}
	stv.SetIsValue()
	n.SetTypeInfo(stv)
	r.info.Uses[n] = obj
	return n
}

// useList is useVar for a list of decls.
func (r *rewriter) useList(decls []*syntax.Name) syntax.Expr {
	var new []syntax.Expr
	for _, decl := range decls {
		new = append(new, r.useVar(decl))
	}
	if len(new) == 1 {
		return new[0]
	}
	return &syntax.ListExpr{ElemList: new}
}

// newVar declares a new variable with the given name and type.
// The returned name must be added to a declaration list.
func (r *rewriter) newVar(pos syntax.Pos, name string, typ types2.Type) (*syntax.Name, *types2.Var) {
	obj := types2.NewVar(pos, r.pkg, name, typ)
	n := syntax.NewName(pos, name)
	stv := syntax.TypeAndValue{Type: typ}
	stv.SetIsValue()
	n.SetTypeInfo(stv)
	r.info.Uses[n] = obj
	return n, obj
}

// declVar declares a variable with a given name type and initializer value.
func (r *rewriter) declVar(name string, typ, init syntax.Expr) (*syntax.Name, *types2.Var) {
	if r.declStmt == nil {
		r.declStmt = &syntax.DeclStmt{}
	}
	stmt := r.declStmt
	n, obj := r.newVar(stmt.Pos(), name, typ.GetTypeInfo().Type)
	r.info.Defs[n] = obj
	stmt.DeclList = append(stmt.DeclList, &syntax.VarDecl{
		NameList: []*syntax.Name{n},
		Type:     typ,
		Values:   init,
	})
	return n, obj
}

// declType declares a type with the given name and type.
// This is more like "type name = typ" than "type name typ".
func declType(pos syntax.Pos, name string, typ types2.Type) *syntax.Name {
	n := syntax.NewName(pos, name)
	n.SetTypeInfo(syntax.TypeAndValue{Type: typ})
	return n
}

// runtimePkg is a fake runtime package that contains what we need to refer to in package runtime.
var runtimePkg = func() *types2.Package {
	var nopos syntax.Pos
	pkg := types2.NewPackage("runtime", "runtime")
	anyType := types2.Universe.Lookup("any").Type()

	// func deferrangefunc() unsafe.Pointer
	obj := types2.NewVar(nopos, pkg, "deferrangefunc", types2.NewSignatureType(nil, nil, nil, nil, types2.NewTuple(types2.NewParam(nopos, pkg, "extra", anyType)), false))
	pkg.Scope().Insert(obj)

	return pkg
}()

// runtimeSym returns a reference to a symbol in the fake runtime package.
func runtimeSym(info *types2.Info, name string) *syntax.Name {
	obj := runtimePkg.Scope().Lookup(name)
	n := syntax.NewName(nopos, "runtime."+name)
	stv := syntax.TypeAndValue{Type: obj.Type()}
	stv.SetIsValue()
	n.SetTypeInfo(stv)
	info.Uses[n] = obj
	return n
}

// setPos walks the top structure of x that has no position assigned
// and assigns it all to have positon pos.
// When setPos encounters a syntax node with a position assigned,
// setPos does not look inside that node.
// setPos only needs to handle syntax we create in this package;
// all other syntax should have positions assigned already.
func setPos(x syntax.Node, pos syntax.Pos) {
	if x == nil || x.Pos() != nopos {
		return
	}
	x.SetPos(pos)
	switch x := x.(type) {
	default:
		panic(fmt.Sprintf("setPos(%T)", x))
	case *syntax.AssignStmt:
		setPos(x.Lhs, pos)
		setPos(x.Rhs, pos)
	case *syntax.BasicLit:
		// nothing
	case *syntax.BlockStmt:
		for _, s := range x.List {
			setPos(s, pos)
		}
		if x.Rbrace != nopos {
			x.Rbrace = pos
		}
	case *syntax.BranchStmt:
		if x.Label != nil {
			setPos(x.Label, pos)
		}
	case *syntax.CallExpr:
		setPos(x.Fun, pos)
		for _, a := range x.ArgList {
			setPos(a, pos)
		}
	case *syntax.DeclStmt:
		for _, d := range x.DeclList {
			setPos(d, pos)
		}
	case *syntax.ExprStmt:
		setPos(x.X, pos)
	case *syntax.Field:
		if x.Name != nil {
			setPos(x.Name, pos)
		}
		setPos(x.Type, pos)
	case *syntax.FuncLit:
		setPos(x.Type, pos)
		setPos(x.Body, pos)
	case *syntax.FuncType:
		for _, f := range x.ParamList {
			setPos(f, pos)
		}
		for _, f := range x.ResultList {
			setPos(f, pos)
		}
	case *syntax.IfStmt:
		setPos(x.Init, pos)
		setPos(x.Cond, pos)
		setPos(x.Then, pos)
		setPos(x.Else, pos)
	case *syntax.ListExpr:
		for _, x := range x.ElemList {
			setPos(x, pos)
		}
	case *syntax.Name:
		// nothing
	case *syntax.Operation:
		setPos(x.X, pos)
		setPos(x.Y, pos)
	case *syntax.ReturnStmt:
		setPos(x.Results, pos)
	case *syntax.VarDecl:
		for _, n := range x.NameList {
			setPos(n, pos)
		}
		setPos(x.Type, pos)
		setPos(x.Values, pos)
	}
}
