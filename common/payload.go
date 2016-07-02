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

func getSlice(offset, start, end int, slice []byte) []byte {
	return slice[offset+start : offset+end]
}

func NewTunPayload(raw []byte, packetLength int) *Payload {
	pkt := getSlice(0, PacketStart, packetLength, raw)

	tag := getSlice(packetLength, TagStart, TagEnd, raw)
	ip := getSlice(packetLength, IpStart, IpEnd, raw)
	nonce := getSlice(packetLength, NonceStart, NonceEnd, raw)

	hash := getSlice(packetLength, HashStart, HashEnd, raw)

	r := getSlice(packetLength, RStart, REnd, raw)
	s := getSlice(packetLength, SStart, SEnd, raw)

	return &Payload{
		Raw:       raw,
		Packet:    pkt,
		Tag:       tag,
		IpAddress: ip,
		Nonce:     nonce,
		Hash:      hash,
		R:         r,
		S:         s,
		Length:    packetLength + FooterSize,
	}
}

func NewSockPayload(raw []byte, packetLength int) *Payload {
	pkt := getSlice(0, PacketStart, packetLength-FooterSize+16, raw)

	offset := packetLength - FooterSize
	tag := getSlice(offset, TagStart, TagEnd, raw)
	ip := getSlice(offset, IpStart, IpEnd, raw)
	nonce := getSlice(offset, NonceStart, NonceEnd, raw)

	hash := getSlice(offset, HashStart, HashEnd, raw)

	r := getSlice(offset, RStart, REnd, raw)
	s := getSlice(offset, SStart, SEnd, raw)

	return &Payload{
		Raw:       raw,
		Packet:    pkt,
		Tag:       tag,
		IpAddress: ip,
		Nonce:     nonce,
		Hash:      hash,
		R:         r,
		S:         s,
		Length:    packetLength,
	}
}
