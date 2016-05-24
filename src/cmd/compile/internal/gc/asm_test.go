// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"bytes"
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

// TestAssembly checks to make sure the assembly generated for
// functions contains certain expected instructions.
func TestAssembly(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir, err := ioutil.TempDir("", "TestAssembly")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	for _, test := range asmTests {
		asm, err := compileToAsm(dir, test.arch, fmt.Sprintf(template, test.function))
		if err != nil {
			t.Error(err)
			continue
		}
		// Get rid of code for "".init. Also gets rid of type algorithms & other junk.
		if i := strings.Index(asm, "\n\"\".init "); i >= 0 {
			asm = asm[:i+1]
		}
		for _, r := range test.regexps {
			if b, err := regexp.MatchString(r, asm); !b || err != nil {
				t.Errorf("expected:%s\ngo:%s\nasm:%s\n", r, test.function, asm)
			}
		}
	}
}

// compile compiles the package pkg for architecture arch and
// returns the generated assembly.  dir is a scratch directory.
func compileToAsm(dir, arch, pkg string) (string, error) {
	// Create source.
	src := filepath.Join(dir, "test.go")
	f, err := os.Create(src)
	if err != nil {
		return "", fmt.Errorf("could not create source file: %v", err)
	}
	f.Write([]byte(pkg))
	f.Close()

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "tool", "compile", "-S", "-o", filepath.Join(dir, "out.o"), src)
	cmd.Env = append(cmd.Env, "GOARCH="+arch)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Failed: %v:\nout: %s\nerr: %s\n", err, &stdout, &stderr)
	}
	if s := stderr.String(); s != "" {
		return "", fmt.Errorf("Stderr = %s\nWant empty", s)
	}
	return stdout.String(), nil
}

// template to convert a function to a full file
const template = `
package main
%s
`

type asmTest struct {
	// architecture to compile to
	arch string
	// function to compile
	function string
	// regexps that must match the generated assembly
	regexps []string
}

var asmTests = [...]asmTest{
	{"amd64", `
func f(x int) int {
	return x * 64
}
`,
		[]string{"\tSHLQ\t\\$6,"},
	},
	{"amd64", `
func f(x int) int {
	return x * 96
}`,
		[]string{"\tSHLQ\t\\$5,", "\tLEAQ\t\\(.*\\)\\(.*\\*2\\),"},
	},
}
