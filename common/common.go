package common

import (
	"encoding/json"
)

const (
	BlockSize       = 16
	MaxPacketLength = 1560
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

type Payload struct {
	Raw       []byte
	Packet    []byte
	IpAddress []byte
	Nonce     []byte
	Length    int
	Mapping   Mapping
}

type Mapping struct {
	Address   string
	PublicKey []byte
	SecretKey []byte `json:"-"`
}

func (m Mapping) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

func ParseMapping(data string) Mapping {
	var mapping Mapping
	json.Unmarshal([]byte(data), &mapping)
	return mapping
}

func NewTunPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IpStart:IpEnd]
	nonce := raw[NonceStart:NonceEnd]
	pkt := raw[PacketStart : PacketStart+packetLength]

	return &Payload{
		Raw:       raw,
		IpAddress: ip,
		Nonce:     nonce,
		Packet:    pkt,
		Length:    PacketStart + packetLength + BlockSize,
	}
}

func NewSockPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IpStart:IpEnd]
	nonce := raw[NonceStart:NonceEnd]
	pkt := raw[PacketStart:packetLength]

	return &Payload{
		Raw:       raw,
		IpAddress: ip,
		Nonce:     nonce,
		Packet:    pkt,
	}
}
