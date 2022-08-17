// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// lateLower applies those rules that need to be run after the general lower rule.
func lateLower(f *Func) {
	// repeat rewrites until we find no more rewrites
	if f.Config.lateLowerValue != nil {
		applyRewrite(f, f.Config.lowerBlock, f.Config.lateLowerValue, removeDeadValues)
	}
}
