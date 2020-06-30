// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package constraint

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var exprStringTests = []struct {
	x   Expr
	out string
}{
	{
		x:   tag("abc"),
		out: "abc",
	},
	{
		x:   not(tag("abc")),
		out: "!abc",
	},
	{
		x:   not(and(tag("abc"), tag("def"))),
		out: "!(abc && def)",
	},
	{
		x:   and(tag("abc"), or(tag("def"), tag("ghi"))),
		out: "abc && (def || ghi)",
	},
	{
		x:   or(and(tag("abc"), tag("def")), tag("ghi")),
		out: "(abc && def) || ghi",
	},
}

func TestExprString(t *testing.T) {
	for i, tt := range exprStringTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			s := tt.x.String()
			if s != tt.out {
				t.Errorf("String() mismatch:\nhave %s\nwant %s", s, tt.out)
			}
		})
	}
}

var lexTests = []struct {
	in  string
	out string
}{
	{"", ""},
	{"x", "x"},
	{"x.y", "x.y"},
	{"x_y", "x_y"},
	{"αx", "αx"},
	{"αx²", "αx err: invalid syntax at ²"},
	{"go1.2", "go1.2"},
	{"x y", "x y"},
	{"x!y", "x ! y"},
	{"&&||!()xy yx ", "&& || ! ( ) xy yx"},
	{"x~", "x err: invalid syntax at ~"},
	{"x ~", "x err: invalid syntax at ~"},
	{"x &", "x err: invalid syntax at &"},
	{"x &y", "x err: invalid syntax at &"},
}

func TestLex(t *testing.T) {
	for i, tt := range lexTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			p := &exprParser{s: tt.in}
			out := ""
			for {
				tok, err := lexHelp(p)
				if tok == "" && err == nil {
					break
				}
				if out != "" {
					out += " "
				}
				if err != nil {
					out += "err: " + err.Error()
					break
				}
				out += tok
			}
			if out != tt.out {
				t.Errorf("lex(%q):\nhave %s\nwant %s", tt.in, out, tt.out)
			}
		})
	}
}

func lexHelp(p *exprParser) (tok string, err error) {
	defer func() {
		if e := recover(); e != nil {
			if e, ok := e.(*SyntaxError); ok {
				err = e
				return
			}
			panic(e)
		}
	}()

	p.lex()
	return p.tok, nil
}

var parseExprTests = []struct {
	in string
	x  Expr
}{
	{"x", tag("x")},
	{"x&&y", and(tag("x"), tag("y"))},
	{"x||y", or(tag("x"), tag("y"))},
	{"(x)", tag("x")},
	{"x||y&&z", or(tag("x"), and(tag("y"), tag("z")))},
	{"x&&y||z", or(and(tag("x"), tag("y")), tag("z"))},
	{"x&&(y||z)", and(tag("x"), or(tag("y"), tag("z")))},
	{"(x||y)&&z", and(or(tag("x"), tag("y")), tag("z"))},
	{"!(x&&y)", not(and(tag("x"), tag("y")))},
}

func TestParseExpr(t *testing.T) {
	for i, tt := range parseExprTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			x, err := parseExpr(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if x.String() != tt.x.String() {
				t.Errorf("parseExpr(%q):\nhave %s\nwant %s", tt.in, x, tt.x)
			}
		})
	}
}

var parseExprErrorTests = []struct {
	in  string
	err error
}{
	{"x && ", &SyntaxError{Offset: 5, msg: "unexpected end of expression"}},
	{"x && (", &SyntaxError{Offset: 6, msg: "missing close paren"}},
	{"x && ||", &SyntaxError{Offset: 5, msg: "unexpected token ||"}},
	{"x && !", &SyntaxError{Offset: 6, msg: "unexpected end of expression"}},
	{"x && !!", &SyntaxError{Offset: 6, msg: "double negation not allowed"}},
	{"x !", &SyntaxError{Offset: 2, msg: "unexpected token !"}},
	{"x && (y", &SyntaxError{Offset: 5, msg: "missing close paren"}},
}

func TestParseError(t *testing.T) {
	for i, tt := range parseExprErrorTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			x, err := parseExpr(tt.in)
			if err == nil {
				t.Fatalf("parseExpr(%q) = %v, want error", tt.in, x)
			}
			if !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("parseExpr(%q): wrong error:\nhave %#v\nwant %#v", tt.in, err, tt.err)
			}
		})
	}
}

var exprEvalTests = []struct {
	in   string
	ok   bool
	tags string
}{
	{"x", false, "x"},
	{"x && y", false, "x y"},
	{"x || y", false, "x y"},
	{"!x && yes", true, "x yes"},
	{"yes || y", true, "y yes"},
}

func TestExprEval(t *testing.T) {
	for i, tt := range exprEvalTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			x, err := parseExpr(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			tags := make(map[string]bool)
			wantTags := make(map[string]bool)
			for _, tag := range strings.Fields(tt.tags) {
				wantTags[tag] = true
			}
			hasTag := func(tag string) bool {
				tags[tag] = true
				return tag == "yes"
			}
			ok := x.Eval(hasTag)
			if ok != tt.ok || !reflect.DeepEqual(tags, wantTags) {
				t.Errorf("Eval(%#q):\nhave ok=%v, tags=%v\nwant ok=%v, tags=%v",
					tt.in, ok, tags, tt.ok, wantTags)
			}
		})
	}
}
