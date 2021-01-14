// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import (
	"bytes"
	"fmt"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

// TestLarge generates a very large file to verify that large
// program builds successfully, in particular, too-far
// conditional branches are fixed, and also verify that the
// instruction's pc can be correctly aligned even when branches
// need to be fixed.
func TestLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip in short mode")
	}
	testenv.MustHaveGoBuild(t)

	dir, err := ioutil.TempDir("", "testlarge")
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	// generate a very large function
	buf := bytes.NewBuffer(make([]byte, 0, 7000000))
	gen(buf)

	tmpfile := filepath.Join(dir, "x.s")
	err = ioutil.WriteFile(tmpfile, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("can't write output: %v\n", err)
	}

	pattern := `0x0080\s00128\s\(.*\)\tMOVD\t\$3,\sR3`

	// assemble generated file
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "asm", "-S", "-o", filepath.Join(dir, "test.o"), tmpfile)
	cmd.Env = append(os.Environ(), "GOARCH=arm64", "GOOS=linux")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Assemble failed: %v, output: %s", err, out)
	}
	matched, err := regexp.MatchString(pattern, string(out))
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("The alignment is not correct: %t, output:%s\n", matched, out)
	}

	// build generated file
	cmd = exec.Command(testenv.GoToolPath(t), "tool", "asm", "-o", filepath.Join(dir, "x.o"), tmpfile)
	cmd.Env = append(os.Environ(), "GOARCH=arm64", "GOOS=linux")
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Build failed: %v, output: %s", err, out)
	}
}

// gen generates a very large program, with a very far conditional branch.
func gen(buf *bytes.Buffer) {
	fmt.Fprintln(buf, "TEXT f(SB),0,$0-0")
	fmt.Fprintln(buf, "TBZ $5, R0, label")
	fmt.Fprintln(buf, "CBZ R0, label")
	fmt.Fprintln(buf, "BEQ label")
	fmt.Fprintln(buf, "PCALIGN $128")
	fmt.Fprintln(buf, "MOVD $3, R3")
	for i := 0; i < 1<<19; i++ {
		fmt.Fprintln(buf, "MOVD R0, R1")
	}
	fmt.Fprintln(buf, "label:")
	fmt.Fprintln(buf, "RET")
}

// Issue 20348.
func TestNoRet(t *testing.T) {
	dir, err := ioutil.TempDir("", "testnoret")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tmpfile := filepath.Join(dir, "x.s")
	if err := ioutil.WriteFile(tmpfile, []byte("TEXT ·stub(SB),$0-0\nNOP\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "asm", "-o", filepath.Join(dir, "x.o"), tmpfile)
	cmd.Env = append(os.Environ(), "GOARCH=arm64", "GOOS=linux")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("%v\n%s", err, out)
	}
}

// TestPCALIGN verifies the correctness of the PCALIGN by checking if the
// code can be aligned to the alignment value.
func TestPCALIGN(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	dir, err := ioutil.TempDir("", "testpcalign")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tmpfile := filepath.Join(dir, "test.s")
	tmpout := filepath.Join(dir, "test.o")

	code1 := []byte("TEXT ·foo(SB),$0-0\nMOVD $0, R0\nPCALIGN $8\nMOVD $1, R1\nRET\n")
	code2 := []byte("TEXT ·foo(SB),$0-0\nMOVD $0, R0\nPCALIGN $16\nMOVD $2, R2\nRET\n")
	// If the output contains this pattern, the pc-offsite of "MOVD $1, R1" is 8 bytes aligned.
	out1 := `0x0008\s00008\s\(.*\)\tMOVD\t\$1,\sR1`
	// If the output contains this pattern, the pc-offsite of "MOVD $2, R2" is 16 bytes aligned.
	out2 := `0x0010\s00016\s\(.*\)\tMOVD\t\$2,\sR2`
	var testCases = []struct {
		name string
		code []byte
		out  string
	}{
		{"8-byte alignment", code1, out1},
		{"16-byte alignment", code2, out2},
	}

	for _, test := range testCases {
		if err := ioutil.WriteFile(tmpfile, test.code, 0644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(testenv.GoToolPath(t), "tool", "asm", "-S", "-o", tmpout, tmpfile)
		cmd.Env = append(os.Environ(), "GOARCH=arm64", "GOOS=linux")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("The %s build failed: %v, output: %s", test.name, err, out)
			continue
		}

		matched, err := regexp.MatchString(test.out, string(out))
		if err != nil {
			t.Fatal(err)
		}
		if !matched {
			t.Errorf("The %s testing failed!\ninput: %s\noutput: %s\n", test.name, test.code, out)
		}
	}
}

// TestVMOVQ checks if the arm64 VMOVQ instruction is working properly.
func TestVMOVQ(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	if runtime.GOARCH != "arm64" || runtime.GOOS != "linux" {
		return
	}
	dir, err := ioutil.TempDir("", "testvmovq")
	if err != nil {
		t.Fatal(err)
	}
	//	defer os.RemoveAll(dir)

	// Create file f.s which contains VMOVQ instruction.
	tmpfile1 := filepath.Join(dir, "f.s")
	s := `TEXT ·f(SB), 4, $0-16
	VMOVQ   $0x7040201008040201, $0x3040201008040201, V1
	VMOV    V1.D[0], R0
	VMOV    V1.D[1], R1
	MOVD    R0, r1+0(FP)
	MOVD    R1, r2+8(FP)
	RET
	`
	if err := ioutil.WriteFile(tmpfile1, []byte(s), 0644); err != nil {
		t.Fatal(err)
	}

	// Create main.go which calls function f.
	tmpfile2 := filepath.Join(dir, "main.go")
	s = `package main
	func f() (r1, r2 uint64)
	func main () {
		a, b := f()
		if a != 0x7040201008040201 || b != 0x3040201008040201 {
			panic("the arm64 instruction VMOVQ is not working as expected")
		}
	}`
	if err := ioutil.WriteFile(tmpfile2, []byte(s), 0644); err != nil {
		t.Fatal(err)
	}

	// Create go.mod
	tmpfile3 := filepath.Join(dir, "go.mod")
	s = "module example.com/main\n"
	if err := ioutil.WriteFile(tmpfile3, []byte(s), 0644); err != nil {
		t.Fatal(err)
	}

	// Build the executable.
	exe := filepath.Join(dir, "main")
	cmd := exec.Command(testenv.GoToolPath(t), "build", "-o", exe)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("%v\n%s", err, out)
	}
	// Run the executable.
	cmd = exec.Command(exe)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("%v\n%s", err, out)
	}
}
