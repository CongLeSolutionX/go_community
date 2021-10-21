// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"cmd/internal/cov"
	"fmt"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const prog1 = `
package main

import "os"

//go:noinline
func first() {
  println("whee")
}

//go:noinline
func second() {
  println("oy")
}

//go:noinline
func third(x int) int {
  if x != 0 {
    return 42
  }
  println("blarg")
  return 0
}

func main() {
  if len(os.Args) > 1 {
    second()
    third(1)
  } else {
    first()
    third(0)
  }
}
`

func gobuild(t *testing.T, src string, dst string, flags []string) {
	t.Helper()

	args := []string{"build", "-o", dst}
	if len(flags) != 0 {
		args = append(args, flags...)
	}
	args = append(args, src)
	cmd := exec.Command(testenv.GoToolPath(t), args...)
	b, err := cmd.CombinedOutput()
	if len(b) != 0 {
		t.Logf("## build output:\n%s", b)
	}
	if err != nil {
		t.Fatalf("build error: %v", err)
	}
}
func buildProg(t *testing.T, progsrc string, tag string, dir string) string {
	// Emit program.
	src := filepath.Join(dir, tag+".go")
	if err := ioutil.WriteFile(src, []byte(progsrc), 0666); err != nil {
		t.Fatal(err)
	}
	exepath := filepath.Join(dir, tag+".exe")
	gobuild(t, src, exepath, []string{"-coverage"})
	return exepath
}

type state struct {
	dir      string
	exepath1 string
	exepath2 string
	tool     string
	outdirs  []string
}

func TestCovTool(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir := t.TempDir()
	if testing.Short() {
		t.Skip()
	}
	if false {
		dir = "/tmp/qqq"
		os.RemoveAll(dir)
		os.Mkdir(dir, 0777)
	}

	exepath1 := buildProg(t, prog1, "prog1", dir)

	// Build the tool.
	tool := filepath.Join(dir, "tool.exe")
	gobuild(t, ".", tool, []string{})

	// Create a couple of coverage output dirs.
	outdirs := []string{
		filepath.Join(dir, "covdata1"),
		filepath.Join(dir, "covdata2"),
		filepath.Join(dir, "covdata3"),
	}
	for _, outdir := range outdirs {
		if err := os.Mkdir(outdir, 0777); err != nil {
			t.Fatalf("can't create outdir %s: %v", outdir, err)
		}
	}

	// Run instrumented program to generate some coverage data output files.
	for k := 0; k < 2; k++ {
		outdir := outdirs[k]
		args := []string{}
		if k != 0 {
			args = append(args, "foo", "bar")
		}
		cmd := exec.Command(exepath1, args...)
		cmd.Env = append(cmd.Env, "GOCOVERDIR="+outdir)
		b, err := cmd.CombinedOutput()
		if len(b) != 0 {
			t.Logf("## instrumented run output:\n%s", b)
		}
		if err != nil {
			t.Fatalf("instrumented run error: %v", err)
		}
	}

	s := state{
		dir:      dir,
		exepath1: exepath1,
		outdirs:  outdirs,
		tool:     tool,
	}

	// At this point we can fork off a couple of child tests
	// to check different tool modes.
	t.Run("Merge", func(t *testing.T) {
		t.Parallel()
		testMerge(t, s)
	})
	t.Run("Dump", func(t *testing.T) {
		t.Parallel()
		testDump(t, s)
	})
}

func testDump(t *testing.T, s state) {
	// Run the dumper on the two dirs we generated.
	dargs := []string{"dump", "-args", "-pkg=^main$", "-live",
		"-i=" + s.outdirs[0] + "," + s.outdirs[1]}
	//t.Logf("dump cmd is: %+v", dargs)
	cmd := exec.Command(s.tool, dargs...)
	b, err := cmd.CombinedOutput()
	if len(b) == 0 {
		t.Fatalf("dump run produced no output")
	}
	if err != nil {
		t.Fatalf("dump run error: %v", err)
	}

	// Sift through the output to make sure it has some key elements.
	output := string(b)
	lines := strings.Split(output, "\n")
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
		fmt.Printf("output from run: %s\n", output)
	}
}

func testMerge(t *testing.T, s state) {
	// Merge the two dirs into a final result.
	ins := fmt.Sprintf("-i=%s,%s", s.outdirs[0], s.outdirs[1])
	out := fmt.Sprintf("-o=%s", s.outdirs[2])
	margs := []string{"merge", ins, out}
	//t.Logf("merge cmd is: %+v", margs)
	cmd := exec.Command(s.tool, margs...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("merge output: %v\n", string(b))
		t.Fatalf("merge run error: %v", err)
	}
	if len(b) != 0 {
		t.Errorf("merge run produced unexpected output: %s", b)
	}

	// We expect the merge tool to produce exacty two files: a meta
	// data file and a counter file. If we get more than just this one
	// pair, something went wrong.
	pods, err := cov.CollectPods([]string{s.outdirs[2]}, true)
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
	dargs := []string{"dump", "-args", "-pkg=^main$", "-live", "-i=" + s.outdirs[2]}
	//t.Logf("dump cmd is: %+v", dargs)
	cmd = exec.Command(s.tool, dargs...)
	b, err = cmd.CombinedOutput()
	if len(b) == 0 {
		t.Fatalf("dump run produced no output")
	}
	if err != nil {
		t.Fatalf("dump run error: %v", err)
	}

	// Sift through the output to make sure it has some key elements.
	// In particular, we want to see entries for all three functions
	// ("first", "second", and "third"
	output := string(b)
	lines := strings.Split(output, "\n")
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
			regexp.MustCompile(`^Func: first\s*$`),
		},
		{
			"third function",
			regexp.MustCompile(`^Func: first\s*$`),
		},
		{
			"third function unit 0",
			regexp.MustCompile(`^0: L26:C3 -- L26:C21 NS=1 = 2$`),
		},
		{
			"third function unit 1",
			regexp.MustCompile(`^1: L27:C11 -- L28:C11 NS=2 = 1$`),
		},
		{
			"third function unit 2",
			regexp.MustCompile(`^2: L30:C10 -- L31:C11 NS=2 = 1$`),
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
		fmt.Printf("output from run: %s\n", output)
	}
}
