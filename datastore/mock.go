// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"sync"

	"github.com/Supernomad/quantum/common"
)

// Mock datastore struct for testing.
type Mock struct {
	InternalMapping *common.Mapping

	wg *sync.WaitGroup
}

// Mapping always returns the internal mapping and true.
func (mock *Mock) Mapping(ip uint32) (*common.Mapping, bool) {
	return mock.InternalMapping, true
}

// Init which is a noop.
func (mock *Mock) Init() error {
	return nil
}

// Start which is a noop.
func (mock *Mock) Start(wg *sync.WaitGroup) {
	mock.wg = wg
}

// Stop which is a noop.
func (mock *Mock) Stop() {
	mock.wg.Done()
}

func newMock(log *common.Logger, cfg *common.Config) (Datastore, error) {
	return &Mock{}, nil
}
