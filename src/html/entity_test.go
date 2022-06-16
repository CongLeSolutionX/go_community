// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"testing"
	"unicode/utf8"
)

func init() {
	// force load of entity maps
	if supportsHtmlAVX512 && !(disableHtmlAVX512 == "1") {
		populateMapsOnce.Do(populateMaps) // force load of entity maps
	} else {
		UnescapeString("")
	}
}

func TestEntitiesAvx(t *testing.T) {
	if supportsHtmlAVX512 && !(disableHtmlAVX512 == "1") {
		// validate avx-optimized entities == reference entities
		// first for the entity map
		for k, v := range entity {
			r0 := make([]byte, 4)
			len := utf8.EncodeRune(r0, v)
			ref := string(r0[:len])
			test := string([]byte(UnescapeString(EscapeString(UnescapeString("&" + k)))))
			if test != ref {
				t.Errorf("optimized unescape(%s) entity error, got %x expected %x", k, []byte(test), []byte(ref))
			}
		}
		// then for the entity2 map
		for k, v := range entity2 {
			r0 := make([]byte, 4)
			r1 := make([]byte, 4)
			l0 := utf8.EncodeRune(r0, v[0])
			l1 := utf8.EncodeRune(r1, v[1])
			ref := string(append(r0[:l0], string(r1[:l1])...))
			test := string([]byte(UnescapeString(EscapeString(UnescapeString("&" + k)))))
			if test != ref {
				t.Errorf("optimized unescape(%s) entity2 error, got %x expected %x", k, []byte(test), []byte(ref))
			}
		}
	}
}

func TestEntityLength(t *testing.T) {
	if len(entity) == 0 || len(entity2) == 0 {
		t.Fatal("maps not loaded")
	}

	// We verify that the length of UTF-8 encoding of each value is <= 1 + len(key).
	// The +1 comes from the leading "&". This property implies that the length of
	// unescaped text is <= the length of escaped text.
	for k, v := range entity {
		if 1+len(k) < utf8.RuneLen(v) {
			t.Error("escaped entity &" + k + " is shorter than its UTF-8 encoding " + string(v))
		}
		if len(k) > longestEntityWithoutSemicolon && k[len(k)-1] != ';' {
			t.Errorf("entity name %s is %d characters, but longestEntityWithoutSemicolon=%d", k, len(k), longestEntityWithoutSemicolon)
		}
	}
	for k, v := range entity2 {
		if 1+len(k) < utf8.RuneLen(v[0])+utf8.RuneLen(v[1]) {
			t.Error("escaped entity &" + k + " is shorter than its UTF-8 encoding " + string(v[0]) + string(v[1]))
		}
	}
}
