// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pty is a simple pseudo-terminal package for Unix systems,
// implemented by calling C functions via cgo.
// This is only used for testing the os/signal package.
package testpty

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

type PtyError struct {
	FuncName    string
	ErrorString string
	Errno       syscall.Errno
}

func ptyError(name string, err error) *PtyError {
	return &PtyError{name, err.Error(), err.(syscall.Errno)}
}

func (e *PtyError) Error() string {
	return fmt.Sprintf("%s: %s", e.FuncName, e.ErrorString)
}

func (e *PtyError) Unwrap() error { return e.Errno }

var ErrNotSupported = errors.New("testpty.Open not implemented on this platform")

// Open returns a control pty and the name of the linked process tty.
//
// If Open is not implemented on this platform, it returns ErrNotSupported.
func Open() (pty *os.File, processTTY string, err error) {
	return open()
}
