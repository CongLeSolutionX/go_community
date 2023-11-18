// test -v -gcflags=-l escape_closure.go

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package escape

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
		{name: "ClosureCallArgs0", f: func() { ClosureCallArgs0() }},
		{name: "ClosureCallArgs1", skip: true, f: func() { ClosureCallArgs1() }},
		{name: "ClosureCallArgs2", skip: true, f: func() { ClosureCallArgs2() }},
		{name: "ClosureCallArgs3", f: func() { ClosureCallArgs3() }},
		{name: "ClosureCallArgs4", f: func() { ClosureCallArgs4() }},
		{name: "ClosureCallArgs5", f: func() { ClosureCallArgs5() }},
		{name: "ClosureCallArgs6", f: func() { ClosureCallArgs6() }},
		{name: "ClosureCallArgs7", skip: true, f: func() { ClosureCallArgs7() }},
		{name: "ClosureCallArgs8", f: func() { ClosureCallArgs8() }},
		{name: "ClosureCallArgs9", skip: true, f: func() { ClosureCallArgs9() }},
		{name: "ClosureCallArgs10", skip: true, f: func() { ClosureCallArgs10() }},
		{name: "ClosureCallArgs11", f: func() { ClosureCallArgs11() }},
		{name: "ClosureCallArgs12", f: func() { ClosureCallArgs12() }},
		{name: "ClosureCallArgs13", f: func() { ClosureCallArgs13() }},
		{name: "ClosureCallArgs14", f: func() { ClosureCallArgs14() }},
		{name: "ClosureCallArgs15", f: func() { ClosureCallArgs15() }},
		{name: "ClosureLeak1 A", f: func() { s := gs; ClosureLeak1(s) }},
		{name: "ClosureLeak1 B", f: func() { s := gs; esc(s); ClosureLeak1(s) }},
		{name: "ClosureLeak2 A", f: func() { s := gs; ClosureLeak2(s) }},
		{name: "ClosureLeak2 B", f: func() { s := gs; esc(s); ClosureLeak2(s) }},
		{name: "ClosureLeak2b A", allowPanic: true, f: func() { var f func() string; ClosureLeak2b(f) }},
		{name: "ClosureLeak2b B", allowPanic: true, f: func() { var f func() string; esc(f); ClosureLeak2b(f) }},
		{name: "ClosureIndirect", f: func() { ClosureIndirect() }},
		{name: "nopFunc A", f: func() { i := gi; pi := &i; nopFunc(pi) }},
		{name: "nopFunc B", f: func() { i := gi; pi := &i; esc(pi); nopFunc(pi) }},
		{name: "ClosureIndirect2", f: func() { ClosureIndirect2() }},
		{name: "nopFunc2 A", f: func() { i := gi; pi := &i; nopFunc2(pi) }},
		{name: "nopFunc2 B", f: func() { i := gi; pi := &i; esc(pi); nopFunc2(pi) }},
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
