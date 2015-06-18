// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big_test

import (
	"fmt"
	"log"
	"math/big"
)

func ExampleRat_SetString() {
	r := new(big.Rat)
	r.SetString("355/113")
	fmt.Println(r.FloatString(3))
	// Output: 3.142
}

func ExampleInt_SetString() {
	i := new(big.Int)
	i.SetString("644", 8) // octal
	fmt.Println(i)
	// Output: 420
}

func ExampleRat_Scan() {
	// The Scan function is rarely used directly;
	// the fmt package recognizes it as an implementation of fmt.Scanner.
	r := new(big.Rat)
	_, err := fmt.Sscan("1.5000", r)
	if err != nil {
		log.Println("error scanning value:", err)
	} else {
		fmt.Println(r)
	}
	// Output: 3/2
}

func ExampleInt_Scan() {
	// The Scan function is rarely used directly;
	// the fmt package recognizes it as an implementation of fmt.Scanner.
	i := new(big.Int)
	_, err := fmt.Sscan("18446744073709551617", i)
	if err != nil {
		log.Println("error scanning value:", err)
	} else {
		fmt.Println(i)
	}
	// Output: 18446744073709551617
}

// This example demonstrates how to use big.Int to compute the smallest
// Fibonacci number with 100 decimal digits and to test whether it is prime.
func Example_fibonacci() {
	// Initialize two big ints with the first two numbers in the sequence.
	a := big.NewInt(0)
	b := big.NewInt(1)

	// Initialize limit as 10^99, the smallest integer with 100 digits.
	var limit big.Int
	limit.Exp(big.NewInt(10), big.NewInt(99), nil)

	// Loop while a is smaller than 1e100.
	for a.Cmp(&limit) < 0 {
		// Compute the next Fibonacci number, storing it in a.
		a.Add(a, b)
		// Swap a and b so that b is the next number in the sequence.
		a, b = b, a
	}
	fmt.Println(a) // 100-digit Fibonacci number

	// Test a for primality.
	// (ProbablyPrimes' argument sets the number of Miller-Rabin
	// rounds to be performed. 20 is a good value.)
	fmt.Println(a.ProbablyPrime(20))

	// Output:
	// 1344719667586153181419716641724567886890850696275767987106294472017884974410332069524504824747437757
	// false
}
