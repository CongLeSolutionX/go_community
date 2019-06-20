// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha512_test

import (
	"crypto/sha512"
	"fmt"
	"io"
	"log"
	"os"
)

func ExampleNew() {
	h := sha512.New()
	h.Write([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", h.Sum(nil))
	// Output: 020da0f4d8a4c8bfbc98274027740061d7df52ee07091ed6595a083e0f45327bbe59424312d86f218b74ed2e25507abaf5c7a5fcf4cafcf9538b705808fd55ec
}

func ExampleNew_file() {
	f, err := os.Open("file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x", h.Sum(nil))
}

func ExampleNew384() {
	h := sha512.New384()
	h.Write([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", h.Sum(nil))
	// Output: d51d28d0141e56f692952ea14861898e2b417b922831e0f4bcdbc326a7fe1e9d9563182e83d3a8af66f68536e0d42b88
}

func ExampleNew512_224() {
	h := sha512.New512_224()
	h.Write([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", h.Sum(nil))
	// Output: 4d90ec85475853bc495a3243d13e664a3af0804705cee3e07edf741b
}

func ExampleNew512_256() {
	h := sha512.New512_256()
	h.Write([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", h.Sum(nil))
	// Output: 862d8c337f9d62ac89aa83d7ffbc2246ed54965684d0877beaf21e9aa7c44852
}

func ExampleSum384() {
	sum := sha512.Sum384([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", sum)
	// Output: d51d28d0141e56f692952ea14861898e2b417b922831e0f4bcdbc326a7fe1e9d9563182e83d3a8af66f68536e0d42b88
}

func ExampleSum512() {
	sum := sha512.Sum512([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", sum)
	// Output: 020da0f4d8a4c8bfbc98274027740061d7df52ee07091ed6595a083e0f45327bbe59424312d86f218b74ed2e25507abaf5c7a5fcf4cafcf9538b705808fd55ec
}

func ExampleSum512_224() {
	sum := sha512.Sum512_224([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", sum)
	// Output: 4d90ec85475853bc495a3243d13e664a3af0804705cee3e07edf741b
}

func ExampleSum512_256() {
	sum := sha512.Sum512_256([]byte("The quick brown fox jumps over the lazy dog.\n"))
	fmt.Printf("%x", sum)
	// Output: 862d8c337f9d62ac89aa83d7ffbc2246ed54965684d0877beaf21e9aa7c44852
}
