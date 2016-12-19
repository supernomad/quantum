// Package device mock struct and func's
// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE
package device

import (
	"github.com/Supernomad/quantum/common"
)

// Mock device
type Mock struct {
}

// Name of the device
func (mock *Mock) Name() string {
	return "Mocked Device"
}

// Read a packet off the device
func (mock *Mock) Read(buf []byte, queue int) (*common.Payload, bool) {
	return common.NewTunPayload(buf, common.MTU), true
}

// Write a packet to the device
func (mock *Mock) Write(payload *common.Payload, queue int) bool {
	return true
}

// Open the device
func (mock *Mock) Open() error {
	return nil
}

// Close the device
func (mock *Mock) Close() error {
	return nil
}

// Queues will return the underlying queue fds
func (mock *Mock) Queues() []int {
	return nil
}

func newMock(cfg *common.Config) *Mock {
	return &Mock{}
}
