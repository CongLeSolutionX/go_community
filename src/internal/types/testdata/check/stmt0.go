// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// statements

package stmt0

func assignments0() (int, int) {
	var a, b, c int
	var ch chan int
	f0 := func() {}
	f1 := func() int { return 1 }
	f2 := func() (int, int) { return 1, 2 }
	f3 := func() (int, int, int) { return 1, 2, 3 }

	a, b, c = 1, 2, 3
	a, b, c = 1 /* ERR assignment mismatch: 3 variables but 2 values */ , 2
	a, b, c = 1 /* ERR assignment mismatch: 3 variables but 4 values */ , 2, 3, 4
	_, _, _ = a, b, c

	a = f0 /* ERR used as value */ ()
	a = f1()
	a = f2 /* ERR assignment mismatch: 1 variable but f2 returns 2 values */ ()
	a, b = f2()
	a, b, c = f2 /* ERR assignment mismatch: 3 variables but f2 returns 2 values */ ()
	a, b, c = f3()
	a, b = f3 /* ERR assignment mismatch: 2 variables but f3 returns 3 values */ ()

	a, b, c = <- /* ERR assignment mismatch: 3 variables but 1 value */ ch

	return /* ERR not enough return values\n\thave ()\n\twant (int, int) */
	return 1 /* ERR not enough return values\n\thave (number)\n\twant (int, int) */
	return 1, 2
	return 1, 2, 3 /* ERR too many return values\n\thave (number, number, number)\n\twant (int, int) */
}

func assignments1() {
	b, i, f, c, s := false, 1, 1.0, 1i, "foo"
	b = i /* ERROR cannot use .* in assignment */
	i = f /* ERROR cannot use .* in assignment */
	f = c /* ERROR cannot use .* in assignment */
	c = s /* ERROR cannot use .* in assignment */
	s = b /* ERROR cannot use .* in assignment */

	v0, v1, v2 := 1 /* ERR assignment mismatch */ , 2, 3, 4
	_, _, _ = v0, v1, v2

	b = true

	i += 1
	i /* ERR mismatched types int and untyped string */+= "foo"

	f -= 1
	f /= 0
	f = float32(0)/0 /* ERR division by zero */
	f /* ERR mismatched types float64 and untyped string */-= "foo"

	c *= 1
	c /= 0

	s += "bar"
	s /* ERR mismatched types string and untyped int */+= 1

	var u64 uint64
	u64 += 1<<u64

	undefined /* ERR undefined */ = 991

	// test cases for issue 5800
	var (
		_ int = nil /* ERR cannot use nil as int value in variable declaration */
		_ [10]int = nil /* ERR cannot use nil as [10]int value in variable declaration */
		_ []byte = nil
		_ struct{} = nil /* ERR cannot use nil as struct{} value in variable declaration */
		_ func() = nil
		_ map[int]string = nil
		_ chan int = nil
	)

	// test cases for issue 5500
	_ = func() (int, bool) {
		var m map[int]int
		return m /* ERR not enough return values */ [0]
	}

	g := func(int, bool){}
	var m map[int]int
	g(m[0]) /* ERR not enough arguments */

	// assignments to _
	_ = nil /* ERR use of untyped nil */
	_ = 1  << /* ERR constant shift overflow */ 1000
	(_) = 0
}

func assignments2() {
	type mybool bool
	var m map[string][]bool
	var s []bool
	var b bool
	var d mybool
	_ = s
	_ = b
	_ = d

	// assignments to map index expressions are ok
	s, b = m["foo"]
	_, d = m["bar"]
	m["foo"] = nil
	m["foo"] = nil /* ERR assignment mismatch: 1 variable but 2 values */ , false
	_ = append(m["foo"])
	_ = append(m["foo"], true)

	var c chan int
	_, b = <-c
	_, d = <-c
	<- /* ERR cannot assign */ c = 0
	<-c = 0 /* ERR assignment mismatch: 1 variable but 2 values */ , false

	var x interface{}
	_, b = x.(int)
	x /* ERR cannot assign */ .(int) = 0
	x.(int) = 0 /* ERR assignment mismatch: 1 variable but 2 values */ , false

	assignments2 /* ERR used as value */ () = nil
	int /* ERR not an expression */ = 0
}

func issue6487() {
	type S struct{x int}
	_ = &S /* ERR cannot take address */ {}.x
	_ = &( /* ERR cannot take address */ S{}.x)
	_ = (&S{}).x
	S /* ERR cannot assign */ {}.x = 0
	(&S{}).x = 0

	type M map[string]S
	var m M
	m /* ERR cannot assign to struct field */ ["foo"].x = 0
	_ = &( /* ERR cannot take address */ m["foo"].x)
	_ = &m /* ERR cannot take address */ ["foo"].x
}

func issue6766a() {
	a, a /* ERR a repeated on left side of := */ := 1, 2
	_ = a
	a, b, b /* ERR b repeated on left side of := */ := 1, 2, 3
	_ = b
	c, c /* ERR c repeated on left side of := */, b := 1, 2, 3
	_ = c
	a, b := /* ERR no new variables */ 1, 2
}

func shortVarDecls1() {
	const c = 0
	type d int
	a, b, c /* ERR cannot assign */ , d /* ERR cannot assign */  := 1, "zwei", 3.0, 4
	var _ int = a // a is of type int
	var _ string = b // b is of type string
}

func incdecs() {
	const c = 3.14
	c /* ERR cannot assign */ ++
	s := "foo"
	s /* ERR invalid operation */ --
	3.14 /* ERR cannot assign */ ++
	var (
		x int
		y float32
		z complex128
	)
	x++
	y--
	z++
}

func sends() {
	var ch chan int
	var rch <-chan int
	var x int
	x <- /* ERR cannot send */ x
	rch <- /* ERR cannot send */ x
	ch <- "foo" /* ERROR cannot use .* in send */
	ch <- x
}

func selects() {
	select {}
	var (
		ch chan int
		sc chan <- bool
	)
	select {
	case <-ch:
	case (<-ch):
	case t := <-ch:
		_ = t
	case t := (<-ch):
		_ = t
	case t, ok := <-ch:
		_, _ = t, ok
	case t, ok := (<-ch):
		_, _ = t, ok
	case <-sc /* ERR cannot receive from send-only channel */ :
	}
	select {
	default:
	default /* ERR multiple defaults */ :
	}
	select {
	case a, b := <-ch:
		_, b = a, b
	case x /* ERR send or receive */ :
	case a /* ERR send or receive */ := ch:
	}

	// test for issue 9570: ch2 in second case falsely resolved to
	// ch2 declared in body of first case
	ch1 := make(chan int)
	ch2 := make(chan int)
	select {
	case <-ch1:
		var ch2 /* ERR ch2 declared and not used */ chan bool
	case i := <-ch2:
		print(i + 1)
	}
}

func gos() {
	go 1; /* ERR must be function call */
	go int /* ERR go requires function call, not conversion */ (0)
	go ( /* ERR expression in go must not be parenthesized */ gos())
	go gos()
	var c chan int
	go close(c)
	go len /* ERR go discards result */ (c)
}

func defers() {
	defer 1; /* ERR must be function call */
	defer int /* ERR defer requires function call, not conversion */ (0)
	defer ( /* ERR expression in defer must not be parenthesized */ defers())
	defer defers()
	var c chan int
	defer close(c)
	defer len /* ERR defer discards result */ (c)
}

func breaks() {
	var x, y int

	break /* ERR break */
	{
		break /* ERR break */
	}
	if x < y {
		break /* ERR break */
	}

	switch x {
	case 0:
		break
	case 1:
		if x == y {
			break
		}
	default:
		break
		break
	}

	var z interface{}
	switch z.(type) {
	case int:
		break
	}

	for {
		break
	}

	var a []int
	for _ = range a {
		break
	}

	for {
		if x == y {
			break
		}
	}

	var ch chan int
	select {
	case <-ch:
		break
	}

	select {
	case <-ch:
		if x == y {
			break
		}
	default:
		break
	}
}

func continues() {
	var x, y int

	continue /* ERR continue */
	{
		continue /* ERR continue */
	}

	if x < y {
		continue /* ERR continue */
	}

	switch x {
	case 0:
		continue /* ERR continue */
	}

	var z interface{}
	switch z.(type) {
	case int:
		continue /* ERR continue */
	}

	var ch chan int
	select {
	case <-ch:
		continue /* ERR continue */
	}

	for i := 0; i < 10; i++ {
		continue
		if x < y {
			continue
			break
		}
		switch x {
		case y:
			continue
		default:
			break
		}
		select {
		case <-ch:
			continue
		}
	}

	var a []int
	for _ = range a {
		continue
		if x < y {
			continue
			break
		}
		switch x {
		case y:
			continue
		default:
			break
		}
		select {
		case <-ch:
			continue
		}
	}
}

func returns0() {
	return
	return 0 /* ERR too many return values */
}

func returns1(x float64) (int, *float64) {
	return 0, &x
	return /* ERR not enough return values */
	return "foo" /* ERROR cannot .* in return statement */, x /* ERROR cannot use .* in return statement */
	return 0, &x, 1 /* ERR too many return values */
}

func returns2() (a, b int) {
	return
	return 1, "foo" /* ERROR cannot use .* in return statement */
	return 1, 2, 3 /* ERR too many return values */
	{
		type a int
		return 1, 2
		return /* ERR a not in scope at return */
	}
}

func returns3() (_ int) {
	return
	{
		var _ int // blank (_) identifiers never shadow since they are in no scope
		return
	}
}

func switches0() {
	var x int

	switch x {
	}

	switch x {
	default:
	default /* ERR multiple defaults */ :
	}

	switch {
	case 1  /* ERR cannot convert */ :
	}

	true := "false"
	_ = true
	// A tagless switch is equivalent to the bool
        // constant true, not the identifier 'true'.
	switch {
	case "false" /* ERR cannot convert */:
	}

	switch int32(x) {
	case 1, 2:
	case x /* ERR invalid case x in switch on int32(x) (mismatched types int and int32) */ :
	}

	switch x {
	case 1 /* ERR overflows */ << 100:
	}

	switch x {
	case 1:
	case 1 /* ERR duplicate case */ :
	case ( /* ERR duplicate case */ 1):
	case 2, 3, 4:
	case 5, 1 /* ERR duplicate case */ :
	}

	switch uint64(x) {
	case 1<<64 - 1:
	case 1 /* ERR duplicate case */ <<64 - 1:
	case 2, 3, 4:
	case 5, 1 /* ERR duplicate case */ <<64 - 1:
	}

	var y32 float32
	switch y32 {
	case 1.1:
	case 11/10: // integer division!
	case 11. /* ERR duplicate case */ /10:
	case 2, 3.0, 4.1:
	case 5.2, 1.10 /* ERR duplicate case */ :
	}

	var y64 float64
	switch y64 {
	case 1.1:
	case 11/10: // integer division!
	case 11. /* ERR duplicate case */ /10:
	case 2, 3.0, 4.1:
	case 5.2, 1.10 /* ERR duplicate case */ :
	}

	var s string
	switch s {
	case "foo":
	case "foo" /* ERR duplicate case */ :
	case "f" /* ERR duplicate case */ + "oo":
	case "abc", "def", "ghi":
	case "jkl", "foo" /* ERR duplicate case */ :
	}

	type T int
	type F float64
	type S string
	type B bool
	var i interface{}
	switch i {
	case nil:
	case nil: // no duplicate detection
	case (*int)(nil):
	case (*int)(nil): // do duplicate detection
	case 1:
	case byte(1):
	case int /* ERR duplicate case */ (1):
	case T(1):
	case 1.0:
	case F(1.0):
	case F /* ERR duplicate case */ (1.0):
	case "hello":
	case S("hello"):
	case S /* ERR duplicate case */ ("hello"):
	case 1==1, B(false):
	case false, B(2==2):
	}

	// switch on array
	var a [3]int
	switch a {
	case [3]int{1, 2, 3}:
	case [3]int{1, 2, 3}: // no duplicate detection
	case [ /* ERR "mismatched types */ 4]int{4, 5, 6}:
	}

	// switch on channel
	var c1, c2 chan int
	switch c1 {
	case nil:
	case c1:
	case c2:
	case c1, c2: // no duplicate detection
	}
}

func switches1() {
	fallthrough /* ERR fallthrough statement out of place */

	var x int
	switch x {
	case 0:
		fallthrough /* ERR fallthrough statement out of place */
		break
	case 1:
		fallthrough
	case 2:
		fallthrough; ; ; // trailing empty statements are ok
	case 3:
	default:
		fallthrough; ;
	case 4:
		fallthrough /* ERR cannot fallthrough final case in switch */
	}

	var y interface{}
	switch y.(type) {
	case int:
		fallthrough /* ERR cannot fallthrough in type switch */ ; ; ;
	default:
	}

	switch x {
	case 0:
		if x == 0 {
			fallthrough /* ERR fallthrough statement out of place */
		}
	}

	switch x {
	case 0:
		goto L1
		L1: fallthrough; ;
	case 1:
		goto L2
		goto L3
		goto L4
		L2: L3: L4: fallthrough
	default:
	}

	switch x {
	case 0:
		goto L5
		L5: fallthrough
	default:
		goto L6
		goto L7
		goto L8
		L6: L7: L8: fallthrough /* ERR cannot fallthrough final case in switch */
	}

	switch x {
	case 0:
		fallthrough; ;
	case 1:
		{
			fallthrough /* ERR fallthrough statement out of place */
		}
	case 2:
		fallthrough
	case 3:
		fallthrough /* ERR fallthrough statement out of place */
		{ /* empty block is not an empty statement */ }; ;
	default:
		fallthrough /* ERR cannot fallthrough final case in switch */
	}

	switch x {
	case 0:
		{
			fallthrough /* ERR fallthrough statement out of place */
		}
	}
}

func switches2() {
	// untyped nil is not permitted as switch expression
	switch nil /* ERR use of untyped nil */ {
	case 1, 2, "foo": // don't report additional errors here
	}

	// untyped constants are converted to default types
	switch 1<<63-1 {
	}
	switch 1 /* ERROR cannot use .* as int value.*\(overflows\) */ << 63 {
	}
	var x int
	switch 1.0 {
	case 1.0, 2.0, x /* ERR mismatched types int and float64 */ :
	}
	switch x {
	case 1.0:
	}

	// untyped bools become of type bool
	type B bool
	var b B = true
	switch x == x {
	case b /* ERR mismatched types B and bool */ :
	}
	switch {
	case b /* ERR mismatched types B and bool */ :
	}
}

func issue11667() {
	switch 9223372036854775808 /* ERROR cannot use .* as int value.*\(overflows\) */ {
	}
	switch 9223372036854775808 /* ERROR cannot use .* as int value.*\(overflows\) */ {
	case 9223372036854775808:
	}
	var x int
	switch x {
	case 9223372036854775808 /* ERR overflows int */ :
	}
	var y float64
	switch y {
	case 9223372036854775808:
	}
}

func issue11687() {
	f := func() (_, _ int) { return }
	switch f /* ERR multiple-value f */ () {
	}
	var x int
	switch f /* ERR multiple-value f */ () {
	case x:
	}
	switch x {
	case f /* ERR multiple-value f */ ():
	}
}

type I interface {
	m()
}

type I2 interface {
	m(int)
}

type T struct{}
type T1 struct{}
type T2 struct{}

func (T) m() {}
func (T2) m(int) {}

func typeswitches() {
	var i int
	var x interface{}

	switch x.(type) {}
	switch (x /* ERR outside type switch */ .(type)) {}

	switch x.(type) {
	default:
	default /* ERR multiple defaults */ :
	}

	switch x /* ERR declared and not used */ := x.(type) {}
	switch _ /* ERR no new variable on left side of := */ := x.(type) {}

	switch x := x.(type) {
	case int:
		var y int = x
		_ = y
	}

	switch x /* ERR x declared and not used */ := i /* ERR not an interface */ .(type) {}

	switch t := x.(type) {
	case nil:
		var v bool = t /* ERROR cannot use .* in variable declaration */
		_ = v
	case int:
		var v int = t
		_ = v
	case float32, complex64:
		var v float32 = t /* ERROR cannot use .* in variable declaration */
		_ = v
	default:
		var v float32 = t /* ERROR cannot use .* in variable declaration */
		_ = v
	}

	var t I
	switch t.(type) {
	case T:
	case T1 /* ERR missing method m */ :
	case T2 /* ERR wrong type for method m */ :
	case I2 /* STRICT "wrong type for method m" */ : // only an error in strict mode (issue 8561)
	}


	{
		x := 1
		v := 2
		switch v /* ERR v (variable of type int) is not an interface */ .(type) {
		case int:
			println(x)
			println(x / 0 /* ERR invalid operation: division by zero */)
		case 1 /* ERR 1 is not a type */:
		}
	}
}

// Test that each case clause uses the correct type of the variable
// declared by the type switch (issue 5504).
func typeswitch0() {
	switch y := interface{}(nil).(type) {
	case int:
		func() int { return y + 0 }()
	case float32:
		func() float32 { return y }()
	}
}

// Test correct scope setup.
// (no redeclaration errors expected in the type switch)
func typeswitch1() {
	var t I
	switch t := t; t := t.(type) {
	case nil:
		var _ I = t
	case T:
		var _ T = t
	default:
		var _ I = t
	}
}

// Test correct typeswitch against interface types.
type A interface { a() }
type B interface { b() }
type C interface { a(int) }

func typeswitch2() {
	switch A(nil).(type) {
	case A:
	case B:
	case C /* STRICT "cannot have dynamic type" */: // only an error in strict mode (issue 8561)
	}
}

func typeswitch3(x interface{}) {
	switch x.(type) {
	case int:
	case float64:
	case int /* ERR duplicate case */ :
	}

	switch x.(type) {
	case nil:
	case int:
	case nil /* ERR duplicate case */ , nil /* ERR duplicate case */ :
	}

	type F func(int)
	switch x.(type) {
	case nil:
	case int, func(int):
	case float32, func /* ERR duplicate case */ (x int):
	case F:
	}
}

func fors1() {
	for {}
	var i string
	_ = i
	for i := 0; i < 10; i++ {}
	for i := 0; i < 10; j /* ERR cannot declare */ := 0 {}
}

func rangeloops1() {
	var (
		x int
		a [10]float32
		b []string
		p *[10]complex128
		pp **[10]complex128
		s string
		m map[int]bool
		c chan int
		sc chan<- int
		rc <-chan int
	)

	for range x /* ERR cannot range over */ {}
	for _ = range x /* ERR cannot range over */ {}
	for i := range x /* ERR cannot range over */ {}

	for range a {}
	for i := range a {
		var ii int
		ii = i
		_ = ii
	}
	for i, x := range a {
		var ii int
		ii = i
		_ = ii
		var xx float64
		xx = x /* ERROR cannot use .* in assignment */
		_ = xx
	}
	var ii int
	var xx float32
	for ii, xx = range a {}
	_, _ = ii, xx

	for range b {}
	for i := range b {
		var ii int
		ii = i
		_ = ii
	}
	for i, x := range b {
		var ii int
		ii = i
		_ = ii
		var xx string
		xx = x
		_ = xx
	}

	for range s {}
	for i := range s {
		var ii int
		ii = i
		_ = ii
	}
	for i, x := range s {
		var ii int
		ii = i
		_ = ii
		var xx rune
		xx = x
		_ = xx
	}

	for range p {}
	for _, x := range p {
		var xx complex128
		xx = x
		_ = xx
	}

	for range pp /* ERR cannot range over */ {}
	for _, x := range pp /* ERR cannot range over */ {}

	for range m {}
	for k := range m {
		var kk int32
		kk = k /* ERROR cannot use .* in assignment */
		_ = kk
	}
	for k, v := range m {
		var kk int
		kk = k
		_ = kk
		if v {}
	}

	for range c {}
	for _, _ /* ERR only one iteration variable */ = range c {}
	for e := range c {
		var ee int
		ee = e
		_ = ee
	}
	for _ = range sc /* ERR cannot range over */ {}
	for _ = range rc {}

	// constant strings
	const cs = "foo"
	for range cs {}
	for range "" {}
	for i, x := range cs { _, _ = i, x }
	for i, x := range "" {
		var ii int
		ii = i
		_ = ii
		var xx rune
		xx = x
		_ = xx
	}
}

func rangeloops2() {
	type I int
	type R rune

	var a [10]int
	var i I
	_ = i
	for i /* ERROR cannot use .* in assignment */ = range a {}
	for i /* ERROR cannot use .* in assignment */ = range &a {}
	for i /* ERROR cannot use .* in assignment */ = range a[:] {}

	var s string
	var r R
	_ = r
	for i /* ERROR cannot use .* in assignment */ = range s {}
	for i /* ERROR cannot use .* in assignment */ = range "foo" {}
	for _, r /* ERROR cannot use .* in assignment */ = range s {}
	for _, r /* ERROR cannot use .* in assignment */ = range "foo" {}
}

func issue6766b() {
	for _ := /* ERR no new variables */ range "" {}
	for a, a /* ERR redeclared */ := range "" { _ = a }
	var a int
	_ = a
	for a, a /* ERR redeclared */ := range []int{1, 2, 3} { _ = a }
}

// Test that despite errors in the range clause,
// the loop body is still type-checked (and thus
// errors reported).
func issue10148() {
	for y /* ERR declared and not used */ := range "" {
		_ = "" /* ERR mismatched types untyped string and untyped int */ + 1
	}
	for range 1 /* ERR cannot range over 1 */ {
		_ = "" /* ERR mismatched types untyped string and untyped int */ + 1
	}
	for y := range 1 /* ERR cannot range over 1 */ {
		_ = "" /* ERR mismatched types untyped string and untyped int */ + 1
	}
}

func labels0() {
	goto L0
	goto L1
	L0:
	L1:
	L1 /* ERR already declared */ :
	if true {
		goto L2
		L2:
		L0 /* ERR already declared */ :
	}
	_ = func() {
		goto L0
		goto L1
		goto L2
		L0:
		L1:
		L2:
	}
}

func expression_statements(ch chan int) {
	expression_statements(ch)
	<-ch
	println()

	0 /* ERR not used */
	1 /* ERR not used */ +2
	cap /* ERR not used */ (ch)
	println /* ERR must be called */
}
