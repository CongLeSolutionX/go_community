// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT (use 'go test -v -update-expected' instead.)
// See cmd/compile/internal/inline/inlheur/testdata/props/README.txt
// for more information on the format of this file.
// =^=^=
package calls

import "os"

var G int

func callee(x int) int {
	return x
}

func callsexit(x int) {
	println(x)
	os.Exit(x)
}

// calls.go T_calls_in_loops 31
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[]}
// callsite: calls.go:33:9|0 "CallSiteInLoop" 1
// callsite: calls.go:36:9|1 "CallSiteInLoop" 1
// =+=+=
// =-=-=
func T_calls_in_loops(x int, q []string) {
	for i := 0; i < x; i++ {
		callee(i)
	}
	for _, s := range q {
		callee(len(s))
	}
}

//
//	0: ParamFeedsIfOrSwitch
//	1: ParamNoInfo
//
// calls.go T_calls_on_panic_paths 52
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[]}
// callsite: calls.go:54:9|0 "CallSiteOnPanicPath" 2
// callsite: calls.go:58:9|1 "CallSiteOnPanicPath" 2
// callsite: calls.go:62:12|2 "CallSiteOnPanicPath" 2
// =+=+=
// =-=-=
func T_calls_on_panic_paths(x int, q []string) {
	if x+G == 101 {
		callee(x)
		panic("ouch")
	}
	if x < G-101 {
		callee(x)
		if len(q) == 0 {
			G++
		}
		callsexit(x)
	}
}

//
//	0: ParamFeedsIfOrSwitch
//	1: ParamNoInfo
//
// calls.go T_calls_not_on_panic_paths 83
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
//   1: ParamNoInfo
// =====
// {"Flags":0,"RecvrParamFlags":[8,0],"ReturnFlags":[]}
// callsite: calls.go:102:9|2 "" 0
// callsite: calls.go:105:9|3 "" 0
// callsite: calls.go:109:12|4 "CallSiteOnPanicPath" 2
// callsite: calls.go:90:9|0 "" 0
// callsite: calls.go:93:9|1 "" 0
// =+=+=
// =-=-=
func T_calls_not_on_panic_paths(x int, q []string) {
	if x != G {
		panic("ouch")
		/* Notes: */
		/* - we only look for post-dominating panic/exit, so */
		/*   this site will on fact not have a panicpath flag */
		/* - vet will complain about this site as unreachable */
		callee(x)
	}
	if x != G {
		callee(x)
		if x < 100 {
			panic("ouch")
		}
	}
	if x+G == 101 {
		if x < 100 {
			panic("ouch")
		}
		callee(x)
	}
	if x < -101 {
		callee(x)
		if len(q) == 0 {
			return
		}
		callsexit(x)
	}
}
