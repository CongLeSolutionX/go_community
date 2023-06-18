// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix

package signal_test

import (
	"context"
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
<<<<<<< PATCH SET (bf891b os/signal: suggestions for NotifyContext cancelation signal )
		cause := context.Cause(ctx)
		if errors.Is(cause, os.SignalError{Signal: os.Interrupt}) {
			fmt.Println(cause)
		}
		stop() // stop receiving signal notifications as soon as possible.
=======
		fmt.Println(ctx.Err()) // prints "context canceled"
		stop()                 // stop receiving signal notifications as soon as possible.
>>>>>>> BASE      (ed8cba cmd/asm: add s390x crypto related instructions)
	}

	// Output:
<<<<<<< PATCH SET (bf891b os/signal: suggestions for NotifyContext cancelation signal )
	// canceled by interrupt signal
=======
	// context canceled
>>>>>>> BASE      (ed8cba cmd/asm: add s390x crypto related instructions)
}
