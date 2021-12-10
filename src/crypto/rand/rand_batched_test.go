// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package rand

import (
	"bytes"
	prand "math/rand"
	"testing"
)

func TestBatched(t *testing.T) {
	fillBatched := batched(func(p []byte) bool {
		for i := range p {
			p[i] = byte(i)
		}
		return true
	}, 5)

	p := make([]byte, 13)
	if !fillBatched(p) {
		t.Fatal("batched function returned false")
	}
	expected := []byte{0, 1, 2, 3, 4, 0, 1, 2, 3, 4, 0, 1, 2}
	if !bytes.Equal(expected, p) {
		t.Errorf("incorrect batch result: got %x, want %x", p, expected)
	}
}

func TestBatchedBuffering(t *testing.T) {
	backingStore := make([]byte, 1<<23)
	prand.Read(backingStore)
	backingMarker := backingStore[:]
	output := make([]byte, len(backingStore))
	outputMarker := output[:]

	fillBatched := batched(func(p []byte) bool {
		n := copy(p, backingMarker)
		backingMarker = backingMarker[n:]
		return true
	}, 731)

	for len(outputMarker) > 0 {
		max := 9200
		if max > len(outputMarker) {
			max = len(outputMarker)
		}
		howMuch := prand.Intn(max + 1)
		if !fillBatched(outputMarker[:howMuch]) {
			t.Fatal("batched function returned false")
		}
		outputMarker = outputMarker[howMuch:]
	}
	if !bytes.Equal(backingStore, output) {
		t.Error("incorrect batch result")
	}
}

func TestBatchedError(t *testing.T) {
	b := batched(func(p []byte) bool { return false }, 5)
	if b(make([]byte, 13)) {
		t.Fatal("batched function should have returned false")
	}
}

func TestBatchedEmpty(t *testing.T) {
	b := batched(func(p []byte) bool { return false }, 5)
	if !b(make([]byte, 0)) {
		t.Fatal("empty slice should always return true")
	}
}
