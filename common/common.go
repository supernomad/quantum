// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"encoding/binary"
	"net"
)

const (
	// RealDeviceNameEnv is the environment variable that the real network device name is stored in for reloads.
	RealDeviceNameEnv = "_QUANTUM_REAL_DEVICE_NAME_"

	// IPStart - The ip start position within a quantum packet.
	IPStart = 0

	// IPEnd - The ip end position within a quantum packet.
	IPEnd = 4

	// IPLength - The length of the private ip header.
	IPLength = 4

	// PacketStart - The real packet start position within a quantum packet.
	PacketStart = 4

	// MaxPacketLength - The maximum packet size to send via the UDP device.
	// StandardMTU(1500) - IPHeader(20) - UDPHeader(8).
	MaxPacketLength = 1472

	// HeaderSize - The size of the data perpended tp the real packet.
	HeaderSize = IPLength

	// OverflowSize - An extra buffer for overflow of the MTU for plugins and other things to use incase its necessary.
	OverflowSize = 35

	// MTU - The max size packet to receive from the TUN device.
	MTU = MaxPacketLength - HeaderSize - OverflowSize
)

// IPtoInt takes an ipv4 net.IP and returns a uint32 that represents it.
func IPtoInt(IP net.IP) uint32 {
	buf := IP.To4()
	return binary.LittleEndian.Uint32(buf)
}

// IncrementIP will increment the given ipv4 net.IP by 1 in place.
func IncrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

// ArrayEquals returns true if both byte slices contain the same data.
//
// NOTE: this is a very slow func and should be limited in use.
func ArrayEquals(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil || len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// StringInSlice returns true if the string 'a' is contained in the string array 'slice'.
func StringInSlice(a string, slice []string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == a {
			return true
		}
	}
	return false
}
