// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gzip_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"os"
)

func ExampleNewWriter() {
	wc := gzip.NewWriter(os.Stdout)
	defer wc.Close()
	wc.Write([]byte("hello, world\n"))
}

func ExampleNewReader() {
	buf := []byte{31, 139, 8, 0, 0, 9, 110, 136, 0, 255, 202, 72, 205, 201, 201, 215, 81, 40, 207, 47, 202, 73, 225, 2, 4, 0, 0, 255, 255, 83, 116, 36, 244, 13, 0, 0, 0}
	b := bytes.NewReader(buf)
	rc, err := gzip.NewReader(b)
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()

	io.Copy(os.Stdout, rc)
}

func ExampleWriter_Reset() {
	var buf bytes.Buffer
	wc := gzip.NewWriter(&buf)
	defer wc.Close()

	wc.Write([]byte("hello, world\n"))

	wc.Reset(os.Stdout)

	wc.Write([]byte("Gophers\n"))
}

func ExampleReader_Reset() {
	buf1 := []byte{31, 139, 8, 0, 0, 9, 110, 136, 0, 255, 202, 72, 205, 201, 201, 215, 81, 40, 207, 47, 202, 73, 225, 2, 4, 0, 0, 255, 255, 83, 116, 36, 244, 13, 0, 0, 0}

	rc, err := gzip.NewReader(bytes.NewReader(buf1))
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()

	io.Copy(os.Stdout, rc)

	buf2 := []byte{31, 139, 8, 0, 0, 9, 110, 136, 0, 255, 114, 207, 47, 200, 72, 45, 42, 230, 2, 4, 0, 0, 255, 255, 248, 97, 174, 38, 8, 0, 0, 0}
	b2 := bytes.NewReader(buf2)
	if err := rc.Reset(b2); err != nil {
		log.Fatal(err)
	}

	io.Copy(os.Stdout, rc)
}
