// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package http

// DefaultTransport is the default implementation of Transport and is
// used by DefaultClient. It implements the RoundTripper interface
// using the WHATWG Fetch API.
var DefaultTransport RoundTripper = (*jsTransport)(nil)
