// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

// A module wrapper adapting the Go FIPS module to the protocol used by the
// BoringSSL project's `acvptool`.
//
// The `acvptool` "lowers" the NIST ACVP server JSON test vectors into a simpler
// stdin/stdout protocol that can be implemented by a module shim. The tool
// will fork this binary, request the supported configuration, and then provide
// test cases over stdin, expecting results to be returned on stdout.
//
// See "Testing other FIPS modules"[0] from the BoringSSL ACVP.md documentation
// for a more detailed description of the protocol used between the acvptool
// and module wrappers.
//
// Example usage:
// ```
//
//	$> # From Go src (note: must specify full path to side-step build ignore)
//	$> go build ./crypto/internal/acvpwrapper/acvpwrapper.go
//
//	$> # From BoringSSL root:
//	$> cd util/fipstools/acvp/acvptool/
//	$> go build ./
//	$> cd test
//	$> go run check_expected.go -tool ../acvptool -module-wrappers go:/path/to/go/acvpwrapper -tests tests.min.json
//
// ```
//
// Example tests.min.json:
// ```json
// [
//
//	{"Wrapper": "go", "In": "vectors/SHA2-224.bz2", "Out": "expected/SHA2-224.bz2"},
//	{"Wrapper": "go", "In": "vectors/SHA2-256.bz2", "Out": "expected/SHA2-256.bz2"},
//	{"Wrapper": "go", "In": "vectors/SHA2-384.bz2", "Out": "expected/SHA2-384.bz2"},
//	{"Wrapper": "go", "In": "vectors/SHA2-512.bz2", "Out": "expected/SHA2-512.bz2"}
//
// ]
// ```
//
// [0]:https://boringssl.googlesource.com/boringssl/+/refs/heads/master/util/fipstools/acvp/ACVP.md#testing-other-fips-modules
package main

import (
	"bufio"
	"crypto/internal/fips/sha256"
	"crypto/internal/fips/sha512"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

func main() {
	writer := bufio.NewWriter(os.Stdout)
	defer func() { _ = writer.Flush() }()

	if err := processingLoop(bufio.NewReader(os.Stdin), writer); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "processing error: %v\n", err)
		os.Exit(1)
	}
}

type request struct {
	name string
	args [][]byte
}

type commandHandler func([][]byte) ([][]byte, error)

type command struct {
	// requiredArgs enforces that an exact number of arguments are provided to the handler.
	requiredArgs int
	handler      commandHandler
}

var (
	// configuration returned from getConfig command. This data represents the algorithms
	// our module supports and is used to determine which test cases are applicable.
	//
	// See https://pages.nist.gov/ACVP/draft-celi-acvp-sha.html#section-7
	config = []interface{}{
		// HASH algorithm capabilities
		// See https://pages.nist.gov/ACVP/draft-celi-acvp-sha.html#section-7.2
		hashCapability("SHA2-224"),
		hashCapability("SHA2-256"),
		hashCapability("SHA2-384"),
		hashCapability("SHA2-512"),
		hashCapability("SHA2-512/256"),
	}

	// commands should reflect what config says we support. E.g. adding a command here will be a NOP
	// unless the configuration indicates the command's associated algorithm is supported.
	commands = map[string]command{
		"getConfig":        cmdGetConfig(),
		"SHA2-224":         cmdHashAft(sha256.New224()),
		"SHA2-224/MCT":     cmdHashMct(sha256.New224()),
		"SHA2-256":         cmdHashAft(sha256.New()),
		"SHA2-256/MCT":     cmdHashMct(sha256.New()),
		"SHA2-384":         cmdHashAft(sha512.New384()),
		"SHA2-384/MCT":     cmdHashMct(sha512.New384()),
		"SHA2-512":         cmdHashAft(sha512.New()),
		"SHA2-512/MCT":     cmdHashMct(sha512.New()),
		"SHA2-512/256":     cmdHashAft(sha512.New512_256()),
		"SHA2-512/256/MCT": cmdHashMct(sha512.New512_256()),
	}
)

func hashCapability(algName string) map[string]interface{} {
	return map[string]interface{}{
		"algorithm": algName,
		"revision":  "1.0",
		// Matching BSSL's config:
		"messageLength": []map[string]int{{
			"min": 0, "max": 65528, "increment": 8,
		}},
	}
}

func processingLoop(reader io.Reader, writer *bufio.Writer) error {
	// Per ACVP.md:
	//   The protocol is request–response: the subprocess only speaks in response to a request
	//   and there is exactly one response for every request.
	for {
		req, err := readRequest(reader)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("reading request: %w", err)
		}

		cmd, exists := commands[req.name]
		if !exists {
			return fmt.Errorf("unknown command: %q", req.name)
		}

		if gotArgs := len(req.args); gotArgs != cmd.requiredArgs {
			return fmt.Errorf("command %q expected %d args, got %d", req.name, cmd.requiredArgs, gotArgs)
		}

		response, err := cmd.handler(req.args)
		if err != nil {
			return fmt.Errorf("command %q failed: %w", req.name, err)
		}

		if err = writeResponse(writer, response); err != nil {
			return fmt.Errorf("command %q response failed: %w", req.name, err)
		}
	}

	return nil
}

func readRequest(reader io.Reader) (*request, error) {
	// Per ACVP.md:
	//   Requests consist of one or more byte strings and responses consist
	//   of zero or more byte strings. A request contains: the number of byte
	//   strings, the length of each byte string, and the contents of each byte
	//   string. All numbers are 32-bit little-endian and values are
	//   concatenated in the order specified.
	var numArgs uint32
	if err := binary.Read(reader, binary.LittleEndian, &numArgs); err != nil {
		return nil, err
	}
	if numArgs == 0 {
		return nil, errors.New("invalid request: zero args")
	}

	args, err := readArgs(reader, numArgs)
	if err != nil {
		return nil, err
	}

	return &request{
		name: string(args[0]),
		args: args[1:],
	}, nil
}

func readArgs(reader io.Reader, requiredArgs uint32) ([][]byte, error) {
	argLengths := make([]uint32, requiredArgs)
	args := make([][]byte, requiredArgs)

	for i := range argLengths {
		if err := binary.Read(reader, binary.LittleEndian, &argLengths[i]); err != nil {
			return nil, fmt.Errorf("invalid request: failed to read %d-th arg len: %w", i, err)
		}
	}

	for i, length := range argLengths {
		buf := make([]byte, length)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return nil, fmt.Errorf("invalid request: failed to read %d-th arg data: %w", i, err)
		}
		args[i] = buf
	}

	return args, nil
}

func writeResponse(writer *bufio.Writer, args [][]byte) error {
	// See `readRequest` for details on the base format. Per ACVP.md:
	//   A response has the same format except that there may be zero byte strings
	//   and the first byte string has no special meaning.
	numArgs := uint32(len(args))
	if err := binary.Write(writer, binary.LittleEndian, numArgs); err != nil {
		return fmt.Errorf("writing arg count: %w", err)
	}

	for i, arg := range args {
		if err := binary.Write(writer, binary.LittleEndian, uint32(len(arg))); err != nil {
			return fmt.Errorf("writing %d-th arg length: %w", i, err)
		}
	}

	for i, b := range args {
		if _, err := writer.Write(b); err != nil {
			return fmt.Errorf("writing %d-th arg data: %w", i, err)
		}
	}

	return writer.Flush()
}

// "All implementations must support the getConfig command
// which takes no arguments and returns a single byte string
// which is a JSON blob of ACVP algorithm configuration."
func cmdGetConfig() command {
	return command{
		handler: func(args [][]byte) ([][]byte, error) {
			configJSON, err := json.Marshal(config)
			if err != nil {
				return nil, err
			}
			return [][]byte{configJSON}, nil
		},
	}
}

// cmdHashAft returns a command handler for the specified hash
// algorithm for algorithm functional test (AFT) test cases.
//
// This shape of command expects a message as the sole argument,
// and writes the resulting digest as a response.
//
// See https://pages.nist.gov/ACVP/draft-celi-acvp-sha.html
func cmdHashAft(h hash.Hash) command {
	return command{
		requiredArgs: 1, // Message to hash.
		handler: func(args [][]byte) ([][]byte, error) {
			h.Reset()
			h.Write(args[0])
			digest := make([]byte, 0, h.Size())
			digest = h.Sum(digest)

			return [][]byte{digest}, nil
		},
	}
}

// cmdHashAft returns a command handler for the specified hash
// algorithm for monte carlo test (MCT) test cases.
//
// This shape of command expects a seed as the sole argument,
// and writes the resulting digest as a response.
//
// This algorithm was ported from `HashMCT` in BSSL's `modulewrapper.cc`
// and is not an exact match to the NIST MCT[0] algorithm due to
// footnote #1 in the ACVP.md docs[1].
//
// [0]: https://pages.nist.gov/ACVP/draft-celi-acvp-sha.html#section-6.2
// [1]: https://boringssl.googlesource.com/boringssl/+/refs/heads/master/util/fipstools/acvp/ACVP.md#testing-other-fips-modules
func cmdHashMct(h hash.Hash) command {
	return command{
		requiredArgs: 1, // Seed message.
		handler: func(args [][]byte) ([][]byte, error) {
			hSize := h.Size()
			seed := args[0]

			if seedLen := len(seed); seedLen != hSize {
				return nil, fmt.Errorf("invalid seed size: expected %d got %d", hSize, seedLen)
			}

			digest := make([]byte, 0, hSize)
			buf := make([]byte, 0, 3*hSize)
			buf = append(buf, seed...)
			buf = append(buf, seed...)
			buf = append(buf, seed...)

			for i := 0; i < 1000; i++ {
				h.Reset()
				h.Write(buf)
				digest = h.Sum(digest)
				h.Sum(digest[:0])

				copy(buf, buf[hSize:])
				copy(buf[2*hSize:], digest)
			}

			return [][]byte{buf[hSize*2:]}, nil
		},
	}
}
