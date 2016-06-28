package common

type Payload struct {
	Raw       []byte
	Packet    []byte
	IpAddress []byte
	Nonce     []byte
	Length    int
	Mapping   *Mapping
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
