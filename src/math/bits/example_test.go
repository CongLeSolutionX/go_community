// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bits_test

import (
	"fmt"
	"math/bits"
)

func ExampleLeadingZeros16() {
	fmt.Println(bits.LeadingZeros16(0))
	fmt.Println(bits.LeadingZeros16(1))
	fmt.Println(bits.LeadingZeros16(256))
	fmt.Println(bits.LeadingZeros16(65535))
	// Output:
	// 16
	// 15
	// 7
	// 0
}

func ExampleLeadingZeros32() {
	fmt.Println(bits.LeadingZeros32(0))
	fmt.Println(bits.LeadingZeros32(1))
	// Output:
	// 32
	// 31
}

func ExampleLeadingZeros64() {
	fmt.Println(bits.LeadingZeros64(0))
	fmt.Println(bits.LeadingZeros64(1))
	// Output:
	// 64
	// 63
}

func ExampleOnesCount() {
	fmt.Printf("%b\n", 14)
	fmt.Println(bits.OnesCount(14))
	// Output:
	// 1110
	// 3
}

func ExampleOnesCount8() {
	fmt.Printf("%b\n", 14)
	fmt.Println(bits.OnesCount8(14))
	// Output:
	// 1110
	// 3
}

func ExampleOnesCount16() {
	fmt.Printf("%b\n", 14)
	fmt.Println(bits.OnesCount16(14))
	// Output:
	// 1110
	// 3
}

func ExampleOnesCount32() {
	fmt.Printf("%b\n", 14)
	fmt.Println(bits.OnesCount32(14))
	// Output:
	// 1110
	// 3
}

func ExampleOnesCount64() {
	fmt.Printf("%b\n", 14)
	fmt.Println(bits.OnesCount64(14))
	// Output:
	// 1110
	// 3
}

func ExampleReverse() {
	fmt.Printf("%b\n", 7)
	fmt.Println(bits.Reverse(7))
	// Output:
	// 111
	// 16140901064495857664
}

func ExampleReverse16() {
	fmt.Printf("%b\n", 7)
	fmt.Println(bits.Reverse16(7))
	// Output:
	// 111
	// 57344
}

func ExampleReverse32() {
	fmt.Printf("%b\n", 7)
	fmt.Println(bits.Reverse32(7))
	// Output:
	// 111
	// 3758096384
}

func ExampleReverse64() {
	fmt.Printf("%b\n", 7)
	fmt.Println(bits.Reverse64(7))
	// Output:
	// 111
	// 16140901064495857664
}

func ExampleReverse8() {
	fmt.Printf("%b\n", 7)
	fmt.Println(bits.Reverse8(7))
	// Output:
	// 111
	// 224
}
