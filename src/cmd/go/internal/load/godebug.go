// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package load

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"cmd/go/internal/modload"
)

var ErrNotGoDebug = errors.New("not //go:debug line")

func ParseGoDebug(text string) (key, value string, err error) {
	if !strings.HasPrefix(text, "//go:debug") {
		return "", "", ErrNotGoDebug
	}
	i := strings.IndexAny(text, " \t")
	if i < 0 {
		if strings.TrimSpace(text) == "//go:debug" {
			return "", "", fmt.Errorf("missing key=value")
		}
		return "", "", ErrNotGoDebug
	}
	k, v, ok := strings.Cut(strings.TrimSpace(text[i:]), "=")
	if !ok {
		return "", "", fmt.Errorf("missing key=value")
	}
	if strings.Contains(k, " \t") {
		return "", "", fmt.Errorf("key contains space")
	}
	if strings.Contains(k, " \t") {
		return "", "", fmt.Errorf("value contains space")
	}
	if strings.Contains(k, ",") {
		return "", "", fmt.Errorf("key contains comma")
	}
	if strings.Contains(k, ",") {
		return "", "", fmt.Errorf("value contains comma")
	}
	return k, v, nil
}

// defaultGODEBUG returns the default GODEBUG setting for the main package p.
func defaultGODEBUG(p *Package) string {
	if p.Name != "main" {
		return ""
	}
	goVersion := modload.MainModules.GoVersion()
	if modload.RootMode == modload.NoRoot && p.Module != nil {
		// This is go install pkg@version or go run pkg@version.
		// Use the Go version from the package.
		// If there isn't one, then
		goVersion = p.Module.GoVersion
		if goVersion == "" {
			goVersion = "1.20"
		}
	}

	def := godebugForGoVersion(goVersion)
	var m map[string]string
	for _, d := range p.Internal.Directives {
		k, v, err := ParseGoDebug(d.Text)
		if err != nil {
			continue
		}
		if m == nil {
			m = make(map[string]string)
			for k, v := range def {
				m[k] = v
			}
		}
		m[k] = v
	}
	if m == nil {
		m = def
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		if b.Len() > 0 {
			b.WriteString(",")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(m[k])
	}
	return b.String()
}

func godebugForGoVersion(v string) map[string]string {
	if strings.Count(v, ".") >= 2 {
		i := strings.Index(v, ".")
		j := i + 1 + strings.Index(v[i+1:], ".")
		v = v[:j]
	}

	if !strings.HasPrefix(v, "1.") {
		return nil
	}
	n, err := strconv.Atoi(v[len("1."):])
	if err != nil {
		return nil
	}
	n = 1_00 + n

	def := make(map[string]string)
	for _, d := range defaultGodebugs {
		if (d.before != 0 && n < d.before) || (d.first != 0 && n >= d.first) {
			def[d.name] = d.value
		}
	}
	return def
}

var defaultGodebugs = []struct {
	before int // applies to Go versions up until this one
	first  int // applies to Go versions starting at this one
	name   string
	value  string
}{
	{before: 1_21, name: "panicnil", value: "1"},
}
