// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

// Building the runtime package with coverage instrumentation enabled
// is tricky.  For all other packages, you can be guaranteed that
// the package init function is run before any functions are executed,
// but this invariant is not maintained for packages such as "runtime",
// "internal/cpu", etc. To handle this, hard-code the package ID for
// the set of packages whose functions may be running before the
// init function of the package is complete.
//
// Hardcoding is unfortunate because it means that the tool that does
// coverage instrumentation has to keep a list of runtime packages,
// meaning that if someone makes changes to the pkg "runtime"
// dependences, an insanity will result for coverage builds. The
// coverage runtime will detect and report the insanity; look for an
// error of this form:
//
//    internal error in coverage meta-data tracking:
//    list of hard-coded runtime package IDs needs revising.
//    registered list:
//    slot: 0 path='internal/cpu'  hard-coded id: 1
//    slot: 1 path='internal/goarch'  hard-coded id: 2
//    slot: 2 path='runtime/internal/atomic'  hard-coded id: 3
//    slot: 3 path='internal/goos'
//    slot: 4 path='runtime/internal/sys'  hard-coded id: 5
//    slot: 5 path='internal/abi'  hard-coded id: 4
//    slot: 6 path='runtime/internal/math'  hard-coded id: 6
//    slot: 7 path='internal/bytealg'  hard-coded id: 7
//    slot: 8 path='internal/goexperiment'
//    slot: 9 path='runtime/internal/syscall'  hard-coded id: 8
//    slot: 10 path='runtime'  hard-coded id: 9
//    fatal error: runtime.addcovmeta: fatal error
//
// For the error above, the hard-coded list is missing "internal/goos"
// and "internal/goexperiment" ; the developer in question will need
// to copy the list above into "rtPkgs" below.
//
// FIXME: the things in this list vary depending on OS/arch. In
// particular, it's a different list for WASM; need to write code
// here that does the right thing.

var rtPkgs = [...]string{
	"internal/cpu",
	"internal/goarch",
	"runtime/internal/atomic",
	"internal/goos",
	"runtime/internal/sys",
	"internal/abi",
	"runtime/internal/math",
	"internal/bytealg",
	"internal/goexperiment",
	"runtime/internal/syscall",
	"runtime",
}

const NotHardCoded = -1

// hardCodedPkgId returns the hard-coded ID for the current package, or
// -1 if we don't use a hard-coded ID. Hard-coded IDs start at -2 and
// decrease as we go down the list.
func HardCodedPkgId(pkgpath string) int {
	for k, p := range rtPkgs {
		if p == pkgpath {
			return (0 - k) - 2
		}
	}
	return NotHardCoded
}
