// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_ "unsafe" // for linkname.
)

// These functions are called from C code via cgo/callbacks.go.

// Panic. The argument is converted into a Go string.

// Call like this in code compiled with gcc:
//   struct { const char *p; } a;
//   a.p = /* string to pass to panic */;
//   crosscall2(_cgo_panic, &a, sizeof a);
//   /* The function call will not return.  */

// TODO: We should export a regular C function to panic, change SWIG
// to use that instead of the above pattern, and then we can drop:
// * backwards-compatibility from crosscall2 and stop exporting it.
// * The _cgo_panic check from cgocallbackg1.
// * The callPanicSWIG test.

//go:linkname _cgo_panic _cgo_panic
func _cgo_panic(a *struct{ cstr *byte }) {
	panic(gostringnocopy(a.cstr))
}
