// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret

func Do(f func()) {
	inc()
	defer dec()
	f()
}

func Enabled() bool {
	return count() > 0
}

// implemented in runtime
func count() int64
func inc()
func dec()
