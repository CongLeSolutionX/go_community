// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

// A NetDevice represents abstracted network device information.
type NetDevice struct {
	Type  int    // see ARPHRD_* in golang.org/x/sys/unix package
	Flags int    // see IFF_* in golang.org/x/sys/unix package
	Alias string // alias for network management
}
