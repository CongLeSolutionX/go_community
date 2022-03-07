// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

// These shared constants are used by both the cmd/cover tool (in
// hybrid instrumentation mode) and cmd/compile (in coverage "fixup" mode).
const MetaVarTag = "metavar"
const MetaHashTag = "metahash"
const MetaLenTag = "metalen"
const PkgIdVarTag = "pkgidvar"
const CounterModeTag = "countermode"
const CounterVarTag = "countervar"
const CounterGranularityTag = "countergranularity"
const CounterPrefixTag = "counterprefix"
