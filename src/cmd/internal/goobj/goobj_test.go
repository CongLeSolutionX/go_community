package goobj

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGoobj(t *testing.T) {
	matches, err := filepath.Glob("testdata/*-goobj")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range matches {
		tmp := strings.Split(filepath.Base(path), "-")
		if len(tmp) != 3 {
			t.Fatalf("unexpected filepath %s", path)
		}
		goarch := tmp[1]
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		p, err := Parse(f, "mypkg")
		if err != nil {
			t.Error(err)
		}
		if p.Arch != goarch {
			t.Errorf("%s: got %v, want %v", path, goarch, p.Arch)
		}
		var found bool
		for _, s := range p.Syms {
			if s.Name == "mypkg.go1" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(`%s: symbol "mypkg.go1" not found`, path)
		}
	}
}

func TestParseArchive(t *testing.T) {
	matches, err := filepath.Glob("testdata/*-archive")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range matches {
		tmp := strings.Split(filepath.Base(path), "-")
		if len(tmp) != 3 {
			t.Fatalf("unexpected filepath %s", path)
		}
		goarch := tmp[1]
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		p, err := Parse(f, "mypkg")
		if err != nil {
			t.Error(err)
		}
		if p.Arch != goarch {
			t.Errorf("%s: got %v, want %v", path, goarch, p.Arch)
		}
		var found1 bool
		var found2 bool
		for _, s := range p.Syms {
			if s.Name == "mypkg.go1" {
				found1 = true
			}
			if s.Name == "mypkg.go2" {
				found2 = true
			}
		}
		if !found1 {
			t.Errorf(`%s: symbol "mypkg.go1" not found`, path)
		}
		if !found2 {
			t.Errorf(`%s: symbol "mypkg.go2" not found`, path)
		}
	}
}
