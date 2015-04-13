// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// detect attempts to autodetect the correct
// values of the environment variables
// used by go_darwin_arm_exec.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func main() {
	devID, err := detectDevID()
	check(err)
	fmt.Printf("export GO_IOS_DEV_ID=%s\n", devID)

	udid, err := detectUDID()
	check(err)

	mp, err := detectMobileProvisionFile(udid)
	check(err)

	f, err := ioutil.TempFile("", "go_ios_detect_")
	check(err)
	fname := f.Name()
	defer os.Remove(fname)

	out, err := parseMobileProvision(mp).CombinedOutput()
	check(err)
	_, err = f.Write(out)
	check(err)
	check(f.Close())

	appID, err := plistExtract(fname, "ApplicationIdentifierPrefix:0")
	check(err)
	fmt.Printf("export GO_IOS_APP_ID=%s\n", appID)

	teamID, err := plistExtract(fname, "Entitlements:com.apple.developer.team-identifier")
	check(err)
	fmt.Printf("export GO_IOS_TEAM_ID=%s\n", teamID)
}

func detectDevID() (string, error) {
	cmd := exec.Command("security", "find-identity", "-p", "codesigning", "-v")
	lines, err := getLines(cmd)
	if err != nil {
		return "", err
	}

	for _, line := range lines {
		if !bytes.Contains(line, []byte("iPhone Developer")) {
			continue
		}
		fields := bytes.Fields(line)
		return string(fields[1]), nil
	}
	return "", errors.New("no code signing identity found")
}

var udidPrefix = []byte("UniqueDeviceID: ")

func detectUDID() ([]byte, error) {
	cmd := exec.Command("ideviceinfo")
	lines, err := getLines(cmd)
	if err != nil {
		return nil, err
	}
	for _, line := range lines {
		if bytes.HasPrefix(line, udidPrefix) {
			return bytes.TrimPrefix(line, udidPrefix), nil
		}
	}
	return nil, errors.New("udid not found; is the device connected?")
}

func detectMobileProvisionFile(udid []byte) (string, error) {
	cmd := exec.Command("mdfind", "-name", ".mobileprovision")
	lines, err := getLines(cmd)
	if err != nil {
		return "", err
	}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		xmlLines, err := getLines(parseMobileProvision(string(line)))
		if err != nil {
			return "", err
		}
		for _, xmlLine := range xmlLines {
			if bytes.Contains(xmlLine, udid) {
				return string(line), nil
			}
		}
	}
	return "", fmt.Errorf("did not find mobile provision matching device udid %s", udid)
}

func parseMobileProvision(fname string) *exec.Cmd {
	return exec.Command("security", "cms", "-D", "-i", string(fname))
}

func plistExtract(fname string, path string) ([]byte, error) {
	out, err := exec.Command("/usr/libexec/PlistBuddy", "-c", "Print "+path, fname).CombinedOutput()
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(out), nil
}

func getLines(cmd *exec.Cmd) ([][]byte, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return bytes.Split(out, []byte("\n")), nil
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
