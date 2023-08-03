// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux
// +build linux

package net

import (
	"bytes"
	"internal/poll"
	"io"
	"math"
	"os"
	"strconv"
	"testing"
)

func BenchmarkSendFile(b *testing.B) {
	for i := 0; i <= 10; i++ {
		size := 1 << (i + 10)
		bench := sendFileBench{chunkSize: size}
		b.Run(strconv.Itoa(size), bench.benchSendFile)
	}
}

type sendFileBench struct {
	chunkSize int
}

func (bench sendFileBench) benchSendFile(b *testing.B) {
	fileSize := b.N * bench.chunkSize
	f := createTempFile(b, make([]byte, fileSize))
	fileName := f.Name()
	defer os.Remove(fileName)
	defer f.Close()

	client, server := spliceTestSocketPair(b, "tcp")
	defer server.Close()

	cleanUp, err := startSpliceClient(client, "r", bench.chunkSize, fileSize)
	if err != nil {
		b.Fatal(err)
	}
	defer cleanUp()

	b.ReportAllocs()
	b.SetBytes(int64(bench.chunkSize))
	b.ResetTimer()

	// Data go from file to socket via sendfile(2).
	sent, err := io.Copy(server, f)
	if err != nil {
		b.Fatalf("failed to copy data with sendfile, error: %v", err)
	}
	if sent != int64(fileSize) {
		b.Fatalf("bytes sent mismatch\n\texpect: %d\n\tgot: %d", fileSize, sent)
	}
}

func createTempFile(b testing.TB, data []byte) *os.File {
	f, err := os.CreateTemp("", "linux-sendfile-test")
	if err != nil {
		b.Fatalf("failed to create temporary file: %v", err)
	}

	if _, err := f.Write(data); err != nil {
		b.Fatalf("failed to create and feed the file: %v", err)
	}
	if err := f.Sync(); err != nil {
		b.Fatalf("failed to save the file: %v", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		b.Fatalf("failed to rewind the file: %v", err)
	}

	return f
}

func TestSendfileSectionReader(t *testing.T) {
	cases := []struct {
		tempFileContents string
		sectionOffset    int64
		sectionLength    int64
		expected         string
	}{
		{
			tempFileContents: "hello world!",
			sectionOffset:    0,
			sectionLength:    5,
			expected:         "hello",
		},
		{
			tempFileContents: "hello world!!!",
			sectionOffset:    6,
			sectionLength:    7,
			expected:         "world!!",
		},
		{
			tempFileContents: "hello world!!!",
			sectionOffset:    0,
			sectionLength:    math.MaxInt64,
			expected:         "hello world!!!",
		},
		{
			tempFileContents: "hello world!!!",
			sectionOffset:    6,
			sectionLength:    math.MaxInt64,
			expected:         "world!!!",
		},
		{
			tempFileContents: "hello world!!!",
			sectionOffset:    0,
			sectionLength:    0,
			expected:         "",
		},
		{
			tempFileContents: "hello world!!!",
			sectionOffset:    5,
			sectionLength:    0,
			expected:         "",
		},
		{
			tempFileContents: "hello world!!!",
			sectionOffset:    math.MaxInt64,
			sectionLength:    0,
			expected:         "",
		},
		{
			tempFileContents: "hello world!!!",
			sectionOffset:    1000,
			sectionLength:    5,
			expected:         "",
		},
	}

	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			tf := createTempFile(t, []byte(c.tempFileContents))
			defer os.Remove(tf.Name())
			defer tf.Close()

			_, err := tf.Seek(3, io.SeekStart)
			if err != nil {
				t.Fatalf("failed to seek file: %v", err)
			}

			client, server := spliceTestSocketPair(t, "tcp")
			defer server.Close()

			doneCh := make(chan struct{})
			buf := &bytes.Buffer{}
			var rerr error
			go func() {
				defer close(doneCh)
				_, rerr = io.Copy(buf, client)
			}()

			sendFileTestHook = func(pfd *poll.FD, f *os.File, pos, remain, written int64, err error, handled bool) {
				if pos != c.sectionOffset {
					t.Errorf("pos mismatch\n\texpect: %d\n\tgot: %d", c.sectionOffset, pos)
				}
				if remain != c.sectionLength {
					t.Errorf("remain mismatch\n\texpect: %d\n\tgot: %d", c.sectionLength, remain)
				}
				if written != int64(len(c.expected)) {
					t.Errorf("written mismatch\n\texpect: %d\n\tgot: %d", len(c.expected), written)
				}
				if err != nil {
					t.Errorf("error mismatch\n\texpect: %v\n\tgot: %v", nil, err)
				}
				if !handled {
					t.Errorf("handled mismatch\n\texpect: %v\n\tgot: %v", true, handled)
				}
				if f != tf {
					t.Errorf("file mismatch\n\texpect: %v\n\tgot: %v", tf, f)
				}
			}
			defer func() { sendFileTestHook = nil }()

			// Data goes from file to socket via sendfile(2).
			sent, err := io.Copy(server, io.NewSectionReader(tf, c.sectionOffset, c.sectionLength))
			if err != nil {
				t.Fatalf("failed to copy data with sendfile, error: %v", err)
			}
			if sent != int64(len(c.expected)) {
				t.Fatalf("bytes sent mismatch\n\texpect: %d\n\tgot: %d", len(c.expected), sent)
			}
			server.Close()

			<-doneCh

			if rerr != nil {
				t.Fatalf("error receiving data: %v", rerr)
			}
			if buf.String() != c.expected {
				t.Fatalf("received data mismatch\n\texpect: %q\n\tgot: %q", c.expected, buf.String())
			}

			tpos, err := tf.Seek(0, io.SeekCurrent)
			if err != nil {
				t.Fatalf("failed to seek file: %v", err)
			}
			if tpos != 3 {
				t.Fatalf("file position mismatch\n\texpect: %d\n\tgot: %d", 3, tpos)
			}
		})
	}
}
