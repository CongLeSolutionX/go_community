// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bytes_test

import (
	"bytes"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// FuzzCompatibility verifies that the semantics of identical functions
// in both the strings and bytes package are identical.
func FuzzCompatibility(f *testing.F) {
	for _, tt := range indexTests {
		f.Add([]byte(tt.a), []byte(tt.b), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range lastIndexTests {
		f.Add([]byte(tt.a), []byte(tt.b), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range indexAnyTests {
		f.Add([]byte(tt.a), []byte(tt.b), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range lastIndexAnyTests {
		f.Add([]byte(tt.a), []byte(tt.b), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range lastIndexByteTests {
		f.Add([]byte(tt.a), []byte(tt.b), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range indexRuneTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(tt.rune))
	}
	for _, tt := range splitTests {
		f.Add([]byte(tt.s), []byte(tt.sep), []byte(nil), []byte(nil), rune(tt.n))
	}
	for _, tt := range splitAfterTests {
		f.Add([]byte(tt.s), []byte(tt.sep), []byte(nil), []byte(nil), rune(tt.n))
	}
	for _, tt := range fieldsTests {
		f.Add([]byte(tt.s), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range upperTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range lowerTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range trimSpaceTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range toValidUTF8Tests {
		f.Add([]byte(tt.in), []byte(tt.repl), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range repeatTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(tt.count))
	}
	for _, tt := range runesTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range trimTests {
		f.Add([]byte(tt.in), []byte(tt.arg), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range trimNilTests {
		f.Add([]byte(tt.in), []byte(tt.arg), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range trimFuncTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range indexFuncTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range replaceTests {
		f.Add([]byte(tt.in), []byte(tt.old), []byte(tt.new), []byte(nil), rune(tt.n))
	}
	for _, tt := range titleTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range toTitleTests {
		f.Add([]byte(tt.in), []byte(nil), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range equalFoldTests {
		f.Add([]byte(tt.s), []byte(tt.t), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range cutTests {
		f.Add([]byte(tt.s), []byte(tt.sep), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range containsAnyTests {
		f.Add([]byte(tt.b), []byte(tt.substr), []byte(nil), []byte(nil), rune(0))
	}
	for _, tt := range containsRuneTests {
		f.Add([]byte(tt.b), []byte(nil), []byte(nil), []byte(nil), rune(tt.r))
	}
	f.Fuzz(func(t *testing.T, a, b, c, d []byte, r rune) {
		funcRuneBool := func(x rune) bool {
			switch uint(r) % 10 {
			case 0:
				return unicode.IsSpace(x)
			case 1:
				return unicode.IsDigit(x)
			case 2:
				return unicode.IsUpper(x)
			case 3:
				return !unicode.IsSpace(x)
			case 4:
				return !unicode.IsDigit(x)
			case 5:
				return !unicode.IsUpper(x)
			case 6:
				return r == utf8.RuneError
			case 7:
				return r != utf8.RuneError
			case 8:
				return x == r
			default:
				return x&97 == r&97
			}

		}
		funcRuneRune := func(x rune) rune {
			switch uint(r)%20 + uint(x)%5 {
			case 1:
				return x + 1
			case 2:
				return x - 1
			case 3:
				return x + x
			case 4:
				return -x
			default:
				return x
			}
		}
		var unicodeSpecialCase unicode.SpecialCase
		switch uint(r) % 3 {
		case 0:
			unicodeSpecialCase = unicode.AzeriCase
		case 1:
			unicodeSpecialCase = unicode.TurkishCase
		}

		t.Run("Compare", func(t *testing.T) {
			retB := bytes.Compare([]byte(a), []byte(b))
			retS := strings.Compare(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("Contains", func(t *testing.T) {
			retB := bytes.Contains([]byte(a), []byte(b))
			retS := strings.Contains(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("ContainsAny", func(t *testing.T) {
			retB := bytes.ContainsAny([]byte(a), string(b))
			retS := strings.ContainsAny(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("ContainsRune", func(t *testing.T) {
			retB := bytes.ContainsRune([]byte(a), r)
			retS := strings.ContainsRune(string(a), r)
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("Count", func(t *testing.T) {
			retB := bytes.Count([]byte(a), []byte(b))
			retS := strings.Count(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("Cut", func(t *testing.T) {
			retB0, retB1, retB2 := bytes.Cut([]byte(a), []byte(b))
			retS0, retS1, retS2 := strings.Cut(string(a), string(b))
			switch {
			case string(retB0) != string(retS0):
				t.Fatalf("result.0 mismatch: %q != %q", retB0, retS0)
			case string(retB1) != string(retS1):
				t.Fatalf("result.1 mismatch: %q != %q", retB1, retS1)
			case retB2 != retS2:
				t.Fatalf("result.2 mismatch: %v != %v", retB2, retS2)
			}
		})

		t.Run("EqualFold", func(t *testing.T) {
			retB := bytes.EqualFold([]byte(a), []byte(b))
			retS := strings.EqualFold(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("Fields", func(t *testing.T) {
			retB := bytes.Fields([]byte(a))
			retS := strings.Fields(string(a))
			if len(retB) != len(retS) {
				t.Fatalf("result mismatch: %v != %v", len(retB), len(retS))
			} else {
				for i := range retB {
					if string(retB[i]) != string(retS[i]) {
						t.Fatalf("result.%d mismatch: %q != %q", i, retB[i], retS[i])
					}
				}
			}
		})

		t.Run("FieldsFunc", func(t *testing.T) {
			retB := bytes.FieldsFunc([]byte(a), funcRuneBool)
			retS := strings.FieldsFunc(string(a), funcRuneBool)
			if len(retB) != len(retS) {
				t.Fatalf("result mismatch: %v != %v", len(retB), len(retS))
			} else {
				for i := range retB {
					if string(retB[i]) != string(retS[i]) {
						t.Fatalf("result.%d mismatch: %q != %q", i, retB[i], retS[i])
					}
				}
			}
		})

		t.Run("HasPrefix", func(t *testing.T) {
			retB := bytes.HasPrefix([]byte(a), []byte(b))
			retS := strings.HasPrefix(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("HasSuffix", func(t *testing.T) {
			retB := bytes.HasSuffix([]byte(a), []byte(b))
			retS := strings.HasSuffix(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("Index", func(t *testing.T) {
			retB := bytes.Index([]byte(a), []byte(b))
			retS := strings.Index(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("IndexAny", func(t *testing.T) {
			retB := bytes.IndexAny([]byte(a), string(b))
			retS := strings.IndexAny(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("IndexByte", func(t *testing.T) {
			retB := bytes.IndexByte([]byte(a), byte(r))
			retS := strings.IndexByte(string(a), byte(r))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("IndexFunc", func(t *testing.T) {
			retB := bytes.IndexFunc([]byte(a), funcRuneBool)
			retS := strings.IndexFunc(string(a), funcRuneBool)
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("IndexRune", func(t *testing.T) {
			retB := bytes.IndexRune([]byte(a), r)
			retS := strings.IndexRune(string(a), r)
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("Join", func(t *testing.T) {
			inB := [][]byte{[]byte(a), []byte(b), []byte(c), []byte(a), []byte(b), []byte(c)}
			inS := []string{string(a), string(b), string(c), string(a), string(b), string(c)}
			inB = inB[:uint(r)%uint(len(inB))]
			inS = inS[:uint(r)%uint(len(inS))]
			retB := bytes.Join(inB, []byte(d))
			retS := strings.Join(inS, string(d))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("LastIndex", func(t *testing.T) {
			retB := bytes.LastIndex([]byte(a), []byte(b))
			retS := strings.LastIndex(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("LastIndexAny", func(t *testing.T) {
			retB := bytes.LastIndexAny([]byte(a), string(b))
			retS := strings.LastIndexAny(string(a), string(b))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("LastIndexByte", func(t *testing.T) {
			retB := bytes.LastIndexByte([]byte(a), byte(r))
			retS := strings.LastIndexByte(string(a), byte(r))
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("LastIndexFunc", func(t *testing.T) {
			retB := bytes.LastIndexFunc([]byte(a), funcRuneBool)
			retS := strings.LastIndexFunc(string(a), funcRuneBool)
			if retB != retS {
				t.Fatalf("result mismatch: %v != %v", retB, retS)
			}
		})

		t.Run("Map", func(t *testing.T) {
			retB := bytes.Map(funcRuneRune, []byte(a))
			retS := strings.Map(funcRuneRune, string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("Repeat", func(t *testing.T) {
			retB := bytes.Repeat([]byte(a), int(uint(r)%10))
			retS := strings.Repeat(string(a), int(uint(r)%10))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("Replace", func(t *testing.T) {
			retB := bytes.Replace([]byte(a), []byte(b), []byte(c), int(uint(r)%10))
			retS := strings.Replace(string(a), string(b), string(c), int(uint(r)%10))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ReplaceAll", func(t *testing.T) {
			retB := bytes.ReplaceAll([]byte(a), []byte(b), []byte(c))
			retS := strings.ReplaceAll(string(a), string(b), string(c))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("Split", func(t *testing.T) {
			retB := bytes.Split([]byte(a), []byte(b))
			retS := strings.Split(string(a), string(b))
			if len(retB) != len(retS) {
				t.Fatalf("result mismatch: %v != %v", len(retB), len(retS))
			} else {
				for i := range retB {
					if string(retB[i]) != string(retS[i]) {
						t.Fatalf("result.%d mismatch: %q != %q", i, retB[i], retS[i])
					}
				}
			}
		})

		t.Run("SplitAfter", func(t *testing.T) {
			retB := bytes.SplitAfter([]byte(a), []byte(b))
			retS := strings.SplitAfter(string(a), string(b))
			if len(retB) != len(retS) {
				t.Fatalf("result mismatch: %v != %v", len(retB), len(retS))
			} else {
				for i := range retB {
					if string(retB[i]) != string(retS[i]) {
						t.Fatalf("result.%d mismatch: %q != %q", i, retB[i], retS[i])
					}
				}
			}
		})

		t.Run("SplitAfterN", func(t *testing.T) {
			retB := bytes.SplitAfterN([]byte(a), []byte(b), int(r%10))
			retS := strings.SplitAfterN(string(a), string(b), int(r%10))
			if len(retB) != len(retS) {
				t.Fatalf("result mismatch: %v != %v", len(retB), len(retS))
			} else {
				for i := range retB {
					if string(retB[i]) != string(retS[i]) {
						t.Fatalf("result.%d mismatch: %q != %q", i, retB[i], retS[i])
					}
				}
			}
		})

		t.Run("SplitN", func(t *testing.T) {
			retB := bytes.SplitN([]byte(a), []byte(b), int(r%10))
			retS := strings.SplitN(string(a), string(b), int(r%10))
			if len(retB) != len(retS) {
				t.Fatalf("result mismatch: %v != %v", len(retB), len(retS))
			} else {
				for i := range retB {
					if string(retB[i]) != string(retS[i]) {
						t.Fatalf("result.%d mismatch: %q != %q", i, retB[i], retS[i])
					}
				}
			}
		})

		t.Run("Title", func(t *testing.T) {
			retB := bytes.Title([]byte(a))
			retS := strings.Title(string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ToLower", func(t *testing.T) {
			retB := bytes.ToLower([]byte(a))
			retS := strings.ToLower(string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ToLowerSpecial", func(t *testing.T) {
			retB := bytes.ToLowerSpecial(unicodeSpecialCase, []byte(a))
			retS := strings.ToLowerSpecial(unicodeSpecialCase, string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ToTitle", func(t *testing.T) {
			retB := bytes.ToTitle([]byte(a))
			retS := strings.ToTitle(string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ToTitleSpecial", func(t *testing.T) {
			retB := bytes.ToTitleSpecial(unicodeSpecialCase, []byte(a))
			retS := strings.ToTitleSpecial(unicodeSpecialCase, string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ToUpper", func(t *testing.T) {
			retB := bytes.ToUpper([]byte(a))
			retS := strings.ToUpper(string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ToUpperSpecial", func(t *testing.T) {
			retB := bytes.ToUpperSpecial(unicodeSpecialCase, []byte(a))
			retS := strings.ToUpperSpecial(unicodeSpecialCase, string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("ToValidUTF8", func(t *testing.T) {
			retB := bytes.ToValidUTF8([]byte(a), []byte(b))
			retS := strings.ToValidUTF8(string(a), string(b))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("Trim", func(t *testing.T) {
			retB := bytes.Trim([]byte(a), string(b))
			retS := strings.Trim(string(a), string(b))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimFunc", func(t *testing.T) {
			retB := bytes.TrimFunc([]byte(a), funcRuneBool)
			retS := strings.TrimFunc(string(a), funcRuneBool)
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimLeft", func(t *testing.T) {
			retB := bytes.TrimLeft([]byte(a), string(b))
			retS := strings.TrimLeft(string(a), string(b))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimLeftFunc", func(t *testing.T) {
			retB := bytes.TrimLeftFunc([]byte(a), funcRuneBool)
			retS := strings.TrimLeftFunc(string(a), funcRuneBool)
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimPrefix", func(t *testing.T) {
			retB := bytes.TrimPrefix([]byte(a), []byte(b))
			retS := strings.TrimPrefix(string(a), string(b))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimRight", func(t *testing.T) {
			retB := bytes.TrimRight([]byte(a), string(b))
			retS := strings.TrimRight(string(a), string(b))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimRightFunc", func(t *testing.T) {
			retB := bytes.TrimRightFunc([]byte(a), funcRuneBool)
			retS := strings.TrimRightFunc(string(a), funcRuneBool)
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimSpace", func(t *testing.T) {
			retB := bytes.TrimSpace([]byte(a))
			retS := strings.TrimSpace(string(a))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})

		t.Run("TrimSuffix", func(t *testing.T) {
			retB := bytes.TrimSuffix([]byte(a), []byte(b))
			retS := strings.TrimSuffix(string(a), string(b))
			if string(retB) != string(retS) {
				t.Fatalf("result mismatch: %q != %q", retB, retS)
			}
		})
	})
}
