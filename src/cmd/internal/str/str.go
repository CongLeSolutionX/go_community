// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package str provides string manipulation utilities.
package str

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// StringList flattens its arguments into a single []string.
// Each argument in args must have type string or []string.
func StringList(args ...interface{}) []string {
	var x []string
	for _, arg := range args {
		switch arg := arg.(type) {
		case []string:
			x = append(x, arg...)
		case string:
			x = append(x, arg)
		default:
			panic("stringList: invalid argument of type " + fmt.Sprintf("%T", arg))
		}
	}
	return x
}

// ToFold returns a string with the property that
//	strings.EqualFold(s, t) iff ToFold(s) == ToFold(t)
// This lets us test a large set of strings for fold-equivalent
// duplicates without making a quadratic number of calls
// to EqualFold. Note that strings.ToUpper and strings.ToLower
// do not have the desired property in some corner cases.
func ToFold(s string) string {
	// Fast path: all ASCII, no upper case.
	// Most paths look like this already.
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf || 'A' <= c && c <= 'Z' {
			goto Slow
		}
	}
	return s

Slow:
	var buf bytes.Buffer
	for _, r := range s {
		// SimpleFold(x) cycles to the next equivalent rune > x
		// or wraps around to smaller values. Iterate until it wraps,
		// and we've found the minimum value.
		for {
			r0 := r
			r = unicode.SimpleFold(r0)
			if r <= r0 {
				break
			}
		}
		// Exception to allow fast path above: A-Z => a-z
		if 'A' <= r && r <= 'Z' {
			r += 'a' - 'A'
		}
		buf.WriteRune(r)
	}
	return buf.String()
}

// FoldDup reports a pair of strings from the list that are
// equal according to strings.EqualFold.
// It returns "", "" if there are no such strings.
func FoldDup(list []string) (string, string) {
	clash := map[string]string{}
	for _, s := range list {
		fold := ToFold(s)
		if t := clash[fold]; t != "" {
			if s > t {
				s, t = t, s
			}
			return s, t
		}
		clash[fold] = s
	}
	return "", ""
}

// Contains reports whether x contains s.
func Contains(x []string, s string) bool {
	for _, t := range x {
		if t == s {
			return true
		}
	}
	return false
}

// Uniq removes consecutive duplicate strings from ss.
func Uniq(ss *[]string) {
	if len(*ss) <= 1 {
		return
	}
	uniq := (*ss)[:1]
	for _, s := range *ss {
		if s != uniq[len(uniq)-1] {
			uniq = append(uniq, s)
		}
	}
	*ss = uniq
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// SplitQuotedFields splits s into a list of fields,
// allowing single or double quotes around elements.
// There is no unescaping or other processing within
// quoted fields.
func SplitQuotedFields(s string) ([]string, error) {
	// Split fields allowing '' or "" around elements.
	// Quotes further inside the string do not count.
	var f []string
	for len(s) > 0 {
		for len(s) > 0 && isSpaceByte(s[0]) {
			s = s[1:]
		}
		if len(s) == 0 {
			break
		}
		// Accepted quoted string. No unescaping inside.
		if s[0] == '"' || s[0] == '\'' {
			quote := s[0]
			s = s[1:]
			i := 0
			for i < len(s) && s[i] != quote {
				i++
			}
			if i >= len(s) {
				return nil, fmt.Errorf("unterminated %c string", quote)
			}
			f = append(f, s[:i])
			s = s[i+1:]
			continue
		}
		i := 0
		for i < len(s) && !isSpaceByte(s[i]) {
			i++
		}
		f = append(f, s[:i])
		s = s[i:]
	}
	return f, nil
}

// SplitQuotedFieldsAndUnescape splits the string s around each instance of one
// or more consecutive white space characters while taking into account quotes
// and escaping, and returns an array of substrings of s or an empty list if s
// contains only white space. Single quotes and double quotes are recognized to
// prevent splitting within the quoted region, and are removed from the
// resulting substrings. If a quote in s isn't closed err will be set and r will
// have the unclosed argument as the last element. The backslash is used for
// escaping.
//
// For example, the following string:
//
//     a b:"c d" 'e''f'  "g\""
//
// Would be parsed as:
//
//     []string{"a", "b:c d", "ef", `g"`}
//
// NOTE: This is a copy of go/build.splitQuoted. Keep in sync.
func SplitQuotedFieldsAndUnescape(s string) (r []string, err error) {
	var args []string
	arg := make([]rune, len(s))
	escaped := false
	quoted := false
	quote := '\x00'
	i := 0
	for _, rune := range s {
		switch {
		case escaped:
			escaped = false
		case rune == '\\':
			escaped = true
			continue
		case quote != '\x00':
			if rune == quote {
				quote = '\x00'
				continue
			}
		case rune == '"' || rune == '\'':
			quoted = true
			quote = rune
			continue
		case unicode.IsSpace(rune):
			if quoted || i > 0 {
				quoted = false
				args = append(args, string(arg[:i]))
				i = 0
			}
			continue
		}
		arg[i] = rune
		i++
	}
	if quoted || i > 0 {
		args = append(args, string(arg[:i]))
	}
	if quote != 0 {
		err = errors.New("unclosed quote")
	} else if escaped {
		err = errors.New("unfinished escaping")
	}
	return args, err
}

// JoinAndQuoteFields joins a list of arguments into a string that can be
// parsed with SplitQuotedFieldsAndUnescape. Arguments containing control
// characters (backslashes, single, and double quotes) are escaped. Arguments
// containing spaces are single quoted. Other arguments are inserted
// without quoting.
func JoinAndQuoteFields(args []string) string {
	containsSpace := func(s string) bool {
		for _, r := range s {
			if unicode.IsSpace(r) {
				return true
			}
		}
		return false
	}

	sb := &strings.Builder{}
	for i, arg := range args {
		if i > 0 {
			sb.WriteString(" ")
		}
		if strings.ContainsAny(arg, `\'"`) {
			// Argument contains characters we need to escape.
			// We won't add quotes, and if we see spaces, we'll escape them, too.
			for _, r := range arg {
				if r == '\\' || r == '\'' || r == '"' || unicode.IsSpace(r) {
					sb.WriteByte('\\')
				}
				sb.WriteRune(r)
			}
		} else if containsSpace(arg) {
			// Argument contains space. Let's add quotes.
			sb.WriteByte('\'')
			sb.WriteString(arg)
			sb.WriteByte('\'')
		} else {
			// Regular old argument. No quote or escape needed.
			sb.WriteString(arg)
		}
	}
	return sb.String()
}

// QuotedStringListFlag is a convenience function for parsing a list of string
// arguments encoded with JoinAndQuoteFields. This is useful for flags like
// cmd/link's -extldflags.
type QuotedStringListFlag []string

var _ flag.Value = (*QuotedStringListFlag)(nil)

func (f *QuotedStringListFlag) Set(v string) error {
	fs, err := SplitQuotedFieldsAndUnescape(v)
	if err != nil {
		return err
	}
	*f = fs[:len(fs):len(fs)]
	return nil
}

func (f *QuotedStringListFlag) String() string {
	if f == nil {
		return ""
	}
	return JoinAndQuoteFields(*f)
}
