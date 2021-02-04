<<<<<<< HEAD   (c83a43 [dev.go2go] go/*: merge parser and types changes from dev.ty)
=======
// UNREVIEWED
>>>>>>> BRANCH (dc122c [dev.typeparams] test: exclude a failing test again (fix 32b)
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// The underlying type of Error is the underlying type of error.
// Make sure we can import this again without problems.
type Error error

func F() Error { return nil }
