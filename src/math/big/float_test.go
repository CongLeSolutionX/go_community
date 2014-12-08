// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"strconv"
	"testing"
)

func fromBinary(s string) int64 {
	x, err := strconv.ParseInt(s, 2, 64)
	if err != nil {
		panic(err)
	}
	return x
}

func toBinary(x int64) string {
	return strconv.FormatInt(x, 2)
}

func testFloatRound(t *testing.T, x, r int64, prec uint, mode RoundingMode) {
	// verify test data
	var ok bool
	switch mode {
	case ToNearest:
		ok = true // nothing to do for now
	case ToZero:
		if x < 0 {
			ok = r >= x
		} else {
			ok = r <= x
		}
	case AwayFromZero:
		if x < 0 {
			ok = r <= x
		} else {
			ok = r >= x
		}
	case ToNegInf:
		ok = r <= x
	case ToPosInf:
		ok = r >= x
	default:
		panic("unreachable")
	}
	if !ok {
		t.Fatalf("incorrect test data for prec = %d, %s: x = %s, r = %s", prec, mode, toBinary(x), toBinary(r))
	}

	// compute expected accuracy
	a := exact
	switch {
	case r < x:
		a = below
	case r > x:
		a = above
	}

	// round
	f := new(Float).SetInt64(x)
	f.Round(f, prec, mode)

	// check result
	r1 := f.Int64()
	p1 := f.Precision()
	a1 := f.Accuracy()
	if r1 != r || p1 != prec || a1 != a {
		t.Errorf("Round(%s, %d, %s): got %s (%d bits, %s); want %s (%d bits, %s)",
			toBinary(x), prec, mode,
			toBinary(r1), p1, a1,
			toBinary(r), prec, a)
	}
}

// TestFloatRound tests basic rounding.
func TestFloatRound(t *testing.T) {
	var tests = []struct {
		prec                   uint
		x, zero, nearest, away string // input, results rounded to prec bits
	}{
		{5, "1000", "1000", "1000", "1000"},
		{5, "1001", "1001", "1001", "1001"},
		{5, "1010", "1010", "1010", "1010"},
		{5, "1011", "1011", "1011", "1011"},
		{5, "1100", "1100", "1100", "1100"},
		{5, "1101", "1101", "1101", "1101"},
		{5, "1110", "1110", "1110", "1110"},
		{5, "1111", "1111", "1111", "1111"},

		{4, "1000", "1000", "1000", "1000"},
		{4, "1001", "1001", "1001", "1001"},
		{4, "1010", "1010", "1010", "1010"},
		{4, "1011", "1011", "1011", "1011"},
		{4, "1100", "1100", "1100", "1100"},
		{4, "1101", "1101", "1101", "1101"},
		{4, "1110", "1110", "1110", "1110"},
		{4, "1111", "1111", "1111", "1111"},

		{3, "1000", "1000", "1000", "1000"},
		{3, "1001", "1000", "1000", "1010"},
		{3, "1010", "1010", "1010", "1010"},
		{3, "1011", "1010", "1100", "1100"},
		{3, "1100", "1100", "1100", "1100"},
		{3, "1101", "1100", "1100", "1110"},
		{3, "1110", "1110", "1110", "1110"},
		{3, "1111", "1110", "10000", "10000"},

		{3, "1000001", "1000000", "1000000", "1010000"},
		{3, "1001001", "1000000", "1010000", "1010000"},
		{3, "1010001", "1010000", "1010000", "1100000"},
		{3, "1011001", "1010000", "1100000", "1100000"},
		{3, "1100001", "1100000", "1100000", "1110000"},
		{3, "1101001", "1100000", "1110000", "1110000"},
		{3, "1110001", "1110000", "1110000", "10000000"},
		{3, "1111001", "1110000", "10000000", "10000000"},

		{2, "1000", "1000", "1000", "1000"},
		{2, "1001", "1000", "1000", "1100"},
		{2, "1010", "1000", "1000", "1100"},
		{2, "1011", "1000", "1100", "1100"},
		{2, "1100", "1100", "1100", "1100"},
		{2, "1101", "1100", "1100", "10000"},
		{2, "1110", "1100", "10000", "10000"},
		{2, "1111", "1100", "10000", "10000"},

		{2, "1000001", "1000000", "1000000", "1100000"},
		{2, "1001001", "1000000", "1000000", "1100000"},
		{2, "1010001", "1000000", "1100000", "1100000"},
		{2, "1011001", "1000000", "1100000", "1100000"},
		{2, "1100001", "1100000", "1100000", "10000000"},
		{2, "1101001", "1100000", "1100000", "10000000"},
		{2, "1110001", "1100000", "10000000", "10000000"},
		{2, "1111001", "1100000", "10000000", "10000000"},

		{1, "1000", "1000", "1000", "1000"},
		{1, "1001", "1000", "1000", "10000"},
		{1, "1010", "1000", "1000", "10000"},
		{1, "1011", "1000", "1000", "10000"},
		{1, "1100", "1000", "10000", "10000"},
		{1, "1101", "1000", "10000", "10000"},
		{1, "1110", "1000", "10000", "10000"},
		{1, "1111", "1000", "10000", "10000"},

		{1, "1000001", "1000000", "1000000", "10000000"},
		{1, "1001001", "1000000", "1000000", "10000000"},
		{1, "1010001", "1000000", "1000000", "10000000"},
		{1, "1011001", "1000000", "1000000", "10000000"},
		{1, "1100001", "1000000", "10000000", "10000000"},
		{1, "1101001", "1000000", "10000000", "10000000"},
		{1, "1110001", "1000000", "10000000", "10000000"},
		{1, "1111001", "1000000", "10000000", "10000000"},
	}

	for _, test := range tests {
		x := fromBinary(test.x)
		z := fromBinary(test.zero)
		n := fromBinary(test.nearest)
		a := fromBinary(test.away)
		prec := test.prec

		testFloatRound(t, x, z, prec, ToZero)
		testFloatRound(t, x, n, prec, ToNearest)
		testFloatRound(t, x, a, prec, AwayFromZero)

		testFloatRound(t, x, z, prec, ToNegInf)
		testFloatRound(t, x, a, prec, ToPosInf)

		testFloatRound(t, -x, -a, prec, ToNegInf)
		testFloatRound(t, -x, -z, prec, ToPosInf)
	}
}

func TestFloatNewFloat(t *testing.T) {
	// TODO(gri) implement this
}

// TestFloatRound24 tests that rounding a float64 to 24 bits
// matches IEEE-754 rounding to nearest when converting a
// float64 to a float32.
func TestFloatRound24(t *testing.T) {
	const x0 = 1<<26 - 0x10 // 11...110000 (26 bits)
	for d := 0; d <= 0x10; d++ {
		x := float64(x0 + d)
		f := new(Float).SetFloat64(x)
		f.Round(f, 24, ToNearest)
		got, _ := f.Float64()
		want := float64(float32(x))
		if got != want {
			t.Errorf("Round(%g, 24) = %g; want %g", x, got, want)
		}
	}
}

func TestFloatSetUint64(t *testing.T) {
	var tests = []uint64{
		0,
		1,
		2,
		10,
		100,
		1<<32 - 1,
		1 << 32,
		1<<64 - 1,
	}
	for _, want := range tests {
		f := new(Float).SetUint64(want)
		if got := f.Uint64(); got != want {
			t.Errorf("got %d (%s); want %d", got, f.PString(), want)
		}
	}
}

func TestFloatSetInt64(t *testing.T) {
	var tests = []int64{
		0,
		1,
		2,
		10,
		100,
		1<<32 - 1,
		1 << 32,
		1<<63 - 1,
	}
	for _, want := range tests {
		for i := range [2]int{} {
			if i&1 != 0 {
				want = -want
			}
			f := new(Float).SetInt64(want)
			if got := f.Int64(); got != want {
				t.Errorf("got %d (%s); want %d", got, f.PString(), want)
			}
		}
	}
}

func TestFloatSetFloat64(t *testing.T) {
	var tests = []float64{
		0,
		1,
		2,
		12345,
		1e10,
		1e100,
		3.14159265e10,
		2.718281828e-123,
		1.0 / 3,
	}
	for _, want := range tests {
		for i := range [2]int{} {
			if i&1 != 0 {
				want = -want
			}
			f := new(Float).SetFloat64(want)
			if got, _ := f.Float64(); got != want {
				t.Errorf("got %g (%s); want %g", got, f.PString(), want)
			}
		}
	}
}

func TestFloatSetInt(t *testing.T) {
}

// TestFloatAdd32 tests that Float.Add/Sub of numbers with
// 24bit mantissa behaves like float32 addition/subtraction.
func TestFloatAdd32(t *testing.T) {
	// chose base such that we cross the mantissa precision limit
	const base = 1<<26 - 0x10 // 11...110000 (26 bits)
	for d := 0; d <= 0x10; d++ {
		for i := range [2]int{} {
			x0, y0 := float64(base), float64(d)
			if i&1 != 0 {
				x0, y0 = y0, x0
			}

			x := new(Float).SetFloat64(x0)
			y := new(Float).SetFloat64(y0)
			var z Float
			z.prec = 24 // TODO(gri) fix this

			z.Add(x, y, ToNearest)
			got, acc := z.Float64()
			want := float64(float32(y0) + float32(x0))
			if got != want || acc != exact {
				t.Errorf("d = %d: %g + %g = %g (%s); want %g exactly", d, x0, y0, got, acc, want)
			}

			z.Sub(&z, y, ToNearest)
			got, acc = z.Float64()
			want = float64(float32(want) - float32(y0))
			if got != want || acc != exact {
				t.Errorf("d = %d: %g - %g = %g (%s); want %g exactly", d, x0+y0, y0, got, acc, want)
			}
		}
	}
}

// TestFloatAdd64 tests that Float.Add/Sub of numbers with
// 53bit mantissa behaves like float64 addition/subtraction.
func TestFloatAdd64(t *testing.T) {
	// chose base such that we cross the mantissa precision limit
	const base = 1<<55 - 0x10 // 11...110000 (55 bits)
	for d := 0; d <= 0x10; d++ {
		for i := range [2]int{} {
			x0, y0 := float64(base), float64(d)
			if i&1 != 0 {
				x0, y0 = y0, x0
			}

			x := new(Float).SetFloat64(x0)
			y := new(Float).SetFloat64(y0)
			var z Float
			z.prec = 53 // TODO(gri) fix this

			z.Add(x, y, ToNearest)
			got, acc := z.Float64()
			want := x0 + y0
			if got != want || acc != exact {
				t.Errorf("d = %d: %g + %g = %g; want %g exactly", d, x0, y0, got, acc, want)
			}

			z.Sub(&z, y, ToNearest)
			got, acc = z.Float64()
			want -= y0
			if got != want || acc != exact {
				t.Errorf("d = %d: %g - %g = %g; want %g exactly", d, x0+y0, y0, got, acc, want)
			}
		}
	}
}

// TestFloatMul64 tests that Float.Mul/Quo of numbers with
// 53bit mantissa behaves like float64 multiplication/division.
func TestFloatMul64(t *testing.T) {
	var tests = []struct {
		x, y float64
	}{
		{0, 0},
		{0, 1},
		{1, 1},
		{1, 1.5},
		{1.234, 0.5678},
		{2.718281828, 3.14159265358979},
		{2.718281828e10, 3.14159265358979e-32},
		{1.0 / 3, 1e200},
	}
	for _, test := range tests {
		for i := range [8]int{} {
			x0, y0 := test.x, test.y
			if i&1 != 0 {
				x0 = -x0
			}
			if i&2 != 0 {
				y0 = -y0
			}
			if i&4 != 0 {
				x0, y0 = y0, x0
			}

			x := new(Float).SetFloat64(x0)
			y := new(Float).SetFloat64(y0)
			var z Float
			z.prec = 53 // TODO(gri) fix this

			z.Mul(x, y, ToNearest)
			got, _ := z.Float64()
			want := x0 * y0
			if got != want {
				t.Errorf("%g * %g = %g; want %g", x0, y0, got, want)
			}

			if y0 == 0 {
				continue // avoid division-by-zero
			}
			z.Quo(&z, y, ToNearest)
			got, _ = z.Float64()
			want /= y0
			if got != want {
				t.Errorf("%g / %g = %g; want %g", x0*y0, y0, got, want)
			}
		}
	}
}

func TestFloatQuo(t *testing.T) {
}

var floatSetStringTests = []struct {
	x, want string
}{
	{"0", "0.0p0"},
	{"-0", "-0.0p0"},
	{"1", "0.1p1"},
	{"-1", "-0.1p1"},
	{"0x123456789", "0.123456789p33"},
}

func TestFloatSetString(t *testing.T) {
	// for _, test := range floatSetStringTests {
	// 	var x Float
	// 	_, ok := x.SetString(test.x)
	// 	if !ok {
	// 		t.Errorf("%s: parse error", test.x)
	// 		continue
	// 	}
	// 	if got := x.PString(); got != test.want {
	// 		t.Errorf("%s: got %s; want %s", test.x, got, test.want)
	// 	}
	// }
}

func TestFloatPString(t *testing.T) {
	var tests = []struct {
		x    Float
		want string
	}{
		{Float{}, "0.0p0"},
		{Float{neg: true}, "-0.0p0"},
		{Float{mant: nat{0x87654321}}, "0.87654321p0"},
		{Float{mant: nat{0x87654321}, exp: -10}, "0.87654321p-10"},
	}
	for _, test := range tests {
		if got := test.x.PString(); got != test.want {
			t.Errorf("%v: got %s; want %s", test.x, got, test.want)
		}
	}
}
