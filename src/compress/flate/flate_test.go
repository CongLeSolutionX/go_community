// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This test tests some internals of the flate package.
// The tests in package compress/gzip serve as the
// end-to-end test of the decompressor.

package flate

import (
	"bytes"
	"io"
	"testing"
)

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

// Verify that flate.Reader.Read returns (n, io.EOF) instead
// of (n, nil) + (0, io.EOF) when possible.
//
// This helps net/http.Transport reuse HTTP/1 connections more
// aggressively.
//
// See https://github.com/google/go-github/pull/317 for background.
func TestReaderEarlyEOF(t *testing.T) {
	testSizes := []int{
		1, 2, 3, 4, 5, 6, 7, 8,
		100, 1000, 10000, 100000,
		128, 1024, 16384, 131072,

		// Testing multiples of windowSize triggers the case
		// where Read will fail to return an early io.EOF.
		windowSize * 1, windowSize * 2, windowSize * 3,
	}

	var maxSize int
	for _, n := range testSizes {
		if maxSize < n {
			maxSize = n
		}
	}

	readBuf := make([]byte, 40)
	data := make([]byte, maxSize)
	for i := range data {
		data[i] = byte(i)
	}

	for _, sz := range testSizes {
		if testing.Short() && sz > windowSize {
			continue
		}
		for _, flush := range []bool{true, false} {
			earlyEOF := true // Do we expect early io.EOF?

			var buf bytes.Buffer
			w, _ := NewWriter(&buf, 5)
			w.Write(data[:sz])
			if flush {
				// If a Flush occurs after all the actual data, the flushing
				// semantics dictate that we will observe a (0, io.EOF) since
				// Read must return data before it knows that the stream ended.
				w.Flush()
				earlyEOF = false
			}
			w.Close()

			r := NewReader(&buf)
			for {
				n, err := r.Read(readBuf)
				if err == io.EOF {
					// If the availWrite == windowSize, then that means that the
					// previous Read returned because the write buffer was full
					// and it just so happened that the stream had no more data.
					// This situation is rare, but unavoidable.
					if r.(*decompressor).dict.availWrite() == windowSize {
						earlyEOF = false
					}

					if n == 0 && earlyEOF {
						t.Errorf("On size:%d flush:%v, Read() = (0, io.EOF), want (n, io.EOF)", sz, flush)
					}
					if n != 0 && !earlyEOF {
						t.Errorf("On size:%d flush:%v, Read() = (%d, io.EOF), want (0, io.EOF)", sz, flush, n)
					}
					break
				}
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
