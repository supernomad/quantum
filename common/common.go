package common

const (
	BlockSize       = 16
	MaxPacketLength = 1596
)

const (
	// IP offsets
	IpStart = 0
	IpEnd   = 4

	// Nonce offsets
	NonceStart = 4
	NonceEnd   = 16

	// R offsets
	RStart = 16
	REnd   = 48

	// S offsets
	SStart = 48
	SEnd   = 80

	// Packet offsets
	PacketStart = 80
)
