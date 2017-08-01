package sanitizers_test

import (
	"strings"
	"testing"
)

func TestMSAN(t *testing.T) {
	t.Parallel()
	requireOvercommit(t)
	config := configure("memory")
	config.skipIfCCompilerBroken(t)

	mustRun(t, config.goCmd("build", "std"))

	cases := []struct {
		src     string
		wantErr bool
	}{
		{src: "msan.go"},
		{src: "msan2.go"},
		{src: "msan2_cmsan.go"},
		{src: "msan3.go"},
		{src: "msan4.go"},
		{src: "msan5.go"},
		{src: "msan_fail.go", wantErr: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.src, func(t *testing.T) {
			t.Parallel()

			cmd := config.goCmd("run", srcPath(tc.src))

			if tc.wantErr {
				out, err := cmd.CombinedOutput()
				if err != nil {
					return
				}
				t.Fatalf("%#q exited without error; want MSAN failure\n%s", strings.Join(cmd.Args, " "), out)
			}

			mustRun(t, cmd)
		})
	}
}
