// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall_test

import (
	"slices"
	"syscall"
	"testing"
)

var wtf8tests = []struct {
	name    string
	content []uint8
	want    []uint16
}{
	{
		name:    "0x0000",
		content: []uint8{0x00},
		want:    []uint16{0x00},
	},
	{
		name:    "0x005C",
		content: []uint8{0x5C},
		want:    []uint16{0x5C},
	},
	{
		name:    "0x007F",
		content: []uint8{0x7F},
		want:    []uint16{0x7F},
	},

	// 2-byte
	{
		name:    "0x0080",
		content: []uint8{0xC2, 0x80},
		want:    []uint16{0x80},
	},
	{
		name:    "0x05CA",
		content: []uint8{0xD7, 0x8A},
		want:    []uint16{0x05CA},
	},
	{
		name:    "0x07FF",
		content: []uint8{0xDF, 0xBF},
		want:    []uint16{0x07FF},
	},

	// 3-byte
	{
		name:    "0x0800",
		content: []uint8{0xE0, 0xA0, 0x80},
		want:    []uint16{0x0800},
	},
	{
		name:    "0x2C3C",
		content: []uint8{0xE2, 0xB0, 0xBC},
		want:    []uint16{0x2C3C},
	},
	{
		name:    "0xFFFF",
		content: []uint8{0xEF, 0xBF, 0xBF},
		want:    []uint16{0xFFFF},
	},
	// unmatched surrogate halves
	// high surrogates: 0xD800 to 0xDBFF
	{
		name:    "0xD800",
		content: []uint8{0xED, 0xA0, 0x80},
		want:    []uint16{0xD800},
	},
	{
		name:    "High surrogate followed by another high surrogate",
		content: []uint8{0xED, 0xA0, 0x80, 0xED, 0xA0, 0x80},
		want:    []uint16{0xD800, 0xD800},
	},
	{
		name:    "High surrogate followed by a symbol that is not a surrogate",
		content: []uint8{0xED, 0xA0, 0x80, 0xA},
		want:    []uint16{0xD800, 0xA},
	},
	{
		name:    "Unmatched high surrogate, followed by a surrogate pair, followed by an unmatched high surrogate",
		content: []uint8{0xED, 0xA0, 0x80, 0xF0, 0x9D, 0x8C, 0x86, 0xED, 0xA0, 0x80},
		want:    []uint16{0xD800, 0xD834, 0xDF06, 0xD800},
	},
	{
		name:    "0xD9AF",
		content: []uint8{0xED, 0xA6, 0xAF},
		want:    []uint16{0xD9AF},
	},
	{
		name:    "0xDBFF",
		content: []uint8{0xED, 0xAF, 0xBF},
		want:    []uint16{0xDBFF},
	},
	// low surrogates: 0xDC00 to 0xDFFF
	{
		name:    "0xDC00",
		content: []uint8{0xED, 0xB0, 0x80},
		want:    []uint16{0xDC00},
	},
	{
		name:    "Low surrogate followed by another low surrogate",
		content: []uint8{0xED, 0xB0, 0x80, 0xED, 0xB0, 0x80},
		want:    []uint16{0xDC00, 0xDC00},
	},
	{
		name:    "Low surrogate followed by a symbol that is not a surrogate",
		content: []uint8{0xED, 0xB0, 0x80, 0xA},
		want:    []uint16{0xDC00, 0xA},
	},
	{
		name:    "Unmatched low surrogate, followed by a surrogate pair, followed by an unmatched low surrogate",
		content: []uint8{0xED, 0xB0, 0x80, 0xF0, 0x9D, 0x8C, 0x86, 0xED, 0xB0, 0x80},
		want:    []uint16{0xDC00, 0xD834, 0xDF06, 0xDC00},
	},
	{
		name:    "0xDEEE",
		content: []uint8{0xED, 0xBB, 0xAE},
		want:    []uint16{0xDEEE},
	},
	{
		name:    "0xDFFF",
		content: []uint8{0xED, 0xBF, 0xBF},
		want:    []uint16{0xDFFF},
	},

	// 4-byte
	{
		name:    "0x010000",
		want:    []uint16{0xD800, 0xDC00},
		content: []uint8{0xF0, 0x90, 0x80, 0x80},
	},
	{
		name:    "0x01D306",
		want:    []uint16{0xD834, 0xDF06},
		content: []uint8{0xF0, 0x9D, 0x8C, 0x86},
	},
	{
		name:    "0x10FFF",
		want:    []uint16{0xDBFF, 0xDFFF},
		content: []uint8{0xF4, 0x8F, 0xBF, 0xBF},
	},
}

func TestWTF16Rountrip(t *testing.T) {
	for _, tt := range wtf8tests {
		t.Run(tt.name, func(t *testing.T) {
			got := syscall.EncodeWTF16(string(tt.content), nil)
			got2 := string(syscall.DecodeWTF16(got, nil))
			if got2 != string(tt.content) {
				t.Errorf("got:\n%swant:\n%s", got2, string(tt.content))
			}
		})
	}
}

func TestWTF16Golden(t *testing.T) {
	for _, tt := range wtf8tests {
		t.Run(tt.name, func(t *testing.T) {
			got := syscall.EncodeWTF16(string(tt.content), nil)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got:\n%vwant:\n%v", got, tt.want)
			}
		})
	}
}
