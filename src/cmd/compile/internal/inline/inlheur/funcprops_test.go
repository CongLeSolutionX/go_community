// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"bufio"
	"cmd/compile/internal/inline/funcprop"
	"encoding/json"
	"flag"
	"fmt"
	"internal/testenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var remasterflag = flag.Bool("update-expected", false, "if true, generate updated golden (*.expected) files for all props tests")

func TestFuncProperties(t *testing.T) {
	td := t.TempDir()
	testenv.MustHaveGoBuild(t)

	// NOTE: this testpoint has the unfortunate characteristic that it
	// relies on the installed compiler, meaning that if you make
	// changes to the inline heuristics code in your working copy
	// and then run the test, it will test the installed compiler
	// and not your local modifications. TODO: decide whether
	// to convert this to building a fresh compiler on the fly, or
	// using some other scheme.

	testcases := []string{"funcflags", "returns", "params"}

	for _, tc := range testcases {
		epath := "testdata/props/" + tc + ".expected"
		dumpfile, err := gatherPropsDumpForFile(t, tc, td)
		if err != nil {
			t.Fatalf("dumping func props for %q: error %v", tc, err)
		}
		if *remasterflag {
			enew := epath + ".new"
			t.Logf("update-expected: copying %s to %s\n", dumpfile, enew)
			t.Logf("please compare the two files, then overwrite %s with %s\n",
				epath, enew)
			run := []string{"cp", dumpfile, enew}
			out, err := testenv.Command(t, run[0], run[1:]...).CombinedOutput()
			t.Logf("%s", out)
			if err != nil {
				t.Fatalf("dump copy failed: %v", err)
			}
			continue
		}
		// Compare the dump file with expected file
		dreader, derr := makeDumpReader(dumpfile)
		if derr != nil {
			t.Fatalf("opening func prop dump: %v", derr)
		}
		ereader, eerr := makeDumpReader(epath)
		if eerr != nil {
			t.Fatalf("opening expected func prop dump: %v", eerr)
		}
		for {
			dfn, dentry, err := dreader.readEntry()
			if err != nil {
				t.Fatalf("reading func prop dump: %v", err)
			}
			if dentry == nil || dfn == "" {
				// end of file
				break
			}
			efn, eentry, err := ereader.readEntry()
			if err != nil {
				t.Fatalf("reading expected func prop dump: %v", err)
			}
			if dfn != efn {
				t.Errorf("got fn %q wanted %q, skipping checks", dfn, efn)
				continue
			}
			if dfn[0] != 'T' {
				// only funcs starting with T are of interest
			}
			compareEntries(t, tc, dfn, dentry, efn, eentry)
		}
	}
}

// TODO: replace returnsToString and paramsToString with a single
// generic function once generics available in Go bootstrap compiler.

func returnsToString(rtns []funcprop.ReturnPropBits) string {
	var sb strings.Builder
	for i, f := range rtns {
		fmt.Fprintf(&sb, "%d: %s\n", i, f.String())
	}
	return sb.String()
}

func paramsToString(params []funcprop.ParamPropBits) string {
	var sb strings.Builder
	for i, f := range params {
		fmt.Fprintf(&sb, "%d: %s\n", i, f.String())
	}
	return sb.String()
}

func compareEntries(t *testing.T, tc string, dfn string, dentry *funcprop.FuncProps, efn string, eentry *funcprop.FuncProps) {
	// Compare function flags.
	if dentry.Flags != eentry.Flags {
		t.Errorf("testcase %s: Flags mismatch for %q: got %s, wanted %s",
			tc, dfn, dentry.Flags.String(), eentry.Flags.String())
	}
	// Compare returns
	rgot := returnsToString(dentry.ReturnFlags)
	rwant := returnsToString(eentry.ReturnFlags)
	if rgot != rwant {
		t.Errorf("Returns mismatch for %q: got:\n%swant:\n%s",
			dfn, rgot, rwant)
	}
	// Compare receiver + params.
	pgot := paramsToString(dentry.RecvrParamFlags)
	pwant := paramsToString(eentry.RecvrParamFlags)
	if pgot != pwant {
		t.Errorf("Params mismatch for %q: got:\n%swant:\n%s",
			dfn, pgot, pwant)
	}
}

type dumpReader struct {
	s *bufio.Scanner
}

func makeDumpReader(path string) (*dumpReader, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := &dumpReader{
		s: bufio.NewScanner(strings.NewReader(string(content))),
	}
	// consume header comment
	r.s.Scan()
	return r, nil
}

// readEntry reads a single function's worth of material from
// a file produced by the "-d=dumpinlfuncprops=..." command line
// flag. It deserializes the json for the func properties and
// returns the resulting FuncProps and function name. EOF is
// signaled by a nil FuncProps return (with no error
func (dr *dumpReader) readEntry() (string, *funcprop.FuncProps, error) {
	dr.s.Scan()
	fname := strings.TrimSpace(dr.s.Text())
	// consume comments
	for {
		dr.s.Scan()
		if !strings.HasPrefix(dr.s.Text(), "//") {
			break
		}
	}
	var sb strings.Builder
	foundDelim := false
	for dr.s.Scan() {
		line := strings.TrimSpace(dr.s.Text())
		if fnDelimiter == line {
			foundDelim = true
			break
		}
		sb.WriteString(dr.s.Text() + "\n")
	}
	if err := dr.s.Err(); err != nil {
		return "", nil, err
	}
	if !foundDelim {
		return "", nil, nil
	}
	var fp funcprop.FuncProps
	if err := json.Unmarshal([]byte(sb.String()), &fp); err != nil {
		return "", nil, err
	}
	return fname, &fp, nil
}

func gatherPropsDumpForFile(t *testing.T, testcase string, td string) (string, error) {
	t.Helper()
	gopath := "testdata/props/" + testcase + ".go"
	outpath := filepath.Join(td, testcase+".a")
	dumpfile := filepath.Join(td, testcase+".dump.txt")
	run := []string{testenv.GoToolPath(t), "build",
		"-gcflags=-m -d=dumpinlfuncprops=" + dumpfile, "-o", outpath, gopath}
	out, err := testenv.Command(t, run[0], run[1:]...).CombinedOutput()
	if strings.TrimSpace(string(out)) != "" {
		t.Logf("%s", out)
	}
	return dumpfile, err
}
