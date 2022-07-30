// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/syscall/unix"
	"syscall"
	"unsafe"
)

// devname returns the name of the block or character device in /dev with a device number of dev.
// To find the right name, devname asks the kernel via the "kern.devname" sysctl.
func devname(dev uint64) (string, error) {
	mib, nerr := unix.SysctlNameToMIB([]byte("kern.devname"))
	if nerr != nil {
		return "", nerr
	}
	miblen := uintptr(len(mib))

	// get name length
	n := uintptr(0)
	_, _, err := syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), miblen, 0, uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&dev)), unsafe.Sizeof(dev))
	if err != 0 {
		return "", err
	}
	if n == 0 {
		return "", nil // shouldn't happen
	}
	// get name
	buf := make([]byte, n)
	_, _, err = syscall.Syscall6(syscall.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), miblen, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&dev)), unsafe.Sizeof(dev))
	if err != 0 {
		return "", err
	}
	if n == 0 {
		return "", nil // shouldn't happen
	}
	return string(buf[:n-1]), nil
}
