package common

type Payload struct {
	Raw       []byte
	Packet    []byte
	Tag       []byte
	IpAddress []byte
	Nonce     []byte
	R         []byte
	S         []byte
	Hash      []byte
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
		Length:    HeaderSize + packetLength + FooterSize,
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
		Length:    packetLength,
	}
}
