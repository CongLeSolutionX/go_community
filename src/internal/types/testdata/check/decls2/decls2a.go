// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// method declarations

package decls2

import (
	"time"
	"unsafe"
)

// T1 declared before its methods.
type T1 struct{
	f int
}

func (T1) m() {}
func (T1) m /* ERR already declared */ () {}
func (x *T1) f /* ERR field and method */ () {}

// Conflict between embedded field and method name,
// with the embedded field being a basic type.
type T1b struct {
	int
}

func (T1b) int /* ERR field and method */ () {}

type T1c struct {
	time.Time
}

func (T1c) Time /* ERR field and method */ () int { return 0 }

// Disabled for now: LookupFieldOrMethod will find Pointer even though
// it's double-declared (it would cost extra in the common case to verify
// this). But the MethodSet computation will not find it due to the name
// collision caused by the double-declaration, leading to an internal
// inconsistency while we are verifying one computation against the other.
// var _ = T1c{}.Pointer

// T2's method declared before the type.
func (*T2) f /* ERR field and method */ () {}

type T2 struct {
	f int
}

// Methods declared without a declared type.
func (undefined /* ERR undefined */) m() {}
func (x *undefined /* ERR undefined */) m() {}

func (pi /* ERR not a type */) m1() {}
func (x pi /* ERR not a type */) m2() {}
func (x *pi /* ERR not a type */ ) m3() {}

// Blank types.
type _ struct { m int }
type _ struct { m int }

func (_ /* ERR cannot use _ */) m() {}
func m(_ /* ERR cannot use _ */) {}

// Methods with receiver base type declared in another file.
func (T3) m1() {}
func (*T3) m2() {}
func (x T3) m3() {}
func (x *T3) f /* ERR field and method */ () {}

// Methods of non-struct type.
type T4 func()

func (self T4) m() func() { return self }

// Methods associated with an interface.
type T5 interface {
	m() int
}

func (T5 /* ERR invalid receiver */ ) m1() {}
func (T5 /* ERR invalid receiver */ ) m2() {}

// Methods associated with a named pointer type.
type ptr *int
func (ptr /* ERR invalid receiver */ ) _() {}
func (* /* ERR invalid receiver */ ptr) _() {}

// Methods with zero or multiple receivers.
func ( /* ERR method has no receiver */ ) _() {}
func (T3, * /* ERR method has multiple receivers */ T3) _() {}
func (T3, T3, T3 /* ERR method has multiple receivers */ ) _() {}
func (a, b /* ERR method has multiple receivers */ T3) _() {}
func (a, b, c /* ERR method has multiple receivers */ T3) _() {}

// Methods associated with non-local or unnamed types.
func (int /* ERR cannot define new methods on non-local type int */ ) m() {}
func ([ /* ERR invalid receiver */ ]int) m() {}
func (time /* ERR cannot define new methods on non-local type time.Time */ .Time) m() {}
func (* /* ERR cannot define new methods on non-local type time.Time */ time.Time) m() {}
func (x /* ERR invalid receiver */ interface{}) m() {}

// Unsafe.Pointer is treated like a pointer when used as receiver type.
type UP unsafe.Pointer
func (UP /* ERR invalid */ ) m1() {}
func (* /* ERR invalid */ UP) m2() {}

// Double declarations across package files
const c_double = 0
type t_double int
var v_double int
func f_double() {}
