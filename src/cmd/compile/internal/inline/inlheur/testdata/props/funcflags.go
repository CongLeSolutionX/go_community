// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT COMMENTS (use 'go test -v -update-expected' instead)

package funcflags

import "os"

// T_simple
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_simple() {
	panic("bad")
}

// T_nested
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_nested(x int) {
	if x < 10 {
		panic("bad")
	} else {
		panic("good")
	}
}

// T_block1
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_block1(x int) {
	panic("bad")
	if x < 10 {
		return
	}
}

// T_block2
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_block2(x int) {
	if x < 10 {
		return
	}
	panic("bad")
}

// T_switches1
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":null,"ReturnFlags":null}
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

// T_switches1a
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_switches1a(x int) {
	switch x {
	case 2:
		panic("two")
	}
}

// T_switches2
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
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

// T_switches3
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_switches3(x any) {
	switch x.(type) {
	case bool:
		panic("one")
	case float32:
		panic("two")
	}
}

// T_switches4
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":null,"ReturnFlags":null}
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

// T_recov
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_recov(x int) {
	if x := recover(); x != nil {
		panic(x)
	}
}

// T_forloops1
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_forloops1(x int) {
	for {
		panic("wokketa")
	}
}

// T_forloops2
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
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

// T_forloops3
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
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

// T_hasgotos
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
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

// T_callsexit
// Flags: FuncPropUnconditionalPanicExit
// =====
// {"Flags":1,"RecvrParamFlags":null,"ReturnFlags":null}
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

// T_exitinexpr
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
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
