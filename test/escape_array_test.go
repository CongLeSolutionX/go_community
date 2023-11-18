// test -v -gcflags=-l escape_array.go

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package foo

import (
	"flag"
	"regexp"
	"runtime"
	"testing"
	"time"
)

var (
	recoverAllFlag  = flag.Bool("recoverall", false, "recover all panics; can be helpful with 'go test -v' to see which functions panic")
	funcTimeoutFlag = flag.Duration("functimeout", time.Minute, "fail if any function takes longer than `duration`")
	runNamesFlag    = flag.String("runnames", "", "test only the names matching `regexp`, ignoring any marked skip")
	listNamesFlag   = flag.String("listnames", "", "list names matching `regexp`, ignoring any marked skip")
)

func TestCallInputFunctions(t *testing.T) {
	tests := []struct {
		name       string
		f          func()
		skip       bool
		allowPanic bool
	}{
		{name: "bar A", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; bar(ps, ps2) }},
		{name: "bar B", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); bar(ps, ps2) }},
		{name: "bar C", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps2); bar(ps, ps2) }},
		{name: "bar D", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); esc(ps2); bar(ps, ps2) }},
		{name: "foo A", f: func() { var x U; foo(x) }},
		{name: "foo B", f: func() { var x U; esc(x); foo(x) }},
		{name: "bff A", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; bff(ps, ps2) }},
		{name: "bff B", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); bff(ps, ps2) }},
		{name: "bff C", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps2); bff(ps, ps2) }},
		{name: "bff D", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); esc(ps2); bff(ps, ps2) }},
		{name: "tbff1", f: func() { tbff1() }},
		{name: "tbff2", f: func() { tbff2() }},
		{name: "car A", f: func() { var x U; car(x) }},
		{name: "car B", f: func() { var x U; esc(x); car(x) }},
		{name: "fun A", f: func() { var x U; s := gs; ps := &s; fun(x, ps) }},
		{name: "fun B", f: func() { var x U; s := gs; ps := &s; esc(x); fun(x, ps) }},
		{name: "fun C", f: func() { var x U; s := gs; ps := &s; esc(ps); fun(x, ps) }},
		{name: "fun D", f: func() { var x U; s := gs; ps := &s; esc(x); esc(ps); fun(x, ps) }},
		{name: "fup A", f: func() { var x U; px := &x; s := gs; ps := &s; fup(px, ps) }},
		{name: "fup B", f: func() { var x U; px := &x; s := gs; ps := &s; esc(px); fup(px, ps) }},
		{name: "fup C", f: func() { var x U; px := &x; s := gs; ps := &s; esc(ps); fup(px, ps) }},
		{name: "fup D", f: func() { var x U; px := &x; s := gs; ps := &s; esc(px); esc(ps); fup(px, ps) }},
		{name: "fum A", f: func() { var x U; px := &x; s := gs; ps := &s; pps := &ps; fum(px, pps) }},
		{name: "fum B", f: func() { var x U; px := &x; s := gs; ps := &s; pps := &ps; esc(px); fum(px, pps) }},
		{name: "fum C", f: func() { var x U; px := &x; s := gs; ps := &s; pps := &ps; esc(pps); fum(px, pps) }},
		{name: "fum D", f: func() { var x U; px := &x; s := gs; ps := &s; pps := &ps; esc(px); esc(pps); fum(px, pps) }},
		{name: "fuo A", f: func() { var x U; px := &x; var y U; py := &y; fuo(px, py) }},
		{name: "fuo B", f: func() { var x U; px := &x; var y U; py := &y; esc(px); fuo(px, py) }},
		{name: "fuo C", f: func() { var x U; px := &x; var y U; py := &y; esc(py); fuo(px, py) }},
		{name: "fuo D", f: func() { var x U; px := &x; var y U; py := &y; esc(px); esc(py); fuo(px, py) }},
		{name: "hugeLeaks1 A", f: func() { s := gs; ps := &s; pps := &ps; s2 := gs; ps2 := &s2; pps2 := &ps2; hugeLeaks1(pps, pps2) }},
		{name: "hugeLeaks1 B", f: func() { s := gs; ps := &s; pps := &ps; s2 := gs; ps2 := &s2; pps2 := &ps2; esc(pps); hugeLeaks1(pps, pps2) }},
		{name: "hugeLeaks1 C", f: func() { s := gs; ps := &s; pps := &ps; s2 := gs; ps2 := &s2; pps2 := &ps2; esc(pps2); hugeLeaks1(pps, pps2) }},
		{name: "hugeLeaks1 D", f: func() { s := gs; ps := &s; pps := &ps; s2 := gs; ps2 := &s2; pps2 := &ps2; esc(pps); esc(pps2); hugeLeaks1(pps, pps2) }},
		{name: "hugeLeaks2 A", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; hugeLeaks2(ps, ps2) }},
		{name: "hugeLeaks2 B", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); hugeLeaks2(ps, ps2) }},
		{name: "hugeLeaks2 C", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps2); hugeLeaks2(ps, ps2) }},
		{name: "hugeLeaks2 D", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); esc(ps2); hugeLeaks2(ps, ps2) }},
		{name: "doesNew1 A", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; doesNew1(ps, ps2) }},
		{name: "doesNew1 B", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); doesNew1(ps, ps2) }},
		{name: "doesNew1 C", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps2); doesNew1(ps, ps2) }},
		{name: "doesNew1 D", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); esc(ps2); doesNew1(ps, ps2) }},
		{name: "doesNew2 A", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; doesNew2(ps, ps2) }},
		{name: "doesNew2 B", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); doesNew2(ps, ps2) }},
		{name: "doesNew2 C", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps2); doesNew2(ps, ps2) }},
		{name: "doesNew2 D", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); esc(ps2); doesNew2(ps, ps2) }},
		{name: "doesMakeSlice A", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; doesMakeSlice(ps, ps2) }},
		{name: "doesMakeSlice B", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); doesMakeSlice(ps, ps2) }},
		{name: "doesMakeSlice C", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps2); doesMakeSlice(ps, ps2) }},
		{name: "doesMakeSlice D", f: func() { s := gs; ps := &s; s2 := gs; ps2 := &s2; esc(ps); esc(ps2); doesMakeSlice(ps, ps2) }},
		{name: "nonconstArray", f: func() { nonconstArray() }},
	}

	listNamesRe := regexp.MustCompile(*listNamesFlag)
	runNamesRe := regexp.MustCompile(*runNamesFlag)
	// Run our test functions one after another in a single goroutine
	// that calls runtime.GC at the end. In some cases, this might help
	// the runtime recognize an illegal heap pointer to a stack variable.
	starting := make(chan string)
	done := make(chan struct{})
	go func() {
		for _, tt := range tests {
			if tt.skip {
				continue
			}
			if *listNamesFlag != "" && listNamesRe.MatchString(tt.name) {
				t.Log("func name:", tt.name)
				continue
			}

			if runNamesRe.MatchString(tt.name) {
				starting <- tt.name
				func() {
					defer func() {
						if r := recover(); r != nil {
							if !tt.allowPanic && !*recoverAllFlag {
								t.Log("panic in:", tt.name)
								panic(r)
							}
							t.Log("recovered panic:", tt.name)
						}
					}()

					// Call the function under test.
					tt.f()
				}()
			}
		}

		for i := 0; i < 5; i++ {
			runtime.GC()
		}
		close(done)
	}()

	var current string
	for {
		select {
		case current = <-starting:
			t.Log("starting func:", current)
		case <-time.After(*funcTimeoutFlag):
			t.Fatal("timed out func:", current)
		case <-done:
			return
		}
	}
}

// Global int and string values to help avoid some heap allocations being optimized away
// via the compiler or runtime recognizing constants or small integer values.
var (
	gi = 1000
	gs = "abcd"
)

// esc forces x to escape.
func esc(x any) {
	if escSink.b {
		escSink.x = x
	}
}

var escSink struct {
	b bool
	x any
}
