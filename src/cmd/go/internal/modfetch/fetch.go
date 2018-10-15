// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modfetch

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/dirhash"
	"cmd/go/internal/lockedfile"
	"cmd/go/internal/module"
	"cmd/go/internal/par"
	"cmd/go/internal/renameio"
)

// withSentinel runs the function f guarded by a sentinel file with the given suffix.
//
// If the sentinel file already exists and is non-empty, withSentinel imports the
// hash of mod and returns nil.
//
// If f returns nil, withSentinel makes the sentinel file non-empty.
func withSentinel(mod module.Version, suffix string, f func() error) error {
	sentinelPath, err := CachePath(mod, suffix)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(sentinelPath), 0777); err != nil {
		return err
	}
	sentinel, err := lockedfile.NewSentinel(sentinelPath)
	if os.IsExist(err) {
		// The sentinel is already present in the cache directory.
		// Verify that the the zipfile's hash matches the main module's go.sum file,
		// adding it if needed.
		return checkSum(mod)
	} else if err != nil {
		return err
	}
	defer sentinel.Close()

	if err := f(); err != nil {
		return err
	}

	_, err = sentinel.WriteString(time.Now().String())
	return err
}

var downloadCache par.Cache

// Download downloads the specific module version to the
// local download cache and returns the name of the directory
// corresponding to the root of the module's file tree.
func Download(mod module.Version) (dir string, err error) {
	if PkgMod == "" {
		// Do not download to current directory.
		return "", fmt.Errorf("missing modfetch.PkgMod")
	}

	// The par.Cache here avoids duplicate work.
	type cached struct {
		dir string
		err error
	}
	c := downloadCache.Do(mod, func() interface{} {
		dir, err := DownloadDir(mod)
		if err != nil {
			return cached{err: err}
		}

		// In theory we could avoid the need for a sentinel file by write-locking
		// the zipfile instead, but that assumes that the zipfile remains writable.
		// If it is only readable — all we really require to unzip it — then we can
		// only obtain a read-lock, which would not exclude other processes from
		// extracting it simultaneously.
		err = withSentinel(mod, "unzipped", func() error { return download(mod, dir) })
		if err != nil {
			return cached{err: err}
		}
		return cached{dir: dir}
	}).(cached)
	return c.dir, c.err
}

func download(mod module.Version, dir string) error {
	zipfile, err := DownloadZip(mod)
	if err != nil {
		return err
	}
	if cfg.CmdName != "mod download" {
		fmt.Fprintf(os.Stderr, "go: extracting %s %s\n", mod.Path, mod.Version)
	}

	modpath := mod.Path + "@" + mod.Version
	if err := Unzip(dir, zipfile, modpath); err != nil {
		fmt.Fprintf(os.Stderr, "-> %s\n", err)
		return err
	}
	return nil
}

var downloadZipCache par.Cache

// DownloadZip downloads the specific module version to the
// local zip cache and returns the name of the zip file.
func DownloadZip(mod module.Version) (zipfile string, err error) {
	// The par.Cache here avoids duplicate work.
	type cached struct {
		zipfile string
		err     error
	}
	c := downloadZipCache.Do(mod, func() interface{} {
		zipfile, err := CachePath(mod, "zip")
		if err != nil {
			return cached{err: err}
		}

		// The zipfile may be read by a GOPROXY HTTP server, so we must write use
		// Rename to write it atomically. Therefore, we must use a separate sentinel
		// file to serialize writes to it.
		err = withSentinel(mod, "downloaded", func() error { return downloadZip(mod, zipfile) })
		if err != nil {
			return cached{err: err}
		}
		return cached{zipfile: zipfile}
	}).(cached)
	return c.zipfile, c.err
}

func downloadZip(mod module.Version, zipfile string) (err error) {
	if cfg.CmdName != "mod download" {
		fmt.Fprintf(os.Stderr, "go: downloading %s %s\n", mod.Path, mod.Version)
	}

	if err := os.MkdirAll(filepath.Dir(zipfile), 0777); err != nil {
		return err
	}

	// Clean up any remaining tempfiles from previous runs.
	// This is only safe to do because the sentinel file ensures that their
	// writers are no longer active.
	for _, base := range []string{zipfile, zipfile + "hash"} {
		if old, err := filepath.Glob(base + ".tmp-*"); err == nil {
			for _, path := range old {
				os.Remove(path) // best effort
			}
		}
	}

	f, err := ioutil.TempFile(filepath.Dir(zipfile), filepath.Base(zipfile)+".tmp-")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	repo, err := Lookup(mod.Path)
	if err != nil {
		return err
	}
	if err := repo.Zip(f, mod.Version); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Hash the zip file and check the sum before returning to write the sentinel.
	// As a side effect, this also ensures that the zipfile is structurally-valid.
	// (We'll check the contents more stringently in Unzip.)
	hash, err := dirhash.HashZip(f.Name(), dirhash.DefaultHash)
	if err != nil {
		return err
	}
	if err := checkOneSum(mod, hash); err != nil {
		return err
	}

	if err := renameio.WriteFile(zipfile+"hash", []byte(hash)); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), zipfile); err != nil {
		return err
	}
	return makeReadonly(zipfile, zipfile+"hash")
}

var GoSumFile string // path to go.sum; set by package modload

type modSum struct {
	mod module.Version
	sum string
}

var goSum struct {
	mu        sync.Mutex
	m         map[module.Version][]string // content of go.sum file (+ go.modverify if present)
	checked   map[modSum]bool             // sums actually checked during execution
	dirty     bool                        // whether we added any new sums to m
	overwrite bool                        // if true, overwrite go.sum without incorporating its contents
	enabled   bool                        // whether to use go.sum at all
	modverify string                      // path to go.modverify, to be deleted
}

// initGoSum initializes the go.sum data.
// It reports whether use of go.sum is now enabled.
// The goSum lock must be held.
func initGoSum() bool {
	if GoSumFile == "" {
		return false
	}
	if goSum.m != nil {
		return true
	}

	goSum.m = make(map[module.Version][]string)
	goSum.checked = make(map[modSum]bool)
	data, err := ioutil.ReadFile(GoSumFile)
	if err != nil && !os.IsNotExist(err) {
		base.Fatalf("go: %v", err)
	}
	goSum.enabled = true
	readGoSum(goSum.m, GoSumFile, data)

	// Add old go.modverify file.
	// We'll delete go.modverify in WriteGoSum.
	alt := strings.TrimSuffix(GoSumFile, ".sum") + ".modverify"
	if data, err := ioutil.ReadFile(alt); err == nil {
		migrate := make(map[module.Version][]string)
		readGoSum(migrate, alt, data)
		for mod, sums := range migrate {
			for _, sum := range sums {
				if err := checkOneSumLocked(mod, sum); err != nil {
					base.Fatalf("go: checksum in go.modverify conflicts with go.sum")
				}
			}
		}
		goSum.modverify = alt
	}
	return true
}

// emptyGoModHash is the hash of a 1-file tree containing a 0-length go.mod.
// A bug caused us to write these into go.sum files for non-modules.
// We detect and remove them.
const emptyGoModHash = "h1:G7mAYYxgmS0lVkHyy2hEOLQCFB0DlQFTMLWggykrydY="

// readGoSum parses data, which is the content of file,
// and adds it to goSum.m. The goSum lock must be held.
func readGoSum(dst map[module.Version][]string, file string, data []byte) {
	lineno := 0
	for len(data) > 0 {
		var line []byte
		lineno++
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			line, data = data, nil
		} else {
			line, data = data[:i], data[i+1:]
		}
		f := strings.Fields(string(line))
		if len(f) == 0 {
			// blank line; skip it
			continue
		}
		if len(f) != 3 {
			base.Fatalf("go: malformed go.sum:\n%s:%d: wrong number of fields %v", file, lineno, len(f))
		}
		if f[2] == emptyGoModHash {
			// Old bug; drop it.
			continue
		}
		mod := module.Version{Path: f[0], Version: f[1]}
		dst[mod] = append(dst[mod], f[2])
	}
}

// checkSum checks the given module's checksum.
func checkSum(mod module.Version) error {
	if PkgMod == "" {
		// Do not use current directory.
		return nil
	}

	// Do the file I/O before acquiring the go.sum lock.
	ziphash, err := CachePath(mod, "ziphash")
	if err != nil {
		return fmt.Errorf("verifying %s@%s: %v", mod.Path, mod.Version, err)
	}
	data, err := ioutil.ReadFile(ziphash)
	if err != nil {
		if os.IsNotExist(err) {
			// This can happen if someone does rm -rf GOPATH/src/cache/download. So it goes.
			return nil
		}
		return fmt.Errorf("verifying %s@%s: %v", mod.Path, mod.Version, err)
	}
	h := strings.TrimSpace(string(data))
	if !strings.HasPrefix(h, "h1:") {
		return fmt.Errorf("verifying %s@%s: unexpected ziphash: %q", mod.Path, mod.Version, h)
	}

	return checkOneSum(mod, h)
}

// goModSum returns the checksum for the go.mod contents.
func goModSum(data []byte) (string, error) {
	return dirhash.Hash1([]string{"go.mod"}, func(string) (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewReader(data)), nil
	})
}

// checkGoMod checks the given module's go.mod checksum;
// data is the go.mod content.
func checkGoMod(path, version string, data []byte) error {
	h, err := goModSum(data)
	if err != nil {
		return fmt.Errorf("verifying %s %s go.mod: %v", path, version, err)
	}

	return checkOneSum(module.Version{Path: path, Version: version + "/go.mod"}, h)
}

// checkOneSum checks that the recorded hash for mod is h.
func checkOneSum(mod module.Version, h string) error {
	goSum.mu.Lock()
	defer goSum.mu.Unlock()
	if !initGoSum() {
		return nil
	}
	return checkOneSumLocked(mod, h)
}

func checkOneSumLocked(mod module.Version, h string) (err error) {
	defer func() {
		if err == nil {
			goSum.checked[modSum{mod, h}] = true
		} else {
			// All checksum errors should cause the go command to exit with a
			// non-success code, but we want to return to the caller rather than exiting
			// right away so that they can clean up any temporary or half-written files.
			base.Errorf("%v", err)
		}
	}()

	for _, vh := range goSum.m[mod] {
		if h == vh {
			return nil
		}
		if strings.HasPrefix(vh, "h1:") {
			return fmt.Errorf("verifying %s@%s: checksum mismatch\n\tdownloaded: %v\n\tgo.sum:     %v", mod.Path, mod.Version, h, vh)
		}
	}
	if len(goSum.m[mod]) > 0 {
		fmt.Fprintf(os.Stderr, "warning: verifying %s@%s: unknown hashes in go.sum: %v; adding %v", mod.Path, mod.Version, strings.Join(goSum.m[mod], ", "), h)
	}
	goSum.m[mod] = append(goSum.m[mod], h)
	goSum.dirty = true
	return nil
}

// Sum returns the checksum for the downloaded copy of the given module,
// if present in the download cache.
func Sum(mod module.Version) string {
	if PkgMod == "" {
		// Do not use current directory.
		return ""
	}

	ziphash, err := CachePath(mod, "ziphash")
	if err != nil {
		return ""
	}
	data, err := ioutil.ReadFile(ziphash)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func serializeGoSum(dst io.Writer) error {
	var mods []module.Version
	for m := range goSum.m {
		mods = append(mods, m)
	}
	module.Sort(mods)
	for _, m := range mods {
		list := goSum.m[m]
		sort.Strings(list)
		for _, h := range list {
			if _, err := fmt.Fprintf(dst, "%s %s %s\n", m.Path, m.Version, h); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteGoSum writes the go.sum file if it needs to be updated.
func WriteGoSum() (err error) {
	goSum.mu.Lock()
	defer func() {
		goSum.mu.Unlock()
		base.ExitIfErrors()
	}()

	if !goSum.enabled {
		// If we haven't read the go.sum file yet, don't bother writing it: at best,
		// we could rename the go.modverify file if it isn't empty, but we haven't
		// needed to touch it so far — how important could it be?
		return nil
	}
	if !goSum.dirty {
		// Don't bother opening the go.sum file if we don't have anything to add.
		return nil
	}

	// We want to avoid races between creating the lockfile and deleting it, but
	// we also don't want to leave a permanent lockfile in the user's repository.
	//
	// On top of that, if we crash while writing go.sum, we don't want to lose the
	// sums that were already present in the file, so it's important that we write
	// the file by renaming rather than truncating — which means that we can't
	// lock the go.sum file itself.
	//
	// Instead, we'll lock a distinguished file in the cache directory: that will
	// only race if the user runs `go clean -modcache` concurrently with a command
	// that updates go.sum, and that's already racy to begin with.
	//
	// We'll end up slightly over-synchronizing go.sum writes if the user runs a
	// bunch of go commands that update sums in separate modules simultaneously,
	// but that's unlikely to matter in practice.

	mu, err := SideLock()
	if err != nil {
		return err
	}
	unlock, err := mu.Lock()
	if err != nil {
		return err
	}
	defer unlock()

	f, err := ioutil.TempFile(filepath.Dir(GoSumFile), filepath.Base(GoSumFile)+".tmp-")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	if !goSum.overwrite {
		// Re-read the go.sum file to incorporate any sums added by other processes
		// in the meantime.
		data, err := ioutil.ReadFile(GoSumFile)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		// Add only the sums that we actually checked: the user may have edited or
		// truncated the file to remove erroneous hashes, and we shouldn't restore
		// them without good reason.
		goSum.m = make(map[module.Version][]string, len(goSum.m))
		readGoSum(goSum.m, GoSumFile, data)
		for ms := range goSum.checked {
			if err := checkOneSumLocked(ms.mod, ms.sum); err != nil {
				return err
			}
		}
	}

	if err := serializeGoSum(f); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), GoSumFile); err != nil {
		return err
	}

	goSum.checked = make(map[modSum]bool)
	goSum.dirty = false
	goSum.overwrite = false

	if goSum.modverify != "" {
		os.Remove(goSum.modverify) // best effort
	}
	return nil
}

// TrimGoSum trims go.sum to contain only the modules for which keep[m] is true.
func TrimGoSum(keep map[module.Version]bool) {
	goSum.mu.Lock()
	defer goSum.mu.Unlock()
	if !initGoSum() {
		return
	}

	for m := range goSum.m {
		// If we're keeping x@v we also keep x@v/go.mod.
		// Map x@v/go.mod back to x@v for the keep lookup.
		noGoMod := module.Version{Path: m.Path, Version: strings.TrimSuffix(m.Version, "/go.mod")}
		if !keep[m] && !keep[noGoMod] {
			delete(goSum.m, m)
			goSum.dirty = true
			goSum.overwrite = true
		}
	}
}
