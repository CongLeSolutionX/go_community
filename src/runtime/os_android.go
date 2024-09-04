// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"unsafe"
)

const PROP_NAME_MAX = 32
const PROP_VALUE_MAX = 92

//go:linkname androidVersion runtime.androidVersion
func androidVersion() int {
	var name [PROP_NAME_MAX]byte
	var value [PROP_VALUE_MAX]byte
	copy(name[:], "ro.build.version.release")
	length := __system_property_get(&name[0], &value[0])
	version, _ := atoi(unsafe.String(&value[0], length))
	return version
}

// Export the main function.
//
// Used by the app package to start all-Go Android apps that are
// loaded via JNI. See golang.org/x/mobile/app.

//go:cgo_export_static main.main
//go:cgo_export_dynamic main.main
