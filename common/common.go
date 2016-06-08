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
	PacketStart = 44

	// Key offsets
	KeyStart = 12
	KeyEnd   = 44

	// Nonce offsets
	NonceStart = 0
	NonceEnd   = 12
)

type Payload struct {
	Raw       []byte
	Packet    []byte
	Nonce     []byte
	Key       []byte
	PublicKey []byte
	Length    int
	Address   string
}

type Mapping struct {
	Address   string
	PublicKey []byte
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
	nonce := raw[NonceStart:NonceEnd]
	key := raw[KeyStart:KeyEnd]
	pkt := raw[PacketStart : PacketStart+packetLength]

	return &Payload{
		Raw:    raw,
		Nonce:  nonce,
		Key:    key,
		Packet: pkt,
		Length: PacketStart + packetLength + BlockSize,
	}
}

func NewSockPayload(raw []byte, packetLength int) *Payload {
	nonce := raw[NonceStart:NonceEnd]
	key := raw[KeyStart:KeyEnd]
	pkt := raw[PacketStart:packetLength]

	return &Payload{
		Raw:    raw,
		Nonce:  nonce,
		Key:    key,
		Packet: pkt,
	}
}
