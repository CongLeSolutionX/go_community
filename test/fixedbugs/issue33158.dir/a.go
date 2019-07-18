// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

import "os"

func M() string {
	if s := os.Getenv("Fred"); s != "" {
		return s
	}
	if s := os.Getenv("Joe"); s != "" {
		return s
	}

	return string("Alex")
}
