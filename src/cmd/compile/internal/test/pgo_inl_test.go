// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"bufio"
	"fmt"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

// TestPGOIntendedInlining tests that specific functions are inlined.
func TestPGOIntendedInlining(t *testing.T) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	// Make a temporary directory to work in.
	tmpdir, err := ioutil.TempDir("", "TestCode")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	want := map[string][]string{
		"cmd/compile/internal/test/testdata/pgo/inline": {
			"(*BS).NS",
		},
	}

	// The functions which are not expected to be inlined are as follows.
	wantNot := map[string][]string{
		"cmd/compile/internal/test/testdata/pgo/inline": {
			// The calling edge main->A is hot and the cost of A is large than
			// inlineHotCalleeMaxBudget.
			"A",
			// The calling edge BenchmarkA" -> benchmarkB is cold
			// and the cost of A is large than inlineMaxBudget.
			"benchmarkB",
		},
	}

	must := map[string]bool{
		"(*BS).NS": true,
	}

	notInlinedReason := make(map[string]string)
	pkgs := make([]string, 0, len(want))
	for pname, fnames := range want {
		pkgs = append(pkgs, pname)
		for _, fname := range fnames {
			fullName := pname + "." + fname
			if _, ok := notInlinedReason[fullName]; ok {
				t.Errorf("duplicate func: %s", fullName)
			}
			notInlinedReason[fullName] = "unknown reason"
		}
	}

	// If the compiler emit "cannot inline for function A", the entry A
	// in expectedNotInlinedList will be removed.
	expectedNotInlinedList := make(map[string]struct{})
	for pname, fnames := range wantNot {
		for _, fname := range fnames {
			fullName := pname + "." + fname
			expectedNotInlinedList[fullName] = struct{}{}
		}
	}

	// go test -bench=. -cpuprofile testdata/pgo/inline/inline_hot.pprof cmd/compile/internal/test/testdata/pgo/inline
	curdir, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting wd: %v", err)
	}
	gcflag := fmt.Sprintf("-gcflags=-m -m -pgoprofile %s/testdata/pgo/inline/inline_hot.pprof", curdir)
	args := append([]string{"test", "-run=nope", gcflag}, pkgs...)
	cmd := testenv.CleanCmdEnv(exec.Command(testenv.GoToolPath(t), args...))

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("error creating pipe: %v", err)
	}
	defer pr.Close()
	cmd.Stdout = pw
	cmd.Stderr = pw

	err = cmd.Start()
	pw.Close()
	if err != nil {
		t.Fatalf("error starting go test: %v", err)
	}

	scanner := bufio.NewScanner(pr)
	curPkg := ""
	canInline := regexp.MustCompile(`: can inline ([^ ]*)`)
	haveInlined := regexp.MustCompile(`: inlining call to ([^ ]*)`)
	cannotInline := regexp.MustCompile(`: cannot inline ([^ ]*): (.*)`)
	for scanner.Scan() {
		line := scanner.Text()
		t.Logf("child: %s", line)
		if strings.HasPrefix(line, "# ") {
			curPkg = line[2:]
			splits := strings.Split(curPkg, " ")
			curPkg = splits[0]
			continue
		}
		if m := haveInlined.FindStringSubmatch(line); m != nil {
			fname := m[1]
			delete(notInlinedReason, curPkg+"."+fname)
			continue
		}
		if m := canInline.FindStringSubmatch(line); m != nil {
			fname := m[1]
			fullname := curPkg + "." + fname
			// If function must be inlined somewhere, being inlinable is not enough
			if _, ok := must[fullname]; !ok {
				delete(notInlinedReason, fullname)
				continue
			}
		}
		if m := cannotInline.FindStringSubmatch(line); m != nil {
			fname, reason := m[1], m[2]
			fullName := curPkg + "." + fname
			if _, ok := notInlinedReason[fullName]; ok {
				// cmd/compile gave us a reason why
				notInlinedReason[fullName] = reason
			}
			delete(expectedNotInlinedList, fullName)
			continue
		}
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("error running go test: %v", err)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading go test output: %v", err)
	}
	for fullName, reason := range notInlinedReason {
		t.Errorf("%s was not inlined: %s", fullName, reason)
	}

	// If the list expectedNotInlinedList is not empty, it indicates
	// the functions in the expectedNotInlinedList are marked with caninline.
	for fullName, _ := range expectedNotInlinedList {
		t.Errorf("%s was expected not inlined", fullName)
	}
}
