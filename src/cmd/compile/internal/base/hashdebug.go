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
