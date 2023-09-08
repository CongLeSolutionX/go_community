// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Patterns for ServeMux routing.

package http

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// A pattern is something that can be matched against an HTTP request.
// It has an optional method, an optional host, and a path.
type pattern struct {
	str    string // original string
	method string
	host   string
	// The representation of a path differs from the surface syntax.
	// Paths ending in '/' are represented with an anonymous "..." wildcard.
	// Paths ending in "{$}" are represented with the literal segment "/".
	// This makes most algorithms simpler.
	segments []segment
	loc      string // source location of registering call, for helpful messages
}

func (p *pattern) String() string { return p.str }

func (p *pattern) lastSegment() segment {
	return p.segments[len(p.segments)-1]
}

// A segment is a pattern piece that matches one or more path segments, or
// a trailing slash.
// If wild is false, it matches a literal segment, or, if s == "/", a trailing slash.
// If wild is true and multi is false, it matches a single path segment.
// If both wild and multi are true, it matches all remaining path segments.
type segment struct {
	s     string // literal or wildcard name or "/" for "/{$}".
	wild  bool
	multi bool // "..." wildcard
}

// parsePattern parses a string into a Pattern.
// The string's syntax is
//
//	[METHOD] [HOST]/[PATH]
//
// where:
//   - METHOD is an HTTP method
//   - HOST is a hostname
//   - PATH consists of slash-separated segments, where each segment is either
//     a literal or a wildcard of the form "{name}", "{name...}", or "{$}".
//
// METHOD, HOST and PATH are all optional; that is, the string can be "/".
// If METHOD is present, it must be followed by a single space.
// Wildcard names must be valid Go identifiers.
// The "{$}" and "{name...}" wildcard must occur at the end of PATH.
// PATH may end with a '/'.
// Wildcard names in a path must be distinct.
func parsePattern(s string) (*pattern, error) {
	if len(s) == 0 {
		return nil, errors.New("empty pattern")
	}
	method, rest, found := strings.Cut(s, " ")
	if !found {
		rest = method
		method = ""
	}
	// A method name can be any valid HTTP token.
	// https://www.rfc-editor.org/rfc/rfc9110.html#name-overview.
	if method != "" && !isValidHTTPToken(method) {
		return nil, fmt.Errorf("bad method %q", method)
	}
	p := &pattern{str: s, method: method}

	i := strings.IndexByte(rest, '/')
	if i < 0 {
		return nil, errors.New("host/path missing /")
	}
	p.host = rest[:i]
	rest = rest[i:]
	if strings.IndexByte(p.host, '{') >= 0 {
		return nil, errors.New("host contains '{' (missing initial '/'?")
	}
	// At this point, rest is the path.

	// An unclean path with a method that is not CONNECT can never match,
	// because paths are cleaned before matching.
	if method != "" && method != "CONNECT" && rest != cleanPath(rest) {
		return nil, errors.New("non-CONNECT pattern with unclean path can never match")
	}

	seenNames := map[string]bool{} // remember wildcard names to catch dups
	for len(rest) > 0 {
		// Invariant: rest[0] == '/'.
		rest = rest[1:]
		if len(rest) == 0 {
			// Trailing slash.
			p.segments = append(p.segments, segment{wild: true, multi: true})
			break
		}
		i := strings.IndexByte(rest, '/')
		if i < 0 {
			i = len(rest)
		}
		var seg string
		seg, rest = rest[:i], rest[i:]
		if i := strings.IndexByte(seg, '{'); i < 0 {
			// Literal.
			p.segments = append(p.segments, segment{s: seg})
		} else {
			// Wildcard.
			if i != 0 {
				return nil, errors.New("bad wildcard segment (must start with '{')")
			}
			if seg[len(seg)-1] != '}' {
				return nil, errors.New("bad wildcard segment (must end with '}')")
			}
			name := seg[1 : len(seg)-1]
			if name == "$" {
				if len(rest) != 0 {
					return nil, errors.New("{$} not at end")
				}
				p.segments = append(p.segments, segment{s: "/"})
				break
			}
			var multi bool
			if strings.HasSuffix(name, "...") {
				multi = true
				name = name[:len(name)-3]
				if len(rest) != 0 {
					return nil, errors.New("{...} wildcard not at end")
				}
			}
			if name == "" {
				return nil, errors.New("empty wildcard")
			}
			if !isValidWildcardName(name) {
				return nil, fmt.Errorf("bad wildcard name %q", name)
			}
			if seenNames[name] {
				return nil, fmt.Errorf("duplicate wildcard name %q", name)
			}
			seenNames[name] = true
			p.segments = append(p.segments, segment{s: name, wild: true, multi: multi})
		}
	}
	return p, nil
}

func isValidHTTPToken(s string) bool {
	if s == "" {
		return false
	}
	// See https://www.rfc-editor.org/rfc/rfc9110#section-5.6.2.
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("!#$%&'*+.^_`|~-", r) {
			return false
		}
	}
	return true
}

func isValidWildcardName(s string) bool {
	if s == "" {
		return false
	}
	// Valid Go identifier.
	for i, c := range s {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return true
}

// relationship is a relationship between two patterns, p1 and p2.
type relationship string

const (
	equivalent   relationship = "equivalent"   // both match the same requests
	moreGeneral  relationship = "moreGeneral"  // p1 matches everything p2 does & more
	moreSpecific relationship = "moreSpecific" // p2 matches everything p1 does & more
	disjoint     relationship = "disjoint"     // there is no request that both match
	overlaps     relationship = "overlaps"     // there is a request that both match, but neither is more specific
)

// conflictsWith reports whether p1 conflicts with p2, that is, whether
// there is a request that both match but where neither is higher precedence
// than the other.
//
//	Precedence is defined by two rules:
//	1. Patterns with a host win over patterns without a host.
//	2. Patterns whose method and path is more specific win. One pattern is more
//	   specific than another if the second matches all the (method, path) pairs
//	   of the first and more.
func (p1 *pattern) conflictsWith(p2 *pattern) bool {
	if p1.host != p2.host {
		// Either one host is empty and the other isn't, in which case the
		// one with the host wins by rule 1, or neither host is empty
		// and they differ, so they won't match the same paths.
		return false
	}
	rel := p1.comparePathsAndMethods(p2)
	return rel == equivalent || rel == overlaps
}

func (p1 *pattern) comparePathsAndMethods(p2 *pattern) relationship {
	mrel := p1.compareMethods(p2)
	// Optimization: avoid a call to comparePaths.
	if mrel == disjoint {
		return disjoint
	}
	prel := p1.comparePaths(p2)
	return combineRelationships(mrel, prel)
}

// compareMethods determines the relationship between the method
// part of patterns p1 and p2.
//
// A method can either be empty, "GET", or something else.
// The empty string matches any method, so it is the most general.
// "GET" matches both GET and HEAD.
// Anything else matches only itself.
func (p1 *pattern) compareMethods(p2 *pattern) relationship {
	if p1.method == p2.method {
		return equivalent
	}
	if p1.method == "" {
		// p1 matches any method, but p2 does not, so p1 is more general.
		return moreGeneral
	}
	if p2.method == "" {
		return moreSpecific
	}
	if p1.method == "GET" && p2.method == "HEAD" {
		// p1 matches GET and HEAD; p2 matches only HEAD.
		return moreGeneral
	}
	if p2.method == "GET" && p1.method == "HEAD" {
		return moreSpecific
	}
	return disjoint
}

// comparePaths determines the relationship between the path
// part of two patterns.
func (p1 *pattern) comparePaths(p2 *pattern) relationship {
	// If a path pattern doesn't end in a multi ("...") wildcard, then it
	// can only match paths with the same number of segments.
	if len(p1.segments) != len(p2.segments) && !p1.lastSegment().multi && !p2.lastSegment().multi {
		return disjoint
	}
	// Track whether a single (non-multi) wildcard in p1 matched
	// a literal in p2, and vice versa.
	// We care about these because if a wildcard matches a literal, then the
	// pattern with the wildcard can't be more specific than the one with the
	// literal.
	wild1MatchedLit2 := false
	wild2MatchedLit1 := false
	var segs1, segs2 []segment
	// Look at corresponding segments in the two path patterns.
	for segs1, segs2 = p1.segments, p2.segments; len(segs1) > 0 && len(segs2) > 0; segs1, segs2 = segs1[1:], segs2[1:] {
		s1 := segs1[0]
		s2 := segs2[0]
		if s1.multi && s2.multi {
			// Two multis match each other.
			continue
		}
		if s1.multi {
			// p1 matches the rest of p2.
			// Does that mean it is more general than p2?
			if !wild2MatchedLit1 {
				// If p2 didn't have any wildcards that matched literals in p1,
				// then yes, p1 is more general.
				return moreGeneral
			}
			// Otherwise neither is more general than the other.
			return overlaps
		}
		if s2.multi {
			// p2 matches the rest of p1. The same logic as above applies.
			if !wild1MatchedLit2 {
				return moreSpecific
			}
			return overlaps
		}
		if s1.s == "/" && s2.s == "/" {
			// Both patterns end in "/{$}"; they match.
			continue
		}
		if s1.s == "/" || s2.s == "/" {
			// One pattern ends in "/{$}", and the other doesn't, nor is the other's
			// corresponding segment a multi. So they are disjoint.
			return disjoint
		}
		if s1.wild && s2.wild {
			// These single-segment wildcards match each other.
		} else if s1.wild {
			// p1's single wildcard matches the corresponding segment of p2.
			wild1MatchedLit2 = true
		} else if s2.wild {
			// p2's single wildcard matches the corresponding segment of p1.
			wild2MatchedLit1 = true
		} else {
			// Two literal segments.
			if s1.s != s2.s {
				return disjoint
			}
		}
	}
	// We've reached the end of the corresponding segments of the patterns.
	if len(segs1) == 0 && len(segs2) == 0 {
		// The patterns matched completely.
		switch {
		case wild1MatchedLit2 && !wild2MatchedLit1:
			// p1 had a wildcard where p2 had a literal, and not vice versa.
			return moreGeneral
		case wild2MatchedLit1 && !wild1MatchedLit2:
			// The opposite.
			return moreSpecific
		case !wild1MatchedLit2 && !wild2MatchedLit1:
			// Neither had a wildcard that matched a literal, so wildcards
			// matched wildcards and literals matched literals.
			return equivalent
		default:
			return overlaps
		}
	}
	// One pattern has more segments than the other.
	// The only way they can fail to be disjoint is if one ends in a multi, but
	// we handled that case in the loop.
	return disjoint
}

// combineRelationships determines the overall relationship of two patterns
// given the relationships of two parts of the patterns.
//
// For example, if p1 is more general than p2 in one way but equivalent
// in another, then it is more general overall.
//
// Or if p1 is more general in one way and more specific in another, then
// they overlap.
func combineRelationships(methodRel, pathRel relationship) relationship {
	switch methodRel {
	case equivalent:
		return pathRel
	case disjoint:
		return disjoint
	case overlaps:
		panic("methods can't overlap")
	case moreGeneral, moreSpecific:
		switch pathRel {
		case equivalent:
			return methodRel
		case inverseRelationship(methodRel):
			return overlaps
		default:
			return pathRel
		}
	default:
		panic(fmt.Sprintf("unknown relationship %q", methodRel))
	}
}

func inverseRelationship(r relationship) relationship {
	switch r {
	case moreSpecific:
		return moreGeneral
	case moreGeneral:
		return moreSpecific
	default:
		return r
	}
}
