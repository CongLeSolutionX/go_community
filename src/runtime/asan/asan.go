// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build asan && linux && (arm64 || amd64)
// +build asan
// +build linux
// +build arm64 amd64

package asan

/*
#cgo CFLAGS: -fsanitize=address
#cgo LDFLAGS: -fsanitize=address

#include <stdint.h>
#include <sanitizer/asan_interface.h>

void __asan_read_go(void *addr, uintptr_t sz) {
	if (__asan_region_is_poisoned(addr, sz)) {
		switch (sz) {
		case 1: __asan_report_load1(addr); break;
		case 2: __asan_report_load2(addr); break;
		case 4: __asan_report_load4(addr); break;
		case 8: __asan_report_load8(addr); break;
		default: __asan_report_load_n(addr, sz); break;
		}
	}
}

void __asan_write_go(void *addr, uintptr_t sz) {
	if (__asan_region_is_poisoned(addr, sz)) {
		switch (sz) {
		case 1: __asan_report_store1(addr); break;
		case 2: __asan_report_store2(addr); break;
		case 4: __asan_report_store4(addr); break;
		case 8: __asan_report_store8(addr); break;
		default: __asan_report_store_n(addr, sz); break;
		}
	}
}

void __asan_unpoison_go(void *addr, uintptr_t sz) {
	__asan_unpoison_memory_region(addr, sz);
}

void __asan_poison_go(void *addr, uintptr_t sz) {
	__asan_poison_memory_region(addr, sz);
}

// Keep in sync with the defination in compiler-rt
// https://github.com/llvm-mirror/compiler-rt/blob/69445f095c22aac2388f939bedebf224a6efcdaf/lib/asan/asan_interface_internal.h#L41
// This structure is used to describe the source location of
// a place where global was defined.
struct _asan_global_source_location {
	const char *filename;
	int line_no;
	int column_no;
};

// Keep in sync with the defination in compiler-rt
// https://github.com/llvm-mirror/compiler-rt/blob/69445f095c22aac2388f939bedebf224a6efcdaf/lib/asan/asan_interface_internal.h#L48
// This structure describes an instrumented global variable.
struct _asan_global {
	uintptr_t beg;
	uintptr_t size;
	uintptr_t size_with_redzone;
	const char *name;
	const char *module_name;
	uintptr_t has_dynamic_init;
	struct _asan_global_source_location *location;
	uintptr_t odr_indicator;
};

// Register global variables.
// The 'globals' is an array of structures describing 'n' globals.
void __asan_register_globals_go(void *addr, uintptr_t n) {
	struct _asan_global *globals = (struct _asan_global *)(addr);
	__asan_register_globals(globals, n);
}
*/
import "C"
