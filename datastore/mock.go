// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package datastore

import (
	"github.com/supernomad/quantum/common"
)

// Mock datastore struct for testing.
type Mock struct {
	InternalMapping        *common.Mapping
	InternalGatewayMapping *common.Mapping
}

// Mapping always returns the internal mapping and true.
func (mock *Mock) Mapping(ip uint32) (*common.Mapping, bool) {
	return mock.InternalMapping, true
}

// GatewayMapping always returns the internal mapping and true.
func (mock *Mock) GatewayMapping() (*common.Mapping, bool) {
	return mock.InternalGatewayMapping, true
}

// Init which is a noop.
func (mock *Mock) Init() error {
	return nil
}

// Start which is a noop.
func (mock *Mock) Start() {
}

// Stop which is a noop.
func (mock *Mock) Stop() {
}

func newMock(cfg *common.Config) (Datastore, error) {
	return &Mock{}, nil
}
