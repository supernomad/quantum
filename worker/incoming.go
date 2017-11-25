// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package worker

import (
	"runtime"

	"github.com/supernomad/quantum/common"
	"github.com/supernomad/quantum/device"
	"github.com/supernomad/quantum/metric"
	"github.com/supernomad/quantum/plugin"
	"github.com/supernomad/quantum/router"
	"github.com/supernomad/quantum/socket"
)

// Incoming packet struct for handleing packets coming in off of a Socket struct which are destined for a Device struct.
type Incoming struct {
	cfg        *common.Config
	aggregator *metric.Aggregator
	plugins    []plugin.Plugin
	dev        device.Device
	sock       socket.Socket
	router     *router.Router
	stop       bool
}

func (incoming *Incoming) resolve(payload *common.Payload) (*common.Payload, *common.Mapping, bool) {
	if mapping, ok := incoming.router.Resolve(payload.IPAddress); ok {
		return payload, mapping, true
	}

	return nil, nil, false
}

func (incoming *Incoming) stats(dropped bool, queue int, payload *common.Payload, mapping *common.Mapping) {
	metric := &metric.Metric{
		Queue:   queue,
		Type:    metric.Rx,
		Dropped: dropped,
	}

	if payload != nil {
		metric.Bytes += uint64(payload.Length)
	}

	if mapping != nil {
		metric.PrivateIP = mapping.PrivateIP.String()
	}

	incoming.aggregator.Metrics <- metric
}

func (incoming *Incoming) pipeline(buf []byte, queue int) bool {
	payload, ok := incoming.sock.Read(queue, buf)
	if !ok {
		incoming.stats(true, queue, payload, nil)
		return ok
	}
	payload, mapping, ok := incoming.resolve(payload)
	if !ok {
		incoming.stats(true, queue, payload, mapping)
		return ok
	}
	for i := 0; i < len(incoming.plugins); i++ {
		payload, mapping, ok = incoming.plugins[i].Apply(plugin.Incoming, payload, mapping)
		if !ok {
			incoming.stats(true, queue, payload, mapping)
			return ok
		}
	}
	ok = incoming.dev.Write(queue, payload)
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
		// We want to pin this routine to a specific thread to reduce switching costs.
		runtime.LockOSThread()

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
func NewIncoming(cfg *common.Config, aggregator *metric.Aggregator, rt *router.Router, plugins []plugin.Plugin, dev device.Device, sock socket.Socket) *Incoming {
	return &Incoming{
		cfg:        cfg,
		aggregator: aggregator,
		plugins:    plugins,
		dev:        dev,
		sock:       sock,
		router:     rt,
		stop:       false,
	}
}
