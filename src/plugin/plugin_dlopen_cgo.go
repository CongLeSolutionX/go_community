// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (linux && cgo) || (freebsd && cgo)

package plugin

/*
#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <limits.h>
#include <stdlib.h>
#include <stdint.h>

#include <stdio.h>

static uintptr_t pluginOpen(const char* path, char** err) {
	void* h = dlopen(path, RTLD_NOW|RTLD_GLOBAL);
	if (h == NULL) {
		*err = (char*)dlerror();
	}
	return (uintptr_t)h;
}

static void* pluginLookup(uintptr_t h, const char* name, char** err) {
	void* r = dlsym((void*)h, name);
	if (r == NULL) {
		*err = (char*)dlerror();
	}
	return r;
}
*/
import "C"
import "unsafe"

type _C_char = C.char
type _C_uintptr_t = C.uintptr_t

const _C_PATH_MAX = C.PATH_MAX

func _C_GoString(p *_C_char) string { return C.GoString(p) }

func _C_pluginOpen(path *_C_char, err **_C_char) _C_uintptr_t {
	return C.pluginOpen(path, err)
}

func _C_pluginLookup(h _C_uintptr_t, name *_C_char, err **_C_char) unsafe.Pointer {
	return C.pluginLookup(h, name, err)
}

func _C_realpath(old, new *_C_char) *_C_char {
	return C.realpath(old, new)
}
