package workers

import (
	"encoding/binary"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
)

// Incoming external packet interface which handles reading packets off of a Socket object
type Incoming struct {
	tunnel *tun.Tun
	sock   *socket.Socket
	store  *backend.Backend
	quit   chan bool
}

// Resolve the incoming payload
func (incoming *Incoming) Resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.IPAddress)

	if mapping, ok := incoming.store.GetMapping(dip); ok {
		return payload, mapping, true
	}

	return payload, nil, false
}

// Unseal the incoming payload
func (incoming *Incoming) Unseal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := mapping.Cipher.Open(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	if err != nil {
		return payload, false
	}

	return payload, true
}

// Start handling packets
func (incoming *Incoming) Start(queue int) {
	go func() {
		buf := make([]byte, common.MaxPacketLength)
		for {
			payload, ok := incoming.sock.Read(buf, queue)
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

// Stop handling packets
func (incoming *Incoming) Stop() {
	go func() {
		incoming.quit <- true
	}()
}

// NewIncoming object
func NewIncoming(privateIP string, store *backend.Backend, tunnel *tun.Tun, sock *socket.Socket) *Incoming {
	return &Incoming{
		tunnel: tunnel,
		sock:   sock,
		store:  store,
		quit:   make(chan bool),
	}
}
