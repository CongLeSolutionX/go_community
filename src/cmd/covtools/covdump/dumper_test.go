// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
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

const himom = `
package main
func main() {
  println("hi mom")
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

func TestBasicDumper(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir := t.TempDir()
	if testing.Short() {
		t.Skip()
	}

	// Emit himom program.
	src := filepath.Join(dir, "test.go")
	if err := ioutil.WriteFile(src, []byte(himom), 0666); err != nil {
		t.Fatal(err)
	}

	// Do a coverage build of our himom program.
	dst := filepath.Join(dir, "out.exe")
	gobuild(t, src, dst, []string{"-coverage"})

	// Build the dumper.
	dumper := filepath.Join(dir, "dumper.exe")
	gobuild(t, ".", dumper, []string{})

	// Create a coverage output dir.
	outdir := filepath.Join(dir, "covdata")
	if err := os.Mkdir(outdir, 0777); err != nil {
		t.Fatalf("can't create outdir %s", outdir)
	}

	// Run instrumented program to generate some coverage data output files.
	cmd := exec.Command(dst, []string{}...)
	cmd.Env = append(cmd.Env, "GOCOVERDIR="+outdir)
	b, err := cmd.CombinedOutput()
	if len(b) != 0 {
		t.Logf("## instrumented run output:\n%s", b)
	}
	if err != nil {
		t.Fatalf("instrumented run error: %v", err)
	}

	// Dump the files in the coverage output dir and
	// examine the output.
	dargs := []string{"-args", "-live", "-h", "-dir", outdir}
	cmd = exec.Command(dumper, dargs...)
	b, err = cmd.CombinedOutput()
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
