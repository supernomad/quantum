package workers

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/datastore"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/socket"
)

// Outgoing internal packet interface which handles reading packets off of a TUN object
type Outgoing struct {
	cfg        *common.Config
	aggregator *agg.Agg
	dev        device.Device
	sock       socket.Socket
	store      datastore.Datastore
	stop       bool
}

func (outgoing *Outgoing) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.Packet[16:20])

	if mapping, ok := outgoing.store.Mapping(dip); ok {
		if outgoing.cfg.IsIPv6Enabled && mapping.IPv6 != nil {
			payload.Sockaddr = mapping.SockaddrInet6
		} else if outgoing.cfg.IsIPv4Enabled && mapping.IPv4 != nil {
			payload.Sockaddr = mapping.SockaddrInet4
		} else {
			return nil, nil, false
		}
		copy(payload.IPAddress, outgoing.cfg.PrivateIP.To4())
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

func (outgoing *Outgoing) stats(dropped bool, queue int, payload *common.Payload, mapping *common.Mapping) {
	aggData := &agg.Data{
		Queue:     queue,
		Direction: agg.Outgoing,
		Dropped:   dropped,
	}

	if payload != nil {
		aggData.Bytes += uint64(payload.Length)
	}

	if mapping != nil {
		aggData.PrivateIP = mapping.PrivateIP.String()
	}

	outgoing.aggregator.Aggs <- aggData
}

func (outgoing *Outgoing) pipeline(buf []byte, queue int) bool {
	payload, ok := outgoing.dev.Read(buf, queue)
	if !ok {
		outgoing.stats(true, queue, payload, nil)
		return ok
	}
	payload, mapping, ok := outgoing.resolve(payload)
	if !ok {
		outgoing.stats(true, queue, payload, mapping)
		return ok
	}
	payload, ok = outgoing.seal(payload, mapping)
	if !ok {
		outgoing.stats(true, queue, payload, mapping)
		return ok
	}
	outgoing.stats(false, queue, payload, mapping)
	return outgoing.sock.Write(payload, queue)
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
func NewOutgoing(cfg *common.Config, aggregator *agg.Agg, store datastore.Datastore, dev device.Device, sock socket.Socket) *Outgoing {
	return &Outgoing{
		cfg:        cfg,
		aggregator: aggregator,
		dev:        dev,
		sock:       sock,
		store:      store,
		stop:       false,
	}
}
