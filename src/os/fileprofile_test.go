package os_test

import (
	"bytes"
	"os"
	"runtime/pprof"
	"testing"
)

func TestFileProfile(t *testing.T) {
	prof := pprof.Lookup("os.File")
	if prof == nil {
		t.Fatalf("no os.File profile")
	}
	profString := func() string {
		var buf bytes.Buffer
		prof.WriteTo(&buf, 1)
		return buf.String()
	}
	baseCount := prof.Count()
	var files []*os.File
	for i := 0; i < 10; i++ {
		files = append(files, newFile("TestFileProfile", t))
		if want, got := baseCount+i+1, prof.Count(); want != got {
			t.Fatalf("want %d profiles, got %d\n%s", want, got, profString())
		}
	}
	for i, f := range files {
		f.Close()
		if want, got := baseCount+len(files)-i-1, prof.Count(); want != got {
			t.Errorf("want %d profiles, got %d\n%s", want, got, profString())
		}
		os.Remove(f.Name())
	}
}
