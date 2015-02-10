// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_vdso "runtime/internal/vdso"
	"unsafe"
)

// nosplit for use in linux/386 startup linux_setup_vdso
//go:nosplit
func argv_index(argv **byte, i int32) *byte {
	return *(**byte)(_core.Add(unsafe.Pointer(argv), uintptr(i)*_core.PtrSize))
}

// Information about what cpu features are available.
// Set on startup in asm_{x86/amd64}.s.
var (
//cpuid_ecx uint32
//cpuid_edx uint32
)

func goargs() {
	if _lock.GOOS == "windows" {
		return
	}

	Argslice = make([]string, _vdso.Argc)
	for i := int32(0); i < _vdso.Argc; i++ {
		Argslice[i] = _lock.Gostringnocopy(argv_index(_vdso.Argv, i))
	}
}

func goenvs_unix() {
	// TODO(austin): ppc64 in dynamic linking mode doesn't
	// guarantee env[] will immediately follow argv.  Might cause
	// problems.
	n := int32(0)
	for argv_index(_vdso.Argv, _vdso.Argc+1+n) != nil {
		n++
	}

	Envs = make([]string, n)
	for i := int32(0); i < n; i++ {
		Envs[i] = _lock.Gostringnocopy(argv_index(_vdso.Argv, _vdso.Argc+1+i))
	}
}

func environ() []string {
	return Envs
}

type dbgVar struct {
	name  string
	value *int32
}

var dbgvars = []dbgVar{
	{"allocfreetrace", &_lock.Debug.Allocfreetrace},
	{"efence", &_lock.Debug.Efence},
	{"gcdead", &_lock.Debug.Gcdead},
	{"gctrace", &_lock.Debug.Gctrace},
	{"invalidptr", &_lock.Debug.Invalidptr},
	{"scavenge", &_lock.Debug.Scavenge},
	{"scheddetail", &_lock.Debug.Scheddetail},
	{"schedtrace", &_lock.Debug.Schedtrace},
	{"wbshadow", &_lock.Debug.Wbshadow},
	{"gccheckmark", &_lock.Debug.Gccheckmark},
}

func parsedebugvars() {
	// gccheckmark is enabled by default for the 1.5 dev cycle
	_lock.Debug.Gccheckmark = 1

	for p := Gogetenv("GODEBUG"); p != ""; {
		field := ""
		i := _lock.Index(p, ",")
		if i < 0 {
			field, p = p, ""
		} else {
			field, p = p[:i], p[i+1:]
		}
		i = _lock.Index(field, "=")
		if i < 0 {
			continue
		}
		key, value := field[:i], field[i+1:]

		// Update MemProfileRate directly here since it
		// is int, not int32, and should only be updated
		// if specified in GODEBUG.
		if key == "memprofilerate" {
			_lock.MemProfileRate = atoi(value)
		} else {
			for _, v := range dbgvars {
				if v.name == key {
					*v.value = int32(atoi(value))
				}
			}
		}
	}

	switch p := Gogetenv("GOTRACEBACK"); p {
	case "":
		_lock.Traceback_cache = 1 << 1
	case "crash":
		_lock.Traceback_cache = 2<<1 | 1
	default:
		_lock.Traceback_cache = uint32(atoi(p)) << 1
	}
}

// TODO: move back into mgc0.c when converted to Go
func readgogc() int32 {
	p := Gogetenv("GOGC")
	if p == "" {
		return 100
	}
	if p == "off" {
		return -1
	}
	return int32(atoi(p))
}
