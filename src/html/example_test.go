// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html_test

import (
	"fmt"
	"html"
)

func Example() {

	unencoded := `The special characters are: <, >, &, ' and "`
	encoded := html.EscapeString(unencoded)

	if unencoded == html.UnescapeString(encoded) {
		fmt.Println("UnescapeString(EscapeString(s)) == s - always true")
	}

	// Output: UnescapeString(EscapeString(s)) == s - always true
}

func ExampleEscapeString() {
	unencoded := `The special characters are: <, >, &, ' and "`
	fmt.Println(html.EscapeString(unencoded))
	// Output: The special characters are: &lt;, &gt;, &amp;, &#39; and &#34;
}

func ExampleUnescapeString() {
	encoded := `The encoded characters are: &lt;, &gt;, &amp;, &#39; and &#34;`
	fmt.Println(html.UnescapeString(encoded))
	// Output: The encoded characters are: <, >, &, ' and "
}
