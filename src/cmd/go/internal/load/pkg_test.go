package load

import (
	"cmd/go/internal/cfg"
	"testing"
)

func TestDefaultExecName(t *testing.T) {
	oldModulesEnabled := cfg.ModulesEnabled
	defer func() { cfg.ModulesEnabled = oldModulesEnabled }()
	for _, tt := range []struct {
		in         string
		wantMod    string
		wantGopath string
	}{
		{"example.com/mycmd", "mycmd", "mycmd"},
		{"example.com/mycmd/v0", "v0", "v0"},
		{"example.com/mycmd/v1", "v1", "v1"},
		{"example.com/mycmd/v2", "mycmd", "v2"}, // Semantic import versioning, use second last element in module mode.
		{"example.com/mycmd/v3", "mycmd", "v3"}, // Semantic import versioning, use second last element in module mode.
		{"mycmd", "mycmd", "mycmd"},
		{"mycmd/v0", "v0", "v0"},
		{"mycmd/v1", "v1", "v1"},
		{"mycmd/v2", "mycmd", "v2"}, // Semantic import versioning, use second last element in module mode.
		{"v0", "v0", "v0"},
		{"v1", "v1", "v1"},
		{"v2", "v2", "v2"},
	} {
		{
			cfg.ModulesEnabled = true
			gotMod := DefaultExecName(tt.in)
			if gotMod != tt.wantMod {
				t.Errorf("DefaultExecName(%q) in module mode = %v; want %v", tt.in, gotMod, tt.wantMod)
			}
		}
		{
			cfg.ModulesEnabled = false
			gotGopath := DefaultExecName(tt.in)
			if gotGopath != tt.wantGopath {
				t.Errorf("DefaultExecName(%q) in gopath mode = %v; want %v", tt.in, gotGopath, tt.wantGopath)
			}
		}
	}
}

func TestIsVersionElement(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		in   string
		want bool
	}{
		{"v0", false},
		{"v05", false},
		{"v1", false},
		{"v2", true},
		{"v3", true},
		{"v9", true},
		{"v10", true},
		{"v11", true},
		{"v", false},
		{"vx", false},
	} {
		got := isVersionElement(tt.in)
		if got != tt.want {
			t.Errorf("isVersionElement(%q) = %v; want %v", tt.in, got, tt.want)
		}
	}
}

func TestLoadPackagesAndErrors(t *testing.T) {
	ModInit = func() {}
	tests := []struct {
		name     string
		patterns []string
		want     int
	}{
		{
			name:     "package end with '.go",
			patterns: []string{"example.com/go-issue-32483/lib.go"},
			want:     1,
		},

		{
			name:     "normal packages",
			patterns: []string{"fmt", "testing"},
			want:     2,
		},

		{
			name: "go files",
			patterns: []string{
				"flag.go",
				"test.go",
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PackagesAndErrors(tt.patterns); len(got) != tt.want {
				t.Errorf("PackagesAndErrors() = %v, want %v packages", got, tt.want)
			}
		})
	}

}
