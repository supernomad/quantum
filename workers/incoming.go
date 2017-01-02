// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package workers

import (
	"encoding/binary"

	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/datastore"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/socket"
)

// Incoming packet struct for handleing packets coming in off of a Socket struct which are destined for a Device struct.
type Incoming struct {
	cfg        *common.Config
	aggregator *agg.Agg
	dev        device.Device
	sock       socket.Socket
	store      datastore.Datastore
	stop       bool
}

func (incoming *Incoming) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.IPAddress)

	if mapping, ok := incoming.store.Mapping(dip); ok {
		return payload, mapping, true
	}

	return nil, nil, false
}

func (incoming *Incoming) unseal(payload *common.Payload, mapping *common.Mapping) (*common.Payload, bool) {
	_, err := mapping.Cipher.Open(payload.Packet[:0], payload.Nonce, payload.Packet, payload.IPAddress)
	if err != nil {
		return nil, false
	}

	return payload, true
}

func (incoming *Incoming) stats(dropped bool, queue int, payload *common.Payload, mapping *common.Mapping) {
	aggStat := &common.Stat{
		Queue:     queue,
		Direction: common.IncomingStat,
		Dropped:   dropped,
	}

	if payload != nil {
		aggStat.Bytes += uint64(payload.Length)
	}

	if mapping != nil {
		aggStat.PrivateIP = mapping.PrivateIP.String()
	}

	incoming.aggregator.Aggs <- aggStat
}

func (incoming *Incoming) pipeline(buf []byte, queue int) bool {
	payload, ok := incoming.sock.Read(buf, queue)
	if !ok {
		incoming.stats(true, queue, payload, nil)
		return ok
	}
	payload, mapping, ok := incoming.resolve(payload)
	if !ok {
		incoming.stats(true, queue, payload, mapping)
		return ok
	}
	payload, ok = incoming.unseal(payload, mapping)
	if !ok {
		incoming.stats(true, queue, payload, mapping)
		return ok
	}
	ok = incoming.dev.Write(payload, queue)
	if !ok {
		incoming.stats(true, queue, payload, mapping)
		return ok
	}
	incoming.stats(false, queue, payload, mapping)
	return true
}

// Start handling packets.
func (incoming *Incoming) Start(queue int) {
	go func() {
		buf := make([]byte, common.MaxPacketLength)
		for !incoming.stop {
			incoming.pipeline(buf, queue)
		}
	}()
}

// Stop handling packets.
func (incoming *Incoming) Stop() {
	incoming.stop = true
}

// NewIncoming generates a new Incoming worker which once started will handle packets coming from the remote nodes in the quantum network destined for the local node.
func NewIncoming(cfg *common.Config, aggregator *agg.Agg, store datastore.Datastore, dev device.Device, sock socket.Socket) *Incoming {
	return &Incoming{
		cfg:        cfg,
		aggregator: aggregator,
		dev:        dev,
		sock:       sock,
		store:      store,
		stop:       false,
	}
}
