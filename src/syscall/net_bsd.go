// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package syscall

// An IFNet represents mixed information consists of network stack and
// device driver on BSD variants.
type IFNet struct {
	Flags int // see IFF_* in golang.org/x/sys/unix
}

// Next implements the Next method of NetworkInterface interface.
func (*IFNet) Next(ifindex int) NetworkInterface {
	return nil
}
