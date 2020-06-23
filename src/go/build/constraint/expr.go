// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package constraint implements parsing and evaluation of build tag constraint lines.
// See TODO URL for details about build constraints.
package constraint

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// An Expr is a build tag constraint expression.
// The underlying concrete type is *AndExpr, *OrExpr, *NotExpr, or *TagExpr.
type Expr interface {
	Eval(ok func(tag string) bool) bool
	String() string
	isExpr()
}

type TagExpr struct {
	Tag string
}

func (x *TagExpr) isExpr() {}

func (x *TagExpr) Eval(ok func(tag string) bool) bool {
	return ok(x.Tag)
}

func (x *TagExpr) String() string {
	return x.Tag
}

func tag(tag string) Expr { return &TagExpr{tag} }

type NotExpr struct {
	X Expr
}

func (x *NotExpr) isExpr() {}

func (x *NotExpr) Eval(ok func(tag string) bool) bool {
	return !x.X.Eval(ok)
}

func (x *NotExpr) String() string {
	s := x.X.String()
	switch x.X.(type) {
	case *AndExpr, *OrExpr:
		s = "(" + s + ")"
	}
	return "!" + s
}

func not(x Expr) Expr { return &NotExpr{x} }

type AndExpr struct {
	X, Y Expr
}

func (x *AndExpr) isExpr() {}

func (x *AndExpr) Eval(ok func(tag string) bool) bool {
	// Note: Eval both, to make sure ok func observes all tags.
	xok := x.X.Eval(ok)
	yok := x.Y.Eval(ok)
	return xok && yok
}

func (x *AndExpr) String() string {
	return andArg(x.X) + " && " + andArg(x.Y)
}

func andArg(x Expr) string {
	s := x.String()
	if _, ok := x.(*OrExpr); ok {
		s = "(" + s + ")"
	}
	return s
}

func and(x, y Expr) Expr {
	return &AndExpr{x, y}
}

type OrExpr struct {
	X, Y Expr
}

func (x *OrExpr) isExpr() {}

func (x *OrExpr) Eval(ok func(tag string) bool) bool {
	// Note: Eval both, to make sure ok func observes all tags.
	xok := x.X.Eval(ok)
	yok := x.Y.Eval(ok)
	return xok || yok
}

func (x *OrExpr) String() string {
	return orArg(x.X) + " || " + orArg(x.Y)
}

func orArg(x Expr) string {
	s := x.String()
	if _, ok := x.(*AndExpr); ok {
		s = "(" + s + ")"
	}
	return s
}

func or(x, y Expr) Expr {
	return &OrExpr{x, y}
}

type exprParser struct {
	s string // input string
	i int    // next read location in s

	tok   string // last token read
	isTag bool
	pos   int // position (start) of last token
}

type SyntaxError struct {
	Offset int
	msg    string
}

func (e *SyntaxError) Error() string {
	return e.msg
}

func TODO() { panic("TODO") }

func (p *exprParser) lex() {
	p.isTag = false
	for p.i < len(p.s) && (p.s[p.i] == ' ' || p.s[p.i] == '\t') {
		p.i++
	}
	if p.i >= len(p.s) {
		p.tok = ""
		p.pos = p.i
		return
	}
	switch p.s[p.i] {
	case '(', ')', '!':
		p.pos = p.i
		p.i++
		p.tok = p.s[p.pos:p.i]
		return

	case '&', '|':
		if p.i+1 >= len(p.s) || p.s[p.i+1] != p.s[p.i] {
			panic(&SyntaxError{Offset: p.i, msg: "invalid syntax at " + string(rune(p.s[p.i]))})
		}
		p.pos = p.i
		p.i += 2
		p.tok = p.s[p.pos:p.i]
		return
	}

	tag := p.s[p.i:]
	for i, c := range tag {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' && c != '.' {
			tag = tag[:i]
			break
		}
	}
	if tag == "" {
		c, _ := utf8.DecodeRuneInString(p.s[p.i:])
		panic(&SyntaxError{Offset: p.i, msg: "invalid syntax at " + string(c)})
	}

	p.pos = p.i
	p.i += len(tag)
	p.tok = p.s[p.pos:p.i]
	p.isTag = true
	return
}

// parseExpr parses a boolean build tag expression.
func parseExpr(text string) (x Expr, err error) {
	defer func() {
		if e := recover(); e != nil {
			if e, ok := e.(*SyntaxError); ok {
				err = e
				return
			}
			panic(e) // unreachable unless parser has a bug
		}
	}()

	p := &exprParser{s: text}
	x = p.or()
	if p.tok != "" {
		panic(&SyntaxError{Offset: p.pos, msg: "unexpected token " + p.tok})
	}
	return x, nil
}

// or parses a sequence of || expressions.
// On entry, the next input token has not yet been lexed.
// On exit, the next input token has been lexed and is in p.tok.
func (p *exprParser) or() Expr {
	x := p.and()
	for p.tok == "||" {
		x = or(x, p.and())
	}
	return x
}

// and parses a sequence of && expressions.
// On entry, the next input token has not yet been lexed.
// On exit, the next input token has been lexed and is in p.tok.
func (p *exprParser) and() Expr {
	x := p.not()
	for p.tok == "&&" {
		x = and(x, p.not())
	}
	return x
}

// not parses a ! expression.
// On entry, the next input token has not yet been lexed.
// On exit, the next input token has been lexed and is in p.tok.
func (p *exprParser) not() Expr {
	p.lex()
	if p.tok == "!" {
		p.lex()
		if p.tok == "!" {
			panic(&SyntaxError{Offset: p.pos, msg: "double negation not allowed"})
		}
		return not(p.atom())
	}
	return p.atom()
}

// atom parses a tag or a parenthesized expression.
// On entry, the next input token HAS been lexed.
// On exit, the next input token has been lexed and is in p.tok.
func (p *exprParser) atom() Expr {
	// first token already in p.tok
	if p.tok == "(" {
		pos := p.pos
		defer func() {
			if e := recover(); e != nil {
				if e, ok := e.(*SyntaxError); ok && e.msg == "unexpected end of expression" {
					e.msg = "missing close paren"
				}
				panic(e)
			}
		}()
		x := p.or()
		if p.tok != ")" {
			panic(&SyntaxError{Offset: pos, msg: "missing close paren"})
		}
		p.lex()
		return x
	}

	if !p.isTag {
		if p.tok == "" {
			panic(&SyntaxError{Offset: p.pos, msg: "unexpected end of expression"})
		}
		panic(&SyntaxError{Offset: p.pos, msg: "unexpected token " + p.tok})
	}
	tok := p.tok
	p.lex()
	return tag(tok)
}

var errNotConstraint = errors.New("not a build constraint")

// Parse parses a single build constraint line of the form "//go:build ..." or "// +build ..."
// and returns the corresponding boolean expression.
func Parse(line string) (Expr, error) {
	if text, ok := splitGoBuild(line); ok {
		return parseExpr(text)
	}
	if text, ok := splitPlusBuild(line); ok {
		return parsePlusBuildExpr(text), nil
	}
	return nil, errNotConstraint
}

// IsGoBuild reports whether the line of text is a "//go:build" constraint.
// It only checks the prefix of the text, not that the expression itself parses.
func IsGoBuild(line string) bool {
	_, ok := splitGoBuild(line)
	return ok
}

func splitGoBuild(line string) (expr string, ok bool) {
	// A single trailing newline is OK; otherwise multiple lines are not.
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if strings.Contains(line, "\n") {
		return "", false
	}

	if !strings.HasPrefix(line, "//go:build") {
		return "", false
	}

	line = strings.TrimSpace(line)
	line = line[len("//go:build"):]

	// If strings.TrimSpace finds more to trim after removing the //go:build prefix,
	// it means that the prefix was followed by a space, making this a //go:build line
	// (as opposed to a //go:buildsomethingelse line).
	// If line is empty, we had "//go:build" by itself, which also counts.
	trim := strings.TrimSpace(line)
	if len(line) == len(trim) && line != "" {
		return "", false
	}

	return trim, true
}

// IsPlusBuild reports whether the line of text is a "// +build" constraint.
// It only checks the prefix of the text, not that the expression itself parses.
func IsPlusBuild(line string) bool {
	_, ok := splitPlusBuild(line)
	return ok
}

func splitPlusBuild(line string) (expr string, ok bool) {
	// A single trailing newline is OK; otherwise multiple lines are not.
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if strings.Contains(line, "\n") {
		return "", false
	}

	if !strings.HasPrefix(line, "//") {
		return "", false
	}
	line = line[len("//"):]
	// Note the space is optional; "//+build" is recognized too.
	line = strings.TrimSpace(line)

	if !strings.HasPrefix(line, "+build") {
		return "", false
	}
	line = line[len("+build"):]

	// If strings.TrimSpace finds more to trim after removing the +build prefix,
	// it means that the prefix was followed by a space, making this a +build line
	// (as opposed to a +buildsomethingelse line).
	// If line is empty, we had "// +build" by itself, which also counts.
	trim := strings.TrimSpace(line)
	if len(line) == len(trim) && line != "" {
		return "", false
	}

	return trim, true
}

func parsePlusBuildExpr(text string) Expr {
	var x Expr
	for _, clause := range strings.Fields(text) {
		var y Expr
		for _, lit := range strings.Split(clause, ",") {
			var z Expr
			var neg bool
			if strings.HasPrefix(lit, "!!") || lit == "!" {
				z = tag("ignore")
			} else {
				if strings.HasPrefix(lit, "!") {
					neg = true
					lit = lit[len("!"):]
				}
				if isValidTag(lit) {
					z = tag(lit)
				} else {
					z = tag("ignore")
				}
				if neg {
					z = not(z)
				}
			}
			if y == nil {
				y = z
			} else {
				y = and(y, z)
			}
		}
		if x == nil {
			x = y
		} else {
			x = or(x, y)
		}
	}
	return x
}

// Tags must be letters, digits, underscores or dots.
// Unlike in Go identifiers, all digits are fine (e.g., "386").
func isValidTag(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' && c != '.' {
			return false
		}
	}
	return true
}
