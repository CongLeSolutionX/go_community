// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"cmd/go/internal/cfg"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// AppendPWD returns the result of appending PWD=dir to the environment base.
//
// The resulting environment makes os.Getwd more efficient for a subprocess
// running in dir, and also improves the accuracy of paths relative to dir
// if one or more elements of dir is a symlink.
func AppendPWD(base []string, dir string) []string {
	// POSIX requires PWD to be absolute.
	// Internally we only use absolute paths, so dir should already be absolute.
	if !filepath.IsAbs(dir) {
		panic(fmt.Sprintf("AppendPWD with relative path %q", dir))
	}
	return append(base, "PWD="+dir)
}

// AppendPATH returns the result of appending PATH=$GOROOT/bin:$PATH
// (or the platform equivalent) to the environment base.
func AppendPATH(base []string) []string {
	if cfg.GOROOTbin == "" {
		return base
	}

	pathVar := "PATH"
	if runtime.GOOS == "plan9" {
		pathVar = "path"
	}

	path := os.Getenv(pathVar)
	if path == "" {
		return append(base, pathVar+"="+cfg.GOROOTbin)
	}
	return append(base, pathVar+"="+cfg.GOROOTbin+string(os.PathListSeparator)+path)
}

// Copied from ../toolchain/toolchain.go.
const gotoolchainCountEnv = "GOTOOLCHAIN_INTERNAL_SWITCH_COUNT"

// ClearSwitchEnv clears the internal environment variable counting
// toolchain switch depth. See ../toolchain/toolchain.go for details.
// All the commands that run user code should call ClearSwitchEnv
// before doing so (go run, go test, go tool).
func ClearSwitchEnv() {
	os.Unsetenv(gotoolchainCountEnv)
}
