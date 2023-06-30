// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"internal/testenv"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var remasterflag = flag.Bool("update-expected", false, "if true, generate updated golden results in testcases for all props tests")

func TestFuncProperties(t *testing.T) {
	td := t.TempDir()
	td = "/tmp/qqq"
	os.RemoveAll(td)
	os.Mkdir(td, 0777)
	testenv.MustHaveGoBuild(t)

	// NOTE: this testpoint has the unfortunate characteristic that it
	// relies on the installed compiler, meaning that if you make
	// changes to the inline heuristics code in your working copy and
	// then run the test, it will test the installed compiler and not
	// your local modifications. TODO: decide whether to convert this
	// to building a fresh compiler on the fly, or using some other
	// scheme.

	testcases := []string{"funcflags", "returns"}

	for _, tc := range testcases {
		dumpfile, err := gatherPropsDumpForFile(t, tc, td)
		if err != nil {
			t.Fatalf("dumping func props for %q: error %v", tc, err)
		}
		// Open the newly generated dump.
		dreader, derr := makeDumpReader(t, dumpfile)
		if derr != nil {
			t.Fatalf("opening func prop dump: %v", derr)
		}
		if *remasterflag {
			updateExpected(t, tc, dreader)
			continue
		}
		// Generate expected dump.
		epath, gerr := genExpected(td, tc)
		if gerr != nil {
			t.Fatalf("generating expected func prop dump: %v", gerr)
		}
		// Create reader for expected results.
		ereader, eerr := makeDumpReader(t, epath)
		if eerr != nil {
			t.Fatalf("opening expected func prop dump: %v", eerr)
		}
		// Compare new vs expected.
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

func returnsToString(rtns []ReturnPropBits) string {
	var sb strings.Builder
	for i, f := range rtns {
		fmt.Fprintf(&sb, "%d: %s\n", i, f.String())
	}
	return sb.String()
}

func compareEntries(t *testing.T, tc string, dfn string, dentry *FuncProps, efn string, eentry *FuncProps) {
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
	// everything else not yet implemented
	if len(dentry.RecvrParamFlags) != 0 || len(eentry.RecvrParamFlags) != 0 {
		t.Fatalf("testcase %s func %q prop miscompare", tc, dfn)
	}
}

type dumpReader struct {
	s *bufio.Scanner
	t *testing.T
	p string
}

func makeDumpReader(t *testing.T, path string) (*dumpReader, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := &dumpReader{
		s: bufio.NewScanner(strings.NewReader(string(content))),
		t: t,
		p: path,
	}
	// consume header comment
	r.s.Scan()
	return r, nil
}

func (dr *dumpReader) curLine() string {
	res := strings.TrimSpace(dr.s.Text())
	if !strings.HasPrefix(res, "// ") {
		dr.t.Fatalf("malformed line %q in path %s, no comment", res, dr.p)
	}
	return res[3:]
}

// readEntry reads a single function's worth of material from
// a file produced by the "-d=dumpinlfuncprops=..." command line
// flag. It deserializes the json for the func properties and
// returns the resulting FuncProps and function name. EOF is
// signaled by a nil FuncProps return (with no error
func (dr *dumpReader) readEntry() (string, *FuncProps, error) {
	if !dr.s.Scan() {
		return "", nil, nil
	}
	fname := dr.curLine()
	//fmt.Fprintf(os.Stderr, "=-= readEntry fname=%q\n", fname)
	// consume comments until delimiter
	for {
		if !dr.s.Scan() {
			break
		}
		if dr.curLine() == comDelimiter {
			break
		}
	}
	var sb strings.Builder
	foundFnDelim := false
	for dr.s.Scan() {
		line := dr.curLine()
		//fmt.Fprintf(os.Stderr, "=-= read line %s from %s\n", line, dr.p)
		if fnDelimiter == line {
			foundFnDelim = true
			break
		}
		sb.WriteString(line + "\n")
	}
	if err := dr.s.Err(); err != nil {
		return "", nil, err
	}
	if !foundFnDelim {
		return "", nil, nil
	}
	fp := &FuncProps{}
	//fmt.Fprintf(os.Stderr, "=-= passing to json: %s\n", sb.String())
	if err := json.Unmarshal([]byte(sb.String()), fp); err != nil {
		return "", nil, err
	}
	//fmt.Fprintf(os.Stderr, "=-= readEntry returning fn=%s fp=%s\n", fname, fp)
	return fname, fp, nil
}

func gatherPropsDumpForFile(t *testing.T, testcase string, td string) (string, error) {
	t.Helper()
	gopath := "testdata/props/" + testcase + ".go"
	outpath := filepath.Join(td, testcase+".a")
	x := strings.ReplaceAll(t.TempDir(), "/", ":")
	dumpfile := filepath.Join(td, testcase+x+".dump.txt")
	run := []string{testenv.GoToolPath(t), "build",
		"-gcflags=-d=dumpinlfuncprops=" + dumpfile, "-o", outpath, gopath}
	out, err := testenv.Command(t, run[0], run[1:]...).CombinedOutput()
	if strings.TrimSpace(string(out)) != "" {
		t.Logf("%s", out)
	}
	return dumpfile, err
}

func genExpected(td string, testcase string) (string, error) {
	epath := filepath.Join(td, testcase+".expected")
	outf, err := os.OpenFile(epath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	gopath := "testdata/props/" + testcase + ".go"
	content, err := os.ReadFile(gopath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines[3:] {
		if !strings.HasPrefix(line, "// ") {
			continue
		}
		fmt.Fprintf(outf, "%s\n", line)
	}
	if err := outf.Close(); err != nil {
		return "", err
	}
	return epath, nil
}

func updateExpected(t *testing.T, testcase string, dr *dumpReader) {
	gopath := "testdata/props/" + testcase + ".go"
	newgopath := "testdata/props/" + testcase + ".go.new"

	// Read the existing Go file.
	content, err := os.ReadFile(gopath)
	if err != nil {
		t.Fatalf("opening %s: %v", gopath, err)
	}
	golines := strings.Split(string(content), "\n")

	newgolines := []string{}

	// Preserve copyright.
	newgolines = append(newgolines, golines[:4]...)
	if !strings.HasPrefix(golines[0], "// Copyright") {
		t.Fatalf("missing copyright from existing testcase")
	}
	golines = golines[4:]

	// Add a "do not edit" message.
	newgolines = append(newgolines,
		"// DO NOT EDIT COMMENTS (use 'go test -v -update-expected' instead)")

	fre := regexp.MustCompile(`^\s*func .*(T_\S+)\(.*\)\s.*{\s*$`)
	for _, line := range golines {

		// Look for the start of an important function in the Go file.
		m := fre.FindStringSubmatch(line)
		if m != nil {
			// Found function start. Locate corresponding function
			// in the dump.
			mfunc := m[1]
			dfn, dentry, err := dr.readEntry()
			if err != nil {
				t.Fatalf("reading func prop dump: %v", err)
			}
			if dentry == nil || dfn != mfunc {
				t.Fatalf("expected %s got fn=%s entry=%v", mfunc, dfn, dentry)
			}

			// Emit preamble for function, then first func line.
			var sb strings.Builder
			dumpFnPreamble(&sb, mfunc, dentry)
			dlines := strings.Split(strings.TrimSpace(sb.String()), "\n")
			newgolines = append(newgolines, dlines...)
		}

		if strings.HasPrefix(line, "//") {
			continue
		}
		newgolines = append(newgolines, line)
	}

	// Open new Go file and write contents.
	of, err := os.OpenFile(newgopath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatalf("opening %s: %v", newgopath, err)
	}
	fmt.Fprintf(of, "%s", strings.Join(newgolines, "\n"))
	if err := of.Close(); err != nil {
		t.Fatalf("closing %s: %v", newgopath, err)
	}

	t.Logf("update-expected: emitted updated file %s", newgopath)
	t.Logf("please compare the two files, then overwrite %s with %s\n",
		gopath, newgopath)
}
