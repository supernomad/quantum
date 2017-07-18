// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package plugin

import (
	"github.com/Supernomad/quantum/common"
)

// Mock plugin struct to use for testing.
type Mock struct {
}

// Apply returns the payload/mapping unchanged and always true.
func (mock *Mock) Apply(direction Direction, payload *common.Payload, mapping *common.Mapping) (*common.Payload, *common.Mapping, bool) {
	return payload, mapping, true
}

// Close which is a noop.
func (mock *Mock) Close() error {
	return nil
}

// Name returns 'mock'.
func (mock *Mock) Name() string {
	return MockPlugin
}

// Order returns the MockPluginOrder value.
func (mock *Mock) Order() int {
	return MockPluginOrder
}

func newMock(cfg *common.Config) (Plugin, error) {
	return &Mock{}, nil
}
