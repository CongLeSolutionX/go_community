package main

import (
	"fmt"
	"reflect"
	"unsafe"
)

func AddInt() {
	v1, v2 := 1, 3
	res := reflect.ValueOf(addInt).Call([]reflect.Value{reflect.ValueOf(v1), reflect.ValueOf(v2)})
	r1 := res[0].Interface().(int)
	if r1 != v1+v2 {
		print("want: ", v1+v2, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func AddFloat64() {
	v1, v2 := float64(2.0), float64(3.5)
	res := reflect.ValueOf(addFloat64).Call([]reflect.Value{reflect.ValueOf(v1), reflect.ValueOf(v2)})
	r1 := res[0].Interface().(float64)
	if r1 != v1+v2 {
		print("want: ", v1+v2, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}

func SumSpillInt() {
	v1, v2, v3, v4, v5, v6, v7, v8, v9, v10 := 1, 2, 3, 4, 5, 6, 7, 8, 9, 10
	res := reflect.ValueOf(sumSpillInt).Call([]reflect.Value{
		reflect.ValueOf(v1),
		reflect.ValueOf(v2),
		reflect.ValueOf(v3),
		reflect.ValueOf(v4),
		reflect.ValueOf(v5),
		reflect.ValueOf(v6),
		reflect.ValueOf(v7),
		reflect.ValueOf(v8),
		reflect.ValueOf(v9),
		reflect.ValueOf(v10),
	})
	r1 := res[0].Interface().(int)
	if want := v1 + v2 + v3 + v4 + v5 + v6 + v7 + v8 + v9 + v10; r1 != want {
		print("want: ", want, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func SumSpillFloat64() {
	f1, f2, f3, f4, f5 := 1.0, 2.0, 3.0, 4.0, 5.0
	f6, f7, f8, f9, f10 := 6.0, 7.0, 8.0, 9.0, 10.0
	f11, f12, f13, f14, f15, f16 := 11.0, 12.0, 13.0, 14.0, 15.0, 16.0
	res := reflect.ValueOf(sumSpillFloat64).Call([]reflect.Value{
		reflect.ValueOf(f1),
		reflect.ValueOf(f2),
		reflect.ValueOf(f3),
		reflect.ValueOf(f4),
		reflect.ValueOf(f5),
		reflect.ValueOf(f6),
		reflect.ValueOf(f7),
		reflect.ValueOf(f8),
		reflect.ValueOf(f9),
		reflect.ValueOf(f10),
		reflect.ValueOf(f11),
		reflect.ValueOf(f12),
		reflect.ValueOf(f13),
		reflect.ValueOf(f14),
		reflect.ValueOf(f15),
		reflect.ValueOf(f16),
	})
	r1 := res[0].Interface().(float64)
	want := f1 + f2 + f3 + f4 + f5 + f6 + f7 + f8 + f9 + f10 +
		f11 + f12 + f13 + f14 + f15 + f16
	if r1 != want {
		print("want: ", want, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func SumSpillMix() {
	v1, v2, v3, v4, v5, v6, v7, v8, v9, v10 := 1, 2, 3, 4, 5, 6, 7, 8, 9, 10
	f1, f2, f3, f4, f5 := 1.0, 2.0, 3.0, 4.0, 5.0
	f6, f7, f8, f9, f10 := 6.0, 7.0, 8.0, 9.0, 10.0
	f11, f12, f13, f14, f15, f16 := 11.0, 12.0, 13.0, 14.0, 15.0, 16.0
	res := reflect.ValueOf(sumSpillMix).Call([]reflect.Value{
		reflect.ValueOf(v1),
		reflect.ValueOf(v2),
		reflect.ValueOf(v3),
		reflect.ValueOf(v4),
		reflect.ValueOf(v5),
		reflect.ValueOf(v6),
		reflect.ValueOf(v7),
		reflect.ValueOf(v8),
		reflect.ValueOf(v9),
		reflect.ValueOf(v10),
		reflect.ValueOf(f1),
		reflect.ValueOf(f2),
		reflect.ValueOf(f3),
		reflect.ValueOf(f4),
		reflect.ValueOf(f5),
		reflect.ValueOf(f6),
		reflect.ValueOf(f7),
		reflect.ValueOf(f8),
		reflect.ValueOf(f9),
		reflect.ValueOf(f10),
		reflect.ValueOf(f11),
		reflect.ValueOf(f12),
		reflect.ValueOf(f13),
		reflect.ValueOf(f14),
		reflect.ValueOf(f15),
		reflect.ValueOf(f16),
	})
	r1 := res[0].Interface().(int)
	r2 := res[1].Interface().(float64)
	want1 := v1 + v2 + v3 + v4 + v5 + v6 + v7 + v8 + v9 + v10
	want2 := f1 + f2 + f3 + f4 + f5 + f6 + f7 + f8 + f9 + f10 +
		f11 + f12 + f13 + f14 + f15 + f16
	if r1 != want1 || r2 != want2 {
		print("(1) want: ", want1, ", got: ", r1, "\n")
		print("(2) want: ", want2, ", got: ", r2, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}

func SplitSpillInt() {
	v1 := 20
	res := reflect.ValueOf(splitSpillInt).Call([]reflect.Value{
		reflect.ValueOf(v1),
	})
	r1 := res[0].Interface().(int)
	r2 := res[1].Interface().(int)
	r3 := res[2].Interface().(int)
	r4 := res[3].Interface().(int)
	r5 := res[4].Interface().(int)
	r6 := res[5].Interface().(int)
	r7 := res[6].Interface().(int)
	r8 := res[7].Interface().(int)
	r9 := res[8].Interface().(int)
	r10 := res[9].Interface().(int)
	got := r1 + r2 + r3 + r4 + r5 + r6 + r7 + r8 + r9 + r10
	if want := v1; got != want {
		print("want: ", want, ", got: ", got, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}

func PassArray1() {
	a := [1]uint32{120}
	res := reflect.ValueOf(passArray1).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().([1]uint32)
	if r1 != a {
		print("want: ", a[0], ", got: ", r1[0], "\n")
		panic("bad reflect call")
	}
	println("PASS")
}

func PassArray() {
	a := [2]uintptr{5, 111}
	res := reflect.ValueOf(passArray).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().([2]uintptr)
	if r1 != a {
		print("(1) want: ", a[0], ", got: ", r1[0], "\n")
		print("(2) want: ", a[1], ", got: ", r1[1], "\n")
		panic("bad reflect call")
	}
	println("PASS")
}

func PassArray1Mix() {
	f := 5
	a := [1]uint32{120}
	g := 0.12
	res := reflect.ValueOf(passArray1Mix).Call([]reflect.Value{
		reflect.ValueOf(f),
		reflect.ValueOf(a),
		reflect.ValueOf(g),
	})
	r1 := res[0].Interface().(int)
	r2 := res[1].Interface().([1]uint32)
	r3 := res[2].Interface().(float64)
	if r1 != f || r2 != a || r3 != g {
		print("want: ", f, ", got: ", r1, "\n")
		print("want: ", a[0], ", got: ", r2[0], "\n")
		print("want: ", g, ", got: ", r3, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassString() {
	a := "PASS"
	res := reflect.ValueOf(passString).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(string)
	if r1 != a {
		print("want: ", a[0], ", got: ", r1[0], "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassInterface() {
	a := 52
	var i interface{}
	i = &a
	res := reflect.ValueOf(passInterface).Call([]reflect.Value{
		reflect.ValueOf(i),
	})
	r1 := res[0].Interface().(interface{})
	a1 := r1.(*int)
	if a1 != &a {
		print("want: ", unsafe.Pointer(&a), ", got: ", unsafe.Pointer(a1), "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassSlice() {
	a := []byte{1, 2, 4, 8, 16, 24, 32}
	res := reflect.ValueOf(passSlice).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().([]byte)
	if &r1[0] != &a[0] || len(r1) != len(a) || cap(r1) != cap(a) {
		print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func SetPointer() {
	var a byte
	res := reflect.ValueOf(setPointer).Call([]reflect.Value{
		reflect.ValueOf(&a),
	})
	r1 := res[0].Interface().(*byte)
	if r1 != &a || a != 231 {
		print("want: ", &a, ", got: ", r1, "\n")
		print("want: ", 231, ", got: ", a, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}

func PassStruct1() {
	a := Struct1{
		a: 12,
		b: 11220,
		c: 2,
	}
	res := reflect.ValueOf(passStruct1).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct1)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct2() {
	a := Struct2{
		a: 9,
		b: 34492,
		c: 1000000,
		d: [2]uint32{1, 2},
	}
	res := reflect.ValueOf(passStruct2).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct2)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct3() {
	a := Struct3{
		a: 198,
		b: 1111,
		c: 9992,
	}
	res := reflect.ValueOf(passStruct3).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct3)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct4() {
	a := Struct4{
		a: -1,
		b: -125,
		c: 251,
		d: 34,
		e: true,
	}
	res := reflect.ValueOf(passStruct4).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct4)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct5() {
	a := Struct5{
		a: 1,
		b: -6666,
		c: 10029,
		d: 762881727,
		e: -92381037,
		f: -2.29e19,
		g: 251.3e-1,
		h: 1.0,
		i: 888.19191,
		j: 0.12918212,
	}
	res := reflect.ValueOf(passStruct5).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct5)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct6() {
	a := Struct6{
		Struct1: Struct1{
			a: 1,
			b: 2,
			c: 3,
		},
	}
	res := reflect.ValueOf(passStruct6).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct6)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct7() {
	a := Struct7{
		Struct1: Struct1{
			a: 1,
			b: 2,
			c: 3,
		},
		Struct2: Struct2{
			a: 5,
			b: 4,
			c: 6,
			d: [2]uint32{2, 5},
		},
	}
	res := reflect.ValueOf(passStruct7).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct7)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct8() {
	a := Struct8{
		Struct5: Struct5{
			a: 1,
			b: -6666,
			c: 10029,
			d: 762881727,
			e: -92381037,
			f: -2.29e19,
			g: 251.3e-1,
			h: 1.0,
			i: 888.19191,
			j: 0.12918212,
		},
		Struct1: Struct1{
			a: 198,
			b: 1111,
			c: 9992,
		},
	}
	res := reflect.ValueOf(passStruct8).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct8)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct9() {
	a := Struct9{
		Struct1: Struct1{
			a: 1,
			b: 6666,
			c: 10029,
		},
		Struct7: Struct7{
			Struct1: Struct1{
				a: 1,
				b: 2,
				c: 3,
			},
			Struct2: Struct2{
				a: 5,
				b: 4,
				c: 6,
				d: [2]uint32{2, 5},
			},
		},
	}
	res := reflect.ValueOf(passStruct9).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct9)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct10() {
	a := Struct10{
		Struct5: Struct5{
			a: 1,
			b: -6666,
			c: 10029,
			d: 762881727,
			e: -92381037,
			f: -2.29e19,
			g: 251.3e-1,
			h: 1.0,
			i: 888.19191,
			j: 0.12918212,
		},
		Struct8: Struct8{
			Struct5: Struct5{
				a: 1,
				b: -6666,
				c: 10029,
				d: 762881727,
				e: -92381037,
				f: -2.29e19,
				g: 251.3e-1,
				h: 1.0,
				i: 888.19191,
				j: 0.12918212,
			},
			Struct1: Struct1{
				a: 198,
				b: 1111,
				c: 9992,
			},
		},
	}
	res := reflect.ValueOf(passStruct10).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct10)
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct11() {
	w := new(byte)
	*w = 10
	x := make(map[string]int)
	x["hello"] = 5
	y := make(chan int, 1)
	y <- 11
	z := func() int {
		return 100
	}
	a := Struct11{
		w: unsafe.Pointer(w),
		x: x,
		y: y,
		z: z,
	}
	res := reflect.ValueOf(passStruct11).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct11)
	if r1.w != unsafe.Pointer(w) || r1.x["hello"] != 5 || <-r1.y != 11 || r1.z() != 100 {
		panic("bad reflect call")
	}
	println("PASS")
}
func PassStruct12() {
	w := new(byte)
	*w = 10
	x := make(map[string]int)
	x["hello"] = 5
	y := make(chan int, 1)
	y <- 11
	z := func() int {
		return 100
	}
	a := Struct12{
		a: -11111,
		Struct11: Struct11{
			w: unsafe.Pointer(w),
			x: x,
			y: y,
			z: z,
		},
	}
	res := reflect.ValueOf(passStruct12).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct12)
	if r1.a != -11111 || r1.w != unsafe.Pointer(w) || r1.x["hello"] != 5 || <-r1.y != 11 || r1.z() != 100 {
		panic("bad reflect call")
	}
	println("PASS")
}
func IncStruct13() {
	a := Struct13{
		a: 12,
		b: 11220,
	}
	res := reflect.ValueOf(incStruct13).Call([]reflect.Value{
		reflect.ValueOf(a),
	})
	r1 := res[0].Interface().(Struct13)
	a.a++
	a.b++
	if r1 != a {
		fmt.Print("want: ", a, ", got: ", r1, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func Pass2Struct1() {
	a := Struct1{
		a: 12,
		b: 11220,
		c: 2,
	}
	b := Struct1{
		a: 5,
		b: 55555,
		c: 55,
	}
	res := reflect.ValueOf(pass2Struct1).Call([]reflect.Value{
		reflect.ValueOf(a),
		reflect.ValueOf(b),
	})
	r1 := res[0].Interface().(Struct1)
	r2 := res[1].Interface().(Struct1)
	if r1 != a || r2 != b {
		fmt.Print("(1) want: ", a, ", got: ", r1, "\n")
		fmt.Print("(2) want: ", b, ", got: ", r2, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}
func PassEmptyStruct() {
	a := 9
	b := struct{}{}
	c := 0.112
	res := reflect.ValueOf(passEmptyStruct).Call([]reflect.Value{
		reflect.ValueOf(a),
		reflect.ValueOf(b),
		reflect.ValueOf(c),
	})
	r1 := res[0].Interface().(int)
	_, ok := res[1].Interface().(struct{})
	r3 := res[2].Interface().(float64)
	if r1 != a || r3 != c || !ok {
		fmt.Print("(1) want: ", a, ", got: ", r1, "\n")
		fmt.Print("(2) ", ok, "\n")
		fmt.Print("(3) want: ", c, ", got: ", r3, "\n")
		panic("bad reflect call")
	}
	println("PASS")
}

func main() {
	AddInt()
	AddFloat64()

	SumSpillInt()
	SumSpillFloat64()
	SumSpillMix()

	SplitSpillInt()

	PassArray1()
	PassArray()
	PassArray1Mix()
	PassString()
	PassInterface()
	PassSlice()
	SetPointer()

	PassStruct1()
	PassStruct2()
	PassStruct3()
	PassStruct4()
	PassStruct5()
	PassStruct6()
	PassStruct7()
	PassStruct8()
	PassStruct9()
	PassStruct10()
	PassStruct11()
	PassStruct12()
	IncStruct13()
	Pass2Struct1()
	PassEmptyStruct()
}
