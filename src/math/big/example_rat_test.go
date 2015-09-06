package big_test

import (
	"fmt"
	"math/big"
)

// eConvergent returns the nth convergent of e.
func eConvergent(n int) *big.Rat {
	return recur(0, int64(n))
}

// Use the classic continued fraction for e
//     e = [1; 0, 1, 1, 2, 1, 1, ... 2n, 1, 1, ...]
// i.e., for the nth term, use
//     1       if    n != 1 (mod 3)
//  2(n-1)/3   if    n == 1 (mod 3)
func recur(n, lim int64) *big.Rat {
	var res *big.Rat
	if n%3 == 1 {
		// Initialize a big.Rat from a couple of int64s
		// (respectively numerator and denominator).
		res = big.NewRat(n/3*2, 1)
	} else {
		res = big.NewRat(1, 1)
	}

	if n > lim {
		return res
	}

	// Directly initialize r as the fractional
	// inverse of the result of recur.
	r := new(big.Rat).Inv(recur(n+1, lim))

	// Add two big.Rat together. The result is
	// guaranteed to be reduced to lowest terms.
	return res.Add(res, r)
}

// This example demonstrates how to use big.Rat to compute and display
// the first 15 terms in the sequence of rational convergents for the
// e constant.
func Example_eConvergents() {
	for i := 1; i <= 15; i++ {
		r := eConvergent(i)

		// Print the big.Rat result in fraction form.
		// We can use the usual fmt.Printf verbs since
		// big.Rat implements fmt.Formatter.
		fmt.Printf("%-13s = ", r)

		// Print a string representation of r in decimal
		// form with 8 digits of precision.
		fmt.Println(r.FloatString(8))
	}

	// Output:
	// 2/1           = 2.00000000
	// 3/1           = 3.00000000
	// 8/3           = 2.66666667
	// 11/4          = 2.75000000
	// 19/7          = 2.71428571
	// 87/32         = 2.71875000
	// 106/39        = 2.71794872
	// 193/71        = 2.71830986
	// 1264/465      = 2.71827957
	// 1457/536      = 2.71828358
	// 2721/1001     = 2.71828172
	// 23225/8544    = 2.71828184
	// 25946/9545    = 2.71828182
	// 49171/18089   = 2.71828183
	// 517656/190435 = 2.71828183
}
