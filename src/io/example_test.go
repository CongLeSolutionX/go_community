// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func ExampleCopy() {
	r := strings.NewReader("some io.Reader stream to be read")

	n, err := io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\n", n)

	// Output:
	// some io.Reader stream to be read
	// 32
}

func ExampleCopyBuffer() {
	r := strings.NewReader("some io.Reader stream to be read")

	buf := make([]byte, 8)
	n, err := io.CopyBuffer(os.Stdout, r, buf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\n", n)

	// Output:
	// some io.Reader stream to be read
	// 32
}

func ExampleCopyN() {
	r := strings.NewReader("some io.Reader stream to be read")

	n, err := io.CopyN(os.Stdout, r, 4)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\n", n)

	// Output:
	// some
	// 4
}

func ExampleReadAtLeast() {
	r := strings.NewReader("some io.Reader stream to be read")

	buf := make([]byte, 32)
	n, err := io.ReadAtLeast(r, buf, 4)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s (%d)\n", buf, n)

	// buffer smaller than minimal read size.
	shortBuf := make([]byte, 3)
	_, errsb := io.ReadAtLeast(r, shortBuf, 4)
	if errsb != nil {
		fmt.Println(errsb)
	}

	// minimal read size bigger than io.Reader stream
	longBuf := make([]byte, 64)
	_, errsr := io.ReadAtLeast(r, longBuf, 64)
	if errsr != nil {
		fmt.Println(errsr)
	}

	// Output:
	// some io.Reader stream to be read (32)
	// short buffer
	// EOF
}

func ExampleReadFull() {
	r := strings.NewReader("some io.Reader stream to be read")

	buf := make([]byte, 4)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s (%d)\n", buf, n)

	// minimal read size bigger than io.Reader stream
	longBuf := make([]byte, 64)
	_, errsr := io.ReadFull(r, longBuf)
	if errsr != nil {
		fmt.Println(errsr)
	}

	// Output:
	// some (4)
	// unexpected EOF
}

func ExampleWriteString() {
	io.WriteString(os.Stdout, "Hello World")

	// Output: Hello World
}

func ExampleLimitReader() {
	r := strings.NewReader("some io.Reader stream to be read")
	buf := make([]byte, 4)

	lmtR := io.LimitReader(r, 4)
	for {
		n, err := lmtR.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s (%d)\n", buf, n)
	}

	// Output:
	// some (4)
}

func ExampleMultiReader() {
	r1 := strings.NewReader("some ")
	r2 := strings.NewReader("io.Reader stream ")
	r3 := strings.NewReader("to be read")

	r := io.MultiReader(r1, r2, r3)
	n, err := io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\n", n)

	// Output:
	// some io.Reader stream to be read
	// 32
}
