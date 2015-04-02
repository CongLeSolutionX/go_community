// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
)

func main() {
	fixcgo()

	// ratec limits the number of tests running concurrently.
	// None of the tests are intensive, so don't bother
	// trying to manually adjust for slow builders.
	ratec := make(chan bool, runtime.NumCPU())
	errc := make(chan error, len(tests))

	for file, want := range tests {
		file, want := file, want
		ratec <- true
		go func() {
			errc <- test(file, want)
			<-ratec
		}()
	}

	var rc int
	for range tests {
		if err := <-errc; err != nil {
			fmt.Println(err)
			rc = 1
		}
	}
	os.Exit(rc)
}

// test builds the test in the given file.
// If want is non-empty, test also runs the test
// and checks that the output matches the regexp want.
func test(file, want string) error {
	// Build the program.
	cmd := exec.Command("go", "build", file+".go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build %s.go failed: %v\nOutput:\n%s", file, err, out)
	}
	defer os.Remove(file)

	// Only run the test if we have output to check.
	if want == "" {
		return nil
	}

	cmd = exec.Command("./" + file)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("./%s failed: %v\nOutput:\n%s", file, err, out)
	}

	// Canonicalize output.
	out = bytes.TrimRight(out, "\n")
	out = bytes.Replace(out, []byte{'\n'}, []byte{' '}, -1)

	// Check the result.
	match, err := regexp.Match(want, out)
	if err != nil {
		return fmt.Errorf("failed to parse regexp %q: %v", want, err)
	}
	if !match {
		return fmt.Errorf("%s.go:\n%q\ndoes not match %s", file, out, want)
	}

	return nil
}

var tests = map[string]string{
	// defer_panic_recover
	"defer":  `^0 3210 2$`,
	"defer2": `^Calling g. Printing in g 0 Printing in g 1 Printing in g 2 Printing in g 3 Panicking! Defer in g 3 Defer in g 2 Defer in g 1 Defer in g 0 Recovered in f 4 Returned normally from f.$`,

	// effective_go
	"eff_bytesize": `^1.00YB 9.09TB$`,
	"eff_qr":       "",
	"eff_sequence": `^\[-1 2 6 16 44\]$`,
	"eff_unused2":  "",

	// error_handling
	"error":  "",
	"error2": "",
	"error3": "",
	"error4": "",

	// law_of_reflection
	"interface":  "",
	"interface2": `^type: float64$`,

	// c_go_cgo
	"cgo1": "",
	"cgo2": "",
	"cgo3": "",
	"cgo4": "",

	// timeout
	"timeout1": "",
	"timeout2": "",

	// gobs
	"gobs1": "",
	"gobs2": "",

	// json
	"json1": `^$`,
	"json2": `the reciprocal of i is`,
	"json3": `Age is int 6`,
	"json4": `^$`,
	"json5": "",

	// image_package
	"image_package1": `^X is 2 Y is 1$`,
	"image_package2": `^3 4 false$`,
	"image_package3": `^3 4 true$`,
	"image_package4": `^image.Point{X:2, Y:1}$`,
	"image_package5": `^{255 0 0 255}$`,
	"image_package6": `^8 4 true$`,

	// other
	"go1":    `^Christmas is a holiday: true Sleeping for 0.123s.*go1.go already exists$`,
	"slices": "",
}

func fixcgo() {
	if os.Getenv("CGO_ENABLED") != "1" {
		delete(tests, "cgo1")
		delete(tests, "cgo2")
		delete(tests, "cgo3")
		delete(tests, "cgo4")
		return
	}

	switch runtime.GOOS {
	case "freebsd":
		// cgo1 and cgo2 don't run on freebsd, srandom has a different signature
		delete(tests, "cgo1")
		delete(tests, "cgo2")
	case "netbsd":
		// cgo1 and cgo2 don't run on netbsd, srandom has a different signature
		delete(tests, "cgo1")
		delete(tests, "cgo2")
		// cgo3 and cgo4 don't run on netbsd, since cgo cannot handle stdout correctly
		delete(tests, "cgo3")
		delete(tests, "cgo4")
	case "openbsd":
		// # cgo3 and cgo4 don't run on openbsd, since cgo cannot handle stdout correctly
		delete(tests, "cgo3")
		delete(tests, "cgo4")
	}
}
