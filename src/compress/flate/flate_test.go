// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This test tests some internals of the flate package.
// The tests in package compress/gzip serve as the
// end-to-end test of the decompressor.

package flate

import (
	"bytes"
	"encoding/hex"
	"io/ioutil"
	"testing"
)

// The following test should not panic.
func TestIssue5915(t *testing.T) {
	bits := []int{4, 0, 0, 6, 4, 3, 2, 3, 3, 4, 4, 5, 0, 0, 0, 0, 5, 5, 6,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 8, 6, 0, 11, 0, 8, 0, 6, 6, 10, 8}
	var h huffmanDecoder
	if h.init(bits) {
		t.Fatalf("Given sequence of bits is bad, and should not succeed.")
	}
}

// The following test should not panic.
func TestIssue5962(t *testing.T) {
	bits := []int{4, 0, 0, 6, 4, 3, 2, 3, 3, 4, 4, 5, 0, 0, 0, 0,
		5, 5, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11}
	var h huffmanDecoder
	if h.init(bits) {
		t.Fatalf("Given sequence of bits is bad, and should not succeed.")
	}
}

// The following test should not panic.
func TestIssue6255(t *testing.T) {
	bits1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 11}
	bits2 := []int{11, 13}
	var h huffmanDecoder
	if !h.init(bits1) {
		t.Fatalf("Given sequence of bits is good and should succeed.")
	}
	if h.init(bits2) {
		t.Fatalf("Given sequence of bits is bad and should not succeed.")
	}
}

func TestInvalidEncoding(t *testing.T) {
	// Initialize Huffman decoder to recognize "0".
	var h huffmanDecoder
	if !h.init([]int{1}) {
		t.Fatal("Failed to initialize Huffman decoder")
	}

	// Initialize decompressor with invalid Huffman coding.
	var f decompressor
	f.r = bytes.NewReader([]byte{0xff})

	_, err := f.huffSym(&h)
	if err == nil {
		t.Fatal("Should have rejected invalid bit sequence")
	}
}

func TestInvalidBits(t *testing.T) {
	oversubscribed := []int{1, 2, 3, 4, 4, 5}
	incomplete := []int{1, 2, 4, 4}
	var h huffmanDecoder
	if h.init(oversubscribed) {
		t.Fatal("Should reject oversubscribed bit-length set")
	}
	if h.init(incomplete) {
		t.Fatal("Should reject incomplete bit-length set")
	}
}

func TestInvalidStreams(t *testing.T) {
	badStreams := []string{
		// over-subscribed HCLenTree
		"344c4a4e494d4b070000ff2e2eff2e2e2e2e2eff",
		// degenerate HCLenTree
		"05e0010000000000100000000000000000000000000000000000000000000000" +
			"00000000000000000004",
		// complete HCLenTree, empty HLitTree, empty HDistTree
		"05e0010400000000000000000000000000000000000000000000000000000000" +
			"00000000000000000010",
		// empty HCLenTree
		"05e0010000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000010",
		// complete HCLenTree, complete HLitTree, empty HDistTree, use missing HDist symbol
		"000100feff000de0010400000000100000000000000000000000000000000000" +
			"0000000000000000000000000000002c",
		// complete HCLenTree, complete HLitTree, degenerate HDistTree, use missing HDist symbol
		"000100feff000de0010000000000000000000000000000000000000000000000" +
			"00000000000000000610000000004070",
		// complete HCLenTree, empty HLitTree, empty HDistTree
		"05e0010400000000100400000000000000000000000000000000000000000000" +
			"0000000000000000000000000008",
		// complete HCLenTree, empty HLitTree, degenerate HDistTree
		"05e0010400000000100400000000000000000000000000000000000000000000" +
			"0000000000000000000800000008",
		// complete HCLenTree, degenerate HLitTree, degenerate HDistTree, use missing HLit symbol
		"05e0010400000000100000000000000000000000000000000000000000000000" +
			"0000000000000000001c",
		// complete HCLenTree, over-subscribed HLitTree, empty HDistTree
		"05e001240000000000fcffffffffffffffffffffffffffffffffffffffffffff" +
			"ffffffffffffffffff07f00f",
		// complete HCLenTree, under-subscribed HLitTree, empty HDistTree
		"05e001240000000000fcffffffffffffffffffffffffffffffffffffffffffff" +
			"fffffffffcffffffff07f00f",
		// complete HCLenTree, complete HLitTree, too large HDistTree
		"edff870500000000200400000000000000000000000000000000000000000000" +
			"000000000000000000080000000000000004",
		// complete HCLenTree, complete HLitTree, empty HDistTree, excessive repeater code
		"edfd870500000000200400000000000000000000000000000000000000000000" +
			"000000000000000000e8b100",
		// fixed block, use reserved symbol 287
		"33180700",
		// issue 10426
		"344c4a4e494d4b070000ff2e2eff2e2e2e2e2eff",
	}

	rd := NewReader(nil)
	rrd := rd.(Resetter)
	for idx, invalid := range badStreams {
		stream, err := hex.DecodeString(invalid)
		if err != nil {
			t.Fatal(err)
		}
		rrd.Reset(bytes.NewReader(stream), nil)
		_, err = ioutil.ReadAll(rd)
		if err == nil {
			t.Fatalf("Failed case %d: expected decoder to reject bad stream", idx)
		}
	}
}

func TestValidStreams(t *testing.T) {
	goodStreams := []string{
		// complete HCLenTree, complete HLitTree, degenerate HDistTree, use valid HDist symbol
		"000100feff000de0010400000000100000000000000000000000000000000000" +
			"0000000000000000000000000000003c",
		// complete HCLenTree, degenerate HLitTree, degenerate HDistTree
		"05e0010400000000100000000000000000000000000000000000000000000000" +
			"0000000000000000000c",
		// complete HCLenTree, degenerate HLitTree, empty HDistTree
		"05e0010400000000100000000000000000000000000000000000000000000000" +
			"00000000000000000004",
		// complete HCLenTree, complete HLitTree with single code, empty HDistTree
		"05e001240000000000f8ffffffffffffffffffffffffffffffffffffffffffff" +
			"ffffffffffffffffff07f00f",
		// complete HCLenTree, complete HLitTree with multiple codes, empty HDistTree
		"05e301240000000000f8ffffffffffffffffffffffffffffffffffffffffffff" +
			"ffffffffffffffffff07807f",
		// complete HCLenTree, complete HLitTree, empty HDistTree, spanning repeater code
		"edfd870500000000200400000000000000000000000000000000000000000000" +
			"000000000000000000e8b000",
		// complete HCLenTree with length codes, complete HLitTree, empty HDistTree
		"ede0010400000000100000000000000000000000000000000000000000000000" +
			"0000000000000000000400004000",
		// complete HCLenTree, complete HLitTree, degenerate HDistTree, use valid HLit and HDist symbols
		"0cc2010d00000082b0ac4aff0eb07d27060000ffff",
		// complete HCLenTree, complete HLitTree, degenerate HDistTree, use valid HLit symbol 284 with count 31
		"000100feff00ede0010400000000100000000000000000000000000000000000" +
			"000000000000000000000000000000040000407f00",
		// raw block
		"010100feff11",
		// issue 11030
		"05c0070600000080400fff37a0ca",
		// issue 11033
		"050fb109c020cca5d017dcbca044881ee1034ec149c8980bbc413c2ab35be9dc" +
			"b1473449922449922411202306ee97b0383a521b4ffdcf3217f9f7d3adb701",
	}
	outputs := []string{
		"00000000",
		"",
		"",
		"01",
		"01",
		"",
		"",
		"616263616263",
		"00000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000" +
			"00000000000000000000000000000000000000000000000000000000000000000000" +
			"000000000000000000000000000000000000000000",
		"11",
		"",
		"3130303634342068652e706870005d05355f7ed957ff084a90925d19e3ebc6d0c6d7",
	}

	rd := NewReader(nil)
	rrd := rd.(Resetter)
	for idx, good := range goodStreams {
		stream, err := hex.DecodeString(good)
		if err != nil {
			t.Fatal(err)
		}
		rrd.Reset(bytes.NewReader(stream), nil)
		data, err := ioutil.ReadAll(rd)
		if err != nil {
			t.Fatalf("Failed case %d: %v", err)
		}
		if out := hex.EncodeToString(data); out != outputs[idx] {
			t.Fatalf("Failed case %d: expected '%s', but got '%s'", idx, outputs[idx], out)
		}
	}
}
