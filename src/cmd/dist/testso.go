// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (t *tester) buildTestSO(dir string) error {
	sosrc := `
// +build ignore

#ifdef WIN32
// A Windows DLL is unable to call an arbitrary function in
// the main executable. Work around that by making the main
// executable pass the callback function pointer to us.
void (*goCallback)(void);
__declspec(dllexport) void setCallback(void *f)
{
	goCallback = (void (*)())f;
}
__declspec(dllexport) void sofunc(void);
#else
extern void goCallback(void);
void setCallback(void *f) { (void)f; }
#endif

// OpenBSD and older Darwin lack TLS support
#if !defined(__OpenBSD__) && !defined(__APPLE__)
__thread int tlsvar = 12345;
#endif

void sofunc(void)
{
	goCallback();
}
`
	output, err := exec.Command("go", "env", "CC").Output()
	if err != nil {
		return fmt.Errorf("Error running go env CC: %v", err)
	}
	cc := strings.TrimSuffix(string(output), "\n")
	if cc == "" {
		return errors.New("CC environment variable (go env CC) cannot be empty")
	}
	output, err = exec.Command("go", "env", "GOGCCFLAGS").Output()
	if err != nil {
		return fmt.Errorf("Error running go env GOGCCFLAGS: %v", err)
	}
	gogccflags := strings.Split(strings.TrimSuffix(string(output), "\n"), " ")

	err = ioutil.WriteFile(filepath.Join(dir, "cgoso_c.c"), []byte(sosrc), 0644)
	if err != nil {
		return err
	}

	ext := "so"
	args := append(gogccflags, "-shared")
	switch t.goos {
	case "darwin":
		ext = "dylib"
		args = append(args, "-undefined", "suppress", "-flat_namespace")
	case "windows":
		ext = "dll"
	}
	args = append(args, "-o", "libcgosotest."+ext, "cgoso_c.c")

	cmd := exec.Command(cc, args...)
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error building shared object: %v\n%v", err, string(output))
	}
	return nil
}

func (t *tester) buildTestSOGo(dir string) error {
	mainsrc := `
// +build ignore

package main

import "."

func main() {
	cgosotest.Test()
}
`
	err := ioutil.WriteFile(filepath.Join(dir, "main.go"), []byte(mainsrc), 0644)
	if err != nil {
		return err
	}

	cgososrc := `
package cgosotest

/*
// intentionally write the same LDFLAGS differently
// to test correct handling of LDFLAGS.
#cgo linux LDFLAGS: -L. -lcgosotest
#cgo dragonfly LDFLAGS: -L. -l cgosotest
#cgo freebsd LDFLAGS: -L. -l cgosotest
#cgo openbsd LDFLAGS: -L. -l cgosotest
#cgo solaris LDFLAGS: -L. -lcgosotest
#cgo netbsd LDFLAGS: -L. libcgosotest.so
#cgo darwin LDFLAGS: -L. libcgosotest.dylib
#cgo windows LDFLAGS: -L. libcgosotest.dll

void init(void);
void sofunc(void);
*/
import "C"

func Test() {
	C.init()
	C.sofunc()
}

//export goCallback
func goCallback() {
}
`
	cgosounixsrc := `
// +build dragonfly freebsd linux netbsd solaris

package cgosotest

/*
extern int __thread tlsvar;
int *getTLS() { return &tlsvar; }
*/
import "C"

func init() {
	if v := *C.getTLS(); v != 12345 {
		println("got", v)
		panic("BAD TLS value")
	}
}
`
	cgosocsrc := `
#include "_cgo_export.h"

#ifdef WIN32
extern void setCallback(void *);
void init() {
	setCallback(goCallback);
}
#else
void init() {}
#endif
`
	err = ioutil.WriteFile(filepath.Join(dir, "cgoso.go"), []byte(cgososrc), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "cgoso_unix.go"), []byte(cgosounixsrc), 0644)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "cgoso.c"), []byte(cgosocsrc), 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", "main.exe", "main.go")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error building go executable: %v\n%v", err, string(output))
	}
	return nil
}

func (t *tester) runTestSO(dir string) error {
	ldlibvar := "LD_LIBRARY_PATH"
	if t.goos == "darwin" {
		ldlibvar = "DYLD_LIBRARY_PATH"
	}
	cmd := exec.Command("./main.exe")
	cmd.Dir = dir
	if t.goos != "windows" {
		cmd.Env = mergeEnvLists([]string{ldlibvar + "=."}, os.Environ())
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error running test: %v\n%v", err, string(output))
	}
	return nil
}

func (t *tester) cgoTestSO() error {
	if t.goos == "android" || t.iOS() {
		// No exec facility on Android or iOS.
		return nil
	}
	if t.goos == "ppc64le" || t.goos == "ppc64" {
		// External linking not implemented on ppc64.
		fmt.Println("skipping test on ppc64 (issue #8912)")
		return nil
	}

	tmpdir, err := ioutil.TempDir("", "testso")
	if err != nil {
		return fmt.Errorf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	err = t.buildTestSO(tmpdir)
	if err != nil {
		return err
	}
	err = t.buildTestSOGo(tmpdir)
	if err != nil {
		return err
	}
	return t.runTestSO(tmpdir)
}
