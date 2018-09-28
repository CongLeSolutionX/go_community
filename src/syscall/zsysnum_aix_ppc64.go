// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ppc64,aix

package syscall

// Those constants are defined for compatibility purpose, but are meaningless
const (
	SYS_EXECVE = 0
	SYS_FCNTL  = 0
)
