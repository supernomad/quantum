package workers

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
	"net"
)

type Outgoing struct {
	tunnel    *tun.Tun
	sock      *socket.Socket
	privateIP []byte
	mappings  map[uint32]*common.Mapping
	quit      chan bool
}

func (out *Outgoing) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.Packet[16:20])

	if mapping, ok := out.mappings[dip]; ok {
		copy(payload.IpAddress, out.privateIP)
		return payload, mapping, true
	}

	return payload, nil, false
}

func (out *Outgoing) seal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := rand.Read(payload.Nonce)
	if err != nil {
		return payload, false
	}

	mapping.Cipher.Seal(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	return payload, true
}

func (out *Outgoing) Start(queue int) {
	go func() {
	loop:
		for {
			select {
			case <-out.quit:
				return
			default:
				payload, ok := out.tunnel.Read(queue)
				if !ok {
					continue loop
				}
				payload, mapping, ok := out.resolve(payload)
				if !ok {
					continue loop
				}
				payload, ok = out.seal(payload, mapping)
				if !ok {
					continue loop
				}
				out.sock.Write(payload, mapping.Address)
			}
		}
	}()
}

func (out *Outgoing) Stop() {
	go func() {
		out.quit <- true
	}()
}

func NewOutgoing(log *logger.Logger, privateIP string, mappings map[uint32]*common.Mapping, tunnel *tun.Tun, sock *socket.Socket) *Outgoing {
	return &Outgoing{
		tunnel:    tunnel,
		sock:      sock,
		privateIP: net.ParseIP(privateIP).To4(),
		mappings:  mappings,
		quit:      make(chan bool),
	}
}
