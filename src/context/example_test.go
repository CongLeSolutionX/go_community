// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// This is a typical use case for context package. It demonstrates the
// abandonment of events generated in the scope of a HTTP request and passing
// of values down the call chain.
func Example() {
	const (
		// Experiment changing timeout to check how the behavior changes.
		timeout = 1 * time.Second

		timerStart = "start"
	)

	slowOp := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "slow operation started")

		ts := ctx.Value(timerStart).(time.Time)

		// Sleep pretending to be a slow operation.
		op := time.After(2 * time.Second)

		select {
		case <-op:
			fmt.Fprintln(w, "long operation duration:", time.Now().Sub(ts))
		case <-ctx.Done():
			fmt.Fprintln(w, "context signal received. error:", ctx.Err())
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timerCtx := context.WithValue(context.Background(), timerStart, time.Now())
		ctx, cancel := context.WithTimeout(timerCtx, timeout)
		defer cancel()

		fmt.Fprintln(w, "request received")
		slowOp(ctx, w, r)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", b)
	// Output:
	// request received
	// slow operation started
	// context signal received. error: context deadline exceeded
}

// This example shows how CancelFunc can be used to abandon several simultaneous
// operations with a single call.
func ExampleWithCancel() {
	ctx, cancel := context.WithCancel(context.Background())

	op := func(ctx context.Context, wg *sync.WaitGroup) {
		t := time.NewTicker(50 * time.Millisecond)
		for range t.C {
			select {
			case <-ctx.Done():
				fmt.Println("canceled:", ctx.Err())
				t.Stop()
				wg.Done()
				return
			default:
				// Do work...
			}
		}
	}

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go op(ctx, &wg)
	}

	cancel()
	wg.Wait()

	// Output:
	// canceled: context canceled
	// canceled: context canceled
	// canceled: context canceled
}

// This example passes a context with a arbitrary deadline to tell a blocking
// function that it should abandon its work as soon as it gets to it.
func ExampleWithDeadline() {
	d := time.Now().Add(50 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), d)

	// Even though ctx will be expired, it is good practice to call its
	// cancelation function in any case. Failure to do so may keep the
	// context and its parent alive longer than necessary.
	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err())
	}

	// Output:
	// context deadline exceeded
}

// This example passes a context with a timeout to tell a blocking function that
// it should abandon its work after the timeout elapses.
func ExampleWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)

	// Even though ctx will be expired, it is good practice to call its
	// cancelation function in any case. Failure to do so may keep the
	// context and its parent alive longer than necessary.
	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err())
	}

	// Output:
	// context deadline exceeded
}

func ExampleWithValue() {
	type key int

	f := func(ctx context.Context, k key) {
		if v := ctx.Value(k); v != nil {
			fmt.Println("found value:", v)
		} else {
			fmt.Println("key not found:", k)
		}
	}

	k := key(5)
	ctx := context.WithValue(context.Background(), k, "Go")

	f(ctx, k)
	f(ctx, key(0))

	// Output:
	// found value: Go
	// key not found: 0
}
