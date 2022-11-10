// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build cgo && !osusergo && unix && !android && !darwin

package user

import (
	"syscall"
	"unsafe"
)

/*
#cgo solaris CFLAGS: -D_POSIX_PTHREAD_SEMANTICS
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>
#include <grp.h>
#include <stdlib.h>

static int mygetpwuid_r(int uid, struct passwd *pwd,
	char *buf, size_t buflen, struct passwd **result) {
	return getpwuid_r(uid, pwd, buf, buflen, result);
}

static int mygetpwnam_r(const char *name, struct passwd *pwd,
	char *buf, size_t buflen, struct passwd **result) {
	return getpwnam_r(name, pwd, buf, buflen, result);
}

static int mygetgrgid_r(int gid, struct group *grp,
	char *buf, size_t buflen, struct group **result) {
 return getgrgid_r(gid, grp, buf, buflen, result);
}

static int mygetgrnam_r(const char *name, struct group *grp,
	char *buf, size_t buflen, struct group **result) {
 return getgrnam_r(name, grp, buf, buflen, result);
}
*/
import "C"

type _C_char = C.char
type _C_int = C.int
type _C_gid_t = C.gid_t
type _C_uid_t = C.uid_t
type _C_size_t = C.size_t
type _C_struct_group = C.struct_group
type _C_struct_passwd = C.struct_passwd
type _C_long = C.long

func _C_pw_uid(p *_C_struct_passwd) _C_uid_t   { return p.pw_uid }
func _C_pw_uidp(p *_C_struct_passwd) *_C_uid_t { return &p.pw_uid }
func _C_pw_gid(p *_C_struct_passwd) _C_gid_t   { return p.pw_gid }
func _C_pw_gidp(p *_C_struct_passwd) *_C_gid_t { return &p.pw_gid }
func _C_pw_name(p *_C_struct_passwd) *_C_char  { return p.pw_name }
func _C_pw_gecos(p *_C_struct_passwd) *_C_char { return p.pw_gecos }
func _C_pw_dir(p *_C_struct_passwd) *_C_char   { return p.pw_dir }

func _C_gr_gid(g *_C_struct_group) _C_gid_t  { return g.gr_gid }
func _C_gr_name(g *_C_struct_group) *_C_char { return g.gr_name }

func _C_GoString(p *_C_char) string { return C.GoString(p) }

func _C_getpwnam_r(name *_C_char, pwd *_C_struct_passwd, buf *_C_char, size _C_size_t, result **_C_struct_passwd) syscall.Errno {
	return syscall.Errno(C.mygetpwnam_r(name, pwd, buf, size, result))
}

func _C_getpwuid_r(uid _C_uid_t, pwd *_C_struct_passwd, buf *_C_char, size _C_size_t, result **_C_struct_passwd) syscall.Errno {
	return syscall.Errno(C.mygetpwuid_r(_C_int(uid), pwd, buf, size, result))
}

func _C_getgrnam_r(name *_C_char, grp *_C_struct_group, buf *_C_char, size _C_size_t, result **_C_struct_group) syscall.Errno {
	return syscall.Errno(C.mygetgrnam_r(name, grp, buf, size, result))
}

func _C_getgrgid_r(gid _C_gid_t, grp *_C_struct_group, buf *_C_char, size _C_size_t, result **_C_struct_group) syscall.Errno {
	return syscall.Errno(C.mygetgrgid_r(_C_int(gid), grp, buf, size, result))
}

const (
	_C__SC_GETPW_R_SIZE_MAX = C._SC_GETPW_R_SIZE_MAX
	_C__SC_GETGR_R_SIZE_MAX = C._SC_GETGR_R_SIZE_MAX
)

func _C_sysconf(key _C_int) _C_long                           { return C.sysconf(key) }
func _C_malloc(n _C_size_t) unsafe.Pointer                    { return C.malloc(n) }
func _C_realloc(p unsafe.Pointer, n _C_size_t) unsafe.Pointer { return C.realloc(p, n) }
func _C_free(p unsafe.Pointer)                                { C.free(p) }
