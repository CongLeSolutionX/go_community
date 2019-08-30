// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"lockcheck/lockgraph"
	"lockcheck/server"
)

// TestRuntime runs the runtime and runtime/debug tests and reports if
// there are lock cycles.
func TestRuntime(t *testing.T) {
	// Build the tests with lock logging enabled and morestack
	// reporting.
	env := append(os.Environ(), "GOFLAGS=-tags=locklog -gcflags=all=-d=maymorestack=runtime.lockLogMoreStack")

	tmpDir, err := ioutil.TempDir("", "lockcheck")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	builder := server.NewGraphBuilder()

	for _, test := range []string{"runtime", "runtime/debug"} {
		// Build test.
		t.Logf("building %s test", test)
		tmpBin := filepath.Join(tmpDir, filepath.Base(test)+".test.exe")
		build := exec.Command("go", "test", "-c", "-o", tmpBin, test)
		build.Env = env
		// Run in GOROOT/src to avoid go.mod errors.
		build.Dir = filepath.Join(runtime.GOROOT(), "src")
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		if err := build.Run(); err != nil {
			t.Fatalf("building %s test: %v", test, err)
		}

		// Run the test and collect the lock graph.
		t.Logf("running %s test", test)
		var out bytes.Buffer
		cmd := exec.Command(tmpBin, "-test.short")
		cmd.Env = env
		cmd.Dir = "../../src/" + test
		cmd.Stdout = &out
		cmd.Stderr = &out
		any, err := server.Run(cmd, builder)
		if err != nil {
			t.Error(out.String())
			t.Error("lock server failed:", err)
			if !any {
				return
			}
			// Otherwise we can analyze what we have.
		}
	}

	// Check the graph for cycles.
	lockGraph := builder.Finish()
	if lockGraph.NumNodes() == 0 {
		// Sanity check.
		t.Error("lock graph contains no locks")
	}
	cNodes, _ := lockgraph.Cycles(lockGraph)
	if len(cNodes) == 0 {
		return
	}

	// There are cycles. Report the lock graph so it can be
	// analyzed later.
	if os.Getenv("GO_BUILDER_NAME") == "" {
		f, err := ioutil.TempFile("", "lockcheck")
		if err != nil {
			t.Fatal(err)
		}
		if err := lockgraph.Dump(f, lockGraph); err != nil {
			t.Fatal("dumping graph failed:", err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		t.Errorf("To analyze lock graph, run:\nlockcheck -load %s -http :8080", f.Name())
	} else {
		// Report graph to stdout so it's easy to get from a
		// builder log.
		var dump bytes.Buffer
		if err := lockgraph.Dump(&dump, lockGraph); err != nil {
			t.Fatal("dumping graph failed:", err)
		}
		t.Errorf("To analyze lock graph, run:\necho \"%s\" |\nlockcheck -load - -http :8080", dump.String())
	}
	t.Error("runtime lock graph contains cycles")
}
