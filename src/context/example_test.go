// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context_test

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func startStreaming(ctx context.Context, cancel context.CancelFunc) error {
	// This is a fake implementation of startStreaming.
	// Actual implementation will start reading from a remote resource
	// and will only cancel if connection is broken, etc.
	cancel()
	return errors.New("cannot fetch the resource")
}

func ExampleWithCancel() {
	// Create a context that can later be cancelable.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// In this example, we will start streaming results from a remote service.
	// If any errors occur, startStreaming will also cancel the context.
	go startStreaming(ctx, cancel)

	select {
	case <-time.After(5 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context canceled"
	}
	// Output:
	// context canceled
}

func ExampleWithTimeout() {
	// Pass a context with a timeout to tell a blocking function that it
	// should abandon its work after the timeout elapses.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context deadline exceeded"
	}

	// Even though ctx should have expired already, it is good
	// practice to call its cancelation function in any case.
	// Failure to do so may keep the context and its parent alive
	// longer than necessary.
	cancel()

	// Output:
	// context deadline exceeded
}
