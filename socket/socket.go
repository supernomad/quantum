// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

// Package socket contains the structs and logic to create, maintain, and operate kernel level sockets
package socket

import (
	"github.com/Supernomad/quantum/common"
)

const (
	// UDPSocket socket type
	UDPSocket int = 0
	// MOCKSocket socket type
	MOCKSocket int = 2
)

// Socket interface for a generic multi-queue socket interface
type Socket interface {
	Read(buf []byte, queue int) (*common.Payload, bool)
	Write(payload *common.Payload, queue int) bool
	Open() error
	Close() error
	GetFDs() []int
}

// New generate a new Socket struct based on the supplied device kind and user configuration
func New(kind int, cfg *common.Config) Socket {
	switch kind {
	case UDPSocket:
		return newUDP(cfg)
	case MOCKSocket:
		return newMock(cfg)
	}
	return nil
}
