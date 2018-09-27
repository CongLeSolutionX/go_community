// runoutput

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test source file beginning with a byte order mark.

package main

import (
	"fmt"
	"strings"
)

func main() {
	prog = strings.ReplaceAll(prog, "BOM", "\uFEFF")
	fmt.Print(prog)
}

var prog = `BOM
package main

func main() {
}
`
