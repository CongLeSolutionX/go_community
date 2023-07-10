// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux && (386 || amd64)

package sysinfo

const (
	// cpuinfo filed name
	ModelName = "model name\t: "
)

func osCpuInfoName() string {
	buf := make([]byte, 512)
	err := readLinuxProcCPUInfo(buf)
	if err != nil {
		return ""
	}

	return findCPUInfoField(buf, ModelName)
}
