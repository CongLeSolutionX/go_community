// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zip

import (
	"bytes"
	"io"
	"testing"
)

func fuzz(t *testing.T, b []byte) {
	r, err := NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return
	}

	type file struct {
		header  *FileHeader
		content []byte
	}
	files := []file{}

	for _, f := range r.File {
		fr, err := f.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(fr)
		if err != nil {
			continue
		}
		files = append(files, file{header: &f.FileHeader, content: content})
		if _, err := r.Open(f.Name); err != nil {
			continue
		}
	}

	w := NewWriter(io.Discard)
	for _, f := range files {
		ww, err := w.CreateHeader(f.header)
		if err != nil {
			t.Fatalf("unable to write previously parsed header: %s", err)
		}
		if _, err := ww.Write(f.content); err != nil {
			t.Fatalf("unable to write previously parsed content: %s", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Unable to write archive: %s", err)
	}

	// TODO: We may want to check if the archive roundtrips.
}

func FuzzReaderEmptySeed(f *testing.F) {
	f.Add([]byte{})

	f.Fuzz(fuzz)
}

func FuzzReaderBasicSeed(f *testing.F) {
	b := bytes.NewBuffer(nil)
	w := NewWriter(b)
	ww, err := w.Create("lorem.txt")
	if err != nil {
		f.Fatalf("failed to create writer: %s", err)
	}
	_, err = ww.Write([]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."))
	if err != nil {
		f.Fatalf("failed to write file to archive: %s", err)
	}
	if err := w.Close(); err != nil {
		f.Fatalf("failed to write archive: %s", err)
	}
	f.Add(b.Bytes())

	f.Fuzz(fuzz)
}
