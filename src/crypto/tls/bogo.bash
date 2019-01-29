#!/usr/bin/env bash
# Copyright 2019 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# This script will run a pinned version of the BoringSSL test suite ("BoGo")
# against the crypto/tls package.

BORINGSSL_COMMIT="8cbb5f8f204d188211b340412cc0dd4d5b09649a"

set -e

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
export GOROOT="$PWD"

if [ ! -d test/boringssl ]; then
    git clone -f https://boringssl.googlesource.com/boringssl test/boringssl
fi
(cd test/boringssl && git fetch && git checkout $BORINGSSL_COMMIT)

bin/go test -c -o bin/crypto-tls-bogo-shim crypto/tls
cd test/boringssl/ssl/test/runner && "$GOROOT/bin/go" test \
    -shim-path "$GOROOT/bin/crypto-tls-bogo-shim" \
    -shim-config "$GOROOT/src/crypto/tls/bogo_shim.json" \
    -allow-unimplemented -loose-errors "$@"
