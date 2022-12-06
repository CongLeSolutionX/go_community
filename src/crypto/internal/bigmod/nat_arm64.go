//go:build arm64 && gc && !purego

package bigmod

func montgomeryLoop(d []uint, a []uint, b []uint, m []uint, m0inv uint) uint
