// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import "io"

func genericReadFrom(f *File, r io.Reader) (int64, error) {
	return io.Copy(onlyWriter{f}, r)
}

type onlyWriter struct {
	io.Writer
}
