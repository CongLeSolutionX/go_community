// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT (use 'go test -v -update-expected' instead.)
// See cmd/compile/internal/inline/inlheur/testdata/props/README.txt
// for more information on the format of this file.
// =^=^=

package funcflags

import "os"

// funcflags.go T_simple 20
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":[],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_simple() {
	panic("bad")
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_nested 35
// Flags: FuncPropUnconditionalPanicExit
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":1,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_nested(x int) {
	if x < 10 {
		panic("bad")
	} else {
		panic("good")
	}
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_block1 54
// Flags: FuncPropUnconditionalPanicExit
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":1,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_block1(x int) {
	panic("bad")
	if x < 10 {
		return
	}
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_block2 71
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_block2(x int) {
	if x < 10 {
		return
	}
	panic("bad")
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_switches1 89
// Flags: FuncPropUnconditionalPanicExit
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":1,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_switches1(x int) {
	switch x {
	case 1:
		panic("one")
	case 2:
		panic("two")
	}
	panic("whatev")
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_switches1a 109
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_switches1a(x int) {
	switch x {
	case 2:
		panic("two")
	}
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_switches2 126
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_switches2(x int) {
	switch x {
	case 1:
		panic("one")
	case 2:
		panic("two")
	default:
		return
	}
	panic("whatev")
}

// funcflags.go T_switches3 143
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_switches3(x any) {
	switch x.(type) {
	case bool:
		panic("one")
	case float32:
		panic("two")
	}
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_switches4 163
// Flags: FuncPropUnconditionalPanicExit
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":1,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_switches4(x int) {
	switch x {
	case 1:
		panic("one")
		fallthrough
	case 2:
		panic("two")
		fallthrough
	default:
		panic("bad")
	}
	panic("whatev")
}

// funcflags.go T_recov 182
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_recov(x int) {
	if x := recover(); x != nil {
		panic(x)
	}
}

// funcflags.go T_forloops1 194
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_forloops1(x int) {
	for {
		panic("wokketa")
	}
}

// funcflags.go T_forloops2 205
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_forloops2(x int) {
	for {
		println("blah")
		if true {
			break
		}
		panic("warg")
	}
}

// funcflags.go T_forloops3 220
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_forloops3(x int) {
	for i := 0; i < 101; i++ {
		println("blah")
		if true {
			continue
		}
		panic("plark")
	}
	for i := range [10]int{} {
		println(i)
		panic("plark")
	}
	panic("whatev")
}

// funcflags.go T_hasgotos 240
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[]}
// =+=+=
// =-=-=
func T_hasgotos(x int, y int) {
	{
		xx := x
		panic("bad")
	lab1:
		goto lab2
	lab2:
		if false {
			goto lab1
		} else {
			goto lab4
		}
	lab4:
		if xx < y {
		lab3:
			if false {
				goto lab3
			}
		}
		println(9)
	}
}

//
//	0: ParamFeedsIfOrSwitch
//
// funcflags.go T_callsexit 276
// Flags: FuncPropUnconditionalPanicExit
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":1,"RecvrParamFlags":[8],"ReturnFlags":[]}
// callsite: funcflags.go:278:10|0 "CallSiteOnPanicPath" 2
// callsite: funcflags.go:280:9|1 "CallSiteOnPanicPath" 2
// =+=+=
// =-=-=
func T_callsexit(x int) {
	if x < 0 {
		os.Exit(1)
	}
	os.Exit(2)
}

func exprcallsexit(x int) int {
	os.Exit(x)
	return x
}

// funcflags.go T_exitinexpr 294
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// callsite: funcflags.go:299:18|0 "CallSiteOnPanicPath" 2
// =+=+=
// =-=-=
func T_exitinexpr(x int) {
	// This function does indeed unconditionally call exit, since the
	// first thing it does is invoke exprcallsexit, however from the
	// perspective of this function, the call is not at the statement
	// level, so we'll wind up missing it.
	if exprcallsexit(x) < 0 {
		println("foo")
	}
}

// funcflags.go T_calls_callsexit 311
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":[0],"ReturnFlags":[]}
// callsite: funcflags.go:312:15|0 "CallSiteOnPanicPath" 2
// =+=+=
// =-=-=
func T_calls_callsexit(x int) {
	exprcallsexit(x)
}
