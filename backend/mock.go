package backend

import (
	"fmt"
	"github.com/Supernomad/quantum/common"
	"sync"
)

// Mock backend
type Mock struct {
	Mappings map[uint32]*common.Mapping
	wg       *sync.WaitGroup
}

// GetMapping from the mock backend
func (mock *Mock) GetMapping(ip uint32) (*common.Mapping, bool) {
	val, ok := mock.Mappings[ip]
	fmt.Println(val, ok)
	return val, ok
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
