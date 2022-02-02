// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// go work use

package workcmd

import (
	"cmd/go/internal/base"
	"cmd/go/internal/fsys"
	"cmd/go/internal/modload"
	"cmd/go/internal/str"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var cmdUse = &base.Command{
	UsageLine: "go work use [-r] [moddirs]",
	Short:     "add modules to workspace file",
	Long: `Use provides a command-line interface for adding
directories, optionally recursively, to a go.work file.

A use directive will be added to the go.work file for each argument
directory listed on the command line go.work file, if it exists on disk,
or removed from the go.work file if it does not exist on disk.

The -r flag searches recursively for modules in the argument
directories, and the use command operates as if each of the directories
were specified as arguments: namely, use directives will be added for
directories that exist, and removed for directories that do not exist.
`,
}

var useR = cmdUse.Flag.Bool("r", false, "")

func init() {
	cmdUse.Run = runUse // break init cycle

	base.AddModCommonFlags(&cmdUse.Flag)
	base.AddWorkfileFlag(&cmdUse.Flag)
}

func runUse(ctx context.Context, cmd *base.Command, args []string) {
	modload.ForceUseModules = true

	var gowork string
	modload.InitWorkfile()
	gowork = modload.WorkFilePath()

	if gowork == "" {
		base.Fatalf("go: no go.work file found\n\t(run 'go work init' first or specify path using -workfile flag)")
	}
	workFile, err := modload.ReadWorkFile(gowork)
	if err != nil {
		base.Fatalf("go: %v", err)
	}

	absWork, err := filepath.Abs(gowork)
	if err != nil {
		base.Fatalf("go: %v", err)
	}
	workDir := filepath.Dir(absWork)

	haveDirs := make(map[string][]string) // absolute path → original(s)
	for _, use := range workFile.Use {
		var abs string
		if filepath.IsAbs(use.Path) {
			abs = filepath.Clean(use.Path)
		} else {
			abs = filepath.Join(workDir, use.Path)
		}
		haveDirs[abs] = append(haveDirs[abs], use.Path)
	}

	staleDirs := make(map[string]bool)
	keepDirs := make(map[string]bool)
	lookDir := func(dir string) {
		absDir, dir := pathRel(workDir, dir)

		fi, err := os.Stat(filepath.Join(absDir, "go.mod"))
		if err != nil {
			if os.IsNotExist(err) {
				for _, origDir := range haveDirs[absDir] {
					staleDirs[origDir] = true
				}
				return
			}
			base.Errorf("go: %v", err)
		}

		if !fi.Mode().IsRegular() {
			base.Errorf("go: %v is not regular", filepath.Join(dir, "go.mod"))
		}

		for _, haveDir := range haveDirs[absDir] {
			if haveDir != dir {
				staleDirs[haveDir] = true
			}
		}
		keepDirs[dir] = true
	}

	for _, useDir := range args {
		if !*useR {
			lookDir(useDir)
			continue
		}

		// Add or remove entries for any subdirectories that still exist.
		err := fsys.Walk(useDir, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() {
				if info.Mode()&fs.ModeSymlink != 0 {
					if target, err := fsys.Stat(path); err == nil && target.IsDir() {
						fmt.Fprintf(os.Stderr, "warning: ignoring symlink %s\n", path)
					}
				}
				return nil
			}
			lookDir(path)
			return nil
		})
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			base.Errorf("go: %v", err)
		}

		absArg, _ := pathRel(workDir, useDir)
		// Remove entries for subdirectories that no longer exist.
		for absDir, dirs := range haveDirs {
			if !str.HasFilePathPrefix(absDir, absArg) {
				continue
			}
			seen := false
			allStale := true
			for _, dir := range dirs {
				// No need to stat the go.mod file for a directory we already walked.
				if _, kept := keepDirs[dir]; kept {
					seen = true
					break
				}
				if _, stale := staleDirs[dir]; !stale {
					allStale = false
				}
			}
			if !seen && !allStale {
				if _, err := os.Stat(filepath.Join(absDir, "go.mod")); errors.Is(err, os.ErrNotExist) {
					for _, dir := range dirs {
						staleDirs[dir] = true
					}
				}
			}
		}
	}

	for dir := range staleDirs {
		if keepDirs[dir] {
			// The user explicitly requested both the relative and absolute path. Keep
			// them both (and probably report an error down the road).
		} else {
			workFile.DropUse(dir)
		}
	}
	for dir := range keepDirs {
		have := false
		for _, haveDir := range haveDirs[dir] {
			if haveDir == dir {
				have = true
				break
			}
		}
		if !have {
			workFile.AddUse(dir, "")
		}
	}
	modload.UpdateWorkFile(workFile)
	modload.WriteWorkFile(gowork, workFile)
}

// goWorkRel returns the absolute and canonical forms of dir for use in a
// go.work file located in directory workDir.
//
// If dir is relative, it is intepreted relative to base.Cwd()
// and its canonical form is relative to gowork if possible.
// If dir is absolute or cannot be made relative to gowork,
// its canonical form is absolute.
//
// Canonical absolute paths are clean.
// Canonical relative paths are clean and slash-separated.
func pathRel(workDir, dir string) (abs, canonical string) {
	if filepath.IsAbs(dir) {
		abs = filepath.Clean(dir)
		return abs, abs
	}

	abs = filepath.Join(base.Cwd(), dir)
	rel, err := filepath.Rel(workDir, abs)
	if err != nil {
		// The path can't be made relative to the go.work file,
		// so it must be kept absolute instead.
		return abs, abs
	}

	// Normalize relative paths to use slashes, so that checked-in go.work
	// files with relative paths within the repo are platform-independent.
	return abs, filepath.ToSlash(rel)
}
