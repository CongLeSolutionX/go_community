package sanitizers_test

import (
	"testing"
)

func TestTSAN(t *testing.T) {
	t.Parallel()
	requireOvercommit(t)
	config := configure("thread")
	config.skipIfCCompilerBroken(t)

	mustRun(t, config.goCmd("build", "std"))

	cases := []struct {
		src          string
		needsRuntime bool
	}{
		{src: "tsan.go"},
		{src: "tsan2.go"},
		{src: "tsan3.go"},
		{src: "tsan4.go"},
		{src: "tsan5.go", needsRuntime: true},
		{src: "tsan6.go", needsRuntime: true},
		{src: "tsan7.go", needsRuntime: true},
		{src: "tsan8.go"},
		{src: "tsan9.go"},
		{src: "tsan10.go", needsRuntime: true},
		{src: "tsan11.go", needsRuntime: true},
		{src: "tsan12.go", needsRuntime: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.src, func(t *testing.T) {
			t.Parallel()

			cmd := config.goCmd("run", srcPath(tc.src))
			if tc.needsRuntime {
				config.skipIfRuntimeIncompatible(t)
			}
			mustRun(t, cmd)
		})
	}
}
