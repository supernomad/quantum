package workers

import (
	"encoding/binary"
	"fmt"
	"github.com/Supernomad/quantum/backend"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/inet"
	"github.com/Supernomad/quantum/socket"
	"sync"
)

// Incoming external packet interface which handles reading packets off of a Socket object
type Incoming struct {
	tunnel     inet.Interface
	sock       socket.Socket
	store      backend.Backend
	stop       bool
	QueueStats []*common.Stats
}

func (incoming *Incoming) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.IPAddress)

	if mapping, ok := incoming.store.GetMapping(dip); ok {
		return payload, mapping, true
	}

	return nil, nil, false
}

func (incoming *Incoming) unseal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := mapping.Cipher.Open(payload.Packet[:0], payload.Nonce, payload.Packet, nil)
	if err != nil {
		return nil, false
	}

	return payload, true
}

func (incoming *Incoming) droppedStats(payload *common.Payload, mapping *common.Mapping, queue int) {
	incoming.QueueStats[queue].DroppedPackets++

	if payload == nil {
		return
	}

	incoming.QueueStats[queue].DroppedBytes += uint64(payload.Length)

	if mapping == nil {
		return
	}

	if link, ok := incoming.QueueStats[queue].Links[mapping.PrivateIP]; !ok {
		incoming.QueueStats[queue].Links[mapping.PrivateIP] = &common.Stats{
			DroppedPackets: 1,
			DroppedBytes:   uint64(payload.Length),
		}
	} else {
		link.DroppedPackets++
		link.DroppedBytes += uint64(payload.Length)
	}
}

func (incoming *Incoming) stats(payload *common.Payload, mapping *common.Mapping, queue int) {
	incoming.QueueStats[queue].Packets++
	incoming.QueueStats[queue].Bytes += uint64(payload.Length)

	if link, ok := incoming.QueueStats[queue].Links[mapping.PrivateIP]; !ok {
		incoming.QueueStats[queue].Links[mapping.PrivateIP] = &common.Stats{
			Packets: 1,
			Bytes:   uint64(payload.Length),
		}
	} else {
		link.Packets++
		link.Bytes += uint64(payload.Length)
	}
}

func (incoming *Incoming) pipeline(buf []byte, queue int) bool {
	payload, ok := incoming.sock.Read(buf, queue)
	if !ok {
		incoming.droppedStats(payload, nil, queue)
		return ok
	}
	payload, mapping, ok := incoming.resolve(payload)
	if !ok {
		incoming.droppedStats(payload, mapping, queue)
		return ok
	}
	payload, ok = incoming.unseal(payload, mapping)
	if !ok {
		incoming.droppedStats(payload, mapping, queue)
		return ok
	}
	incoming.stats(payload, mapping, queue)
	return incoming.tunnel.Write(payload, queue)
}

// Start handling packets
func (incoming *Incoming) Start(queue int, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		buf := make([]byte, common.MaxPacketLength)
		for !incoming.stop {
			incoming.pipeline(buf, queue)
		}
		fmt.Println("[INCOMING]", "Queue:", queue, "Exiting")
	}()
}

// Stop handling packets
func (incoming *Incoming) Stop() {
	incoming.stop = true
}

// NewIncoming object
func NewIncoming(privateIP string, numWorkers int, store backend.Backend, tunnel inet.Interface, sock socket.Socket) *Incoming {
	stats := make([]*common.Stats, numWorkers)
	for i := 0; i < numWorkers; i++ {
		stats[i] = common.NewStats()
	}
	return &Incoming{
		tunnel:     tunnel,
		sock:       sock,
		store:      store,
		stop:       false,
		QueueStats: stats,
	}
}
