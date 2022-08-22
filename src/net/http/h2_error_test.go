// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !nethttpomithttp2
// +build !nethttpomithttp2

package http

import (
	"errors"
	"testing"
)

func TestStreamError(t *testing.T) {
	streamErr := http2streamError(42, http2ErrCodeProtocol)
	t.Run("vendored", func(t *testing.T) {
		var target http2StreamError
		ok := errors.As(streamErr, &target)
		if !ok {
			t.Fatalf("errors.As failed")
		}
		if target.StreamID != streamErr.StreamID {
			t.Errorf("got StreamID %v, expected %v", target.StreamID, streamErr.StreamID)
		}
		if target.Cause != streamErr.Cause {
			t.Errorf("got Cause %v, expected %v", target.Cause, streamErr.Cause)
		}
		if target.Code != streamErr.Code {
			t.Errorf("got Code %v, expected %v", target.Code, streamErr.Code)
		}
	})
	t.Run("external", func(t *testing.T) {
		type externalCode uint32
		type externalStreamError struct {
			StreamID uint32
			Cause    error
			Code     externalCode
		}
		var target externalStreamError
		ok := asHTTP2StreamError(streamErr, &target, "net/http", "externalStreamError")
		if !ok {
			t.Fatalf("errors.As failed")
		}
		if target.StreamID != streamErr.StreamID {
			t.Errorf("got StreamID %v, expected %v", target.StreamID, streamErr.StreamID)
		}
		if target.Cause != streamErr.Cause {
			t.Errorf("got Cause %v, expected %v", target.Cause, streamErr.Cause)
		}
		if uint32(target.Code) != uint32(streamErr.Code) {
			t.Errorf("got Code %v, expected %v", target.Code, streamErr.Code)
		}
	})
}
