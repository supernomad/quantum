package inet

import (
	"github.com/Supernomad/quantum/common"
)

const (
	// TUNInterface interface type
	TUNInterface int = 0
	// TAPInterface interface type
	TAPInterface  int = 1
	mockInterface int = 2
)

// Interface is a generic multi-queue network interface
type Interface interface {
	Name() string
	Read(buf []byte, queue int) (*common.Payload, bool)
	Write(payload *common.Payload, queue int) bool
	Open() error
	Close() error
}

// New Interface object
func New(kind int, cfg *common.Config) Interface {
	switch kind {
	case TUNInterface:
		return newTUN(cfg)
	case mockInterface:
		return newMock(cfg)
	}
	return nil
}
