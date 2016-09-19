package common

// Payload represents a packet going through either the incoming or outgoing pipelines
type Payload struct {
	Raw       []byte
	Packet    []byte
	IPAddress []byte
	Nonce     []byte
	Length    int
	Mapping   *Mapping
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

func NewIPPayload(raw []byte, packetLength int) *Payload {
	offset := raw[0] & 15 * 4
	ip := raw[offset+IPStart : offset+IPEnd]
	nonce := raw[offset+NonceStart : offset+NonceEnd]
	pkt := raw[offset+PacketStart : packetLength]

	return &Payload{
		Raw:       raw,
		IPAddress: ip,
		Nonce:     nonce,
		Packet:    pkt,
		Length:    packetLength - int(offset),
	}
}
