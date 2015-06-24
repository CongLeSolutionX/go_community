// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestLimitWriter(t *testing.T) {
	var buf bytes.Buffer
	lw := &limitWriter{&buf, 5}
	if n, err := lw.Write([]byte("Hello, world!")); n != 13 || err != nil {
		t.Errorf("Write = %v, %v; want 13, nil", n, err)
	}
	if buf.String() != "Hello" {
		t.Errorf("Buf = %q; want Hello", buf.String())
	}

	buf.Reset()
	pr, pw := io.Pipe()
	someErr := errors.New("some error")
	pr.CloseWithError(someErr)
	lw = &limitWriter{io.MultiWriter(&buf, pw), 4}
	if n, err := lw.Write([]byte("some failure")); n != 0 || err != someErr {
		t.Errorf("Write = %v, %v; want 0, someErr", n, err)
	}
	if buf.String() != "some" {
		t.Errorf("Buf = %q; want 'some'", buf.String())
	}
}
