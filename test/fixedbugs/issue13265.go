// errorcheck -0 -race

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 13265: nil pointer deref.

package main

var SINK interface{}
var Var1 = (*((((<-make(chan [2][]*bool, 1))[(<-make(chan int))])[:((<-make(chan [2]map[[2]int]int, 1))[(int)((len)(Var51))])[Var52]])[(int)((*(([]*float32{})[((func([2]struct{}) int)(nil))([2]struct{}{})])))]))
var Var52 = ([][2]int{4: ((func(float64, chan chan int) func(map[struct{}]int16, chan complex64, [2]interface{}) [2]int)(nil))(1.0, make(chan chan int, 1))(([0]map[struct{}]int16{})[((func(uint, struct{}) int)(nil))(uint(1), (struct{}{}))], make(chan complex64, 1), [2]interface{}{})})[((func(byte, uintptr, complex128) int)(nil))(byte(0), uintptr(0), 1i)]

func init() {
	return
}

var Var454 = (**[0]uintptr)(nil)
var Var475 = [1][1][0]int16{}
var Var476 = 1

func main() {
	// race instrumentation segfaults here
	for ; false; (*(((<-Var343)[(<-((*(Var393))[(((make([][0]int, 0, 1))[(*(*(([]**int{})[(*((Var406)[Var476]))-(1)])))+(1)])[(<-make(chan int, 1))])+(1)])[(*(([]*int{})[(<-make(chan int))]))+(1)])])[(int)((Var667)[((func(chan string, uint) int)(nil))(make(chan string), uint(1))])])) = (*[]interface{})(nil) {
	}
	return
}

var Var406 = (<-(*((([]*chan [2]*int{})[((Var453)[(int)((*(*(Var454)))[(*(*(([]**[1]int{})[(int)(([]byte{})[(int)((((Var475)[Var476])[(int)((len)(([]string{})[(struct {
	Field481 int16
	Field482 int
	Field483 float32
	Field484 rune
	Field485 chan bool
	Field486 interface {
		Method487(bool, int16, rune, byte) (bool, float32, uintptr, uintptr)
		Method488(int16) (uint, byte, rune, int16)
		Method489(float32, float32) (uint, uintptr, interface{}, uintptr)
	}
	Field490 struct {
		Field491 error
		Field492 error
	}
	Field493 float32
}{}).Field482]))])[(int)(([]rune{})[((func(uintptr, interface {
	Method494(map[int]int16, chan float32, map[complex64]complex128, interface{}) rune
	Method495(interface {
		Method496(complex128, byte) (float32, complex64, uint, uint)
		Method497(string, rune, int16) string
		Method498(complex64, byte, uintptr, int16) (error, byte)
	}, [2]uintptr, chan uintptr, [2]float64) (func(rune, error, byte) (complex64, string), struct {
		Field499 float32
		Field500 uintptr
		Field501 uint
		Field502 uint
		Field503 complex64
	}, string, struct {
		Field504 string
		Field505 interface{}
		Field506 complex128
		Field507 string
		Field508 rune
	})
}) int)(nil))(uintptr(0), interface {
	Method494(map[int]int16, chan float32, map[complex64]complex128, interface{}) rune
	Method495(interface {
		Method496(complex128, byte) (float32, complex64, uint, uint)
		Method497(string, rune, int16) string
		Method498(complex64, byte, uintptr, int16) (error, byte)
	}, [2]uintptr, chan uintptr, [2]float64) (func(rune, error, byte) (complex64, string), struct {
		Field499 float32
		Field500 uintptr
		Field501 uint
		Field502 uint
		Field503 complex64
	}, string, struct {
		Field504 string
		Field505 interface{}
		Field506 complex128
		Field507 string
		Field508 rune
	})
}(nil))])])])])))[(<-make(chan int))]])])[((func(byte, rune, map[string]bool) int)(nil))(byte(0), rune(0), map[string]bool{})]:])[(int)(Var623)])))
var Var667 = [2]float32{}

func init() {
	return
}

var Var901 = ((*((((((func(bool) [][0][2]*[1][]*[0]int)(nil))(false))[([2]func(bool, struct{}) int{})[(int)((Var1012).Field1011)](Var1014, (struct{}{}))])[(struct {
	Field1041 complex128
	Field1042 []interface{}
	Field1043 []complex64
	Field1044 interface{}
	Field1045 float32
	Field1046 interface{}
	Field1047 func(int, bool, float32) (string, string, float32)
	Field1048 int
}{}).Field1048])[(int)(Var1120)]))[(Var1121)-(1)])[(((([][1][0]int{})[(int)((*(Var1140)))])[(int)((cap)((*(([]*[1]chan byte{})[(make(map[interface {
	Method2621(uintptr, bool) (error, int16, int, error)
}]int, 1))[interface {
	Method2621(uintptr, bool) (error, int16, int, error)
}(nil)]]))[(*((([][1]*int{})[((func(bool, string, byte) int)(nil))(false, "foo", byte(0))])[((func(error) int)(nil))(error(nil))]))-(1)]))])[((func(string, rune) int)(nil))("foo", rune(0))])-(1)]
var Var1012 = Var1013
var Var1121 = copy((<-((func(complex64, func(rune, uint, interface {
	Method1128(error, complex64) complex128
	Method1129(interface{}, complex64) (int, int16)
}) (int16, interface{}), error, [1]*float64) chan func(uintptr) []chan int16)(nil))(complex64(1i), ((func(rune, uint, interface {
	Method1128(error, complex64) complex128
	Method1129(interface{}, complex64) (int, int16)
}) (int16, interface{}))(nil)), error(nil), [1]*float64{}))(uintptr(0)), []chan int16{})
var Var1140 = ([]*float32{})[(*(((([][2][1]*int{(*((((func(error) []*[2][1]*int)(nil))(error(nil)))[(int)((len)((*(Var1468)).Field1286))])), (<-make(chan [][2][1]*int))[(int)(Var1121)]})[(int)(((([][1][2]int16{})[((Var1669)[((([]struct {
	Field1695 interface{}
	Field1696 []bool
	Field1697 int
}{})[([]int{})[(<-make(chan int))]]).Field1697)+(1)])-(1)])[(*(Var1733))])[(*(Var1813))])])[(int)(((Var1817)[(int)(([]float64{})[(<-make(chan int, 1))])])[(int)((len)((*(([]*string{})[(((*((Var2348)[(struct {
	Field2405 int16
	Field2406 complex64
	Field2407 int
	Field2408 rune
}{}).Field2407]))[(int)((*(Var2409)))])[(int)((Var2465).Field2464)])+(1)]))))])])[Var1121]))-(1)]
var Var2348 = [1]*[0][1]int{}
var Var2465 = (struct {
	Field2464 rune
}{})
var Var51 = "foo"

var Var1468 = (*struct {
	Field1285 [2]float32
	Field1286 string
})(&(Var1537))
var Var393 = (*[][2]chan int)(nil)

var Var623 = (uintptr)(([]func(rune, func(uintptr) (*byte, map[int16]int), byte, map[interface{}]func(uintptr, rune) (float32, complex128, interface{})) func(string, string) float64{})[((*(Var901))[((func(int16, *interface{}, func([]error, int) (uintptr, float64, [1]complex128, interface{}), rune) int)(nil))(int16(1), (*interface{})(nil), ((func([]error, int) (uintptr, float64, [1]complex128, interface{}))(nil)), rune(0))])-(1)](rune(0), ((func(uintptr) (*byte, map[int16]int))(nil)), byte(0), map[interface{}]func(uintptr, rune) (float32, complex128, interface{}){})("foo", "foo"))
var Var1817 = (*(*(([]**[1][0][0]rune{})[((Var2005)[(([]int{})[(struct {
	Field2023 int
	Field2024 chan byte
	Field2025 struct {
		Field2026 interface{}
		Field2027 complex64
	}
	Field2028 int
}{}).Field2023])-(1)])[(((([][2][0]int{})[((func(uintptr) int)(nil))(uintptr(0))])[(int)((*(([]*uintptr{})[((func(*interface{}, error, bool, complex128) int)(nil))((*interface{})(nil), error(nil), false, 1i)])))])[(int)((([][1]float64{})[(Var2148)[((func(map[*error]interface{}) int)(nil))(map[*error]interface{}{})]])[(<-make(chan int))])])-(1)]])))[(([]int{})[((func(interface{}, interface{}, *chan uint, rune) int)(nil))(interface{}(nil), interface{}(nil), (*chan uint)(nil), rune(0))])-(1)]
var Var2005 = [2][2]int{}
var Var453 = [][2]int{}
var Var342 = ((func(bool, int, rune, int16) (chan complex128, chan complex64, []int))(nil))

var Var2884 = [2][2]int16{0: ((*(Var2885))("foo", []chan string{}, complex64(1i)))[((([][0][0]int{})[(<-make(chan int, 1))])[(struct {
	Field5883 struct{}
	Field5884 int
}{}).Field5884])[(Var3667)-(1)]], 1: [2]int16{}}
var Var2885 = Var2886
var Var2886 = Var2887
var Var3107 = Var3116
var Var3372 = (*(([]*[1][1]*int{})[(([]int{})[((Var3386)[((func(struct {
	Field3387 complex128
	Field3388 [2]string
}) int)(nil))((struct {
	Field3387 complex128
	Field3388 [2]string
}{}))])-(1)])+(1)]))
var Var3714 = [1]int{}
var Var4742 = uint(1)
var Var4875 = (*[0]uintptr)(nil)
var Var5598 = (*int)(nil)
var Var5690 = (*int)(nil)
var Var6657 = (***[2][2]string)(nil)
var Var6658 = [0]int{}
var Var2148 = [1]int{}
var Var1014 = false
var Var1013 = (struct {
	Field1011 rune
}{})
var Var2887 = ((Var2962)[(*((*((Var3083)[(make(map[error]int))[error(nil)]]))[((func(struct{}, func(complex128) (*byte, uint, int, struct {
	Field5689 interface{}
})) int)(nil))((struct{}{}), ((func(complex128) (*byte, uint, int, struct {
	Field5689 interface{}
}))(nil)))]))[(struct{}{})]])[(*(Var5690))]
var Var3116 = func(Param3117 **error, Param3118 error, Param3119 uintptr, Param3120 uintptr) *[2]*map[struct{}]int {
	((*(((*(*((([2]func([]float64, complex128, bool) []**[][0]*[1][2][0]complex128{})[(*(((([][1][0]*int{})[((([][2][2]int{})[(*(([]*int{})[(make(map[chan int16]int, 1))[make(chan int16)]]))])[Var3305])[(*(([]*int{})[(int)((*((([][0]*int{})[(int)((len)((*((((Var4283)[(int)(Var3667)])[(<-make(chan int, 1))])[(int)((len)((*(((([][1][2]*[0]map[rune]uintptr{})[Var3667])[(int)(([]uint{})[((func(complex128) int)(nil))(1i)])])[(int)((len)((*(*(*(Var4503))))))]))[((*(([]*[2]int{})[(int)((len)(((*(([]*[0][1]string{})[(int)((len)(Var4560))]))[(int)((([][1]int{})[(<-make(chan int))])[(map[interface{}]int{})[interface{}(nil)]])])[(*(*(Var4741)))]))]))[(int)(((([][1][0]rune{})[(int)(Var4742)])[(<-make(chan int, 1))])[(((Var4743)[(int)(([]float32{})[(int)(([]float64{})[(struct {
		Field4749 bool
		Field4750 int
	}{}).Field4750])])])[(int)((*(Var4875))[((func(*func(complex128, bool, error, complex64) (complex128, string), uintptr, interface {
		Method4876(complex64, interface {
			Method4877(error) (uintptr, uintptr)
			Method4878(byte, complex64, float32, byte) (string, bool, float64)
			Method4879(uint, bool, int, uint) (complex128, float64, byte, error)
		}, []float64) (uint, int, *string)
	}) int)(nil))((*func(complex128, bool, error, complex64) (complex128, string))(nil), uintptr(0), interface {
		Method4876(complex64, interface {
			Method4877(error) (uintptr, uintptr)
			Method4878(byte, complex64, float32, byte) (string, bool, float64)
			Method4879(uint, bool, int, uint) (complex128, float64, byte, error)
		}, []float64) (uint, int, *string)
	}(nil))])])-(1)])])+(1)]))]))[(int)((cap)(Var4880))]))])[([]int{})[(int)((len)((([][0]string{})[(struct {
		Field4935 *complex128
		Field4936 rune
		Field4937 byte
		Field4938 struct{}
		Field4939 rune
		Field4940 int
		Field4941 func(string, string, interface{}) (complex64, interface{}, float64)
	}{}).Field4940])[(<-make(chan int))]))]])))]))+(1)]])[(int)((cap)((*((([][1]*chan interface {
		Method4944(byte) (float64, int)
	}{})[(<-make(chan int, 1))])[((func(*func(float32, uintptr, uintptr) int16, float64, func(*string, []rune, *complex128) int) int)(nil))((*func(float32, uintptr, uintptr) int16)(nil), 1.0, ((func(*string, []rune, *complex128) int)(nil)))]))))])[(int)((Var5063)[(int)((([][0]float32{})[((func(string) int)(nil))("foo")])[(int)((*((Var5096)[(*(Var5122))-(1)])))])])]))]([]float64{}, 1i, false))[((func(func(error, map[bool]string, bool, error) (bool, bool), struct{}, complex64) int)(nil))(((func(error, map[bool]string, bool, error) (bool, bool))(nil)), (struct{}{}), complex64(1i))])))[(int)(((([][2]struct {
		Field5144 interface {
			Method5145(complex64) rune
		}
		Field5146 float64
	}{})[(<-make(chan int, 1))])[(int)(((([][0][0]float32{})[((((([][0][1][2]int{})[(int)((len)(([][]chan byte{})[(struct {
		Field5492 func(int, bool, int16) (rune, complex128)
		Field5493 int
	}{}).Field5493]))])[(*(([]*int{})[(Var3667)+(1)]))+(1)])[((func(uint, [0]map[float32]bool, interface{}) int)(nil))(uint(1), [0]map[float32]bool{}, interface{}(nil))])[(map[int16]int{})[int16(1)]])-(1)])[Var3667])[(<-make(chan int))])]).Field5146)])[(int)((len)((*(([]*[1]string{})[(*(Var5598))]))[(<-make(chan int, 1))]))]))[((func(uint, rune, []float32, interface{}) int)(nil))(uint(1), rune(0), []float32{}, interface{}(nil))])[(*(([]*int{})[(Var3667)+(1)]))], _ = <-(*((make([]*chan [0]complex128, (int)((<-make(chan uintptr, 1)))))[(int)(Param3119)]))
	switch uintptr(0) {
	case uintptr(0):
		for false {
			break
		}
		make(chan uintptr, 1) <- uintptr(0)
		if Var5661 := interface{}(nil); false {
			_ = Var5661
		}
		if false {
			if false {
				type Type5668 float32
				Var5669, Var5670 := <-make(chan byte)
				_ = Var5669
				_ = Var5670
			} else {
			}
		}
		fallthrough
	default:
	}
	_ = Param3117
	_ = Param3118
	_ = Param3119
	_ = Param3120
	return (*[2]*map[struct{}]int)(nil)
}((**error)(nil), error(nil), uintptr(0), uintptr(0))
var Var3305 = ((((1) + (1)) + (([]int{})[(*(((Var3372)[((([][0][2]int{})[(struct {
	Field3545 *uintptr
	Field3546 int16
	Field3547 map[rune]error
	Field3548 int
}{}).Field3548])[((((Var3570)[(<-make(chan int))])[(int)((*((([][1]*float32{})[(Var3667)+(1)])[(Var3714)[(int)(([]float64{})[(*(*(Var3715)))])]])))])[(int)((cap)((*(*(*(Var3791))))[(Var3667)+(1)]))])-(1)])[(<-make(chan int, 1))]])[(*(([]*int{})[(<-make(chan int, 1))]))]))])) + (1)) + (1)
var Var3570 = [1][2][0]int{}
var Var3667 = 1
var Var3715 = (**int)(nil)
var Var4283 = [0][0][2]*[2]map[*bool]chan string{}
var Var4560 = "foo"
var Var4741 = (**int)(nil)
var Var4743 = [1][0]int{}
var Var4880 = []func(complex64) (int16, complex64){}
var Var5063 = [1]byte{}
var Var5096 = [1]*uintptr{}
var Var5122 = (*int)(nil)
var Var6119 = [2][]*rune{}
var Var6617 = byte(0)
var Var1813 = (*int)(nil)
var Var1120 = rune(0)
var Var2409 = (*(*(((Var2766)[((func(interface{}, struct {
	Field6786 complex64
}, int16) int)(nil))(interface{}(nil), (struct {
	Field6786 complex64
}{}), int16(1))])[(*(([]*int{})[(make(map[string]int, 1))["foo"]]))+(1)])))
var Var2766 = Var2767
var Var2767 = (<-([]chan [2]func(byte, *[1]error, map[struct{}]func(float64, uint, bool) (byte, float64, byte, complex64), uint) [0][2]***float64{})[(int)(((Var2884)[(*(([]*int{})[(<-make(chan int, 1))]))+(1)])[(([]int{})[(struct {
	Field5982 int
}{}).Field5982])-(1)])])[(int)((*(*(*(([]***[0]float32{})[(((*(([]*[0][2]int{})[(<-make(chan int, 1))]))[((func([0]int16) int)(nil))([0]int16{})])[(int)((cap)((Var6119)[((*((([][2]*[1]struct {
	Field6150 uintptr
	Field6151 byte
	Field6152 struct {
		Field6153 complex64
		Field6154 int16
		Field6155 int
	}
	Field6156 *complex128
	Field6157 int
	Field6158 complex64
}{})[(Var3667)+(1)])[(*(*(*((([][2]***int{})[((func(interface{}, *chan int, []uintptr, chan uintptr) int)(nil))(interface{}(nil), (*chan int)(nil), []uintptr{}, make(chan uintptr))])[(([]int{})[(int)((len)(Var4560))])+(1)]))))]))[(int)((*(*(([]**float64{})[(int)((*(([]*struct {
	Field6461 float32
	Field6462 int16
	Field6463 rune
}{})[Var3667])).Field6462)]))))]).Field6157]))])-(1)]))))[(([]int{})[(*(*((([][0]**int{})[Var3667])[(int)(((([][2][1]uint{})[(Var3667)-(1)])[(int)(Var6617)])[(int)((len)(((*(*(*(Var6657))))[([]int{})[(Var6658)[(<-make(chan int, 1))]]])[(int)(Var6617)]))])])))])+(1)])](byte(0), (*[1]error)(nil), map[struct{}]func(float64, uint, bool) (byte, float64, byte, complex64){}, uint(1))
var Var2962 = [][1]*func(string, []chan string, complex64) [0][2]int16{}
var Var3083 = []*[2]*map[struct{}]int{0: Var3107, 9: (*[2]*map[struct{}]int)(nil)}
var Var3386 = [0]int{}
var Var3791 = (***[1][]*bool)(nil)
var Var4503 = (***string)(nil)
var Var1733 = (*int)(nil)
var Var1669 = [0]int{}
var Var1537 = (struct {
	Field1285 [2]float32
	Field1286 string
}{})
var Var343 = make(chan [][]**[]interface{})

