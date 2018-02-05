// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package common

// Payload represents a packet traversing the quantum network.
type Payload struct {
	// The raw byte array representing the payload, which includes all necessary metadata.
	Raw []byte

	// The packet data within the raw payload.
	Packet []byte

	// The private ip address of the remote peer within the raw payload.
	IPAddress []byte

	// The total length of the payload.
	Length int
}

// NewTunPayload is used to generate a payload based on a received TUN packet.
func NewTunPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IPStart:IPEnd]
	pkt := raw[PacketStart : PacketStart+packetLength]

	return &Payload{
		Raw:       raw,
		IPAddress: ip,
		Packet:    pkt,
		Length:    HeaderSize + packetLength,
	}
}

// NewSockPayload is used to generate a payload based on a received Socket packet.
func NewSockPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IPStart:IPEnd]
	pkt := raw[PacketStart:packetLength]

	return &Payload{
		Raw:       raw,
		IPAddress: ip,
		Packet:    pkt,
		Length:    packetLength,
	}
}
