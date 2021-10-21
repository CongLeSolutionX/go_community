// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"cmd/internal/cov"
	"fmt"
	"internal/goexperiment"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const debugtrace = true

func gobuild(t *testing.T, indir string, dst string, flags []string) {
	t.Helper()

	args := []string{"build", "-o", dst}
	if len(flags) != 0 {
		args = append(args, flags...)
	}
	if debugtrace {
		if indir != "" {
			t.Logf("in dir %s: ", indir)
		}
		t.Logf("cmd: %s %+v\n", testenv.GoToolPath(t), args)
	}
	cmd := exec.Command(testenv.GoToolPath(t), args...)
	cmd.Dir = indir
	b, err := cmd.CombinedOutput()
	if len(b) != 0 {
		t.Logf("## build output:\n%s", b)
	}
	if err != nil {
		t.Fatalf("build error: %v", err)
	}
}

func buildProg(t *testing.T, prog string, dir string) (string, string) {

	// Create subdir.
	subdir := filepath.Join(dir, prog+"dir")
	if err := os.Mkdir(subdir, 0777); err != nil {
		t.Fatalf("can't create outdir %s: %v", subdir, err)
	}

	// Emit program.
	insrc := filepath.Join("testdata", prog+".go")
	payload, err := ioutil.ReadFile(insrc)
	if err != nil {
		t.Fatalf("error reading %q: %v", insrc, err)
	}
	src := filepath.Join(subdir, prog+".go")
	if err := ioutil.WriteFile(src, payload, 0666); err != nil {
		t.Fatalf("writing %q: %v", src, err)
	}

	// Emit go.mod.
	mod := filepath.Join(subdir, "go.mod")
	modsrc := fmt.Sprintf("\nmodule %s\n\ngo 1.18\n", prog)
	if err := ioutil.WriteFile(mod, []byte(modsrc), 0666); err != nil {
		t.Fatal(err)
	}
	exepath := filepath.Join(subdir, prog+".exe")
	gobuild(t, subdir, exepath, []string{"-cover"})
	return exepath, subdir
}

type state struct {
	dir      string
	exedir1  string
	exedir2  string
	exepath1 string
	exepath2 string
	tool     string
	outdirs  [2]string
}

const debugWorkDir = false

func TestCovTool(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	if !goexperiment.CoverageRedesign {
		t.Skipf("stubbed out due to goexperiment.CoverageRedesign=false")
	}
	dir := t.TempDir()
	if testing.Short() {
		t.Skip()
	}
	if debugWorkDir {
		// debugging
		dir = "/tmp/qqq"
		os.RemoveAll(dir)
		os.Mkdir(dir, 0777)
	}

	s := state{
		dir: dir,
	}
	s.exepath1, s.exedir1 = buildProg(t, "prog1", dir)

	// Build the tool.
	s.tool = filepath.Join(dir, "tool.exe")
	gobuild(t, "", s.tool, []string{"."})

	// Create a few coverage output dirs.
	for i := 0; i < 2; i++ {
		d := filepath.Join(dir, fmt.Sprintf("covdata%d", i))
		s.outdirs[i] = d
		if err := os.Mkdir(d, 0777); err != nil {
			t.Fatalf("can't create outdir %s: %v", d, err)
		}
	}

	// Run instrumented program to generate some coverage data output files.
	for k := 0; k < 2; k++ {
		args := []string{}
		if k != 0 {
			args = append(args, "foo", "bar")
		}
		cmd := exec.Command(s.exepath1, args...)
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+s.outdirs[k])
		b, err := cmd.CombinedOutput()
		if len(b) != 0 {
			t.Logf("## instrumented run output:\n%s", b)
		}
		if err != nil {
			t.Fatalf("instrumented run error: %v", err)
		}
	}

	// At this point we can fork off a bunch of child tests
	// to check different tool modes.
	t.Run("MergeSimple", func(t *testing.T) {
		t.Parallel()
		testMergeSimple(t, s)
	})
	t.Run("MergePcombine", func(t *testing.T) {
		t.Parallel()
		testMergeCombinePrograms(t, s)
	})
	t.Run("Dump", func(t *testing.T) {
		t.Parallel()
		testDump(t, s)
	})
	t.Run("Subtract", func(t *testing.T) {
		t.Parallel()
		testSubtract(t, s)
	})
}

const showToolInvocations = true

func runToolOp(t *testing.T, s state, op string, args []string) []string {
	// Perform tool run.
	t.Helper()
	args = append([]string{op}, args...)
	if showToolInvocations {
		t.Logf("%s cmd is: %s %+v", op, s.tool, args)
	}
	cmd := exec.Command(s.tool, args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "## %s output: %s\n", op, string(b))
		t.Fatalf("%q run error: %v", op, err)
	}
	output := strings.TrimSpace(string(b))
	lines := strings.Split(output, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	return lines
}

func testDump(t *testing.T, s state) {
	// Run the dumper on the two dirs we generated.
	dargs := []string{"-args", "-emitperc", "-emitdump", "-pkg=^main$", "-live",
		"-i=" + s.outdirs[0] + "," + s.outdirs[1]}
	lines := runToolOp(t, s, "dump", dargs)

	// Sift through the output to make sure it has some key elements.
	testpoints := []struct {
		tag string
		re  *regexp.Regexp
	}{
		{
			"args",
			regexp.MustCompile(`^data file \S+ program args: .+$`),
		},
		{
			"main package",
			regexp.MustCompile(`^Package: main\s*$`),
		},
		{
			"main function",
			regexp.MustCompile(`^Func: main\s*$`),
		},
		{
			"statement coverage percent",
			regexp.MustCompile(`coverage: \d+\.\d% of statements\s*$`),
		},
	}

	bad := false
	for _, testpoint := range testpoints {
		found := false
		for _, line := range lines {
			if m := testpoint.re.FindStringSubmatch(line); m != nil {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("dump output regexp match failed for %s", testpoint.tag)
			bad = true
		}
	}
	if bad {
		fmt.Printf("output from run: %v\n", lines)
	}
}

func dumplines(lines []string) {
	for i := range lines {
		fmt.Fprintf(os.Stderr, "%s\n", lines[i])
	}
}

func testMergeSimple(t *testing.T, s state) {

	outdir := filepath.Join(s.dir, "simpleMergeOut")
	if err := os.Mkdir(outdir, 0777); err != nil {
		t.Fatalf("can't create outdir %s: %v", outdir, err)
	}

	// Merge the two dirs into a final result.
	ins := fmt.Sprintf("-i=%s,%s", s.outdirs[0], s.outdirs[1])
	out := fmt.Sprintf("-o=%s", outdir)
	margs := []string{ins, out}
	lines := runToolOp(t, s, "merge", margs)
	if len(lines) != 0 {
		t.Errorf("merge run produced %d lines of unexpected output", len(lines))
		dumplines(lines)
	}

	// We expect the merge tool to produce exacty two files: a meta
	// data file and a counter file. If we get more than just this one
	// pair, something went wrong.
	pods, err := cov.CollectPods([]string{outdir}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d pods", len(pods))
	}
	ncdfs := len(pods[0].CounterDataFiles)
	if ncdfs != 1 {
		t.Fatalf("expected 1 counter data file, got %d", ncdfs)
	}

	// Dump the files in the merged output dir and examine the result.
	dargs := []string{"-args", "-emitdump", "-pkg=^main$", "-live", "-i=" + outdir}
	lines = runToolOp(t, s, "dump", dargs)
	if len(lines) == 0 {
		t.Fatalf("dump run produced no output")
	}

	// Sift through the output to make sure it has some key elements.
	// In particular, we want to see entries for all three functions
	// ("first", "second", and "third"
	testpoints := []struct {
		tag string
		re  *regexp.Regexp
	}{
		{
			"first function",
			regexp.MustCompile(`^Func: first\s*$`),
		},
		{
			"second function",
			regexp.MustCompile(`^Func: second\s*$`),
		},
		{
			"third function",
			regexp.MustCompile(`^Func: third\s*$`),
		},
		{
			"third function unit 0",
			regexp.MustCompile(`^0: L20:C23 -- L21:C12 NS=1 = 1$`),
		},
		{
			"third function unit 1",
			regexp.MustCompile(`^1: L24:C2 -- L25:C10 NS=2 = 1$`),
		},
		{
			"third function unit 2",
			regexp.MustCompile(`^2: L21:C12 -- L23:C3 NS=1 = 1$`),
		},
	}

	bad := false
	for _, testpoint := range testpoints {
		found := false
		for _, line := range lines {
			if m := testpoint.re.FindStringSubmatch(line); m != nil {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("dump output regexp match failed for %s", testpoint.tag)
			bad = true
		}
	}
	if bad {
		fmt.Printf("output from 'dump' run:\n")
		dumplines(lines)
	}
}

func testMergeCombinePrograms(t *testing.T, s state) {

	// Build a new test program.
	s.exepath2, s.exedir2 = buildProg(t, "prog2", s.dir)

	// Run the new program, emitting output into a new set
	// of outdirs.
	runout := [2]string{}
	for k := 0; k < 2; k++ {
		runout[k] = filepath.Join(s.dir, fmt.Sprintf("newcovdata%d", k))
		if err := os.Mkdir(runout[k], 0777); err != nil {
			t.Fatalf("can't create outdir %s: %v", runout[k], err)
		}
		args := []string{}
		if k != 0 {
			args = append(args, "foo", "bar")
		}
		cmd := exec.Command(s.exepath2, args...)
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+runout[k])
		b, err := cmd.CombinedOutput()
		if len(b) != 0 {
			t.Logf("## instrumented run output:\n%s", b)
		}
		if err != nil {
			t.Fatalf("instrumented run error: %v", err)
		}
	}

	// Create out dir for -pcombine merge.
	moutdir := filepath.Join(s.dir, "mergeCombineOut")
	if err := os.Mkdir(moutdir, 0777); err != nil {
		t.Fatalf("can't create outdir %s: %v", moutdir, err)
	}

	// Run a merge over both programs, using the -pcombine
	// flag to do maximal combining.
	ins := fmt.Sprintf("-i=%s,%s,%s,%s", s.outdirs[0], s.outdirs[1],
		runout[0], runout[1])
	out := fmt.Sprintf("-o=%s", moutdir)
	margs := []string{"-pcombine", ins, out}
	lines := runToolOp(t, s, "merge", margs)
	if len(lines) != 0 {
		t.Errorf("merge run produced unexpected output: %v", lines)
	}

	// We expect the merge tool to produce exacty two files: a meta
	// data file and a counter file. If we get more than just this one
	// pair, something went wrong.
	pods, err := cov.CollectPods([]string{moutdir}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(pods) != 1 {
		t.Fatalf("expected 1 pod, got %d pods", len(pods))
	}
	ncdfs := len(pods[0].CounterDataFiles)
	if ncdfs != 1 {
		t.Fatalf("expected 1 counter data file, got %d", ncdfs)
	}

	// Dump the files in the merged output dir and examine the result.
	dargs := []string{"-args", "-emitdump", "-pkg=^main$", "-live", "-i=" + moutdir}
	lines = runToolOp(t, s, "dump", dargs)
	if len(lines) == 0 {
		t.Fatalf("dump run produced no output")
	}

	// Sift through the output to make sure it has some key elements.
	testpoints := []struct {
		tag string
		re  *regexp.Regexp
	}{
		{
			"first function",
			regexp.MustCompile(`^Func: first\s*$`),
		},
		{
			"sixth function",
			regexp.MustCompile(`^Func: sixth\s*$`),
		},
	}

	bad := false
	for _, testpoint := range testpoints {
		found := false
		for _, line := range lines {
			if m := testpoint.re.FindStringSubmatch(line); m != nil {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("dump output regexp match failed for %s", testpoint.tag)
			bad = true
		}
	}
	if bad {
		fmt.Printf("output from run: %v\n", lines)
	}
}

func testSubtract(t *testing.T, s state) {
	// Create out dir for subtract merge.
	soutdir := filepath.Join(s.dir, "subtractOut")
	if err := os.Mkdir(soutdir, 0777); err != nil {
		t.Fatalf("can't create outdir %s: %v", soutdir, err)
	}

	// Subtract the two dirs into a final result.
	ins := fmt.Sprintf("-i=%s,%s", s.outdirs[0], s.outdirs[1])
	out := fmt.Sprintf("-o=%s", soutdir)
	sargs := []string{ins, out}
	lines := runToolOp(t, s, "subtract", sargs)
	if len(lines) != 0 {
		t.Errorf("subtract run produced unexpected output: %+v", lines)
	}

	// Dump the files in the subtract output dir and examine the result.
	dargs := []string{"-args", "-emitdump", "-pkg=^main$", "-live", "-i=" + soutdir}
	lines = runToolOp(t, s, "dump", dargs)
	if len(lines) == 0 {
		t.Errorf("dump run produced no output")
	}

	// Vet the output.
	testpoints := []struct {
		tag       string
		re        *regexp.Regexp
		wantfound bool
	}{
		{
			tag:       "first function",
			re:        regexp.MustCompile(`^Func: first\s*$`),
			wantfound: true,
		},
		{
			tag:       "second function",
			re:        regexp.MustCompile(`^Func: second\s*$`),
			wantfound: false,
		},
		{
			tag:       "third function",
			re:        regexp.MustCompile(`^Func: third\s*$`),
			wantfound: true,
		},
		{
			tag:       "third function unit 0",
			re:        regexp.MustCompile(`^0: L20:C23 -- L21:C12 NS=1 = 0$`),
			wantfound: true,
		},
		{
			tag:       "third function unit 1",
			re:        regexp.MustCompile(`^1: L24:C2 -- L25:C10 NS=2 = 1$`),
			wantfound: true,
		},
		{
			tag:       "third function unit 2",
			re:        regexp.MustCompile(`^2: L21:C12 -- L23:C3 NS=1 = 0$`),
			wantfound: true,
		},
	}

	bad := false
	for _, testpoint := range testpoints {
		found := false
		for _, line := range lines {
			if m := testpoint.re.FindStringSubmatch(line); m != nil {
				found = true
				break
			}
		}

		if found != testpoint.wantfound {
			t.Errorf("dump output regexp match failed for %s", testpoint.tag)
			bad = true
		}
	}
	if bad {
		fmt.Printf("output from run:\n")
		dumplines(lines)
	}
}
