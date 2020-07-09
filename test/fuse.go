// +build amd64 arm64 linux
// errorcheck -0 -d=ssa/late_fuse/debug=1

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

const Cf2 = 2.0

// TypEq
func fEqEq(a int, f float64) bool {
	return a == 0 && f > Cf2 || a == 0 && f < -Cf2 // ERROR "Redirect Eq64 based on Eq64$"
}

func fEqNeq(a int32, f float64) bool {
	return a == 0 && f > Cf2 || a != 0 && f < -Cf2 // ERROR "Redirect Neq32 based on Eq32$"
}

func fEqLess(a int8, f float64) bool {
	return a == 0 && f > Cf2 || a < 0 && f < -Cf2
}

func fEqLeq(a float64, f float64) bool {
	return a == 0 && f > Cf2 || a <= 0 && f < -Cf2
}

func fEqLessU(a uint, f float64) bool {
	return a == 0 && f > Cf2 || a < 0 && f < -Cf2
}

func fEqLeqU(a uint64, f float64) bool {
	return a == 0 && f > Cf2 || a <= 0 && f < -Cf2
}

// TypNeq
func fNeqEq(a int, f float64) bool {
	return a != 0 && f > Cf2 || a == 0 && f < -Cf2 // ERROR "Redirect Eq64 based on Neq64$"
}

func fNeqNeq(a int32, f float64) bool {
	return a != 0 && f > Cf2 || a != 0 && f < -Cf2 // ERROR "Redirect Neq32 based on Neq32$"
}

func fNeqLess(a float32, f float64) bool {
	return a != 0 && f > Cf2 || a < 0 && f < -Cf2 // ERROR "Redirect Less32F based on Neq32F$"
}

func fNeqLeq(a int16, f float64) bool {
	return a != 0 && f > Cf2 || a <= 0 && f < -Cf2 // ERROR "Redirect Leq16 based on Neq16$"
}

func fNeqLessU(a uint, f float64) bool {
	return a != 0 && f > Cf2 || a < 0 && f < -Cf2 // ERROR "Redirect Less64U based on Neq64$"
}

func fNeqLeqU(a uint32, f float64) bool {
	return a != 0 && f > Cf2 || a <= 0 && f < -Cf2 // ERROR "Redirect Leq32U based on Neq32$"
}

// TypLess
func fLessEq(a int, f float64) bool {
	return a < 0 && f > Cf2 || a == 0 && f < -Cf2
}

func fLessNeq(a int32, f float64) bool {
	return a < 0 && f > Cf2 || a != 0 && f < -Cf2
}

func fLessLess(a float32, f float64) bool {
	return a < 0 && f > Cf2 || a < 0 && f < -Cf2 // ERROR "Redirect Less32F based on Less32F$"
}

func fLessLeq(a float64, f float64) bool {
	return a < 0 && f > Cf2 || a <= 0 && f < -Cf2
}

// TypLeq
func fLeqEq(a float64, f float64) bool {
	return a <= 0 && f > Cf2 || a == 0 && f < -Cf2 // ERROR "Redirect Eq64F based on Leq64F$"
}

func fLeqNeq(a int16, f float64) bool {
	return a <= 0 && f > Cf2 || a != 0 && f < -Cf2 // ERROR "Redirect Neq16 based on Leq16$"
}

func fLeqLess(a float32, f float64) bool {
	return a <= 0 && f > Cf2 || a < 0 && f < -Cf2 // ERROR "Redirect Less32F based on Leq32F$"
}

func fLeqLeq(a int8, f float64) bool {
	return a <= 0 && f > Cf2 || a <= 0 && f < -Cf2 // ERROR "Redirect Leq8 based on Leq8$"
}

// TypLessU
func fLessUEq(a uint8, f float64) bool {
	return a < 0 && f > Cf2 || a == 0 && f < -Cf2
}

func fLessUNeq(a uint16, f float64) bool {
	return a < 0 && f > Cf2 || a != 0 && f < -Cf2
}

func fLessULessU(a uint32, f float64) bool {
	return a < 0 && f > Cf2 || a < 0 && f < -Cf2 // ERROR "Redirect Less32U based on Less32U$"
}

func fLessULeqU(a uint64, f float64) bool {
	return a < 0 && f > Cf2 || a <= 0 && f < -Cf2
}

// TypLeqU
func fLeqUEq(a uint8, f float64) bool {
	return a <= 0 && f > Cf2 || a == 0 && f < -Cf2 // ERROR "Redirect Eq8 based on Leq8U$"
}

func fLeqUNeq(a uint16, f float64) bool {
	return a <= 0 && f > Cf2 || a != 0 && f < -Cf2 // ERROR "Redirect Neq16 based on Leq16U$"
}

func fLeqLessU(a uint32, f float64) bool {
	return a <= 0 && f > Cf2 || a < 0 && f < -Cf2 // ERROR "Redirect Less32U based on Leq32U$"
}

func fLeqLeqU(a uint64, f float64) bool {
	return a <= 0 && f > Cf2 || a <= 0 && f < -Cf2 // ERROR "Redirect Leq64U based on Leq64U$"
}

// Cases for TypEqB and TypNeqB are hard to construct,
// the following cases actually match OpArg.
// TypEqB
func fEqBEqB(a bool, f float64) bool { // ERROR "Redirect Arg based on Arg$"
	return a && f > Cf2 || a && f < -Cf2
}

func fEqBNeqB(a bool, f float64) bool { // ERROR "Redirect Arg based on Arg$"
	return a && f > Cf2 || !a && f < -Cf2
}

// TypNeqB
func fNeqBEqB(a bool, f float64) bool { // ERROR "Redirect Arg based on Arg$"
	return !a && f > Cf2 || a && f < -Cf2
}

func fNeqBNeqB(a bool, f float64) bool { // ERROR "Redirect Arg based on Arg$"
	return !a && f > Cf2 || !a && f < -Cf2
}

// TypEqPtr
func fEqPtrEqPtr(a, b *int, f float64) bool {
	return a == b && f > Cf2 || a == b && f < -Cf2 // ERROR "Redirect EqPtr based on EqPtr$"
}

func fEqPtrNeqPtr(a, b *int, f float64) bool {
	return a == b && f > Cf2 || a != b && f < -Cf2 // ERROR "Redirect NeqPtr based on EqPtr$"
}

// TypNeqPtr
func fNeqPtrEqPtr(a, b *int, f float64) bool {
	return a != b && f > Cf2 || a == b && f < -Cf2 // ERROR "Redirect EqPtr based on NeqPtr$"
}

func fNeqPtrNeqPtr(a, b *int, f float64) bool {
	return a != b && f > Cf2 || a != b && f < -Cf2 // ERROR "Redirect NeqPtr based on NeqPtr$"
}

// TypEqInter, TypNeqInter, TypEqSlice and TypNeqSlice are
// converted to TypIsNonNil, so they match TypIsNonNil.
// TypEqInter
func fEqInterEqInter(a interface{}, f float64) bool {
	return a == nil && f > Cf2 || a == nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

func fEqInterNeqInter(a interface{}, f float64) bool {
	return a == nil && f > Cf2 || a != nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

// TypNeqInter
func fNeqInterEqInter(a interface{}, f float64) bool {
	return a != nil && f > Cf2 || a == nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

func fNeqInterNeqInter(a interface{}, f float64) bool {
	return a != nil && f > Cf2 || a != nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

// TypEqSlice
func fEqSliceEqSlice(a []int, f float64) bool {
	return a == nil && f > Cf2 || a == nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

func fEqSliceNeqSlice(a []int, f float64) bool {
	return a == nil && f > Cf2 || a != nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

// TypNeqSlice
func fNeqSliceEqSlice(a []int, f float64) bool {
	return a != nil && f > Cf2 || a == nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

func fNeqSliceNeqSlice(a []int, f float64) bool {
	return a != nil && f > Cf2 || a != nil && f < -Cf2 // ERROR "Redirect IsNonNil based on IsNonNil$"
}

// TypPhi
func fPhi(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	un := len(s)
	if !neg && un > 10 || neg && un < 5 { // ERROR "Redirect Phi based on Phi$"
		return 0, false
	}
	return un, true
}

// TypArg
func fArg(neg bool, a int) (int, bool) { // ERROR "Redirect Arg based on Arg$"
	if !neg && a > 10 || neg && a < 5 {
		return a, false
	}
	return a, true
}

func main() {
}
