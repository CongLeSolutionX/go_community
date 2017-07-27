// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ioutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

func checkSize(t *testing.T, path string, size int64) {
	dir, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat %q (looking for size %d): %s", path, size, err)
	}
	if dir.Size() != size {
		t.Errorf("Stat %q: size %d want %d", path, dir.Size(), size)
	}
}

func TestReadFile(t *testing.T) {
	filename := "rumpelstilzchen"
	contents, err := ReadFile(filename)
	if err == nil {
		t.Fatalf("ReadFile %s: error expected, none found", filename)
	}

	filename = "ioutil_test.go"
	contents, err = ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", filename, err)
	}

	checkSize(t, filename, int64(len(contents)))
}

func TestWriteFile(t *testing.T) {
	f, err := TempFile("", "ioutil-test")
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	data := "Programming today is a race between software engineers striving to " +
		"build bigger and better idiot-proof programs, and the Universe trying " +
		"to produce bigger and better idiots. So far, the Universe is winning."

	if err := WriteFile(filename, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", filename, err)
	}

	contents, err := ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", filename, err)
	}

	if string(contents) != data {
		t.Fatalf("contents = %q\nexpected = %q", string(contents), data)
	}

	// cleanup
	f.Close()
	os.Remove(filename) // ignore error
}

func TestReadDir(t *testing.T) {
	dirname := "rumpelstilzchen"
	_, err := ReadDir(dirname)
	if err == nil {
		t.Fatalf("ReadDir %s: error expected, none found", dirname)
	}

	dirname = ".."
	list, err := ReadDir(dirname)
	if err != nil {
		t.Fatalf("ReadDir %s: %v", dirname, err)
	}

	foundFile := false
	foundSubDir := false
	for _, dir := range list {
		switch {
		case !dir.IsDir() && dir.Name() == "io_test.go":
			foundFile = true
		case dir.IsDir() && dir.Name() == "ioutil":
			foundSubDir = true
		}
	}
	if !foundFile {
		t.Fatalf("ReadDir %s: io_test.go file not found", dirname)
	}
	if !foundSubDir {
		t.Fatalf("ReadDir %s: ioutil directory not found", dirname)
	}
}

type delayReader struct {
	br    *bytes.Reader
	delay time.Duration
}

func newDelayReader(size int64, delay time.Duration) delayReader {
	return delayReader{
		br:    bytes.NewReader(make([]byte, size)),
		delay: delay,
	}
}

func (r delayReader) Read(p []byte) (int, error) {
	time.Sleep(r.delay)
	return r.br.Read(p)
}

func BenchmarkDiscard(b *testing.B) {
	for _, tc := range []struct {
		readDelay time.Duration
	}{
		{readDelay: 0},
		{readDelay: time.Microsecond},
		{readDelay: time.Millisecond},
		{readDelay: 5 * time.Millisecond},
	} {
		b.Run(fmt.Sprintf("delay=%s", tc.readDelay), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				n, err := io.Copy(Discard, newDelayReader(10<<20, tc.readDelay))
				if err != nil {
					b.Fatalf("Failed to copy random data to discard: %v", err)
				}
				b.SetBytes(n)
			}
		})

	}
}
