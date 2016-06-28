package workers

import (
	"encoding/binary"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/logger"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
)

type Incoming struct {
	tunnel   *tun.Tun
	sock     *socket.Socket
	mappings map[uint32]*common.Mapping
	quit     chan bool
}

func (incoming *Incoming) Resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.IpAddress)

	if mapping, ok := incoming.mappings[dip]; ok {
		return payload, mapping, true
	}

	return payload, nil, false
}

func (incoming *Incoming) Unseal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := mapping.Cipher.Open(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	if err != nil {
		return payload, false
	}

	return payload, true
}

func (incoming *Incoming) Start(queue int) {
	go func() {
		for {
			payload, ok := incoming.sock.Read(queue)
			if !ok {
				continue
			}
			payload, mapping, ok := incoming.Resolve(payload)
			if !ok {
				continue
			}
			payload, ok = incoming.Unseal(payload, mapping)
			if !ok {
				continue
			}
			incoming.tunnel.Write(payload, queue)
		}
	}()
}

func (incoming *Incoming) Stop() {
	go func() {
		incoming.quit <- true
	}()
}

func NewIncoming(log *logger.Logger, privateIP string, mappings map[uint32]*common.Mapping, tunnel *tun.Tun, sock *socket.Socket) *Incoming {
	return &Incoming{
		tunnel:   tunnel,
		sock:     sock,
		mappings: mappings,
		quit:     make(chan bool),
	}
}
