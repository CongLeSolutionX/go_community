// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package windows

import (
	"syscall"
	"unsafe"
)

//go:generate go run ../../../syscall/mksyscall_windows.go -output zsyscall_windows.go syscall_windows.go

const GAA_FLAG_INCLUDE_PREFIX = 0x00000010

const (
	IF_TYPE_OTHER              = 1
	IF_TYPE_ETHERNET_CSMACD    = 6
	IF_TYPE_ISO88025_TOKENRING = 9
	IF_TYPE_PPP                = 23
	IF_TYPE_SOFTWARE_LOOPBACK  = 24
	IF_TYPE_ATM                = 37
	IF_TYPE_IEEE80211          = 71
	IF_TYPE_TUNNEL             = 131
	IF_TYPE_IEEE1394           = 144
)

type SocketAddress struct {
	Sockaddr       *syscall.RawSockaddrAny
	SockaddrLength int32
}

type IpAdapterUnicastAddress struct {
	Length             uint32
	Flags              uint32
	Next               *IpAdapterUnicastAddress
	Address            SocketAddress
	PrefixOrigin       int32
	SuffixOrigin       int32
	DadState           int32
	ValidLifetime      uint32
	PreferredLifetime  uint32
	LeaseLifetime      uint32
	OnLinkPrefixLength uint8
}

type IpAdapterAnycastAddress struct {
	Length  uint32
	Flags   uint32
	Next    *IpAdapterAnycastAddress
	Address SocketAddress
}

type IpAdapterMulticastAddress struct {
	Length  uint32
	Flags   uint32
	Next    *IpAdapterMulticastAddress
	Address SocketAddress
}

type IpAdapterDnsServerAdapter struct {
	Length   uint32
	Reserved uint32
	Next     *IpAdapterDnsServerAdapter
	Address  SocketAddress
}

type IpAdapterPrefix struct {
	Length       uint32
	Flags        uint32
	Next         *IpAdapterPrefix
	Address      SocketAddress
	PrefixLength uint32
}

type IpAdapterAddresses struct {
	Length                uint32
	IfIndex               uint32
	Next                  *IpAdapterAddresses
	AdapterName           *byte
	FirstUnicastAddress   *IpAdapterUnicastAddress
	FirstAnycastAddress   *IpAdapterAnycastAddress
	FirstMulticastAddress *IpAdapterMulticastAddress
	FirstDnsServerAddress *IpAdapterDnsServerAdapter
	DnsSuffix             *uint16
	Description           *uint16
	FriendlyName          *uint16
	PhysicalAddress       [syscall.MAX_ADAPTER_ADDRESS_LENGTH]byte
	PhysicalAddressLength uint32
	Flags                 uint32
	Mtu                   uint32
	IfType                uint32
	OperStatus            uint32
	Ipv6IfIndex           uint32
	ZoneIndices           [16]uint32
	FirstPrefix           *IpAdapterPrefix
	/* more fields might be present here. */
}

const (
	IfOperStatusUp             = 1
	IfOperStatusDown           = 2
	IfOperStatusTesting        = 3
	IfOperStatusUnknown        = 4
	IfOperStatusDormant        = 5
	IfOperStatusNotPresent     = 6
	IfOperStatusLowerLayerDown = 7
)

//sys	GetAdaptersAddresses(family uint32, flags uint32, reserved uintptr, adapterAddresses *IpAdapterAddresses, sizePointer *uint32) (errcode error) = iphlpapi.GetAdaptersAddresses
//sys	GetComputerNameEx(nameformat uint32, buf *uint16, n *uint32) (err error) = GetComputerNameExW
//sys	MoveFileEx(from *uint16, to *uint16, flags uint32) (err error) = MoveFileExW

const (
	ComputerNameNetBIOS                   = 0
	ComputerNameDnsHostname               = 1
	ComputerNameDnsDomain                 = 2
	ComputerNameDnsFullyQualified         = 3
	ComputerNamePhysicalNetBIOS           = 4
	ComputerNamePhysicalDnsHostname       = 5
	ComputerNamePhysicalDnsDomain         = 6
	ComputerNamePhysicalDnsFullyQualified = 7
	ComputerNameMax                       = 8

	MOVEFILE_REPLACE_EXISTING      = 0x1
	MOVEFILE_COPY_ALLOWED          = 0x2
	MOVEFILE_DELAY_UNTIL_REBOOT    = 0x4
	MOVEFILE_WRITE_THROUGH         = 0x8
	MOVEFILE_CREATE_HARDLINK       = 0x10
	MOVEFILE_FAIL_IF_NOT_TRACKABLE = 0x20
)

func Rename(oldpath, newpath string) error {
	from, err := syscall.UTF16PtrFromString(oldpath)
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(newpath)
	if err != nil {
		return err
	}
	return MoveFileEx(from, to, MOVEFILE_REPLACE_EXISTING)
}

//sys	GetACP() (acp uint32) = kernel32.GetACP
//sys	MultiByteToWideChar(codePage uint32, dwFlags uint32, str *byte, nstr int32, wchar *uint16, nwchar int32) (nwrite int32, err error) = kernel32.MultiByteToWideChar

const (
	FILE_NAME_NORMALIZED = 0x0
	FILE_NAME_OPEND      = 0x8

	VOLUME_NAME_DOS  = 0x0
	VOLUME_NAME_GUID = 0x1
	VOLUME_NAME_NONE = 0x4
	VOLUME_NAME_NT   = 0x2
)

const (
	ObjectBasicInformation = iota
	ObjectNameInformation
	ObjectTypeInformation
	ObjectAllInformation
	ObjectDataInformation
)

const (
	ERROR_NOT_ENOUGH_MEMORY syscall.Errno = 0x8
)

const (
	STATUS_BUFFER_OVERFLOW syscall.Errno = 0x80000005
)

//sys	ntQueryObject(handle syscall.Handle, infoClass uint32, info *byte, infoLen uint32, retLen *uint32) (lasterr error) = ntdll.NtQueryObject
//sys	getFinalPathNameByHandle(handle syscall.Handle, path *uint16, pathLen uint32, flag uint32) (n uint32, err error) = kernel32.GetFinalPathNameByHandleW
//sys	getLogicalDriveStrings(bufLen uint32, buffer *uint16) (n uint32, err error) = kernel32.GetLogicalDriveStringsW
//sys	queryDosDevice(drive *uint16, volume *uint16, volumeLen uint32) (n uint32, err error) = kernel32.QueryDosDeviceW

var (
	unc, _                = syscall.UTF16FromString(`UNC\`)
	ntLanmanRedirector, _ = syscall.UTF16FromString(`\Device\LanmanRedirector\`)
	ntMup, _              = syscall.UTF16FromString(`\Device\Mup\`)
)

func wIndex(ws []uint16, w uint16) int {
	for i, c := range ws {
		if c == w {
			return i
		}
	}

	return -1
}

func wEqual(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}

	for i, w := range a {
		if b[i] != w {
			return false
		}
	}

	return true
}

func wHasPrefix(ws, prefix []uint16) bool {
	return len(ws) >= len(prefix) && wEqual(ws[0:len(prefix)], prefix)
}

func ntPathByHandle(fd syscall.Handle) (path []uint16, err error) {
	type unicode struct {
		Length        uint16
		MaximumLength uint16
		Buffer        *uint16
	}

	type objNameInfo struct {
		Name       unicode
		NameBuffer uint32
	}

	var n uint32

	bufLen := uint32(syscall.MAX_PATH*2 + 8)
	for {
		buf := make([]byte, bufLen)

		err := ntQueryObject(fd, ObjectNameInformation, &buf[0], bufLen, &n)

		if err != STATUS_BUFFER_OVERFLOW {
			if err != nil {
				return nil, err
			}

			info := (*objNameInfo)(unsafe.Pointer(&buf[0]))

			path = (*[0xffff]uint16)(unsafe.Pointer(info.Name.Buffer))[:info.Name.Length/2]

			return path, nil
		}

		bufLen = n
	}
}

func getWin32Devices() ([]uint16, error) {
	win32DevicesLen := uint32(100)

	for {
		win32Devices := make([]uint16, win32DevicesLen)

		n, err := getLogicalDriveStrings(win32DevicesLen, &win32Devices[0])
		if err != nil {
			return nil, err
		}

		if n < win32DevicesLen {
			win32Devices = win32Devices[:n]

			return win32Devices, nil
		}

		win32DevicesLen *= 2
	}
}

func FinalPathByHandle(fd syscall.Handle) (path string, err error) {
	if fd == syscall.InvalidHandle {
		return "", syscall.EINVAL
	}

	// GetFinalPathNameByHandle is not supported before Windows Vista
	if procGetFinalPathNameByHandleW.Find() == nil {
		pathLen := uint32(syscall.MAX_PATH)

		for {
			path := make([]uint16, pathLen)

			n, err := getFinalPathNameByHandle(fd, &path[0], pathLen, FILE_NAME_NORMALIZED|VOLUME_NAME_DOS)
			if err == ERROR_NOT_ENOUGH_MEMORY {
				pathLen *= 2

				continue
			}

			if err != nil {
				break
			}

			path = path[4:n] // trim long path prefix \\?\

			if wHasPrefix(path, unc) {
				path[len(unc)-2] = uint16('\\') // replace UNC\ to \\

				return syscall.UTF16ToString(path[len(unc)-2:]), nil
			}

			return syscall.UTF16ToString(path), nil
		}
	}

	// fallback

	// https://msdn.microsoft.com/en-us/library/windows/desktop/aa365247(v=vs.85).aspx
	// NtQueryObject return the path as NT Namespace.
	// We need to convert it to Win32 File Namespace.

	ntPath, err := ntPathByHandle(fd)
	if err != nil {
		return "", err
	}

	// is UNC path? (XP)
	if wHasPrefix(ntPath, ntLanmanRedirector) {
		ntPath[len(ntLanmanRedirector)-2] = uint16('\\')

		return syscall.UTF16ToString(ntPath[len(ntLanmanRedirector)-2:]), nil
	}

	// is UNC path? (Vista+)
	if wHasPrefix(ntPath, ntMup) {
		ntPath[len(ntMup)-2] = uint16('\\')

		return syscall.UTF16ToString(ntPath[len(ntMup)-2:]), nil
	}

	// get win32Devices (i.e C:\<null>D:\<null>...<null>)

	win32Devices, err := getWin32Devices()
	if err != nil {
		return "", err
	}

	// traverse win32Devices and get ntDevice (i.e. \Device\HarddiskVolume1) per win32Device (i.e C:\)
	// if ntDevice match ntPath's prefix, then replace it by win32Device and return ntPath.

	{
		ntDeviceBufLen := uint32(50)

		ntDeviceBuf := make([]uint16, ntDeviceBufLen)

		for {
			i := wIndex(win32Devices, 0)

			if i == -1 { // unknown ntDevice
				return "", nil
			}

			win32Devices[i-1] = 0

			win32Device := win32Devices[:i-1]

			var ntDevice []uint16

			for {
				n, err := queryDosDevice(&win32Device[0], &ntDeviceBuf[0], ntDeviceBufLen)
				if err != syscall.ERROR_INSUFFICIENT_BUFFER {
					if err != nil {
						return "", err
					}

					ntDevice = ntDeviceBuf[:n-2]

					break
				}

				ntDeviceBufLen *= 2

				ntDeviceBuf = make([]uint16, ntDeviceBufLen)
			}

			if wHasPrefix(ntPath, ntDevice) {
				ntPath = ntPath[len(ntDevice)-len(win32Device):]

				copy(ntPath, win32Device)

				return syscall.UTF16ToString(ntPath), nil
			}

			win32Devices = win32Devices[i+1:]
		}
	}
}
