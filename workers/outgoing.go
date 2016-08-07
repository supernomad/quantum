package workers

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/socket"
	"github.com/Supernomad/quantum/tun"
	"net"
)

// Outgoing internal packet interface which handles reading packets off of a TUN object
type Outgoing struct {
	tunnel    *tun.Tun
	sock      *socket.Socket
	privateIP []byte
	store     *backend.Backend
	quit      chan bool
}

// Resolve the outgoing payload
func (outgoing *Outgoing) Resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.Packet[16:20])

	if mapping, ok := outgoing.store.GetMapping(dip); ok {
		copy(payload.IPAddress, outgoing.privateIP)
		return payload, mapping, true
	}

	return payload, nil, false
}

// Seal the outgoing payload
func (outgoing *Outgoing) Seal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := rand.Read(payload.Nonce)
	if err != nil {
		return payload, false
	}

	mapping.Cipher.Seal(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	return payload, true
}

// Start handling packets
func (outgoing *Outgoing) Start(queue int) {
	go func() {
		buf := make([]byte, common.MaxPacketLength)
		for {
			payload, ok := outgoing.tunnel.Read(buf, queue)
			if !ok {
				continue
			}
			payload, mapping, ok := outgoing.Resolve(payload)
			if !ok {
				continue
			}
			payload, ok = outgoing.Seal(payload, mapping)
			if !ok {
				continue
			}
			outgoing.sock.Write(payload, mapping.Sockaddr, queue)
		}
	}()
}

// Stop handling packets
func (outgoing *Outgoing) Stop() {
	go func() {
		outgoing.quit <- true
	}()
}

// NewOutgoing object
func NewOutgoing(privateIP string, store *backend.Backend, tunnel *tun.Tun, sock *socket.Socket) *Outgoing {
	return &Outgoing{
		tunnel:    tunnel,
		sock:      sock,
		privateIP: net.ParseIP(privateIP).To4(),
		store:     store,
		quit:      make(chan bool),
	}
}
