package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestTooFewArgs(t *testing.T) {
	commands["test"] = command{
		requiredArgs: 1,
		handler: func(args [][]byte) ([][]byte, error) {
			if gotArgs := len(args); gotArgs != 1 {
				return nil, fmt.Errorf("expected 1 args, got %d", gotArgs)
			}
			return nil, nil
		},
	}

	var output bytes.Buffer
	err := processingLoop(mockRequest(t, "test", nil), bufio.NewWriter(&output))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	expectedErr := "expected 1 args, got 0"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %v", expectedErr, err)
	}
}

func TestTooManyArgs(t *testing.T) {
	commands["test"] = command{
		requiredArgs: 1,
		handler: func(args [][]byte) ([][]byte, error) {
			if gotArgs := len(args); gotArgs != 1 {
				return nil, fmt.Errorf("expected 1 args, got %d", gotArgs)
			}
			return nil, nil
		},
	}

	var output bytes.Buffer
	err := processingLoop(mockRequest(
		t, "test", [][]byte{[]byte("one"), []byte("two")}), bufio.NewWriter(&output))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	expectedErr := "expected 1 args, got 2"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got %v", expectedErr, err)
	}
}

func TestGetConfig(t *testing.T) {
	var output bytes.Buffer
	err := processingLoop(mockRequest(t, "getConfig", nil), bufio.NewWriter(&output))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	respArgs := readResponse(t, &output)
	if len(respArgs) != 1 {
		t.Fatalf("expected 1 response arg, got %d", len(respArgs))
	}

	expectedConfig, _ := json.Marshal(config)
	if string(respArgs[0]) != string(expectedConfig) {
		t.Errorf("expected config %s, got %s", expectedConfig, respArgs[0])
	}
}

func TestSha2256(t *testing.T) {
	testMessage := []byte("gophers eat grass")
	expectedDigest := []byte{
		0x67, 0x6f, 0x70, 0x68, 0x65, 0x72, 0x73, 0x20,
		0x65, 0x61, 0x74, 0x20, 0x67, 0x72, 0x61, 0x73,
		0x73, 0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c,
		0x14, 0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9,
		0x24, 0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93,
		0x4c, 0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8,
		0x55,
	}

	var output bytes.Buffer
	err := processingLoop(mockRequest(t, "SHA2-256", [][]byte{testMessage}), bufio.NewWriter(&output))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	respArgs := readResponse(t, &output)
	if len(respArgs) != 1 {
		t.Fatalf("expected 1 response arg, got %d", len(respArgs))
	}

	if !bytes.Equal(respArgs[0], expectedDigest) {
		t.Errorf("expected digest %x, got %x", expectedDigest, respArgs[0])
	}
}

func mockRequest(t *testing.T, cmd string, args [][]byte) io.Reader {
	t.Helper()

	msgData := append([][]byte{[]byte(cmd)}, args...)

	var buf bytes.Buffer
	if err := writeResponse(bufio.NewWriter(&buf), msgData); err != nil {
		t.Fatalf("writeResponse error: %v", err)
	}

	return &buf
}

func readResponse(t *testing.T, reader io.Reader) [][]byte {
	var numArgs uint32
	if err := binary.Read(reader, binary.LittleEndian, &numArgs); err != nil {
		t.Fatalf("failed to read response args count: %v", err)
	}

	args, err := readArgs(reader, numArgs)
	if err != nil {
		t.Fatalf("failed to read %d response args: %v", numArgs, err)
	}

	return args
}
