// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements various error reporters.

package types

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

func assert(p bool) {
	if !p {
		panic("assertion failed")
	}
}

func unreachable() {
	panic("unreachable")
}

func fullyQualified(pkg *Package) string {
	return strconv.Quote(pkg.path)
}

// qualifyScope is used to disambiguate package qualifiers within a single
// logical scope (for example an error message).
type qualifyScope struct {
	check     *Checker
	ambiguous bool                // whether ambiguity has been encountered
	names     map[string]*Package // package name->package, to detect ambiguities
}

func (q *qualifyScope) qualifier(pkg *Package) string {
	if pkg == q.check.pkg {
		return ""
	}
	if ambiguous := q.update(pkg.name, pkg); ambiguous {
		return fullyQualified(pkg)
	}
	return pkg.name
}

func (q *qualifyScope) update(name string, pkg *Package) bool {
	if q.names == nil {
		q.names = make(map[string]*Package)
	}
	existing, ok := q.names[name]
	if !ok {
		q.names[name] = pkg
		// We may be setting a nil pkg when merging scopes, in which case this
		// scope is now ambiguous.
		if pkg == nil {
			q.ambiguous = true
		}
		return pkg == nil
	}
	if existing == nil {
		// We've already detected ambiguity for this package.
		return true
	}
	if existing == pkg {
		// existing != nil && existing == pkg. No ambiguity.
		return false
	}
	// exising != nil && existing != pkg. This is the first time we've detected
	// ambiguity for this package name.
	q.ambiguous = true
	// Setting q.names[pkg.name] to nil forces all subsequent qualifications of
	// this package name to be considered ambiguous.
	q.names[name] = nil
	return true
}

func (q *qualifyScope) merge(s qualifiedString) qualifiedString {
	for name, pkg := range s.scope.names {
		q.update(name, pkg)
	}
	newArgs := make([]interface{}, len(s.formattedArgs))
	copy(newArgs, s.formattedArgs)
	s.formattedArgs = newArgs
	s.scope = q
	return s
}

type qualifiedString struct {
	format        string
	args          []interface{}
	formattedArgs []interface{}
	scope         *qualifyScope
}

func (s qualifiedString) String() string {
	if s.scope != nil && s.scope.ambiguous {
		for i, arg := range s.args {
			switch a := arg.(type) {
			case *operand:
				arg = operandString(a, s.scope.qualifier)
			case Object:
				arg = ObjectString(a, s.scope.qualifier)
			case Type:
				arg = TypeString(a, s.scope.qualifier)
			default:
				continue
			}
			s.formattedArgs[i] = arg
		}
	}
	return fmt.Sprintf(s.format, s.formattedArgs...)
}

type unqualified string

func (s unqualified) String() string {
	return string(s)
}

func (q *qualifyScope) sprintf(format string, args ...interface{}) qualifiedString {
	formattedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		switch a := arg.(type) {
		case nil:
			arg = "<nil>"
		case operand:
			panic("internal error: should always pass *operand")
		case *operand:
			arg = operandString(a, q.qualifier)
		case token.Pos:
			arg = q.check.fset.Position(a).String()
		case ast.Expr:
			arg = ExprString(a)
		case Object:
			arg = ObjectString(a, q.qualifier)
		case Type:
			arg = TypeString(a, q.qualifier)
		case qualifiedString:
			arg = q.merge(a)
		}
		formattedArgs[i] = arg
	}
	return qualifiedString{
		format:        format,
		args:          args,
		formattedArgs: formattedArgs,
		scope:         q,
	}
}

func (check *Checker) newQualifyScope() *qualifyScope {
	return &qualifyScope{check: check}
}

func (check *Checker) sprintf(format string, args ...interface{}) string {
	return check.newQualifyScope().sprintf(format, args...).String()
}

func (check *Checker) trace(pos token.Pos, format string, args ...interface{}) {
	fmt.Printf("%s:\t%s%s\n",
		check.fset.Position(pos),
		strings.Repeat(".  ", check.indent),
		check.sprintf(format, args...),
	)
}

// dump is only needed for debugging
func (check *Checker) dump(format string, args ...interface{}) {
	fmt.Println(check.sprintf(format, args...))
}

func (check *Checker) err(err error) {
	if err == nil {
		return
	}
	var e Error
	isInternal := errors.As(err, &e)
	// Cheap trick: Don't report errors with messages containing
	// "invalid operand" or "invalid type" as those tend to be
	// follow-on errors which don't add useful information. Only
	// exclude them if these strings are not at the beginning,
	// and only if we have at least one error already reported.
	isInvalidErr := isInternal && (strings.Index(e.Msg, "invalid operand") > 0 || strings.Index(e.Msg, "invalid type") > 0)
	if check.firstErr != nil && isInvalidErr {
		return
	}

	if check.errpos != nil && isInternal {
		// If we have an internal error and the errpos override is set, use it to
		// augment our error positioning.
		// TODO(rFindley) we may also want to augment the error message and refer
		// to the position (pos) in the original expression.
		span := spanOf(check.errpos)
		e.Pos = span.pos
		e.go116start = span.start
		e.go116end = span.end
		err = e
	}

	if check.firstErr == nil {
		check.firstErr = err
	}

	if trace {
		pos := e.Pos
		msg := e.Msg
		if !isInternal {
			msg = err.Error()
			pos = token.NoPos
		}
		check.trace(pos, "ERROR: %s", msg)
	}

	f := check.conf.Error
	if f == nil {
		panic(bailout{}) // report only first error
	}
	f(err)
}

func (check *Checker) newError(at positioner, code errorCode, soft bool, msg string) error {
	span := spanOf(at)
	return Error{
		Fset:       check.fset,
		Pos:        span.pos,
		Msg:        msg,
		Soft:       soft,
		go116code:  code,
		go116start: span.start,
		go116end:   span.end,
	}
}

// newErrorf creates a new Error, but does not handle it.
func (check *Checker) newErrorf(at positioner, code errorCode, soft bool, format string, args ...interface{}) error {
	msg := check.sprintf(format, args...)
	return check.newError(at, code, soft, msg)
}

func (check *Checker) error(at positioner, code errorCode, msg string) {
	check.err(check.newError(at, code, false, msg))
}

func (check *Checker) errorf(at positioner, code errorCode, format string, args ...interface{}) {
	msg := check.sprintf(format, args...)
	check.error(at, code, msg)
}

func (check *Checker) softErrorf(at positioner, code errorCode, format string, args ...interface{}) {
	check.err(check.newErrorf(at, code, true, format, args...))
}

func (check *Checker) invalidAST(at positioner, format string, args ...interface{}) {
	check.errorf(at, 0, "invalid AST: "+format, args...)
}

func (check *Checker) invalidArg(at positioner, code errorCode, format string, args ...interface{}) {
	check.errorf(at, code, "invalid argument: "+format, args...)
}

func (check *Checker) invalidOp(at positioner, code errorCode, format string, args ...interface{}) {
	check.errorf(at, code, "invalid operation: "+format, args...)
}

// The positioner interface is used to extract the position of type-checker
// errors.
type positioner interface {
	Pos() token.Pos
}

// posSpan holds a position range along with a highlighted position within that
// range. This is used for positioning errors, with pos by convention being the
// first position in the source where the error is known to exist, and start
// and end defining the full span of syntax being considered when the error was
// detected. Invariant: start <= pos < end || start == pos == end.
type posSpan struct {
	start, pos, end token.Pos
}

func (e posSpan) Pos() token.Pos {
	return e.pos
}

// inNode creates a posSpan for the given node.
// Invariant: node.Pos() <= pos < node.End() (node.End() is the position of the
// first byte after node within the source).
func inNode(node ast.Node, pos token.Pos) posSpan {
	start, end := node.Pos(), node.End()
	if debug {
		assert(start <= pos && pos < end)
	}
	return posSpan{start, pos, end}
}

// atPos wraps a token.Pos to implement the positioner interface.
type atPos token.Pos

func (s atPos) Pos() token.Pos {
	return token.Pos(s)
}

// spanOf extracts an error span from the given positioner. By default this is
// the trivial span starting and ending at pos, but this span is expanded when
// the argument naturally corresponds to a span of source code.
func spanOf(at positioner) posSpan {
	switch x := at.(type) {
	case nil:
		panic("internal error: nil")
	case posSpan:
		return x
	case ast.Node:
		pos := x.Pos()
		return posSpan{pos, pos, x.End()}
	case *operand:
		if x.expr != nil {
			pos := x.Pos()
			return posSpan{pos, pos, x.expr.End()}
		}
		return posSpan{token.NoPos, token.NoPos, token.NoPos}
	default:
		pos := at.Pos()
		return posSpan{pos, pos, pos}
	}
}
