package backend

import (
	"github.com/Supernomad/quantum/common"
)

// Mock backend
type Mock struct {
	Mappings map[uint32]*common.Mapping
}

// GetMapping from the mock backend
func (mock *Mock) GetMapping(ip uint32) (*common.Mapping, bool) {
	val, ok := mock.Mappings[ip]
	return val, ok
}

// Init the mock backend
func (mock *Mock) Init() error {
	return nil
}

// Start the mock backend
func (mock *Mock) Start() {

}

// Stop the mock backend
func (mock *Mock) Stop() {

}
