// test -v -gcflags=-l escape_field.go

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
		{name: "field0", f: func() { field0() }},
		{name: "field1", f: func() { field1() }},
		{name: "field3", f: func() { field3() }},
		{name: "field4", f: func() { field4() }},
		{name: "field5", f: func() { field5() }},
		{name: "field6 A", f: func() { var x X; px := &x; field6(px) }},
		{name: "field6 B", f: func() { var x X; px := &x; esc(px); field6(px) }},
		{name: "field6a", f: func() { field6a() }},
		{name: "field7", f: func() { field7() }},
		{name: "field8", f: func() { field8() }},
		{name: "field9", f: func() { field9() }},
		{name: "field10", f: func() { field10() }},
		{name: "field11", f: func() { field11() }},
		{name: "field12", f: func() { field12() }},
		{name: "field13", f: func() { field13() }},
		{name: "field14", f: func() { field14() }},
		{name: "field15", f: func() { field15() }},
		{name: "field16", f: func() { field16() }},
		{name: "field17", f: func() { field17() }},
		{name: "field18", f: func() { field18() }},
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
