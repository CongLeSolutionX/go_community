// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux || freebsd || dragonfly || solaris

package rand

import (
	"bytes"
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
	if err := fillBatched(p); err != nil {
		t.Fatalf("batched function returned error: %s", err)
	}
	expected := []byte{0, 1, 2, 3, 4, 0, 1, 2, 3, 4, 0, 1, 2}
	if !bytes.Equal(expected, p) {
		t.Errorf("incorrect batch result: got %x, want %x", p, expected)
	}
}

<<<<<<< HEAD   (8ed0e5 [release-branch.go1.18] go1.18.2)
=======
func TestBatchedBuffering(t *testing.T) {
	var prandSeed [8]byte
	Read(prandSeed[:])
	prand.Seed(int64(binary.LittleEndian.Uint64(prandSeed[:])))

	backingStore := make([]byte, 1<<23)
	prand.Read(backingStore)
	backingMarker := backingStore[:]
	output := make([]byte, len(backingStore))
	outputMarker := output[:]

	fillBatched := batched(func(p []byte) error {
		n := copy(p, backingMarker)
		backingMarker = backingMarker[n:]
		return nil
	}, 731)

	for len(outputMarker) > 0 {
		max := 9200
		if max > len(outputMarker) {
			max = len(outputMarker)
		}
		howMuch := prand.Intn(max + 1)
		if err := fillBatched(outputMarker[:howMuch]); err != nil {
			t.Fatalf("batched function returned error: %s", err)
		}
		outputMarker = outputMarker[howMuch:]
	}
	if !bytes.Equal(backingStore, output) {
		t.Error("incorrect batch result")
	}
}

>>>>>>> CHANGE (bb1f44 crypto/rand: properly handle large Read on windows)
func TestBatchedError(t *testing.T) {
<<<<<<< HEAD   (8ed0e5 [release-branch.go1.18] go1.18.2)
	b := batched(func(p []byte) bool { return false }, 5)
	if b(make([]byte, 13)) {
		t.Fatal("batched function should have returned false")
=======
	b := batched(func(p []byte) error { return errors.New("failure") }, 5)
	if b(make([]byte, 13)) == nil {
		t.Fatal("batched function should have returned an error")
>>>>>>> CHANGE (bb1f44 crypto/rand: properly handle large Read on windows)
	}
}

func TestBatchedEmpty(t *testing.T) {
<<<<<<< HEAD   (8ed0e5 [release-branch.go1.18] go1.18.2)
	b := batched(func(p []byte) bool { return false }, 5)
	if !b(make([]byte, 0)) {
		t.Fatal("empty slice should always return true")
=======
	b := batched(func(p []byte) error { return errors.New("failure") }, 5)
	if b(make([]byte, 0)) != nil {
		t.Fatal("empty slice should always return successful")
>>>>>>> CHANGE (bb1f44 crypto/rand: properly handle large Read on windows)
	}
}
