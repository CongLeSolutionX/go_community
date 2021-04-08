package genChecker42

import "bad_select_crash.dir/genUtils"

type StructF0S0 struct {
}

type ArrayF0S0E2 [2]int16

type ArrayF0S1E1 [1]StructF0S0

type StructF1S0 struct {
F0 StructF1S1
_ ArrayF1S0E4
}

type StructF1S1 struct {
}

type StructF1S2 struct {
F0 uint32
F1 uint8
F2 string
F3 string
F4 ArrayF1S1E1
}

type StructF1S3 struct {
F0 float64
}

type StructF1S4 struct {
_ int32
F1 float32
}

type StructF1S5 struct {
F0 uint16
}

type StructF1S6 struct {
F0 uint8
F1 uint32
}

type ArrayF1S0E4 [4]float64

type ArrayF1S1E1 [1]StructF1S3

type ArrayF1S2E2 [2]StructF1S4

type ArrayF1S3E2 [2]StructF1S5

type ArrayF1S4E4 [4]ArrayF1S5E3

type ArrayF1S5E3 [3]string

type ArrayF1S6E1 [1]float64

type StructF2S0 struct {
F0 ArrayF2S1E1
}

// equal func for StructF2S0
//go:noinline
func EqualStructF2S0(left StructF2S0, right StructF2S0) bool {
  return   EqualArrayF2S1E1(left.F0, right.F0)
}

type StructF2S1 struct {
_ string
F1 string
}

type ArrayF2S0E0 [0]int8

type ArrayF2S1E1 [1]*float64

// equal func for ArrayF2S1E1
//go:noinline
func EqualArrayF2S1E1(left ArrayF2S1E1, right ArrayF2S1E1) bool {
  return *left[0] == *right[0]
}

type ArrayF2S2E1 [1]StructF2S1

// 5 returns 4 params
//go:registerparams
//go:noinline
func Test2(p0 ArrayF2S0E0, p1 uint8, _ uint16, p3 float64) (r0 StructF2S0, r1 ArrayF2S2E1, r2 int16, r3 float32, r4 int64) {
  // consume some stack space, so as to trigger morestack
  var pad [16]uint64
  pad[genUtils.FailCount&0x1]++
  rc0 := StructF2S0{F0: ArrayF2S1E1{New_3(float64(-0.4418990509835844))}}
  rc1 := ArrayF2S2E1{StructF2S1{/* _: "񊶿(z̽|" */F1: "􂊇񊶿"}}
  rc2 := int16(4162)
  rc3 := float32(-7.667096e+37)
  rc4 := int64(3202175648847048679)
  p1f0c := uint8(57)
  if p1 != p1f0c {
    genUtils.NoteFailureElem(9, 42, 2, "genChecker42", "parm", 1, 0, false, pad[0])
    return
  }
  _ = uint16(10920)
  p3f0c := float64(-1.597256501942112)
  if p3 != p3f0c {
    genUtils.NoteFailureElem(9, 42, 2, "genChecker42", "parm", 3, 0, false, pad[0])
    return
  }
  defer func(p0 ArrayF2S0E0, p1 uint8) {
  // check parm passed
  // check parm passed
  if p1 != p1f0c {
    genUtils.NoteFailureElem(9, 42, 2, "genChecker42", "parm", 1, 0, false, pad[0])
    return
  }
  // check parm captured
  if p3 != p3f0c {
    genUtils.NoteFailureElem(9, 42, 2, "genChecker42", "parm", 3, 0, false, pad[0])
    return
  }
  } (p0, p1)

  return rc0, rc1, rc2, rc3, rc4
  // 0 addr-taken params, 0 addr-taken returns
}


//go:noinline
func New_3(i float64)  *float64 {
  x := new( float64)
  *x = i
  return x
}


type StructF3S0 struct {
F0 uint8
F1 *uint8
F2 uint16
}

type ArrayF3S0E2 [2]string

type ArrayF3S1E2 [2]complex64

type StructF4S0 struct {
}

type ArrayF4S0E0 [0]float32

type ArrayF4S1E4 [4]StructF4S0

type ArrayF4S2E0 [0]float32

type ArrayF4S3E1 [1]float64

type ArrayF4S4E4 [4]complex64

type ArrayF4S5E1 [1]uint16

var _ *uint16

var _ *ArrayF4S5E1

type StructF5S0 struct {
F0 int64
F1 byte
F2 uint16
F3 *uint16
F4 StructF5S1
}

type StructF5S1 struct {
_ uint32
F1 StructF5S2
}

type StructF5S2 struct {
}

type StructF5S3 struct {
F0 int8
F1 float32
F2 StructF5S4
}

type StructF5S4 struct {
}

type StructF5S5 struct {
F0 uint32
F1 uint16
F2 ArrayF5S0E1
_ int32
}

type StructF5S6 struct {
}

type StructF5S7 struct {
F0 uint8
}

type StructF5S8 struct {
F0 uint8
F1 StructF5S9
F2 StructF5S10
_ complex128
F4 *int8
F5 StructF5S11
}

type StructF5S9 struct {
F0 int16
}

type StructF5S10 struct {
F0 complex64
}

type StructF5S11 struct {
F0 uint64
}

type ArrayF5S0E1 [1]StructF5S6

type ArrayF5S1E4 [4]uint64

type ArrayF5S2E2 [2]uint8

type StructF6S0 struct {
}

type StructF6S1 struct {
_ byte
}

type StructF6S2 struct {
F0 StructF6S3
}

type StructF6S3 struct {
F0 int8
F1 uint64
}

type ArrayF6S0E3 [3]float64

type ArrayF6S1E0 [0]StructF6S1

type StructF7S0 struct {
_ float64
}

type StructF7S1 struct {
F0 StructF7S2
F1 uint64
F2 uint16
F3 complex64
}

type StructF7S2 struct {
F0 int32
F1 string
}

type StructF7S3 struct {
_ ArrayF7S0E0
F1 float32
F2 StructF7S4
}

type StructF7S4 struct {
}

type StructF7S5 struct {
F0 complex64
_ ArrayF7S1E1
F2 StructF7S6
}

type StructF7S6 struct {
F0 uint64
}

type ArrayF7S0E0 [0]uint8

type ArrayF7S1E1 [1]int32

type StructF8S0 struct {
}

type ArrayF8S0E2 [2]string

type ArrayF8S1E0 [0]uint64

type ArrayF8S2E2 [2]uint8

type StructF9S0 struct {
}

type ArrayF9S0E4 [4]**string

type ArrayF9S1E0 [0]float64

type StructF10S0 struct {
_ ArrayF10S1E1
F1 ArrayF10S2E3
F2 uint8
F3 ArrayF10S3E4
F4 int32
F5 uint8
}

type StructF10S1 struct {
}

type ArrayF10S0E0 [0]float32

type ArrayF10S1E1 [1]int8

type ArrayF10S2E3 [3]string

type ArrayF10S3E4 [4]ArrayF10S4E0

type ArrayF10S4E0 [0]uint16

type StructF11S0 struct {
F0 int16
F1 int8
F2 float64
F3 float32
F4 ArrayF11S0E3
}

type StructF11S1 struct {
F0 StructF11S2
}

type StructF11S2 struct {
F0 uint16
_ int64
}

type StructF11S3 struct {
F0 int64
F1 StructF11S4
F2 StructF11S5
_ StructF11S6
F4 *uint64
}

type StructF11S4 struct {
F0 string
}

type StructF11S5 struct {
F0 uint64
}

type StructF11S6 struct {
F0 string
}

type StructF11S7 struct {
F0 ArrayF11S2E0
}

type StructF11S8 struct {
F0 uint16
F1 int16
F2 ArrayF11S4E3
F3 uint8
F4 string
}

type StructF11S9 struct {
_ float32
}

type StructF11S10 struct {
F0 uint64
F1 float32
F2 byte
F3 ArrayF11S7E2
F4 complex128
}

type StructF11S11 struct {
}

type ArrayF11S0E3 [3]ArrayF11S1E0

type ArrayF11S1E0 [0]float64

type ArrayF11S2E0 [0]float64

type ArrayF11S3E4 [4]float32

type ArrayF11S4E3 [3]ArrayF11S5E1

type ArrayF11S5E1 [1]float32

type ArrayF11S6E3 [3]float32

type ArrayF11S7E2 [2]StructF11S11

type StructF12S0 struct {
F0 byte
F1 int32
_ string
F3 StructF12S1
}

type StructF12S1 struct {
_ *int8
}

type StructF13S0 struct {
F0 complex64
F1 *float32
_ StructF13S1
}

type StructF13S1 struct {
_ ArrayF13S0E4
}

type StructF13S2 struct {
F0 string
}

type ArrayF13S0E4 [4]uint16

type ArrayF13S1E4 [4]StructF13S2

type ArrayF13S2E4 [4]ArrayF13S3E0

type ArrayF13S3E0 [0]float32

type StructF14S0 struct {
F0 StructF14S1
F1 uint8
}

type StructF14S1 struct {
F0 byte
_ StructF14S2
}

type StructF14S2 struct {
}

type StructF14S3 struct {
F0 uint16
F1 StructF14S4
}

type StructF14S4 struct {
_ ArrayF14S0E3
}

type StructF14S5 struct {
_ byte
F1 ArrayF14S4E3
}

type StructF14S6 struct {
F0 string
}

type ArrayF14S0E3 [3]float32

type ArrayF14S1E3 [3]byte

type ArrayF14S2E4 [4]complex128

type ArrayF14S3E3 [3]string

type ArrayF14S4E3 [3]StructF14S6

type StructF15S0 struct {
F0 uint64
F1 ArrayF15S0E2
}

type StructF15S1 struct {
F0 float64
}

type StructF15S2 struct {
F0 StructF15S3
_ StructF15S4
}

type StructF15S3 struct {
F0 ArrayF15S2E0
}

type StructF15S4 struct {
F0 ArrayF15S3E4
F1 string
}

type StructF15S5 struct {
_ int8
F1 string
F2 uint16
}

type StructF15S6 struct {
F0 StructF15S7
F1 float64
F2 string
F3 *uint32
F4 StructF15S8
}

type StructF15S7 struct {
}

type StructF15S8 struct {
}

type ArrayF15S0E2 [2]StructF15S1

type ArrayF15S1E1 [1]byte

type ArrayF15S2E0 [0]float32

type ArrayF15S3E4 [4]string

type ArrayF15S4E2 [2]int64


type MyTypeF16S0 complex64

// 1 returns 1 params
//go:noinline
func (rcvr MyTypeF16S0) Test16(p0 int8) (r0 byte) {
  // consume some stack space, so as to trigger morestack
  var pad [4]uint64
  pad[genUtils.FailCount&0x1]++
  rc0 := byte(41)
  p0f0c := int8(9)
  if p0 != p0f0c {
    genUtils.NoteFailureElem(3, 42, 16, "genChecker42", "parm", 0, 0, false, pad[0])
    return
  }
  rcvrf0c := MyTypeF16S0(complex(float32(-2.66459e+37),float32(6.0335833e+37)))
  if rcvr != rcvrf0c {
    genUtils.NoteFailureElem(3, 42, 16, "genChecker42", "rcvr", 0, -1, false, pad[0])
    return
  }
  return rc0
  // 0 addr-taken params, 0 addr-taken returns
}


type StructF17S0 struct {
F0 ArrayF17S1E0
F1 int32
F2 uint32
F3 uint64
F4 StructF17S1
F5 ArrayF17S2E3
}

type StructF17S1 struct {
F0 StructF17S2
F1 StructF17S3
}

type StructF17S2 struct {
}

type StructF17S3 struct {
_ float32
}

type StructF17S4 struct {
F0 ArrayF17S4E1
F1 StructF17S5
F2 uint8
}

type StructF17S5 struct {
F0 string
}

type StructF17S6 struct {
F0 ArrayF17S6E4
}

type StructF17S7 struct {
F0 int16
F1 StructF17S8
F2 float64
}

type StructF17S8 struct {
F0 ArrayF17S7E3
}

type StructF17S9 struct {
F0 float64
F1 complex128
F2 uint64
F3 string
F4 uint8
}

type StructF17S10 struct {
}

type StructF17S11 struct {
F0 StructF17S12
}

type StructF17S12 struct {
}

type StructF17S13 struct {
F0 uint8
F1 ArrayF17S8E1
F2 uint16
F3 byte
F4 byte
}

type StructF17S14 struct {
F0 ArrayF17S9E3
_ ArrayF17S10E2
}

type StructF17S15 struct {
}

type StructF17S16 struct {
}

type StructF17S17 struct {
F0 ArrayF17S15E2
F1 byte
_ ArrayF17S16E0
F3 *float64
F4 StructF17S19
F5 byte
}

type StructF17S18 struct {
}

type StructF17S19 struct {
F0 string
}

type StructF17S20 struct {
}

type StructF17S21 struct {
F0 ArrayF17S17E0
F1 int8
_ StructF17S22
F3 string
F4 ArrayF17S18E3
}

type StructF17S22 struct {
F0 uint16
F1 float64
}

type ArrayF17S0E1 [1]int8

type ArrayF17S1E0 [0]int8

type ArrayF17S2E3 [3]ArrayF17S3E4

type ArrayF17S3E4 [4]float32

type ArrayF17S4E1 [1]ArrayF17S5E4

type ArrayF17S5E4 [4]string

type ArrayF17S6E4 [4]complex128

type ArrayF17S7E3 [3]*complex128

type ArrayF17S8E1 [1]byte

type ArrayF17S9E3 [3]uint8

type ArrayF17S10E2 [2]uint32

type ArrayF17S11E1 [1]ArrayF17S12E4

type ArrayF17S12E4 [4]StructF17S15

type ArrayF17S13E4 [4]StructF17S16

type ArrayF17S14E1 [1]uint32

type ArrayF17S15E2 [2]float64

type ArrayF17S16E0 [0]StructF17S18

type ArrayF17S17E0 [0]complex64

type ArrayF17S18E3 [3]complex128

type StructF18S0 struct {
F0 StructF18S1
F1 ArrayF18S0E0
F2 ArrayF18S1E3
F3 ArrayF18S2E0
_ StructF18S3
}

type StructF18S1 struct {
F0 float32
}

type StructF18S2 struct {
F0 float64
}

type StructF18S3 struct {
F0 StructF18S4
F1 StructF18S5
}

type StructF18S4 struct {
F0 float64
}

type StructF18S5 struct {
F0 float64
}

type StructF18S6 struct {
F0 ArrayF18S4E1
}

type StructF18S7 struct {
F0 *StructF18S8
}

type StructF18S8 struct {
F0 uint8
}

type ArrayF18S0E0 [0]uint16

type ArrayF18S1E3 [3]StructF18S2

type ArrayF18S2E0 [0]float32

type ArrayF18S3E3 [3]float32

type ArrayF18S4E1 [1]uint16

type ArrayF18S5E4 [4]string

type ArrayF18S6E2 [2]string

type StructF19S0 struct {
}

type StructF19S1 struct {
F0 complex128
}

type ArrayF19S0E4 [4]complex64

var _ **float32

var _ *ArrayF19S0E4

// dummy
var _ genUtils.UtilsType
