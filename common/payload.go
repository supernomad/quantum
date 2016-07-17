package common

type Payload struct {
	Raw       []byte
	Packet    []byte
	IPAddress []byte
	Nonce     []byte
	Length    int
	Mapping   *Mapping
}

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
