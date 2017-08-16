// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fmtcmd implements the ``go fmt'' command.
package fmtcmd

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/load"
	"cmd/go/internal/str"
)

func init() {
	base.AddBuildFlagsNX(&CmdFmt.Flag)
}

var CmdFmt = &base.Command{
	Run:       runFmt,
	UsageLine: "fmt [-n] [-x] [packages]",
	Short:     "run gofmt on package sources",
	Long: `
Fmt runs the command 'gofmt -l -w' on the packages named
by the import paths. It prints the names of the files that are modified.

For more about gofmt, see 'go doc cmd/gofmt'.
For more about specifying packages, see 'go help packages'.

The -n flag prints commands that would be executed.
The -x flag prints commands as they are executed.

To run gofmt with specific options, run gofmt itself.

See also: go fix, go vet.
	`,
}

func runFmt(cmd *base.Command, args []string) {
	gofmt := gofmtPath()
	var allFiles []string
	for _, pkg := range load.Packages(args) {
		// Use pkg.gofiles instead of pkg.Dir so that
		// the command only applies to this package,
		// not to packages in subdirectories.
		files := base.RelPaths(pkg.InternalAllGoFiles())
		allFiles = append(allFiles, files...)
	}

	// Send files to different gofmts to overlap I/O.
	procs := 2 * runtime.GOMAXPROCS(0)
	perProc := (len(allFiles) + procs - 1) / procs
	var wg sync.WaitGroup
	for len(allFiles) > 0 {
		files := allFiles
		if len(files) > perProc {
			files = files[:perProc]
		}
		allFiles = allFiles[len(files):]
		wg.Add(1)
		go func() {
			base.Run(str.StringList(gofmt, "-l", "-w", files))
			wg.Done()
		}()
	}
	wg.Wait()
}

func gofmtPath() string {
	gofmt := "gofmt"
	if base.ToolIsWindows {
		gofmt += base.ToolWindowsExtension
	}

	gofmtPath := filepath.Join(cfg.GOBIN, gofmt)
	if _, err := os.Stat(gofmtPath); err == nil {
		return gofmtPath
	}

	gofmtPath = filepath.Join(cfg.GOROOT, "bin", gofmt)
	if _, err := os.Stat(gofmtPath); err == nil {
		return gofmtPath
	}

	// fallback to looking for gofmt in $PATH
	return "gofmt"
}
