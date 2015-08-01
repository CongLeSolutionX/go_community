// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

// An errorString represents a runtime error described by a single string.
type ErrorString string

func (e ErrorString) RuntimeError() {}

func (e ErrorString) Error() string {
	return "runtime error: " + string(e)
}
