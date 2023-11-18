// test -v -gcflags=-l escape_indir.go

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
		{name: "constptr0", f: func() { constptr0() }},
		{name: "constptr01", f: func() { constptr01() }},
		{name: "constptr02", f: func() { constptr02() }},
		{name: "constptr03", f: func() { constptr03() }},
		{name: "constptr1", f: func() { constptr1() }},
		{name: "constptr2", f: func() { constptr2() }},
		{name: "constptr4", f: func() { constptr4() }},
		{name: "constptr5", f: func() { constptr5() }},
		{name: "constptr6 A", f: func() { var p ConstPtr; pp := &p; constptr6(pp) }},
		{name: "constptr6 B", f: func() { var p ConstPtr; pp := &p; esc(pp); constptr6(pp) }},
		{name: "constptr7", f: func() { constptr7() }},
		{name: "constptr8", f: func() { constptr8() }},
		{name: "constptr9", f: func() { constptr9() }},
		{name: "constptr10", f: func() { constptr10() }},
		{name: "constptr11", f: func() { constptr11() }},
		{name: "foo A", f: func() { i := gi; pi := &i; ppi := &pi; foo(ppi) }},
		{name: "foo B", f: func() { i := gi; pi := &i; ppi := &pi; esc(ppi); foo(ppi) }},
		{name: "foo1 A", f: func() { i := gi; pi := &i; foo1(pi) }},
		{name: "foo1 B", f: func() { i := gi; pi := &i; esc(pi); foo1(pi) }},
		{name: "foo2", f: func() { foo2() }},
		{name: "f", f: func() { f() }},
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
