// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func ExampleWrite() {
	buf := new(bytes.Buffer)
	var pi float64 = math.Pi
	err := binary.Write(buf, binary.LittleEndian, pi)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	fmt.Printf("% x", buf.Bytes())
	// Output: 18 2d 44 54 fb 21 09 40
}

func ExampleWrite_multi() {
	buf := new(bytes.Buffer)
	var data = []interface{}{
		uint16(61374),
		int8(-54),
		uint8(254),
	}
	for _, v := range data {
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
	}
	fmt.Printf("%x", buf.Bytes())
	// Output: beefcafe
}

func ExampleRead() {
	var pi float64
	b := []byte{0x18, 0x2d, 0x44, 0x54, 0xfb, 0x21, 0x09, 0x40}
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.LittleEndian, &pi)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	fmt.Print(pi)
	// Output: 3.141592653589793
}

func ExamplePutUvarint() {
	x := uint64(0x42f0e1eba9ea3693)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, x)
	fmt.Printf("%x", buf[:n])
	// Output: 93eda8cfbabdb8f842
}

func ExamplePutVarint() {
	x := int64(0x42f0e1eba9ea3693)
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, x)
	fmt.Printf("%x", buf[:n])
	// Output: a6dad19ef5faf0f08501
}

func ExampleUvarint() {
	b := []byte{0x93, 0xed, 0xa8, 0xcf, 0xba, 0xbd, 0xb8, 0xf8, 0x42}
	x, n := binary.Uvarint(b)
	if n != len(b) {
		fmt.Println("Uvarint did not consume all of b")
	}
	fmt.Printf("%#x", x)
	// Output: 0x42f0e1eba9ea3693
}

func ExampleVarint() {
	b := []byte{0xa6, 0xda, 0xd1, 0x9e, 0xf5, 0xfa, 0xf0, 0xf0, 0x85, 0x01}
	x, n := binary.Varint(b)
	if n != len(b) {
		fmt.Println("Uvarint did not consume all of b")
	}
	fmt.Printf("%#x", x)
	// Output: 0x42f0e1eba9ea3693
}
