// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"internal/testenv"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/build/relnote"
	"golang.org/x/net/html"
	"rsc.io/markdown"
)

var flagCheck = flag.Bool("check", false, "run API release note checks")

// Check that each file in api/next has corresponding release note files in doc/next.
func TestCheckAPIFragments(t *testing.T) {
	if !*flagCheck {
		t.Skip("-check not specified")
	}
	root := testenv.GOROOT(t)
	rootFS := os.DirFS(root)
	files, err := fs.Glob(rootFS, "api/next/*.txt")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("checking release notes for %d files in api/next", len(files))
	docFS := os.DirFS(filepath.Join(root, "doc", "next"))
	// Check that each api/next file has a corresponding release note fragment.
	for _, apiFile := range files {
		if err := relnote.CheckAPIFile(rootFS, apiFile, docFS, "doc/next"); err != nil {
			t.Errorf("%s: %v", apiFile, err)
		}
	}
}

// TestLinks is a subset of TestAll in x/website/cmd/golangorg
// that runs in the main Go repo, and tries to report problems
// with links in release note fragments in doc/next.
func TestLinks(t *testing.T) {
	if !*flagCheck {
		t.Skip("-check not specified")
	}
	root := testenv.GOROOT(t)
	docFS := os.DirFS(filepath.Join(root, "doc", "next"))
	if _, err := fs.Stat(docFS, "."); errors.Is(err, fs.ErrNotExist) {
		t.Log("no next release note fragments")
		return
	}
	docMD, err := relnote.Merge(docFS)
	if err != nil {
		t.Fatalf("relnote.Merge: %v", err)
	}
	docHTML := markdown.ToHTML(docMD)

	// Check that page is valid HTML.
	// First check for over- or under-escaped HTML.
	bad := findBad(docHTML)
	if bad != "" {
		t.Fatalf("doc/next: contains improperly escaped HTML\n%s", bad)
	}

	// Now check all the links to other pages on this server.
	// (Pages on other servers are too expensive to check
	// and would cause test failures if servers went down
	// or moved their contents.)
	doc, err := html.Parse(strings.NewReader(docHTML))
	if err != nil {
		t.Fatalf("doc/next: parsing HTML: %v", err)
	}

	// Walk HTML looking for <a href=...>, <img src=...>, and <script src=...>.
	var checkLinks func(*html.Node)
	checkLinks = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			checkLinks(c)
		}
		var targ string
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a":
				targ = findAttr(n, "href")
			case "img", "script":
				targ = findAttr(n, "src")
			}
		}
		// Ignore no target or #fragment.
		if targ == "" || strings.HasPrefix(targ, "#") {
			return
		}

		// Parse target as URL.
		u, err := url.Parse(targ)
		if err != nil {
			t.Errorf("doc/next: found unparseable URL %s: %v", targ, err)
			return
		}

		// Check whether URL is canonicalized properly.
		if fix := fixURL(u); fix != "" {
			t.Errorf("doc/next: found link to %s, should be %s", targ, fix)
		}
	}
	checkLinks(doc)
}

// fixURL returns the corrected URL for u,
// or the empty string if u is fine.
func fixURL(u *url.URL) string {
	switch u.Host {
	case "golang.org":
		if strings.HasPrefix(u.Path, "/x/") {
			return ""
		}
		fallthrough
	case "go.dev":
		u.Host = ""
		u.Scheme = ""
		if u.Path == "" {
			u.Path = "/"
		}
		return u.String()
	case "blog.golang.org",
		"blog.go.dev",
		"learn.golang.org",
		"learn.go.dev",
		"play.golang.org",
		"play.go.dev",
		"tour.golang.org",
		"tour.go.dev",
		"talks.golang.org",
		"talks.go.dev":
		name, _, _ := strings.Cut(u.Host, ".")
		u.Host = ""
		u.Scheme = ""
		u.Path = "/" + name + u.Path
		return u.String()
	case "github.com":
		if strings.HasPrefix(u.Path, "/golang/go/issues/") {
			u.Host = "go.dev"
			u.Path = "/issue/" + strings.TrimPrefix(u.Path, "/golang/go/issues/")
			return u.String()
		}
		if strings.HasPrefix(u.Path, "/golang/go/wiki/") {
			u.Host = "go.dev"
			u.Path = "/wiki/" + strings.TrimPrefix(u.Path, "/golang/go/wiki/")
			return u.String()
		}
	}
	return ""
}

// findAttr returns the value for n's attribute with the given name.
func findAttr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

// findBad returns (only) the lines containing badly escaped HTML in body.
// If findBad returns the empty string, there is no badly escaped HTML.
func findBad(body string) string {
	lines := strings.SplitAfter(body, "\n")
	var out []string
	for _, line := range lines {
		for _, b := range bads {
			if strings.Contains(line, b) {
				out = append(out, line)
				break
			}
		}
	}
	return strings.Join(out, "")
}

var bads = []string{
	"&amp;lt;",
	"&amp;gt;",
	"&amp;amp;",
	" < ",
	"<-",
	"& ",
}
