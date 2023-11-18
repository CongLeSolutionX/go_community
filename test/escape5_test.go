// test -v -gcflags=-l escape5.go

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
		{name: "noleak A", f: func() { i := gi; pi := &i; noleak(pi) }},
		{name: "noleak B", f: func() { i := gi; pi := &i; esc(pi); noleak(pi) }},
		{name: "leaktoret A", f: func() { i := gi; pi := &i; leaktoret(pi) }},
		{name: "leaktoret B", f: func() { i := gi; pi := &i; esc(pi); leaktoret(pi) }},
		{name: "leaktoret2 A", f: func() { i := gi; pi := &i; leaktoret2(pi) }},
		{name: "leaktoret2 B", f: func() { i := gi; pi := &i; esc(pi); leaktoret2(pi) }},
		{name: "leaktoret22 A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leaktoret22(pi, pi2) }},
		{name: "leaktoret22 B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leaktoret22(pi, pi2) }},
		{name: "leaktoret22 C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leaktoret22(pi, pi2) }},
		{name: "leaktoret22 D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leaktoret22(pi, pi2) }},
		{name: "leaktoret22b A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leaktoret22b(pi, pi2) }},
		{name: "leaktoret22b B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leaktoret22b(pi, pi2) }},
		{name: "leaktoret22b C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leaktoret22b(pi, pi2) }},
		{name: "leaktoret22b D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leaktoret22b(pi, pi2) }},
		{name: "leaktoret22c A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leaktoret22c(pi, pi2) }},
		{name: "leaktoret22c B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leaktoret22c(pi, pi2) }},
		{name: "leaktoret22c C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leaktoret22c(pi, pi2) }},
		{name: "leaktoret22c D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leaktoret22c(pi, pi2) }},
		{name: "leaktoret22d A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leaktoret22d(pi, pi2) }},
		{name: "leaktoret22d B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leaktoret22d(pi, pi2) }},
		{name: "leaktoret22d C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leaktoret22d(pi, pi2) }},
		{name: "leaktoret22d D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leaktoret22d(pi, pi2) }},
		{name: "leaktoret22e A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leaktoret22e(pi, pi2) }},
		{name: "leaktoret22e B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leaktoret22e(pi, pi2) }},
		{name: "leaktoret22e C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leaktoret22e(pi, pi2) }},
		{name: "leaktoret22e D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leaktoret22e(pi, pi2) }},
		{name: "leaktoret22f A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leaktoret22f(pi, pi2) }},
		{name: "leaktoret22f B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leaktoret22f(pi, pi2) }},
		{name: "leaktoret22f C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leaktoret22f(pi, pi2) }},
		{name: "leaktoret22f D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leaktoret22f(pi, pi2) }},
		{name: "leaktosink A", f: func() { i := gi; pi := &i; leaktosink(pi) }},
		{name: "leaktosink B", f: func() { i := gi; pi := &i; esc(pi); leaktosink(pi) }},
		{name: "f1", f: func() { f1() }},
		{name: "f2", f: func() { f2() }},
		{name: "f3", f: func() { f3() }},
		{name: "f4", f: func() { f4() }},
		{name: "f5", f: func() { f5() }},
		{name: "f6", f: func() { f6() }},
		{name: "f7", f: func() { f7() }},
		{name: "leakrecursive1 A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leakrecursive1(pi, pi2) }},
		{name: "leakrecursive1 B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leakrecursive1(pi, pi2) }},
		{name: "leakrecursive1 C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leakrecursive1(pi, pi2) }},
		{name: "leakrecursive1 D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leakrecursive1(pi, pi2) }},
		{name: "leakrecursive2 A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; leakrecursive2(pi, pi2) }},
		{name: "leakrecursive2 B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); leakrecursive2(pi, pi2) }},
		{name: "leakrecursive2 C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); leakrecursive2(pi, pi2) }},
		{name: "leakrecursive2 D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); leakrecursive2(pi, pi2) }},
		{name: "f8 A", f: func() { var p T1; pp := &p; f8(pp) }},
		{name: "f8 B", f: func() { var p T1; pp := &p; esc(pp); f8(pp) }},
		{name: "f9", f: func() { f9() }},
		{name: "f10", skip: true, f: func() { f10() }},
		{name: "f11 A", f: func() { i := gi; pi := &i; ppi := &pi; f11(ppi) }},
		{name: "f11 B", f: func() { i := gi; pi := &i; ppi := &pi; esc(ppi); f11(ppi) }},
		{name: "f12 A", f: func() { i := gi; pi := &i; ppi := &pi; f12(ppi) }},
		{name: "f12 B", f: func() { i := gi; pi := &i; ppi := &pi; esc(ppi); f12(ppi) }},
		{name: "f13", f: func() { f13() }},
		{name: "fbad24305a", f: func() { fbad24305a() }},
		{name: "fbad24305b", f: func() { fbad24305b() }},
		{name: "f29000 A", f: func() { i := gi; var x any = gi; f29000(i, x) }},
		{name: "f29000 C", f: func() { i := gi; var x any = gi; esc(x); f29000(i, x) }},
		{name: "g29000", f: func() { g29000() }},
		{name: "f28369", f: func() { i := gi; f28369(i) }},
		{name: "f A", f: func() { i := gi; pi := &i; f(pi) }},
		{name: "f B", f: func() { i := gi; pi := &i; esc(pi); f(pi) }},
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
