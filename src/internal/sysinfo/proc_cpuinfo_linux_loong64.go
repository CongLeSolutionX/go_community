// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sysinfo

const (
	// cpuinfo filed name
	ModelName = "\nModel Name\t\t: "
	CPUMHz    = "\nCPU MHz\t\t\t: "
)

func osCpuInfoName() string {
	// The 512-byte buffer is enough to hold the contents of CPU0
	buf := make([]byte, 512)
	err := readLinuxProcCPUInfo(buf)
	if err != nil {
		return ""
	}

	modelName := findCPUInfoField(buf, ModelName)
	cpuMHz := findCPUInfoField(buf, CPUMHz)

	if modelName == "" {
		return ""
	}

	if cpuMHz == "" {
		return modelName
	}

	return modelName + " @ " + cpuMHz + "MHz"
}
