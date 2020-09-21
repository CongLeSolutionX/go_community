// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd js,wasm !android,linux netbsd openbsd solaris
// +build !cgo osusergo

package user

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const groupFile = "/etc/group"
const userFile = "/etc/passwd"

var colon = []byte{':'}

func init() {
	groupImplemented = false
}

// lineFunc returns a value, an error, or (nil, nil) to skip the row.
type lineFunc func(line []byte) (v interface{}, err error)

// readColonFile parses r as an /etc/group or /etc/passwd style file, running
// fn for each row. readColonFile returns a value, an error, or (nil, nil) if
// the end of the file is reached without a match.
func readColonFile(r io.Reader, fn lineFunc) (v interface{}, err error) {
	bs := bufio.NewScanner(r)
	for bs.Scan() {
		line := bs.Bytes()
		// There's no spec for /etc/passwd or /etc/group, but we try to follow
		// the same rules as the glibc parser, which allows comments and blank
		// space at the beginning of a line.
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		v, err = fn(line)
		if v != nil || err != nil {
			return
		}
	}
	return nil, bs.Err()
}

func matchGroupIndexValue(value string, idx int) lineFunc {
	var leadColon string
	if idx > 0 {
		leadColon = ":"
	}
	substr := []byte(leadColon + value + ":")
	return func(line []byte) (v interface{}, err error) {
		if !bytes.Contains(line, substr) || bytes.Count(line, colon) < 3 {
			return
		}
		// wheel:*:0:root
		parts := strings.SplitN(string(line), ":", 4)
		if len(parts) < 4 || parts[0] == "" || parts[idx] != value ||
			// If the file contains +foo and you search for "foo", glibc
			// returns an "invalid argument" error. Similarly, if you search
			// for a gid for a row where the group name starts with "+" or "-",
			// glibc fails to find the record.
			parts[0][0] == '+' || parts[0][0] == '-' {
			return
		}
		if _, err := strconv.Atoi(parts[2]); err != nil {
			return nil, nil
		}
		return &Group{Name: parts[0], Gid: parts[2]}, nil
	}
}

func findGroupId(id string, r io.Reader) (*Group, error) {
	if v, err := readColonFile(r, matchGroupIndexValue(id, 2)); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*Group), nil
	}
	return nil, UnknownGroupIdError(id)
}

func findGroupName(name string, r io.Reader) (*Group, error) {
	if v, err := readColonFile(r, matchGroupIndexValue(name, 0)); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*Group), nil
	}
	return nil, UnknownGroupError(name)
}

// returns a *User for a row if that row's has the given value at the
// given index.
func matchUserIndexValue(value string, idx int) lineFunc {
	var leadColon string
	if idx > 0 {
		leadColon = ":"
	}
	substr := []byte(leadColon + value + ":")
	return func(line []byte) (v interface{}, err error) {
		if !bytes.Contains(line, substr) || bytes.Count(line, colon) < 6 {
			return
		}
		// kevin:x:1005:1006::/home/kevin:/usr/bin/zsh
		parts := strings.SplitN(string(line), ":", 7)
		if len(parts) < 6 || parts[idx] != value || parts[0] == "" ||
			parts[0][0] == '+' || parts[0][0] == '-' {
			return
		}
		if _, err := strconv.Atoi(parts[2]); err != nil {
			return nil, nil
		}
		if _, err := strconv.Atoi(parts[3]); err != nil {
			return nil, nil
		}
		u := &User{
			Username: parts[0],
			Uid:      parts[2],
			Gid:      parts[3],
			Name:     parts[4],
			HomeDir:  parts[5],
		}
		// The pw_gecos field isn't quite standardized. Some docs
		// say: "It is expected to be a comma separated list of
		// personal data where the first item is the full name of the
		// user."
		if i := strings.Index(u.Name, ","); i >= 0 {
			u.Name = u.Name[:i]
		}
		return u, nil
	}
}

func findUserId(uid string, r io.Reader) (*User, error) {
	i, e := strconv.Atoi(uid)
	if e != nil {
		return nil, errors.New("user: invalid userid " + uid)
	}
	if v, err := readColonFile(r, matchUserIndexValue(uid, 2)); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*User), nil
	}
	return nil, UnknownUserIdError(i)
}

func findUsername(name string, r io.Reader) (*User, error) {
	if v, err := readColonFile(r, matchUserIndexValue(name, 0)); err != nil {
		return nil, err
	} else if v != nil {
		return v.(*User), nil
	}
	return nil, UnknownUserError(name)
}

func lookupGroup(groupname string) (*Group, error) {
	if g, err := defaultUserdbClient.lookupGroup(groupname); err != nil {
		var connectError *errConnectUserdb
		if !errors.As(err, &connectError) {
			// systemd userdb is available per se, but an error occurred:
			return nil, err
		}
		// fallthrough to parsing files ourselves:
	} else {
		return g, nil
	}

	f, err := os.Open(groupFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findGroupName(groupname, f)
}

func lookupGroupId(id string) (*Group, error) {
	if g, err := defaultUserdbClient.lookupGroupId(id); err != nil {
		var connectError *errConnectUserdb
		if !errors.As(err, &connectError) {
			// systemd userdb is available per se, but an error occurred:
			return nil, err
		}
		// fallthrough to parsing files ourselves:
	} else {
		return g, nil
	}

	f, err := os.Open(groupFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findGroupId(id, f)
}

func lookupUser(username string) (*User, error) {
	if u, err := defaultUserdbClient.lookupUser(username); err != nil {
		var connectError *errConnectUserdb
		if !errors.As(err, &connectError) {
			// systemd userdb is available per se, but an error occurred:
			return nil, err
		}
		// fallthrough to parsing files ourselves:
	} else {
		return u, nil
	}

	f, err := os.Open(userFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findUsername(username, f)
}

func lookupUserId(uid string) (*User, error) {
	if u, err := defaultUserdbClient.lookupUserId(uid); err != nil {
		var connectError *errConnectUserdb
		if !errors.As(err, &connectError) {
			// systemd userdb is available per se, but an error occurred:
			return nil, err
		}
		// fallthrough to parsing files ourselves:
	} else {
		return u, nil
	}

	f, err := os.Open(userFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return findUserId(uid, f)
}

type errConnectUserdb struct {
	underlying error
}

func (e *errConnectUserdb) Error() string {
	return fmt.Sprintf("could not connect to systemd user database: %v", e.underlying)
}

// userdbClient queries the io.systemd.UserDatabase service provided by
// systemd-userdbd.service(8) for obtaining full user/group details even when
// cgo is not available.
type userdbClient struct {
	address string
}

var defaultUserdbClient = &userdbClient{
	address: "/run/systemd/userdb/io.systemd.NameServiceSwitch",
}

func (c *userdbClient) query(method string, unmarshal func([]byte) (bool, error)) error {
	conn, err := net.Dial("unix", c.address)
	if err != nil {
		return &errConnectUserdb{err}
	}
	defer conn.Close()

	// The other end of this socket is implemented in
	// https://github.com/systemd/systemd/tree/v245/src/userdb

	type params struct {
		Service string `json:"service"`
	}
	req := struct {
		Method     string `json:"method"`
		Parameters params `json:"parameters"`
		More       bool   `json:"more"`
	}{
		Method: method,
		Parameters: params{
			Service: "io.systemd.NameServiceSwitch",
		},
		More: true,
	}

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if _, err := conn.Write(append(b, 0)); err != nil {
		return err
	}
	sc := bufio.NewScanner(conn)
	// This is bufio.ScanLines, but looking for a 0-byte instead of '\n':
	sc.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, 0); i >= 0 {
			// We have a full 0-byte-terminated line.
			return i + 1, data[0:i], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	})
	for sc.Scan() {
		continues, err := unmarshal(sc.Bytes())
		if err != nil {
			return nil
		}
		if !continues {
			break
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return nil
}

func (cl *userdbClient) queryGroupDb(predicate func(*groupRecord) bool) (*Group, error) {
	const method = "io.systemd.UserDatabase.GetGroupRecord"
	var group *Group
	unmarshal := func(b []byte) (bool, error) {
		var reply struct {
			Parameters struct {
				Record groupRecord `json:"record"`
			} `json:"parameters"`
			Continues bool `json:"continues"`
		}
		if err := json.Unmarshal(b, &reply); err != nil {
			return false, err
		}
		r := reply.Parameters.Record // for convenience
		if !predicate(&r) {
			return reply.Continues, nil // skip
		}
		group = &Group{
			Name: r.GroupName,
			Gid:  strconv.FormatInt(r.Gid, 10),
		}
		return reply.Continues, nil
	}

	if err := cl.query(method, unmarshal); err != nil {
		return nil, err
	}
	return group, nil
}

func (cl *userdbClient) queryUserDb(predicate func(*userRecord) bool) (*User, error) {
	const method = "io.systemd.UserDatabase.GetUserRecord"
	var u *User
	unmarshal := func(b []byte) (bool, error) {
		var reply struct {
			Parameters struct {
				Record userRecord `json:"record"`
			} `json:"parameters"`
			Continues bool `json:"continues"`
		}
		if err := json.Unmarshal(b, &reply); err != nil {
			return false, err
		}
		r := reply.Parameters.Record // for convenience
		if !predicate(&r) {
			return reply.Continues, nil // skip
		}
		u = &User{
			Uid:      strconv.FormatInt(r.Uid, 10),
			Gid:      strconv.FormatInt(r.Gid, 10),
			Username: r.UserName,
			Name:     r.RealName,
			HomeDir:  r.HomeDirectory,
		}
		return reply.Continues, nil
	}
	if err := cl.query(method, unmarshal); err != nil {
		return nil, err
	}
	return u, nil
}

type groupRecord struct {
	GroupName string `json:"groupName"`
	Gid       int64  `json:"gid"`
}

func (cl *userdbClient) lookupGroup(groupname string) (*Group, error) {
	return cl.queryGroupDb(func(g *groupRecord) bool {
		return g.GroupName == groupname
	})
}

func (cl *userdbClient) lookupGroupId(id string) (*Group, error) {
	return cl.queryGroupDb(func(g *groupRecord) bool {
		return strconv.FormatInt(g.Gid, 10) == id
	})
}

type userRecord struct {
	UserName      string `json:"userName"`
	RealName      string `json:"realName"`
	Uid           int64  `json:"uid"`
	Gid           int64  `json:"gid"`
	HomeDirectory string `json:"homeDirectory"`
}

func (cl *userdbClient) lookupUser(username string) (*User, error) {
	return cl.queryUserDb(func(u *userRecord) bool {
		return u.UserName == username
	})
}

func (cl *userdbClient) lookupUserId(uid string) (*User, error) {
	return cl.queryUserDb(func(u *userRecord) bool {
		return strconv.FormatInt(u.Uid, 10) == uid
	})
}
