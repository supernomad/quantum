// Copyright (c) 2016 Christian Saide <Supernomad>
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

	// NonceStart - The nonce start position within a quantum packet.
	NonceStart = 4

	// NonceEnd - The nonce end position within a quantum packet.
	NonceEnd = 16

	// NonceLength - The nonce header length.
	NonceLength = 12

	// TagLength - The GCM tag header length.
	TagLength = 16

	// PacketStart - The real packet start position within a quantum packet.
	PacketStart = 16

	// MaxPacketLength - The maximum packet size to send via the UDP device.
	// StandardMTU(1500) - IPHeader(20) - UDPHeader(8) - SafetyNet(2).
	MaxPacketLength = 1470

	// HeaderSize - The size of the data perpended tp the real packet.
	HeaderSize = IPLength + NonceLength

	// FooterSize - The size of the data appended to the real packet.
	FooterSize = TagLength

	// MTU - The max size packet to receive from the TUN device.
	MTU = MaxPacketLength - HeaderSize - FooterSize
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
