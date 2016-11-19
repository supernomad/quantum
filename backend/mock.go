package backend

import (
	"github.com/Supernomad/quantum/common"
	"sync"
)

// Mock backend
type Mock struct {
	Mapping *common.Mapping
	wg      *sync.WaitGroup
}

// GetMapping from the mock backend
func (mock *Mock) GetMapping(ip uint32) (*common.Mapping, bool) {
	return mock.Mapping, true
}

// Init the mock backend
func (mock *Mock) Init() error {
	return nil
}

// Start the mock backend
func (mock *Mock) Start(wg *sync.WaitGroup) {
	mock.wg = wg
}

// Stop the mock backend
func (mock *Mock) Stop() {
	mock.wg.Done()
}
