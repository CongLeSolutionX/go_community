// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sysinfo

import (
	"io"
	"os"
	"strings"
)

func readLinuxProcCPUInfo(buf []byte) error {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}

	return nil
}

func findCPUInfoField(buf []byte, fieldName string) string {
	filedValue := string(buf[:len(buf)])
	n := strings.Index(filedValue, fieldName)
	if n == -1 {
		return ""
	}

	filedValue = filedValue[n+len(fieldName):]
	if n := strings.Index(filedValue, "\n"); n != -1 {
		filedValue = filedValue[:n]
	}

	return filedValue
}
