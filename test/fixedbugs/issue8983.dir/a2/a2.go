// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a2

type Closer interface {
	Close() error
}

func NilCloser() Closer {
	return nil
}
