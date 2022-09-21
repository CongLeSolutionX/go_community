// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package buildinternal provides internal functions used by go/build
// that need to be used by other packages too.
package buildinternal

// IsShared is used to give go/build access to information of whether
// the configuration we're running in is buildshared or linkshared.
var IsShared bool

// NeedsInstalledDotA returns true if the given stdlib package
// needs an installed .a file in the stdlib.
func NeedsInstalledDotA(importPath string) bool {
	if IsShared {
		return importPath != "unsafe"
	}
	return importPath == "net" || importPath == "os/signal" || importPath == "os/user" || importPath == "plugin" ||
		importPath == "runtime/cgo"
}
