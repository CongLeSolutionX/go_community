package get

import (
	"cmd/go/internal/load"
	"testing"
)

func Test_runGet(t *testing.T) {
	load.ModInit = func() {} // set module init to no-op

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "cmd/go/internal/load",
			args: []string{"cmd/go/internal/load"},
		},

		{
			name: "pkg end with '.go'",
			args: []string{"github.com/letientai299/go-issue-32483/lib.go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runGet(nil, tt.args)
		})
	}
}
