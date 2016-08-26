// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,cgo

package plugin

/*
#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <limits.h>
#include <stdlib.h>
#include <stdint.h>

static uintptr_t pluginOpen(const char* path, char** err) {
	void* h = dlopen(path, RTLD_NOW|RTLD_GLOBAL);
	if (h == 0) {
		*err = dlerror();
	}
	return (uintptr_t)h;
}

static void* pluginLookup(uintptr_t h, const char* name, char** err) {
	void* r = dlsym((void*)h, name);
	if (r == 0) {
		*err = dlerror();
	}
	return r;
}
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"
)

func open(name string) (*Plugin, error) {
	pluginsMu.Lock()
	cRelName := C.CString(name)
	cPath := C.realpath(cRelName, nil)
	C.free(unsafe.Pointer(cRelName))
	defer C.free(unsafe.Pointer(cPath))
	path := C.GoString(cPath)
	if p, ok := plugins[path]; ok {
		pluginsMu.Unlock()
		if p == nil {
			return nil, errors.New("plugin.Open: plugin is loading: " + path)
		}
		return p, nil
	}
	if plugins == nil {
		plugins = make(map[string]*Plugin)
	}
	// This function can be called from the init function of a plugin.
	// Drop a placehodler in the map, so that a plugin cannot open an
	// opening plugin.
	plugins[path] = nil
	var cErr *C.char
	h := C.pluginOpen(cPath, &cErr)
	if h == 0 {
		return nil, errors.New("plugin.Open: " + C.GoString(cErr))
	}
	// TODO(crawshaw): look for plugin note, confirm it is a Go plugin
	// and it was built with the correct toolchain.
	// TODO(crawshaw): get full plugin name from note.
	syms := lastmoduleinit()
	pluginsMu.Unlock()

	if len(name) > 3 && name[len(name)-3:] == ".so" {
		name = name[:len(name)-3]
	}

	initStr := C.CString(name + ".init")
	initFuncPC := C.pluginLookup(h, initStr, &cErr)
	C.free(unsafe.Pointer(initStr))
	if initFuncPC != nil {
		initFuncP := &initFuncPC
		initFunc := *(*func())(unsafe.Pointer(&initFuncP))
		initFunc()
	}

	// Fill out the value of each plugin symbol.
	for symName, sym := range syms {
		isFunc := symName[0] == '.'
		if isFunc {
			delete(syms, symName)
			symName = symName[1:]
		}

		cname := C.CString(name + "." + symName)
		p := C.pluginLookup(h, cname, &cErr)
		C.free(unsafe.Pointer(cname))
		if p == nil {
			return nil, errors.New("plugin.Open: could not find symbol " + symName + ": " + C.GoString(cErr))
		}
		valp := (*[2]unsafe.Pointer)(unsafe.Pointer(&sym))
		if isFunc {
			(*valp)[1] = unsafe.Pointer(&p)
		} else {
			(*valp)[1] = p
		}
		syms[symName] = sym
	}
	p := &Plugin{
		name: name,
		syms: syms,
	}

	pluginsMu.Lock()
	plugins[path] = p
	pluginsMu.Unlock()

	return p, nil
}

func lookup(p *Plugin, symName string) (Symbol, error) {
	if s := p.syms[symName]; s != nil {
		return s, nil
	}
	return nil, errors.New("plugin: symbol " + symName + " not found in plugin " + p.name)
}

var (
	pluginsMu sync.Mutex
	plugins   map[string]*Plugin
)

func lastmoduleinit() map[string]interface{} // in package runtime
