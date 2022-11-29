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

func defaultGODEBUG(p *Package) string {
	def := godebugForGoVersion(p)
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

func godebugForGoVersion(p *Package) map[string]string {
	if p.Standard { // std and cmd are missing p.Module
		return nil
	}
	v := "go1.16"
	if p.Module != nil && p.Module.GoVersion != "" {
		v = p.Module.GoVersion
	}

	if strings.Count(v, ".") >= 2 {
		i := strings.Index(v, ".")
		j := i + 1 + strings.Index(v[i+1:], ".")
		v = v[:j]
	}

	if !strings.HasPrefix(v, "go1.") {
		return nil
	}
	n, err := strconv.Atoi(v[len("go1."):])
	if err != nil {
		return nil
	}

	switch {
	case n <= 19:
		return godebugForGo1_19

	case n <= 20:
		return godebugForGo1_20

	case n == 21:
		return godebugForGo1_21

	case n >= 22:
		return godebugForGo1_22
	}
	panic("unreachable")
}

// Note: When a new entry is added to godebugForGo1_N,
// it should also be added to godebugForGo1_M for all M < N.

var godebugForGo1_19 = map[string]string{
	"randautoseed": "0",
}

var godebugForGo1_20 = map[string]string{}

var godebugForGo1_21 = map[string]string{}

var godebugForGo1_22 = map[string]string{}
