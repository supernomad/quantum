package common

type Payload struct {
	Raw       []byte
	Packet    []byte
	IpAddress []byte
	Nonce     []byte
	R         []byte
	S         []byte
	Length    int
	Mapping   *Mapping
}

func NewTunPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IpStart:IpEnd]
	nonce := raw[NonceStart:NonceEnd]
	r := raw[RStart:REnd]
	s := raw[SStart:SEnd]
	pkt := raw[PacketStart : PacketStart+packetLength]

	return &Payload{
		Raw:       raw,
		IpAddress: ip,
		Nonce:     nonce,
		R:         r,
		S:         s,
		Packet:    pkt,
		Length:    PacketStart + packetLength + BlockSize,
	}
}

func NewSockPayload(raw []byte, packetLength int) *Payload {
	ip := raw[IpStart:IpEnd]
	nonce := raw[NonceStart:NonceEnd]
	r := raw[RStart:REnd]
	s := raw[SStart:SEnd]
	pkt := raw[PacketStart:packetLength]

	return &Payload{
		Raw:       raw,
		IpAddress: ip,
		Nonce:     nonce,
		R:         r,
		S:         s,
		Packet:    pkt,
	}
}
