// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package auth provides access to user-provided authentication credentials.
package auth

import (
	"bufio"
	"cmd/go/internal/base"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os/exec"
	"strings"
)

// runAuthCommand executes a user provided GOAUTH command, parses its output, and
// stores credentials.
func runAuthCommand(cmdStr, prefix string) {
	cmdParts := strings.Fields(cmdStr)
	if len(cmdParts) == 0 {
		panic("GOAUTH invoked an empty authenticator command:" + cmdStr) // This should be caught earlier.
	}
	var allArgs []string
	cmdName := cmdParts[0]
	if len(cmdParts) > 1 {
		allArgs = cmdParts[1:]
	}
	if prefix != "" {
		allArgs = append(allArgs, prefix)
	}
	cmd := exec.Command(cmdName, allArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		base.Fatalf("could not run command %s: %v", cmdName, err)
	}
	if err := parseAuthOutput(stdout); err != nil {
		base.Fatalf("could not run command %s: %v", cmdName, err)
	}
	cmd.Wait()
}

// parseGoAuthOutput will parse the output from a GOAUTH command and
// update the credential cache.
// The expected format is:
// 1. one per line, one or more URL prefixes to which the credential should apply, beginning with the string https:// and ending at a complete path element.
// 2. an empty line.
// 3. one per line, zero or more request headers to set to apply the credential. (zero headers removes previous credentials from the cache.)
// 4. another empty line.
func parseAuthOutput(stdout io.ReadCloser) error {
	scanner := bufio.NewScanner(stdout)
	var prefixes []string
	var parsedPrefixes bool
	var headerBuilder strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "https://") {
			prefixes = append(prefixes, line)
			parsedPrefixes = true
		} else if len(strings.TrimSpace(line)) == 0 { // Encountered an empty line.
			if parsedPrefixes { // Skip the first new line.
				parsedPrefixes = false
				continue
			}
			err := processHeaders(prefixes, headerBuilder.String()) // A second empty line indicates the end of headers.
			if err != nil {
				return err
			}
			prefixes = nil // Reset prefixes for the new block.
			headerBuilder.Reset()
		} else {
			headerBuilder.WriteString(line)
			headerBuilder.WriteString("\r\n")
		}
	}
	if headerBuilder.Len() > 0 { // Strictly reject commands that do not follow the format.
		return fmt.Errorf("could not parse headers: missing new line")
	}
	return nil
}

// processHeaders parses headers from the GOAUTH command output, and stores
// them in the credential cache.
func processHeaders(prefixes []string, httpHeaders string) error {
	reader := bufio.NewReader(strings.NewReader(httpHeaders + "\r\n"))
	tp := textproto.NewReader(reader)
	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return fmt.Errorf("error parsing headers: %v", err)
	}
	httpHeader := http.Header(mimeHeader)
	storeCredential(prefixes, httpHeader)
	return nil
}
