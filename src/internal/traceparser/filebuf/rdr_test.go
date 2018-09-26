// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filebuf

// These Test routines are copies of the ones in filebuf_test.go, except
// that they use a Reader. The code duplication sucks. // PJW

import (
	"bytes"
	"io"
	"testing"
)

func get(n int) io.Reader {
	create()
	if n <= len(contents) {
		return bytes.NewReader(contents[:n])
	}
	return bytes.NewReader(contents)
}

func TestRdrSmall(t *testing.T) {
	f, err := FromReader(get(7))
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 23)
	n, err := f.Read(buf)
	if n != 7 || err != io.EOF {
		t.Errorf("got %d, expected 7, %v", n, err)
	}
	m, err := f.Seek(0, io.SeekCurrent)
	if m != 7 || err != nil {
		t.Errorf("got %d, expected 7, %v", m, err)
	}
	m, err = f.Seek(1, io.SeekStart)
	if m != 1 || err != nil {
		t.Errorf("got %d expected 1, %v", m, err)
	}
	n, err = f.Read(buf)
	if n != 6 || err != io.EOF {
		t.Errorf("got %d, expected 6, %v", n, err)
	}
	for i := 0; i < 6; i++ {
		if buf[i] != byte(i+1) {
			t.Fatalf("byte %d is %d, not %d, %v", i, buf[i], i+1, buf)
		}
	}
}

func TestReaderLarge(t *testing.T) {
	f, err := FromReader(get(1 << 30))
	if err != nil {
		t.Fatal(err)
	}

	x := Buflen - 7
	n, err := f.Seek(int64(x), io.SeekStart)
	if n != Buflen-7 || err != nil {
		t.Fatalf("expected %d, got %d, %v", x, n, err)
	}
	buf := make([]byte, 23)
	m, err := f.Read(buf)
	if m != len(buf) || err != nil {
		t.Fatalf("expected %d, got %d, %v", len(buf), m, err)
	}
	for i := 0; i < 23; i++ {
		if buf[i] != byte(x+i) {
			t.Fatalf("byte %d, got %d, wanted %d", i, buf[i],
				byte(x+i))
		}
	}
	m, err = f.Read(buf)
	if m != len(buf) || err != nil {
		t.Fatalf("got %d, expected %d, %v", m, len(buf), err)
	}
	x += len(buf)
	for i := 0; i < 23; i++ {
		if buf[i] != byte(x+i) {
			t.Fatalf("byte %d, got %d, wanted %d", i, buf[i],
				byte(x+i))
		}
	}
}
