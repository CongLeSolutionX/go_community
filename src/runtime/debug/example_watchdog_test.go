// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug_test

import (
	"io"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
)

// Example_watchdog shows an example of using debug.SetCrashOutput to
// direct crashes to a "watchdog" process, for automated crash
// reporting. The watchdog is the same executable, invoked with
// WATCHDOG=1 in the environment.
func Example_watchdog() {
	main()

	// This Example doesn't actually run as a test because its
	// purpose is to crash, so the Output comment below is
	// commented out.
	//
	// To observe the watchdog in action, uncomment the Output
	// comment below and run this command:
	//
	//    $ go test -run=Example_watchdog runtime/debug
	//    panic: oops
	//    ...stack...
	//    watchdog: saved crash report at /tmp/10804884239807998216.crash

	// // Output:
}

func main() {
	watchdog()

	// Run the application.
	println("hello")
	panic("oops")
}

// watchdog starts the watchdog process, which performs automated
// crash reporting. Call this function immediately within main.
//
// This function re-executes the same executable as a child process,
// in a special mode. In that mode, the call to watchdog will never
// return.
func watchdog() {
	if os.Getenv("WATCHDOG") != "" {
		// This is the watchdog (child) process.
		log.SetFlags(0)
		log.SetPrefix("watchdog: ")

		log.Println("woof")
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("failed to read from input pipe: %v", err)
		}

		// Save the crash report securely in the file system.
		f, err := os.CreateTemp("", "*.crash")
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
		log.Fatalf("saved crash report at %s", f.Name())
	}

	// This is the application process.
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=Example_watchdog")
	cmd.Env = append(os.Environ(), "WATCHDOG=1")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr
	pipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("StdinPipe: %v", err)
	}
	debug.SetCrashOutput(pipe.(*os.File)) // (this conversion is safe)
	if err := cmd.Start(); err != nil {
		log.Fatalf("can't start watchdog: %v", err)
	}
}
