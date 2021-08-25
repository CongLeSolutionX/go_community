// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tar

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func FuzzReader(f *testing.F) {
	testdata, err := os.ReadDir("testdata")
	if err != nil {
		f.Fatalf("failed to read testdata directory: %s", err)
	}
	for _, de := range testdata {
		if de.IsDir() {
			continue
		}
		if de.Name() == "gnu-sparse-big.tar" || de.Name() == "pax-sparse-big.tar" {
			// these two are too big to be useful, and often will just
			// timeout the target trying to load the archive contents
			continue
		}
		b, err := os.ReadFile(filepath.Join("testdata", de.Name()))
		if err != nil {
			f.Fatalf("failed to read testdata: %s", err)
		}
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		r := NewReader(bytes.NewReader(b))
		type file struct {
			header  *Header
			content []byte
		}
		files := []file{}
		for {
			hdr, err := r.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
			buf := bytes.NewBuffer(nil)
			if _, err := io.Copy(buf, r); err != nil {
				return
			}
			files = append(files, file{header: hdr, content: buf.Bytes()})
		}

		out := bytes.NewBuffer(nil)
		w := NewWriter(out)
		for _, f := range files {
			if err := w.WriteHeader(f.header); err != nil {
				t.Fatalf("unable to write previously parsed header: %s", err)
			}
			if _, err := w.Write(f.content); err != nil {
				t.Fatalf("unable to write previously parsed content: %s", err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Unable to write archive: %s", err)
		}

		// TODO: We may want to check if the archive roundtrips. This would require
		// taking into account addition of the two zero trailer blocks that Writer.Close
		// appends.
	})
}
