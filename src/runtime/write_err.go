// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !android

package runtime

// Precondition: len(b) > 0.
func writeErr(b []byte) {
	writeErrData(&b[0], int32(len(b)))
}
