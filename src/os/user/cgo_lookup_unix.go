// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (cgo || darwin) && !osusergo && unix && !android

package user

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func current() (*User, error) {
	return lookupUnixUid(syscall.Getuid())
}

func lookupUser(username string) (*User, error) {
	var pwd _C_struct_passwd
	var result *_C_struct_passwd
	nameC := make([]byte, len(username)+1)
	copy(nameC, username)

	buf := alloc(userBuffer)
	defer buf.free()

	err := retryWithBuffer(buf, func() syscall.Errno {
		return syscall.Errno(_C_getpwnam_r((*_C_char)(unsafe.Pointer(&nameC[0])),
			&pwd,
			(*_C_char)(buf.ptr),
			_C_size_t(buf.size),
			&result))
	})
	if err != nil {
		return nil, fmt.Errorf("user: lookup username %s: %v", username, err)
	}
	if result == nil {
		return nil, UnknownUserError(username)
	}
	return buildUser(&pwd), err
}

func lookupUserId(uid string) (*User, error) {
	i, e := strconv.Atoi(uid)
	if e != nil {
		return nil, e
	}
	return lookupUnixUid(i)
}

func lookupUnixUid(uid int) (*User, error) {
	var pwd _C_struct_passwd
	var result *_C_struct_passwd

	buf := alloc(userBuffer)
	defer buf.free()

	err := retryWithBuffer(buf, func() syscall.Errno {
		return syscall.Errno(_C_getpwuid_r(_C_uid_t(uid),
			&pwd,
			(*_C_char)(buf.ptr),
			_C_size_t(buf.size),
			&result))
	})
	if err != nil {
		return nil, fmt.Errorf("user: lookup userid %d: %v", uid, err)
	}
	if result == nil {
		return nil, UnknownUserIdError(uid)
	}
	return buildUser(&pwd), nil
}

func buildUser(pwd *_C_struct_passwd) *User {
	u := &User{
		Uid:      strconv.FormatUint(uint64(_C_pw_uid(pwd)), 10),
		Gid:      strconv.FormatUint(uint64(_C_pw_gid(pwd)), 10),
		Username: _C_GoString(_C_pw_name(pwd)),
		Name:     _C_GoString(_C_pw_gecos(pwd)),
		HomeDir:  _C_GoString(_C_pw_dir(pwd)),
	}
	// The pw_gecos field isn't quite standardized. Some docs
	// say: "It is expected to be a comma separated list of
	// personal data where the first item is the full name of the
	// user."
	u.Name, _, _ = strings.Cut(u.Name, ",")
	return u
}

func lookupGroup(groupname string) (*Group, error) {
	var grp _C_struct_group
	var result *_C_struct_group

	buf := alloc(groupBuffer)
	defer buf.free()
	cname := make([]byte, len(groupname)+1)
	copy(cname, groupname)

	err := retryWithBuffer(buf, func() syscall.Errno {
		return syscall.Errno(_C_getgrnam_r((*_C_char)(unsafe.Pointer(&cname[0])),
			&grp,
			(*_C_char)(buf.ptr),
			_C_size_t(buf.size),
			&result))
	})
	if err != nil {
		return nil, fmt.Errorf("user: lookup groupname %s: %v", groupname, err)
	}
	if result == nil {
		return nil, UnknownGroupError(groupname)
	}
	return buildGroup(&grp), nil
}

func lookupGroupId(gid string) (*Group, error) {
	i, e := strconv.Atoi(gid)
	if e != nil {
		return nil, e
	}
	return lookupUnixGid(i)
}

func lookupUnixGid(gid int) (*Group, error) {
	var grp _C_struct_group
	var result *_C_struct_group

	buf := alloc(groupBuffer)
	defer buf.free()

	err := retryWithBuffer(buf, func() syscall.Errno {
		return syscall.Errno(_C_getgrgid_r(_C_gid_t(gid),
			&grp,
			(*_C_char)(buf.ptr),
			_C_size_t(buf.size),
			&result))
	})
	if err != nil {
		return nil, fmt.Errorf("user: lookup groupid %d: %v", gid, err)
	}
	if result == nil {
		return nil, UnknownGroupIdError(strconv.Itoa(gid))
	}
	return buildGroup(&grp), nil
}

func buildGroup(grp *_C_struct_group) *Group {
	g := &Group{
		Gid:  strconv.Itoa(int(_C_gr_gid(grp))),
		Name: _C_GoString(_C_gr_name(grp)),
	}
	return g
}

type bufferKind _C_int

const (
	userBuffer  = bufferKind(_C__SC_GETPW_R_SIZE_MAX)
	groupBuffer = bufferKind(_C__SC_GETGR_R_SIZE_MAX)
)

func (k bufferKind) initialSize() _C_size_t {
	sz := _C_sysconf(_C_int(k))
	if sz == -1 {
		// DragonFly and FreeBSD do not have _SC_GETPW_R_SIZE_MAX.
		// Additionally, not all Linux systems have it, either. For
		// example, the musl libc returns -1.
		return 1024
	}
	if !isSizeReasonable(int64(sz)) {
		// Truncate.  If this truly isn't enough, retryWithBuffer will error on the first run.
		return maxBufferSize
	}
	return _C_size_t(sz)
}

type memBuffer struct {
	ptr  unsafe.Pointer
	size _C_size_t
}

func alloc(kind bufferKind) *memBuffer {
	sz := kind.initialSize()
	return &memBuffer{
		ptr:  _C_malloc(sz),
		size: sz,
	}
}

func (mb *memBuffer) resize(newSize _C_size_t) {
	mb.ptr = _C_realloc(mb.ptr, newSize)
	mb.size = newSize
}

func (mb *memBuffer) free() {
	_C_free(mb.ptr)
}

// retryWithBuffer repeatedly calls f(), increasing the size of the
// buffer each time, until f succeeds, fails with a non-ERANGE error,
// or the buffer exceeds a reasonable limit.
func retryWithBuffer(buf *memBuffer, f func() syscall.Errno) error {
	for {
		errno := f()
		if errno == 0 {
			return nil
		} else if errno != syscall.ERANGE {
			return errno
		}
		newSize := buf.size * 2
		if !isSizeReasonable(int64(newSize)) {
			return fmt.Errorf("internal buffer exceeds %d bytes", maxBufferSize)
		}
		buf.resize(newSize)
	}
}

const maxBufferSize = 1 << 20

func isSizeReasonable(sz int64) bool {
	return sz > 0 && sz <= maxBufferSize
}

// Because we can't use cgo in tests:
func structPasswdForNegativeTest() _C_struct_passwd {
	sp := _C_struct_passwd{}
	*_C_pw_uidp(&sp) = 1<<32 - 2
	*_C_pw_gidp(&sp) = 1<<32 - 3
	return sp
}
