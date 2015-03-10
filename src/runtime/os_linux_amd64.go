// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// Export the runtime entry point symbol, used by external packages to start
// the Go runtime after loading a Go shared library.

//go:cgo_export_static _rt0_amd64_linux1
//go:cgo_export_dynamic _rt0_amd64_linux1
