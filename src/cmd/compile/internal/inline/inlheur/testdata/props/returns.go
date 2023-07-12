// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT (use 'go test -v -update-expected' instead.)
// See cmd/compile/internal/inline/inlheur/testdata/props/README.txt
// for more information on the format of this file.
// =^=^=

package returns1

import "unsafe"

type Bar struct {
	x int
	y string
}

func (b *Bar) Plark() {
}

type Q int

func (q *Q) Plark() {
}

type Itf interface {
	Plark()
}

//
//	0: ReturnIsAllocatedMem
//
// returns.go T_simple_allocmem 41
// ReturnFlags:
//   0: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":[],"ReturnFlags":[2]}
// =+=+=
// =-=-=
func T_simple_allocmem() *Bar {
	return &Bar{}
}

//
//	0: ParamFeedsIfOrSwitch
//
//
//	0: ReturnIsAllocatedMem
//
// returns.go T_allocmem_two_returns 60
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// ReturnFlags:
//   0: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[2]}
// =+=+=
// =-=-=
func T_allocmem_two_returns(x int) *Bar {
	// multiple returns
	if x < 0 {
		return new(Bar)
	} else {
		return &Bar{x: 2}
	}
}

//
//	0: ParamFeedsIfOrSwitch
//
//
//	0: ReturnIsAllocatedMem
//
// returns.go T_allocmem_three_returns 84
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// ReturnFlags:
//   0: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[2]}
// =+=+=
// =-=-=
func T_allocmem_three_returns(x int) []*Bar {
	// more multiple returns
	switch x {
	case 10, 11, 12:
		return make([]*Bar, 10)
	case 13:
		fallthrough
	case 15:
		return []*Bar{&Bar{x: 15}}
	}
	return make([]*Bar, 0, 10)
}

//
//	0: ReturnAlwaysSameConstant
//
// returns.go T_return_nil 107
// ReturnFlags:
//   0: ReturnAlwaysSameConstant
// =====
// {"Flags":0,"RecvrParamFlags":[],"ReturnFlags":[8]}
// =+=+=
// =-=-=
func T_return_nil() *Bar {
	// simple case: no alloc
	return nil
}

//
//	0: ReturnAlwaysSameConstant
//
// returns.go T_multi_return_nil 122
// ReturnFlags:
//   0: ReturnAlwaysSameConstant
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[8]}
// =+=+=
// =-=-=
func T_multi_return_nil(x, y bool) *Bar {
	if x && y {
		return nil
	}
	return nil
}

// returns.go T_multi_return_nil_anomoly 134
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[0]}
// =+=+=
// =-=-=
func T_multi_return_nil_anomoly(x, y bool) Itf {
	if x && y {
		var qnil *Q
		return qnil
	}
	var barnil *Bar
	return barnil
}

// returns.go T_multi_return_some_nil 148
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[0]}
// =+=+=
// =-=-=
func T_multi_return_some_nil(x, y bool) *Bar {
	if x && y {
		return nil
	} else {
		return &GB
	}
}

var GB Bar

//
//	0: ParamFeedsIfOrSwitch
//
// returns.go T_mixed_returns 168
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[0]}
// =+=+=
// =-=-=
func T_mixed_returns(x int) *Bar {
	// mix of alloc and non-alloc
	if x < 0 {
		return new(Bar)
	} else {
		return &GB
	}
}

//
//	0: ParamFeedsIfOrSwitch
//
// returns.go T_mixed_returns_slice 187
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[0]}
// =+=+=
// =-=-=
func T_mixed_returns_slice(x int) []*Bar {
	// mix of alloc and non-alloc
	switch x {
	case 10, 11, 12:
		return make([]*Bar, 10)
	case 13:
		fallthrough
	case 15:
		return []*Bar{&Bar{x: 15}}
	}
	ba := [...]*Bar{&GB, &GB}
	return ba[:]
}

//
//	0: ReturnNoInfo
//	1: ReturnNoInfo
//	2: ReturnNoInfo
//	3: ReturnAlwaysSameConstant
//
// returns.go T_maps_and_channels 217
// ReturnFlags:
//   0: ReturnNoInfo
//   1: ReturnNoInfo
//   2: ReturnNoInfo
//   3: ReturnAlwaysSameConstant
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[0,0,0,8]}
// =+=+=
// =-=-=
func T_maps_and_channels(x int, b bool) (bool, map[int]int, chan bool, unsafe.Pointer) {
	// maps and channels
	return b, make(map[int]int), make(chan bool), nil
}

//
//	0: ParamFeedsIfOrSwitch
//
// returns.go T_assignment_to_named_returns 232
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[0,0]}
// =+=+=
// =-=-=
func T_assignment_to_named_returns(x int) (r1 *uint64, r2 *uint64) {
	// assignments to named returns and then "return" not supported
	r1 = new(uint64)
	if x < 1 {
		*r1 = 2
	}
	r2 = new(uint64)
	return
}

//
//	0: ParamFeedsIfOrSwitch
//
//
//	0: ReturnIsAllocatedMem
//	1: ReturnIsAllocatedMem
//
// returns.go T_named_returns_but_return_explicit_values 259
// RecvrParamFlags:
//   0: ParamFeedsIfOrSwitch
// ReturnFlags:
//   0: ReturnIsAllocatedMem
//   1: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":[8],"ReturnFlags":[2,2]}
// =+=+=
// =-=-=
func T_named_returns_but_return_explicit_values(x int) (r1 *uint64, r2 *uint64) {
	// named returns ok if all returns are non-empty
	rx1 := new(uint64)
	if x < 1 {
		*rx1 = 2
	}
	rx2 := new(uint64)
	return rx1, rx2
}

//
//	0: ReturnIsConcreteTypeConvertedToInterface
//
// returns.go T_return_concrete_type_to_itf 279
// ReturnFlags:
//   0: ReturnIsConcreteTypeConvertedToInterface
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[4]}
// =+=+=
// =-=-=
func T_return_concrete_type_to_itf(x, y int) Itf {
	return &Bar{}
}

//
//	0: ReturnIsConcreteTypeConvertedToInterface
//
// returns.go T_return_concrete_type_to_itfwith_copy 293
// ReturnFlags:
//   0: ReturnIsConcreteTypeConvertedToInterface
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[4]}
// =+=+=
// =-=-=
func T_return_concrete_type_to_itfwith_copy(x, y int) Itf {
	b := &Bar{}
	println("whee")
	return b
}

// returns.go T_return_concrete_type_to_itf_mixed 304
// =====
// {"Flags":0,"RecvrParamFlags":[0,0],"ReturnFlags":[0]}
// =+=+=
// =-=-=
func T_return_concrete_type_to_itf_mixed(x, y int) Itf {
	if x < y {
		b := &Bar{}
		return b
	}
	return nil
}
