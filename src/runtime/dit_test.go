package runtime_test

import (
	"internal/cpu"
	"testing"
	"runtime"
)

func TestWithDataIndependentTiming(t *testing.T) {
	if !cpu.ARM64.HasDIT {
		t.Skipf("DIT is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	runtime.WithDataIndependentTiming(func() {
		if !runtime.DataIndependentTimingEnabled() {
			t.Fatal("dit not enabled within WithDataIndependentTiming closure")
		}

		runtime.WithDataIndependentTiming(func() {
			if !runtime.DataIndependentTimingEnabled() {
				t.Fatal("dit not enabled within nested WithDataIndependentTiming closure")
			}
		})

		if !runtime.DataIndependentTimingEnabled() {
			t.Fatal("dit not enabled after return from nested WithDataIndependentTiming closure")
		}
	})
	
	if runtime.DataIndependentTimingEnabled() {
		t.Fatal("dit not unset after returning from WithDataIndependentTiming closure")
	}
}
