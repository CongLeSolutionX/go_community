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

// testPGODevirtualize tests that specific PGO devirtualize rewrites are performed.
func testPGODevirtualize(t *testing.T, dir string) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	const pkg = "example.com/pgo/devirtualize"

	// Add a go.mod so we have a consistent symbol names in this temp dir.
	goMod := fmt.Sprintf(`module %s
go 1.19
`, pkg)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("error writing go.mod: %v", err)
	}

	// Build the test with the profile.
	pprof := filepath.Join(dir, "shape.pprof")
	gcflag := fmt.Sprintf("-gcflags=-m=2 -pgoprofile=%s -d=pgodebug=2", pprof)
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

	type devirtualization struct {
		pos    string
		callee string
	}

	want := devirtualization{
		pos:    "./shape.go:77:19",
		callee: "Circle.Area",
	}

	devirtualizedLine := regexp.MustCompile(`(.*): PGO devirtualizing call to (.*)`)

	found := false
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		line := scanner.Text()
		t.Logf("child: %s", line)
		if found {
			// If done, keep looping just to log all output.
			continue
		}

		m := devirtualizedLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		sp := devirtualization{
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
		t.Errorf("%v was not devirtualized", want)
	}
}

// TestPGODevirtualize tests that specific functions are devirtualized when PGO
// is applied to the exact source that was profiled.
func TestPGODevirtualize(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting wd: %v", err)
	}
	srcDir := filepath.Join(wd, "testdata/pgo/devirtualize")

	// Copy the module to a scratch location so we can add a go.mod.
	dir := t.TempDir()

	for _, file := range []string{"shape.go", "shape_test.go", "shape.pprof"} {
		if err := copyFile(filepath.Join(dir, file), filepath.Join(srcDir, file)); err != nil {
			t.Fatalf("error copying %s: %v", file, err)
		}
	}

	testPGODevirtualize(t, dir)
}
