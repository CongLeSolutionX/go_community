package sanitizers_test

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"
)

func TestShared(t *testing.T) {
	t.Parallel()
	requireOvercommit(t)

	GOOS, err := goEnv("GOOS")
	if err != nil {
		t.Fatal(err)
	}
	libExt := "so"
	if GOOS == "darwin" {
		libExt = "dylib"
	}

	cases := []struct {
		sanitizer string
		srcFile   string
	}{
		{
			sanitizer: "memory",
			srcFile:   "msan_shared.go",
		},
		{
			sanitizer: "thread",
			srcFile:   "tsan_shared.go",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.sanitizer, func(t *testing.T) {
			t.Parallel()
			config := configure(tc.sanitizer)
			config.skipIfCCompilerBroken(t)

			dir := newTempDir(t)
			defer dir.RemoveAll(t)

			lib := dir.Join(fmt.Sprintf("lib%s.%s", tc.sanitizer, libExt))

			cmd := config.goCmd("build", "-buildmode=c-shared", "-o", lib, srcPath(tc.srcFile))
			mustRun(t, cmd)

			cSrc := dir.Join("main.c")
			if err := ioutil.WriteFile(cSrc, cMain, 0600); err != nil {
				t.Fatalf("failed to write C source file: %v", err)
			}

			dstBin := dir.Join("test" + tc.sanitizer)
			cmd, err := cc(config.cFlags...)
			if err != nil {
				t.Fatal(err)
			}
			cmd.Args = append(cmd.Args, config.ldFlags...)
			cmd.Args = append(cmd.Args, "-o", dstBin, cSrc, lib)
			mustRun(t, cmd)

			cmd = exec.Command(dstBin)
			replaceEnv(cmd, "LD_LIBRARY_PATH", ".")
			mustRun(t, cmd)
		})
	}
}
