// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// matcher sanitizes, uniques, and filters names of subtests and subbenchmarks.
type matcher struct {
	filter    []string
	matchFunc func(pat, str string) (bool, error)

	mu       sync.Mutex
	subNames map[string]int64
}

// TODO: fix test_main to avoid race and improve caching, also allowing to
// eliminate this Mutex.
var matchMutex sync.Mutex

func newMatcher(matchString func(pat, str string) (bool, error), patterns, name string) *matcher {
	var filter []string
	if patterns != "" {
		filter = splitRegexp(patterns)
		for i, s := range filter {
			filter[i] = rewrite(s)
		}
		// Verify filters before doing any processing.
		for i, s := range filter {
			if _, err := matchString(s, "non-empty"); err != nil {
				fmt.Fprintf(os.Stderr, "testing: invalid regexp for element %d of %s (%q): %s\n", i, name, s, err)
				os.Exit(1)
			}
		}
	}
	return &matcher{
		filter:    filter,
		matchFunc: matchString,
		subNames:  map[string]int64{},
	}
}

func (m *matcher) fullName(c *common, subname string) (name string, ok, partial bool) {
	name = subname

	m.mu.Lock()
	defer m.mu.Unlock()

	if c != nil && c.level > 0 {
		name = m.unique(c.name, rewrite(subname))
	}

	matchMutex.Lock()
	defer matchMutex.Unlock()

	// We check the full array of paths each time to allow for the case that
	// a pattern contains a '/'.
	elem := strings.Split(name, "/")
	for i, s := range elem {
		if i >= len(m.filter) {
			break
		}
		if ok, _ := m.matchFunc(m.filter[i], s); !ok {
			return name, false, false
		}
	}
	return name, true, len(elem) < len(m.filter)
}

func splitRegexp(s string) []string {
	a := make([]string, 0, strings.Count(s, "/"))
	cs := 0
	cp := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '[':
			cs++
		case ']':
			if cs--; cs < 0 { // An unmatched ']' is legal.
				cs = 0
			}
		case '(':
			if cs == 0 {
				cp++
			}
		case ')':
			if cs == 0 {
				cp--
			}
		case '\\':
			i++
		case '/':
			if cs == 0 && cp == 0 {
				a = append(a, s[:i])
				s = s[i+1:]
				i = 0
				continue
			}
		}
		i++
	}
	return append(a, s)
}

// unique creates a unique name for the given parent and subname by affixing it
// with one or more counts, if necessary.
func (m *matcher) unique(parent, subname string) string {
	name := fmt.Sprintf("%s/%s", parent, subname)

	// Check whether the name looks like an auto-generated name.
	// If so, check that it does not overlap with any preexisting names.
	var subMatch bool
	var suffixLen int
	for !subMatch {
		suffix, curr := stripNumberSuffix(name[:len(name)-suffixLen])
		if suffix == "" {
			break
		}
		suffixLen += len(suffix)
		next := m.subNames[name[:len(name)-suffixLen]]
		subMatch = curr < next
	}

	// Return the user-specified name verbatim if unique.
	// The subname cannot be empty.
	next, exists := m.subNames[name]
	if subname != "" && !exists && !subMatch {
		m.subNames[name] = 1 // next count is 1
		return name
	}
	if next == 0 && subname != "" {
		next++ // zero-index only permitted for an empty name
	}

	// Name was already used.
	// We increment with the count and append a string with the count.
	m.subNames[name] = next + 1
	name = fmt.Sprintf("%s#%02d", name, next)
	if _, exists := m.subNames[name]; !exists {
		return name
	}

	// The auto-generated name was alread used.
	// Recursively call matcher.unique if the new number suffix.
	return m.unique(parent, name[len(parent)+len("/"):])
}

// stripNumberSuffix strips the number suffix if it has one,
// returning the suffix and the number in the suffix.
// The number suffix must be of the format "#%02d".
func stripNumberSuffix(name string) (string, int64) {
	// Count number of trailing digits.
	var numDigit int
	for i := len(name) - 1; i >= 0 && '0' <= name[i] && name[i] <= '9'; i-- {
		numDigit++
	}
	// Check for '#' delimiter.
	if numDigit == len(name) || name[len(name)-len("#")-numDigit] != '#' {
		return "", 0
	}
	suffix := name[len(name)-len("#")-numDigit:]
	// Parse the number.
	n, err := strconv.ParseInt(suffix[len("#"):], 10, 64)
	if err != nil {
		return "", 0
	}
	// The "%02d" format requires two digits and cannot have leading zeros.
	if numDigit < 2 || (numDigit > 2 && suffix[len("#")] == '0') {
		return "", 0
	}
	return suffix, n
}

// rewrite rewrites a subname to having only printable characters and no white
// space.
func rewrite(s string) string {
	b := []byte{}
	for _, r := range s {
		switch {
		case isSpace(r):
			b = append(b, '_')
		case !strconv.IsPrint(r):
			s := strconv.QuoteRune(r)
			b = append(b, s[1:len(s)-1]...)
		default:
			b = append(b, string(r)...)
		}
	}
	return string(b)
}

func isSpace(r rune) bool {
	if r < 0x2000 {
		switch r {
		// Note: not the same as Unicode Z class.
		case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0, 0x1680:
			return true
		}
	} else {
		if r <= 0x200a {
			return true
		}
		switch r {
		case 0x2028, 0x2029, 0x202f, 0x205f, 0x3000:
			return true
		}
	}
	return false
}
