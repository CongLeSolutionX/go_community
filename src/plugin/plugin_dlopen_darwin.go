// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plugin

import (
	"internal/syscall/unix"
	"unsafe"
)

type _C_char = byte
type _C_uintptr_t = uintptr

const _C_PATH_MAX = unix.PATH_MAX

func _C_GoString(p *_C_char) string { return unix.GoString(p) }

func _C_pluginOpen(path *_C_char, err **_C_char) _C_uintptr_t {
	return unix.Dlopen(path, unix.RTLD_NOW|unix.RTLD_GLOBAL, err)
}

func _C_pluginLookup(h _C_uintptr_t, name *_C_char, err **_C_char) unsafe.Pointer {
	return unix.Dlsym(h, name, err)
}

func _C_realpath(old, new *_C_char) *_C_char {
	p, _ := unix.Realpath(old, new)
	return p
}
