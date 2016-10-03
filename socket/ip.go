package socket

import (
	"encoding/binary"
	"fmt"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/packet"
	"golang.org/x/net/ipv4"
)

type IP struct {
	id     int
	packet *packet.Packet
	cfg    *common.Config
}

func checksum(buf []byte) {
	buf[10] = 0
	buf[11] = 0

	var csum uint32
	for i := 0; i < len(buf); i += 2 {
		csum += uint32(buf[i]) << 8
		csum += uint32(buf[i+1])
	}
	for {
		if csum <= 65535 {
			break
		}
		csum = (csum >> 16) + uint32(uint16(csum))
	}
	binary.BigEndian.PutUint16(buf[10:12], ^uint16(csum))
}

func (sock *IP) Open() error {
	return sock.packet.Open()
}

func (sock *IP) Close() error {
	return sock.packet.Close()
}

func (sock *IP) GetFDs() []int {
	return sock.packet.GetFDs()
}

func (sock *IP) Read(buf []byte, queue int) (*common.Payload, bool) {
	err := sock.packet.Recv(buf, queue)
	if err != nil {
		return nil, false
	}
	header, err := ipv4.ParseHeader(buf)
	return common.NewIPPayload(buf[20:header.TotalLen], header.TotalLen-20), true
}

func (sock *IP) Write(payload *common.Payload, mapping *common.Mapping, queue int) bool {
	iph := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: payload.Length + 20,
		ID:       sock.id,
		Flags:    0,
		FragOff:  0,
		TTL:      64,
		Protocol: 138,
		Checksum: 0,
		Src:      sock.cfg.PublicIPAddr,
		Dst:      mapping.Addr,
		Options:  nil,
	}
	sock.id++

	buf, _ := iph.Marshal()
	checksum(buf)
	buf = append(buf, payload.Raw[:payload.Length]...)

	err := sock.packet.Send(buf, queue)
	if err != nil {
		fmt.Println(err)
		return false
	}

	err = sock.packet.Flush(queue)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func newIP(cfg *common.Config) *IP {
	return &IP{
		id:     0,
		packet: packet.New(cfg),
		cfg:    cfg,
	}
}
