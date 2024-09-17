// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package test

import (
	"errors"
	"syscall"
)

func isETXTBSY(err error) bool {
	return errors.Is(err, syscall.ETXTBSY)
}

func init() {
	// Record the initial umask for the cache.
	initialUmask = syscall.Umask(0)
	syscall.Umask(initialUmask)
}
