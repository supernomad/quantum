package socket

import (
	"github.com/Supernomad/quantum/common"
)

// Mock socket
type Mock struct {
}

// Name of the socket
func (mock *Mock) Name() string {
	return "Mocked Socket"
}

// Read a packet off the socket
func (mock *Mock) Read(buf []byte, queue int) (*common.Payload, bool) {
	return common.NewSockPayload(buf, len(buf)), true
}

// Write a packet to the socket
func (mock *Mock) Write(payload *common.Payload, mapping *common.Mapping, queue int) bool {
	return true
}

// Open the socket
func (mock *Mock) Open() error {
	return nil
}

// Close the socket
func (mock *Mock) Close() error {
	return nil
}

// GetFDs will return the underlying queue fds
func (mock *Mock) GetFDs() []int {
	return nil
}

func newMock(cfg *common.Config) *Mock {
	return &Mock{}
}
