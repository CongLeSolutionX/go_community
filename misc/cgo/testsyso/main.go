// This test tests that we can link in-package syso files that provides symbols
// for cgo. See issue 29253.
package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var files = map[string]string{
	"ext.c": `
// +build ignore

int f() { return 42; }
`,
	"pkg/pkg.go": `
package pkg

// extern int f(void);
import "C"

func init() {
	if v := C.f(); v != 42 {
		panic(v)
	}
}
`,
	"main.go": `
package main

import _ "./pkg"

func main() {}
`}

func runGoTool(dir string, args ...string) (string, error) {
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(data), "\n"), nil
}

func main() {
	cc, err := runGoTool("", "env", "CC")
	if err != nil {
		log.Fatal(err)
	}
	cflags, err := runGoTool("", "env", "GOGCCFLAGS")
	if err != nil {
		log.Fatal(err)
	}
	dir, err := ioutil.TempDir("", "testsyso")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	err = os.Mkdir(filepath.Join(dir, "pkg"), 0700)
	if err != nil {
		log.Fatal(err)
	}
	for name, data := range files {
		err := ioutil.WriteFile(filepath.Join(dir, name), []byte(data), 0600)
		if err != nil {
			log.Fatal(err)
		}
	}
	args := append(strings.Fields(cflags), "-o", "pkg/o.syso", "-c", "ext.c")
	cmd := exec.Command(cc, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to compile ext.syso: %v", err)
	}
	_, err = runGoTool(dir, "run", "main.go")
	if err != nil {
		log.Fatalf("failed to run test: %v", err)
	}
}
