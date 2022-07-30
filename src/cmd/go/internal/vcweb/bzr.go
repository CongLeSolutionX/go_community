// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcweb

import "net/http"

type bzrHandler struct{}

func (*bzrHandler) Available() bool { return true }

func (*bzrHandler) Handler(dir string, env []string) (http.Handler, error) {
	return http.FileServer(http.Dir(dir)), nil
}
