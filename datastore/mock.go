package datastore

import (
	"sync"

	"github.com/Supernomad/quantum/common"
)

// Mock datastore
type Mock struct {
	InternalMapping *common.Mapping

	wg *sync.WaitGroup
}

// GetMapping from the mock datastore
func (mock *Mock) Mapping(ip uint32) (*common.Mapping, bool) {
	return mock.InternalMapping, true
}

// Init the mock datastore
func (mock *Mock) Init() error {
	return nil
}

// Start the mock datastore
func (mock *Mock) Start(wg *sync.WaitGroup) {
	mock.wg = wg
}

// Stop the mock datastore
func (mock *Mock) Stop() {
	mock.wg.Done()
}

func newMock(log *common.Logger, cfg *common.Config) (Datastore, error) {
	return &Mock{}, nil
}
