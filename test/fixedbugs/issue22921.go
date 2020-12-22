// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "bytes"

type _ struct{ bytes.nonexist } // ERROR "unexported|undefined"

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
type _ interface{ bytes.nonexist } // ERROR "unexported|undefined"
=======
type _ interface{ bytes.nonexist } // ERROR "unexported|undefined|expected signature or type name"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

func main() {
	var _ bytes.Buffer
	var _ bytes.buffer // ERROR "unexported|undefined"
}
