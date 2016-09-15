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

	// Setting the Header fields is optional.
	wc.Name = "a-new-hope.txt"
	wc.Comment = "an epic space opera by George Lucas"
	wc.ModTime = time.Date(1977, time.May, 25, 0, 0, 0, 0, time.UTC)

	_, err := io.WriteString(wc, "A long time ago in a galaxy far, far away...")
	if err != nil {
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

	if err := rc.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s\nComment: %s\nModTime: %s\n", rc.Name, rc.Comment, rc.ModTime.UTC())
	fmt.Printf("Data: %s\n", slurp)

	// Output:
	// Name: a-new-hope.txt
	// Comment: an epic space opera by George Lucas
	// ModTime: 1977-05-25 00:00:00 +0000 UTC
	// Data: A long time ago in a galaxy far, far away...
}

func ExampleReader_Multistream() {
	var buf bytes.Buffer
	wc := gzip.NewWriter(&buf)

	// Setting the Header fields is optional.
	wc.Name = "file-1.txt"
	wc.Comment = "file-header-1"
	wc.ModTime = time.Date(2006, time.February, 1, 3, 4, 5, 0, time.UTC)

	if _, err := io.WriteString(wc, "Hello Gophers - 1"); err != nil {
		log.Fatal(err)
	}

	if err := wc.Close(); err != nil {
		log.Fatal(err)
	}

	wc.Reset(&buf)

	wc.Name = "file-2.txt"
	wc.Comment = "file-header-2"
	wc.ModTime = time.Date(2007, time.March, 2, 4, 5, 6, 1, time.UTC)

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
	defer rc.Close()

	for {
		rc.Multistream(false)
		file, err := ioutil.ReadAll(rc)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Name: %s\nComment: %s\nModTime: %s\nData: %s\n\n", rc.Name, rc.Comment, rc.ModTime.UTC(), file)

		err = rc.Reset(&buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	// Output:
	// Name: file-1.txt
	// Comment: file-header-1
	// ModTime: 2006-02-01 03:04:05 +0000 UTC
	// Data: Hello Gophers - 1
	//
	// Name: file-2.txt
	// Comment: file-header-2
	// ModTime: 2007-03-02 04:05:06 +0000 UTC
	// Data: Hello Gophers - 2
}
