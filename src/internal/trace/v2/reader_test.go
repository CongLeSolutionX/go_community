// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace_test

import (
	"bytes"
	"flag"
	"fmt"
	"internal/txtar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"internal/trace/v2"
	"internal/trace/v2/internal/testtrace"
	"internal/trace/v2/raw"
	"internal/trace/v2/version"
)

var (
	logEvents  = flag.Bool("log-events", false, "whether to log high-level events; significantly slows down tests")
	dumpTraces = flag.Bool("dump-traces", false, "dump traces even on success")
)

func TestReaderGolden(t *testing.T) {
	matches, err := filepath.Glob("./testdata/tests/*.test")
	if err != nil {
		t.Fatalf("failed to glob for tests: %v", err)
	}
	for _, testPath := range matches {
		testPath := testPath
		testName, err := filepath.Rel("./testdata", testPath)
		if err != nil {
			t.Fatalf("failed to relativize testdata path: %v", err)
		}
		t.Run(testName, func(t *testing.T) {
			tr, exp := parseTestFile(t, testPath)
			testReader(t, tr, exp)
		})
	}
}

func testReader(t *testing.T, tr io.Reader, exp *testtrace.Expectation) {
	r, err := trace.NewReader(tr)
	if err != nil {
		if err := exp.Check(err); err != nil {
			t.Error(err)
		}
		return
	}
	v := testtrace.NewValidator()
	for {
		ev, err := r.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			if err := exp.Check(err); err != nil {
				t.Error(err)
			}
			return
		}
		if *logEvents {
			t.Log(ev.String())
		}
		if err := v.Event(ev); err != nil {
			t.Error(err)
		}
	}
	if err := exp.Check(nil); err != nil {
		t.Error(err)
	}
}

func parseTestFile(t *testing.T, testPath string) (io.Reader, *testtrace.Expectation) {
	t.Helper()

	ar, err := txtar.ParseFile(testPath)
	if err != nil {
		t.Fatalf("failed to read test file for %s: %v", testPath, err)
	}
	if len(ar.Files) != 2 {
		t.Fatalf("malformed test %s: wrong number of files", testPath)
	}
	if ar.Files[0].Name != "expect" {
		t.Fatalf("malformed test %s: bad filename %s", testPath, ar.Files[0].Name)
	}
	if ar.Files[1].Name != "trace" {
		t.Fatalf("malformed test %s: bad filename %s", testPath, ar.Files[1].Name)
	}
	tr, err := raw.NewTextReader(bytes.NewReader(ar.Files[1].Data))
	if err != nil {
		t.Fatalf("malformed test %s: bad trace file: %v", testPath, err)
	}
	var buf bytes.Buffer
	tw, err := raw.NewWriter(&buf, tr.Version())
	if err != nil {
		t.Fatalf("failed to create trace byte writer: %v", err)
	}
	for {
		ev, err := tr.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("malformed test %s: bad trace file: %v", testPath, err)
		}
		if err := tw.WriteEvent(ev); err != nil {
			t.Fatalf("internal error during %s: failed to write trace bytes: %v", testPath, err)
		}
	}
	exp, err := testtrace.ParseExpectation(ar.Files[0].Data)
	if err != nil {
		t.Fatalf("internal error during %s: failed to parse expectation %q: %v", testPath, string(ar.Files[0].Data), err)
	}
	return &buf, exp
}

func dumpTraceToText(t *testing.T, b []byte) string {
	t.Helper()

	br, err := raw.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatalf("dumping trace: %v", err)
	}
	var sb strings.Builder
	tw, err := raw.NewTextWriter(&sb, version.Go122)
	if err != nil {
		t.Fatalf("dumping trace: %v", err)
	}
	for {
		ev, err := br.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("dumping trace: %v", err)
		}
		if err := tw.WriteEvent(ev); err != nil {
			t.Fatalf("dumping trace: %v", err)
		}
	}
	return sb.String()
}

func dumpTraceToFile(t *testing.T, testName string, stress bool, b []byte) string {
	t.Helper()

	desc := "default"
	if stress {
		desc = "stress"
	}
	relname := fmt.Sprintf("./%s.%s.trace", testName, desc)
	fname, err := filepath.Abs(relname)
	if err != nil {
		t.Fatalf("absolutizing trace file name %q: %v", relname, err)
	}
	if err := os.WriteFile(fname, b, 0o666); err != nil {
		t.Fatalf("writing trace dump to %q: %v", fname, err)
	}
	return fname
}
