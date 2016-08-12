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

// This is a summarized example of a complete use case for context. It
// demonstrates the cancellation of events generated in the scope of a HTTP
// request and passing of values along the call chain.
func Example() {
	const (
		// Experiment changing timeout to check how the behavior changes.
		timeout = 1 * time.Second

		timerStart = "start"
	)

	slowOp := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "slow operation started")

		// Sleep pretending to be a slow operation.
		opwait, timer := time.After(2*time.Second), ctx.Value(timerStart).(time.Time)

		select {
		case <-opwait:
			fmt.Fprintln(w, "long operation duration:", time.Now().Sub(timer))
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

func ExampleWithCancel() {
	ctx, cancel := context.WithCancel(context.Background())

	op := func(ctx context.Context, wg *sync.WaitGroup) {
		t := time.NewTicker(50 * time.Millisecond)
		for range t.C {
			select {
			case <-ctx.Done():
				fmt.Println("cancelled:", ctx.Err())
				t.Stop()
				wg.Done()
				return
			default:
				// Do work...
			}
		}
	}

	var wg sync.WaitGroup

	// Three simultaneous operations that are going to be cancelled by a single cancel() call.
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go op(ctx, &wg)
	}

	cancel()
	wg.Wait()

	// Output:
	// cancelled: context canceled
	// cancelled: context canceled
	// cancelled: context canceled
}

func ExampleWithDeadline() {
	// Pass a context with a timeout to tell a blocking function that it
	// should abandon its work after the timeout elapses.
	d := time.Now().Add(50 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), d)

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
