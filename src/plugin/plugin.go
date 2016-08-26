// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package plugin implements loading and symbol resolution of Go plugins.
//
// A plugin is a Go main package with exported functions and variables that
// has been built with
//
//	go build -buildmode=plugin.
//
// When a plugin is first opened, the init functions of all packages not
// already part of the program are called. The main function is not run.
// A plugin is only initialized once, and cannot be closed.
package plugin

// Plugin is a loaded Go plugin.
type Plugin struct {
	name string
	syms map[string]interface{}
}

// Open opens a Go plugin.
func Open(path string) (*Plugin, error) {
	return open(path)
}

// Lookup searches for a symbol named symName in plugin p.
// It reports an error if the symbol is not found.
func (p *Plugin) Lookup(symName string) (interface{}, error) {
	return lookup(p, symName)
}
