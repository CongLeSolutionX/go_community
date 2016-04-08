// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"bytes"
	"strings"
	"testing"
)

func TestCorruptedInputs(t *testing.T) {
	// These inputs crashed parser previously.
	tests := []string{
		"gotrace\x00\x020",
		"gotrace\x00Q00\x020",
		"gotrace\x00T00\x020",
		"gotrace\x00\xc3\x0200",
		"go 1.5 trace\x00\x00\x00\x00\x020",
		"go 1.5 trace\x00\x00\x00\x00Q00\x020",
		"go 1.5 trace\x00\x00\x00\x00T00\x020",
		"go 1.5 trace\x00\x00\x00\x00\xc3\x0200",
	}
	for _, data := range tests {
		events, err := Parse(strings.NewReader(data))
		if err == nil || events != nil {
			t.Fatalf("no error on input: %q\n", data)
		}
	}
}

func TestTimestampOverflow(t *testing.T) {
	// Test that parser correctly handles large timestamps (long tracing).
	w := newWriter()
	w.emit(EvBatch, 0, 0, 0)
	w.emit(EvFrequency, 1e9, 0)
	for ts := uint64(1); ts < 1e16; ts *= 2 {
		w.emit(EvGoCreate, 1, ts, ts, 1, 0)
	}
	if _, err := Parse(w); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
}

type writer struct {
	bytes.Buffer
}

func newWriter() *writer {
	w := new(writer)
	w.Write([]byte("go 1.5 trace\x00\x00\x00\x00"))
	return w
}

func (w *writer) emit(typ byte, args ...uint64) {
	nargs := byte(len(args)) - 2
	if nargs > 3 {
		nargs = 3
	}
	buf := []byte{typ | nargs<<6}
	if nargs == 3 {
		buf = append(buf, 0)
	}
	for _, a := range args {
		buf = appendVarint(buf, a)
	}
	if nargs == 3 {
		buf[1] = byte(len(buf) - 2)
	}
	n, err := w.Write(buf)
	if n != len(buf) || err != nil {
		panic("failed to write")
	}
}

func appendVarint(buf []byte, v uint64) []byte {
	for ; v >= 0x80; v >>= 7 {
		buf = append(buf, 0x80|byte(v))
	}
	buf = append(buf, byte(v))
	return buf
}
