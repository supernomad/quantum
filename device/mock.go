// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package device

import (
	"github.com/Supernomad/quantum/common"
)

// Mock device struct to use for testing.
type Mock struct {
}

// Name of the mock device.
func (mock *Mock) Name() string {
	return "Mocked Device"
}

// Read which just returns the supplied buffer in the form of a *common.Payload.
func (mock *Mock) Read(buf []byte, queue int) (*common.Payload, bool) {
	return common.NewTunPayload(buf, common.MTU), true
}

// Write which is a noop.
func (mock *Mock) Write(payload *common.Payload, queue int) bool {
	return true
}

// Close which is a noop.
func (mock *Mock) Close() error {
	return nil
}

// Queues which is a noop.
func (mock *Mock) Queues() []int {
	return nil
}

func newMock(cfg *common.Config) (Device, error) {
	return &Mock{}, nil
}
