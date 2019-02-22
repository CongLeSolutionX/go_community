// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo,!ios,!android

package so_test

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func requireTestSOSupported(t *testing.T) {
	t.Helper()
	switch runtime.GOARCH {
	case "ppc64":
		t.Skip("External linking not implemented on ppc64 (issue #8912).")
	case "mips64le", "mips64":
		t.Skip("External linking not implemented on mips64.")
	}
}

func TestSO(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	GOPATH, err := ioutil.TempDir("", "cgosotest")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(GOPATH)
	os.Setenv("GOPATH", GOPATH)

	modRoot := filepath.Join(GOPATH, "src", "cgosotest")
	if err := os.MkdirAll(modRoot, 0777); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(modRoot, "go.mod"), []byte("module cgosotest\n"), 0666); err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil || path == "testdata" {
			return err
		}

		dstPath := modRoot + strings.TrimPrefix(path, "testdata")
		if info.IsDir() {
			return os.Mkdir(dstPath, 0777)
		}

		data, err := ioutil.ReadFile(filepath.Join(cwd, path))
		if err != nil {
			return err
		}
		return ioutil.WriteFile(dstPath, data, info.Mode())
	})
	if err != nil {
		log.Fatal(err)
	}

	goEnv := map[string]string{}
	cmd := exec.Command("go", "env")
	cmd.Dir = modRoot
	cmd.Stderr = new(strings.Builder)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Error running go env: %v\n%s", err, cmd.Stderr)
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		frags := strings.SplitN(line, "=", 2)
		if len(frags) != 2 {
			continue
		}
		k := frags[0]
		v, err := strconv.Unquote(frags[1])
		if err != nil {
			t.Fatalf("Malformed output from 'go env': %s\n%s", err, line)
		}
		goEnv[k] = v
	}

	cc := goEnv["CC"]
	if cc == "" {
		t.Fatal("CC environment variable (go env CC) cannot be empty")
	}
	gogccflags := strings.Split(goEnv["GOGCCFLAGS"], " ")
	goos := goEnv["GOOS"]
	goarch := goEnv["GOARCH"]

	// build shared object
	ext := "so"
	args := append(gogccflags, "-shared")
	switch goos {
	case "darwin":
		ext = "dylib"
		args = append(args, "-undefined", "suppress", "-flat_namespace")
	case "windows":
		ext = "dll"
		args = append(args, "-DEXPORT_DLL")
	}
	sofname := "libcgosotest." + ext
	args = append(args, "-o", sofname, "cgoso_c.c")

	cmd = exec.Command(cc, args...)
	cmd.Dir = modRoot
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s\n%s", strings.Join(cmd.Args, " "), err, out)
	}
	t.Logf("%s:\n%s", strings.Join(cmd.Args, " "), out)

	cmd = exec.Command("go", "build", "-o", "main.exe", "main.go")
	cmd.Dir = modRoot
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s\n%s", strings.Join(cmd.Args, " "), err, out)
	}
	t.Logf("%s:\n%s", strings.Join(cmd.Args, " "), out)

	cmd = exec.Command("./main.exe")
	cmd.Dir = modRoot
	if goos != "windows" {
		s := "LD_LIBRARY_PATH"
		if goos == "darwin" {
			s = "DYLD_LIBRARY_PATH"
		}
		cmd.Env = append(os.Environ(), s+"=.")

		// On FreeBSD 64-bit architectures, the 32-bit linker looks for
		// different environment variables.
		if goos == "freebsd" && goarch == "386" {
			cmd.Env = append(cmd.Env, "LD_32_LIBRARY_PATH=.")
		}
	}
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %s\n%s", strings.Join(cmd.Args, " "), err, out)
	}
	t.Logf("%s:\n%s", strings.Join(cmd.Args, " "), out)
}
