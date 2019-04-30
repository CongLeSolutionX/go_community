// errorcheck -0 -m -l -newescape=true

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test escape analysis for function parameters.

// In this test almost everything is BAD except the simplest cases
// where input directly flows to output.

package foo

import "fmt"

func f(buf []byte) []byte { // ERROR "leaking param: buf to result ~r1 level=0$"
	return buf
}

func g(*byte) string

func h(e int) {
	var x [32]byte // ERROR "moved to heap: x$"
	g(&f(x[:])[0])
}

type Node struct {
	s           string
	left, right *Node
}

func walk(np **Node) int { // ERROR "leaking param content: np$"
	n := *np
	w := len(n.s)
	if n == nil {
		return 0
	}
	wl := walk(&n.left)
	wr := walk(&n.right)
	if wl < wr {
		n.left, n.right = n.right, n.left // ERROR "walk ignoring self-assignment in n.left, n.right = n.right, n.left$"
		wl, wr = wr, wl
	}
	*np = n
	return w + wl + wr
}

// Test for bug where func var f used prototype's escape analysis results.
func prototype(xyz []string) {} // ERROR "prototype xyz does not escape$"
func bar() {
	var got [][]string
	f := prototype
	f = func(ss []string) { got = append(got, ss) } // ERROR "bar func literal does not escape$" "leaking param: ss$"
	s := "string"
	f([]string{s}) // ERROR "\[\]string literal escapes to heap$"
}

// Test for special treatment of arguments to fmt.Printf etc.
type fooi int

func (x fooi) String() string {
	return "I am a foo"
}

type bari int

func (x bari) NotString() string {
	return "I am a bar"
}

var one = 1
var you = "you"

func fmtcaller() {
	fmt.Printf("hi %s %s %d %d %s %v\n", "there", you, 1, one, fooi(11), bari(12)) // ERROR "fmtcaller ... argument does not escape$" "fmtcaller .there. does not escape$" "fmtcaller 1 does not escape$" "fmtcaller bari\(12\) does not escape$" "fmtcaller one does not escape$" "fmtcaller you does not escape$" "fooi\(11\) escapes to heap$"
}
