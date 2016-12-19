// Package common payload struct and func's
// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package common

import (
	"syscall"
)

// Payload represents a packet going through either the incoming or outgoing pipelines
type Payload struct {
	Raw       []byte
	Packet    []byte
	IPAddress []byte
	Nonce     []byte
	Sockaddr  syscall.Sockaddr
	Length    int
}

// NewTunPayload is used to generate a payload based on a TUN packet
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
