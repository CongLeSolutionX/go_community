// run

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test the 'for range' construct.

package main

// test range over channels

func gen(c chan int, lo, hi int) {
	for i := lo; i <= hi; i++ {
		c <- i
	}
	close(c)
}

func seq(lo, hi int) chan int {
	c := make(chan int)
	go gen(c, lo, hi)
	return c
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func testblankvars() {
	n := 0
	for range alphabet {
		n++
	}
	if n != 26 {
		println("for range: wrong count", n, "want 26")
		panic("fail")
	}
	n = 0
	for _ = range alphabet {
		n++
	}
	if n != 26 {
		println("for _ = range: wrong count", n, "want 26")
		panic("fail")
	}
	n = 0
	for _, _ = range alphabet {
		n++
	}
	if n != 26 {
		println("for _, _ = range: wrong count", n, "want 26")
		panic("fail")
	}
	s := 0
	for i, _ := range alphabet {
		s += i
	}
	if s != 325 {
		println("for i, _ := range: wrong sum", s, "want 325")
		panic("fail")
	}
	r := rune(0)
	for _, v := range alphabet {
		r += v
	}
	if r != 2847 {
		println("for _, v := range: wrong sum", r, "want 2847")
		panic("fail")
	}
}

func testchan() {
	s := ""
	for i := range seq('a', 'z') {
		s += string(i)
	}
	if s != alphabet {
		println("Wanted lowercase alphabet; got", s)
		panic("fail")
	}
	n := 0
	for range seq('a', 'z') {
		n++
	}
	if n != 26 {
		println("testchan wrong count", n, "want 26")
		panic("fail")
	}
}

// test that range over slice only evaluates
// the expression after "range" once.

var nmake = 0

func makeslice() []int {
	nmake++
	return []int{1, 2, 3, 4, 5}
}

func testslice() {
	s := 0
	nmake = 0
	for _, v := range makeslice() {
		s += v
	}
	if nmake != 1 {
		println("range called makeslice", nmake, "times")
		panic("fail")
	}
	if s != 15 {
		println("wrong sum ranging over makeslice", s)
		panic("fail")
	}

	x := []int{10, 20}
	y := []int{99}
	i := 1
	for i, x[i] = range y {
		break
	}
	if i != 0 || x[0] != 10 || x[1] != 99 {
		println("wrong parallel assignment", i, x[0], x[1])
		panic("fail")
	}
}

func testslice1() {
	s := 0
	nmake = 0
	for i := range makeslice() {
		s += i
	}
	if nmake != 1 {
		println("range called makeslice", nmake, "times")
		panic("fail")
	}
	if s != 10 {
		println("wrong sum ranging over makeslice", s)
		panic("fail")
	}
}

func testslice2() {
	n := 0
	nmake = 0
	for range makeslice() {
		n++
	}
	if nmake != 1 {
		println("range called makeslice", nmake, "times")
		panic("fail")
	}
	if n != 5 {
		println("wrong count ranging over makeslice", n)
		panic("fail")
	}
}

// test that range over []byte(string) only evaluates
// the expression after "range" once.

func makenumstring() string {
	nmake++
	return "\x01\x02\x03\x04\x05"
}

func testslice3() {
	s := byte(0)
	nmake = 0
	for _, v := range []byte(makenumstring()) {
		s += v
	}
	if nmake != 1 {
		println("range called makenumstring", nmake, "times")
		panic("fail")
	}
	if s != 15 {
		println("wrong sum ranging over []byte(makenumstring)", s)
		panic("fail")
	}
}

// test that range over array only evaluates
// the expression after "range" once.

func makearray() [5]int {
	nmake++
	return [5]int{1, 2, 3, 4, 5}
}

func testarray() {
	s := 0
	nmake = 0
	for _, v := range makearray() {
		s += v
	}
	if nmake != 1 {
		println("range called makearray", nmake, "times")
		panic("fail")
	}
	if s != 15 {
		println("wrong sum ranging over makearray", s)
		panic("fail")
	}
}

func testarray1() {
	s := 0
	nmake = 0
	for i := range makearray() {
		s += i
	}
	if nmake != 1 {
		println("range called makearray", nmake, "times")
		panic("fail")
	}
	if s != 10 {
		println("wrong sum ranging over makearray", s)
		panic("fail")
	}
}

func testarray2() {
	n := 0
	nmake = 0
	for range makearray() {
		n++
	}
	if nmake != 1 {
		println("range called makearray", nmake, "times")
		panic("fail")
	}
	if n != 5 {
		println("wrong count ranging over makearray", n)
		panic("fail")
	}
}

func makearrayptr() *[5]int {
	nmake++
	return &[5]int{1, 2, 3, 4, 5}
}

func testarrayptr() {
	nmake = 0
	x := len(makearrayptr())
	if x != 5 || nmake != 1 {
		println("len called makearrayptr", nmake, "times and got len", x)
		panic("fail")
	}
	nmake = 0
	x = cap(makearrayptr())
	if x != 5 || nmake != 1 {
		println("cap called makearrayptr", nmake, "times and got len", x)
		panic("fail")
	}
	s := 0
	nmake = 0
	for _, v := range makearrayptr() {
		s += v
	}
	if nmake != 1 {
		println("range called makearrayptr", nmake, "times")
		panic("fail")
	}
	if s != 15 {
		println("wrong sum ranging over makearrayptr", s)
		panic("fail")
	}
}

func testarrayptr1() {
	s := 0
	nmake = 0
	for i := range makearrayptr() {
		s += i
	}
	if nmake != 1 {
		println("range called makearrayptr", nmake, "times")
		panic("fail")
	}
	if s != 10 {
		println("wrong sum ranging over makearrayptr", s)
		panic("fail")
	}
}

func testarrayptr2() {
	n := 0
	nmake = 0
	for range makearrayptr() {
		n++
	}
	if nmake != 1 {
		println("range called makearrayptr", nmake, "times")
		panic("fail")
	}
	if n != 5 {
		println("wrong count ranging over makearrayptr", n)
		panic("fail")
	}
}

// test that range over string only evaluates
// the expression after "range" once.

func makestring() string {
	nmake++
	return "abcd☺"
}

func teststring() {
	var s rune
	nmake = 0
	for _, v := range makestring() {
		s += v
	}
	if nmake != 1 {
		println("range called makestring", nmake, "times")
		panic("fail")
	}
	if s != 'a'+'b'+'c'+'d'+'☺' {
		println("wrong sum ranging over makestring", s)
		panic("fail")
	}

	x := []rune{'a', 'b'}
	i := 1
	for i, x[i] = range "c" {
		break
	}
	if i != 0 || x[0] != 'a' || x[1] != 'c' {
		println("wrong parallel assignment", i, x[0], x[1])
		panic("fail")
	}

	y := []int{1, 2, 3}
	r := rune(1)
	for y[r], r = range "\x02" {
		break
	}
	if r != 2 || y[0] != 1 || y[1] != 0 || y[2] != 3 {
		println("wrong parallel assignment", r, y[0], y[1], y[2])
		panic("fail")
	}
}

func teststring1() {
	s := 0
	nmake = 0
	for i := range makestring() {
		s += i
	}
	if nmake != 1 {
		println("range called makestring", nmake, "times")
		panic("fail")
	}
	if s != 10 {
		println("wrong sum ranging over makestring", s)
		panic("fail")
	}
}

func teststring2() {
	n := 0
	nmake = 0
	for range makestring() {
		n++
	}
	if nmake != 1 {
		println("range called makestring", nmake, "times")
		panic("fail")
	}
	if n != 5 {
		println("wrong count ranging over makestring", n)
		panic("fail")
	}
}

// test that range over map only evaluates
// the expression after "range" once.

func makemap() map[int]int {
	nmake++
	return map[int]int{0: 'a', 1: 'b', 2: 'c', 3: 'd', 4: '☺'}
}

func testmap() {
	s := 0
	nmake = 0
	for _, v := range makemap() {
		s += v
	}
	if nmake != 1 {
		println("range called makemap", nmake, "times")
		panic("fail")
	}
	if s != 'a'+'b'+'c'+'d'+'☺' {
		println("wrong sum ranging over makemap", s)
		panic("fail")
	}
}

func testmap1() {
	s := 0
	nmake = 0
	for i := range makemap() {
		s += i
	}
	if nmake != 1 {
		println("range called makemap", nmake, "times")
		panic("fail")
	}
	if s != 10 {
		println("wrong sum ranging over makemap", s)
		panic("fail")
	}
}

func testmap2() {
	n := 0
	nmake = 0
	for range makemap() {
		n++
	}
	if nmake != 1 {
		println("range called makemap", nmake, "times")
		panic("fail")
	}
	if n != 5 {
		println("wrong count ranging over makemap", n)
		panic("fail")
	}
}

func testint1() {
	bad := false
	j := 0
	for i := range int(4) {
		if i != j {
			println("range var", i, "want", j)
			bad = true
		}
		j++
	}
	if j != 4 {
		println("wrong count ranging over 4:", j)
		bad = true
	}
	if bad {
		panic("testint1")
	}
}

func testint2() {
	bad := false
	j := 0
	for i := range 4 {
		if i != j {
			println("range var", i, "want", j)
			bad = true
		}
		j++
	}
	if j != 4 {
		println("wrong count ranging over 4:", j)
		bad = true
	}
	if bad {
		panic("testint2")
	}
}

func testint3() {
	bad := false
	type MyInt int
	j := MyInt(0)
	for i := range MyInt(4) {
		if i != j {
			println("range var", i, "want", j)
			bad = true
		}
		j++
	}
	if j != 4 {
		println("wrong count ranging over 4:", j)
		bad = true
	}
	if bad {
		panic("testint3")
	}
}

var gj int

func yield4x(yield func() bool) bool {
	return yield() && yield() && yield() && yield()
}

func yield4(yield func(int) bool) bool {
	return yield(1) && yield(2) && yield(3) && yield(4)
}

func yield3(yield func(int) bool) bool {
	return yield(1) && yield(2) && yield(3)
}

func yield2(yield func(int) bool) bool {
	return yield(1) && yield(2)
}

func testfunc0() {
	j := 0
	for range yield4x {
		j++
	}
	if j != 4 {
		println("wrong count ranging over yield4x:", j)
		panic("testfunc0")
	}

	j = 0
	for range yield4 {
		j++
	}
	if j != 4 {
		println("wrong count ranging over yield4:", j)
		panic("testfunc0")
	}
}

func testfunc1() {
	bad := false
	j := 1
	for i := range yield4 {
		if i != j {
			println("range var", i, "want", j)
			bad = true
		}
		j++
	}
	if j != 5 {
		println("wrong count ranging over f:", j)
		bad = true
	}
	if bad {
		panic("testfunc1")
	}
}

func testfunc2() {
	bad := false
	j := 1
	var i int
	for i = range yield4 {
		if i != j {
			println("range var", i, "want", j)
			bad = true
		}
		j++
	}
	if j != 5 {
		println("wrong count ranging over f:", j)
		bad = true
	}
	if i != 4 {
		println("wrong final i ranging over f:", i)
		bad = true
	}
	if bad {
		panic("testfunc2")
	}
}

func testfunc3() {
	bad := false
	j := 1
	var i int
	for i = range yield4 {
		if i != j {
			println("range var", i, "want", j)
			bad = true
		}
		j++
		if i == 2 {
			break
		}
		continue
	}
	if j != 3 {
		println("wrong count ranging over f:", j)
		bad = true
	}
	if i != 2 {
		println("wrong final i ranging over f:", i)
		bad = true
	}
	if bad {
		panic("testfunc3")
	}
}

func testfunc4() {
	bad := false
	j := 1
	var i int
	func() {
		for i = range yield4 {
			if i != j {
				println("range var", i, "want", j)
				bad = true
			}
			j++
			if i == 2 {
				return
			}
		}
	}()
	if j != 3 {
		println("wrong count ranging over f:", j)
		bad = true
	}
	if i != 2 {
		println("wrong final i ranging over f:", i)
		bad = true
	}
	if bad {
		panic("testfunc3")
	}
}

func func5() (int, int) {
	for i := range yield4 {
		return 10, i
	}
	panic("still here")
}

func testfunc5() {
	x, y := func5()
	if x != 10 || y != 1 {
		println("wrong results", x, y, "want", 10, 1)
		panic("testfunc5")
	}
}

func func6() (z, w int) {
	for i := range yield4 {
		z = 10
		w = i
		return
	}
	panic("still here")
}

func testfunc6() {
	x, y := func6()
	if x != 10 || y != 1 {
		println("wrong results", x, y, "want", 10, 1)
		panic("testfunc6")
	}
}

var saved []int

func save(x int) {
	saved = append(saved, x)
}

func printslice(s []int) {
	print("[")
	for i, x := range s {
		if i > 0 {
			print(", ")
		}
		print(x)
	}
	print("]")
}

func eqslice(s, t []int) bool {
	if len(s) != len(t) {
		return false
	}
	for i, x := range s {
		if x != t[i] {
			return false
		}
	}
	return true
}

func func7() {
	defer save(-1)
	for i := range yield4 {
		defer save(i)
	}
	defer save(5)
}

func checkslice(name string, saved, want []int) {
	if !eqslice(saved, want) {
		print("wrong results ")
		printslice(saved)
		print(" want ")
		printslice(want)
		print("\n")
		panic(name)
	}
}

func testfunc7() {
	saved = nil
	func7()
	want := []int{5, 4, 3, 2, 1, -1}
	checkslice("testfunc7", saved, want)
}

func func8() {
	defer save(-1)
	for i := range yield2 {
		for j := range yield3 {
			defer save(i*10 + j)
		}
		defer save(i)
	}
	defer save(-2)
	for i := range yield4 {
		defer save(i)
	}
	defer save(-3)
}

func testfunc8() {
	saved = nil
	func8()
	want := []int{-3, 4, 3, 2, 1, -2, 2, 23, 22, 21, 1, 13, 12, 11, -1}
	checkslice("testfunc8", saved, want)
}

func func9() {
	n := 0
	for range yield2 {
		for range yield3 {
			n++
			defer save(n)
			return
		}
	}
}

func testfunc9() {
	saved = nil
	func9()
	want := []int{6, 5, 4, 3, 2, 1}
	checkslice("testfunc9", saved, want)
}

// test that range evaluates the index and value expressions
// exactly once per iteration.

var ncalls = 0

func getvar(p *int) *int {
	ncalls++
	return p
}

func testcalls() {
	var i, v int
	si := 0
	sv := 0
	for *getvar(&i), *getvar(&v) = range [2]int{1, 2} {
		si += i
		sv += v
	}
	if ncalls != 4 {
		println("wrong number of calls:", ncalls, "!= 4")
		panic("fail")
	}
	if si != 1 || sv != 3 {
		println("wrong sum in testcalls", si, sv)
		panic("fail")
	}

	ncalls = 0
	for *getvar(&i), *getvar(&v) = range [0]int{} {
		println("loop ran on empty array")
		panic("fail")
	}
	if ncalls != 0 {
		println("wrong number of calls:", ncalls, "!= 0")
		panic("fail")
	}

	ncalls = 0
	si = 0
	sv = 0
	for *getvar(&i), *getvar(&v) = range iter2(1, 2) {
		si += i
		sv += v
	}
	if ncalls != 4 {
		println("wrong number of calls:", ncalls, "!= 4")
		panic("fail")
	}
	if si != 1 || sv != 3 {
		println("wrong sum in testcalls", si, sv)
		panic("fail")
	}
}

func iter2(list ...int) func(func(int, int) bool) bool {
	return func(yield func(int, int) bool) bool {
		for i, x := range list {
			if !yield(i, x) {
				return false
			}
		}
		return true
	}
}

func main() {
	testblankvars()
	testchan()
	testarray()
	testarray1()
	testarray2()
	testarrayptr()
	testarrayptr1()
	testarrayptr2()
	testslice()
	testslice1()
	testslice2()
	testslice3()
	teststring()
	teststring1()
	teststring2()
	testmap()
	testmap1()
	testmap2()
	testint1()
	testint2()
	testint3()
	testfunc0()
	testfunc1()
	testfunc2()
	testfunc3()
	testfunc4()
	testfunc5()
	testfunc6()
	testfunc7()
	testfunc8()
	testcalls()
}
