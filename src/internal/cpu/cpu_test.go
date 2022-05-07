// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu_test

import (
	. "internal/cpu"
	"internal/godebug"
	"internal/testenv"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func MustHaveDebugOptionsSupport(t *testing.T) {
	if !DebugOptions {
		t.Skipf("skipping test: cpu feature options not supported by OS")
	}
}

func MustSupportFeatureDectection(t *testing.T) {
	// TODO: add platforms that do not have CPU feature detection support.
}

// isPassingInOutput looks at the last few lines of output from a
// completed test, checking to see if the test passed. It skips over
// code coverage metrics and blank lines in the process, returning
// true if the test passed.
func isPassingInOutput(t *testing.T, output string) bool {
	t.Helper()
	lines := strings.Split(string(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "coverage: ") || line == "" {
			continue
		}
		return line == "PASS"
	}
	return false
}

func runDebugOptionsTest(t *testing.T, test string, options string) {
	MustHaveDebugOptionsSupport(t)

	testenv.MustHaveExec(t)

	env := "GODEBUG=" + options

	cmd := exec.Command(os.Args[0], "-test.run="+test)
	cmd.Env = append(cmd.Env, env)

	output, err := cmd.CombinedOutput()
	if err != nil || !isPassingInOutput(t, string(output)) {
		t.Fatalf("%s with %s: run failed: %v output:\n%s\n",
			test, env, err, string(output))
	}
}

func TestDisableAllCapabilities(t *testing.T) {
	MustSupportFeatureDectection(t)
	runDebugOptionsTest(t, "TestAllCapabilitiesDisabled", "cpu.all=off")
}

func TestAllCapabilitiesDisabled(t *testing.T) {
	MustHaveDebugOptionsSupport(t)

	if godebug.Get("cpu.all") != "off" {
		t.Skipf("skipping test: GODEBUG=cpu.all=off not set")
	}

	for _, o := range Options {
		want := false
		if got := *o.Feature; got != want {
			t.Errorf("%v: expected %v, got %v", o.Name, want, got)
		}
	}
}
