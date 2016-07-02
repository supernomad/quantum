package common

const (
	FooterSize      = 96
	MaxPacketLength = 1596
)

const (
	// Tag offsets x16
	TagStart = 0
	TagEnd   = 16

	// IP offsets x4
	IpStart = 16
	IpEnd   = 20

	// Nonce offsets x12
	NonceStart = 20
	NonceEnd   = 32

	// R offsets x32
	RStart = 32
	REnd   = 64

	// S offsets x32
	SStart = 64
	SEnd   = 96

	// Hash offsets x32 (tag + ip + nonce)
	HashStart = 0
	HashEnd   = 32

	// Packet offsets
	PacketStart = 0
)
