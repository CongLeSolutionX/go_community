// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io_test

import (
	"bytes"
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

func ExampleTeeReader() {
	f := func(r io.Reader) {
		n, err := io.Copy(os.Stdout, r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("\n", n, "\n")
	}

	r := strings.NewReader("some io.Reader stream to be read")
	tbuf := bytes.NewBuffer([]byte{})
	tee := io.TeeReader(r, tbuf)

	f(tee)
	f(tbuf)

	// Output:
	// some io.Reader stream to be read
	// 32
	// some io.Reader stream to be read
	// 32
}

func ExampleSectionReader() {
	r := strings.NewReader("some io.Reader stream to be read")
	s := io.NewSectionReader(r, 5, 16)

	n, err := io.Copy(os.Stdout, s)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\n", n, s.Size(), "\n")

	// Output:
	// io.Reader stream
	// 16 16
}

func ExampleSectionReader_ReadAt() {
	r := strings.NewReader("some io.Reader stream to be read")
	s := io.NewSectionReader(r, 5, 16)

	tmp := make([]byte, 6)
	nReadAt, err := s.ReadAt(tmp, 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d %s\n", nReadAt, tmp)

	// Output:
	// 6 stream
}

func ExampleSectionReader_Seek() {
	r := strings.NewReader("some io.Reader stream to be read")
	s := io.NewSectionReader(r, 5, 16)

	rightBuf := make([]byte, 6)
	nSeek, err := s.Seek(10, 0)
	if err != nil {
		log.Fatal(err)
	}

	nRead, err := s.Read(rightBuf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d %d %s\n", nSeek, nRead, rightBuf)

	// Output:
	// 10 6 stream
}

func ExampleMultiWriter() {
	r := strings.NewReader("some io.Reader stream to be read")

	buf1 := new(bytes.Buffer)
	buf2 := new(bytes.Buffer)

	w := io.MultiWriter(buf1, buf2)
	n, err := io.Copy(w, r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(buf1.String())
	fmt.Println(buf2.String())
	fmt.Println(n)

	// Output:
	// some io.Reader stream to be read
	// some io.Reader stream to be read
	// 32

}
