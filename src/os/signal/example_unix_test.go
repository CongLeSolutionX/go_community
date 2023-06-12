// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package signal_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

// This example passes a context with a signal to tell a blocking function that
// it should abandon its work after a signal is received.
func ExampleNotifyContext() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Fatal(err)
	}

	// On a Unix-like system, pressing Ctrl+C on a keyboard sends a
	// SIGINT signal to the process of the program in execution.
	//
	// This example simulates that by sending a SIGINT signal to itself.
	if err := p.Signal(os.Interrupt); err != nil {
		log.Fatal(err)
	}

	select {
	case <-time.After(time.Second):
		fmt.Println("missed signal")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context canceled"
		stop()                 // stop receiving signal notifications as soon as possible.

		cause := context.Cause(ctx)
		fmt.Println(cause) // prints "context canceled by signal interrupt"
		var cse *signal.CanceledBySignalError
		if errors.As(cause, &cse) {
			fmt.Println("Signal:", cse.Signal) // prints "Signal: interrupt"
		}
	}

	// Output:
	// context canceled
	// context canceled by signal interrupt
	// Signal: interrupt
}
