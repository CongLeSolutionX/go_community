// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	. "net/http"
	"testing"
)

func TestStatusType(t *testing.T) {
	code := StatusContinue
	if !IsInformational(code) {
		t.Errorf("Status %s is not 1xx: Informational", StatusText(code))
	}

	code = StatusCreated
	if IsInformational(code) {
		t.Errorf("Status %s is 1xx: Informational", StatusText(code))
	}

	code = StatusOK
	if !IsSuccess(code) {
		t.Errorf("Status %s is not 2xx: Success", StatusText(code))
	}

	code = StatusProcessing
	if IsSuccess(code) {
		t.Errorf("Status %s is 2xx: Success", StatusText(code))
	}

	code = StatusFound
	if !IsRedirection(code) {
		t.Errorf("Status %s is not 3xx: Redirection", StatusText(code))
	}

	code = StatusNotFound
	if IsRedirection(code) {
		t.Errorf("Status %s is 3xx: Redirection", StatusText(code))
	}

	code = StatusBadRequest
	if !IsClientError(code) {
		t.Errorf("Status %s is not 3xx: ClientError", StatusText(code))
	}

	code = StatusGatewayTimeout
	if IsClientError(code) {
		t.Errorf("Status %s is 3xx: ClientError", StatusText(code))
	}

	code = StatusInternalServerError
	if !IsServerError(code) {
		t.Errorf("Status %s is not 3xx: ServerError", StatusText(code))
	}

	code = StatusRequestURITooLong
	if IsServerError(code) {
		t.Errorf("Status %s is 3xx: ServerError", StatusText(code))
	}
}
