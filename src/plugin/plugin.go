// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,cgo

// Package plugin implements loading and symbol resolution of Go plugins.
package plugin

/*
#cgo linux LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include <stdint.h>

uintptr_t pluginOpen(const char* name) {
	void* h = dlopen(name, RTLD_NOW|RTLD_GLOBAL);
	return (uintptr_t)h;
}

void* pluginLookup(uintptr_t h, const char* name) {
	return dlsym((void*)h, name);
}
*/
import "C"
import (
	"errors"
	"os"
	"sync"
	"unsafe"
)

// Plugin is a loaded Go plugin.
type Plugin struct {
	name string
	syms map[string]interface{}
}

// Lookup searches for a symbol named symName in plugin p.
// It reports an error if the symbol is not found.
func (p *Plugin) Lookup(symName string) (interface{}, error) {
	if s := p.syms[symName]; s != nil {
		return s, nil
	}
	return nil, errors.New("plugin: symbol " + symName + " not found in plugin " + p.name)
}

// Open opens a Go plugin.
func Open(name string) (*Plugin, error) {
	i := len(name) - 1
	for i >= 0 && !os.IsPathSeparator(name[i]) {
		i--
	}
	if i >= 0 {
		name = name[i+1:]
	}
	name = trimSuffix(name, ".so")

	pluginsMu.Lock()
	defer pluginsMu.Unlock()

	if p := plugins[name]; p != nil {
		return p, nil
	}

	cname := C.CString(name)
	h := C.pluginOpen(cname)
	C.free(unsafe.Pointer(cname))
	if h == 0 {
		return nil, errors.New("plugin.Open: " + C.GoString(C.dlerror()))
	}

	// TODO(crawshaw): look for plugin note, confirm it is a Go plugin
	// and it was built with the correct toolchain.

	syms := lastmoduleinit()

	p := &Plugin{
		name: name,
		syms: syms,
	}
	if plugins == nil {
		plugins = make(map[string]*Plugin)
	}
	plugins[name] = p

	initStr := C.CString(name + ".init")
	initFuncPC := C.pluginLookup(h, initStr)
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
		p := C.pluginLookup(h, cname)
		C.free(unsafe.Pointer(cname))
		valp := (*[2]unsafe.Pointer)(unsafe.Pointer(&sym))
		if isFunc {
			(*valp)[1] = unsafe.Pointer(&p)
		} else {
			(*valp)[1] = p
		}
		syms[symName] = sym
	}

	return p, nil
}

func trimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}

var (
	pluginsMu sync.Mutex
	plugins   map[string]*Plugin
)

func lastmoduleinit() map[string]interface{} // in package runtime
