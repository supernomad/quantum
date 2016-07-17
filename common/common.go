package common

import (
	"encoding/binary"
	"net"
)

const (
	MTU             = 65475
	HeaderSize      = 16
	FooterSize      = 16
	MaxPacketLength = HeaderSize + MTU + FooterSize
)

const (
	// IP offsets x4
	IPStart = 0
	IPEnd   = 4

	// Nonce offsets x12
	NonceStart = 4
	NonceEnd   = 16

	// Packet offsets
	PacketStart = 16
)

func IPtoInt(IP string) uint32 {
	buf := net.ParseIP(IP).To4()
	return binary.LittleEndian.Uint32(buf)
}
