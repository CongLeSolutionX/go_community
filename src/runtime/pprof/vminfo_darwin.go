// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"unsafe"
)

func isExecutable(protection int32) bool {
	return (protection&_VM_PROT_EXECUTE) != 0 && (protection&_VM_PROT_READ) != 0
}

func isWritable(protection int32) bool {
	return (protection & _VM_PROT_WRITE) != 0
}

// machVMInfo uses the mach_vm_region region system call to add mapping entries
// for the text region of the running process.
func machVMInfo(addMapping func(lo, hi, offset uint64, file, buildID string)) bool {
	var addr uint64 = 0x1
	var page_size uint64 = 0x0
	var info machVMRegionBasicInfoData

	// Get the first address and page size.
	kr := mach_vm_region(
		uintptr(unsafe.Pointer(&addr)),
		uintptr(unsafe.Pointer(&page_size)),
		uintptr(unsafe.Pointer(&info)))
	if kr != 0 {
		return false
	}

	prev_addr := addr
	prev_page_size := page_size
	prev_info := info

	for {
		addr = prev_addr + prev_page_size
		kr := mach_vm_region(
			uintptr(unsafe.Pointer(&addr)),
			uintptr(unsafe.Pointer(&page_size)),
			uintptr(unsafe.Pointer(&info)))
		if kr != 0 {
			if kr == _MACH_SEND_INVALID_DEST {
				// No more memory regions.
				return true
			}
			// Unexpected error.
			return false
		}
		if !isWritable(prev_info.Protection) && isExecutable(prev_info.Protection) {
			addMapping(prev_addr, addr, uint64(prev_info.Offset), "", "")
		}
		prev_addr = addr
		prev_page_size = page_size
		prev_info = info
	}
}

// mach_vm_region is implemented the runtime package (runtime/sys_darwin.go).
func mach_vm_region(address, page_size, info uintptr) int32
