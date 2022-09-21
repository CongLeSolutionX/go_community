// Package buildinternal provides internal functions used by go/build
// that need to be used by other packages too.
package buildinternal

var IsShared bool // Set by cmd/go if applicable.

// NeedsInstalledDotA returns true if the given stdlib package
// needs an installed .a file in the stdlib.
func NeedsInstalledDotA(importPath string) bool {
	if IsShared {
		return true
	}
	return importPath == "net" || importPath == "os/signal" || importPath == "os/user" || importPath == "plugin" ||
		importPath == "runtime/cgo"
}
