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

// calls.go init.0 30
// =====
// {"Flags":0,"RecvrParamFlags":[],"ReturnFlags":[]}
// callsite: calls.go:31:16|0 "CallSiteInInitFunc" 4
// =+=+=
// =-=-=
func init() {
	println(callee(5))
}

// calls.go T_call_in_panic_arg 40
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// callsite: calls.go:42:15|0 "CallSiteOnPanicPath" 2
// =+=+=
// =-=-=
func T_call_in_panic_arg(x int) {
	if x < G {
		panic(callee(x))
	}
}

// calls.go T_calls_in_loops 53
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[]}
// callsite: calls.go:55:9|0 "CallSiteInLoop" 1
// callsite: calls.go:58:9|1 "CallSiteInLoop" 1
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

// calls.go T_calls_on_panic_paths 70
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[]}
// callsite: calls.go:72:9|0 "CallSiteOnPanicPath" 2
// callsite: calls.go:76:9|1 "CallSiteOnPanicPath" 2
// callsite: calls.go:80:12|2 "CallSiteOnPanicPath" 2
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

// calls.go T_calls_not_on_panic_paths 97
// RecvrParamFlags
//   0 ParamFeedsIfOrSwitch
//   1 ParamNoInfo
// =====
// {"Flags":0,"RecvrParamFlags":[8,0],"ReturnFlags":[]}
// callsite: calls.go:104:9|0 "" 0
// callsite: calls.go:107:9|1 "" 0
// callsite: calls.go:116:9|2 "" 0
// callsite: calls.go:119:9|3 "" 0
// callsite: calls.go:123:12|4 "CallSiteOnPanicPath" 2
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
