// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package syscall

// An IFNet represents mixed information consists of network stack and
// device driver on BSD variants.
type IFNet struct {
	Type  int // see IFT_* in golang.org/x/sys/unix package
	Flags int // see IFF_* in golang.org/x/sys/unix package
}
