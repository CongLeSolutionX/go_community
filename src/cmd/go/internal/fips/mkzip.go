// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/module"
	modzip "golang.org/x/mod/zip"
)

var flagBranch = flag.String("b", "master", "origin branch to use")

func usage() {
	fmt.Fprintf(os.Stderr, "usage: go run mkzip.go [-b branch] vX.Y\n")
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("mkzip: ")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		usage()
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if !strings.HasSuffix(filepath.ToSlash(wd), "lib/fips140") {
		log.Fatalf("must be run in lib/fips140 directory")
	}

	version := flag.Arg(0)
	if !regexp.MustCompile(`^v\d+\.\d+\.\d+$`).MatchString(version) {
		log.Fatalf("invalid version %q; must be vX.Y.Z", version)
	}
	if _, err := os.Stat(version + ".zip"); err == nil {
		log.Fatalf("%s.zip already exists", version)
	}

	// Make module zip file in memory.
	goroot := "../.."
	var zbuf bytes.Buffer
	err = modzip.CreateFromVCS(&zbuf,
		module.Version{Path: "golang.org/fips140", Version: version},
		goroot, *flagBranch, "src/crypto/internal/fips")
	if err != nil {
		log.Fatal(err)
	}

	// Write new zip file with longer paths: fips/v1.2.3/foo.go instead of foo.go.
	// That way we can bind the fips directory onto the
	// GOROOT/src/crypto/internal/fips directory.
	zr, err := zip.NewReader(bytes.NewReader(zbuf.Bytes()), int64(zbuf.Len()))
	if err != nil {
		log.Fatal(err)
	}

	var zbuf2 bytes.Buffer
	zw := zip.NewWriter(&zbuf2)
	for _, f := range zr.File {
		// golang.org/fips140@v1.2.3/dir/file.go ->
		// golang.org/fips140@v1.2.3/fips140/v1.2.3/dir/file.go
		if f.Name != "golang.org/fips140@"+version+"/LICENSE" {
			f.Name = "golang.org/fips140@" + version + "/fips140/" + version +
				strings.TrimPrefix(f.Name, "golang.org/fips140@"+version)
		}
		wf, err := zw.CreateRaw(&f.FileHeader)
		if err != nil {
			log.Fatal(err)
		}
		rf, err := f.OpenRaw()
		if err != nil {
			log.Fatal(err)
		}
		if _, err := io.Copy(wf, rf); err != nil {
			log.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(version+".zip", zbuf2.Bytes(), 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("wrote %s.zip", version)
}
