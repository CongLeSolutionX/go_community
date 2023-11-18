// test -v -gcflags=-l escape_slice.go

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
		{name: "slice0", f: func() { slice0() }},
		{name: "slice1", f: func() { slice1() }},
		{name: "slice2", f: func() { slice2() }},
		{name: "slice3", f: func() { slice3() }},
		{name: "slice4 A", f: func() { s := make([]*int, 10); slice4(s) }},
		{name: "slice4 B", f: func() { s := make([]*int, 10); esc(s); slice4(s) }},
		{name: "slice5 A", f: func() { s := make([]*int, 10); slice5(s) }},
		{name: "slice5 B", f: func() { s := make([]*int, 10); esc(s); slice5(s) }},
		{name: "slice6", f: func() { slice6() }},
		{name: "slice7", f: func() { slice7() }},
		{name: "slice8", f: func() { slice8() }},
		{name: "slice9", f: func() { slice9() }},
		{name: "slice10", f: func() { slice10() }},
		{name: "slice11", allowPanic: true, f: func() { slice11() }},
		{name: "slice12 A", f: func() { x := make([]int, 10); slice12(x) }},
		{name: "slice12 B", f: func() { x := make([]int, 10); esc(x); slice12(x) }},
		{name: "slice13 A", f: func() { x := make([]*int, 10); slice13(x) }},
		{name: "slice13 B", f: func() { x := make([]*int, 10); esc(x); slice13(x) }},
		{name: "envForDir A", f: func() { s := gs; envForDir(s) }},
		{name: "envForDir B", f: func() { s := gs; esc(s); envForDir(s) }},
		{name: "mergeEnvLists A", f: func() { i := make([]string, 10); o := make([]string, 10); mergeEnvLists(i, o) }},
		{name: "mergeEnvLists B", f: func() { i := make([]string, 10); o := make([]string, 10); esc(i); mergeEnvLists(i, o) }},
		{name: "mergeEnvLists C", f: func() { i := make([]string, 10); o := make([]string, 10); esc(o); mergeEnvLists(i, o) }},
		{name: "mergeEnvLists D", f: func() { i := make([]string, 10); o := make([]string, 10); esc(i); esc(o); mergeEnvLists(i, o) }},
		{name: "IPv4", f: func() { var b byte; var b2 byte; var b3 byte; var b4 byte; IPv4(b, b2, b3, b4) }},
		{name: "setupTestData", f: func() { setupTestData() }},
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
