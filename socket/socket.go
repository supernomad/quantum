// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package socket

import (
	"github.com/Supernomad/quantum/common"
)

// Type defines what kind of socket to use.
type Type int

const (
	// UDPSocket type creates and manages a UDP based socket.
	UDPSocket Type = iota

	// MOCKSocket type creates and manages a mocked out socket for testing.
	MOCKSocket
)

// Socket interface for a generic multi-queue socket interface.
type Socket interface {
	Read(buf []byte, queue int) (*common.Payload, bool)
	Write(payload *common.Payload, queue int) bool
	Open() error
	Close() error
	GetFDs() []int
}

// New generate a new Socket struct based on the supplied device socketType and user configuration.
func New(socketType Type, cfg *common.Config) Socket {
	switch socketType {
	case UDPSocket:
		return newUDP(cfg)
	case MOCKSocket:
		return newMock(cfg)
	}
	return nil
}
