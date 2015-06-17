// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big_test

import (
	"fmt"
	"math/big"
)

// ExampleFibonacci demonstrates how to use big.Int to compute the smallest
// Fibonacci number with 1000 decimal digits, and find out whether it is prime.
func Example_fibonacci() {
	// create and initialize big.Ints from int64s
	fib1 := big.NewInt(1)
	fib2 := big.NewInt(2)

	// initialize limit as 10^999 (the smallest integer
	// with 1000 digits) using the Exp function
	var limit big.Int
	limit.Exp(big.NewInt(10), big.NewInt(999), nil)

	// loop until fib1 is smaller than 1e1000
	for fib1.Cmp(&limit) < 0 {
		fib1, fib2 = fib2, fib1.Add(fib1, fib2)
	}

	fmt.Println(fib1) // 1000-digits fibonacci number

	// test fib1 for primality. The ProbablyPrimes parameter sets the number
	// of Miller-Rabin rounds to be performed. 20 is a good value
	isPrime := fib1.ProbablyPrime(20)
	fmt.Println(isPrime) // false
}
