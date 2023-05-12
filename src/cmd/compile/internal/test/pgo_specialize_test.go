// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"bufio"
	"fmt"
	"internal/testenv"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// testPGOSpecialize tests that specific specializations are performed.
func testPGOSpecialize(t *testing.T, dir string) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	const pkg = "example.com/pgo/specialize"

	// Add a go.mod so we have a consistent symbol names in this temp dir.
	goMod := fmt.Sprintf(`module %s
go 1.19
`, pkg)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("error writing go.mod: %v", err)
	}

	// Build the test with the profile.
	pprof := filepath.Join(dir, "shape.pprof")
	gcflag := fmt.Sprintf("-gcflags=-m=2 -pgoprofile=%s", pprof)
	out := filepath.Join(dir, "test.exe")
	cmd := testenv.CleanCmdEnv(testenv.Command(t, testenv.GoToolPath(t), "test", "-c", "-o", out, gcflag, "."))
	cmd.Dir = dir

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

	type specialization struct {
		pos    string
		callee string
	}

	want := specialization{
		pos:    "./shape.go:76:19",
		callee: "Circle.Area",
	}

	specializedLine := regexp.MustCompile(`(.*): specializing call to (.*)`)

	found := false
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		line := scanner.Text()
		t.Logf("child: %s", line)
		if found {
			// If done, keep looping just to log all output.
			continue
		}

		m := specializedLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		sp := specialization{
			pos:    m[1],
			callee: m[2],
		}
		if sp == want {
			found = true
		}
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("error running go test: %v", err)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading go test output: %v", err)
	}
	if !found {
		t.Errorf("%v was not specialized", want)
	}
}

// TestPGOSpecialize tests that specific functions are specialized when PGO is
// applied to the exact source that was profiled.
func TestPGOSpecialize(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting wd: %v", err)
	}
	srcDir := filepath.Join(wd, "testdata/pgo/specialize")

	// Copy the module to a scratch location so we can add a go.mod.
	dir := t.TempDir()

	for _, file := range []string{"shape.go", "shape_test.go", "shape.pprof"} {
		if err := copyFile(filepath.Join(dir, file), filepath.Join(srcDir, file)); err != nil {
			t.Fatalf("error copying %s: %v", file, err)
		}
	}

	testPGOSpecialize(t, dir)
}
