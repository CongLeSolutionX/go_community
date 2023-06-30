// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT COMMENTS (use 'go test -v -update-expected' instead)
package params

import "os"

var G int

type T int

type I interface {
	Blarg()
}

func (r T) Blarg() {
}

// T_feeds_if_simple
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_simple(x int) {
	if x < 100 {
		os.Exit(1)
	}
	println(x)
}

// T_feeds_if_pointer
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_pointer(xp *int) {
	if xp != nil {
		os.Exit(1)
	}
	println(xp)
}

// T.T_feeds_if_simple_method
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
//   1: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8,8],"ReturnFlags":[]}
// =-=-=
func (r T) T_feeds_if_simple_method(x int) {
	if x < 100 {
		os.Exit(1)
	}
	if r != 99 {
		os.Exit(2)
	}
	println(x)
}

// T_feeds_if_blanks
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_blanks(_ string, x int, _ bool, _ bool) {
	// blanks ignored; from a props perspective "x" is param 0
	if x < 100 {
		os.Exit(1)
	}
	println(x)
}

// T_feeds_if_with_copy
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_with_copy(x int) {
	// simple copy here -- we get this case
	xx := x
	if xx < 100 {
		os.Exit(1)
	}
	println(x)
}

// T_feeds_if_with_copy_expr
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_with_copy_expr(x int) {
	// this case (copy of expression) currently not handled.
	xx := x < 100
	if xx {
		os.Exit(1)
	}
	println(x)
}

// T_feeds_switch
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =-=-=
func T_feeds_switch(x int) {
	switch x {
	case 101:
		println(101)
	case 202:
		panic("bad")
	}
	println(x)
}

// T_feeds_if_toocomplex
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_toocomplex(x int, y int) {
	// not handled at the moment; we only look for cases where
	// an "if" or "switch" can be simplified based on a single
	// constant param, not a combination of constant params.
	if x < y {
		panic("bad")
	}
	println(x + y)
}

// T_feeds_if_redefined
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_redefined(x int) {
	if x < G {
		x++
	}
	if x == 101 {
		panic("bad")
	}
}

// T_feeds_if_redefined2
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_redefined2(x int) {
	// this currently classifies "x" as "no info", since the analysis we
	// use to check for reassignments/redefinitions is not flow-sensitive,
	// but we could probably catch this case with better analysis or
	// high-level SSA.
	if x == 101 {
		panic("bad")
	}
	if x < G {
		x++
	}
}

// T_feeds_multi_if
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
//   1: ParamNoInfo
// =====
// {"Flags":0,"RecvrParamFlags":[8,0],"ReturnFlags":[]}
// =-=-=
func T_feeds_multi_if(x int, y int) {
	// Here we have one "if" that is too complex (x < y) but one that is
	// simple enough. Currently we enable the heuristic for this. It's
	// possible to imagine this being a bad thing if the function in
	// question is sufficiently large, but if it's too large we probably
	// can't inline it anyhow.
	if x < y {
		panic("bad")
	}
	if x < 10 {
		panic("whatev")
	}
	println(x + y)
}

// T_feeds_if_redefined_indirectwrite
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_redefined_indirectwrite(x int) {
	ax := &x
	if G != 2 {
		*ax = G
	}
	if x == 101 {
		panic("bad")
	}
}

// T_feeds_if_redefined_indirectwrite_copy
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_redefined_indirectwrite_copy(x int) {
	// we don't catch this case, "x" is marked as no info,
	// since we're conservative about redefinitions.
	ax := &x
	cx := x
	if G != 2 {
		*ax = G
	}
	if cx == 101 {
		panic("bad")
	}
}

// T_feeds_if_expr1
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_expr1(x int) {
	if x == 101 || x == 102 || x&0xf == 0 {
		panic("bad")
	}
}

// T_feeds_if_expr2
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_expr2(x int) {
	if (x*x)-(x+x)%x == 101 || x&0xf == 0 {
		panic("bad")
	}
}

// T_feeds_if_expr3
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_expr3(x int) {
	if x-(x&0x1)^378 > (1 - G) {
		panic("bad")
	}
}

// T_feeds_if_shift_may_panic
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[0]}
// =-=-=
func T_feeds_if_shift_may_panic(x int) *int {
	// here if "x" is a constant like 2, we could simplify the "if",
	// but if we were to pass in a negative value for "x" we can't
	// fold the condition due to the need to panic on negative shift.
	if 1<<x > 1024 {
		return nil
	}
	return &G
}

// T_feeds_if_maybe_divide_by_zero
// =====
// {"Flags":0,"RecvrParamFlags":[0],"ReturnFlags":[]}
// =-=-=
func T_feeds_if_maybe_divide_by_zero(x int) {
	if 99/x == 3 {
		return
	}
	println("blarg")
}

// T_feeds_indcall
// RecvrParamFlags:
//   0: ParamFeedsIndirectCall
// =====
// {"Flags":0,"RecvrParamFlags":[4],"ReturnFlags":[]}
// =-=-=
func T_feeds_indcall(x func()) {
	if G != 20 {
		x()
	}
}

// T_feeds_indcall_and_if
// RecvrParamFlags:
//   0: ParamFeedsIndirectCall|ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[12],"ReturnFlags":[]}
// =-=-=
func T_feeds_indcall_and_if(x func()) {
	if x != nil {
		x()
	}
}

// T_feeds_indcall_with_copy
// RecvrParamFlags:
//   0: ParamFeedsIndirectCall
// =====
// {"Flags":0,"RecvrParamFlags":[4],"ReturnFlags":[]}
// =-=-=
func T_feeds_indcall_with_copy(x func()) {
	xx := x
	if G < 10 {
		G--
	}
	xx()
}

// T_feeds_interface_method_call
// RecvrParamFlags:
//   0: ParamFeedsInterfaceMethodCall
// =====
// {"Flags":0,"RecvrParamFlags":[2],"ReturnFlags":[]}
// =-=-=
func T_feeds_interface_method_call(i I) {
	i.Blarg()
}
