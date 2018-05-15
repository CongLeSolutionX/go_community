// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arch

import (
	"cmd/internal/obj"
	"errors"
	"fmt"
	"strings"
)

// This file encapsulates some of the odd characteristics of the AMD64/I386
// instruction set, to minimize its interaction with the core of the
// assembler.

// X86Suffix handles the special suffix for the AMD64/I386.
// Leading "." in cond is ignored.
func X86Suffix(prog *obj.Prog, cond string) error {
	cond = strings.TrimPrefix(cond, ".")

	suffix := obj.NewX86suffix(cond)
	if !suffix.IsValid() {
		return x86inferSuffixError(cond)
	}

	prog.Scond = uint8(suffix)
	return nil
}

// x86inferSuffixError returns non-nil error that describes what could be
// the cause of suffix parse failure.
//
// At the point this function is executed there is already assembly error,
// so we can burn some clocks to construct good error message.
//
// Reported issues:
//	- duplicated suffixes
//	- illegal rounding/SAE+broadcast combinations
//	- unknown suffixes
//	- misplaced suffix (e.g. wrong Z suffix position)
func x86inferSuffixError(cond string) error {
	suffixSet := make(map[string]bool)  // Set for duplicates detection.
	unknownSet := make(map[string]bool) // Set of unknown suffixes.
	hasBcst := false
	hasRoundSae := false
	var msg []string // Error message parts

	suffixes := strings.Split(cond, ".")
	for i, suffix := range suffixes {
		switch suffix {
		case "Z":
			if i != len(suffixes)-1 {
				msg = append(msg, "Z suffix should be the last")
			}
		case "BCST":
			hasBcst = true
		case "SAE", "RN_SAE", "RZ_SAE", "RD_SAE", "RU_SAE":
			hasRoundSae = true
		default:
			if !unknownSet[suffix] {
				msg = append(msg, fmt.Sprintf("unknown %q", suffix))
			}
			unknownSet[suffix] = true
		}

		if suffixSet[suffix] {
			msg = append(msg, fmt.Sprintf("duplicate %q", suffix))
		}
		suffixSet[suffix] = true
	}

	if hasBcst && hasRoundSae {
		msg = append(msg, "can't combine rounding/SAE and broadcast")
	}

	if len(msg) == 0 {
		return errors.New("bad suffix combination")
	}
	return errors.New(strings.Join(msg, "; "))
}
