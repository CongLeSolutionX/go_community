// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sort"
	"unicode"
)

func ExampleBuffer() {
	var b bytes.Buffer // A Buffer needs no initialization.
	b.Write([]byte("Hello "))
	fmt.Fprintf(&b, "world!")
	b.WriteTo(os.Stdout)
	// Output: Hello world!
}

func ExampleBuffer_reader() {
	// A Buffer can turn a string or a []byte into an io.Reader.
	buf := bytes.NewBufferString("R29waGVycyBydWxlIQ==")
	dec := base64.NewDecoder(base64.StdEncoding, buf)
	io.Copy(os.Stdout, dec)
	// Output: Gophers rule!
}

func ExampleBuffer_Grow() {
	var b bytes.Buffer
	b.Grow(64)
	bb := b.Bytes()
	b.Write([]byte("64 bytes or fewer"))
	fmt.Printf("%q", bb[:b.Len()])
	// Output: "64 bytes or fewer"
}

func ExampleCompare() {
	// Interpret Compare's result by comparing it to zero.
	var a, b []byte
	if bytes.Compare(a, b) < 0 {
		// a less b
	}
	if bytes.Compare(a, b) <= 0 {
		// a less or equal b
	}
	if bytes.Compare(a, b) > 0 {
		// a greater b
	}
	if bytes.Compare(a, b) >= 0 {
		// a greater or equal b
	}

	// Prefer Equal to Compare for equality comparisons.
	if bytes.Equal(a, b) {
		// a equal b
	}
	if !bytes.Equal(a, b) {
		// a not equal b
	}
}

func ExampleCompare_search() {
	// Binary search to find a matching byte slice.
	var needle []byte
	var haystack [][]byte // Assume sorted
	i := sort.Search(len(haystack), func(i int) bool {
		// Return haystack[i] >= needle.
		return bytes.Compare(haystack[i], needle) >= 0
	})
	if i < len(haystack) && bytes.Equal(haystack[i], needle) {
		// Found it!
	}
}

func ExampleTrimSuffix() {
	var b = []byte("Hello, goodbye, etc!")
	b = bytes.TrimSuffix(b, []byte("goodbye, etc!"))
	b = bytes.TrimSuffix(b, []byte("gopher"))
	b = append(b, bytes.TrimSuffix([]byte("world!"), []byte("x!"))...)
	os.Stdout.Write(b)
	// Output: Hello, world!
}

func ExampleTrimPrefix() {
	var b = []byte("Goodbye,, world!")
	b = bytes.TrimPrefix(b, []byte("Goodbye,"))
	b = bytes.TrimPrefix(b, []byte("See ya,"))
	fmt.Printf("Hello%s", b)
	// Output: Hello, world!
}

func ExampleFields() {
	fmt.Printf("Fields are: %q", bytes.Fields([]byte("  foo bar  baz   ")))
	// Output: Fields are: ["foo" "bar" "baz"]
}

func ExampleFieldsFunc() {
	f := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	fmt.Printf("Fields are: %q", bytes.FieldsFunc([]byte("  foo1;bar2,baz3..."), f))
	// Output: Fields are: ["foo1" "bar2" "baz3"]
}

func ExampleContains() {
	fmt.Println(bytes.Contains([]byte("seafood"), []byte("foo")))
	fmt.Println(bytes.Contains([]byte("seafood"), []byte("bar")))
	fmt.Println(bytes.Contains([]byte("seafood"), []byte("")))
	fmt.Println(bytes.Contains([]byte(""), []byte("")))
	// Output:
	// true
	// false
	// true
	// true
}

func ExampleContainsAny() {
	fmt.Println(bytes.ContainsAny([]byte("gopher"), "ph"))
	fmt.Println(bytes.ContainsAny([]byte("👍🕺🏾👍🏽👍🏿"), "👍🏽"))
	fmt.Println(bytes.ContainsAny([]byte("gopher"), "xyz"))
	// Output:
	// true
	// true
	// false
}

func ExampleContainsRune() {
	fmt.Println(bytes.ContainsRune([]byte("gopher"), 'g'))
	fmt.Println(bytes.ContainsRune([]byte("gopher"), 'x'))
	// Output:
	// true
	// false
}

func ExampleEqual() {
	fmt.Println(bytes.Equal([]byte("dog"), []byte("dog")))
	fmt.Println(bytes.Equal([]byte("dog"), []byte("cat")))
	fmt.Println(bytes.Equal([]byte("rehpog"), nil))
	fmt.Println(bytes.Equal(nil, nil))
	// Output:
	// true
	// false
	// false
	// true
}

func ExampleCount() {
	fmt.Println(bytes.Count([]byte("cheese"), []byte("e")))
	fmt.Println(bytes.Count([]byte("five"), []byte(""))) // before & after each rune
	// Output:
	// 3
	// 5
}

func ExampleEqualFold() {
	fmt.Println(bytes.EqualFold([]byte("Go"), []byte("go")))
	// Output: true
}

func ExampleHasPrefix() {
	fmt.Println(bytes.HasPrefix([]byte("Gopher"), []byte("Go")))
	fmt.Println(bytes.HasPrefix([]byte("Gopher"), []byte("C")))
	fmt.Println(bytes.HasPrefix([]byte("Gopher"), []byte("")))
	// Output:
	// true
	// false
	// true
}

func ExampleHasSuffix() {
	fmt.Println(bytes.HasSuffix([]byte("Amigo"), []byte("go")))
	fmt.Println(bytes.HasSuffix([]byte("Amigo"), []byte("O")))
	fmt.Println(bytes.HasSuffix([]byte("Amigo"), []byte("Ami")))
	fmt.Println(bytes.HasSuffix([]byte("Amigo"), []byte("")))
	// Output:
	// true
	// false
	// false
	// true
}

func ExampleIndex() {
	fmt.Println(bytes.Index([]byte("chicken"), []byte("ken")))
	fmt.Println(bytes.Index([]byte("chicken"), []byte("dmr")))
	// Output:
	// 4
	// -1
}

func ExampleIndexFunc() {
	f := func(c rune) bool {
		return unicode.Is(unicode.Han, c)
	}
	fmt.Println(bytes.IndexFunc([]byte("Hello, 世界"), f))
	fmt.Println(bytes.IndexFunc([]byte("Hello, world"), f))
	// Output:
	// 7
	// -1
}

func ExampleIndexAny() {
	fmt.Println(bytes.IndexAny([]byte("chicken"), "aeiouy"))
	fmt.Println(bytes.IndexAny([]byte("crwth"), "aeiouy"))
	// Output:
	// 2
	// -1
}

func ExampleIndexRune() {
	fmt.Println(bytes.IndexRune([]byte("chicken"), 'k'))
	fmt.Println(bytes.IndexRune([]byte("chicken"), 'd'))
	// Output:
	// 4
	// -1
}

func ExampleLastIndex() {
	fmt.Println(bytes.Index([]byte("go gopher"), []byte("go")))
	fmt.Println(bytes.LastIndex([]byte("go gopher"), []byte("go")))
	fmt.Println(bytes.LastIndex([]byte("go gopher"), []byte("rodent")))
	// Output:
	// 0
	// 3
	// -1
}

func ExampleJoin() {
	s := [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	fmt.Printf("%s", bytes.Join(s, []byte(", ")))
	// Output: foo, bar, baz
}

func ExampleRepeat() {
	fmt.Printf("ba%s", bytes.Repeat([]byte("na"), 2))
	// Output: banana
}

func ExampleReplace() {
	fmt.Printf("%s\n", bytes.Replace([]byte("oink oink oink"), []byte("k"), []byte("ky"), 2))
	fmt.Printf("%s\n", bytes.Replace([]byte("oink oink oink"), []byte("oink"), []byte("moo"), -1))
	// Output:
	// oinky oinky oink
	// moo moo moo
}

func ExampleSplit() {
	fmt.Printf("%q\n", bytes.Split([]byte("a,b,c"), []byte(",")))
	fmt.Printf("%q\n", bytes.Split([]byte("a man a plan a canal panama"), []byte("a ")))
	fmt.Printf("%q\n", bytes.Split([]byte(" xyz "), []byte("")))
	fmt.Printf("%q\n", bytes.Split([]byte(""), []byte("Bernardo O'Higgins")))
	// Output:
	// ["a" "b" "c"]
	// ["" "man " "plan " "canal panama"]
	// [" " "x" "y" "z" " "]
	// [""]
}

func ExampleIndexByte() {
	fmt.Println(bytes.IndexByte([]byte("dog"), []byte("d")[0]))
	fmt.Println(bytes.IndexByte([]byte("marsupial"), byte(117)))
	// Output:
	// 0
	// 4
}

func ExampleLastIndexAny() {
	fmt.Println(bytes.LastIndexAny([]byte("supercalifragilisticexpialidocious"), "c"))
	fmt.Println(bytes.LastIndexAny([]byte("Ti esrever dna ti pilf, nwod gniht ym tup"), "e"))
	fmt.Println(bytes.LastIndexAny([]byte("和小"), ""))
	// Output:
	// 29
	// 8
	// -1
}

func ExampleLastIndexByte() {
	fmt.Println(bytes.LastIndexByte([]byte("supercalifragilisticexpialidocious"), byte(99)))
	fmt.Println(bytes.LastIndexByte([]byte("supercalifragilisticexpialidocious"), byte(104)))
	// Output:
	// 29
	// -1
}

func ExamplelastIndexFunc() {
	b := []byte("holeinone")
	index := bytes.LastIndexFunc(b, func(r rune) bool {
		return r == 'e'
	})
	fmt.Println(index)
	b = []byte("holeinone")
	index = bytes.LastIndexFunc(b, func(r rune) bool {
		return r == 'z'
	})
	fmt.Println(index)
	// Output:
	// 8
	// -1
}

func ExampleSplitN() {
	fmt.Printf("%q\n", bytes.SplitN([]byte("a,b,c"), []byte(","), 2))
	z := bytes.SplitN([]byte("a,b,c"), []byte(","), 0)
	fmt.Printf("%q (nil = %v)\n", z, z == nil)
	// Output:
	// ["a" "b,c"]
	// [] (nil = true)
}

func ExampleSplitAfter() {
	fmt.Printf("%q\n", bytes.SplitAfter([]byte("a,b,c"), []byte(",")))
	// Output: ["a," "b," "c"]
}

func ExampleSplitAfterN() {
	fmt.Printf("%q\n", bytes.SplitAfterN([]byte("a,b,c"), []byte(","), 2))
	// Output: ["a," "b,c"]
}

func ExampleTitle() {
	fmt.Printf("%s", bytes.Title([]byte("her royal highness")))
	// Output: Her Royal Highness
}

func ExampleToTitle() {
	fmt.Printf("%s\n", bytes.ToTitle([]byte("loud noises")))
	fmt.Printf("%s\n", bytes.ToTitle([]byte("хлеб")))
	// Output:
	// LOUD NOISES
	// ХЛЕБ
}

func ExampleTrim() {
	fmt.Printf("[%q]", bytes.Trim([]byte(" !!! Achtung! Achtung! !!! "), "! "))
	// Output: ["Achtung! Achtung"]
}

func ExampleMap() {
	rot13 := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		}
		return r
	}
	fmt.Printf("%s", bytes.Map(rot13, []byte("'Twas brillig and the slithy gopher...")))
	// Output: 'Gjnf oevyyvt naq gur fyvgul tbcure...
}

func ExampleTrimSpace() {
	fmt.Printf("%s", bytes.TrimSpace([]byte(" \t\n a lone gopher \n\t\r\n")))
	// Output: a lone gopher
}

func ExampleToUpper() {
	fmt.Printf("%s", bytes.ToUpper([]byte("Gopher")))
	// Output: GOPHER
}

func ExampleToLower() {
	fmt.Printf("%s", bytes.ToLower([]byte("Gopher")))
	// Output: gopher
}
