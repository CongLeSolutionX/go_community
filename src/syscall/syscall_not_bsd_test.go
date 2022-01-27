// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !darwin && !dragonfly && !freebsd && !openbsd

package syscall_test

func MaxFileLimit() (uint32, error) {
	return 0, nil
}
