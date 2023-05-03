// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"bufio"
	"fmt"
	"internal/testenv"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// testPGOSpecialize tests that specific specializations are performed.
func testPGOSpecialize(t *testing.T, dir string) {
	testenv.MustHaveGoRun(t)
	t.Parallel()

	const pkg = "example.com/pgo/specialize"

	// Add a go.mod so we have a consistent symbol names in this temp dir.
	goMod := fmt.Sprintf(`module %s
go 1.19
`, pkg)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("error writing go.mod: %v", err)
	}

	// Build the test with the profile.
	pprof := filepath.Join(dir, "shape.pprof")
	gcflag := fmt.Sprintf("-gcflags=-m -m -pgoprofile=%s -pgospecialize", pprof)
	out := filepath.Join(dir, "test.exe")
	cmd := testenv.CleanCmdEnv(testenv.Command(t, testenv.GoToolPath(t), "test", "-c", "-o", out, gcflag, "."))
	cmd.Dir = dir

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("error creating pipe: %v", err)
	}
	defer pr.Close()
	cmd.Stdout = pw
	cmd.Stderr = pw

	err = cmd.Start()
	pw.Close()
	if err != nil {
		t.Fatalf("error starting go test: %v", err)
	}

	scanner := bufio.NewScanner(pr)
	curPkg := ""

	origAST := "sumA += i2.Area()"
	specializedASTCond := ":= i2.(Circle);"
	specializedASTBody := "sumA += Circle.Area(i2.(Circle))"
	notSpecializedASTPerim := "i1.Perimeter()"
	notSpecializedASTArea := "i2.Area()"

	specializedLine := regexp.MustCompile(`: specializing (.*): (.*)`)
	notSpecializedLine := regexp.MustCompile(`: cannot specialize hot interface method call (.*)`)

	var origASTFound string
	var specializedASTFound string

	notSpecialized := make(map[string]struct{})

	for scanner.Scan() {
		line := scanner.Text()
		t.Logf("child: %s", line)
		if strings.HasPrefix(line, "# ") {
			curPkg = line[2:]
			splits := strings.Split(curPkg, " ")
			curPkg = splits[0]
			continue
		}
		if m := specializedLine.FindStringSubmatch(line); m != nil {
			origASTFound, specializedASTFound = m[1], m[2]
			continue
		}
		if m := notSpecializedLine.FindStringSubmatch(line); m != nil {
			notSpecialized[m[1]] = struct{}{}
			continue
		}
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("error running go test: %v", err)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading go test output: %v", err)
	}
	if origASTFound == origAST {
		if !strings.Contains(specializedASTFound, specializedASTCond) || !strings.Contains(specializedASTFound, specializedASTBody) {
			t.Errorf("%s was not specialized", origAST)
		}
	} else {
		t.Errorf("%s was not specialized", origAST)
	}
	if _, ok := notSpecialized[notSpecializedASTPerim]; !ok {
		t.Errorf("%s should not be specialized", notSpecializedASTPerim)
	}
	if _, ok := notSpecialized[notSpecializedASTArea]; !ok {
		t.Errorf("%s should not be specialized", notSpecializedASTArea)
	}
}
