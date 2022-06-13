// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"cmd/internal/notsha256"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

const GOSSAHASH = "GOSSAHASH"

type writeSyncer interface {
	io.Writer
	Sync() error
}

type HashDebug struct {
	// what file (if any) receives the yes/no logging?
	// default is os.Stdout
	logfile writeSyncer
}

var mu sync.Mutex
var hd HashDebug

// DebugHashMatch reports whether environment variable GOSSAHASH
//  1. is empty (this is a special more-quickly implemented case of 3)
//  2. is "y" or "Y"
//  3. is a suffix of the sha1 hash of name
//  4. is a suffix of the environment variable
//     fmt.Sprintf("%s%d", evname, n)
//     provided that all such variables are nonempty for 0 <= i <= n
//
// Otherwise it returns false.
//
// Unless GOSSAHASH is empty, when DebugHashMatch returns true the message
//
//	"%s triggered %s\n", evname, name
//
// is printed on the file named in environment variable GSHS_LOGFILE,
// or standard out if that is empty or there is an error opening the file.
//
// Typical use:
//
//  1. you make a change to the compiler, say, adding a new phase
//
//  2. it is broken in some mystifying way, for example, make.bash builds a broken
//     compiler that almost works, but crashes compiling a test in run.bash.
//
//  3. add this guard to the code, which by default leaves it broken, but
//     does not run the broken new code if GOSSAHASH is non-empty and non-matching:
//
//     if !base.DebugHashMatch(ir.PkgFuncName(fn)) {
//     return nil // early exit, do nothing
//     }
//  4. use github.com/dr2chase/gossahash to search for the error:
//     gossahash -- <the thing that fails>; e.g., gossahash -- ./all.bash
//  5. gossahash should return the single (hopefully) function whose miscompilation
//     causes the problem.
//
// Pedantic nit; actually a series of environment variables GOSSAHASH, GOSSAHASH0,
// GOSSAHASH1, etc are tried for a suffix match until GOSSAHASH_k is empty.
// This is used by automated search for failures with multiple causes.
func DebugHashMatch(pkgAndName string) bool {
	return hd.DebugHashMatch(pkgAndName)
}

func (d *HashDebug) DebugHashMatch(pkgAndName string) bool {
	evname := GOSSAHASH
	evhash := os.Getenv(evname)
	hstr := ""

	switch evhash {
	case "":
		return true // default behavior with no EV is "on"
	case "n", "N":
		return false
	}

	// Check the hash of the name against a partial input hash.
	// We use this feature to do a binary search to
	// find a function that is incorrectly compiled.
	for _, b := range notsha256.Sum256([]byte(pkgAndName)) {
		hstr += fmt.Sprintf("%08b", b)
	}

	if evhash == "y" || evhash == "Y" || strings.HasSuffix(hstr, evhash) {
		d.logDebugHashMatch(evname, pkgAndName, hstr)
		return true
	}

	// Iteratively try additional hashes to allow tests for multi-point
	// failure.
	for i := 0; true; i++ {
		ev := fmt.Sprintf("%s%d", evname, i)
		evv := os.Getenv(ev)
		if evv == "" {
			break
		}
		if strings.HasSuffix(hstr, evv) {
			d.logDebugHashMatch(ev, pkgAndName, hstr)
			return true
		}
	}
	return false
}

func (d *HashDebug) logDebugHashMatch(evname, name, hstr string) {
	mu.Lock()
	defer mu.Unlock()
	file := d.logfile
	if file == nil {
		file = os.Stdout
		if tmpfile := os.Getenv("GSHS_LOGFILE"); tmpfile != "" {
			var err error
			file, err = os.OpenFile(tmpfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				Fatalf("could not open hash-testing logfile %s", tmpfile)
				return
			}
		}
		d.logfile = file
	}
	if len(hstr) > 24 {
		hstr = hstr[len(hstr)-24:]
	}
	fmt.Fprintf(file, "%s triggered %s %s\n", evname, name, hstr)
	file.Sync()
}
