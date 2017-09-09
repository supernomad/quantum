// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package socket

import (
	"errors"
	"syscall"

	"github.com/supernomad/quantum/common"
)

// UDP socket struct for managing a multi-queue udp socket.
type UDP struct {
	cfg    *common.Config
	queues []int
}

// Close the UDP socket and removes associated network configuration.
func (udp *UDP) Close() error {
	for i := 0; i < len(udp.queues); i++ {
		if err := syscall.Close(udp.queues[i]); err != nil {
			return errors.New("error closing the socket queues: " + err.Error())
		}
	}
	return nil
}

// Queues will return the underlying UDP socket file descriptors.
func (udp *UDP) Queues() []int {
	return udp.queues
}

// Read a packet off the specified UDP socket queue and return a *common.Payload representation of the packet.
func (udp *UDP) Read(queue int, buf []byte) (*common.Payload, bool) {
	n, _, err := syscall.Recvfrom(udp.queues[queue], buf, 0)
	if err != nil {
		return nil, false
	}
	return common.NewSockPayload(buf, n), true
}

// Write a *common.Payload to the specified UDP socket queue.
func (udp *UDP) Write(queue int, payload *common.Payload, mapping *common.Mapping) bool {
	err := syscall.Sendto(udp.queues[queue], payload.Raw[:payload.Length], 0, mapping.Sockaddr)
	return err == nil
}

func newUDP(cfg *common.Config) (*UDP, error) {
	udp := &UDP{
		cfg:    cfg,
		queues: make([]int, cfg.NumWorkers),
	}

	for i := 0; i < udp.cfg.NumWorkers; i++ {
		var queue int
		var err error

		if !udp.cfg.ReuseFDS {
			queue, err = createUDPSocket(udp.cfg.IsIPv6Enabled, udp.cfg.ListenAddr)
			if err != nil {
				return udp, errors.New("error creating the UDP socket: " + err.Error())
			}
		} else {
			queue = 3 + udp.cfg.NumWorkers + i
		}
		udp.queues[i] = queue
	}
	return udp, nil
}
