// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "os"

var _ = os.Open // avoid imported and not used error

type T struct {
	File int
}

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	_ = T {
		os.File: 1, // ERROR "unknown T? ?field|invalid field"
=======
	_ = T{
		os.File: 1, // ERROR "invalid field name os.File|unknown field"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	}
}
