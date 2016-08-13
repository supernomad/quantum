package workers

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/inet"
	"github.com/Supernomad/quantum/socket"
	"net"
)

// Outgoing internal packet interface which handles reading packets off of a TUN object
type Outgoing struct {
	tunnel     inet.Interface
	sock       socket.Socket
	privateIP  []byte
	store      *backend.Backend
	quit       chan bool
	QueueStats []*common.Stats
}

func (outgoing *Outgoing) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.Packet[16:20])

	if mapping, ok := outgoing.store.GetMapping(dip); ok {
		copy(payload.IPAddress, outgoing.privateIP)
		return payload, mapping, true
	}

	return payload, nil, false
}

func (outgoing *Outgoing) seal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := rand.Read(payload.Nonce)
	if err != nil {
		return payload, false
	}

	mapping.Cipher.Seal(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	return payload, true
}

func (outgoing *Outgoing) stats(payload *common.Payload, mapping *common.Mapping, queue int) {
	outgoing.QueueStats[queue].Packets++
	outgoing.QueueStats[queue].Bytes += uint64(payload.Length)

	if link, ok := outgoing.QueueStats[queue].Links[mapping.PrivateIP]; !ok {
		outgoing.QueueStats[queue].Links[mapping.PrivateIP] = &common.Stats{
			Packets: 1,
			Bytes:   uint64(payload.Length),
		}
	} else {
		link.Packets++
		link.Bytes += uint64(payload.Length)
	}
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
			payload, mapping, ok := outgoing.resolve(payload)
			if !ok {
				continue
			}
			payload, ok = outgoing.seal(payload, mapping)
			if !ok {
				continue
			}
			outgoing.stats(payload, mapping, queue)
			outgoing.sock.Write(payload, mapping, queue)
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
func NewOutgoing(privateIP string, numWorkers int, store *backend.Backend, tunnel inet.Interface, sock socket.Socket) *Outgoing {
	stats := make([]*common.Stats, numWorkers)
	for i := 0; i < numWorkers; i++ {
		stats[i] = common.NewStats()
	}
	return &Outgoing{
		tunnel:     tunnel,
		sock:       sock,
		privateIP:  net.ParseIP(privateIP).To4(),
		store:      store,
		quit:       make(chan bool),
		QueueStats: stats,
	}
}
