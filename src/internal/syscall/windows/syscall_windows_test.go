// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

package windows_test

import (
	"internal/syscall/windows"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestUTF16PtrToStringAllocs(t *testing.T) {
	msg := "Hello, world 🐻"
	testUTF16PtrToStringAllocs(t, msg)
	testUTF16PtrToStringAllocs(t, strings.Repeat(msg, 10))
}

func testUTF16PtrToStringAllocs(t *testing.T, msg string) {
	in := utf16.Encode([]rune(msg + "\x00"))
	var out string
	alloccnt := testing.AllocsPerRun(1000, func() {
		out = windows.UTF16PtrToString(&in[0])
	})
	if out != msg {
		t.Errorf("windows.UTF16PtrToString(%q) returned %q; want %q", in, out, msg)
	}
	if alloccnt > 1.01 {
		t.Errorf("windows.UTF16PtrToString(%q) made %v allocs per call; want 1", in, alloccnt)
	}
}
