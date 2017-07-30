// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package worker

import (
	"encoding/binary"
	"runtime"

	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/datastore"
	"github.com/Supernomad/quantum/device"
	"github.com/Supernomad/quantum/metric"
	"github.com/Supernomad/quantum/plugin"
	"github.com/Supernomad/quantum/socket"
)

// Outgoing packet struct for handleing packets coming in off of a Device which are destined for a Socket.
type Outgoing struct {
	cfg        *common.Config
	aggregator *metric.Aggregator
	plugins    []plugin.Plugin
	dev        device.Device
	sock       socket.Socket
	store      datastore.Datastore
	stop       bool
}

func (outgoing *Outgoing) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	dip := binary.LittleEndian.Uint32(payload.Packet[16:20])

	if mapping, ok := outgoing.store.Mapping(dip); ok {
		copy(payload.IPAddress, outgoing.cfg.PrivateIP.To4())
		return payload, mapping, true
	}

	return nil, nil, false
}

func (outgoing *Outgoing) stats(dropped bool, queue int, payload *common.Payload, mapping *common.Mapping) {
	metric := &metric.Metric{
		Queue:   queue,
		Type:    metric.Tx,
		Dropped: dropped,
	}

	if payload != nil {
		metric.Bytes += uint64(payload.Length)
	}

	if mapping != nil {
		metric.PrivateIP = mapping.PrivateIP.String()
	}

	outgoing.aggregator.Metrics <- metric
}

func (outgoing *Outgoing) pipeline(buf []byte, queue int) bool {
	payload, ok := outgoing.dev.Read(queue, buf)
	if !ok {
		outgoing.stats(true, queue, payload, nil)
		return ok
	}
	payload, mapping, ok := outgoing.resolve(payload)
	if !ok {
		outgoing.stats(true, queue, payload, mapping)
		return ok
	}
	for i := 0; i < len(outgoing.plugins); i++ {
		payload, mapping, ok = outgoing.plugins[i].Apply(plugin.Outgoing, payload, mapping)
		if !ok {
			outgoing.stats(true, queue, payload, mapping)
			return ok
		}
	}
	ok = outgoing.sock.Write(queue, payload, mapping)
	if !ok {
		outgoing.stats(true, queue, payload, mapping)
		return ok
	}
	outgoing.stats(false, queue, payload, mapping)
	return true
}

// Start handling packets.
func (outgoing *Outgoing) Start(queue int) {
	go func() {
		// We want to pin this routine to a specific thread to reduce switching costs.
		runtime.LockOSThread()

		buf := make([]byte, common.MaxPacketLength)
		for !outgoing.stop {
			outgoing.pipeline(buf, queue)
		}
	}()
}

// Stop handling packets and shutdown.
func (outgoing *Outgoing) Stop() {
	outgoing.stop = true
}

// NewOutgoing generates an Outgoing worker which once started will handle packets coming from the local node destined for remote nodes in the quantum network.
func NewOutgoing(cfg *common.Config, aggregator *metric.Aggregator, store datastore.Datastore, plugins []plugin.Plugin, dev device.Device, sock socket.Socket) *Outgoing {
	return &Outgoing{
		cfg:        cfg,
		aggregator: aggregator,
		plugins:    plugins,
		dev:        dev,
		sock:       sock,
		store:      store,
		stop:       false,
	}
}
