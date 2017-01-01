// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package common

import (
	"syscall"
)

// Payload represents a packet traversing the quantum network.
type Payload struct {
	// The raw byte array representing the payload, which includes all necesary metadata.
	Raw []byte

	// The TUN packet data within the raw payload.
	Packet []byte

	// The private ip address of the sender within the raw payload.
	IPAddress []byte

	// The cryptographic nonce value within the raw payload.
	Nonce []byte

	// The syscall.Sockaddr representation of the destination address of the raw payload.
	Sockaddr syscall.Sockaddr

	// The total length of the payload.
	Length int
}

// NewTunPayload is used to generate a payload based on a received TUN packet.
func NewTunPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IPStart:IPEnd]
	nonce := raw[NonceStart:NonceEnd]
	pkt := raw[PacketStart : PacketStart+packetLength]

	return &Payload{
		Raw:       raw,
		IPAddress: ip,
		Nonce:     nonce,
		Packet:    pkt,
		Length:    HeaderSize + packetLength + FooterSize,
	}
}

// NewSockPayload is used to generate a payload based on a received Socket packet.
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
