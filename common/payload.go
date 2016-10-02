package common

import (
	"encoding/binary"
)

// Payload represents a packet going through either the incoming or outgoing pipelines
type Payload struct {
	Raw       []byte
	Packet    []byte
	IPAddress []byte
	Nonce     []byte
	IPHeader  []byte
	Length    int
}

// NewTunPayload is used to generate a payload based on a TUN packet
func NewTunPayload(raw []byte, packetLength int) *Payload {
	ipHeader := raw[IPHeaderStart:IPHeaderEnd]
	ip := raw[IPStart:IPEnd]
	nonce := raw[NonceStart:NonceEnd]
	pkt := raw[PacketStart : PacketStart+packetLength]

	return &Payload{
		Raw:       raw,
		IPHeader:  ipHeader,
		IPAddress: ip,
		Nonce:     nonce,
		Packet:    pkt,
		Length:    IPHeaderSize + HeaderSize + packetLength + FooterSize,
	}
}

// NewSockPayload is used to generate a payload based on a Socket packet
func NewSockPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IPStart:IPEnd]
	nonce := raw[NonceStart:NonceEnd]
	pkt := raw[PacketStart:packetLength]

	return &Payload{
		Raw:       raw,
		IPAddress: ip,
		Nonce:     nonce,
		Packet:    pkt,
		Length:    packetLength,
	}
}

func NewIPPayload(raw []byte) *Payload {
	ipHeader := raw[IPHeaderStart:IPHeaderEnd]
	packetLength := int(binary.BigEndian.Uint16(ipHeader[2:4]))
	ip := raw[IPStart:IPEnd]
	nonce := raw[NonceStart:NonceEnd]
	pkt := raw[PacketStart:packetLength]

	return &Payload{
		Raw:       raw,
		IPHeader:  ipHeader,
		IPAddress: ip,
		Nonce:     nonce,
		Packet:    pkt,
		Length:    packetLength,
	}
}
