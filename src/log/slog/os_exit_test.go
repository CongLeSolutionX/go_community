package slog

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"internal/testenv"
)

const os_exit_EnvVar = "SLOGTEST_LOG_OS_EXIT_TEST"

func TestSlogOsExit(t *testing.T) {
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
		{"fatal", func() { log.Fatal("fatal") }},
		{"fatalf", func() { log.Fatalf("fatalf") }},
		{"fatalln", func() { log.Fatalln("fatalln") }},
		{"panic", func() { log.Panic("panic") }},
		{"panicf", func() { log.Panicf("panicf") }},
		{"panicln", func() { log.Panicf("panicln") }},
		{"output", func() { log.Output(1, "output") }},
		{"default.fatal", func() { log.Default().Fatal("default.fatal") }},
		{"default.fatalf", func() { log.Default().Fatalf("default.fatalf") }},
		{"default.fatalln", func() { log.Default().Fatalln("default.fatalln") }},
		{"default.panic", func() { log.Default().Panic("default.panic") }},
		{"default.panicf", func() { log.Default().Panicf("default.panicf") }},
		{"default.panicln", func() { log.Default().Panicf("default.panicln") }},
		{"default.output", func() { log.Default().Output(1, "default.output") }},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer SetDefault(Default())
			SetDefault(New(
				NewTextHandler(os.Stderr, &HandlerOptions{
					AddSource: true,
					ReplaceAttr: func(groups []string, a Attr) Attr {
						if (a.Key == MessageKey || a.Key == SourceKey) && len(groups) == 0 {
							return a
						}
						return Attr{}
					},
				}),
			))

			// inside spawned test executable
			if os.Getenv(os_exit_EnvVar) == "1" {
				log.SetFlags(log.Lshortfile)
				tt.osExitCall()
				panic("expected os.Exit() to already have be called")
			}

			// spawn test executable
			var stderr bytes.Buffer
			cmd := testenv.Command(t, ep, "-test.run=TestSlogOsExit/"+tt.name+"$")
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

			lineNumberOffset := 33 // line number of the first test case
			expected := fmt.Sprintf(
				`source=:0 msg="os_exit_test.go:%d: %s"`,
				lineNumberOffset+i, tt.name,
			)
			if string(firstLine) != expected {
				t.Errorf("log filename missmatch: wanted %q, got %q", expected, firstLine)
			}
		})
	}
}
