// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"internal/coverage/slicewriter"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime/coverage"
)

// Good cases to handle:
// - emit to directory
// - emit to writer, then dump to named file
//
// Bad cases to handle:
// - emit to non-existent dir
// - emit to non-writable dir
// - emit to nil writer
// - emit to writer that fails somewhere along the line

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var testpointflag = flag.String("tp", "", "Testpoint to run")
var outdirflag = flag.String("o", "", "Output dir into which to emit")

func emitToWriter() {
	var slwm slicewriter.WriteSeeker
	if err := coverage.CoverageMetaDataEmitToWriter(&slwm); err != nil {
		log.Fatalf("error: CoverageMetaDataEmitToWriter returns %v", err)
	}
	mf := filepath.Join(*outdirflag, "covmeta.0abcdef")
	if err := ioutil.WriteFile(mf, slwm.Payload(), 0666); err != nil {
		log.Fatalf("error: writing %s: %v", mf, err)
	}
	var slwc slicewriter.WriteSeeker
	if err := coverage.CoverageCounterDataEmitToWriter(&slwc); err != nil {
		log.Fatalf("error: CoverageCounterDataEmitToWriter returns %v", err)
	}
	cf := filepath.Join(*outdirflag, "covcounters.0abcdef.99.77")
	if err := ioutil.WriteFile(cf, slwc.Payload(), 0666); err != nil {
		log.Fatalf("error: writing %s: %v", cf, err)
	}
}

func emitToDir() {
	if err := coverage.CoverageMetaDataEmitToDir(*outdirflag); err != nil {
		log.Fatalf("error: CoverageMetaDataEmitToDir returns %v", err)
	}
	if err := coverage.CoverageCounterDataEmitToDir(*outdirflag); err != nil {
		log.Fatalf("error: CoverageCounterDataEmitToDir returns %v", err)
	}
}

func final() int {
	println("I run last.")
	return 43
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("harness: ")
	flag.Parse()
	if *testpointflag == "" {
		log.Fatalf("error: no testpoint (use -tp flag)")
	}
	if *outdirflag == "" {
		log.Fatalf("error: no output dir specified (use -o flag)")
	}
	switch *testpointflag {
	case "emitToDir":
		emitToDir()
	case "emitToWriter":
		emitToWriter()
	default:
		log.Fatalf("error: unknown testpoint %q", *testpointflag)
	}
	final()
}
