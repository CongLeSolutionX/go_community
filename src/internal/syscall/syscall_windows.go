package syscall

import (
	stdsyscall "syscall"
)

//go:generate go run mksyscall_windows.go -output zsyscall_windows.go syscall_windows.go

const GAA_FLAG_INCLUDE_PREFIX = 0x00000010

const IF_TYPE_SOFTWARE_LOOPBACK = 24

const (
	IF_ACCESS_LOOPBACK             = 1
	IF_ACCESS_BROADCAST            = 2
	IF_ACCESS_POINT_TO_POINT       = 3
	IF_ACCESS_POINT_TO_MULTI_POINT = 4
)

const (
	IP_ADAPTER_ADDRESS_DNS_ELIGIBLE = 0x00000001
	IP_ADAPTER_ADDRESS_TRANSIENT    = 0x00000002
)

type SocketAddress struct {
	Sockaddr       *stdsyscall.RawSockaddrAny
	SockaddrLength int32
}

type IpAdapterUnicastAddress struct {
	Length            uint32
	Flags             uint32
	Next              uintptr
	Address           SocketAddress
	PrefixOrigin      int32
	SuffixOrigin      int32
	DadState          int32
	ValidLifetime     uint32
	PreferredLifetime uint32
	LeaseLifetime     uint32
}

type IpAdapterAnycastAddress struct {
	Length  uint32
	Flags   uint32
	Next    uintptr
	Address SocketAddress
}

type IpAdapterMulticastAddress struct {
	Length  uint32
	Flags   uint32
	Next    uintptr
	Address SocketAddress
}

type IpAdapterDnsServerAdapter struct {
	Length   uint32
	Reserved uint32
	Next     uintptr
	Address  SocketAddress
}

type IpAdapterPrefix struct {
	Length       uint32
	Flags        uint32
	Next         uintptr
	Address      SocketAddress
	PrefixLength uint32
}

type IpAdapterAddresses struct {
	Length                uint32
	IfIndex               uint32
	Next                  uintptr
	AdapterName           *byte
	FirstUnicastAddress   *IpAdapterUnicastAddress
	FirstAnycastAddress   *IpAdapterAnycastAddress
	FirstMulticastAddress *IpAdapterMulticastAddress
	FirstDnsServerAddress *IpAdapterDnsServerAdapter
	DnsSuffix             *uint16
	Description           *uint16
	FriendlyName          *uint16
	PhysicalAddress       [stdsyscall.MAX_ADAPTER_ADDRESS_LENGTH]byte
	PhysicalAddressLength uint32
	Flags                 uint32
	Mtu                   uint32
	IfType                uint32
	OperStatus            uint32
	Ipv6IfIndex           uint32
	ZoneIndices           [16]uint32
	FirstPrefix           *IpAdapterPrefix
}

type IpInterfaceNameInfo struct {
	Index          uint32
	MediaType      uint32
	ConnectionType uint8
	AccessType     uint8
	DeviceGuid     [16]byte
	InterfaceGuid  [16]byte
}

const (
	IFF_UP          = 1 << 0
	IFF_LOOPBACK    = 1 << 1
	IFF_BROADCAST   = 1 << 2
	IFF_POINTOPOINT = 1 << 3
	IFF_MULTICAST   = 1 << 4
)

const (
	IfOperStatusUp             = 1
	IfOperStatusDown           = 2
	IfOperStatusTesting        = 3
	IfOperStatusUnknown        = 4
	IfOperStatusDormant        = 5
	IfOperStatusNotPresent     = 6
	IfOperStatusLowerLayerDown = 7
)

//sys GetAdaptersAddresses(family uint32, flags uint32, reserved uintptr, adapterAddresses *IpAdapterAddresses, sizeOfPointer *uint32) (errcode error) = iphlpapi.GetAdaptersAddresses
//sys NhpAllocateAndGetInterfaceInfoFromStack(ppTable **IpInterfaceNameInfo, count *uint32, order bool, handle stdsyscall.Handle, flags uint32) (errcode error) = iphlpapi.NhpAllocateAndGetInterfaceInfoFromStack
//sys GetProcessHeap() (handle stdsyscall.Handle) = kernel32.GetProcessHeap
