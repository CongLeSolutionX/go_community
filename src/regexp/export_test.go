// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package regexp

var (
	CompileOnePass     = compileOnePass
	MaxBacktrackVector = maxBacktrackVector
	Special            = special
	MinInputLen        = minInputLen
	MergeFailed        = mergeFailed
	CompileInternal    = compile
	MergeRuneSets      = mergeRuneSets
)

func SetLongest(r *Regexp, l bool) {
	r.longest = l
}

func IsOnePassNil(r *Regexp) bool {
	return r.onepass == nil
}
