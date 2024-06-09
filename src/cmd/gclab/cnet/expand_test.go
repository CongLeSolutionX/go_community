// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

// TODO: None of these tests actually test anything! They were just a handy way
// to manually inspect results.

func printBits(x []uint64) {
	var buf bytes.Buffer
	// Print the header line
	fmt.Fprintf(&buf, "    ")
	for i := range 64 {
		if i%16 == 0 {
			buf.WriteByte(' ')
		}
		fmt.Fprintf(&buf, "%x", i%16)
	}
	buf.WriteByte('\n')
	for i, v := range x {
		fmt.Fprintf(&buf, "%04x", i*64)
		for i := range 64 {
			if i%16 == 0 {
				buf.WriteByte(' ')
			}
			buf.WriteByte('0' + byte((v>>i)&1))
		}
		buf.WriteByte('\n')
	}
	fmt.Print(buf.String())
}

func TestExpandX3(t *testing.T) {
	t.Skip()
	var inp [6]uint64
	var outp ptrMask
	for i := range inp {
		inp[i] = 0b_10101010_10101010_10101010_10101010_10101010_10101010_10101010_10101010
	}
	expandX3(&inp, &outp)
	log.Fatalf("%064b\n%064b", inp, outp)
}

func TestExpand3(t *testing.T) {
	t.Skip()
	var inp [6]uint64
	var outp ptrMask
	for i := range inp {
		inp[i] = 0b_10101010_10101010_10101010_10101010_10101010_10101010_10101010_10101010
	}
	expandAsm(3, &inp, &outp)
	printBits(inp[:])
	printBits(outp[:])

	var inp2 [6]uint64
	for i := range len(inp) * 64 / 2 {
		bit := i*2 + 1
		inp2 = inp
		inp2[bit/64] &^= 1 << (bit % 64)
		clear(outp[:])
		expandAsm(3, &inp2, &outp)
		printBits(inp2[:])
		printBits(outp[:])
	}
	//log.Fatalf("%064b\n%064b", inp, outp)
}

func TestExpandX6(t *testing.T) {
	t.Skip()
	var inp [3]uint64
	var outp ptrMask
	for i := range inp {
		inp[i] = 0b_10101010_10101010_10101010_10101010_10101010_10101010_10101010_10101010
	}
	expandX6AVX512(&inp, &outp)
	log.Fatalf("%064b\n%064b", inp, outp)
}
