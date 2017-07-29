// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package socket

import (
	"errors"

	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/crypto"
)

// Quic socket struct for managing a multi-queue quic socket.
type Quic struct {
	cfg     *common.Config
	stop    bool
	queues  []int
	servers []*crypto.QuicContext
	clients []*crypto.QuicContext
}

// Close the Quic socket and removes associated network configuration.
func (quic *Quic) Close() error {
	return errors.New("quic socket is not implemented yet")
}

// Queues will return the underlying Quic socket file descriptors.
func (quic *Quic) Queues() []int {
	return quic.queues
}

// Read a packet off the specified Quic socket queue and return a *common.Payload representation of the packet.
func (quic *Quic) Read(queue int, buf []byte) (*common.Payload, bool) {
	return nil, false
}

// Write a *common.Payload to the specified Quic socket queue.
func (quic *Quic) Write(queue int, payload *common.Payload, mapping *common.Mapping) bool {
	return false
}

func newQuic(cfg *common.Config) (*Quic, error) {
	return nil, errors.New("quic socket is not implemented yet")
}
