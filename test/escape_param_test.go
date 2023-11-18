// test -v -gcflags=-l escape_param.go

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
		{name: "zero", f: func() { zero() }},
		{name: "param0 A", f: func() { i := gi; pi := &i; param0(pi) }},
		{name: "param0 B", f: func() { i := gi; pi := &i; esc(pi); param0(pi) }},
		{name: "caller0a", f: func() { caller0a() }},
		{name: "caller0b", f: func() { caller0b() }},
		{name: "param1 A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; param1(pi, pi2) }},
		{name: "param1 B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); param1(pi, pi2) }},
		{name: "param1 C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi2); param1(pi, pi2) }},
		{name: "param1 D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; esc(pi); esc(pi2); param1(pi, pi2) }},
		{name: "caller1", f: func() { caller1() }},
		{name: "param2 A", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; ppi2 := &pi2; param2(pi, ppi2) }},
		{name: "param2 B", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; ppi2 := &pi2; esc(pi); param2(pi, ppi2) }},
		{name: "param2 C", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; ppi2 := &pi2; esc(ppi2); param2(pi, ppi2) }},
		{name: "param2 D", f: func() { i := gi; pi := &i; i2 := gi; pi2 := &i2; ppi2 := &pi2; esc(pi); esc(ppi2); param2(pi, ppi2) }},
		{name: "caller2a", f: func() { caller2a() }},
		{name: "caller2b", f: func() { caller2b() }},
		{name: "paramArraySelfAssign A", f: func() { var p PairOfPairs; pp := &p; paramArraySelfAssign(pp) }},
		{name: "paramArraySelfAssign B", f: func() { var p PairOfPairs; pp := &p; esc(pp); paramArraySelfAssign(pp) }},
		{name: "paramArraySelfAssignUnsafeIndex A", f: func() { var p PairOfPairs; pp := &p; paramArraySelfAssignUnsafeIndex(pp) }},
		{name: "paramArraySelfAssignUnsafeIndex B", f: func() { var p PairOfPairs; pp := &p; esc(pp); paramArraySelfAssignUnsafeIndex(pp) }},
		{name: "leakParam A", f: func() { var x any = gi; leakParam(x) }},
		{name: "leakParam B", f: func() { var x any = gi; esc(x); leakParam(x) }},
		{name: "sinkAfterSelfAssignment1 A", allowPanic: true, f: func() { var b BoxedPair; pb := &b; sinkAfterSelfAssignment1(pb) }},
		{name: "sinkAfterSelfAssignment1 B", allowPanic: true, f: func() { var b BoxedPair; pb := &b; esc(pb); sinkAfterSelfAssignment1(pb) }},
		{name: "sinkAfterSelfAssignment2 A", allowPanic: true, f: func() { var b BoxedPair; pb := &b; sinkAfterSelfAssignment2(pb) }},
		{name: "sinkAfterSelfAssignment2 B", allowPanic: true, f: func() { var b BoxedPair; pb := &b; esc(pb); sinkAfterSelfAssignment2(pb) }},
		{name: "sinkAfterSelfAssignment3 A", allowPanic: true, f: func() { var b BoxedPair; pb := &b; sinkAfterSelfAssignment3(pb) }},
		{name: "sinkAfterSelfAssignment3 B", allowPanic: true, f: func() { var b BoxedPair; pb := &b; esc(pb); sinkAfterSelfAssignment3(pb) }},
		{name: "sinkAfterSelfAssignment4 A", allowPanic: true, f: func() { var b BoxedPair; pb := &b; sinkAfterSelfAssignment4(pb) }},
		{name: "sinkAfterSelfAssignment4 B", allowPanic: true, f: func() { var b BoxedPair; pb := &b; esc(pb); sinkAfterSelfAssignment4(pb) }},
		{name: "selfAssignmentAndUnrelated A", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; selfAssignmentAndUnrelated(pb, pb2) }},
		{name: "selfAssignmentAndUnrelated B", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; esc(pb); selfAssignmentAndUnrelated(pb, pb2) }},
		{name: "selfAssignmentAndUnrelated C", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; esc(pb2); selfAssignmentAndUnrelated(pb, pb2) }},
		{name: "selfAssignmentAndUnrelated D", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; esc(pb); esc(pb2); selfAssignmentAndUnrelated(pb, pb2) }},
		{name: "notSelfAssignment1 A", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; notSelfAssignment1(pb, pb2) }},
		{name: "notSelfAssignment1 B", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; esc(pb); notSelfAssignment1(pb, pb2) }},
		{name: "notSelfAssignment1 C", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; esc(pb2); notSelfAssignment1(pb, pb2) }},
		{name: "notSelfAssignment1 D", allowPanic: true, f: func() { var b BoxedPair; pb := &b; var b2 BoxedPair; pb2 := &b2; esc(pb); esc(pb2); notSelfAssignment1(pb, pb2) }},
		{name: "notSelfAssignment2 A", f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; notSelfAssignment2(pp, pp2) }},
		{name: "notSelfAssignment2 B", f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; esc(pp); notSelfAssignment2(pp, pp2) }},
		{name: "notSelfAssignment2 C", f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; esc(pp2); notSelfAssignment2(pp, pp2) }},
		{name: "notSelfAssignment2 D", f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; esc(pp); esc(pp2); notSelfAssignment2(pp, pp2) }},
		{name: "notSelfAssignment3 A", allowPanic: true, f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; notSelfAssignment3(pp, pp2) }},
		{name: "notSelfAssignment3 B", allowPanic: true, f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; esc(pp); notSelfAssignment3(pp, pp2) }},
		{name: "notSelfAssignment3 C", allowPanic: true, f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; esc(pp2); notSelfAssignment3(pp, pp2) }},
		{name: "notSelfAssignment3 D", allowPanic: true, f: func() { var p PairOfPairs; pp := &p; var p2 PairOfPairs; pp2 := &p2; esc(pp); esc(pp2); notSelfAssignment3(pp, pp2) }},
		{name: "boxedPairSelfAssign A", allowPanic: true, f: func() { var b BoxedPair; pb := &b; boxedPairSelfAssign(pb) }},
		{name: "boxedPairSelfAssign B", allowPanic: true, f: func() { var b BoxedPair; pb := &b; esc(pb); boxedPairSelfAssign(pb) }},
		{name: "wrappedPairSelfAssign A", f: func() { var w WrappedPair; pw := &w; wrappedPairSelfAssign(pw) }},
		{name: "wrappedPairSelfAssign B", f: func() { var w WrappedPair; pw := &w; esc(pw); wrappedPairSelfAssign(pw) }},
		{name: "param3 A", f: func() { var p Pair; pp := &p; param3(pp) }},
		{name: "param3 B", f: func() { var p Pair; pp := &p; esc(pp); param3(pp) }},
		{name: "caller3a", f: func() { caller3a() }},
		{name: "caller3b", f: func() { caller3b() }},
		{name: "caller4a", f: func() { caller4a() }},
		{name: "caller4b", f: func() { caller4b() }},
		{name: "param5 A", f: func() { i := gi; pi := &i; param5(pi) }},
		{name: "param5 B", f: func() { i := gi; pi := &i; esc(pi); param5(pi) }},
		{name: "caller5", f: func() { caller5() }},
		{name: "param6 A", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; param6(pppi) }},
		{name: "param6 B", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; esc(pppi); param6(pppi) }},
		{name: "caller6a", f: func() { caller6a() }},
		{name: "param7 A", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; param7(pppi) }},
		{name: "param7 B", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; esc(pppi); param7(pppi) }},
		{name: "caller7", f: func() { caller7() }},
		{name: "param8 A", f: func() { i := gi; pi := &i; ppi := &pi; param8(ppi) }},
		{name: "param8 B", f: func() { i := gi; pi := &i; ppi := &pi; esc(ppi); param8(ppi) }},
		{name: "caller8", f: func() { caller8() }},
		{name: "param9 A", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; param9(pppi) }},
		{name: "param9 B", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; esc(pppi); param9(pppi) }},
		{name: "caller9a", f: func() { caller9a() }},
		{name: "caller9b", f: func() { caller9b() }},
		{name: "param10 A", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; param10(pppi) }},
		{name: "param10 B", f: func() { i := gi; pi := &i; ppi := &pi; pppi := &ppi; esc(pppi); param10(pppi) }},
		{name: "caller10a", f: func() { caller10a() }},
		{name: "caller10b", f: func() { caller10b() }},
		{name: "param11 A", f: func() { i := gi; pi := &i; ppi := &pi; param11(ppi) }},
		{name: "param11 B", f: func() { i := gi; pi := &i; ppi := &pi; esc(ppi); param11(ppi) }},
		{name: "caller11a", f: func() { caller11a() }},
		{name: "caller11b", f: func() { caller11b() }},
		{name: "caller11c", f: func() { caller11c() }},
		{name: "caller11d", f: func() { caller11d() }},
		{name: "caller12a", f: func() { caller12a() }},
		{name: "caller12b", f: func() { caller12b() }},
		{name: "caller12c", f: func() { caller12c() }},
		{name: "caller12d", f: func() { caller12d() }},
		{name: "caller13a", f: func() { caller13a() }},
		{name: "caller13b", f: func() { caller13b() }},
		{name: "caller13c", f: func() { caller13c() }},
		{name: "caller13d", f: func() { caller13d() }},
		{name: "caller13e", f: func() { caller13e() }},
		{name: "caller13f", f: func() { caller13f() }},
		{name: "caller13g", f: func() { caller13g() }},
		{name: "caller13h", f: func() { caller13h() }},
		{name: "f A", f: func() { var x Node; px := &x; f(px) }},
		{name: "f B", f: func() { var x Node; px := &x; esc(px); f(px) }},
		{name: "g A", f: func() { var x Node; px := &x; g(px) }},
		{name: "g B", f: func() { var x Node; px := &x; esc(px); g(px) }},
		{name: "h A", f: func() { var x Node; px := &x; h(px) }},
		{name: "h B", f: func() { var x Node; px := &x; esc(px); h(px) }},
		{name: "param14a A", f: func() { var x [4]*int; param14a(x) }},
		{name: "param14a B", f: func() { var x [4]*int; esc(x); param14a(x) }},
		{name: "param14b A", f: func() { i := gi; pi := &i; param14b(pi) }},
		{name: "param14b B", f: func() { i := gi; pi := &i; esc(pi); param14b(pi) }},
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
