// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import (
	"net/http"
	"testing"
)

func TestStatusType(t *testing.T) {
	code := http.StatusContinue
	if !IsInformational(code) {
		t.Errorf("Status %s is not 1xx: Informational", StatusText(code))
	}

	code = http.StatusCreated
	if IsInformational(code) {
		t.Errorf("Status %s is 1xx: Informational", StatusText(code))
	}

	code = http.StatusOK
	if !IsSuccess(code) {
		t.Errorf("Status %s is not 2xx: Success", StatusText(code))
	}

	code = http.StatusProcessing
	if IsSuccess(code) {
		t.Errorf("Status %s is 2xx: Success", StatusText(code))
	}

	code = http.StatusFound
	if !IsRedirection(code) {
		t.Errorf("Status %s is not 3xx: Redirection", StatusText(code))
	}

	code = http.StatusNotFound
	if IsRedirection(code) {
		t.Errorf("Status %s is 3xx: Redirection", StatusText(code))
	}

	code = http.StatusBadRequest
	if !IsClientError(code) {
		t.Errorf("Status %s is not 4xx: ClientError", StatusText(code))
	}

	code = http.StatusGatewayTimeout
	if IsClientError(code) {
		t.Errorf("Status %s is 4xx: ClientError", StatusText(code))
	}

	code = http.StatusInternalServerError
	if !IsServerError(code) {
		t.Errorf("Status %s is not 5xx: ServerError", StatusText(code))
	}

	code = http.StatusRequestURITooLong
	if IsServerError(code) {
		t.Errorf("Status %s is 5xx: ServerError", StatusText(code))
	}
}
