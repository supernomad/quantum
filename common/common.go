package common

const (
	BlockSize       = 16
	MaxPacketLength = 1532
)

const (
	// Packet offsets
	PacketStart = 16

	// IP offsets
	IpStart = 0
	IpEnd   = 4

	// Nonce offsets
	NonceStart = 4
	NonceEnd   = 16
)
