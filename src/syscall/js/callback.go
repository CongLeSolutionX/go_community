// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package js

import "sync"

var (
	callbacksMu    sync.Mutex
	callbacks             = make(map[uint32]func(Value, []Value) interface{})
	nextCallbackID uint32 = 1
)

// Callback is a Go function that got wrapped for use as a JavaScript callback.
type Callback struct {
	Value // the JavaScript function that queues the callback for execution
	id    uint32
}

// NewCallback returns a wrapped callback function.
//
// Invoking the callback in JavaScript will synchronously call the Go function fn.
// Blocking operations on the callback will block the JavaScript event loop.
// As a consequence, if one callback blocks, other callbacks will not be processed.
// A blocking callback should therefore explicitly start a new goroutine.
//
// Callback.Release must be called to free up resources when the callback will not be used any more.
func NewCallback(fn func(this Value, args []Value) interface{}) Callback {
	callbacksMu.Lock()
	id := nextCallbackID
	nextCallbackID++
	callbacks[id] = fn
	callbacksMu.Unlock()
	return Callback{
		id:    id,
		Value: jsGo.Call("_makeCallbackHelper", id),
	}
}

// Release frees up resources allocated for the callback.
// The callback must not be invoked after calling Release.
func (c Callback) Release() {
	callbacksMu.Lock()
	delete(callbacks, c.id)
	callbacksMu.Unlock()
}

// setCallbackHandler is defined in the runtime package
func setCallbackHandler(fn func())

func init() {
	setCallbackHandler(handleCallback)
}

func handleCallback() {
	cb := jsGo.Get("_pendingCallback")
	if cb == Null() {
		return
	}
	jsGo.Set("_pendingCallback", Null())

	id := uint32(cb.Get("id").Int())
	if id == 0 { // zero indicates deadlock
		select {}
	}
	callbacksMu.Lock()
	f, ok := callbacks[id]
	callbacksMu.Unlock()
	if !ok {
		Global().Get("console").Call("error", "call to closed callback")
		return
	}

	this := cb.Get("this")
	argsObj := cb.Get("args")
	args := make([]Value, argsObj.Length())
	for i := range args {
		args[i] = argsObj.Index(i)
	}
	result := f(this, args)
	cb.Set("result", result)
}
