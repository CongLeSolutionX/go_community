#!/usr/bin/env bash
# Copyright 2014 The Go Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# For testing Android.
# The compiler runs locally, then a copy of the GOROOT is pushed to a
# target device using adb, and the tests are run there.

set -e
ulimit -c 0 # no core files

if [ ! -f make.bash ]; then
	echo 'androidtest.bash must be run from $GOROOT/src' 1>&2
	exit 1
fi

if [ -z $GOOS ]; then
	export GOOS=android
fi
if [ "$GOOS" != "android" ]; then
	echo "androidtest.bash requires GOOS=android, got GOOS=$GOOS" 1>&2
	exit 1
fi

if [ -z $GOARM ]; then
	export GOARM=7
fi
if [ "$GOARM" != "7" ]; then
	echo "android only supports GOARM=7, got GOARM=$GOARM" 1>&2
	exit 1
fi

export CGO_ENABLED=1
unset GOBIN

export NDK=/Users/jbd/pkg/gomobile/android-ndk-r10e
export SYSROOT=$NDK/arm/sysroot
export CC_FOR_TARGET="/Users/jbd/pkg/gomobile/android-ndk-r10e/arm/bin/arm-linux-androideabi-gcc --sysroot=$SYSROOT"
export CXX_FOR_TARGET=/Users/jbd/pkg/gomobile/android-ndk-r10e/arm/bin/arm-linux-androideabi-g++
export GOOS=android
export GOARM=7
export GOARCH=arm


adb shell rm -rf /mnt/media_rw/goroot
adb shell rm -rf /data/media/tmp


# Do the build first, so we can build go_android_exec and cleaner.
# Also lets us fail early before the (slow) adb push if the build is broken.
. ./make.bash --no-banner
export GOROOT=$(dirname $(pwd))
export PATH=$GOROOT/bin:$PATH
GOOS=$GOHOSTOS GOARCH=$GOHOSTARCH go build \
	-o ../bin/go_android_${GOARCH}_exec \
	../misc/android/go_android_exec.go

export ANDROID_TEST_DIR=/tmp/androidtest-$$

function cleanup() {
	rm -rf ${ANDROID_TEST_DIR}
}

trap cleanup EXIT

# Push GOROOT to target device.
#
# The adb sync command will sync either the /system or /data
# directories of an android device from a similar directory
# on the host. We copy the files required for running tests under
# /data/local/tmp/goroot. The adb sync command does not follow
# symlinks so we have to copy.
export ANDROID_PRODUCT_OUT="${ANDROID_TEST_DIR}/out"
FAKE_GOROOT=$ANDROID_PRODUCT_OUT/mnt/media_rw/goroot
mkdir -p $FAKE_GOROOT/src
mkdir -p $FAKE_GOROOT/pkg
cp -a "${GOROOT}/src" "${FAKE_GOROOT}"
cp -a "${GOROOT}/test" "${FAKE_GOROOT}/"
cp -a "${GOROOT}/lib" "${FAKE_GOROOT}/"

echo $FAKE_GOROOT

# For android, the go tool will install the compiled package in
# pkg/android_${GOARCH}_shared directory by default, not in
# the usual pkg/${GOOS}_${GOARCH}. Some tests in src/go/* assume
# the compiled packages were installed in the usual places.
# Instead of reflecting this exception into the go/* packages,
# we copy the compiled packages into the usual places.
cp -a "${GOROOT}/pkg/android_${GOARCH}_shared" "${FAKE_GOROOT}/pkg/"
mv "${FAKE_GOROOT}/pkg/android_${GOARCH}_shared" "${FAKE_GOROOT}/pkg/android_${GOARCH}"

echo '# Syncing test files to android device'
adb shell mkdir -p /mnt/media_rw/goroot
time adb push ${FAKE_GOROOT} /mnt/media_rw/goroot

# export CLEANER=${ANDROID_TEST_DIR}/androidcleaner-$$
# cp ../misc/android/cleaner.go $CLEANER.go
# echo 'var files = `' >> $CLEANER.go
# (cd $ANDROID_PRODUCT_OUT/data/local/tmp/goroot; find . >> $CLEANER.go)
# echo '`' >> $CLEANER.go
# go build -o $CLEANER $CLEANER.go
# adb push $CLEANER /data/local/tmp/cleaner
# adb shell /data/local/tmp/cleaner

# Run standard tests.
bash run.bash --no-rebuild
