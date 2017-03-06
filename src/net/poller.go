// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "internal/poll"

// A PolledOp represents a type of IO operation using the
// runtime-integrated network poller.
type PolledOp = poll.Op

const (
	PolledRead  = poll.ReadOp
	PolledWrite = poll.WriteOp
)

// A Poller is a handle for using the runtime-integrated network
// poller with a user-defined function.
type Poller struct {
	fd *netFD
}

// Call calls the user-defined function fn for non-IO operation.
//
// The user-defined function fn may do any non-IO operation using the
// passed socket descriptor.
// Calling any IO method or function provided by the net package from
// inside fn is prohibited.
func (p *Poller) Call(fn func(s uintptr) error) error {
	// see https://go-review.googlesource.com/c/37039/
	return nil
}

// Run starts IO operation associated with the user-defined function
// fn.
//
// The user-defined function fn may do read or write operation using
// the passed socket descriptor, and must return true when
// accomplished.
// Calling any IO method or function provided by the net package from
// inside fn is prohibited.
func (p *Poller) Run(op PolledOp, fn func(s uintptr) (done bool, err error)) error {
	// see https://go-review.googlesource.com/c/37039/
	return nil
}

// Poller returns a poller associated with c.
func (c *UDPConn) Poller() *Poller {
	return &Poller{fd: c.fd}
}
