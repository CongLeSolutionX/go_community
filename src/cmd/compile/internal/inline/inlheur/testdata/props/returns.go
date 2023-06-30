// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT COMMENTS (use 'go test -v -update-expected' instead)

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

// T_simple_allocmem
// ReturnFlags:
//   0: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[2]}
// =-=-=
func T_simple_allocmem() *Bar {
	return &Bar{}
}

// T_allocmem_two_returns
// ReturnFlags:
//   0: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[2]}
// =-=-=
func T_allocmem_two_returns(x int) *Bar {
	// multiple returns
	if x < 0 {
		return new(Bar)
	} else {
		return &Bar{x: 2}
	}
}

// T_allocmem_three_returns
// ReturnFlags:
//   0: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[2]}
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

// T_return_nil
// ReturnFlags:
//   0: ReturnAlwaysSameConstant
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[8]}
// =-=-=
func T_return_nil() *Bar {
	// simple case: no alloc
	return nil
}

// T_multi_return_nil
// ReturnFlags:
//   0: ReturnAlwaysSameConstant
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[8]}
// =-=-=
func T_multi_return_nil(x, y bool) *Bar {
	if x && y {
		return nil
	}
	return nil
}

// T_multi_return_nil_anomoly
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[0]}
// =-=-=
func T_multi_return_nil_anomoly(x, y bool) Itf {
	if x && y {
		var qnil *Q
		return qnil
	}
	var barnil *Bar
	return barnil
}

// T_multi_return_some_nil
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[0]}
// =-=-=
func T_multi_return_some_nil(x, y bool) *Bar {
	if x && y {
		return nil
	} else {
		return &GB
	}
}

var GB Bar

// T_mixed_returns
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[0]}
// =-=-=
func T_mixed_returns(x int) *Bar {
	// mix of alloc and non-alloc
	if x < 0 {
		return new(Bar)
	} else {
		return &GB
	}
}

// T_mixed_returns_slice
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[0]}
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

// T_maps_and_channels
// ReturnFlags:
//   0: ReturnNoInfo
//   1: ReturnNoInfo
//   2: ReturnNoInfo
//   3: ReturnAlwaysSameConstant
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[0,0,0,8]}
// =-=-=
func T_maps_and_channels(x int, b bool) (bool, map[int]int, chan bool, unsafe.Pointer) {
	// maps and channels
	return b, make(map[int]int), make(chan bool), nil
}

// T_assignment_to_named_returns
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[0,0]}
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

// T_named_returns_but_return_explicit_values
// ReturnFlags:
//   0: ReturnIsAllocatedMem
//   1: ReturnIsAllocatedMem
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[2,2]}
// =-=-=
func T_named_returns_but_return_explicit_values(x int) (r1 *uint64, r2 *uint64) {
	// named returns ok if we don't use them by name
	rx1 := new(uint64)
	if x < 1 {
		*rx1 = 2
	}
	rx2 := new(uint64)
	return rx1, rx2
}

// T_return_concrete_type_to_itf
// ReturnFlags:
//   0: ReturnIsConcreteTypeConvertedToInterface
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[4]}
// =-=-=
func T_return_concrete_type_to_itf(x, y int) Itf {
	return &Bar{}
}

// T_return_concrete_type_to_itfwith_copy
// ReturnFlags:
//   0: ReturnIsConcreteTypeConvertedToInterface
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[4]}
// =-=-=
func T_return_concrete_type_to_itfwith_copy(x, y int) Itf {
	b := &Bar{}
	println("whee")
	return b
}

// T_return_concrete_type_to_itf_mixed
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":[0]}
// =-=-=
func T_return_concrete_type_to_itf_mixed(x, y int) Itf {
	if x < y {
		b := &Bar{}
		return b
	}
	return nil
}
