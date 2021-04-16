// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objabi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"internal/buildcfg"
)

const (
	ElfRelocOffset   = 256
	MachoRelocOffset = 2048 // reserve enough space for ELF relocations
)

// HeaderString returns the toolchain configuration string written in
// Go object headers. This string ensures we don't attempt to import
// or link object files that are incompatible with each other. This
// string always starts with "go object ".
func HeaderString() string {
	return fmt.Sprintf("go object %s %s %s X:%s\n", buildcfg.GOOS, buildcfg.GOARCH, buildcfg.Version, strings.Join(buildcfg.EnabledExperiments(), ","))
}

func init() {
	// Diagnose invalid configuration. Delayed from internal/buildcfg,
	// which can be linked into programs using go/build.
	if buildcfg.Error != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(os.Args[0]), buildcfg.Error)
		os.Exit(2)
	}
}
