// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build gc

//go:generate go run mkstdpkgs.go

package goroot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var stdMap = func() map[string]bool {
	m := map[string]bool{}
	for _, k := range std {
		m[k] = true
	}
	return m
}()

// find the goroot for this go binary. stdMap is only
// valid for that goroot. Don't use runtime.GOROOT() because
// respects the GOROOT environment variable.
var binarygoroot = func() string {
	e, err := os.Executable()
	if err != nil {
		return ""
	}
	goroot := strings.TrimSuffix(e, filepath.FromSlash("/bin/go"))
	if goroot == e {
		return ""
	}
	return goroot
}()

// IsStandardPackage reports whether path is a standard package,
// given goroot and compiler.
func IsStandardPackage(goroot, compiler, path string) bool {
	switch compiler {
	case "gc":
		// stdMap is only valid for true stdlib packages (that is, not packages
		// contained in testdata directories like runtime/testdata/testprog)
		// that belong to the same GOROOT as this binary.
		if goroot == binarygoroot && !strings.Contains(path, "testdata") {
			return stdMap[path]
		}
		dir := filepath.Join(goroot, "src", path)
		_, err := os.Stat(dir)
		return err == nil
	case "gccgo":
		return gccgoSearch.isStandard(path)
	default:
		panic("unknown compiler " + compiler)
	}
}

// gccgoSearch holds the gccgo search directories.
type gccgoDirs struct {
	once sync.Once
	dirs []string
}

// gccgoSearch is used to check whether a gccgo package exists in the
// standard library.
var gccgoSearch gccgoDirs

// init finds the gccgo search directories. If this fails it leaves dirs == nil.
func (gd *gccgoDirs) init() {
	gccgo := os.Getenv("GCCGO")
	if gccgo == "" {
		gccgo = "gccgo"
	}
	bin, err := exec.LookPath(gccgo)
	if err != nil {
		return
	}

	allDirs, err := exec.Command(bin, "-print-search-dirs").Output()
	if err != nil {
		return
	}
	versionB, err := exec.Command(bin, "-dumpversion").Output()
	if err != nil {
		return
	}
	version := strings.TrimSpace(string(versionB))
	machineB, err := exec.Command(bin, "-dumpmachine").Output()
	if err != nil {
		return
	}
	machine := strings.TrimSpace(string(machineB))

	dirsEntries := strings.Split(string(allDirs), "\n")
	const prefix = "libraries: ="
	var dirs []string
	for _, dirEntry := range dirsEntries {
		if strings.HasPrefix(dirEntry, prefix) {
			dirs = filepath.SplitList(strings.TrimPrefix(dirEntry, prefix))
			break
		}
	}
	if len(dirs) == 0 {
		return
	}

	var lastDirs []string
	for _, dir := range dirs {
		goDir := filepath.Join(dir, "go", version)
		if fi, err := os.Stat(goDir); err == nil && fi.IsDir() {
			gd.dirs = append(gd.dirs, goDir)
			goDir = filepath.Join(goDir, machine)
			if fi, err = os.Stat(goDir); err == nil && fi.IsDir() {
				gd.dirs = append(gd.dirs, goDir)
			}
		}
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			lastDirs = append(lastDirs, dir)
		}
	}
	gd.dirs = append(gd.dirs, lastDirs...)
}

// isStandard reports whether path is a standard library for gccgo.
func (gd *gccgoDirs) isStandard(path string) bool {
	// Quick check: if the first path component has a '.', it's not
	// in the standard library. This skips most GOPATH directories.
	i := strings.Index(path, "/")
	if i < 0 {
		i = len(path)
	}
	if strings.Contains(path[:i], ".") {
		return false
	}

	if path == "unsafe" {
		// Special case.
		return true
	}

	gd.once.Do(gd.init)
	if gd.dirs == nil {
		// We couldn't find the gccgo search directories.
		// Best guess, since the first component did not contain
		// '.', is that this is a standard library package.
		return true
	}

	for _, dir := range gd.dirs {
		full := filepath.Join(dir, path) + ".gox"
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			return true
		}
	}

	return false
}
