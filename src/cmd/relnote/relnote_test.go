// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"internal/testenv"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIFragments(t *testing.T) {
	root := testenv.GOROOT(t)
	nextFiles, err := filepath.Glob(filepath.Join(root, "api", "next", "*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	// Check that each api/next file has a corresponding release note fragment.
	for _, apiFile := range nextFiles {
		relnoteFile := filepath.Join(root, "doc", "next", filepath.Base(strings.TrimSuffix(apiFile, ".txt"))+".md")
		data, err := os.ReadFile(relnoteFile)
		if errors.Is(err, fs.ErrNotExist) {
			t.Errorf("API file %s needs a corresponding valid release-note file %s", apiFile, relnoteFile)
		} else if err != nil {
			t.Fatal(err)
		} else if err := checkValid(data); err != nil {
			t.Errorf("%s: %v", relnoteFile, err)
		}
	}
}

func checkValid(contents []byte) error {
	// Require that the file contains at least "TODO" or a sentence-ending punctuation mark.
	// TODO(jba): improve these checks.
	if !bytes.Contains(contents, []byte("TODO")) && !bytes.ContainsAny(contents, ".?!") {
		return errors.New("must contain 'TODO' or a sentence")
	}
	return nil
}
