// errorcheckdir -goexperiment aliastypeparams

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test case for issue #68526: make sure that the compiler
// doesn't panic when instantiating an imported generic
// type alias.
// TODO: remove/adjust this test for Go1.24.

package ignored
