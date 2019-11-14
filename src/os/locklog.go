// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// When runtime lock logging is enabled, this connects to the lock
// logging server and initializes runtime lock logging.

package os

func runtime_lockLogInit()

func init() {
	sockPath := Getenv("GOLOCKLOG")
	if sockPath == "" {
		return
	}
	runtime_lockLogInit()
}
