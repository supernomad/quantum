package common

// Payload represents a packet going through either the incoming or outgoing pipelines
type Payload struct {
	Raw       []byte
	Packet    []byte
	IPAddress []byte
	Nonce     []byte
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

// NewIPPayload is used to generate a payload based on a IP packet
func NewIPPayload(raw []byte, packetLength int) *Payload {
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
