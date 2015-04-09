#!/usr/bin/env bash
# Copyright 2015 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

if [ "$(go env GOOS)" == "linux" ]; then
	GOPATH=$(pwd) go build src/signal.go
	./signal
fi