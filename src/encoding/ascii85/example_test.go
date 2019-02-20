// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ascii85_test

import (
	"bytes"
	"encoding/ascii85"
	"fmt"
)

func ExampleEncode() {
	// plain message
	src := []byte("This is simple text for encoding.")
	buffer := make([]byte, ascii85.MaxEncodedLen(len(src)))
	ascii85.Encode(buffer, src)

	fmt.Println(string(buffer))
	// Output:
	// <+oue+DGm>F(oK1Ch4`2AU&;>AoD]4ASu!rA8,po/cYkO

}

func ExampleDecode() {
	// encoded message
	src := "<+oue+DGm>F(oK1Ch4`2AU&;>AoD]4ASu!rA8,po/cYkO"

	buffer := make([]byte, len(src))
	_, _, err := ascii85.Decode(buffer, []byte(src), true)

	if err != nil {
		fmt.Println(err)
	}
	// resize bytes buffer
	b := bytes.Trim(buffer, "\x00")
	fmt.Println(string(b))

	// Output:
	// This is simple text for encoding.
}
