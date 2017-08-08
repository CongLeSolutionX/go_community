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

func ExampleTrailingZeros() {
	fmt.Printf("%b\n", 0)
	fmt.Println(bits.TrailingZeros(0))
	fmt.Printf("%b\n", 1)
	fmt.Println(bits.TrailingZeros(1))
	fmt.Printf("%b\n", 256)
	fmt.Println(bits.TrailingZeros(256))
	fmt.Printf("%b\n", 65535)
	fmt.Println(bits.TrailingZeros(65535))
	// Output:
	// 0
	// 64
	// 1
	// 0
	// 100000000
	// 8
	// 1111111111111111
	// 0
}

func ExampleTrailingZeros8() {
	fmt.Printf("%b\n", 0)
	fmt.Println(bits.TrailingZeros8(0))
	fmt.Printf("%b\n", 254)
	fmt.Println(bits.TrailingZeros8(254))
	// Output:
	// 0
	// 8
	// 11111110
	// 1
}

func ExampleTrailingZeros16() {
	fmt.Printf("%b\n", 0)
	fmt.Println(bits.TrailingZeros16(0))
	fmt.Printf("%b\n", 254)
	fmt.Println(bits.TrailingZeros16(254))
	fmt.Printf("%b\n", 255)
	fmt.Println(bits.TrailingZeros16(255))
	fmt.Printf("%b\n", 65534)
	fmt.Println(bits.TrailingZeros16(65534))
	fmt.Printf("%b\n", 65535)
	fmt.Println(bits.TrailingZeros16(65535))
	// Output:
	// 0
	// 16
	// 11111110
	// 1
	// 11111111
	// 0
	// 1111111111111110
	// 1
	// 1111111111111111
	// 0
}

func ExampleTrailingZeros32() {
	fmt.Printf("%b\n", 0)
	fmt.Println(bits.TrailingZeros32(0))
	fmt.Printf("%b\n", 254)
	fmt.Println(bits.TrailingZeros32(254))
	fmt.Printf("%b\n", 255)
	fmt.Println(bits.TrailingZeros32(255))
	fmt.Printf("%b\n", 65534)
	fmt.Println(bits.TrailingZeros32(65534))
	fmt.Printf("%b\n", 65535)
	fmt.Println(bits.TrailingZeros32(65535))
	fmt.Printf("%b\n", 2147483648)
	fmt.Println(bits.TrailingZeros32(2147483648))
	fmt.Printf("%b\n", 4294967295)
	fmt.Println(bits.TrailingZeros32(4294967295))
	// Output:
	// 0
	// 32
	// 11111110
	// 1
	// 11111111
	// 0
	// 1111111111111110
	// 1
	// 1111111111111111
	// 0
	// 10000000000000000000000000000000
	// 31
	// 11111111111111111111111111111111
	// 0
}

func ExampleTrailingZeros64() {
	fmt.Printf("%b\n", 0)
	fmt.Println(bits.TrailingZeros64(0))
	fmt.Printf("%b\n", 254)
	fmt.Println(bits.TrailingZeros64(254))
	fmt.Printf("%b\n", 255)
	fmt.Println(bits.TrailingZeros64(255))
	fmt.Printf("%b\n", 65534)
	fmt.Println(bits.TrailingZeros64(65534))
	fmt.Printf("%b\n", 65535)
	fmt.Println(bits.TrailingZeros64(65535))
	fmt.Printf("%b\n", 2147483648)
	fmt.Println(bits.TrailingZeros64(2147483648))
	fmt.Printf("%b\n", 4294967295)
	fmt.Println(bits.TrailingZeros64(4294967295))
	fmt.Printf("%b\n", 9223372036854775807)
	fmt.Println(bits.TrailingZeros64(9223372036854775807))
	// Output:
	// 0
	// 64
	// 11111110
	// 1
	// 11111111
	// 0
	// 1111111111111110
	// 1
	// 1111111111111111
	// 0
	// 10000000000000000000000000000000
	// 31
	// 11111111111111111111111111111111
	// 0
	// 111111111111111111111111111111111111111111111111111111111111111
	// 0
}
