package socket

import (
	"github.com/Supernomad/quantum/common"
)

const (
	// UDPSocket socket type
	UDPSocket int = 0
	// IPSocket socket type
	IPSocket int = 1
	// MOCKSocket socket type
	MOCKSocket int = 2

	ipProto = 138
)

// Socket is a generic multi-queue socket interface
type Socket interface {
	Read(buf []byte, queue int) (*common.Payload, bool)
	Write(payload *common.Payload, queue int) bool
	Open() error
	Close() error
	GetFDs() []int
}

// New Socket object
func New(kind int, cfg *common.Config) Socket {
	switch kind {
	case UDPSocket:
		return newUDP(cfg)
	case MOCKSocket:
		return newMock(cfg)
	}
	return nil
}
