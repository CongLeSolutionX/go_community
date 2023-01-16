package x86_test

import (
	"internal/testenv"
	"os"
	"path/filepath"
	"testing"
)

func asmSehOutput(t *testing.T, s string) []byte {
	tmpdir, err := os.MkdirTemp("", "sehtest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	tmpfile, err := os.Create(filepath.Join(tmpdir, "input.s"))
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()
	_, err = tmpfile.WriteString(s)
	if err != nil {
		t.Fatal(err)
	}
	cmd := testenv.Command(t,
		testenv.GoToolPath(t), "tool", "asm", "-S", "-dynlink",
		"-o", filepath.Join(tmpdir, "output.6"), tmpfile.Name())

	cmd.Env = append(os.Environ(),
		"GOARCH=amd64", "GOOS=linux", "GOPATH="+filepath.Join(tmpdir, "_gopath"))
	asmout, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("error %s output %s", err, asmout)
	}
	return asmout
}

func TestSeh(t *testing.T) {

}
