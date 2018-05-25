// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package http

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"syscall/js"
)

// RoundTrip implements the RoundTripper interface using the WHATWG Fetch API.
// It supports streaming response bodies.
// See https://fetch.spec.whatwg.org/ for more information on this standard.
func (*Transport) RoundTrip(req *Request) (*Response, error) {
	if isLocal(req) {
		return t.roundTrip(req) // use fake in-memory network for tests
	}
	headers := js.Global.Get("Headers").New()
	for key, values := range req.Header {
		for _, value := range values {
			headers.Call("append", key, value)
		}
	}

	ac := js.Global.Get("AbortController").New()

	opt := js.Global.Get("Object").New()
	opt.Set("headers", headers)
	opt.Set("method", req.Method)
	opt.Set("credentials", "same-origin")
	opt.Set("signal", ac.Get("signal"))

	if req.Body != nil {
		// TODO(johanbrandhorst): Stream request body when possible.
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			req.Body.Close() // RoundTrip must always close the body, including on errors.
			return nil, err
		}
		req.Body.Close()
		opt.Set("body", body)
	}
	respPromise := js.Global.Call("fetch", req.URL.String(), opt)
	var (
		respCh = make(chan *Response, 1)
		errCh  = make(chan error, 1)
	)
	success := js.NewCallback(func(args []js.Value) {
		result := args[0]
		header := Header{}
		writeHeaders := js.NewCallback(func(args []js.Value) {
			key, value := args[0].String(), args[1].String()
			ck := CanonicalHeaderKey(key)
			header[ck] = append(header[ck], value)
		})
		defer writeHeaders.Close()
		result.Get("headers").Call("forEach", writeHeaders)

		contentLength := int64(-1)
		if cl, err := strconv.ParseInt(header.Get("Content-Length"), 10, 64); err == nil {
			contentLength = cl
		}

		b := result.Get("body")
		var body io.ReadCloser
		if b != js.Undefined {
			body = &streamReader{stream: b.Call("getReader")}
		} else {
			// Fall back to using ArrayBuffer
			// https://developer.mozilla.org/en-US/docs/Web/API/Body/arrayBuffer
			body = &arrayReader{arrayPromise: result.Call("arrayBuffer")}
		}

		select {
		case respCh <- &Response{
			Status:        result.Get("status").String() + " " + StatusText(result.Get("status").Int()),
			StatusCode:    result.Get("status").Int(),
			Header:        header,
			ContentLength: contentLength,
			Body:          body,
			Request:       req,
		}:
		case <-req.Context().Done():
		}
	})
	defer success.Close()
	failure := js.NewCallback(func(args []js.Value) {
		err := fmt.Errorf("net/http: fetch() failed: %s", args[0].String())
		select {
		case errCh <- err:
		case <-req.Context().Done():
		}
	})
	defer failure.Close()
	respPromise.Call("then", success, failure)
	select {
	case <-req.Context().Done():
		// Abort the Fetch request
		ac.Call("abort")
		return nil, req.Context().Err()
	case resp := <-respCh:
		return resp, nil
	case err := <-errCh:
		return nil, err
	}
}

func isLocal(req *Request) bool {
	switch req.Host {
	case "127.0.0.1": // TODO(johanbrandhorst): Add more local hosts?
		return true
	}

	return false
}

// streamReader implements an io.ReadCloser wrapper for ReadableStream.
// See https://fetch.spec.whatwg.org/#readablestream for more information.
type streamReader struct {
	pending []byte
	stream  js.Value
	err     error // sticky read error
}

func (r *streamReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	if len(r.pending) == 0 {
		var (
			bCh   = make(chan []byte, 1)
			errCh = make(chan error, 1)
		)
		success := js.NewCallback(func(args []js.Value) {
			result := args[0]
			if result.Get("done").Bool() {
				errCh <- io.EOF
				return
			}
			value := make([]byte, result.Get("value").Get("byteLength").Int())
			js.ValueOf(value).Call("set", result.Get("value"))
			bCh <- value
		})
		defer success.Close()
		failure := js.NewCallback(func(args []js.Value) {
			// Assumes it's a TypeError. See
			// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/TypeError
			// for more information on this type. See
			// https://streams.spec.whatwg.org/#byob-reader-read for the spec on
			// the read method.
			errCh <- errors.New(args[0].Get("message").String())
		})
		defer failure.Close()
		r.stream.Call("read").Call("then", success, failure)
		select {
		case b := <-bCh:
			r.pending = b
		case err := <-errCh:
			r.err = err
			return 0, err
		}
	}
	n = copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

func (r *streamReader) Close() error {
	// This ignores any error returned from cancel method. So far, I did not encounter any concrete
	// situation where reporting the error is meaningful. Most users ignore error from resp.Body.Close().
	// If there's a need to report error here, it can be implemented and tested when that need comes up.
	r.stream.Call("cancel")
	return nil
}

// arrayReader implements an io.ReadCloser wrapper for ArrayBuffer.
// https://developer.mozilla.org/en-US/docs/Web/API/Body/arrayBuffer.
type arrayReader struct {
	arrayPromise js.Value
	pending      []byte
	read         bool
	err          error // sticky read error
}

func (r *arrayReader) Read(p []byte) (n int, err error) {
	if r.err != nil {
		return 0, r.err
	}
	if !r.read {
		r.read = true
		var (
			bCh   = make(chan []byte, 1)
			errCh = make(chan error, 1)
		)
		success := js.NewCallback(func(args []js.Value) {
			// Wrap the input ArrayBuffer with a Uint8Array
			uint8arrayWrapper := js.Global.Get("Uint8Array").New(args[0])
			value := make([]byte, uint8arrayWrapper.Get("byteLength").Int())
			js.ValueOf(value).Call("set", uint8arrayWrapper)
			bCh <- value
		})
		defer success.Close()
		failure := js.NewCallback(func(args []js.Value) {
			// Assumes it's a TypeError. See
			// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/TypeError
			// for more information on this type.
			// See https://fetch.spec.whatwg.org/#concept-body-consume-body for reasons this might error.
			errCh <- errors.New(args[0].Get("message").String())
		})
		defer failure.Close()
		r.arrayPromise.Call("then", success, failure)
		select {
		case b := <-bCh:
			r.pending = b
		case err := <-errCh:
			return 0, err
		}
	}
	if len(r.pending) == 0 {
		return 0, io.EOF
	}
	n = copy(p, r.pending)
	r.pending = r.pending[n:]
	return n, nil
}

func (r *arrayReader) Close() error {
	// This is a noop
	return nil
}
