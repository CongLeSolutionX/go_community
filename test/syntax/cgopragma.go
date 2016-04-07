// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

//go:cgo_export_dynamic              // ERROR "usage: //go:cgo_export_dynamic local \[remote\]"
//go:cgo_export_dynamic "arg"        // ERROR "usage: //go:cgo_export_dynamic local \[remote\]"
//go:cgo_export_dynamic arg "arg"    // ERROR "usage: //go:cgo_export_dynamic local \[remote\]"
//go:cgo_export_dynamic arg arg arg  // ERROR "usage: //go:cgo_export_dynamic local \[remote\]"

//go:cgo_export_static              // ERROR "usage: //go:cgo_export_static local \[remote\]"
//go:cgo_export_static "arg"        // ERROR "usage: //go:cgo_export_static local \[remote\]"
//go:cgo_export_static arg "arg"    // ERROR "usage: //go:cgo_export_static local \[remote\]"
//go:cgo_export_static arg arg arg  // ERROR "usage: //go:cgo_export_static local \[remote\]"

//go:cgo_import_dynamic                     // ERROR "usage: //go:cgo_import_dynamic local \[remote \[\042library\042\]\]"
//go:cgo_import_dynamic "arg"               // ERROR "usage: //go:cgo_import_dynamic local \[remote \[\042library\042\]\]"
//go:cgo_import_dynamic arg "arg"           // ERROR "usage: //go:cgo_import_dynamic local \[remote \[\042library\042\]\]"
//go:cgo_import_dynamic arg arg arg         // ERROR "usage: //go:cgo_import_dynamic local \[remote \[\042library\042\]\]"
//go:cgo_import_dynamic arg arg "arg" arg   // ERROR "usage: //go:cgo_import_dynamic local \[remote \[\042library\042\]\]"
//go:cgo_import_dynamic arg arg "arg""arg"  // ERROR "usage: //go:cgo_import_dynamic local \[remote \[\042library\042\]\]"

//go:cgo_import_static          // ERROR "usage: //go:cgo_import_static local"
//go:cgo_import_static "arg"    // ERROR "usage: //go:cgo_import_static local"
//go:cgo_import_static arg arg  // ERROR "usage: //go:cgo_import_static local"

//go:cgo_dynamic_linker            // ERROR "usage: //go:cgo_dynamic_linker \042path\042"
//go:cgo_dynamic_linker arg        // ERROR "usage: //go:cgo_dynamic_linker \042path\042"
//go:cgo_dynamic_linker "arg" arg  // ERROR "usage: //go:cgo_dynamic_linker \042path\042"

//go:cgo_ldflag            // ERROR "usage: //go:cgo_ldflag \042arg\042"
//go:cgo_ldflag arg        // ERROR "usage: //go:cgo_ldflag \042arg\042"
//go:cgo_ldflag "arg" arg  // ERROR "usage: //go:cgo_ldflag \042arg\042"
