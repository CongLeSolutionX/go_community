// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package signal_test

import (
	"fmt"
	"os"
	"os/signal"
)

func ExampleNotify() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)
}

func ExampleNotify_noSignalDefined() {
	// Set up channel on which to send signal notifications.
	// The buffered channel is used
	// to avoid missing the signal.
	c := make(chan os.Signal, 1)

	// No signals are defined in
	// Notify arguments. This way
	// all signals would be catched.
	signal.Notify(c)

	// Block until any signal is received.
	s := <-c
	fmt.Println("Got signal:", s)
}
