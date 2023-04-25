// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(windows || js || wasip1)

package main

import (
	"strings"
	"testing"
)

func TestExitCodeFilter(t *testing.T) {
	const pre = `abcexitcode=123def`
	const text = pre + `exitcode=1`

	// Write text to the filter one character at a time.
	var out strings.Builder
	f := newExitCodeFilter(&out, "exitcode=")
	for i := 0; i < len(text); i++ {
		_, err := f.Write([]byte{text[i]})
		if err != nil {
			t.Fatal(err)
		}
	}

	// The "pre" output should all have been flushed already.
	if want, got := pre, out.String(); want != got {
		t.Errorf("filter should have already flushed %q, but flushed %q", want, got)
	}

	code, err := f.Finish()
	if err != nil {
		t.Fatal(err)
	}

	// Nothing more should have been written to out.
	if want, got := pre, out.String(); want != got {
		t.Errorf("want output %q, got %q", want, got)
	}
	if want := 1; want != code {
		t.Errorf("want exit code %d, got %d", want, code)
	}
}

func TestExitCodeMissing(t *testing.T) {
	check := func(text string) {
		t.Helper()
		var out strings.Builder
		f := newExitCodeFilter(&out, "exitcode=")
		f.Write([]byte(text))
		_, err := f.Finish()
		// We should get a no exit code error
		if err == nil || !strings.HasPrefix(err.Error(), "no exit code") {
			t.Errorf("want 'no exit code' error, got %s", err)
		}
		// And it should flush all output (even if it looks
		// like we may be getting an exit code)
		if got := out.String(); text != got {
			t.Errorf("want full output %q, got %q", text, got)
		}
	}
	check("abc")
	check("exitcode")
	check("exitcode=")
	check("exitcode=123\n")
}
