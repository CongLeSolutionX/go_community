// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The unfaketime command strips faketime headers from the stdout and stderr of
// a subprocess. It is intended to be used as an exec filter when diagnosing
// faketime deadlocks.
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		if err := filter(os.Stdout, os.Stdin); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	path := os.Args[1]
	args := os.Args[2:]

	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	os.Stdin.Close() // Let the parent process see EPIPE when the child closes stdin.

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc)
	go func() {
		for sig := range sigc {
			cmd.Process.Signal(sig)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := filter(os.Stdout, stdout); err != nil {
			log.Print(err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := filter(os.Stderr, stderr); err != nil {
			log.Print(err)
		}
	}()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.Exited() {
			os.Exit(ee.ExitCode())
		}
		log.Fatal(err)
	}
}

func filter(dst *os.File, src io.ReadCloser) error {
	defer src.Close()

	const magic = "\x00\x00PB"

	var prevT time.Time
	var t time.Time
	scanner := bufio.NewScanner(src)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) < len(magic) && !atEOF && bytes.HasPrefix([]byte(magic), data) {
			return 0, nil, nil
		}
		if !bytes.HasPrefix(data, []byte(magic)) {
			if atEOF && len(data) == 0 {
				return 0, nil, io.EOF
			}
			return 1, data[:1], nil
		}

		if len(data) < 16 {
			// Too short for a complete header.
			if atEOF {
				return len(data), data, io.ErrUnexpectedEOF
			}
			return 0, nil, nil
		}

		nanos := binary.BigEndian.Uint64(data[4:12])
		t = time.Unix(int64(nanos/1e9), int64(nanos%1e9))

		dlen := binary.BigEndian.Uint32(data[12:16])
		if int(dlen) < 0 {
			return len(data), data, errors.New("unfaketime: dlen overflow")
		}

		if len(data)-16 < int(dlen) {
			if atEOF {
				return len(data), data[16:], io.ErrUnexpectedEOF
			}
			return 0, nil, nil
		}
		return int(dlen) + 16, data[16 : 16+dlen], nil
	})

	for scanner.Scan() {
		if !t.Equal(prevT) {
			if _, err := fmt.Fprintln(dst, t.Format(time.RFC3339Nano)); err != nil {
				return err
			}
			prevT = t
		}
		if _, err := dst.Write(scanner.Bytes()); err != nil {
			return err
		}
	}
	return scanner.Err()
}
