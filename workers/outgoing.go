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
	cfg        *common.Config
	tunnel     inet.Interface
	sock       socket.Socket
	privateIP  []byte
	store      backend.Backend
	stop       bool
	QueueStats []*common.Stats
}

func (outgoing *Outgoing) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.Packet[16:20])

	if mapping, ok := outgoing.store.GetMapping(dip); ok {
		copy(payload.IPAddress, outgoing.privateIP)
		return payload, mapping, true
	}

	return nil, nil, false
}

func (outgoing *Outgoing) seal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := rand.Read(payload.Nonce)
	if err != nil {
		return nil, false
	}

	mapping.Cipher.Seal(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	return payload, true
}

func (outgoing *Outgoing) droppedStats(payload *common.Payload, mapping *common.Mapping, queue int) {
	outgoing.QueueStats[queue].DroppedPackets++
	if payload == nil {
		return
	}

	outgoing.QueueStats[queue].DroppedBytes += uint64(payload.Length)

	if mapping == nil {
		return
	}

	if link, ok := outgoing.QueueStats[queue].Links[mapping.PrivateIP]; !ok {
		outgoing.QueueStats[queue].Links[mapping.PrivateIP] = &common.Stats{
			DroppedPackets: 1,
			DroppedBytes:   uint64(payload.Length),
		}
	} else {
		link.DroppedPackets++
		link.DroppedBytes += uint64(payload.Length)
	}
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

func (outgoing *Outgoing) pipeline(buf []byte, queue int) bool {
	payload, ok := outgoing.tunnel.Read(buf, queue)
	if !ok {
		outgoing.droppedStats(payload, nil, queue)
		return ok
	}
	payload, mapping, ok := outgoing.resolve(payload)
	if !ok {
		outgoing.droppedStats(payload, mapping, queue)
		return ok
	}
	payload, ok = outgoing.seal(payload, mapping)
	if !ok {
		outgoing.droppedStats(payload, mapping, queue)
		return ok
	}
	outgoing.stats(payload, mapping, queue)
	return outgoing.sock.Write(payload, mapping, queue)
}

// Start handling packets
func (outgoing *Outgoing) Start(queue int) {
	go func() {
		buf := make([]byte, common.MaxPacketLength)
		for !outgoing.stop {
			outgoing.pipeline(buf, queue)
		}
	}()
}

// Stop handling packets
func (outgoing *Outgoing) Stop() {
	outgoing.stop = true
}

// NewOutgoing object
func NewOutgoing(cfg *common.Config, store backend.Backend, tunnel inet.Interface, sock socket.Socket) *Outgoing {
	stats := make([]*common.Stats, cfg.NumWorkers)
	for i := 0; i < cfg.NumWorkers; i++ {
		stats[i] = common.NewStats()
	}
	return &Outgoing{
		cfg:        cfg,
		tunnel:     tunnel,
		sock:       sock,
		privateIP:  net.ParseIP(cfg.PrivateIP).To4(),
		store:      store,
		stop:       false,
		QueueStats: stats,
	}
}
