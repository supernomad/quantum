package socket

import (
	"github.com/Supernomad/quantum/common"
)

const (
	// UDPSock socket type
	UDPSock int = 0
	// TCPSock socket type
	TCPSock int = 1
	// TUNSock socket type
	TUNSock int = 2
	// TAPSock socket type
	TAPSock int = 3
)

// Socket is a generic multi-queue socket interface
type Socket interface {
	Name() string
	Read(buf []byte, queue int) (*common.Payload, bool)
	Write(payload *common.Payload, mapping *common.Mapping, queue int) bool
	Open() error
	Close() error
}

// New Socket object
func New(kind int, cfg *common.Config) Socket {
	switch kind {
	case UDPSock:
		return newUDP(cfg)
	case TUNSock:
		return newTUN(cfg)
	}
	return nil
}
