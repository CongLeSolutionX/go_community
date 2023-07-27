// run -ldflags=-prunedeadmeth=0

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test conversion from function to single method interface.

package main

func Case1() {
	fn := func() {}
	type iface interface{ f() }
	fnc := func(i iface) {
		i.f()
	}
	fnc(iface(fn))
}

func Case2() {
	fn := func(arg0 int) {}
	type iface interface{ f(arg0 int) }
	fnc := func(i iface) {
		i.f(15)
	}
	fnc(iface(fn))
}

func Case3() {
	fn := func(arg0 int, arg1 rune) {}
	type iface interface{ f(arg0 int, arg1 rune) }
	fnc := func(i iface) {
		i.f(15, 'b')
	}
	fnc(iface(fn))
}

func Case4() {
	fn := func(arg0 int, arg1 rune) int { return 20 }
	type iface interface{ f(arg0 int, arg1 rune) int }
	fnc := func(i iface) {
		r := i.f(15, 'b')
		if r != 20 {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case5() {
	fn := func(arg0 int) (int, rune) { return arg0, 'd' }
	type iface interface{ f(arg0 int) (int, rune) }
	fnc := func(i iface) {
		r0, r1 := i.f(15)
		if r0 != 15 || r1 != 'd' {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case6() {
	fn := func(arg0 ...int) int { return arg0[0] }
	type iface interface{ f(arg0 ...int) int }
	fnc := func(i iface) {
		r := i.f(15)
		if r != 15 {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case7() {
	fn := func(arg0 int, arg1 rune, arg2 ...int) (int, rune) { return arg0, arg1 }
	type iface interface {
		f(arg0 int, arg1 rune, arg2 ...int) (int, rune)
	}
	fnc := func(i iface) {
		r0, r1 := i.f(15, 'c', 20)
		if r0 != 15 || r1 != 'c' {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case8() {
	fn := func() int { return 1 }
	type iface interface {
		f() int
	}
	fnc := func(i iface) {
		r0 := i.f()
		if r0 != 1 {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case9() {
	fn := func() (int, rune) { return 1, 'c' }
	type iface interface {
		f() (int, rune)
	}
	fnc := func(i iface) {
		r0, r1 := i.f()
		if r0 != 1 || r1 != 'c' {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case10() {
	fn := func() (int, rune, int) { return 1, 'c', 1 }
	type iface interface {
		f() (int, rune, int)
	}
	fnc := func(i iface) {
		r0, r1, r2 := i.f()
		if r0 != 1 || r1 != 'c' || r2 != 1 {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case11() {
	fn := func() (int, int, int) { return 1, 2, 3 }
	type iface interface {
		f() (int, int, int)
	}
	fnc := func(i iface) {
		r0, r1, r2 := i.f()
		if r0 != 1 || r1 != 2 || r2 != 3 {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

func Case12() {
	fn := func() (int, int, int, int) { return 1, 2, 3, 4 }
	type iface interface {
		f() (int, int, int, int)
	}
	fnc := func(i iface) {
		r0, r1, r2, r3 := i.f()
		if r0 != 1 || r1 != 2 || r2 != 3 || r3 != 4 {
			panic("fail")
		}
	}
	fnc(iface(fn))
}

// variable of a function type
func Case13() {
	var fn = func() {}
	type iface interface {
		f()
	}
	fnc := func(i iface) {
		i.f()
	}
	fnc(iface(fn))
}

// method expression
type S struct{}

func (s *S) bar(arg0 int) int { return 1 }

func Case14() {
	type sb struct{ F func() }
	type iface interface {
		f(arg0 int) int
	}
	is := &S{}
	fnc := func(i iface) {
		i.f(1)
	}
	fnc(iface(is.bar))
}

// function is a field
type SB struct{ F func() }

func Case15() {
	s := &SB{F: func() {}}
	type iface interface {
		f()
	}
	_ = iface(s.F)
}

type IT int

func (i IT) b(arg0 int) {}

func Case16() {
	var s IT = 5
	type iface interface {
		f(arg0 int)
	}
	_ = iface(s.b)
}

func Case17() {
	i.M()
}

var i = I(F(nil))

type I interface{ M() }

type F func()

func (F) M() {}

func main() {
	Case1()
	Case2()
	Case3()
	Case4()
	Case5()
	Case6()
	Case7()
	Case8()
	Case9()
	Case10()
	Case11()
	Case12()
	Case13()
	Case14()
	Case15()
	Case16()
	Case17()
}
