package log

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"internal/testenv"
)

const os_exit_EnvVar = "LOGTEST_LOG_OS_EXIT_TEST"

func TestLogOsExit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	testenv.MustHaveExec(t)
	ep, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable failed: %v", err)
	}

	tests := []struct {
		name       string
		osExitCall func()
	}{
		{"fatal", func() { Fatal("fatal") }},
		{"fatalf", func() { Fatalf("fatalf") }},
		{"fatalln", func() { Fatalln("fatalln") }},
		{"panic", func() { Panic("panic") }},
		{"panicf", func() { Panicf("panicf") }},
		{"panicln", func() { Panicf("panicln") }},
		{"output", func() { Output(1, "output") }},
		{"default.fatal", func() { Default().Fatal("default.fatal") }},
		{"default.fatalf", func() { Default().Fatalf("default.fatalf") }},
		{"default.fatalln", func() { Default().Fatalln("default.fatalln") }},
		{"default.panic", func() { Default().Panic("default.panic") }},
		{"default.panicf", func() { Default().Panicf("default.panicf") }},
		{"default.panicln", func() { Default().Panicf("default.panicln") }},
		{"default.output", func() { Default().Output(1, "default.output") }},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// inside spawned test executable
			if os.Getenv(os_exit_EnvVar) == "1" {
				SetFlags(Lshortfile)
				tt.osExitCall()
				panic("expected os.Exit() to already have be called")
			}

			// spawn test executable
			var stderr bytes.Buffer
			cmd := testenv.Command(t, ep, "-test.run=TestLogOsExit/"+tt.name+"$")
			cmd.Env = append(cmd.Environ(), os_exit_EnvVar+"=1")
			cmd.Stderr = &stderr

			var exitErr *exec.ExitError
			if err := cmd.Run(); !errors.As(err, &exitErr) {
				t.Fatalf("exec(self) expected exec.ExitError: %v", err)
			}

			_, firstLine, err := bufio.ScanLines(stderr.Bytes(), true)
			if err != nil {
				t.Fatalf("exec(self) failed to split line: %v", err)
			}

			lineNumberOffset := 32 // line number of the first test case
			expected := fmt.Sprintf("os_exit_test.go:%d: %s", lineNumberOffset+i, tt.name)
			if string(firstLine) != expected {
				t.Errorf("log filename missmatch: wanted %q, got %q", expected, firstLine)
			}
		})
	}
}
