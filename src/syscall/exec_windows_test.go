// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestEscapeArg(t *testing.T) {
	var tests = []struct {
		input, output string
	}{
		{``, `""`},
		{`a`, `a`},
		{` `, `" "`},
		{`\`, `\`},
		{`"`, `\"`},
		{`\"`, `\\\"`},
		{`\\"`, `\\\\\"`},
		{`\\ `, `"\\ "`},
		{` \\`, `" \\\\"`},
		{`a `, `"a "`},
		{`C:\`, `C:\`},
		{`C:\Program Files (x32)\Common\`, `"C:\Program Files (x32)\Common\\"`},
		{`C:\Users\Игорь\`, `C:\Users\Игорь\`},
		{`Андрей\file`, `Андрей\file`},
		{`C:\Windows\temp`, `C:\Windows\temp`},
		{`c:\temp\newfile`, `c:\temp\newfile`},
		{`\\?\C:\Windows`, `\\?\C:\Windows`},
		{`\\?\`, `\\?\`},
		{`\\.\C:\Windows\`, `\\.\C:\Windows\`},
		{`\\server\share\file`, `\\server\share\file`},
		{`\\newserver\tempshare\really.txt`, `\\newserver\tempshare\really.txt`},
	}
	for _, test := range tests {
		if got := syscall.EscapeArg(test.input); got != test.output {
			t.Errorf("EscapeArg(%#q) = %#q, want %#q", test.input, got, test.output)
		}
	}
}

func testStartProcessBatchFile(t *testing.T, name string, argsInCmdLine bool, args ...string) {
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	defer pw.Close()
	attr := &os.ProcAttr{Files: []*os.File{nil, pw, pw}}
	var argv []string
	if argsInCmdLine {
		attr.Sys = &syscall.SysProcAttr{CmdLine: `"` + name + `"`}
		for _, v := range args {
			attr.Sys.CmdLine += " " + syscall.EscapeArg(v)
		}
	} else {
		argv = append([]string{name}, args...)
	}
	p, err := os.StartProcess(name, argv, attr)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Wait()
	if err != nil {
		t.Fatal(err)
	}
	pw.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, pr)
	if err != nil {
		t.Fatal(err)
	}
	var want string
	for _, v := range args {
		want += syscall.EscapeArg(v) + " "
	}
	if len(want) > 0 {
		want = want[:len(want)-1] // remove trailing space
	} else {
		want = "ECHO is on."
	}
	got := string(bytes.TrimSpace(buf.Bytes()))
	if got != want {
		t.Fatalf("StartProcess(%#q, %#q) = %#q, want %#q", name, args, got, want)
	}
}

func TestStartProcessWithSpaceInPath(t *testing.T) {
	dir := t.TempDir()
	cmd := path.Join(dir, "foo bar.cmd")
	err := os.WriteFile(cmd, []byte("@echo %*"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"args with no spaces", []string{"ab"}},
		{"args with spaces", []string{"a b"}},
		{"args with and without spaces", []string{"ab", "a b"}},
	}
	for _, test := range tests {
		t.Run(test.name+"/argv", func(t *testing.T) {
			testStartProcessBatchFile(t, cmd, false, test.args...)
		})
		t.Run(test.name+"/CmdLine", func(t *testing.T) {
			testStartProcessBatchFile(t, cmd, true, test.args...)
		})
	}
}

func TestChangingProcessParent(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "parent" {
		// in parent process

		// Parent does nothing. It is just used as a parent of a child process.
		time.Sleep(time.Minute)
		os.Exit(0)
	}

	if os.Getenv("GO_WANT_HELPER_PROCESS") == "child" {
		// in child process
		dumpPath := os.Getenv("GO_WANT_HELPER_PROCESS_FILE")
		if dumpPath == "" {
			fmt.Fprintf(os.Stderr, "Dump file path cannot be blank.")
			os.Exit(1)
		}
		err := os.WriteFile(dumpPath, []byte(fmt.Sprintf("%d", os.Getppid())), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing dump file: %v", err)
			os.Exit(2)
		}
		os.Exit(0)
	}

	// run parent process

	parent := exec.Command(os.Args[0], "-test.run=^TestChangingProcessParent$")
	parent.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=parent")
	err := parent.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		parent.Process.Kill()
		parent.Wait()
	}()

	// run child process

	const _PROCESS_CREATE_PROCESS = 0x0080
	const _PROCESS_DUP_HANDLE = 0x0040
	childDumpPath := filepath.Join(t.TempDir(), "ppid.txt")
	ph, err := syscall.OpenProcess(_PROCESS_CREATE_PROCESS|_PROCESS_DUP_HANDLE|syscall.PROCESS_QUERY_INFORMATION,
		false, uint32(parent.Process.Pid))
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.CloseHandle(ph)

	child := exec.Command(os.Args[0], "-test.run=^TestChangingProcessParent$")
	child.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=child",
		"GO_WANT_HELPER_PROCESS_FILE="+childDumpPath)
	child.SysProcAttr = &syscall.SysProcAttr{ParentProcess: ph}
	childOutput, err := child.CombinedOutput()
	if err != nil {
		t.Errorf("child failed: %v: %v", err, string(childOutput))
	}
	childOutput, err = os.ReadFile(childDumpPath)
	if err != nil {
		t.Fatalf("reading child output failed: %v", err)
	}
	if got, want := string(childOutput), fmt.Sprintf("%d", parent.Process.Pid); got != want {
		t.Fatalf("child output: want %q, got %q", want, got)
	}
}
