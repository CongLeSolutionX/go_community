// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gzip_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"
)

func ExampleWriterReader() {
	var buf bytes.Buffer
	wc := gzip.NewWriter(&buf)
	wc.Header.Comment = "gzip example"
	wc.Header.Name = "example-file"
	wc.Header.ModTime = time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)

	if err := io.WriteString(wc, "Hello, Gophers!"); err != nil {
		log.Fatal(err)
	}

	if err := wc.Close(); err != nil {
		log.Fatal(err)
	}

	rc, err := gzip.NewReader(&buf)
	if err != nil {
		log.Fatal(err)
	}

	slurp, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Decompressed: %s\n", slurp)

	if err := rc.Close(); err != nil {
		log.Fatal(err)
	}

	header := rc.Header
	fmt.Printf("Header fields:\nComment: %s\nName: %s\nModTime: %s\n", header.Comment, header.Name, header.ModTime.UTC())

	// Output:
	// Decompressed: Hello, Gophers!
	// Header fields:
	// Comment: gzip example
	// Name: example-file
	// ModTime: 2006-02-01 03:04:05 +0000 UTC
}

func ExampleReader_Multistream() {
	var buf bytes.Buffer
	wc := gzip.NewWriter(&buf)
	wc.Header.Comment = "file-header-1"
	wc.Header.Name = "file-1"
	wc.Header.ModTime = time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)

	if _, err := io.WriteString(wc, "Hello Gophers - 1"); err != nil {
		log.Fatal(err)
	}

	if err := wc.Close(); err != nil {
		log.Fatal(err)
	}

	if err := wc.Reset(&buf); err != nil {
		log.Fatal(err)
	}

	wc.Header.Comment = "file-header-2"
	wc.Header.Name = "file-2"
	wc.Header.ModTime = time.Date(2007, time.March, 2, 4, 5, 6, 1, time.UTC)

	if _, err := io.WriteString(wc, "Hello Gophers - 2"); err != nil {
		log.Fatal(err)
	}

	if err := wc.Close(); err != nil {
		log.Fatal(err)
	}

	rc, err := gzip.NewReader(&buf)
	if err != nil {
		log.Fatal(err)
	}

	rc.Multistream(false)
	file1, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("File 1 Decompressed: %s\n", file1)

	header1 := rc.Header
	fmt.Printf("Header1 fields:\nComment: %s\nName: %s\nModTime: %s\n", header1.Comment, header1.Name, header1.ModTime.UTC())

	if err := rc.Reset(&buf); err != nil {
		log.Fatal(err)
	}

	rc.Multistream(false)
	file2, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("File 2 Decompressed: %s\n", file2)
	if err := rc.Close(); err != nil {
		log.Fatal(err)
	}

	header2 := rc.Header
	fmt.Printf("Header2 fields:\nComment: %s\nName: %s\nModTime: %s\n", header2.Comment, header2.Name, header2.ModTime.UTC())

	// Output:
	// File 1 Decompressed: Hello Gophers - 1
	// Header1 fields:
	// Comment: file-header-1
	// Name: file-1
	// ModTime: 2006-02-01 03:04:05 +0000 UTC
	// File 2 Decompressed: Hello Gophers - 2
	// Header2 fields:
	// Comment: file-header-2
	// Name: file-2
	// ModTime: 2007-03-02 04:05:06 +0000 UTC
}
