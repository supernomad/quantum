package inet

import (
	"github.com/Supernomad/quantum/common"
)

// Mock interface
type Mock struct {
}

// Name of the interface
func (mock *Mock) Name() string {
	return "Mocked Interface"
}

// Read a packet off the interface
func (mock *Mock) Read(buf []byte, queue int) (*common.Payload, bool) {
	return nil, false
}

// Write a packet to the interface
func (mock *Mock) Write(payload *common.Payload, queue int) bool {
	return false
}

// Open the interface
func (mock *Mock) Open() error {
	return nil
}

// Close the interface
func (mock *Mock) Close() error {
	return nil
}

func newMock(cfg *common.Config) *Mock {
	return &Mock{}
}
