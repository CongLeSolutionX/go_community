// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fakenet

import "sync/atomic"

// This variable controls whether the fake networking faciliy used by some
// platforms gets enabled.
//
// The fake network is used for js and wasip1 to execute tests in net
// subpackages like net/http.
//
// We have to use a runtime check because there is no mechanism by which we can
// detect at compile time that we are building a test.
var enabled atomic.Bool

func IsEnabled() bool { return enabled.Load() }

func SetEnabled(set bool) { enabled.Store(set) }
