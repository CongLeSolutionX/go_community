// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file contains various utility functions used by profile-guided optimization (PGO) framework.
package pgo

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
)

// CanInlineOrSpecialize determines whether fn is inlineable or specializable.
func CanInlineOrSpecialize(fn *ir.Func) string {
	var reason string // reason, if any, that the function can not be inlined or specialized
	if fn.Nname == nil {
		reason = "no name"
		return reason
	}

	// If marked "go:noinline", don't inline or specialize
	if fn.Pragma&ir.Noinline != 0 {
		reason = "marked go:noinline"
		return reason
	}

	// If marked "go:norace" and -race compilation, don't inline.
	if base.Flag.Race && fn.Pragma&ir.Norace != 0 {
		reason = "marked go:norace with -race compilation"
		return reason
	}

	// If marked "go:nocheckptr" and -d checkptr compilation, don't inline or specialize.
	if base.Debug.Checkptr != 0 && fn.Pragma&ir.NoCheckPtr != 0 {
		reason = "marked go:nocheckptr"
		return reason
	}

	// If marked "go:cgo_unsafe_args", don't inline or specialize, since the
	// function makes assumptions about its argument frame layout.
	if fn.Pragma&ir.CgoUnsafeArgs != 0 {
		reason = "marked go:cgo_unsafe_args"
		return reason
	}

	// If marked as "go:uintptrkeepalive", don't inline or specialize, since the
	// keep alive information is lost during inlining.
	//
	// TODO(prattmic): This is handled on calls during escape analysis,
	// which is after inlining. Move prior to inlining so the keep-alive is
	// maintained after inlining.
	if fn.Pragma&ir.UintptrKeepAlive != 0 {
		reason = "marked as having a keep-alive uintptr argument"
		return reason
	}

	// If marked as "go:uintptrescapes", don't inline or specialize, since the
	// escape information is lost during inlining.
	if fn.Pragma&ir.UintptrEscapes != 0 {
		reason = "marked as having an escaping uintptr argument"
		return reason
	}

	// The nowritebarrierrec checker currently works at function
	// granularity, so inlining yeswritebarrierrec functions can
	// confuse it (#22342). As a workaround, disallow inlining or specializing
	// them for now.
	if fn.Pragma&ir.Yeswritebarrierrec != 0 {
		reason = "marked go:yeswritebarrierrec"
		return reason
	}

	// If fn has no body (is defined outside of Go), cannot inline or specialize it.
	if len(fn.Body) == 0 {
		reason = "no function body"
		return reason
	}

	// If fn is synthetic hash or eq function, cannot inline it.
	// The function is not generated in Unified IR frontend at this moment.
	if ir.IsEqOrHashFunc(fn) {
		reason = "type eq/hash function"
		return reason
	}

	return ""
}
