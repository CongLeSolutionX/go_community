// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"reflect"
	"testing"
)

// Signal size changes of important structures.

func TestSizeof(t *testing.T) {
	const _64bit = ^uint(0)>>32 != 0

	var tests = []struct {
		val    interface{} // type as a value
		_32bit uintptr     // size on 32bit platforms
		_64bit uintptr     // size on 64bit platforms
	}{
		// TODO(gri) fill in the 32-bit sizes
		// Types
		{Basic{}, 0, 32},
		{Array{}, 0, 24},
		{Slice{}, 0, 16},
		{Struct{}, 0, 48},
		{Pointer{}, 0, 16},
		{Tuple{}, 0, 24},
		{Signature{}, 0, 88},
		{Sum{}, 0, 24},
		{Interface{}, 0, 120},
		{Map{}, 0, 32},
		{Chan{}, 0, 24},
		{Named{}, 0, 136},
		{TypeParam{}, 0, 48},
		{instance{}, 0, 96},
		{bottom{}, 0, 0},
		{top{}, 0, 0},

		// Objects
		{PkgName{}, 0, 104},
		{Const{}, 0, 104},
		{TypeName{}, 0, 88},
		{Var{}, 0, 96},
		{Func{}, 0, 96},
		{Label{}, 0, 96},
		{Builtin{}, 0, 96},
		{Nil{}, 0, 88},

		// Misc
		{Scope{}, 0, 96},
		{Package{}, 0, 80},
	}

	for _, test := range tests {
		got := reflect.TypeOf(test.val).Size()
		want := test._32bit
		if _64bit {
			want = test._64bit
		}
		if got != want {
			t.Errorf("unsafe.Sizeof(%T) = %d, want %d", test.val, got, want)
		}
	}
}
