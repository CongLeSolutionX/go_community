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
	"strings"
	"testing"
	"time"
)

var remasterflag = flag.Bool("update-expected", false, "if true, generate updated golden results in testcases for all props tests")

func interestingToCompare(fname string) bool {
	if strings.HasPrefix(fname, "init.") {
		return true
	}
	if strings.HasPrefix(fname, "T_") {
		return true
	}
	f := strings.Split(fname, ".")
	if len(f) == 2 && strings.HasPrefix(f[1], "T_") {
		return true
	}
	return false
}

func TestFuncProperties(t *testing.T) {
	td := t.TempDir()
	//td = "/tmp/qqq"
	//os.RemoveAll(td)
	//os.Mkdir(td, 0777)
	testenv.MustHaveGoBuild(t)

	// NOTE: this testpoint has the unfortunate characteristic that it
	// relies on the installed compiler, meaning that if you make
	// changes to the inline heuristics code in your working copy and
	// then run the test, it will test the installed compiler and not
	// your local modifications. TODO: decide whether to convert this
	// to building a fresh compiler on the fly, or using some other
	// scheme.

	testcases := []string{"funcflags"}

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
			dentry, err := dreader.readEntry()
			if err != nil {
				t.Fatalf("reading func prop dump: %v", err)
			}
			if dentry == nil || dentry.fname == "" {
				// end of file
				break
			}
			if !interestingToCompare(dentry.fname) {
				continue
			}
			eentry, err := ereader.readEntry()
			if err != nil {
				t.Fatalf("reading expected func prop dump: %v", err)
			}
			if eentry == nil {
				t.Errorf("missing expected results for %q, skipping",
					dentry.fname)
				continue
			}
			if dentry.fname != eentry.fname {
				t.Errorf("got fn %q wanted %q, skipping checks",
					dentry.fname, eentry.fname)
				continue
			}
			compareEntries(t, tc, dentry, eentry)
		}
	}
}

// TODO: replace returnsToString and paramsToString with a single
// generic function once generics available in Go bootstrap compiler.

func returnsToString(rtns []ReturnPropBits) string {
	var sb strings.Builder
	for i, f := range rtns {
		fmt.Fprintf(&sb, "%d: %s\n", i, f.String())
	}
	return sb.String()
}

func paramsToString(params []ParamPropBits) string {
	var sb strings.Builder
	for i, f := range params {
		fmt.Fprintf(&sb, "%d: %s\n", i, f.String())
	}
	return sb.String()
}

func compareEntries(t *testing.T, tc string, dentry *fnInlHeur, eentry *fnInlHeur) {
	dfp := dentry.props
	efp := eentry.props
	dfn := dentry.fname

	// Compare function flags.
	if dfp.Flags != efp.Flags {
		t.Errorf("testcase %s: Flags mismatch for %q: got %s, wanted %s",
			tc, dfn, dfp.Flags.String(), efp.Flags.String())
	}
	// Compare returns
	rgot := returnsToString(dfp.ReturnFlags)
	rwant := returnsToString(efp.ReturnFlags)
	if rgot != rwant {
		t.Errorf("Returns mismatch for %q: got:\n%swant:\n%s",
			dfn, rgot, rwant)
	}
	// Compare receiver + params.
	pgot := paramsToString(dfp.RecvrParamFlags)
	pwant := paramsToString(efp.RecvrParamFlags)
	if pgot != pwant {
		t.Errorf("Params mismatch for %q: got:\n%swant:\n%s",
			dfn, pgot, pwant)
	}
}

type dumpReader struct {
	s  *bufio.Scanner
	t  *testing.T
	p  string
	ln int
}

func makeDumpReader(t *testing.T, path string) (*dumpReader, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := &dumpReader{
		s:  bufio.NewScanner(strings.NewReader(string(content))),
		t:  t,
		p:  path,
		ln: 1,
	}
	// consume header comment until preamble delimiter.
	found := false
	for r.scan() {
		if r.curLine() == preambleDelimiter {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("malformed testcase file %s, missing preamble delimiter",
			path)
	}
	return r, nil
}

func (dr *dumpReader) scan() bool {
	v := dr.s.Scan()
	if v {
		dr.ln++
	}
	return v
}

func (dr *dumpReader) curLine() string {
	res := strings.TrimSpace(dr.s.Text())
	if !strings.HasPrefix(res, "// ") {
		dr.t.Fatalf("malformed line %s:%d, no comment: %s", dr.p, dr.ln, res)
	}
	return res[3:]
}

// readObjBlob reads in a series of commented lines until
// it hits a delimiter, then returns the contents of the comments.
func (dr *dumpReader) readObjBlob(delim string) (string, error) {
	var sb strings.Builder
	foundDelim := false
	for dr.scan() {
		line := dr.curLine()
		if delim == line {
			foundDelim = true
			break
		}
		sb.WriteString(line + "\n")
	}
	if err := dr.s.Err(); err != nil {
		return "", err
	}
	if !foundDelim {
		return "", fmt.Errorf("malformed input %s, missing delimiter %q",
			dr.p, delim)
	}
	return sb.String(), nil
}

// readEntry reads a single function's worth of material from
// a file produced by the "-d=dumpinlfuncprops=..." command line
// flag. It deserializes the json for the func properties and
// returns the resulting properties and function name. EOF is
// signaled by a nil FuncProps return (with no error
func (dr *dumpReader) readEntry() (*fnInlHeur, error) {
	var fih fnInlHeur
	if !dr.scan() {
		return nil, nil
	}
	// first line contains info about function: file/name/line
	info := dr.curLine()
	chunks := strings.Fields(info)
	fih.file = chunks[0]
	fih.fname = chunks[1]
	if _, err := fmt.Sscanf(chunks[2], "%d", &fih.line); err != nil {
		return nil, err
	}
	// consume comments until and including delimiter
	for {
		if !dr.scan() {
			break
		}
		if dr.curLine() == comDelimiter {
			break
		}
	}

	// Consume JSON for encoded props.
	dr.scan()
	line := dr.curLine()
	fp := &FuncProps{}
	if err := json.Unmarshal([]byte(line), fp); err != nil {
		return nil, err
	}
	fih.props = fp

	// Consume delimiter.
	dr.scan()
	line = dr.curLine()
	if line != fnDelimiter {
		return nil, fmt.Errorf("malformed testcase file %q, missing delimiter %q", dr.p, fnDelimiter)
	}

	return &fih, nil
}

// gatherPropsDumpForFile builds the specified testcase 'testcase' from
// testdata/props passing the "-d=dumpinlfuncprops=..." compiler option,
// to produce a properties dump, then returns the path of the newly
// created file. NB: we can't use "go tool compile" here, since
// some of the test cases import stdlib packages (such as "os").
// This means using "go build", which is problematic since the
// Go command can potentially cache the results of the compile step,
// causing the test to fail when being run interactively. E.g.
//
//	$ rm -f dump.txt
//	$ go build -gcflags=-d=dumpinlfuncprops=dump.txt foo.go
//	$ rm -f dump.txt
//	$ echo "var G int" >> foo.go
//	$ go build -gcflags=-d=dumpinlfuncprops=dump.txt foo.go
//	$ ls dump.txt
//	ls: cannot access 'dump.txt': No such file or directory
//	$
//
// For this reason, pick a unique filename for the dump, so as to
// defeat the caching.
func gatherPropsDumpForFile(t *testing.T, testcase string, td string) (string, error) {
	t.Helper()
	gopath := "testdata/props/" + testcase + ".go"
	outpath := filepath.Join(td, testcase+".a")
	salt := fmt.Sprintf(".p%dt%d", os.Getpid(), time.Now().UnixNano())
	dumpfile := filepath.Join(td, testcase+salt+".dump.txt")
	run := []string{testenv.GoToolPath(t), "build",
		"-gcflags=-d=dumpinlfuncprops=" + dumpfile, "-o", outpath, gopath}
	out, err := testenv.Command(t, run[0], run[1:]...).CombinedOutput()
	if strings.TrimSpace(string(out)) != "" {
		t.Logf("%s", out)
	}
	return dumpfile, err
}

// genExpected reads in a given Go testcase file, strips out all the
// unindented (column 0) commands, writes them out to a new file, and
// returns the path of that new file.
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

// updateExpected takes a given Go testcase file X.go and writes out
// a new/updated version of the file to X.go.new, where the
// column-0 "expected" comments have been updated.
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

	// Write file preamble with "DO NOT EDIT" message and such.
	var sb strings.Builder
	dumpFilePreamble(&sb)
	newgolines = append(newgolines,
		strings.Split(strings.TrimSpace(sb.String()), "\n")...)

	for _, line := range golines {
		if strings.HasPrefix(line, "func ") {

			// We have a function definition. Read the corresponding entry in the dump.
			dentry, err := dr.readEntry()
			if err != nil {
				t.Fatalf("reading func prop dump: %v", err)
			}
			if dentry == nil {
				// short dump? should never happen.
				t.Fatalf("reading func prop dump: short read")
			}

			// Decide whether this func is compare-worthy.
			if interestingToCompare(dentry.fname) {

				// Emit preamble for function.
				var sb strings.Builder
				dumpFnPreamble(&sb, dentry)
				newgolines = append(newgolines,
					strings.Split(strings.TrimSpace(sb.String()), "\n")...)
			}
		}

		// Consume all existing comments comments.
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
