package runtime_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func checkGdbPython(t *testing.T) {
	cmd := exec.Command("gdb", "-nx", "-q", "--batch", "-ex", "python import sys; print('golang gdb python support')")
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Skipf("skipping due to issue running gdb%v", err)
	}
	if string(out) != "golang gdb python support\n" {
		t.Skipf("skipping due to lack of python gdb support: %s", out)
	}
}

const helloSource = `
package main
import "fmt"
func main() {
	fmt.Println("hi")
}
`

func TestGdbLoadRuntimeSupport(t *testing.T) {

	checkGdbPython(t)

	dir, err := ioutil.TempDir("", "go-build")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "main.go")
	f, err := os.Create(src)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_, err = f.WriteString(helloSource)
	if err != nil {
		f.Close()
		t.Fatalf("failed to write file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", "a.exe")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building source %v\n%s", err, out)
	}

	got, _ := exec.Command("gdb", "-nx", "-q", "--batch", "-ex", "source runtime-gdb.py",
		filepath.Join(dir, "a.exe")).CombinedOutput()
	if string(got) != "Loading Go Runtime support.\n" {
		t.Fatalf("%s", got)
	}
}
