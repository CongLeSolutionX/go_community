// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// Check compatibility with IEEE-754 arithmetic when called on a Float
// with a 53-bit mantissa
func TestSqrtf53(t *testing.T) {
	for i := 0; i < 1e5; i++ {
		r := rand.Float64()
		z := NewFloat(0) // prec defaults to 53

		got, _ := z.Sqrt(NewFloat(r)).Float64()
		want := math.Sqrt(r)
		if got != want {
			t.Fatalf("Sqrt(%g) =\n got %g;\nwant %g", z, got, want)
		}
	}
}

func TestSqrtf(t *testing.T) {
	for _, test := range []struct {
		x    string
		want string
	}{
		// Test values were generated on Wolfram Alpha using query
		//   'sqrt(N) to 350 digits'
		// 350 decimal digits give up to 1000 binary digits.
		{"0.5", "0.70710678118654752440084436210484903928483593768847403658833986899536623923105351942519376716382078636750692311545614851246241802792536860632206074854996791570661133296375279637789997525057639103028573505477998580298513726729843100736425870932044459930477616461524215435716072541988130181399762570399484362669827316590441482031030762917619752737287514"},
		{"2.0", "1.4142135623730950488016887242096980785696718753769480731766797379907324784621070388503875343276415727350138462309122970249248360558507372126441214970999358314132226659275055927557999505011527820605714701095599716059702745345968620147285174186408891986095523292304843087143214508397626036279952514079896872533965463318088296406206152583523950547457503"},
		{"3.0", "1.7320508075688772935274463415058723669428052538103806280558069794519330169088000370811461867572485756756261414154067030299699450949989524788116555120943736485280932319023055820679748201010846749232650153123432669033228866506722546689218379712270471316603678615880190499865373798593894676503475065760507566183481296061009476021871903250831458295239598"},
		{"4.0", "2.0"},

		{"1p512", "1p256"},
		{"4p1024", "2p512"},
		{"9p2048", "3p1024"},

		{"1p-1024", "1p-512"},
		{"4p-2048", "2p-1024"},
		{"9p-4096", "3p-2048"},
	} {
		for _, prec := range []uint{24, 53, 64, 100, 128, 129, 200, 256, 400, 600, 800, 1000} {
			want := new(Float).SetPrec(prec)
			want.Parse(test.want, 10)
			x := new(Float).SetPrec(prec)
			x.Parse(test.x, 10)

			z := new(Float).SetPrec(prec).Sqrt(x)
			if z.Cmp(want) != 0 {
				t.Errorf("prec = %d, Sqrt(%v) =\ngot  %g;\nwant %g",
					prec, test.x, z, want)
			}
		}
	}
}

func TestSqrtfSpecial(t *testing.T) {
	for _, test := range []struct {
		x    *Float
		want *Float
	}{
		{NewFloat(+0), NewFloat(+0)},
		{NewFloat(-0), NewFloat(-0)},
		{NewFloat(math.Inf(+1)), NewFloat(math.Inf(+1))},
	} {
		z := new(Float).Sqrt(test.x)
		if z.neg != test.want.neg || z.form != test.want.form {
			t.Errorf("Sqrt(%v) = %v (neg: %v); want %v (neg: %v)",
				test.x, z, z.neg, test.want, test.want.neg)
		}
	}

}

// Benchmarks

func BenchmarkSqrtf(b *testing.B) {
	for _, prec := range []uint{64, 128, 1e3, 1e4, 1e5, 1e6} {
		x := NewFloat(2)
		z := new(Float).SetPrec(prec)
		b.Run(fmt.Sprintf("%v", prec), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				z.Sqrt(x)
			}
		})
	}
}
