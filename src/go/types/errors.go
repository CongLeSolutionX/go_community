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

func (check *Checker) qualifier(pkg *Package) string {
	// Qualify the package unless it's the package being type-checked.
	if pkg != check.pkg {
		// If the same package name was used by multiple packages, display the full path.
		if check.pkgCnt[pkg.name] > 1 {
			return fullyQualified(pkg)
		}
		return pkg.name
	}
	return ""
}

func fullyQualified(pkg *Package) string {
	return strconv.Quote(pkg.path)
}

// qualifyScope is used to disambiguate package qualifiers within a single
// logical scope (for example an error message).
type qualifyScope struct {
	ambiguous bool // whether ambiguity has been encountered
	names     map[string]*Package
	qf        Qualifier
}

func (q *qualifyScope) qualifier(pkg *Package) string {
	if q.names == nil {
		q.names = make(map[string]*Package)
	}
	if existing, ok := q.names[pkg.name]; ok && existing != pkg {
		q.ambiguous = true
		// Setting q.names[pkg.name] to nil forces all subsequant qualifications of
		// this package name to be considered ambiguous.
		q.names[pkg.name] = nil
		return fullyQualified(pkg)
	}
	q.names[pkg.name] = pkg
	return q.qf(pkg)
}

func (check *Checker) sprintf(format string, args ...interface{}) string {
	// Until formatting all arguments, it is not known whether any packages
	// referenced by args are mutually ambiguous, so we use a qualifyScope
	// identify and record ambiguities. If any ambiguities are encountered we
	// perform a second pass so that all ambiguous package names are fully
	// qualified.
	formattedArgs := make([]interface{}, len(args))
	qs := &qualifyScope{qf: check.qualifier}
	for pass := 0; pass < 2; pass++ {
		for i, arg := range args {
			switch a := arg.(type) {
			case nil:
				arg = "<nil>"
			case operand:
				panic("internal error: should always pass *operand")
			case *operand:
				arg = operandString(a, qs.qualifier)
			case token.Pos:
				arg = check.fset.Position(a).String()
			case ast.Expr:
				arg = ExprString(a)
			case Object:
				arg = ObjectString(a, qs.qualifier)
			case Type:
				arg = TypeString(a, qs.qualifier)
			}
			formattedArgs[i] = arg
		}
		if !qs.ambiguous {
			break
		}
	}
	return fmt.Sprintf(format, formattedArgs...)
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
	check.error(at, code, check.sprintf(format, args...))
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
