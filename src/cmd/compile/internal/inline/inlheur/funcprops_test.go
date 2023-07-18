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
	"strconv"
	"strings"
	"testing"
)

var remasterflag = flag.Bool("update-expected", false, "if true, generate updated golden results in testcases for all props tests")

func TestFuncProperties(t *testing.T) {
	td := t.TempDir()
	// td = "/tmp/qqq"
	// os.RemoveAll(td)
	// os.Mkdir(td, 0777)
	testenv.MustHaveGoBuild(t)

	// NOTE: this testpoint has the unfortunate characteristic that it
	// relies on the installed compiler, meaning that if you make
	// changes to the inline heuristics code in your working copy and
	// then run the test, it will test the installed compiler and not
	// your local modifications. TODO: decide whether to convert this
	// to building a fresh compiler on the fly, or using some other
	// scheme.

	testcases := []string{"funcflags", "returns", "params", "calls"}

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
			if dentry == nil || dentry.fih.fname == "" {
				// end of file
				break
			}
			if !interestingToCompare(dentry.fih.fname) {
				continue
			}
			eentry, err := ereader.readEntry()
			if err != nil {
				t.Fatalf("reading expected func prop dump: %v", err)
			}
			if eentry == nil {
				t.Errorf("missing expected results for %q, skipping",
					dentry.fih.fname)
				continue
			}
			if dentry.fih.fname != eentry.fih.fname {
				t.Errorf("got fn %q wanted %q, skipping checks",
					dentry.fih.fname, eentry.fih.fname)
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

func compareEntries(t *testing.T, tc string, dentry *dumpEntry, eentry *dumpEntry) {
	dfp := dentry.fih.props
	efp := eentry.fih.props
	dfn := dentry.fih.fname

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
	// Compare call sites.
	for k, ve := range eentry.callsites {
		if vd, ok := dentry.callsites[k]; !ok {
			t.Errorf("missing expected callsite %q in func %q",
				dfn, k)
			continue
		} else {
			if vd != ve {
				t.Errorf("callsite %q in func %q: got %s want %s",
					k, dfn, vd.String(), ve.String())
			}
		}
	}
	for k := range dentry.callsites {
		if _, ok := eentry.callsites[k]; !ok {
			t.Errorf("unexpected extra callsite %q in func %q",
				dfn, k)
		}
	}
}

type dumpReader struct {
	s  *bufio.Scanner
	t  *testing.T
	p  string
	ln int
}

type callSiteResult struct {
	flags int
}

type dumpEntry struct {
	fih       fnInlHeur
	callsites encodedCallSiteTab
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
func (dr *dumpReader) readEntry() (*dumpEntry, error) {
	var de dumpEntry
	if !dr.scan() {
		return nil, nil
	}
	// first line contains info about function: file/name/line
	info := dr.curLine()
	chunks := strings.Fields(info)
	de.fih.file = chunks[0]
	de.fih.fname = chunks[1]
	if _, err := fmt.Sscanf(chunks[2], "%d", &de.fih.line); err != nil {
		return nil, fmt.Errorf("scanning line %q: %v", info, err)
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
	de.fih.props = fp

	// Consume callsites.
	de.callsites = make(encodedCallSiteTab)
	for dr.scan() {
		line := dr.curLine()
		if line == csDelimiter {
			break
		}
		// expected format: "// callsite: <expanded pos> <desc> <flags>"
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return nil, fmt.Errorf("malformed callsite %s line %d: %s",
				dr.p, dr.ln, line)
		}
		tag := fields[1]
		flagstr := fields[3]
		flags, err := strconv.Atoi(flagstr)
		if err != nil {
			return nil, fmt.Errorf("bad flags val %s line %d: %q err=%v",
				dr.p, dr.ln, line, err)
		}
		de.callsites[tag] = CSPropBits(flags)
	}

	// Consume function delimiter.
	dr.scan()
	line = dr.curLine()
	if line != fnDelimiter {
		return nil, fmt.Errorf("malformed testcase file %q, missing delimiter %q", dr.p, fnDelimiter)
	}

	return &de, nil
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
				t.Fatalf("reading func prop dump for %q: %v", gopath, err)
			}
			if dentry == nil {
				// short dump? should never happen.
				t.Fatalf("reading func prop dump: short read")
			}

			// Decide whether this func is compare-worthy.
			if interestingToCompare(dentry.fih.fname) {

				// Emit preamble for function.
				var sb strings.Builder
				dumpFnPreamble(&sb, &dentry.fih, dentry.callsites)
				newgolines = append(newgolines,
					strings.Split(strings.TrimSpace(sb.String()), "\n")...)
			}
		}

		// Consume all existing comments comments.
		if strings.HasPrefix(line, "// ") {
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
