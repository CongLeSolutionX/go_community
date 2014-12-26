#!/bin/sh
# NOTE: please update deployment target and -isysroot to match your setup.
export IPHONEOS_DEPLOYMENT_TARGET=5.1
# cmd/cgo doesn't support llvm-gcc-4.2, so we have to use clang.
exec clang -arch armv7 -isysroot /Applications/Xcode.app/Contents/Developer/Platforms/iPhoneOS.platform/Developer/SDKs/iPhoneOS6.1.sdk "$@"
