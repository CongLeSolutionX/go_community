// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package filebuf

import (
	"bytes"
	"io"
)

func get(n int) io.Reader {
	if n <= len(contents) {
		return bytes.NewReader(contents[:n])
	}
	return bytes.NewReader(contents)
}
